package domain_test

// Note the package name: domain_test, not domain.
//
// Go lets a test file live in either. Using the _test suffix package means this
// file can only touch the EXPORTED API -- exactly what a real caller sees. If
// your tests need private fields to be useful, that is a signal the public API
// is wrong.
//
// Table-driven tests are the Go convention: a slice of cases, one loop, one
// t.Run per case. No pytest.mark.parametrize, no describe/it -- just a for loop
// over a struct slice, which is why every Go codebase's tests look the same.
//
// TODO(you): write these once the domain is implemented. Start here, before the
// HTTP layer -- the domain is pure, has zero dependencies, and needs no
// database to test. That property is the whole point of the architecture.

// func TestNewTitle(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		input   string
// 		wantErr error
// 	}{
// 		{"valid title", "Buy milk", nil},
// 		{"empty string", "", domain.ErrTitleEmpty},
// 		{"whitespace only", "   ", domain.ErrTitleEmpty},
// 		{"too long", strings.Repeat("a", 201), domain.ErrTitleTooLong},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			_, err := domain.NewTitle(tt.input)
// 			if !errors.Is(err, tt.wantErr) {
// 				t.Errorf("got %v, want %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
