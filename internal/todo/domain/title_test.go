package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
)

// Table-driven tests are THE Go convention. A slice of cases, one loop, one
// t.Run per case. There is no parametrize decorator and no describe/it -- it is
// a for loop over a struct slice, which is why Go test files across every
// codebase you will ever read look nearly identical.
//
// t.Run creates a subtest with its own name, so a failure reports as
// TestNewTitle/whitespace_only rather than just "TestNewTitle failed".

func TestNewTitle(t *testing.T) {
	t.Skip("TODO(you): remove this line once domain.NewTitle is implemented")

	tests := []struct {
		name    string
		input   string
		want    string // expected normalised value
		wantErr error  // nil means success expected
	}{
		{"simple title", "Buy milk", "Buy milk", nil},
		{"trims surrounding space", "  Buy milk  ", "Buy milk", nil},
		{"at max length", strings.Repeat("a", domain.TitleMaxLength), strings.Repeat("a", domain.TitleMaxLength), nil},
		{"empty string", "", "", domain.ErrTitleEmpty},
		{"whitespace only", "   ", "", domain.ErrTitleEmpty},
		{"one over max", strings.Repeat("a", domain.TitleMaxLength+1), "", domain.ErrTitleTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewTitle(tt.input)

			// errors.Is walks the wrapped chain, so this still matches even if
			// the implementation wraps the sentinel with extra context.
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewTitle(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return // nothing more to check on the failure path
			}
			if got.String() != tt.want {
				t.Errorf("NewTitle(%q) = %q, want %q", tt.input, got.String(), tt.want)
			}
		})
	}
}

// TODO(you): add TestTitle_Equals -- value objects compare by value, so two
// Titles built from the same string must be equal.
