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
// Codes are COARSE: invalid_title covers both the empty and the too-long case,
// and the `message` carries the specifics for a human. Fewer strings frozen
// into the contract forever, at the cost of a client not being able to tell the
// two title failures apart without reading prose.
//
// The `error` field is a STABLE MACHINE CODE clients branch on, so treat it as
// part of the contract: renaming "not_found" is a breaking API change, exactly
// like renaming a JSON field.

// Machine-readable error codes. Constants rather than inline strings so a typo
// is a compile error and every use site is greppable.
const (
	codeNotFound         = "not_found"
	codeInvalidTitle     = "invalid_title"
	codeInvalidID        = "invalid_id"
	codeAlreadyCompleted = "already_completed"
	codeNotCompleted     = "not_completed"
	codeInternal         = "internal_error"
)

// classify turns any error into the status and payload the client should see.
//
// errors.Is walks the wrapped chain, so a repository that wrapped the sentinel
// with extra context still matches here.
func classify(err error) (int, ErrorResponse) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nethttp.StatusNotFound, ErrorResponse{
			Error:   codeNotFound,
			Message: "todo not found",
		}

	case errors.Is(err, domain.ErrTitleEmpty):
		return nethttp.StatusBadRequest, ErrorResponse{
			Error:   codeInvalidTitle,
			Message: "title must not be empty",
		}

	case errors.Is(err, domain.ErrTitleTooLong):
		return nethttp.StatusBadRequest, ErrorResponse{
			Error: codeInvalidTitle,
			// Built from the domain's own constant, so the limit cannot drift
			// out of sync with the rule it describes.
			Message: "title must be at most " + itoa(domain.TitleMaxLength) + " characters",
		}

	case errors.Is(err, domain.ErrInvalidID):
		return nethttp.StatusBadRequest, ErrorResponse{
			Error:   codeInvalidID,
			Message: "id must be a valid UUID",
		}

	case errors.Is(err, domain.ErrAlreadyComplete):
		return nethttp.StatusConflict, ErrorResponse{
			Error:   codeAlreadyCompleted,
			Message: "todo is already completed",
		}

	case errors.Is(err, domain.ErrNotCompleted):
		return nethttp.StatusConflict, ErrorResponse{
			Error:   codeNotCompleted,
			Message: "todo is not completed",
		}

	default:
		// Anything unrecognised is a bug or an outage, and the client is told
		// nothing about it. The real error is logged server-side by the
		// response error handler in router.go.
		//
		// Leaking `pq: relation "todos" does not exist` would hand an attacker
		// the schema; handler_test.go asserts this body stays generic.
		return nethttp.StatusInternalServerError, ErrorResponse{
			Error:   codeInternal,
			Message: "an unexpected error occurred",
		}
	}
}

// isInternal reports whether an error has no meaningful HTTP representation, so
// a handler should hand it back to the strict runtime (which logs it and
// answers 500) rather than turning it into a typed response.
func isInternal(err error) bool {
	status, _ := classify(err)
	return status == nethttp.StatusInternalServerError
}

// itoa avoids pulling strconv in for one call site.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
