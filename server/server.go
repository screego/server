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

var (
	// These variables are extracted for testing purposes
	notifySignal   = signal.Notify
	serverShutdown = (*http.Server).Shutdown
)

// Config holds server configuration parameters
type Config struct {
	Address      string
	CertFile     string
	KeyFile      string
	ShutdownWait time.Duration
}

// Start initializes and runs the HTTP server with graceful shutdown capabilities.
func Start(router *mux.Router, cfg Config) error {
	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	// Channel to receive server errors
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		listener, err := createListener(cfg.Address)
		if err != nil {
			serverErrors <- err
			return
		}

		log.Info().
			Str("address", cfg.Address).
			Bool("tls", cfg.CertFile != "" && cfg.KeyFile != "").
			Msg("Starting server")

		if cfg.CertFile != "" && cfg.KeyFile != "" {
			serverErrors <- server.ServeTLS(listener, cfg.CertFile, cfg.KeyFile)
		} else {
			serverErrors <- server.Serve(listener)
		}
	}()

	// Setup interrupt handler
	interrupt := make(chan os.Signal, 1)
	notifySignal(interrupt, os.Interrupt)

	// Use a WaitGroup to ensure we don't exit before shutdown completes
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		
		select {
		case err := <-serverErrors:
			if err != nil && err != http.ErrServerClosed {
				log.Error().Err(err).Msg("Server error")
			}
		case <-interrupt:
			log.Info().Msg("Received interrupt signal, initiating graceful shutdown")
			
			ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownWait)
			defer cancel()
			
			if err := serverShutdown(server, ctx); err != nil {
				log.Error().Err(err).Msg("Graceful shutdown failed")
				serverErrors <- err
			}
		}
	}()

	wg.Wait()
	return <-serverErrors
}

// createListener creates either a TCP or Unix socket listener based on the address format
func createListener(address string) (net.Listener, error) {
	if strings.HasPrefix(address, "unix:") {
		socketPath := strings.TrimPrefix(address, "unix:")
		
		// Remove existing socket file if present
		if _, err := os.Stat(socketPath); err == nil {
			if err := os.Remove(socketPath); err != nil {
				return nil, err
			}
		}
		
		return net.Listen("unix", socketPath)
	}
	return net.Listen("tcp", address)
}
