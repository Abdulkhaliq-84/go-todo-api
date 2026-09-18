// Package domain holds the Todo bounded context's business model.
//
// THE ONE RULE: this package imports only the standard library (plus a UUID
// generator, which is identity, not infrastructure). No pgx. No net/http. No
// encoding/json. If you ever want to put a `json:"..."` tag on a struct in
// here, that is the design telling you the transport layer needs its own DTO.
package domain

import "time"

// Todo is the aggregate root of this context.
//
// Every field is lowercase, so they are private to this package. Nothing
// outside domain/ can write `t.completed = true` and skip the rules. This is
// Go's encapsulation unit -- the package, not the struct.
//
// Coming from FastAPI: this is NOT a SQLModel class and NOT a Pydantic schema.
// It is a third, separate thing that knows only about todo-ness.
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
// TIME
//
// Every method that reads or writes a timestamp takes `now` as a parameter.
// Nothing in this package calls time.Now().
//
// The service calls time.Now() once per request and threads it down. Three
// things fall out of that:
//   - the domain is a pure function of its inputs, so tests state the instant
//     they mean rather than sleeping or freezing a clock
//   - every todo in a list is judged against the SAME instant
//   - "a failed transition must not move updatedAt" becomes an exact
//     assertion rather than a timing-dependent one
//
// All stored times are normalised to UTC on the way in. Postgres TIMESTAMPTZ
// stores UTC regardless, so normalising here means the value a test sees and
// the value the database returns are the same value.
// ---------------------------------------------------------------------------

// New creates a valid Todo. There is deliberately no other way to build one
// from outside this package, so an invalid Todo cannot exist.
//
// A due date in the past is accepted -- see errors.go for why.
func New(title Title, description string, dueDate *time.Time, now time.Time) (*Todo, error) {
	if title.value == "" {
		return nil, ErrTitleEmpty
	}

	ts := now.UTC()
	t := &Todo{
		id:          NewID(),
		title:       title,
		description: description,
		completed:   false,
		createdAt:   ts,
		updatedAt:   ts,
	}
	t.setDueDate(dueDate)
	return t, nil
}

// Reconstitute rebuilds a Todo from stored state.
//
// Why this exists: New applies CREATION rules -- it mints a fresh ID and sets
// createdAt to now. The repository is loading a row that already passed those
// rules, possibly years ago, and must not re-run them. This is the standard DDD
// escape hatch for persistence, and it is what lets the repository live outside
// this package without needing access to private fields.
func Reconstitute(
	id ID,
	title Title,
	description string,
	completed bool,
	dueDate *time.Time,
	createdAt, updatedAt time.Time,
) *Todo {
	t := &Todo{
		id:          id,
		title:       title,
		description: description,
		completed:   completed,
		createdAt:   createdAt.UTC(),
		updatedAt:   updatedAt.UTC(),
	}
	t.setDueDate(dueDate)
	return t
}

// ---------------------------------------------------------------------------
// STATE MACHINE
//
//	        Complete()                  Reopen()
//	 ACTIVE ──────────▶ COMPLETED   COMPLETED ──────────▶ ACTIVE
//	 (completed=false)  (=true)      (=true)              (=false)
//
// Both transitions are STRICT: calling one from the wrong state is an error,
// not a silent no-op. On the error path NOTHING changes,
// updatedAt included -- a failed transition that still bumped a timestamp would
// be a partial mutation, and preventing those is what an aggregate is for.
// ---------------------------------------------------------------------------

// Complete transitions ACTIVE -> COMPLETED.
//
// The due date is irrelevant here: an overdue todo can still be completed.
// Lateness is an observation about time, never a permission check.
func (t *Todo) Complete(now time.Time) error {
	if t.completed {
		return ErrAlreadyComplete
	}
	t.completed = true
	t.touch(now)
	return nil
}

// Reopen transitions COMPLETED -> ACTIVE. The exact mirror of Complete.
//
// Deliberately does not touch dueDate. A todo reopened after its deadline is
// immediately overdue again, which is truthful. Clearing the date to be kind
// would be the domain inventing a rule nobody asked for -- Reschedule exists
// for that, visibly.
func (t *Todo) Reopen(now time.Time) error {
	if !t.completed {
		return ErrNotCompleted
	}
	t.completed = false
	t.touch(now)
	return nil
}

// ---------------------------------------------------------------------------
// Mutators
//
// These return an error they currently never produce. That is a deliberate
// choice for the aggregate's mutating API: adding a rule later (say, a
// completed todo may not be renamed) then costs no caller a signature change.
// Elsewhere in Go, returning an error you cannot produce is noise -- here the
// uniformity across an aggregate's methods is worth more.
// ---------------------------------------------------------------------------

// UpdateTitle replaces the title. Takes an already-validated Title, so this
// method cannot receive garbage.
func (t *Todo) UpdateTitle(title Title, now time.Time) error {
	if title.Equals(t.title) {
		return nil // no change, so no timestamp bump
	}
	t.title = title
	t.touch(now)
	return nil
}

// UpdateDescription replaces the free-text description.
func (t *Todo) UpdateDescription(description string, now time.Time) error {
	if description == t.description {
		return nil
	}
	t.description = description
	t.touch(now)
	return nil
}

// Reschedule changes or clears the due date. Pass nil to clear it.
func (t *Todo) Reschedule(dueDate *time.Time, now time.Time) error {
	t.setDueDate(dueDate)
	t.touch(now)
	return nil
}

// IsOverdue reports whether this todo needs attention as of `now`.
//
// A COMPLETED todo is never overdue. "Overdue" here
// means outstanding and past its deadline, so ?overdue=true returns exactly the
// list a user should act on, with no client-side filtering.
func (t *Todo) IsOverdue(now time.Time) bool {
	if t.completed || t.dueDate == nil {
		return false
	}
	return now.UTC().After(*t.dueDate)
}

// ---------------------------------------------------------------------------
// Getters
//
// Go has no `public readonly` and no @property. Private field plus exported
// getter is the whole mechanism: the outside world can read state and cannot
// write it.
// ---------------------------------------------------------------------------

func (t *Todo) ID() ID               { return t.id }
func (t *Todo) Title() Title         { return t.title }
func (t *Todo) Description() string  { return t.description }
func (t *Todo) IsCompleted() bool    { return t.completed }
func (t *Todo) CreatedAt() time.Time { return t.createdAt }
func (t *Todo) UpdatedAt() time.Time { return t.updatedAt }

// DueDate returns a COPY, not the stored pointer.
//
// Without the copy a caller could write *todo.DueDate() = someOtherTime and
// mutate the aggregate from outside, defeating every private field above. This
// is the pointer-shaped hole in Go's encapsulation, and copying is the patch.
func (t *Todo) DueDate() *time.Time {
	if t.dueDate == nil {
		return nil
	}
	d := *t.dueDate
	return &d
}

// ---------------------------------------------------------------------------
// internals
// ---------------------------------------------------------------------------

// setDueDate stores a normalised COPY of the caller's time, for the same reason
// DueDate hands one out: a stored pointer the caller still holds is a field the
// caller can still change.
func (t *Todo) setDueDate(dueDate *time.Time) {
	if dueDate == nil {
		t.dueDate = nil
		return
	}
	d := dueDate.UTC()
	t.dueDate = &d
}

func (t *Todo) touch(now time.Time) { t.updatedAt = now.UTC() }
