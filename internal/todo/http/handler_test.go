package http_test

// Handler tests use net/http/httptest, which is in the STANDARD LIBRARY.
//
// httptest.NewRecorder() is a fake ResponseWriter that captures what a handler
// wrote; httptest.NewRequest() builds a request without a network. So these
// tests exercise real routing, real JSON encoding and real status codes, in
// process, in microseconds. No supertest, no TestClient, no running server.
//
// What belongs here: TRANSLATION correctness.
//   - a malformed body returns 400, not 500
//   - domain.ErrNotFound becomes 404
//   - the JSON response has the field names the API contract promises
//
// What does NOT belong here: business rules, again. Test those once, in domain.

import (
	"net/http/httptest"
	"testing"
)

func TestHandler_Create(t *testing.T) {
	t.Skip("TODO(you): remove once the handler is implemented")

	_ = httptest.NewRecorder

	// TODO(you): POST a valid JSON body, assert 201 and that the response body
	// decodes into the expected shape.
}

func TestHandler_Create_MalformedJSON(t *testing.T) {
	t.Skip("TODO(you)")

	// TODO(you): send `{"title":` -- broken JSON. Must be 400, never 500.
	// A 500 here means an error escaped instead of being handled.
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	t.Skip("TODO(you)")

	// TODO(you): this is the test that proves writeError's mapping works.
	// domain.ErrNotFound in, 404 out.
}

func TestHandler_ErrorStatusMapping(t *testing.T) {
	t.Skip("TODO(you)")

	// TODO(you): table-driven over every domain error -> expected status.
	// One table here documents your whole error contract in a readable place,
	// and fails loudly the day someone adds an error and forgets to map it.
	//
	//   domain.ErrNotFound         -> 404
	//   domain.ErrTitleEmpty       -> 400
	//   domain.ErrInvalidID        -> 400
	//   domain.ErrAlreadyComplete  -> 409
	//   errors.New("boom")         -> 500
}

// TODO(you): TestHandler_Create_LeaksNoInternalError -- when the repository
// fails with a database error, the 500 response body must NOT contain the
// driver's message. Leaking "pq: relation todos does not exist" tells an
// attacker your schema. Log it server-side; return something generic.
