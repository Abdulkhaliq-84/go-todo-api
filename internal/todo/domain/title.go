package domain

import (
	"strings"
	"unicode/utf8"
)

// Title is a value object: no identity, immutable, and two titles with the same
// text ARE the same title.
//
// Why wrap a string? Because func UpdateTitle(s string) accepts "", 5000
// characters of spam, or a password pasted by mistake. func UpdateTitle(t Title)
// accepts only something that already passed NewTitle. The type carries the
// validation, so it cannot be forgotten at a call site.
type Title struct {
	value string
}

// Title length bounds, in RUNES. Exported so the transport layer can mention
// them in a message without hardcoding the numbers a second time.
const (
	TitleMinLength = 1
	TitleMaxLength = 200
)

// NewTitle is the only way to build a Title.
//
// Length is counted in runes, not bytes, to match the database: Postgres
// VARCHAR(200) counts characters too. len("café") is 5 bytes but 4 characters --
// validating in bytes here would reject titles the column would have accepted,
// and the two limits would disagree for every non-ASCII user.
func NewTitle(value string) (Title, error) {
	trimmed := strings.TrimSpace(value)

	if utf8.RuneCountInString(trimmed) < TitleMinLength {
		return Title{}, ErrTitleEmpty
	}
	if utf8.RuneCountInString(trimmed) > TitleMaxLength {
		return Title{}, ErrTitleTooLong
	}
	return Title{value: trimmed}, nil
}

// String makes Title satisfy fmt.Stringer.
func (t Title) String() string { return t.value }

// Equals -- value objects compare by value, never by identity.
func (t Title) Equals(other Title) bool { return t.value == other.value }
