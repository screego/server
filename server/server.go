package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

var (
	notifySignal   = signal.Notify
	serverShutdown = func(server *http.Server, ctx context.Context) error {
		return server.Shutdown(ctx)
	}
)

// Start starts the http server.
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
		err := listenAndServe(srv, address, cert, key)
		shutdown <- err
	}()
	return srv, shutdown
}

func listenAndServe(srv *http.Server, address, cert, key string) error {
	var err error
	var listener net.Listener

	if strings.HasPrefix(address, "unix:") {
		listener, err = net.Listen("unix", strings.TrimPrefix(address, "unix:"))
	} else {
		listener, err = net.Listen("tcp", address)
	}
	if err != nil {
		return err
	}

	if cert != "" || key != "" {
		log.Info().Str("addr", address).Msg("Start HTTP with tls")
		return srv.ServeTLS(listener, cert, key)
	} else {
		log.Info().Str("addr", address).Msg("Start HTTP")
		return srv.Serve(listener)
	}
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
