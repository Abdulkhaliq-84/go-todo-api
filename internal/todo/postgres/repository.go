// Package postgres implements domain.Repository against PostgreSQL.
//
// This is the OUTER ring. It imports domain; domain does not import it.
// Try reversing that and Go will refuse to compile -- import cycle.
//
// No ORM. Hand-written SQL, scanned into structs by hand. Coming from
// SQLAlchemy/SQLModel this feels like a step backwards for about a week, and
// then you notice you can read every query your app runs.
package postgres

import (
	"context"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is the concrete implementation.
//
// Nowhere in this file does it say "implements domain.Repository". Go
// interfaces are satisfied implicitly: if the method set matches, it fits.
// The compile-time assertion below is how Go programmers make that intent
// explicit and get a build error the moment it stops being true.
type Repository struct {
	pool *pgxpool.Pool
}

// Compile-time interface check. Costs nothing at runtime; the blank identifier
// discards the value and the compiler still type-checks the conversion.
var _ domain.Repository = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Save upserts.
//
//	INSERT INTO todos (...) VALUES ($1, ...)
//	ON CONFLICT (id) DO UPDATE SET
//	    title = EXCLUDED.title, ... , updated_at = EXCLUDED.updated_at
//
// The trade accepted with this choice: inserting an ID that already exists
// silently overwrites rather than failing loudly. Acceptable here because IDs
// are UUIDs minted by the domain, so a collision means a bug so severe that a
// unique-violation error would not be the thing that saves you.
//
// Do NOT include created_at in the DO UPDATE SET list -- an update must not
// rewrite the creation timestamp.
//
// TODO(you): write the SQL, pass todo's getters as parameters.
// Use $1, $2 placeholders -- never string concatenation, that is SQL injection.
func (r *Repository) Save(ctx context.Context, todo *domain.Todo) error {
	return nil // TODO
}

// FindByID loads one todo.
//
// TODO(you): SELECT the row, scan it, hand the values to toDomain().
// pgx returns pgx.ErrNoRows when nothing matched -- translate that into
// domain.ErrNotFound here. The domain must never see a pgx error type;
// that is what "infrastructure does not leak" means in practice.
func (r *Repository) FindByID(ctx context.Context, id domain.ID) (*domain.Todo, error) {
	return nil, nil // TODO
}

// FindAll loads todos matching a filter.
//
// TODO(you): translate domain.Filter into WHERE clauses. Building SQL
// conditionally is fiddly in Go -- a []string of conditions joined with " AND "
// plus a parallel []any of args is the common pattern.
func (r *Repository) FindAll(ctx context.Context, filter domain.Filter) ([]*domain.Todo, error) {
	return nil, nil // TODO
}

// Delete removes a todo.
//
// TODO(you): DELETE, then check the command tag's RowsAffected. Zero rows means
// it was not there -- return domain.ErrNotFound so the caller can 404.
func (r *Repository) Delete(ctx context.Context, id domain.ID) error {
	return nil // TODO
}
