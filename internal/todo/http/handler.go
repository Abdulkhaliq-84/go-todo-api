// Package http adapts HTTP to the application layer.
//
// Its entire job is translation, in both directions:
//   inbound:  JSON body + URL params  ->  app commands/queries
//   outbound: app DTOs or errors      ->  JSON + status codes
//
// It imports app. It must NOT import domain directly for business decisions,
// and it must NEVER import postgres.
//
// Package name collides with net/http, so the import below is aliased. Slightly
// awkward, and entirely normal in Go codebases organised this way.
package http

import (
	nethttp "net/http"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/app"
)

// Handler holds the dependencies every route needs.
//
// In Express you would close over `service` in a route callback; in FastAPI you
// would Depends() it in. In Go you put it in a struct and hang methods off that
// struct -- the methods then satisfy http.HandlerFunc. Same idea, no framework.
type Handler struct {
	service *app.Service
}

func NewHandler(service *app.Service) *Handler {
	return &Handler{service: service}
}

// Every handler has the same signature: func(ResponseWriter, *Request).
// No return value -- you WRITE the response rather than returning it. That is
// the biggest surface-level difference from Express/FastAPI, and it means
// forgetting to return after an error write is a real bug Go will not catch.

// Create handles POST /api/v1/todos
//
// TODO(you): decode the body, validate, call service.Create, write 201.
func (h *Handler) Create(w nethttp.ResponseWriter, r *nethttp.Request) {
	// TODO
}

// List handles GET /api/v1/todos
//
// TODO(you): parse ?completed= &overdue= &limit= &offset= into a ListTodosQuery.
func (h *Handler) List(w nethttp.ResponseWriter, r *nethttp.Request) {
	// TODO
}

// GetByID handles GET /api/v1/todos/{id}
//
// Go 1.22+ gives you r.PathValue("id") from the stdlib router -- no gorilla/mux,
// no chi needed for this.
//
// TODO(you)
func (h *Handler) GetByID(w nethttp.ResponseWriter, r *nethttp.Request) {
	// TODO
}

// Update handles PATCH /api/v1/todos/{id}
//
// TODO(you)
func (h *Handler) Update(w nethttp.ResponseWriter, r *nethttp.Request) {
	// TODO
}

// Complete handles POST /api/v1/todos/{id}/complete
//
// Why a sub-resource verb instead of PATCH {"completed": true}? Because
// completing is a domain ACTION with its own rules, not a field assignment.
// The URL reflects the model. Change it if you disagree -- but decide on purpose.
//
// TODO(you)
func (h *Handler) Complete(w nethttp.ResponseWriter, r *nethttp.Request) {
	// TODO
}

// Reopen handles POST /api/v1/todos/{id}/reopen
//
// TODO(you)
func (h *Handler) Reopen(w nethttp.ResponseWriter, r *nethttp.Request) {
	// TODO
}

// Delete handles DELETE /api/v1/todos/{id}
//
// TODO(you): 204 No Content on success.
func (h *Handler) Delete(w nethttp.ResponseWriter, r *nethttp.Request) {
	// TODO
}
