package app_test

// Tests for the application layer: is the ORCHESTRATION right?
//
// What belongs here:
//   - the service calls the domain and persists the result
//   - domain errors propagate UNWRAPPED, so errors.Is still matches downstream
//   - a failed operation does not persist anything
//
// What does NOT belong here: business rules. "Cannot complete a completed todo"
// is tested in domain/todo_test.go. Testing it again here would mean the rule
// exists in two places, or worse, was implemented in the wrong layer.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/app"
	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// Dates far from "now" in both directions, so these tests are deterministic
// whenever they run without needing to control the clock.
var (
	longPast   = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	longFuture = time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
)

func newService(t *testing.T) (*app.Service, *fakeRepository) {
	t.Helper()
	repo := newFakeRepository()
	return app.NewService(repo), repo
}

func ptr[T any](v T) *T { return &v }

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestService_Create(t *testing.T) {
	svc, repo := newService(t)

	dto, err := svc.Create(context.Background(), app.CreateTodoCommand{
		Title:       "  Learn Go interfaces  ",
		Description: "implicit satisfaction",
		DueDate:     &longFuture,
	})
	if err != nil {
		t.Fatalf("Create: unexpected error %v", err)
	}

	if dto.ID == "" {
		t.Error("returned DTO has no id")
	}
	if dto.Title != "Learn Go interfaces" {
		t.Errorf("Title = %q, want the trimmed form", dto.Title)
	}
	if dto.Completed {
		t.Error("a new todo must not be completed")
	}
	if dto.Overdue {
		t.Error("a todo due in 2099 is not overdue")
	}
	if repo.count() != 1 {
		t.Errorf("repository holds %d todos, want 1", repo.count())
	}
}

// An invalid title must fail AND persist nothing. The second assertion is the
// one people forget -- it proves validation runs before the write, not after.
func TestService_Create_InvalidTitle(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr error
	}{
		{"empty", "", domain.ErrTitleEmpty},
		{"whitespace only", "   ", domain.ErrTitleEmpty},
		{"too long", strings.Repeat("a", domain.TitleMaxLength+1), domain.ErrTitleTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newService(t)

			_, err := svc.Create(context.Background(), app.CreateTodoCommand{Title: tt.title})

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create error = %v, want %v", err, tt.wantErr)
			}
			if repo.saveCalls != 0 {
				t.Errorf("repository was written to %d times despite invalid input", repo.saveCalls)
			}
		})
	}
}

func TestService_Create_RepositoryFailure(t *testing.T) {
	svc, repo := newService(t)
	boom := errors.New("connection refused")
	repo.saveErr = boom

	_, err := svc.Create(context.Background(), app.CreateTodoCommand{Title: "Task"})

	if !errors.Is(err, boom) {
		t.Errorf("Create error = %v, want the repository error to surface", err)
	}
}

// ---------------------------------------------------------------------------
// Read
// ---------------------------------------------------------------------------

func TestService_GetByID(t *testing.T) {
	svc, _ := newService(t)
	created, err := svc.Create(context.Background(), app.CreateTodoCommand{Title: "Task"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != created.ID || got.Title != created.Title {
		t.Errorf("GetByID returned %+v, want %+v", got, created)
	}
}

// domain.ErrNotFound must arrive UNWRAPPED, or errors.Is in the HTTP layer
// stops matching and the 404 silently becomes a 500.
func TestService_GetByID_NotFound(t *testing.T) {
	svc, _ := newService(t)

	_, err := svc.GetByID(context.Background(), domain.NewID().String())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID error = %v, want domain.ErrNotFound", err)
	}
}

// A malformed id is a 400, not a 404 -- so it must not be reported as "missing".
func TestService_GetByID_MalformedID(t *testing.T) {
	svc, _ := newService(t)

	_, err := svc.GetByID(context.Background(), "not-a-uuid")

	if !errors.Is(err, domain.ErrInvalidID) {
		t.Errorf("GetByID error = %v, want domain.ErrInvalidID", err)
	}
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestService_List_Filters(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()

	active, _ := svc.Create(ctx, app.CreateTodoCommand{Title: "Active"})
	overdue, _ := svc.Create(ctx, app.CreateTodoCommand{Title: "Overdue", DueDate: &longPast})
	done, _ := svc.Create(ctx, app.CreateTodoCommand{Title: "Done"})
	if _, err := svc.Complete(ctx, done.ID); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	tests := []struct {
		name  string
		query app.ListTodosQuery
		want  []string
	}{
		{"no filter returns all", app.ListTodosQuery{}, []string{active.ID, overdue.ID, done.ID}},
		{"completed only", app.ListTodosQuery{Completed: ptr(true)}, []string{done.ID}},
		{"active only", app.ListTodosQuery{Completed: ptr(false)}, []string{active.ID, overdue.ID}},
		{"overdue only", app.ListTodosQuery{Overdue: ptr(true)}, []string{overdue.ID}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.List(ctx, tt.query)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("List returned %d todos, want %d", len(got), len(tt.want))
			}
			ids := make(map[string]bool, len(got))
			for _, d := range got {
				ids[d.ID] = true
			}
			for _, want := range tt.want {
				if !ids[want] {
					t.Errorf("expected todo %s in results", want)
				}
			}
		})
	}
}

// A completed todo is never overdue, even one finished after its deadline --
// so filtering on overdue must not return it.
func TestService_List_CompletedIsNotOverdue(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()

	late, _ := svc.Create(ctx, app.CreateTodoCommand{Title: "Finished late", DueDate: &longPast})
	if _, err := svc.Complete(ctx, late.ID); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	got, err := svc.List(ctx, app.ListTodosQuery{Overdue: ptr(true)})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("overdue list returned %d todos, want 0 -- a completed todo is not overdue", len(got))
	}
}

func TestService_List_ClampsPagination(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	for i := 0; i < 25; i++ {
		if _, err := svc.Create(ctx, app.CreateTodoCommand{Title: "Task"}); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	// Limit 0 means "unspecified" and must become the default of 20, NOT zero
	// results -- the difference between a sensible page and an empty one.
	got, err := svc.List(ctx, app.ListTodosQuery{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 20 {
		t.Errorf("unspecified limit returned %d, want the default 20", len(got))
	}

	// An absurd limit is capped rather than honoured.
	got, err = svc.List(ctx, app.ListTodosQuery{Limit: 10000})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 25 {
		t.Errorf("huge limit returned %d, want all 25 (capped at 100)", len(got))
	}
}

// A list with no matches must serialise as [] rather than null, so clients
// calling data.map(...) do not break.
func TestService_List_EmptyIsNotNil(t *testing.T) {
	svc, _ := newService(t)

	got, err := svc.List(context.Background(), app.ListTodosQuery{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if got == nil {
		t.Error("List returned a nil slice; it must be empty but non-nil")
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestService_Update_PartialFields(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	created, _ := svc.Create(ctx, app.CreateTodoCommand{
		Title: "Original", Description: "keep me", DueDate: &longFuture,
	})

	// Only the title is sent. Everything else must survive untouched.
	got, err := svc.Update(ctx, created.ID, app.UpdateTodoCommand{Title: ptr("Renamed")})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if got.Title != "Renamed" {
		t.Errorf("Title = %q, want %q", got.Title, "Renamed")
	}
	if got.Description != "keep me" {
		t.Errorf("Description = %q, want it unchanged", got.Description)
	}
	if got.DueDate == nil || !got.DueDate.Equal(longFuture) {
		t.Errorf("DueDate = %v, want it unchanged", got.DueDate)
	}
}

// The three-state distinction that ClearDueDate exists for.
func TestService_Update_DueDateThreeStates(t *testing.T) {
	ctx := context.Background()

	t.Run("absent leaves it alone", func(t *testing.T) {
		svc, _ := newService(t)
		created, _ := svc.Create(ctx, app.CreateTodoCommand{Title: "T", DueDate: &longFuture})

		got, err := svc.Update(ctx, created.ID, app.UpdateTodoCommand{Title: ptr("New")})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if got.DueDate == nil {
			t.Error("an omitted due_date must not clear the stored one")
		}
	})

	t.Run("explicit clear removes it", func(t *testing.T) {
		svc, _ := newService(t)
		created, _ := svc.Create(ctx, app.CreateTodoCommand{Title: "T", DueDate: &longFuture})

		got, err := svc.Update(ctx, created.ID, app.UpdateTodoCommand{ClearDueDate: true})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if got.DueDate != nil {
			t.Errorf("DueDate = %v, want nil after an explicit clear", got.DueDate)
		}
	})

	t.Run("a value sets it", func(t *testing.T) {
		svc, _ := newService(t)
		created, _ := svc.Create(ctx, app.CreateTodoCommand{Title: "T"})

		got, err := svc.Update(ctx, created.ID, app.UpdateTodoCommand{DueDate: &longFuture})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if got.DueDate == nil || !got.DueDate.Equal(longFuture) {
			t.Errorf("DueDate = %v, want %v", got.DueDate, longFuture)
		}
	})
}

// An invalid title must abort the whole update -- not apply the description and
// then fail. Partial application is the bug this test exists to prevent.
func TestService_Update_InvalidTitleChangesNothing(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	created, _ := svc.Create(ctx, app.CreateTodoCommand{Title: "Original", Description: "original"})

	_, err := svc.Update(ctx, created.ID, app.UpdateTodoCommand{
		Title:       ptr(""),
		Description: ptr("should not be applied"),
	})
	if !errors.Is(err, domain.ErrTitleEmpty) {
		t.Fatalf("Update error = %v, want domain.ErrTitleEmpty", err)
	}

	after, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if after.Description != "original" {
		t.Errorf("Description = %q, want it untouched after a failed update", after.Description)
	}
}

// ---------------------------------------------------------------------------
// Complete / Reopen
// ---------------------------------------------------------------------------

func TestService_CompleteAndReopen(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	created, _ := svc.Create(ctx, app.CreateTodoCommand{Title: "Task"})

	done, err := svc.Complete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if !done.Completed {
		t.Error("Completed = false after Complete")
	}

	again, err := svc.Reopen(ctx, created.ID)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if again.Completed {
		t.Error("Completed = true after Reopen")
	}
}

// The strict state machine must reach the caller unwrapped, so the HTTP layer
// can map it to a 409.
func TestService_Complete_Twice(t *testing.T) {
	svc, repo := newService(t)
	ctx := context.Background()
	created, _ := svc.Create(ctx, app.CreateTodoCommand{Title: "Task"})
	if _, err := svc.Complete(ctx, created.ID); err != nil {
		t.Fatalf("first Complete: %v", err)
	}
	savesBefore := repo.saveCalls

	_, err := svc.Complete(ctx, created.ID)

	if !errors.Is(err, domain.ErrAlreadyComplete) {
		t.Fatalf("second Complete error = %v, want domain.ErrAlreadyComplete", err)
	}
	// A rejected transition must not reach the database.
	if repo.saveCalls != savesBefore {
		t.Errorf("a failed Complete still wrote to the repository")
	}
}

func TestService_Reopen_NotCompleted(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	created, _ := svc.Create(ctx, app.CreateTodoCommand{Title: "Task"})

	_, err := svc.Reopen(ctx, created.ID)

	if !errors.Is(err, domain.ErrNotCompleted) {
		t.Errorf("Reopen error = %v, want domain.ErrNotCompleted", err)
	}
}

func TestService_Complete_NotFound(t *testing.T) {
	svc, _ := newService(t)

	_, err := svc.Complete(context.Background(), domain.NewID().String())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Complete error = %v, want domain.ErrNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestService_Delete(t *testing.T) {
	svc, repo := newService(t)
	ctx := context.Background()
	created, _ := svc.Create(ctx, app.CreateTodoCommand{Title: "Task"})

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if repo.count() != 0 {
		t.Errorf("repository still holds %d todos", repo.count())
	}

	// Deleting again must report not-found, not succeed silently.
	if err := svc.Delete(ctx, created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("second Delete error = %v, want domain.ErrNotFound", err)
	}
}

func TestService_Delete_MalformedID(t *testing.T) {
	svc, _ := newService(t)

	if err := svc.Delete(context.Background(), "nope"); !errors.Is(err, domain.ErrInvalidID) {
		t.Errorf("Delete error = %v, want domain.ErrInvalidID", err)
	}
}
