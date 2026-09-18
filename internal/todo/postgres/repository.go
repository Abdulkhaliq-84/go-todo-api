// Package postgres implements domain.Repository against PostgreSQL.
//
// This is the OUTER ring. It imports domain; domain does not import it. Try
// reversing that and Go refuses to compile -- import cycle.
//
// No ORM. Hand-written SQL, scanned into structs by hand. Coming from
// SQLAlchemy or Prisma this feels like a step backwards for about a week, and
// then you notice you can read every query the application runs.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Selected columns. id is cast to text so it scans into a plain string and the
// domain stays free of any UUID type belonging to a driver.
const selectColumns = `id::text, title, description, completed, due_date, created_at, updated_at`

// Repository is the concrete implementation.
//
// Nowhere does it say "implements domain.Repository" -- Go interfaces are
// satisfied structurally, exactly like TypeScript's. The assertion below is how
// Go programmers make the intent explicit and get a build error the moment it
// stops being true.
type Repository struct {
	pool *pgxpool.Pool
}

var _ domain.Repository = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Save upserts.
//
// One method rather than Insert + Update, so the caller never tracks whether an
// entity is new -- that is a persistence concern and it stays on this side of
// the boundary.
//
// created_at is deliberately absent from the DO UPDATE SET list: an update must
// never rewrite when the row was created.
func (r *Repository) Save(ctx context.Context, todo *domain.Todo) error {
	const query = `
		INSERT INTO todos (id, title, description, completed, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			title       = EXCLUDED.title,
			description = EXCLUDED.description,
			completed   = EXCLUDED.completed,
			due_date    = EXCLUDED.due_date,
			updated_at  = EXCLUDED.updated_at`

	row := toRow(todo)

	// $1, $2 placeholders -- never string concatenation. The driver sends
	// values out of band, so a title of "'; DROP TABLE todos; --" is stored as
	// that literal text and never parsed as SQL.
	_, err := r.pool.Exec(ctx, query,
		row.ID, row.Title, row.Description, row.Completed,
		row.DueDate, row.CreatedAt, row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres: save todo %s: %w", row.ID, err)
	}
	return nil
}

// FindByID loads one todo.
func (r *Repository) FindByID(ctx context.Context, id domain.ID) (*domain.Todo, error) {
	query := `SELECT ` + selectColumns + ` FROM todos WHERE id = $1`

	var row todoRow
	err := r.pool.QueryRow(ctx, query, id.String()).Scan(
		&row.ID, &row.Title, &row.Description, &row.Completed,
		&row.DueDate, &row.CreatedAt, &row.UpdatedAt,
	)

	// THE BOUNDARY. pgx.ErrNoRows is a driver detail; the domain must never see
	// it. Translating here is what lets the service and the HTTP layer speak
	// only domain errors, and what would let a different database be swapped in
	// without touching a line outside this package.
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: find todo %s: %w", id, err)
	}

	return toDomain(row)
}

// FindAll loads todos matching a filter.
//
// Building SQL conditionally is fiddly in Go: a slice of condition strings and
// a parallel slice of arguments, joined at the end. Verbose next to a query
// builder, and the resulting SQL is exactly what you wrote.
func (r *Repository) FindAll(ctx context.Context, filter domain.Filter) ([]*domain.Todo, error) {
	var (
		conditions []string
		args       []any
	)

	if filter.Completed != nil {
		args = append(args, *filter.Completed)
		conditions = append(conditions, fmt.Sprintf("completed = $%d", len(args)))
	}

	if filter.Overdue != nil {
		// "Overdue" is DERIVED, never stored -- the same definition the domain
		// uses, expressed in SQL: outstanding, has a deadline, and past it.
		// A completed todo is never overdue, however late it was finished.
		args = append(args, time.Now().UTC())
		overdue := fmt.Sprintf(
			"(completed = FALSE AND due_date IS NOT NULL AND due_date < $%d)", len(args))

		if *filter.Overdue {
			conditions = append(conditions, overdue)
		} else {
			// The IS NOT NULL guard keeps this three-valued-logic-safe: with a
			// NULL due_date the inner expression is FALSE, not NULL, so NOT of
			// it is TRUE and rows without deadlines are correctly included.
			conditions = append(conditions, "NOT "+overdue)
		}
	}

	query := `SELECT ` + selectColumns + ` FROM todos`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// id is the tiebreak so the order is total. Without it, rows sharing a
	// created_at come back in an arbitrary order and pagination can show or
	// skip the same row twice.
	query += " ORDER BY created_at DESC, id"

	args = append(args, filter.Limit)
	query += fmt.Sprintf(" LIMIT $%d", len(args))
	args = append(args, filter.Offset)
	query += fmt.Sprintf(" OFFSET $%d", len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres: list todos: %w", err)
	}
	defer rows.Close() // releases the connection back to the pool

	var collected []todoRow
	for rows.Next() {
		var row todoRow
		if err := rows.Scan(
			&row.ID, &row.Title, &row.Description, &row.Completed,
			&row.DueDate, &row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("postgres: scan todo: %w", err)
		}
		collected = append(collected, row)
	}
	// rows.Err() reports a failure that happened mid-iteration. Skipping this
	// check is a classic Go bug: the loop simply ends early and you silently
	// return a partial result as if it were the whole set.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterate todos: %w", err)
	}

	return toDomainAll(collected)
}

// Delete removes a todo.
func (r *Repository) Delete(ctx context.Context, id domain.ID) error {
	const query = `DELETE FROM todos WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("postgres: delete todo %s: %w", id, err)
	}

	// A DELETE that matches nothing is not an error as far as Postgres is
	// concerned -- you have to notice yourself. Without this check the API
	// would answer 204 for a todo that was never there.
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
