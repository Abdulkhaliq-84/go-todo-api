//go:build e2e

// End-to-end tests: the real router, the real service, the real repository,
// the real database. The only thing faked is nothing.
//
// Run with: make test-e2e
//
// These are the slowest and most brittle tests you own, so keep them few and
// keep them about WORKFLOWS rather than edge cases. Edge cases belong in the
// unit tests below them. A good rule: if an e2e test fails, it should be
// because a wire came loose between layers -- every other kind of bug should
// have been caught earlier and faster.
//
//   domain tests       hundreds, microseconds, no dependencies
//   app tests          dozens, microseconds, fake repository
//   handler tests      dozens, microseconds, httptest
//   integration tests  a dozen, seconds, real Postgres
//   e2e tests          a handful, seconds, everything real
//
// That shape is the testing pyramid, and Go's build tags are what let you run
// each layer independently.

package e2e_test

import "testing"

// setupAPI boots the whole application against a test database and returns a
// base URL. This is essentially main.go's wiring, which is a good reason to
// keep that wiring short and boring.
//
// TODO(you): load test config, connect, migrate, build the handler chain, then
// httptest.NewServer(mux). Return the server so the test can Close() it.
func setupAPI(t *testing.T) string {
	t.Helper()
	return "" // TODO
}

// TestTodoLifecycle walks one todo through its entire life.
//
// This single test is worth more than a dozen narrow ones: it is the only place
// that proves the layers actually connect end to end.
func TestTodoLifecycle(t *testing.T) {
	t.Skip("TODO(you)")

	// TODO(you):
	//   1. POST   /api/v1/todos                -> 201, capture the id
	//   2. GET    /api/v1/todos/{id}           -> 200, completed == false
	//   3. PATCH  /api/v1/todos/{id}           -> 200, title changed
	//   4. POST   /api/v1/todos/{id}/complete  -> 200, completed == true
	//   5. GET    /api/v1/todos?completed=true -> contains this todo
	//   6. DELETE /api/v1/todos/{id}           -> 204
	//   7. GET    /api/v1/todos/{id}           -> 404
	//
	// Step 7 matters as much as step 6: it proves the delete actually happened
	// rather than merely returning a cheerful status code.
}

func TestHealthEndpoint(t *testing.T) {
	t.Skip("TODO(you)")

	// TODO(you): GET /health -> 200. Trivial, but it is what your container
	// orchestrator polls, so it is worth one test.
}
