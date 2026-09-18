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
// DueDate is nullable.Nullable[time.Time], not *time.Time, because the spec
// declares it optional AND nullable. Build it with
// nullable.NewNullableWithValue(t) or nullable.NewNullNullable().
//
// TODO(you): parse dto.ID into a uuid.UUID (openapi_types.UUID is an alias for
// it) and copy the rest across.
func toTodo(dto app.TodoDTO) Todo {
	return Todo{} // TODO
}

// toUpdateCommand resolves the wire format's THREE states into the application
// layer's two fields. This function is the entire reason nullable-type is
// enabled -- it is where "the client did not mention due_date" and "the client
// asked to clear due_date" stop being the same thing.
//
//	body.DueDate.IsSpecified() == false  -> absent        -> ClearDueDate=false, DueDate=nil
//	body.DueDate.IsNull()      == true   -> explicit null -> ClearDueDate=true,  DueDate=nil
//	otherwise                            -> a value       -> ClearDueDate=false, DueDate=&v
//
// TODO(you): implement the three-way resolution.
func toUpdateCommand(body UpdateTodoRequest) app.UpdateTodoCommand {
	return app.UpdateTodoCommand{} // TODO
}

// toTodos maps a slice.
//
// TODO(you)
func toTodos(dtos []app.TodoDTO) []Todo {
	return nil // TODO
}
