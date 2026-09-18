package app

import (
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// TodoDTO is the OUTPUT boundary of the application layer.
//
// Why not just return *domain.Todo?
//  1. The entity's fields are private -- callers cannot read them except
//     through getters.
//  2. Returning the entity would let the HTTP layer call todo.Complete() and
//     mutate the domain from inside a JSON handler. A DTO is inert data.
//  3. It decouples the API's shape from the model's. Rename a domain field and
//     the JSON contract does not silently break.
//
// No json tags here -- those live on the generated wire types in
// internal/todo/http. This layer does not know it is being serialised, or to
// what.
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
// Takes `now` explicitly rather than calling time.Now() in here, so the domain
// and this mapping stay pure functions of their inputs. The service calls
// time.Now() ONCE per request and threads it down -- which also means every
// todo in a list is judged against the same instant rather than each against a
// slightly different one.
//
// Lowercase = private to package app, so mapping can only flow outward.
func toDTO(t *domain.Todo, now time.Time) TodoDTO {
	return TodoDTO{
		ID:          t.ID().String(),
		Title:       t.Title().String(),
		Description: t.Description(),
		Completed:   t.IsCompleted(),
		DueDate:     t.DueDate(), // already a defensive copy
		Overdue:     t.IsOverdue(now),
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   t.UpdatedAt(),
	}
}

// toDTOs maps a slice against a single instant.
func toDTOs(todos []*domain.Todo, now time.Time) []TodoDTO {
	// Return an empty slice, never nil: the transport layer serialises this
	// straight into a JSON array, and a nil slice marshals to `null` rather
	// than `[]`. Clients doing data.map(...) break on null.
	dtos := make([]TodoDTO, 0, len(todos))
	for _, t := range todos {
		dtos = append(dtos, toDTO(t, now))
	}
	return dtos
}
