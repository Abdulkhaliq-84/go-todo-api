package app

import (
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// TodoDTO is the OUTPUT boundary of the application layer.
//
// Why not just return *domain.Todo and be done with it?
//   1. The domain entity's fields are private -- callers literally cannot read
//      them without going through getters.
//   2. Returning the entity would let the HTTP layer call todo.Complete(),
//      mutating the domain from inside a JSON handler. The DTO makes that
//      impossible: it is inert data.
//   3. It decouples your API's shape from your model's shape. Rename a domain
//      field and your JSON contract does not silently break.
//
// Still no json tags here -- those live in internal/todo/http/response.go.
// This layer does not know it is being serialised, or to what.
type TodoDTO struct {
	ID          string
	Title       string
	Description string
	Completed   bool
	DueDate     *time.Time
	Overdue     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// toDTO maps a domain entity to its output representation.
//
// Lowercase = private to package app. Nothing outside can call it, which keeps
// the mapping direction one-way by construction.
//
// TODO(you): read each field off the entity's getters. Decide what `now` to
// pass to IsOverdue -- time.Now() is easy but makes tests time-dependent;
// threading a clock through is testable but more plumbing.
func toDTO(t *domain.Todo) TodoDTO {
	return TodoDTO{} // TODO
}

// toDTOs maps a slice.
//
// TODO(you)
func toDTOs(todos []*domain.Todo) []TodoDTO {
	return nil // TODO
}
