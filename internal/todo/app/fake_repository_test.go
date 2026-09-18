package app_test

import (
	"context"
	"sync"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// fakeRepository is an in-memory stand-in for the real Postgres repository.
//
// THIS IS THE PAYOFF FOR THE WHOLE ARCHITECTURE. Because app.Service depends on
// the domain.Repository INTERFACE rather than on *postgres.Repository, the
// entire application layer can be tested with no database, no Docker, and no
// network -- in microseconds.
//
// Note there is no mocking library. No unittest.mock, no jest.mock, no
// gomock-generated file. Go's implicit interfaces mean a hand-written struct
// with the right methods simply IS a domain.Repository. Most Go codebases write
// fakes like this by hand and are better off for it: a fake has real behaviour
// you can reason about, where a mock only has recorded expectations.
type fakeRepository struct {
	mu    sync.RWMutex // handlers run concurrently; tests with -race will catch you
	items map[string]*domain.Todo

	// Error injection: set these to force a failure path in a test.
	saveErr error
	findErr error
}

// Compile-time proof the fake satisfies the interface the service needs. If you
// add a method to domain.Repository, this line breaks here rather than in a
// confusing error somewhere else.
var _ domain.Repository = (*fakeRepository)(nil)

// TODO(you)
func newFakeRepository() *fakeRepository {
	return nil // TODO
}

// TODO(you): honour saveErr, then store the todo by id.String().
func (f *fakeRepository) Save(ctx context.Context, todo *domain.Todo) error {
	return nil // TODO
}

// TODO(you): honour findErr, look up by id, return domain.ErrNotFound when absent.
func (f *fakeRepository) FindByID(ctx context.Context, id domain.ID) (*domain.Todo, error) {
	return nil, nil // TODO
}

// TODO(you): apply the filter in memory. Map iteration order in Go is
// RANDOMISED on purpose, so sort the results before returning or your tests
// will pass and fail at random.
func (f *fakeRepository) FindAll(ctx context.Context, filter domain.Filter) ([]*domain.Todo, error) {
	return nil, nil // TODO
}

// TODO(you)
func (f *fakeRepository) Delete(ctx context.Context, id domain.ID) error {
	return nil // TODO
}
