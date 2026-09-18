// Package server wraps net/http with sane timeouts and graceful shutdown.
package server

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/platform/config"
)

// Server owns the HTTP listener lifecycle.
type Server struct {
	httpServer *http.Server
}

// New builds the server.
//
// ALWAYS set ReadTimeout / WriteTimeout / IdleTimeout. Go's zero value for each
// is NO TIMEOUT AT ALL -- one slow client can hold a connection open forever,
// and enough of them exhaust the process. Node and uvicorn both ship defaults
// here; Go deliberately does not, and it bites people in production.
func New(cfg config.ServerConfig, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         net.JoinHostPort("", cfg.Port),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

// Addr reports the address the server listens on.
func (s *Server) Addr() string { return s.httpServer.Addr }

// Start begins listening and blocks until the server stops.
//
// ErrServerClosed is what ListenAndServe returns after a deliberate Shutdown,
// so it is filtered out here: a clean stop is not a failure, and reporting it
// as one would make every graceful shutdown look like a crash in the logs.
func (s *Server) Start() error {
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Shutdown stops accepting new connections and waits for in-flight requests to
// finish, up to the deadline on ctx.
//
// This is what makes a deploy not drop requests mid-flight: the container gets
// SIGTERM, the listener closes, work already in progress completes, the process
// exits. Without it, every deploy severs whatever was being served at that
// instant.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
