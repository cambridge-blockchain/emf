package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// ShutdownPeriod is the duration to wait before a blocking function should abandon its work.
const ShutdownPeriod = 5 * time.Minute

// Option provides the client a callback that is used dynamically to specify attributes for a Server.
type Option func(*Server)

// Server is a robust server that can be easily be spun up and shut down with well implemented error handeling.
type Server struct {
	endpoint *http.Server
}

// WithServer creates an Option that is used for specifying the http.Server for a Server.
func WithServer(serv *http.Server) Option {
	return func(s *Server) { s.endpoint = serv }
}

// New is a variadic constructor for a Server.
func New(opts ...Option) *Server {
	var s = &Server{
		endpoint: &http.Server{},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Startup starts a robust server with well defined error handling.
func (s *Server) Startup() (quit chan bool) {
	var (
		err  error
		sigs chan os.Signal
	)

	sigs = make(chan os.Signal, 1)
	quit = make(chan bool, 1)

	signal.Notify(sigs, os.Interrupt)

	go func() {
		s.Shutdown(<-sigs, quit)
	}()

	if err = s.endpoint.ListenAndServe(); err != nil {
		panic(fmt.Errorf("failed to start server with error: '%s'", err))
	}
	return quit
}

// Shutdown stops the server gracefully.
func (s *Server) Shutdown(signal os.Signal, done chan bool) {
	var err error
	var ctx, cancel = context.WithTimeout(context.Background(), ShutdownPeriod)

	fmt.Printf("Shutting down server due to signal %v ....", signal)

	defer cancel()

	if err = s.endpoint.Shutdown(ctx); err != nil {
		panic(fmt.Errorf("failed to shutdown server with error: '%s'", err))
	}

	done <- true
}
