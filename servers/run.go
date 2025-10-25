package servers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// RunWithGracefulShutdown starts the server and handles SIGINT/SIGTERM gracefully.
// - server: the http.Server to run
// - appName: for logging
// - cleanup: optional cleanup function to release resources
// - timeout: max duration for shutdown
func RunWithGracefulShutdown(server *http.Server, appName string, cleanup func(), timeout time.Duration) error {
	// Channel to capture server errors
	serverErrChan := make(chan error, 1) // error channel

	// Start the server in a goroutine
	go func() {
		log.Printf("[INFO] %q listening on %s ...", appName, server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		} else {
			serverErrChan <- nil
		}
	}()

	// Channel to listen for OS signals
	osSignalChan := make(chan os.Signal, 1) // os.Signal channel
	signal.Notify(osSignalChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal
	sig := <-osSignalChan
	log.Printf("[INFO] got signal [%s]. shutting down the app [%s] ...", sig, appName)

	// Run cleanup if provided
	if cleanup != nil {
		cleanup()
	}

	// Shutdown HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// Server Shutdown to Stop Accepting New HTTP Requests Immediately
	// But with the context with timeout, requests already being processed get time to finish
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] server shutdown failed: %v", err)
	}

	// Wait for server goroutine to return
	if err := <-serverErrChan; err != nil {
		return err
	}

	log.Printf("[INFO] %q shutdown complete", appName)
	return nil
}
