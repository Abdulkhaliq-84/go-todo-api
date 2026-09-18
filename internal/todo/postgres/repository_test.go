//go:build integration

// The build tag means `go test ./...` SKIPS this file entirely. It compiles
// only with `go test -tags=integration ./...`.
//
// Why: these tests need a real PostgreSQL. Unit tests must stay fast enough to
// run on every save; integration tests run when you ask, and in CI. Mixing the
// two is how suites get slow enough that people stop running them.
//
// The tag must be the first line, followed by a blank line, before `package`.

package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// defaultTestDB points at a local PostgreSQL. Override with TEST_DATABASE_URL
// to run against Docker or CI instead.
const defaultTestDB = "postgres://localhost:5432/todos_test?sslmode=disable"

// setupTestDB gives each test a clean table.
//
// Chosen approach: one shared database, truncated before each test. The
// alternatives were testcontainers (zero setup for a fresh clone, several
// seconds per run) and a transaction rolled back per test (fastest and
// perfectly isolated, but the repository would have to accept a tx instead of a
// pool, changing its signature to suit the tests).
//
// TRUNCATE rather than DELETE: it does not scan the table and it resets cleanly.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = defaultTestDB
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connect to %s: %v", url, err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping %s: %v (is postgres running? createdb todos_test)", url, err)
	}
	if _, err := pool.Exec(ctx, "TRUNCATE todos"); err != nil {
		t.Fatalf("truncate: %v (did you apply migrations/000001_create_todos.up.sql?)", err)
	}

	// t.Cleanup runs after the test whether it passed, failed or panicked --
	// Go's answer to a pytest fixture teardown, and more reliable than defer in
	// a helper, which would run before the test body even started.
	t.Cleanup(pool.Close)
	return pool
}

func newTodo(t *testing.T, title string, due *time.Time) *domain.Todo {
	t.Helper()
	name, err := domain.NewTitle(title)
	if err != nil {
		t.Fatalf("NewTitle: %v", err)
	}
	todo, err := domain.New(name, "desc", due, time.Now())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return todo
}

func ptr[T any](v T) *T { return &v }

// ---------------------------------------------------------------------------
// Round trip -- the most valuable test in this file
// ---------------------------------------------------------------------------

func TestRepository_SaveAndFindByID(t *testing.T) {
	repo := postgres.NewRepository(setupTestDB(t))
	ctx := context.Background()

	// A due date with non-zero sub-second precision and a non-UTC zone: the two
	// things that get silently mangled between Go and Postgres.
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	due := time.Date(2026, 7, 4, 15, 30, 45, 123456000, tokyo)

	original := newTodo(t, "Round trip", &due)
	if err := repo.Save(ctx, original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := repo.FindByID(ctx, original.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}

	if !loaded.ID().Equals(original.ID()) {
		t.Errorf("ID = %v, want %v", loaded.ID(), original.ID())
	}
	if loaded.Title().String() != original.Title().String() {
		t.Errorf("Title = %q, want %q", loaded.Title(), original.Title())
	}
	if loaded.Description() != original.Description() {
		t.Errorf("Description = %q, want %q", loaded.Description(), original.Description())
	}
	if loaded.IsCompleted() != original.IsCompleted() {
		t.Errorf("Completed = %v, want %v", loaded.IsCompleted(), original.IsCompleted())
	}

	// THE TIMEZONE ASSERTION. This is what catches TIMESTAMP-instead-of-
	// TIMESTAMPTZ, and it is why .Equal is used rather than ==. Two time.Time
	// values can name the same instant in different zones: == compares the
	// struct (zone included) and fails; .Equal compares the instant and passes.
	// Using == here would produce a test that fails for the wrong reason, and
	// the usual "fix" is to weaken the assertion and lose the real check.
	if loaded.DueDate() == nil {
		t.Fatal("DueDate came back nil")
	}
	if !loaded.DueDate().Equal(due) {
		t.Errorf("DueDate = %v, want the same instant as %v", loaded.DueDate(), due)
	}
	if !loaded.CreatedAt().Equal(original.CreatedAt()) {
		t.Errorf("CreatedAt = %v, want %v", loaded.CreatedAt(), original.CreatedAt())
	}
}

func TestRepository_SaveAndFindByID_NoDueDate(t *testing.T) {
	repo := postgres.NewRepository(setupTestDB(t))
	ctx := context.Background()

	original := newTodo(t, "No deadline", nil)
	if err := repo.Save(ctx, original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := repo.FindByID(ctx, original.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if loaded.DueDate() != nil {
		t.Errorf("DueDate = %v, want nil -- a NULL column must stay nil", loaded.DueDate())
	}
}

// Save is an upsert: saving twice updates rather than failing on the primary
// key, and must not rewrite created_at.
func TestRepository_Save_Upserts(t *testing.T) {
	repo := postgres.NewRepository(setupTestDB(t))
	ctx := context.Background()

	todo := newTodo(t, "First", nil)
	if err := repo.Save(ctx, todo); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	createdAt := todo.CreatedAt()

	later := time.Now().Add(time.Hour)
	if err := todo.Complete(later); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if err := repo.Save(ctx, todo); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	loaded, err := repo.FindByID(ctx, todo.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if !loaded.IsCompleted() {
		t.Error("the second Save did not update the row")
	}
	if !loaded.CreatedAt().Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want the original %v -- an update must not rewrite it",
			loaded.CreatedAt(), createdAt)
	}
	if !loaded.UpdatedAt().Equal(later.UTC()) {
		t.Errorf("UpdatedAt = %v, want %v", loaded.UpdatedAt(), later.UTC())
	}
}

// An unknown id must return domain.ErrNotFound, NOT pgx.ErrNoRows. This test
// guards the infrastructure boundary: it proves driver types do not leak out of
// this package and into the domain.
func TestRepository_FindByID_NotFound(t *testing.T) {
	repo := postgres.NewRepository(setupTestDB(t))

	_, err := repo.FindByID(context.Background(), domain.NewID())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("FindByID error = %v, want domain.ErrNotFound", err)
	}
}

// Values are sent out of band as query parameters, never concatenated into SQL,
// so this title is stored as literal text.
func TestRepository_Save_SQLInjectionIsStoredAsText(t *testing.T) {
	repo := postgres.NewRepository(setupTestDB(t))
	ctx := context.Background()

	nasty := "'; DROP TABLE todos; --"
	todo := newTodo(t, nasty, nil)
	if err := repo.Save(ctx, todo); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := repo.FindByID(ctx, todo.ID())
	if err != nil {
		t.Fatalf("FindByID: %v (the table may be gone)", err)
	}
	if loaded.Title().String() != nasty {
		t.Errorf("Title = %q, want it stored verbatim", loaded.Title())
	}
}

// ---------------------------------------------------------------------------
// Queries
// ---------------------------------------------------------------------------

func TestRepository_FindAll_Filters(t *testing.T) {
	repo := postgres.NewRepository(setupTestDB(t))
	ctx := context.Background()
	now := time.Now()

	past := now.Add(-48 * time.Hour)
	future := now.Add(48 * time.Hour)

	active := newTodo(t, "Active", nil)
	overdue := newTodo(t, "Overdue", &past)
	upcoming := newTodo(t, "Upcoming", &future)
	doneLate := newTodo(t, "Done late", &past)
	if err := doneLate.Complete(now); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	for _, todo := range []*domain.Todo{active, overdue, upcoming, doneLate} {
		if err := repo.Save(ctx, todo); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	tests := []struct {
		name   string
		filter domain.Filter
		want   []domain.ID
	}{
		{"no filter", domain.Filter{Limit: 10},
			[]domain.ID{active.ID(), overdue.ID(), upcoming.ID(), doneLate.ID()}},
		{"completed", domain.Filter{Completed: ptr(true), Limit: 10},
			[]domain.ID{doneLate.ID()}},
		{"active", domain.Filter{Completed: ptr(false), Limit: 10},
			[]domain.ID{active.ID(), overdue.ID(), upcoming.ID()}},
		// doneLate is past its deadline but finished, so it is NOT overdue --
		// the same rule the domain applies, here expressed in SQL.
		{"overdue", domain.Filter{Overdue: ptr(true), Limit: 10},
			[]domain.ID{overdue.ID()}},
		{"not overdue", domain.Filter{Overdue: ptr(false), Limit: 10},
			[]domain.ID{active.ID(), upcoming.ID(), doneLate.ID()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindAll(ctx, tt.filter)
			if err != nil {
				t.Fatalf("FindAll: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("FindAll returned %d todos, want %d", len(got), len(tt.want))
			}
			found := make(map[string]bool, len(got))
			for _, todo := range got {
				found[todo.ID().String()] = true
			}
			for _, want := range tt.want {
				if !found[want.String()] {
					t.Errorf("expected %s in results", want)
				}
			}
		})
	}
}

// Pagination off-by-ones are common and invisible until production.
func TestRepository_FindAll_Pagination(t *testing.T) {
	repo := postgres.NewRepository(setupTestDB(t))
	ctx := context.Background()

	// Distinct creation times so the ORDER BY is meaningful rather than relying
	// on the id tiebreak.
	base := time.Now().Add(-10 * time.Hour)
	for i := 0; i < 5; i++ {
		name, err := domain.NewTitle("Task")
		if err != nil {
			t.Fatalf("NewTitle: %v", err)
		}
		todo, err := domain.New(name, "", nil, base.Add(time.Duration(i)*time.Hour))
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if err := repo.Save(ctx, todo); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	page1, err := repo.FindAll(ctx, domain.Filter{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("FindAll page 1: %v", err)
	}
	page2, err := repo.FindAll(ctx, domain.Filter{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("FindAll page 2: %v", err)
	}
	page3, err := repo.FindAll(ctx, domain.Filter{Limit: 2, Offset: 4})
	if err != nil {
		t.Fatalf("FindAll page 3: %v", err)
	}

	if len(page1) != 2 || len(page2) != 2 || len(page3) != 1 {
		t.Fatalf("page sizes = %d/%d/%d, want 2/2/1", len(page1), len(page2), len(page3))
	}

	// No row may appear on two pages -- the bug an unstable ORDER BY causes.
	seen := map[string]bool{}
	for _, page := range [][]*domain.Todo{page1, page2, page3} {
		for _, todo := range page {
			if seen[todo.ID().String()] {
				t.Errorf("todo %s appeared on two pages", todo.ID())
			}
			seen[todo.ID().String()] = true
		}
	}
	if len(seen) != 5 {
		t.Errorf("paged through %d distinct todos, want 5", len(seen))
	}

	// Newest first.
	if !page1[0].CreatedAt().After(page1[1].CreatedAt()) {
		t.Error("results are not ordered newest first")
	}
}

func TestRepository_FindAll_OffsetPastEnd(t *testing.T) {
	repo := postgres.NewRepository(setupTestDB(t))
	ctx := context.Background()
	if err := repo.Save(ctx, newTodo(t, "Only one", nil)); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindAll(ctx, domain.Filter{Limit: 10, Offset: 100})
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("FindAll returned %d todos past the end, want 0", len(got))
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestRepository_Delete(t *testing.T) {
	repo := postgres.NewRepository(setupTestDB(t))
	ctx := context.Background()

	todo := newTodo(t, "Doomed", nil)
	if err := repo.Save(ctx, todo); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := repo.Delete(ctx, todo.ID()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByID(ctx, todo.ID()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("after Delete, FindByID error = %v, want domain.ErrNotFound", err)
	}
}

// Deleting a missing row must report not-found. Postgres considers a DELETE
// matching nothing a success, so the repository has to check RowsAffected --
// without it the API answers 204 for a todo that never existed.
func TestRepository_Delete_NotFound(t *testing.T) {
	repo := postgres.NewRepository(setupTestDB(t))

	err := repo.Delete(context.Background(), domain.NewID())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete error = %v, want domain.ErrNotFound", err)
	}
}
