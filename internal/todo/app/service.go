// Package app holds the use cases of the Todo context.
//
// The rule for this layer: ORCHESTRATE, never DECIDE.
// "Can this todo be completed?" is a domain decision -- it belongs on the
// entity. "Load it, tell it to complete, save it" is orchestration -- that
// belongs here.
//
// If you find an `if` statement in this package that encodes a business rule,
// it escaped from the domain and should be moved back.
//
// This layer imports domain. It does NOT import postgres or net/http.
package app

import (
	"context"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// Service is the entry point for every use case in this context.
//
// Look at the field: it is domain.Repository, the INTERFACE -- not
// *postgres.Repository, the concrete type. This package therefore has no idea
// Postgres exists, and your tests can pass a fake repository with no database
// running at all.
//
// This is dependency injection in Go: a struct field. No @Depends, no
// container, no decorators. Wiring happens once, by hand, in cmd/api/main.go.
type Service struct {
	repo domain.Repository
}

// NewService is the constructor. Accepting an interface and storing it is the
// entire pattern -- "accept interfaces, return structs" is the Go proverb.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// ---------------------------------------------------------------------------
// Use cases
//
// One method per thing a user can do. Each follows the same rhythm:
//   1. turn primitive input into domain types (validation happens there)
//   2. load the aggregate, if the use case needs one
//   3. call a method ON the aggregate -- the rules run inside the domain
//   4. persist
//   5. map to an output DTO
// ---------------------------------------------------------------------------

// Create adds a new todo.
//
// TODO(you): build a domain.Title from cmd.Title, call domain.New, save, map out.
func (s *Service) Create(ctx context.Context, cmd CreateTodoCommand) (TodoDTO, error) {
	return TodoDTO{}, nil // TODO
}

// GetByID fetches one todo.
//
// TODO(you): parse the id, find it, map it. Let domain.ErrNotFound bubble up
// unchanged -- the HTTP layer is what turns it into a 404.
func (s *Service) GetByID(ctx context.Context, id string) (TodoDTO, error) {
	return TodoDTO{}, nil // TODO
}

// List returns todos matching a filter.
//
// TODO(you): build a domain.Filter from the query, fetch, map the slice.
func (s *Service) List(ctx context.Context, q ListTodosQuery) ([]TodoDTO, error) {
	return nil, nil // TODO
}

// Update changes a todo's editable fields.
//
// >>> THIS IS ONE OF YOUR DECISIONS -- see the note I left you. <<<
// The shape of this method depends on whether you want PUT (full replace) or
// PATCH (partial update) semantics.
//
// TODO(you)
func (s *Service) Update(ctx context.Context, id string, cmd UpdateTodoCommand) (TodoDTO, error) {
	return TodoDTO{}, nil // TODO
}

// Complete marks a todo done.
//
// Notice how thin this should be: load, todo.Complete(), save. If it grows
// past five lines, business logic is leaking out of the domain.
//
// TODO(you)
func (s *Service) Complete(ctx context.Context, id string) (TodoDTO, error) {
	return TodoDTO{}, nil // TODO
}

// Reopen moves a completed todo back to active.
//
// TODO(you)
func (s *Service) Reopen(ctx context.Context, id string) (TodoDTO, error) {
	return TodoDTO{}, nil // TODO
}

// Delete removes a todo.
//
// TODO(you)
func (s *Service) Delete(ctx context.Context, id string) error {
	return nil // TODO
}
