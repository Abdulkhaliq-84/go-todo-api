package http

import (
	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/app"
)

// Mapping between the application layer's DTOs and the GENERATED wire types.
//
// Note there are no hand-written request/response structs any more -- Todo,
// CreateTodoRequest, TodoListResponse and ErrorResponse all come from
// openapi_gen.go. Writing them by hand would mean two definitions of the same
// contract, free to drift.
//
// The layering still holds. Three distinct representations, each owned by one
// layer, each free to change without breaking the others:
//
//	domain.Todo   business object, private fields, behaviour
//	app.TodoDTO   inert output of a use case
//	http.Todo     the wire format, generated from the spec

// toTodo converts an application DTO into the generated wire type.
//
// TODO(you): parse dto.ID into a uuid.UUID (openapi_types.UUID is an alias for
// it) and copy the rest across.
func toTodo(dto app.TodoDTO) Todo {
	return Todo{} // TODO
}

// toTodos maps a slice.
//
// TODO(you)
func toTodos(dtos []app.TodoDTO) []Todo {
	return nil // TODO
}
