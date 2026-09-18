//go:build integration

// The build tag above means `go test ./...` SKIPS this file entirely.
// It only compiles with `go test -tags=integration ./...`.
//
// Why: these tests need a real PostgreSQL. Unit tests must stay fast enough to
// run on every save; integration tests run when you ask for them, and in CI.
// Mixing the two is how test suites become slow enough that people stop running
// them. See the Makefile: `make test` is unit only, `make test-integration`
// runs these.
//
// The tag must be the first line, followed by a blank line, before `package`.

package postgres_test

import "testing"

// setupTestDB gives each test a clean database.
//
// >>> A DECISION FOR YOU. Three common approaches:
//
//  1. testcontainers-go -- spins up a throwaway Postgres container per test run.
//     Zero setup for anyone cloning the repo, works identically in CI, slower
//     to start (a few seconds). Most popular choice in Go today.
//
//  2. docker-compose Postgres + truncate tables between tests -- fast, but
//     requires `make db-up` first and tests can pollute each other if you
//     forget to clean up.
//
//  3. Transaction rollback per test -- start a tx, run the test inside it, roll
//     back. Very fast and perfectly isolated, but your repository must accept a
//     tx rather than a pool, which changes its signature.
//
// TODO(you): pick one and implement.
func setupTestDB(t *testing.T) any {
	t.Helper()
	return nil // TODO
}

func TestRepository_SaveAndFindByID(t *testing.T) {
	t.Skip("TODO(you)")

	// TODO(you): the round-trip test. Save a todo, load it back, assert every
	// field survived -- especially the due date.
	//
	// This is the test that catches TIMEZONE bugs, and it is the single most
	// valuable test in this file. Save a due date, read it back, and assert the
	// instants are equal. If you used TIMESTAMP instead of TIMESTAMPTZ, or
	// compared with == instead of .Equal(), this is where you find out --
	// rather than from a user in three months.
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	t.Skip("TODO(you)")

	// TODO(you): an unknown id must return domain.ErrNotFound -- NOT pgx.ErrNoRows.
	// This test is the guard on your infrastructure boundary: it proves driver
	// types are not leaking out of this package and into the domain.
}

func TestRepository_FindAll_Filters(t *testing.T) {
	t.Skip("TODO(you)")

	// TODO(you): seed a known set, then assert each domain.Filter combination
	// returns exactly the right subset. Include limit/offset -- pagination
	// off-by-ones are extremely common and invisible until production.
}

func TestRepository_Delete(t *testing.T) {
	t.Skip("TODO(you)")

	// TODO(you): deleting a missing row must return domain.ErrNotFound, which
	// means checking RowsAffected. A DELETE that matches nothing is NOT an
	// error as far as Postgres is concerned -- you have to notice yourself.
}
