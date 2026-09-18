// Package server wraps net/http with sane timeouts and graceful shutdown.
package server

import (
	"context"
	"net/http"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/platform/config"
)

// Server owns the HTTP listener lifecycle.
type Server struct {
	httpServer *http.Server
}

// New builds the server.
//
// IMPORTANT: always set ReadTimeout/WriteTimeout/IdleTimeout. Go's
// http.ListenAndServe default is NO timeouts at all -- one slow client can hold
// a connection open forever. Node and uvicorn both ship defaults here; Go
// deliberately does not, and it bites people in production.
//
// TODO(you): construct the http.Server from cfg and handler.
func New(cfg config.ServerConfig, handler http.Handler) *Server {
	return nil // TODO
}

// Start begins listening. Blocks until the server stops.
//
// TODO(you): return ListenAndServe's error, but treat http.ErrServerClosed as
// a normal shutdown rather than a failure.
func (s *Server) Start() error {
	return nil // TODO
}

// Shutdown stops accepting new connections and waits for in-flight requests.
//
// This is what makes a deploy not drop requests mid-flight -- the container
// gets SIGTERM, you stop accepting, you finish what you started, you exit.
//
// TODO(you)
func (s *Server) Shutdown(ctx context.Context) error {
	return nil // TODO
}
