package http

import (
	"errors"
	nethttp "net/http"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// This file is the ONLY place in the codebase that knows both domain errors and
// HTTP status codes. Concentrating the mapping here is what lets the domain
// stay transport-agnostic -- swap in gRPC tomorrow and you rewrite this one
// file and nothing else.
//
// errors.Is walks the wrapped chain, so a repository that wraps ErrNotFound
// with extra context still matches.

// classify turns any error into the status and payload the client should see.
//
// Codes are COARSE: invalid_title covers both the empty and the too-long case,
// and the `message` carries the specifics for a human. Fewer strings frozen
// into the contract forever, at the cost of a client not being able to tell the
// two title failures apart without reading prose.
//
// The `error` field is a STABLE MACHINE CODE clients may branch on, so treat it
// as part of the contract: renaming "not_found" is a breaking API change, the
// same as renaming a JSON field.
//
// TODO(you): switch on the domain errors. A starting point:
//
//	domain.ErrNotFound                          404  "not_found"
//	domain.ErrTitleEmpty, domain.ErrTitleTooLong 400  "invalid_title"
//	domain.ErrInvalidID                          400  "invalid_id"
//	domain.ErrAlreadyComplete                    409  "already_completed"
//	domain.ErrNotCompleted                       409  "not_completed"
//	anything else                                500  "internal_error"
//
// For the 500 case: LOG the real error server-side and return a generic
// message. Leaking `pq: relation "todos" does not exist` hands an attacker your
// schema, and handler_test.go has a test for exactly that.
func classify(err error) (int, ErrorResponse) {
	// TODO(you): remove these two lines -- they only keep the imports alive
	_ = errors.Is
	_ = domain.ErrNotFound

	return nethttp.StatusInternalServerError, ErrorResponse{
		Error:   "internal_error",
		Message: "an unexpected error occurred",
	}
}
