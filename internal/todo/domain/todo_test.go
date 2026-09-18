package domain_test

// Package name is domain_test, not domain. Go allows either; the _test suffix
// means this file may only touch the EXPORTED API -- exactly what a real caller
// sees. If a test needs private fields to be useful, that is a signal the
// public API is wrong, not a reason to drop the suffix.
//
// These tests need NO database, NO HTTP server, and NO mocks. That is not an
// accident -- it is the payoff for keeping domain/ free of dependencies. If you
// ever find yourself needing a mock to test the domain, something leaked in.

import (
	"errors"
	"testing"
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// newTestTodo is a helper that builds a valid Todo for tests that are about
// something else. t.Helper() makes failure line numbers point at the CALLER,
// not at this function -- small thing, saves real debugging time.
//
// TODO(you)
func newTestTodo(t *testing.T, title string) *domain.Todo {
	t.Helper()
	return nil // TODO
}

func TestNew(t *testing.T) {
	t.Skip("TODO(you): remove once domain.New is implemented")

	// TODO(you): assert that a new Todo
	//   - has a non-zero ID
	//   - is not completed
	//   - has createdAt == updatedAt
}

// TestTodo_Complete is where YOUR design decision gets pinned down.
//
// Write the cases for the behaviour you chose in docs/DECISIONS.md #1, then
// make them pass. Whichever you picked, the test is the thing that stops the
// rule from quietly changing later.
func TestTodo_Complete(t *testing.T) {
	t.Skip("TODO(you): remove once Complete() is implemented")

	tests := []struct {
		name      string
		completed bool  // starting state
		wantErr   error // nil, or domain.ErrAlreadyComplete -- depends on your choice
	}{
		{"completing an active todo", false, nil},
		{"completing an already-completed todo", true, domain.ErrAlreadyComplete}, // <- change if you chose idempotent
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt
			_ = errors.Is
			// TODO(you): build a todo in the starting state, call Complete(),
			// assert on the error and on IsCompleted().
		})
	}
}

// TODO(you): TestTodo_Reopen -- must mirror Complete's rule exactly.

func TestTodo_IsOverdue(t *testing.T) {
	t.Skip("TODO(you): remove once IsOverdue is implemented")

	// Note that IsOverdue takes `now` as a parameter rather than calling
	// time.Now() internally. That is what makes this testable without sleeping
	// or freezing the clock -- you just pass the moment you want to ask about.
	_ = time.Now

	// TODO(you): cover
	//   - no due date        -> never overdue
	//   - due in the future  -> not overdue
	//   - due in the past    -> overdue
	//   - due in the past but already completed -> ??? (your call: is a
	//     finished task still "overdue"? Decide, then encode it here.)
}
