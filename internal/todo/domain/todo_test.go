package domain_test

// Package name is domain_test, not domain. Go allows either; the _test suffix
// means this file may only touch the EXPORTED API -- exactly what a real caller
// sees. If a test needs private fields to be useful, that is a signal the public
// API is wrong, not a reason to drop the suffix.
//
// These tests need no database, no HTTP server and no mocks. That is not an
// accident -- it is the payoff for keeping domain/ free of dependencies.

import (
	"errors"
	"testing"
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// Fixed instants. Every test states the moment it means, so nothing here
// depends on when the suite runs.
var (
	tCreate = time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	tLater  = time.Date(2026, 3, 2, 9, 0, 0, 0, time.UTC)
	tPast   = time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC)
	tFuture = time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)
)

func newTestTodo(t *testing.T, title string) *domain.Todo {
	t.Helper()
	todo, err := domain.New(mustTitle(t, title), "", nil, tCreate)
	if err != nil {
		t.Fatalf("New: unexpected error %v", err)
	}
	return todo
}

func TestNew(t *testing.T) {
	due := tFuture
	todo, err := domain.New(mustTitle(t, "Learn Go"), "interfaces", &due, tCreate)
	if err != nil {
		t.Fatalf("New: unexpected error %v", err)
	}

	if todo.ID().IsZero() {
		t.Error("a new todo must have an identity before it is persisted")
	}
	if todo.IsCompleted() {
		t.Error("a new todo must start active")
	}
	if !todo.CreatedAt().Equal(tCreate) || !todo.UpdatedAt().Equal(tCreate) {
		t.Errorf("timestamps = %v / %v, want both %v", todo.CreatedAt(), todo.UpdatedAt(), tCreate)
	}
	if todo.Title().String() != "Learn Go" || todo.Description() != "interfaces" {
		t.Errorf("content not stored: %q / %q", todo.Title(), todo.Description())
	}
}

func TestNew_UniqueIdentities(t *testing.T) {
	a := newTestTodo(t, "one")
	b := newTestTodo(t, "two")
	if a.ID().Equals(b.ID()) {
		t.Error("each todo must get its own identity")
	}
}

// A due date in the past is allowed: logging a task you already missed is
// normal, and importing history would be impossible otherwise.
func TestNew_AcceptsPastDueDate(t *testing.T) {
	due := tPast
	if _, err := domain.New(mustTitle(t, "Overdue import"), "", &due, tCreate); err != nil {
		t.Errorf("a past due date must be accepted, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// State machine -- strict: a transition from the wrong state is an error
// ---------------------------------------------------------------------------

func TestTodo_Complete(t *testing.T) {
	tests := []struct {
		name          string
		startCompleted bool
		wantErr       error
		wantCompleted bool
	}{
		{"completing an active todo", false, nil, true},
		{"completing an already-completed todo", true, domain.ErrAlreadyComplete, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo := newTestTodo(t, "Buy milk")
			if tt.startCompleted {
				if err := todo.Complete(tCreate); err != nil {
					t.Fatalf("setup: %v", err)
				}
			}
			before := todo.UpdatedAt()

			err := todo.Complete(tLater)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Complete() error = %v, want %v", err, tt.wantErr)
			}
			if todo.IsCompleted() != tt.wantCompleted {
				t.Errorf("IsCompleted() = %v, want %v", todo.IsCompleted(), tt.wantCompleted)
			}

			// The half people forget: a FAILED transition must change nothing
			// at all. A rejected call that still bumped updatedAt would be a
			// partial mutation, which is exactly what an aggregate prevents.
			if tt.wantErr != nil && !todo.UpdatedAt().Equal(before) {
				t.Errorf("failed Complete moved updatedAt: %v -> %v", before, todo.UpdatedAt())
			}
			if tt.wantErr == nil && !todo.UpdatedAt().Equal(tLater) {
				t.Errorf("updatedAt = %v, want %v", todo.UpdatedAt(), tLater)
			}
		})
	}
}

func TestTodo_Reopen(t *testing.T) {
	tests := []struct {
		name           string
		startCompleted bool
		wantErr        error
		wantCompleted  bool
	}{
		{"reopening a completed todo", true, nil, false},
		{"reopening an active todo", false, domain.ErrNotCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo := newTestTodo(t, "Buy milk")
			if tt.startCompleted {
				if err := todo.Complete(tCreate); err != nil {
					t.Fatalf("setup: %v", err)
				}
			}
			before := todo.UpdatedAt()

			err := todo.Reopen(tLater)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Reopen() error = %v, want %v", err, tt.wantErr)
			}
			if todo.IsCompleted() != tt.wantCompleted {
				t.Errorf("IsCompleted() = %v, want %v", todo.IsCompleted(), tt.wantCompleted)
			}
			if tt.wantErr != nil && !todo.UpdatedAt().Equal(before) {
				t.Errorf("failed Reopen moved updatedAt: %v -> %v", before, todo.UpdatedAt())
			}
		})
	}
}

// Reopening must not clear the due date -- a todo reopened past its deadline is
// immediately overdue again, which is truthful.
func TestTodo_Reopen_KeepsDueDate(t *testing.T) {
	due := tPast
	todo, err := domain.New(mustTitle(t, "Late task"), "", &due, tCreate)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := todo.Complete(tCreate); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if err := todo.Reopen(tLater); err != nil {
		t.Fatalf("Reopen: %v", err)
	}

	if todo.DueDate() == nil || !todo.DueDate().Equal(tPast) {
		t.Errorf("DueDate() = %v, want it preserved as %v", todo.DueDate(), tPast)
	}
	if !todo.IsOverdue(tLater) {
		t.Error("a reopened todo past its due date should be overdue again")
	}
}

// ---------------------------------------------------------------------------
// Derived state -- a completed todo is never overdue
// ---------------------------------------------------------------------------

func TestTodo_IsOverdue(t *testing.T) {
	tests := []struct {
		name      string
		dueDate   *time.Time
		completed bool
		now       time.Time
		want      bool
	}{
		{"no due date", nil, false, tLater, false},
		{"due in the future", &tFuture, false, tLater, false},
		{"due in the past", &tPast, false, tLater, true},
		// Decision #7: "overdue" means needs attention, and a finished task
		// needs none -- even one finished after its deadline.
		{"due in the past but completed", &tPast, true, tLater, false},
		{"due in the future and completed", &tFuture, true, tLater, false},
		// Exactly at the deadline is not yet late -- After is strict.
		{"exactly at the due date", &tLater, false, tLater, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo, err := domain.New(mustTitle(t, "Task"), "", tt.dueDate, tCreate)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if tt.completed {
				if err := todo.Complete(tCreate); err != nil {
					t.Fatalf("Complete: %v", err)
				}
			}
			if got := todo.IsOverdue(tt.now); got != tt.want {
				t.Errorf("IsOverdue(%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Mutators
// ---------------------------------------------------------------------------

func TestTodo_UpdateTitle(t *testing.T) {
	todo := newTestTodo(t, "Old")
	if err := todo.UpdateTitle(mustTitle(t, "New"), tLater); err != nil {
		t.Fatalf("UpdateTitle: %v", err)
	}
	if todo.Title().String() != "New" {
		t.Errorf("Title() = %q, want %q", todo.Title(), "New")
	}
	if !todo.UpdatedAt().Equal(tLater) {
		t.Errorf("updatedAt = %v, want %v", todo.UpdatedAt(), tLater)
	}
}

// Writing the same value is not a change, so it must not bump updatedAt --
// otherwise a no-op PATCH looks like an edit to every client watching the field.
func TestTodo_UpdateTitle_NoChangeDoesNotTouch(t *testing.T) {
	todo := newTestTodo(t, "Same")
	if err := todo.UpdateTitle(mustTitle(t, "Same"), tLater); err != nil {
		t.Fatalf("UpdateTitle: %v", err)
	}
	if !todo.UpdatedAt().Equal(tCreate) {
		t.Errorf("updatedAt moved on a no-op update: %v", todo.UpdatedAt())
	}
}

func TestTodo_Reschedule(t *testing.T) {
	todo := newTestTodo(t, "Task")

	if err := todo.Reschedule(&tFuture, tLater); err != nil {
		t.Fatalf("Reschedule: %v", err)
	}
	if todo.DueDate() == nil || !todo.DueDate().Equal(tFuture) {
		t.Errorf("DueDate() = %v, want %v", todo.DueDate(), tFuture)
	}

	if err := todo.Reschedule(nil, tLater); err != nil {
		t.Fatalf("Reschedule(nil): %v", err)
	}
	if todo.DueDate() != nil {
		t.Errorf("DueDate() = %v, want nil after clearing", todo.DueDate())
	}
}

// The aggregate must hand out a COPY of its due date. Returning the stored
// pointer would let any caller mutate the entity from outside, defeating every
// private field in the struct.
func TestTodo_DueDate_IsDefensiveCopy(t *testing.T) {
	due := tFuture
	todo, err := domain.New(mustTitle(t, "Task"), "", &due, tCreate)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Mutate through the caller's own pointer and through the returned one.
	due = tPast
	if got := todo.DueDate(); got != nil {
		*got = tPast
	}

	if todo.DueDate() == nil || !todo.DueDate().Equal(tFuture) {
		t.Errorf("DueDate() = %v, want the original %v -- the aggregate leaked a pointer", todo.DueDate(), tFuture)
	}
}

func TestReconstitute(t *testing.T) {
	id := domain.NewID()
	due := tFuture

	todo := domain.Reconstitute(id, mustTitle(t, "Stored"), "desc", true, &due, tPast, tCreate)

	// Reconstitute must take every value verbatim: it is rebuilding a row that
	// already passed the creation rules, so it must not mint a new ID or reset
	// createdAt the way New does.
	if !todo.ID().Equals(id) {
		t.Errorf("ID() = %v, want the stored %v", todo.ID(), id)
	}
	if !todo.IsCompleted() {
		t.Error("stored completion state was lost")
	}
	if !todo.CreatedAt().Equal(tPast) || !todo.UpdatedAt().Equal(tCreate) {
		t.Errorf("timestamps = %v / %v, want %v / %v", todo.CreatedAt(), todo.UpdatedAt(), tPast, tCreate)
	}
}
