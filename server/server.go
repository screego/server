package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/gorilla/mux"
)

var notifySignal = signal.Notify
var serverShutdown = func(server *http.Server, ctx context.Context) error {
	return server.Shutdown(ctx)
}

// Start starts the http server
func Start(mux *mux.Router, address, cert, key string) error {
	server, shutdown := startServer(mux, address, cert, key)
	shutdownOnInterruptSignal(server, 2*time.Second, shutdown)
	return waitForServerToClose(shutdown)
}

func startServer(mux *mux.Router, address, cert, key string) (*http.Server, chan error) {
	srv := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	shutdown := make(chan error)
	go func() {
		if cert != "" || key != "" {
			log.Info().Str("addr", address).Msg("Start HTTP with tls")
			err := srv.ListenAndServeTLS(cert, key)
			shutdown <- err
		} else {
			log.Info().Str("addr", address).Msg("Start HTTP")
			err := srv.ListenAndServe()
			shutdown <- err
		}
	}()
	return srv, shutdown
}

func shutdownOnInterruptSignal(server *http.Server, timeout time.Duration, shutdown chan<- error) {
	interrupt := make(chan os.Signal, 1)
	notifySignal(interrupt, os.Interrupt)

	go func() {
		<-interrupt
		log.Info().Msg("Received interrupt. Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := serverShutdown(server, ctx); err != nil {
			shutdown <- err
		}
	}()
}

func waitForServerToClose(shutdown <-chan error) error {
	err := <-shutdown
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
