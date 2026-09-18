package domain

// Title is a value object: it has no identity, it is immutable, and two
// titles with the same text ARE the same title.
//
// Why bother wrapping a string? Because `func UpdateTitle(s string)` accepts
// "", 5000 characters of spam, or a user's password pasted by mistake.
// `func UpdateTitle(t Title)` accepts only something that already passed
// NewTitle(). The type system carries the validation for you.
//
// This is the closest Go gets to a Pydantic constrained type -- except you
// write the check yourself and it runs exactly where you put it.
type Title struct {
	value string
}

// Title length bounds. Exported so the HTTP layer can mention them in error
// messages without hardcoding the numbers twice.
const (
	TitleMinLength = 1
	TitleMaxLength = 200
)

// NewTitle is the only way to build a Title. The zero value Title{} is
// technically constructible, which is a known Go wart -- the domain guards
// against it by never accepting a Title it did not receive from here.
//
// TODO(you): trim whitespace, enforce the bounds, return ErrTitleEmpty /
// ErrTitleTooLong from errors.go on failure.
func NewTitle(value string) (Title, error) {
	return Title{}, nil // TODO
}

// String makes Title satisfy fmt.Stringer, so it prints nicely and can be
// passed straight to the SQL driver later.
func (t Title) String() string { return t.value }

// Equals -- value objects compare by value, never by identity.
func (t Title) Equals(other Title) bool { return t.value == other.value }
