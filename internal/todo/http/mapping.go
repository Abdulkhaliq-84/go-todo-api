package http

import (
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/app"
	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
)

// Mapping between the application layer's DTOs and the GENERATED wire types.
//
// There are no hand-written request/response structs: Todo, CreateTodoRequest,
// TodoListResponse and ErrorResponse all come from openapi_gen.go. Writing them
// by hand would mean two definitions of one contract, free to drift.
//
// The layering still holds -- three representations, each owned by one layer,
// each free to change without breaking the others:
//
//	domain.Todo   business object, private fields, behaviour
//	app.TodoDTO   inert output of a use case
//	http.Todo     the wire format, generated from the spec

// ---------------------------------------------------------------------------
// outbound
// ---------------------------------------------------------------------------

// toTodo converts an application DTO into the generated wire type.
//
// The id is parsed rather than assigned: openapi_types.UUID is an alias for
// uuid.UUID, not a string. A parse failure here would mean the application
// layer produced a malformed id, which cannot happen -- domain.ID is only ever
// built through NewID or ParseID. The zero UUID on failure is a deliberate
// tombstone: visibly wrong in a response rather than a panic in production.
func toTodo(dto app.TodoDTO) Todo {
	id, err := uuid.Parse(dto.ID)
	if err != nil {
		id = uuid.Nil
	}

	return Todo{
		Id:          id,
		Title:       dto.Title,
		Description: dto.Description,
		Completed:   dto.Completed,
		DueDate:     toNullableTime(dto.DueDate),
		Overdue:     dto.Overdue,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}
}

func toTodos(dtos []app.TodoDTO) []Todo {
	// Non-nil, so an empty page serialises as [] rather than null.
	todos := make([]Todo, 0, len(dtos))
	for _, dto := range dtos {
		todos = append(todos, toTodo(dto))
	}
	return todos
}

// toNullableTime turns an optional time into the wire's three-state value.
// A missing due date is emitted as JSON null rather than omitted, so the field
// is always present and clients need not distinguish absent from cleared.
func toNullableTime(t *time.Time) nullable.Nullable[time.Time] {
	if t == nil {
		return nullable.NewNullNullable[time.Time]()
	}
	return nullable.NewNullableWithValue(*t)
}

// ---------------------------------------------------------------------------
// inbound
// ---------------------------------------------------------------------------

// toCreateCommand maps the request body onto the use case's input.
func toCreateCommand(body CreateTodoRequest) app.CreateTodoCommand {
	cmd := app.CreateTodoCommand{Title: body.Title}

	if body.Description != nil {
		cmd.Description = *body.Description
	}
	if v, ok := fromNullableTime(body.DueDate); ok {
		cmd.DueDate = v
	}
	return cmd
}

// toUpdateCommand resolves the wire format's THREE states into the application
// layer's two fields.
//
// This function is the entire reason nullable-type is enabled in the codegen
// config. It is where "the client did not mention due_date" and "the client
// asked to clear due_date" stop being the same thing -- with a plain
// *time.Time, both arrive as nil and a PATCH that omitted the field would
// silently wipe the stored date.
//
//	not specified -> absent        -> DueDate=nil, ClearDueDate=false
//	specified null-> explicit null -> DueDate=nil, ClearDueDate=true
//	specified value               -> DueDate=&v,  ClearDueDate=false
func toUpdateCommand(body UpdateTodoRequest) app.UpdateTodoCommand {
	cmd := app.UpdateTodoCommand{
		Title:       body.Title,
		Description: body.Description,
	}

	switch {
	case !body.DueDate.IsSpecified():
		// absent: leave the stored value alone
	case body.DueDate.IsNull():
		cmd.ClearDueDate = true
	default:
		if v, err := body.DueDate.Get(); err == nil {
			cmd.DueDate = &v
		}
	}
	return cmd
}

// toListQuery maps parsed query parameters onto the use case's input.
//
// Nothing is defaulted here. Pagination bounds live in the application layer so
// every caller inherits them -- an HTTP handler, a CLI, a future gRPC server.
// Defaulting in both places is how the two quietly disagree.
func toListQuery(params ListTodosParams) app.ListTodosQuery {
	q := app.ListTodosQuery{
		Completed: params.Completed,
		Overdue:   params.Overdue,
	}
	if params.Limit != nil {
		q.Limit = *params.Limit
	}
	if params.Offset != nil {
		q.Offset = *params.Offset
	}
	return q
}

// fromNullableTime reads the wire's three-state value. ok is false when the
// field was absent or explicitly null.
func fromNullableTime(n nullable.Nullable[time.Time]) (*time.Time, bool) {
	if !n.IsSpecified() || n.IsNull() {
		return nil, false
	}
	v, err := n.Get()
	if err != nil {
		return nil, false
	}
	return &v, true
}
