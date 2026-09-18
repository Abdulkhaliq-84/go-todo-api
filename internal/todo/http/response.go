package http

import (
	nethttp "net/http"
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/app"
)

// Response DTOs: the public shape of your API, versioned independently of the
// domain model. Renaming a domain field should never break a client.

type todoResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	DueDate     *time.Time `json:"due_date"`
	Overdue     bool       `json:"overdue"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// listResponse wraps collections in an object rather than returning a bare
// JSON array. Bare top-level arrays leave you nowhere to add pagination
// metadata later without breaking every client.
type listResponse struct {
	Data  []todoResponse `json:"data"`
	Count int            `json:"count"`
}

// errorResponse is the single error shape every failure uses. One shape means
// clients write one error handler.
type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// TODO(you): map app.TodoDTO -> todoResponse.
func toResponse(dto app.TodoDTO) todoResponse {
	return todoResponse{} // TODO
}

// TODO(you)
func toResponses(dtos []app.TodoDTO) []todoResponse {
	return nil // TODO
}

// writeJSON serialises a payload with the right headers and status.
//
// TODO(you): set Content-Type BEFORE WriteHeader -- headers written after the
// status line are silently discarded, which is a classic Go gotcha.
func writeJSON(w nethttp.ResponseWriter, status int, payload any) {
	// TODO
}
