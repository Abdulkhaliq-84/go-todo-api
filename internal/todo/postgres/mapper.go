package postgres

import (
	"fmt"
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// todoRow is the DATABASE's shape of a todo: flat, primitive, nullable.
//
// Kept separate from domain.Todo on purpose. The schema and the domain model
// are allowed to drift apart, and this struct is the one place they reconcile.
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
// the creation rules when it was written. Re-running them would mint a fresh ID
// and reset createdAt, which is simply wrong.
//
// The tension worth naming: Title can only be built through NewTitle, so
// reconstruction does re-validate the title even though reconstruction is meant
// to bypass creation rules. The consequence is real -- tighten TitleMaxLength
// later and old rows stop loading. The alternative (an unvalidated
// ReconstituteTitle) lets corrupt data into the domain silently, which is
// worse. A row that cannot form a valid Title is treated as corruption and
// surfaced loudly rather than papered over.
func toDomain(row todoRow) (*domain.Todo, error) {
	id, err := domain.ParseID(row.ID)
	if err != nil {
		return nil, fmt.Errorf("corrupt row: id %q: %w", row.ID, err)
	}
	title, err := domain.NewTitle(row.Title)
	if err != nil {
		return nil, fmt.Errorf("corrupt row %s: title: %w", row.ID, err)
	}

	return domain.Reconstitute(
		id,
		title,
		row.Description,
		row.Completed,
		row.DueDate,
		row.CreatedAt,
		row.UpdatedAt,
	), nil
}

// toRow flattens a domain entity for storage.
func toRow(t *domain.Todo) todoRow {
	return todoRow{
		ID:          t.ID().String(),
		Title:       t.Title().String(),
		Description: t.Description(),
		Completed:   t.IsCompleted(),
		DueDate:     t.DueDate(),
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   t.UpdatedAt(),
	}
}

// toDomainAll maps a set of rows, failing on the first corrupt one.
func toDomainAll(rows []todoRow) ([]*domain.Todo, error) {
	todos := make([]*domain.Todo, 0, len(rows))
	for _, row := range rows {
		todo, err := toDomain(row)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, nil
}
