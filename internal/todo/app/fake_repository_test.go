package app_test

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// fakeRepository is an in-memory stand-in for the Postgres repository.
//
// THIS IS THE PAYOFF FOR THE WHOLE ARCHITECTURE. Because app.Service depends on
// the domain.Repository INTERFACE rather than on *postgres.Repository, the
// entire application layer is testable with no database, no Docker and no
// network -- in microseconds.
//
// Note there is no mocking library: no unittest.mock, no jest.mock, no
// generated gomock file. Go's implicit interfaces mean a hand-written struct
// with the right method set simply IS a domain.Repository. Most Go codebases
// write fakes by hand and are better off for it -- a fake has real behaviour
// you can reason about, where a mock only has recorded expectations.
type fakeRepository struct {
	mu    sync.RWMutex // handlers run concurrently; `go test -race` will catch you
	items map[string]*domain.Todo

	// Error injection: set these to force a failure path.
	saveErr   error
	findErr   error
	deleteErr error

	// Call counts, for asserting that something did NOT happen.
	saveCalls int
}

// Compile-time proof the fake satisfies the interface the service needs. Add a
// method to domain.Repository and this line breaks here, rather than somewhere
// confusing later.
var _ domain.Repository = (*fakeRepository)(nil)

func newFakeRepository() *fakeRepository {
	return &fakeRepository{items: make(map[string]*domain.Todo)}
}

func (f *fakeRepository) Save(ctx context.Context, todo *domain.Todo) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.saveCalls++
	if f.saveErr != nil {
		return f.saveErr
	}
	f.items[todo.ID().String()] = todo
	return nil
}

func (f *fakeRepository) FindByID(ctx context.Context, id domain.ID) (*domain.Todo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.findErr != nil {
		return nil, f.findErr
	}
	todo, ok := f.items[id.String()]
	if !ok {
		// Returns the DOMAIN error, exactly as the Postgres repository must
		// translate pgx.ErrNoRows. If the fake returned something else, tests
		// would pass against behaviour the real repository does not have.
		return nil, domain.ErrNotFound
	}
	return todo, nil
}

func (f *fakeRepository) FindAll(ctx context.Context, filter domain.Filter) ([]*domain.Todo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.findErr != nil {
		return nil, f.findErr
	}

	now := time.Now()
	matched := make([]*domain.Todo, 0, len(f.items))
	for _, todo := range f.items {
		if filter.Completed != nil && todo.IsCompleted() != *filter.Completed {
			continue
		}
		if filter.Overdue != nil && todo.IsOverdue(now) != *filter.Overdue {
			continue
		}
		matched = append(matched, todo)
	}

	// Go RANDOMISES map iteration order on purpose, so results must be sorted
	// or the tests pass and fail at random. Newest first, matching the
	// (completed, created_at DESC) index the real table carries.
	sort.Slice(matched, func(i, j int) bool {
		if !matched[i].CreatedAt().Equal(matched[j].CreatedAt()) {
			return matched[i].CreatedAt().After(matched[j].CreatedAt())
		}
		return matched[i].ID().String() < matched[j].ID().String() // stable tiebreak
	})

	return paginate(matched, filter.Offset, filter.Limit), nil
}

func (f *fakeRepository) Delete(ctx context.Context, id domain.ID) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.deleteErr != nil {
		return f.deleteErr
	}
	if _, ok := f.items[id.String()]; !ok {
		return domain.ErrNotFound
	}
	delete(f.items, id.String())
	return nil
}

// count is a test helper, not part of the interface.
func (f *fakeRepository) count() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.items)
}

// paginate slices defensively. Go panics on an out-of-range slice, so an offset
// past the end must be clamped rather than trusted -- the same care the real
// SQL needs around OFFSET.
func paginate(items []*domain.Todo, offset, limit int) []*domain.Todo {
	if offset >= len(items) {
		return []*domain.Todo{}
	}
	items = items[offset:]
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	return items
}
