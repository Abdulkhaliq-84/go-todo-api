// Package app holds the use cases of the Todo context.
//
// The rule for this layer: ORCHESTRATE, never DECIDE.
// "Can this todo be completed?" is a domain decision and belongs on the entity.
// "Load it, tell it to complete, save it" is orchestration and belongs here.
//
// If you find an `if` in this package that encodes a business rule, it escaped
// from the domain and should go back.
//
// This layer imports domain. It does NOT import postgres or net/http.
package app

import (
	"context"
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// Pagination bounds, applied here rather than in the transport layer so every
// caller gets them -- an HTTP handler, a CLI, a future gRPC server, a test.
// A transport-level default would have to be repeated in each one.
const (
	defaultLimit = 20
	maxLimit     = 100
)

// Service is the entry point for every use case in this context.
//
// The field is domain.Repository, the INTERFACE -- not *postgres.Repository.
// This package therefore has no idea Postgres exists, and its tests pass a fake
// with no database running.
//
// Dependency injection in Go is a struct field. No container, no decorators.
// Wiring happens once, by hand, in cmd/api/main.go.
type Service struct {
	repo domain.Repository
}

// NewService accepts an interface and returns a struct -- the Go proverb.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// ---------------------------------------------------------------------------
// ERRORS
//
// Domain errors are returned UNWRAPPED, deliberately.
//
// Wrapping ErrAlreadyComplete in a service-level error would stop
// errors.Is(err, domain.ErrAlreadyComplete) from matching in the HTTP layer,
// and the 409 would silently become a 500 -- no compile error, no test failure
// unless one was written for it. This is the sharpest edge of errors-as-values
// versus exceptions: nothing propagates automatically, so a careless wrap is a
// real regression.
// ---------------------------------------------------------------------------

// Create adds a new todo.
func (s *Service) Create(ctx context.Context, cmd CreateTodoCommand) (TodoDTO, error) {
	title, err := domain.NewTitle(cmd.Title)
	if err != nil {
		return TodoDTO{}, err
	}

	now := time.Now()

	todo, err := domain.New(title, cmd.Description, cmd.DueDate, now)
	if err != nil {
		return TodoDTO{}, err
	}
	if err := s.repo.Save(ctx, todo); err != nil {
		return TodoDTO{}, err
	}
	return toDTO(todo, now), nil
}

// GetByID fetches one todo.
func (s *Service) GetByID(ctx context.Context, id string) (TodoDTO, error) {
	todo, err := s.find(ctx, id)
	if err != nil {
		return TodoDTO{}, err
	}
	return toDTO(todo, time.Now()), nil
}

// List returns todos matching a filter.
func (s *Service) List(ctx context.Context, q ListTodosQuery) ([]TodoDTO, error) {
	filter := domain.Filter{
		Completed: q.Completed,
		Overdue:   q.Overdue,
		Limit:     clampLimit(q.Limit),
		Offset:    max(q.Offset, 0),
	}

	todos, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// One time.Now() for the whole page, so every todo in it is judged against
	// the same instant.
	return toDTOs(todos, time.Now()), nil
}

// Update changes a todo's editable fields.
//
// PATCH semantics: only the fields the client actually sent are touched.
// cmd.Title and cmd.Description are nil when absent; cmd.ClearDueDate carries
// "the client explicitly asked to clear it", which nil alone cannot express.
func (s *Service) Update(ctx context.Context, id string, cmd UpdateTodoCommand) (TodoDTO, error) {
	todo, err := s.find(ctx, id)
	if err != nil {
		return TodoDTO{}, err
	}

	now := time.Now()

	// Validate BEFORE mutating anything. An invalid title must leave the
	// aggregate untouched, not half-updated with the description applied.
	var title domain.Title
	if cmd.Title != nil {
		title, err = domain.NewTitle(*cmd.Title)
		if err != nil {
			return TodoDTO{}, err
		}
	}

	if cmd.Title != nil {
		if err := todo.UpdateTitle(title, now); err != nil {
			return TodoDTO{}, err
		}
	}
	if cmd.Description != nil {
		if err := todo.UpdateDescription(*cmd.Description, now); err != nil {
			return TodoDTO{}, err
		}
	}

	// ClearDueDate wins over DueDate: a client that sends both is contradicting
	// itself, and "clear it" is the less destructive reading.
	switch {
	case cmd.ClearDueDate:
		if err := todo.Reschedule(nil, now); err != nil {
			return TodoDTO{}, err
		}
	case cmd.DueDate != nil:
		if err := todo.Reschedule(cmd.DueDate, now); err != nil {
			return TodoDTO{}, err
		}
	}

	if err := s.repo.Save(ctx, todo); err != nil {
		return TodoDTO{}, err
	}
	return toDTO(todo, now), nil
}

// Complete marks a todo done.
//
// Note how thin this is: load, tell the aggregate, save. All the rules live in
// todo.Complete. If this method ever grows past a dozen lines, business logic
// is leaking out of the domain.
func (s *Service) Complete(ctx context.Context, id string) (TodoDTO, error) {
	todo, err := s.find(ctx, id)
	if err != nil {
		return TodoDTO{}, err
	}

	now := time.Now()
	if err := todo.Complete(now); err != nil {
		return TodoDTO{}, err // unwrapped: the HTTP layer maps this to 409
	}
	if err := s.repo.Save(ctx, todo); err != nil {
		return TodoDTO{}, err
	}
	return toDTO(todo, now), nil
}

// Reopen moves a completed todo back to active.
func (s *Service) Reopen(ctx context.Context, id string) (TodoDTO, error) {
	todo, err := s.find(ctx, id)
	if err != nil {
		return TodoDTO{}, err
	}

	now := time.Now()
	if err := todo.Reopen(now); err != nil {
		return TodoDTO{}, err
	}
	if err := s.repo.Save(ctx, todo); err != nil {
		return TodoDTO{}, err
	}
	return toDTO(todo, now), nil
}

// Delete removes a todo.
func (s *Service) Delete(ctx context.Context, id string) error {
	todoID, err := domain.ParseID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, todoID)
}

// ---------------------------------------------------------------------------
// internals
// ---------------------------------------------------------------------------

// find parses an id and loads the aggregate. Every use case but Create and List
// starts this way, so it lives in one place -- and a malformed id returns
// ErrInvalidID (400) rather than ErrNotFound (404), which is the honest answer.
func (s *Service) find(ctx context.Context, id string) (*domain.Todo, error) {
	todoID, err := domain.ParseID(id)
	if err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, todoID)
}

func clampLimit(limit int) int {
	switch {
	case limit <= 0:
		return defaultLimit
	case limit > maxLimit:
		return maxLimit
	default:
		return limit
	}
}
