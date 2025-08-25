package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	ErrNilHandler = errors.New("given handler is nil")
)

type Server struct {
	Server *http.Server
}

func NewServer(port int, handler http.Handler) (*Server, error) {
	if handler == nil {
		return nil, ErrNilHandler
	}
	return &Server{
		Server: &http.Server{
			ReadHeaderTimeout: time.Minute,
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           handler,
		},
	}, nil
}

// Run the server, blocking until ctx is cancelled or until Stop() is called.
//
// If ctx is cancelled, gracefully shut down the server with the given timeout.
// Gracefully shutting down means that requests which are in progress are still finished.
// So the timeout should be long enough to process a reasonable request, but
// shorter than the Kubernetes terminationGracePeriodSeconds (30 seconds by default)
// for force-killing the pod.
//
// Return an error if either running the server or gracefully shutting it down failed.
// No error is returned for a successful graceful shutdown.
func (s *Server) Run(ctx context.Context, shutdownTimeout time.Duration) error {
	// This is essentially the proposed implementation from https://github.com/golang/go/issues/52805

	// This channel must be buffered because we don't receive from it in the <-ctx.Done() case.
	// In that case, the goroutine below just exits and the error from ListenAndServe
	// is ignored (if would be ErrServerClosed anyway).
	errchan := make(chan error, 1)

	go func() {
		err := s.Server.ListenAndServe()
		errchan <- err
	}()

	select {
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return s.Stop(shutdownContext)
	case err := <-errchan:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("unable to serve http: %w", err)
		}
	}
	return nil
}

// Stop gracefully shuts down the http server.
//
// Note that most applications don't need this: the simplest way to
// shut down the server is using the context passed to Run().
func (s *Server) Stop(ctx context.Context) error {
	if err := s.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to gracefully shut down http server: %w", err)
	}
	return nil
}
