package http

import (
	"errors"
	nethttp "net/http"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// This file is the ONLY place in the codebase that knows both domain errors
// and HTTP status codes. Concentrating the mapping here is what lets the domain
// stay transport-agnostic -- swap in a gRPC layer tomorrow and you write a new
// version of this one file, and change nothing else.
//
// errors.Is walks the wrapped-error chain, so a repository that wraps
// ErrNotFound with extra context still matches here.

// writeError translates any error into an HTTP response.
//
// >>> THIS IS ONE OF YOUR DECISIONS -- see the note I left you. <<<
//
// TODO(you): switch on the domain error and pick statuses. A starting point:
//
//   domain.ErrNotFound                        -> 404
//   domain.ErrTitleEmpty, ErrTitleTooLong     -> 400
//   domain.ErrInvalidID                       -> 400
//   domain.ErrAlreadyComplete, ErrNotCompleted-> 409
//   anything unrecognised                     -> 500, and log the real error
//                                                while returning something
//                                                generic to the client
func writeError(w nethttp.ResponseWriter, err error) {
	// TODO(you): remove both lines -- they only keep the imports alive
	_ = errors.Is
	_ = domain.ErrNotFound
	// TODO
}
