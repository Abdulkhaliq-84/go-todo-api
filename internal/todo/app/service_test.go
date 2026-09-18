package app_test

// Tests for the application layer: is the ORCHESTRATION right?
//
// What belongs here:
//   - the service calls the domain and persists the result
//   - domain errors propagate up untouched (not swallowed, not re-wrapped
//     into something the HTTP layer cannot recognise)
//   - repository failures surface rather than being ignored
//
// What does NOT belong here: business rules. "Cannot complete a completed todo"
// is tested in domain/todo_test.go. Testing it again here means you have
// duplicated the rule, or worse, implemented it in the wrong layer.

import (
	"context"
	"testing"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/app"
)

func TestService_Create(t *testing.T) {
	t.Skip("TODO(you): remove once Service.Create is implemented")

	_ = context.Background
	_ = app.CreateTodoCommand{}

	// TODO(you): given a fake repo, when Create is called with a valid command,
	// assert the returned DTO carries the title and that the todo was persisted.
}

func TestService_Create_InvalidTitle(t *testing.T) {
	t.Skip("TODO(you)")

	// TODO(you): an empty title must fail with domain.ErrTitleEmpty AND must
	// not persist anything. That second assertion is the one people forget --
	// it is what proves validation happens before the write, not after.
}

func TestService_Complete_NotFound(t *testing.T) {
	t.Skip("TODO(you)")

	// TODO(you): completing an unknown id must return domain.ErrNotFound
	// unchanged, so the HTTP layer can map it to a 404. If the service wraps it
	// in its own error type, errors.Is stops matching and you silently start
	// returning 500s.
}

// TODO(you): TestService_List_FiltersPassThrough -- assert that a
// ListTodosQuery becomes the equivalent domain.Filter. This is pure translation
// and exactly the kind of thing that breaks silently.
