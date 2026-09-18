// Package domain holds the Todo bounded context's business model.
//
// THE ONE RULE: this package imports ONLY the standard library.
// No pgx. No net/http. No encoding/json. If you ever feel the urge to add
// a `json:"..."` tag to a struct in here, that is the design telling you
// the transport layer needs its own DTO instead.
package domain

import "time"

// Todo is the aggregate root of this context.
//
// Note the lowercase fields: they are private to this package. Nothing
// outside domain/ can do `t.completed = true` and skip the business rules.
// This is Go's encapsulation unit -- the package, not the struct.
//
// Coming from FastAPI: this is NOT your SQLModel class and NOT your Pydantic
// schema. It is a third, separate thing that knows only about todo-ness.
type Todo struct {
	id          ID
	title       Title
	description string
	completed   bool
	dueDate     *time.Time // nil = no due date. Pointer is how Go says "optional".
	createdAt   time.Time
	updatedAt   time.Time
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// New creates a valid Todo or fails. There is deliberately no other way to
// build one from outside this package, so an invalid Todo cannot exist.
//
// TODO(you): construct the Todo, set timestamps, return it.
func New(title Title, description string, dueDate *time.Time) (*Todo, error) {
	return nil, nil // TODO
}

// Reconstitute rebuilds a Todo from stored state.
//
// Why this exists: New() applies creation rules (fresh ID, createdAt = now).
// The Postgres repository is loading a row that ALREADY passed those rules
// years ago -- it must not re-run them. This is the standard DDD escape hatch
// for the persistence layer, and it is why the repository can live outside
// this package without needing access to private fields.
//
// TODO(you): populate every field verbatim from the arguments.
func Reconstitute(
	id ID,
	title Title,
	description string,
	completed bool,
	dueDate *time.Time,
	createdAt, updatedAt time.Time,
) *Todo {
	return nil // TODO
}

// ---------------------------------------------------------------------------
// Behaviour -- the business rules live HERE, not in the service, not in the handler
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// STATE MACHINE
//
// A Todo has exactly two states and two legal transitions:
//
//	        Complete()                  Reopen()
//	 ACTIVE ──────────▶ COMPLETED   COMPLETED ──────────▶ ACTIVE
//	 (completed=false)  (=true)      (=true)              (=false)
//
// Both transitions are STRICT: calling one from the wrong state is an error,
// not a silent no-op. Decided deliberately -- see docs/DECISIONS.md #1.
//
// Consequences that ripple outward, all of them already in place:
//   - api/openapi.yaml promises 409 on both endpoints
//   - http/errors.go maps ErrAlreadyComplete / ErrNotCompleted to 409
//   - app/service.go must let those errors propagate UNWRAPPED, or errors.Is
//     stops matching and the client silently starts receiving 500s
// ---------------------------------------------------------------------------

// Complete transitions the todo from ACTIVE to COMPLETED.
//
// Contract:
//
//	Precondition   t.completed == false
//	Postcondition  t.completed == true, t.updatedAt refreshed
//	Error          ErrAlreadyComplete if already completed
//
// On the error path NOTHING may change -- not completed, and not updatedAt.
// A failed transition that still bumped a timestamp would be a partial
// mutation, and the aggregate's job is to make those impossible.
//
// Note the due date is irrelevant here. An overdue todo can still be completed;
// lateness is an observation about time, never a permission check.
//
// TODO(you): implement to this contract.
func (t *Todo) Complete() error {
	return nil // TODO
}

// Reopen transitions the todo from COMPLETED back to ACTIVE.
//
// Contract -- the exact mirror of Complete:
//
//	Precondition   t.completed == true
//	Postcondition  t.completed == false, t.updatedAt refreshed
//	Error          ErrNotCompleted if it was never completed
//
// The mirroring matters. If Complete is strict and Reopen is forgiving, the two
// halves of the state machine disagree, and which half a bug lands in becomes
// a coin flip.
//
// Deliberately NOT part of this contract: reopening does not touch dueDate. A
// todo reopened after its deadline is immediately overdue again, which is
// truthful. Clearing the date to be kind would be the domain quietly inventing
// a rule nobody asked for -- if rescheduling is wanted, the caller calls
// Reschedule, visibly.
//
// TODO(you): implement to this contract.
func (t *Todo) Reopen() error {
	return nil // TODO
}

// UpdateTitle replaces the title. Takes a Title (already-validated value
// object), not a string -- so this method cannot receive garbage.
//
// TODO(you): assign and touch updatedAt.
func (t *Todo) UpdateTitle(title Title) error {
	return nil // TODO
}

// UpdateDescription replaces the free-text description.
//
// TODO(you)
func (t *Todo) UpdateDescription(description string) error {
	return nil // TODO
}

// Reschedule changes or clears the due date. Pass nil to clear it.
//
// TODO(you): consider whether a due date in the past should be rejected.
func (t *Todo) Reschedule(dueDate *time.Time) error {
	return nil // TODO
}

// IsOverdue is a derived property -- computed, never stored.
//
// TODO(you)
func (t *Todo) IsOverdue(now time.Time) bool {
	return false // TODO
}

// ---------------------------------------------------------------------------
// Getters
//
// Go has no `public readonly` and no @property. Private field + exported
// getter method is the whole mechanism. Verbose, but it means the outside
// world can read state and cannot write it.
// ---------------------------------------------------------------------------

func (t *Todo) ID() ID                { return t.id }
func (t *Todo) Title() Title          { return t.title }
func (t *Todo) Description() string   { return t.description }
func (t *Todo) IsCompleted() bool     { return t.completed }
func (t *Todo) DueDate() *time.Time   { return t.dueDate }
func (t *Todo) CreatedAt() time.Time  { return t.createdAt }
func (t *Todo) UpdatedAt() time.Time  { return t.updatedAt }
