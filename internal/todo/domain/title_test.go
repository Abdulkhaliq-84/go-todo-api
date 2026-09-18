package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// Table-driven tests are THE Go convention: a slice of cases, one loop, one
// t.Run per case. No parametrize decorator, no describe/it -- a for loop over a
// struct slice, which is why Go test files everywhere look nearly identical.

func TestNewTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{"simple title", "Buy milk", "Buy milk", nil},
		{"trims surrounding space", "  Buy milk  ", "Buy milk", nil},
		{"keeps interior space", "Buy  milk", "Buy  milk", nil},
		{"at max length", strings.Repeat("a", domain.TitleMaxLength), strings.Repeat("a", domain.TitleMaxLength), nil},
		// Counted in runes, not bytes: 200 multi-byte characters is 600 bytes
		// and must still be accepted, or the limit would disagree with the
		// VARCHAR(200) column for every non-ASCII user.
		{"200 multi-byte runes", strings.Repeat("é", 200), strings.Repeat("é", 200), nil},
		{"arabic title", "اشترِ الحليب", "اشترِ الحليب", nil},
		{"empty string", "", "", domain.ErrTitleEmpty},
		{"whitespace only", "   ", "", domain.ErrTitleEmpty},
		{"tabs and newlines only", "\t\n ", "", domain.ErrTitleEmpty},
		{"one over max", strings.Repeat("a", domain.TitleMaxLength+1), "", domain.ErrTitleTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewTitle(tt.input)

			// errors.Is walks the wrapped chain, so this still matches if an
			// implementation wraps the sentinel with extra context.
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewTitle(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if got.String() != tt.want {
				t.Errorf("NewTitle(%q) = %q, want %q", tt.input, got.String(), tt.want)
			}
		})
	}
}

func TestTitle_Equals(t *testing.T) {
	a := mustTitle(t, "Buy milk")
	b := mustTitle(t, "  Buy milk  ") // trimmed to the same text
	c := mustTitle(t, "Buy bread")

	if !a.Equals(b) {
		t.Error("titles with the same text after trimming should be equal")
	}
	if a.Equals(c) {
		t.Error("titles with different text should not be equal")
	}
}

// mustTitle builds a Title or fails the test. t.Helper() makes failure line
// numbers point at the CALLER rather than at this function.
func mustTitle(t *testing.T, s string) domain.Title {
	t.Helper()
	title, err := domain.NewTitle(s)
	if err != nil {
		t.Fatalf("NewTitle(%q): unexpected error %v", s, err)
	}
	return title
}
