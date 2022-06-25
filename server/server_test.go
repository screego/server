package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestShutdownOnErrorWhileShutdown(t *testing.T) {
	disposeInterrupt := fakeInterrupt(t)
	defer disposeInterrupt()

	shutdownError := errors.New("shutdown error")
	disposeShutdown := fakeShutdownError(shutdownError)
	defer disposeShutdown()

	finished := make(chan error)

	go func() {
		finished <- Start(mux.NewRouter(), ":"+strconv.Itoa(port()), "", "")
	}()

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Server should be closed")
	case err := <-finished:
		assert.Equal(t, shutdownError, err)
	}
}

func TestShutdownAfterError(t *testing.T) {
	finished := make(chan error)

	go func() {
		finished <- Start(mux.NewRouter(), ":-5", "", "")
	}()

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Server should be closed")
	case err := <-finished:
		assert.NotNil(t, err)
	}
}

func TestShutdown(t *testing.T) {
	dispose := fakeInterrupt(t)
	defer dispose()

	finished := make(chan error)

	go func() {
		finished <- Start(mux.NewRouter(), ":"+strconv.Itoa(port()), "", "")
	}()

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Server should be closed")
	case err := <-finished:
		assert.Nil(t, err)
	}
}

func fakeInterrupt(t *testing.T) func() {
	oldNotify := notifySignal
	notifySignal = func(c chan<- os.Signal, sig ...os.Signal) {
		assert.Contains(t, sig, os.Interrupt)
		go func() {
			time.Sleep(100 * time.Millisecond)
			c <- os.Interrupt
		}()
	}
	return func() {
		notifySignal = oldNotify
	}
}

func fakeShutdownError(err error) func() {
	old := serverShutdown
	serverShutdown = func(server *http.Server, ctx context.Context) error {
		return err
	}
	return func() {
		serverShutdown = old
	}
}

func port() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
