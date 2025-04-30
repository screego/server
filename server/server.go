package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// Config holds server configuration parameters
type Config struct {
	Address         string        // Listen address (e.g., ":8080" or "unix:/tmp/server.sock")
	CertFile        string        // TLS certificate file (optional)
	KeyFile         string        // TLS private key file (optional)
	ShutdownTimeout time.Duration // Graceful shutdown timeout (default: 2 seconds)
}

var (
	// Extracted for testing
	notifySignal   = signal.Notify
	serverShutdown = (*http.Server).Shutdown
)

// Start initializes and runs the HTTP server with graceful shutdown.
func Start(router *mux.Router, cfg Config) error {
	// Set default shutdown timeout if not specified
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 2 * time.Second
	}

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	listener, err := createListener(cfg.Address)
	if err != nil {
		return err
	}

	// Channel for server errors
	serverErrors := make(chan error, 1)

	// Start server
	go func() {
		log.Info().
			Str("address", cfg.Address).
			Bool("tls", cfg.CertFile != "" && cfg.KeyFile != "").
			Msg("Starting server")

		var serveErr error
		if cfg.CertFile != "" && cfg.KeyFile != "" {
			serveErr = server.ServeTLS(listener, cfg.CertFile, cfg.KeyFile)
		} else {
			serveErr = server.Serve(listener)
		}

		if serveErr != nil && serveErr != http.ErrServerClosed {
			serverErrors <- serveErr
		}
	}()

	// Handle shutdown signals
	interrupt := make(chan os.Signal, 1)
	notifySignal(interrupt, os.Interrupt)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		select {
		case <-interrupt:
			log.Info().Msg("Received interrupt signal, shutting down...")
			ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				log.Error().Err(err).Msg("Graceful shutdown failed")
				serverErrors <- err
			}
		case err := <-serverErrors:
			if err != nil {
				log.Error().Err(err).Msg("Server error occurred")
			}
		}
	}()

	wg.Wait()
	return nil
}

// createListener creates a TCP or Unix domain socket listener.
// Returns an error if the Unix socket already exists (safer than auto-removal).
func createListener(address string) (net.Listener, error) {
	if strings.HasPrefix(address, "unix:") {
		socketPath := strings.TrimPrefix(address, "unix:")
		return net.Listen("unix", socketPath) // Let OS return error if socket exists
	}
	return net.Listen("tcp", address)
}
