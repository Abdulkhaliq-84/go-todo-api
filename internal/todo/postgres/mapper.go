package postgres

import (
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// todoRow is the database's shape of a todo: flat, primitive, nullable.
// Kept separate from domain.Todo on purpose -- the DB schema and the domain
// model are allowed to drift apart, and this struct is where they reconcile.
type todoRow struct {
	ID          string
	Title       string
	Description string
	Completed   bool
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// toDomain rebuilds a domain entity from a database row.
//
// It calls domain.Reconstitute, NOT domain.New -- this row already satisfied
// the creation rules when it was first written. Re-running them here would be
// wrong (it would mint a new ID and reset createdAt) and could even reject rows
// that were valid under older rules.
//
// TODO(you): parse the ID, rebuild the Title, call Reconstitute.
func toDomain(row todoRow) (*domain.Todo, error) {
	return nil, nil // TODO
}

// toRow flattens a domain entity for storage.
//
// TODO(you): read the entity's getters into the row struct.
func toRow(t *domain.Todo) todoRow {
	return todoRow{} // TODO
}
