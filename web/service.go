package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zeptools/gw-core/svc"
)

type Service struct {
	Ctx    context.Context    // Service Context
	cancel context.CancelFunc // Service Context CancelFunc
	state  int                // internal service state
	done   chan error         // Shutdown Error Channel
	Server *http.Server
}

func (s *Service) Name() string {
	return "WebService"
}

func NewService(parentCtx context.Context, addr string, router http.Handler) *Service {
	svcCtx, svcCancel := context.WithCancel(parentCtx)
	return &Service{
		Ctx:    svcCtx,
		cancel: svcCancel,
		state:  svc.StateREADY,
		done:   make(chan error, 1),
		Server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}

func (s *Service) Start() error {
	if s.state == svc.StateRUNNING {
		return fmt.Errorf("already started")
	}
	if s.state != svc.StateREADY {
		return fmt.Errorf("cannot start. not ready")
	}
	s.state = svc.StateRUNNING
	log.Println("[INFO][HTTP] service started")
	go s.run()
	return nil
}

func (s *Service) Stop() {
	if s.state != svc.StateRUNNING {
		log.Println("[ERROR][HTTP] cannot stop. not running")
		return
	}
	s.cancel()
	s.state = svc.StateSTOPPED
	log.Println("[INFO][HTTP] service stopped")
}

func (s *Service) Done() <-chan error {
	return s.done
}

func (s *Service) run() {
	// clean up with graceful shutdown
	go func() {
		<-s.Ctx.Done()
		log.Println("[INFO][HTTPServer] stopping...")
		gracefulShutdownCtx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		if err := s.Server.Shutdown(gracefulShutdownCtx); err != nil {
			log.Printf("[ERROR][HTTPServer] shutdown failed: %v", err)
		}
	}()
	// run the server in background, push the server error to the error channel
	go func() {
		log.Printf("[INFO][HTTPServer] listening on %s ...", s.Server.Addr)
		if err := s.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.done <- err
		} else {
			s.done <- nil
		}
	}()
}
