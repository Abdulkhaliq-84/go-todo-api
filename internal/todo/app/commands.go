package app

import "time"

// Commands and queries are the INPUT boundary of the application layer.
//
// They carry primitives (string, bool, *time.Time) rather than domain value
// objects, because the caller -- an HTTP handler, a CLI, a gRPC server, a test
// -- should not have to know how to build a domain.Title. Converting primitive
// to domain type is the service's job, and doing it there means validation
// runs no matter which transport called in.
//
// CQRS vocabulary, lightly applied: a Command changes state, a Query reads it.

// CreateTodoCommand is the input to Service.Create.
type CreateTodoCommand struct {
	Title       string
	Description string
	DueDate     *time.Time
}

// UpdateTodoCommand is the input to Service.Update.
//
// Pointer fields encode three-state intent, which a plain string cannot:
//   nil            -> field absent, leave it alone
//   pointer to ""  -> field present and explicitly cleared
//   pointer to "x" -> set it to "x"
//
// This is Go's answer to Pydantic's Optional + exclude_unset. Same problem,
// solved with pointers instead of sentinel objects.
type UpdateTodoCommand struct {
	Title       *string
	Description *string
	DueDate     **time.Time // pointer-to-pointer: outer = "was it sent?", inner = "is it null?"
}

// ListTodosQuery is the input to Service.List.
type ListTodosQuery struct {
	Completed *bool
	Overdue   *bool
	Limit     int
	Offset    int
}
