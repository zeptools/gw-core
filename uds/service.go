package uds

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

	"github.com/zeptools/gw-core/svc"
)

type Service struct {
	Ctx        context.Context    // Service Context
	cancel     context.CancelFunc // Service Context CancelFunc
	state      int                // internal service state
	done       chan error         // Shutdown Error Channel
	SocketPath string
	CmdMap     map[string]CmdHnd
	listener   net.Listener
}

func (s *Service) Name() string {
	return "UDSService"
}

func NewService(parentCtx context.Context, sockPath string, cmdMap map[string]CmdHnd) *Service {
	svcCtx, svcCancel := context.WithCancel(parentCtx)
	return &Service{
		Ctx:        svcCtx,
		cancel:     svcCancel,
		state:      svc.StateREADY,
		done:       make(chan error, 1),
		SocketPath: sockPath,
		CmdMap:     cmdMap,
	}
}

// Start the unix socket service in the background.
// Bootstrapping errors are returned immediately.
// Runtime errors are pushed into Done().
func (s *Service) Start() error {
	// clean up old socket if any
	_ = os.Remove(s.SocketPath)
	listener, err := net.Listen("unix", s.SocketPath)
	if err != nil {
		return fmt.Errorf("listen(%q) failed: %v", s.SocketPath, err)
	}
	s.listener = listener
	// tighten permissions immediately after binding
	if err = os.Chmod(s.SocketPath, 0600); err != nil {
		_ = s.listener.Close()
		_ = os.Remove(s.SocketPath)
		return fmt.Errorf("chmod(%q) failed: %w", s.SocketPath, err)
	}
	go s.run()
	return nil
}

func (s *Service) Stop() {
	s.cancel()
	s.state = svc.StateSTOPPED
	log.Println("[INFO][UDS] service stopped")
}

func (s *Service) Done() <-chan error {
	return s.done
}

// run - internal run loop
func (s *Service) run() {
	// goroutine to clean up when context is done
	go func() {
		<-s.Ctx.Done()
		log.Printf("[INFO][UDS] stopping")
		if err := s.listener.Close(); err != nil {
			log.Printf("[ERROR][UDS] cannot close listener: %v", err)
		}
		// To avoid TOCTOU race, just try removing before checking if it exists.
		if err := os.Remove(s.SocketPath); err != nil && !os.IsNotExist(err) {
			log.Printf("[ERROR][UDS] cannot remove socket file: %v", err)
		}
	}()

	// --- Serving loop ---
	log.Printf("[INFO][UDS] listening on %q ...\n", s.SocketPath)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Printf("[INFO][UDS] socket closed")
				s.done <- nil // also a clean shutdown
				return
			}
			// For transient errors, don’t kill the loop
			log.Println("[ERROR][UDS] accept failed:", err)
			continue
		}
		log.Println("[INFO][UDS] new connection")
		go s.handleConn(conn)
	}
}

func (s *Service) handleConn(c net.Conn) {
	go func() {
		<-s.Ctx.Done()
		_ = c.Close()
	}()

	defer func() {
		if err := c.Close(); err != nil {
			if !errors.Is(err, net.ErrClosed) { // && !strings.Contains(err.Error(), "use of closed network connection")
				log.Printf("[ERROR][UDS] closing connection: %v\n", err)
			}
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
		args := strings.Fields(line)
		cmdStr := args[0]
		if cmdStr == "quit" {
			return
		}
		if cmdStr == "help" {
			_, _ = fmt.Fprintln(c, "")
			for cmdKey, cmdHnd := range s.CmdMap {
				_, _ = fmt.Fprintf(c, "%-36s %s\n", cmdKey, cmdHnd.Desc)
			}
			_, _ = fmt.Fprintln(c, "")
			continue
		}
		// look it up in the command map
		if cmdHnd, ok := s.CmdMap[cmdStr]; ok {
			log.Printf("[INFO][UDS] requested command `%s`\n", line)
			cmdHnd.Fn(args[1:], c)
			log.Printf("[INFO][UDS] command `%s` done\n", line)
			return
		} else {
			_, _ = fmt.Fprintf(c, "unknown command: %s\n", cmdStr)
			continue // give another chance
		}
	}

}
