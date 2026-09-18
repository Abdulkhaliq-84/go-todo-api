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
// PATCH semantics: only the fields the client actually
// sent are changed. Pointers encode that:
//
//	nil            -> absent, leave it alone
//	pointer to "x" -> set it to "x"
//
// Go's answer to Pydantic's Optional + exclude_unset, using pointers instead of
// a sentinel object.
//
// Due date needs a THIRD state -- "clear it" -- which a single pointer cannot
// express. An earlier draft used **time.Time, where the outer pointer meant
// "was it sent?" and the inner meant "is it null?". Correct, and unreadable.
//
// An explicit flag says the same thing in a form you can read at a glance:
//
//	DueDate=nil,  ClearDueDate=false  -> absent, unchanged
//	DueDate=&t,   ClearDueDate=false  -> set to t
//	DueDate=nil,  ClearDueDate=true   -> cleared
//
// The transport layer resolves the wire format's three states into these two
// fields; see internal/todo/http/mapping.go.
type UpdateTodoCommand struct {
	Title        *string
	Description  *string
	DueDate      *time.Time
	ClearDueDate bool
}

// ListTodosQuery is the input to Service.List.
type ListTodosQuery struct {
	Completed *bool
	Overdue   *bool
	Limit     int
	Offset    int
}
