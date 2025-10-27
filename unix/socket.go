package unix

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type CmdHnd struct {
	Fn    func(ctx context.Context, args []string, out io.Writer)
	Desc  string
	Usage string
}

type Socket struct {
	Ctx        context.Context    // Service Context
	Cancel     context.CancelFunc // Service Context CancelFunc
	SocketPath string
	CmdMap     map[string]CmdHnd
	listener   net.Listener
}

func NewSocket(parentCtx context.Context, sockPath string) *Socket {
	svcCtx, svcCancel := context.WithCancel(parentCtx)
	return &Socket{
		Ctx:        svcCtx,
		Cancel:     svcCancel,
		SocketPath: sockPath,
		// ToDo: Set Command Map
	}
}

// StartService the socket service asynchronously in a background goroutine
// It returns immediately and the service will continue to run until the provided context is canceled.
// Typical usage is in applications that have other long-running services
// (e.g. HTTP server) where the admin socket is just an auxiliary control channel.
func (s *Socket) StartService() {
	go s.Run()
}

// Run starts the service and blocks until the provided context is canceled or the listener is closed
// This method is synchronous and will not return until shutdown
// Typical usage is in tests or when the socket should be the primary long-running process.
// The method will clean up the socket file on exit.
func (s *Socket) Run() {
	// clean up old socket if any
	_ = os.Remove(s.SocketPath)
	var err error
	s.listener, err = net.Listen("unix", s.SocketPath)
	if err != nil {
		log.Fatalf("[FATAL][UDS] listen(%q) failed: %v", s.SocketPath, err)
	}

	// tighten permissions immediately after binding
	if err = os.Chmod(s.SocketPath, 0600); err != nil {
		_ = s.listener.Close()
		_ = os.Remove(s.SocketPath)
		log.Fatalf("[FATAL][UDS] chmod(%q) failed: %v", s.SocketPath, err)
	}

	// goroutine for conext done to close the listener and clean up
	go func() {
		<-s.Ctx.Done()
		if err := s.listener.Close(); err != nil {
			log.Printf("[ERROR][UDS] cannot close listener: %v", err)
		}
		if err := os.Remove(s.SocketPath); err != nil {
			log.Printf("[ERROR][UDS] cannot remove socket file: %v", err)
		}
	}()

	// --- Serving loop ---
	log.Printf("[INFO][UDS] listening on %q ...\n", s.SocketPath)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// when closed via ctx, return ctx.Err()
			if s.Ctx.Err() != nil {
				log.Printf("[INFO][UDS] shutting down by the context. %v", s.Ctx.Err())
				return
			}
			if errors.Is(err, net.ErrClosed) {
				log.Printf("[INFO][UDS] admin socket closed")
				return
			}
			log.Println("[ERROR][UDS] accept failed:", err)
			continue
		}
		log.Println("[INFO][UDS] new connection")
		go handleAdmin(conn)
	}
}

func handleAdmin(c net.Conn) {
	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("[ERROR][UDS] closing admin connection: %v\n", err)
		} else {
			log.Println("[INFO][UDS] connection closed")
		}
	}()

	reader := bufio.NewReader(io.LimitReader(c, 1<<20)) // 1 MB max per line

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("[INFO][UDS] client disconnected")
			} else {
				log.Printf("[ERROR][UDS] read error: %v\n", err)
			}
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// args := strings.Fields(line)
		log.Printf("[INFO][UDS] requested command: `%s`\n", line)

		// Temp echoing. ToDo: map args[0] as a command in command map
		if _, err := fmt.Fprintf(c, "%s\n", line); err != nil {
			log.Printf("[ERROR][UDS] %v", err)
		}
	}

}
