package domain

import "errors"

// Domain errors are part of the domain's vocabulary, so they are declared here
// -- not in the HTTP layer, and definitely not as HTTP status codes.
//
// The transport layer's job is to TRANSLATE these into 404 / 400 / 409.
// The domain must never know that HTTP, or status codes, exist.
//
// Go idiom: these are sentinel values compared with errors.Is(err, ErrNotFound),
// which walks wrapped errors. Closest analogue to catching a custom exception
// class -- except errors here are ordinary values you return, not control flow
// you throw.
var (
	ErrNotFound        = errors.New("todo not found")
	ErrTitleEmpty      = errors.New("title must not be empty")
	ErrTitleTooLong    = errors.New("title exceeds maximum length")
	ErrInvalidID       = errors.New("invalid todo id")
	ErrAlreadyComplete = errors.New("todo is already completed")
	ErrNotCompleted    = errors.New("todo is not completed")
)

// Deliberately absent: an ErrDueDateInPast.
//
// A due date in the past is allowed, on creation and on reschedule. Logging a
// task you already missed is a normal thing to do, and importing history would
// be impossible otherwise. A sentinel error the code never returns is worse
// than no sentinel at all -- it implies a rule that does not exist.
