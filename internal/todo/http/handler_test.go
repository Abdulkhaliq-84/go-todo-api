package http_test

// Handler tests use net/http/httptest, which is in the STANDARD LIBRARY.
//
// httptest.NewRecorder() is a fake ResponseWriter that captures what a handler
// wrote; httptest.NewRequest() builds a request without a network. So these
// tests exercise real routing, real JSON decoding and real status codes, in
// process, in microseconds. No supertest, no TestClient, no running server.
//
// What belongs here: TRANSLATION correctness.
//   - malformed input returns 400, never 500
//   - domain.ErrNotFound becomes 404, ErrAlreadyComplete becomes 409
//   - the JSON response carries the field names the contract promises
//   - a 500 body never leaks an internal error message
//
// What does NOT belong here: business rules. Those are tested once, in
// domain/todo_test.go.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/app"
	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/domain"
	todohttp "github.com/Abdulkhaliq-84/go-todo-api/internal/todo/http"
)

// ---------------------------------------------------------------------------
// test harness
// ---------------------------------------------------------------------------

// memRepo is a minimal in-memory repository. The app package has its own fake,
// but a _test package cannot be imported from another package's tests -- Go
// deliberately keeps test helpers from leaking between packages.
type memRepo struct {
	mu    sync.RWMutex
	items map[string]*domain.Todo
	fail  error // when set, every method returns it
}

func newMemRepo() *memRepo {
	return &memRepo{items: make(map[string]*domain.Todo)}
}

var _ domain.Repository = (*memRepo)(nil)

func (m *memRepo) Save(ctx context.Context, todo *domain.Todo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fail != nil {
		return m.fail
	}
	m.items[todo.ID().String()] = todo
	return nil
}

func (m *memRepo) FindByID(ctx context.Context, id domain.ID) (*domain.Todo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.fail != nil {
		return nil, m.fail
	}
	todo, ok := m.items[id.String()]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return todo, nil
}

func (m *memRepo) FindAll(ctx context.Context, filter domain.Filter) ([]*domain.Todo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.fail != nil {
		return nil, m.fail
	}
	todos := make([]*domain.Todo, 0, len(m.items))
	for _, todo := range m.items {
		todos = append(todos, todo)
	}
	return todos, nil
}

func (m *memRepo) Delete(ctx context.Context, id domain.ID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fail != nil {
		return m.fail
	}
	if _, ok := m.items[id.String()]; !ok {
		return domain.ErrNotFound
	}
	delete(m.items, id.String())
	return nil
}

// newTestServer wires the real router, the real service and a fake repository,
// then returns the mux. Everything above the repository is production code.
func newTestServer(t *testing.T) (*nethttp.ServeMux, *memRepo) {
	t.Helper()

	repo := newMemRepo()
	service := app.NewService(repo)
	srv := todohttp.NewServer(service)

	// Discard logs so a deliberately failing test does not print a scary error.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mux := nethttp.NewServeMux()
	todohttp.RegisterRoutes(mux, srv, logger)
	return mux, repo
}

// do issues a request against the mux and returns the recorder.
func do(t *testing.T, mux *nethttp.ServeMux, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// decode unmarshals a response body, failing the test on bad JSON.
func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decoding response: %v\nbody was: %s", err, rec.Body.String())
	}
	return out
}

// createTodo is a fixture: POSTs a todo and returns its id.
func createTodo(t *testing.T, mux *nethttp.ServeMux, title string) string {
	t.Helper()
	rec := do(t, mux, "POST", "/api/v1/todos", `{"title":"`+title+`"}`)
	if rec.Code != nethttp.StatusCreated {
		t.Fatalf("fixture create failed: %d %s", rec.Code, rec.Body.String())
	}
	return decode[map[string]any](t, rec)["id"].(string)
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestHandler_Create(t *testing.T) {
	mux, _ := newTestServer(t)

	rec := do(t, mux, "POST", "/api/v1/todos",
		`{"title":"Learn Go","description":"interfaces","due_date":"2099-01-01T00:00:00Z"}`)

	if rec.Code != nethttp.StatusCreated {
		t.Fatalf("status = %d, want 201\nbody: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	body := decode[map[string]any](t, rec)

	// Field names are the CONTRACT. Asserting on them here is what catches a
	// rename that would silently break every client.
	for _, field := range []string{"id", "title", "description", "completed", "overdue", "created_at", "updated_at"} {
		if _, ok := body[field]; !ok {
			t.Errorf("response is missing field %q\nbody: %s", field, rec.Body.String())
		}
	}
	if body["title"] != "Learn Go" {
		t.Errorf("title = %v, want %q", body["title"], "Learn Go")
	}
	if body["completed"] != false {
		t.Errorf("completed = %v, want false", body["completed"])
	}
}

// Broken JSON must be a 400. A 500 here would mean an error escaped instead of
// being handled.
func TestHandler_Create_MalformedJSON(t *testing.T) {
	mux, _ := newTestServer(t)

	rec := do(t, mux, "POST", "/api/v1/todos", `{"title":`)

	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("status = %d, want 400\nbody: %s", rec.Code, rec.Body.String())
	}
}

// A domain rule violation must surface as a 400 with the coarse code, not as a
// generic failure.
func TestHandler_Create_EmptyTitle(t *testing.T) {
	mux, _ := newTestServer(t)

	rec := do(t, mux, "POST", "/api/v1/todos", `{"title":"   "}`)

	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("status = %d, want 400\nbody: %s", rec.Code, rec.Body.String())
	}
	body := decode[map[string]any](t, rec)
	if body["error"] != "invalid_title" {
		t.Errorf("error = %v, want %q", body["error"], "invalid_title")
	}
}

func TestHandler_Create_TitleTooLong(t *testing.T) {
	mux, _ := newTestServer(t)

	rec := do(t, mux, "POST", "/api/v1/todos",
		`{"title":"`+strings.Repeat("a", domain.TitleMaxLength+1)+`"}`)

	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	// Coarse codes: empty and too-long share one code, and the message carries
	// the difference.
	body := decode[map[string]any](t, rec)
	if body["error"] != "invalid_title" {
		t.Errorf("error = %v, want %q", body["error"], "invalid_title")
	}
	if msg, _ := body["message"].(string); !strings.Contains(msg, "200") {
		t.Errorf("message = %q, want it to mention the 200 character limit", msg)
	}
}

// ---------------------------------------------------------------------------
// Read
// ---------------------------------------------------------------------------

func TestHandler_GetByID(t *testing.T) {
	mux, _ := newTestServer(t)
	id := createTodo(t, mux, "Task")

	rec := do(t, mux, "GET", "/api/v1/todos/"+id, "")

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200\nbody: %s", rec.Code, rec.Body.String())
	}
	if got := decode[map[string]any](t, rec)["id"]; got != id {
		t.Errorf("id = %v, want %v", got, id)
	}
}

// This is the test that proves classify()'s mapping works: domain.ErrNotFound
// in, 404 out.
func TestHandler_GetByID_NotFound(t *testing.T) {
	mux, _ := newTestServer(t)

	rec := do(t, mux, "GET", "/api/v1/todos/"+domain.NewID().String(), "")

	if rec.Code != nethttp.StatusNotFound {
		t.Fatalf("status = %d, want 404\nbody: %s", rec.Code, rec.Body.String())
	}
	if got := decode[map[string]any](t, rec)["error"]; got != "not_found" {
		t.Errorf("error = %v, want %q", got, "not_found")
	}
}

// A malformed id is a 400 (you sent nonsense), not a 404 (it is not here).
// The generated code rejects it before the handler runs.
func TestHandler_GetByID_MalformedID(t *testing.T) {
	mux, _ := newTestServer(t)

	rec := do(t, mux, "GET", "/api/v1/todos/not-a-uuid", "")

	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("status = %d, want 400\nbody: %s", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestHandler_List(t *testing.T) {
	mux, _ := newTestServer(t)
	createTodo(t, mux, "One")
	createTodo(t, mux, "Two")

	rec := do(t, mux, "GET", "/api/v1/todos", "")

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := decode[struct {
		Data  []map[string]any `json:"data"`
		Count int              `json:"count"`
	}](t, rec)

	if body.Count != 2 || len(body.Data) != 2 {
		t.Errorf("count = %d, len(data) = %d, want 2 and 2", body.Count, len(body.Data))
	}
}

// An empty collection must serialise as [] rather than null -- clients calling
// data.map(...) break on null.
func TestHandler_List_EmptyIsArrayNotNull(t *testing.T) {
	mux, _ := newTestServer(t)

	rec := do(t, mux, "GET", "/api/v1/todos", "")

	if !strings.Contains(rec.Body.String(), `"data":[]`) {
		t.Errorf("empty list body = %s, want \"data\":[]", rec.Body.String())
	}
}

func TestHandler_List_RejectsBadQueryParam(t *testing.T) {
	mux, _ := newTestServer(t)

	rec := do(t, mux, "GET", "/api/v1/todos?completed=maybe", "")

	if rec.Code != nethttp.StatusBadRequest {
		t.Errorf("status = %d, want 400 for an unparseable boolean", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Update -- the three-state due_date, end to end over the wire
// ---------------------------------------------------------------------------

func TestHandler_Update_DueDateThreeStates(t *testing.T) {
	const withDue = `{"title":"Task","due_date":"2099-01-01T00:00:00Z"}`

	t.Run("omitting due_date leaves it alone", func(t *testing.T) {
		mux, _ := newTestServer(t)
		rec := do(t, mux, "POST", "/api/v1/todos", withDue)
		id := decode[map[string]any](t, rec)["id"].(string)

		rec = do(t, mux, "PATCH", "/api/v1/todos/"+id, `{"title":"Renamed"}`)

		if rec.Code != nethttp.StatusOK {
			t.Fatalf("status = %d, want 200\nbody: %s", rec.Code, rec.Body.String())
		}
		// THIS is the bug nullable-type exists to prevent. With a plain
		// *time.Time, an omitted field and an explicit null both arrive as nil
		// and this PATCH would silently wipe the stored date.
		if got := decode[map[string]any](t, rec)["due_date"]; got == nil {
			t.Error("an omitted due_date cleared the stored one")
		}
	})

	t.Run("explicit null clears it", func(t *testing.T) {
		mux, _ := newTestServer(t)
		rec := do(t, mux, "POST", "/api/v1/todos", withDue)
		id := decode[map[string]any](t, rec)["id"].(string)

		rec = do(t, mux, "PATCH", "/api/v1/todos/"+id, `{"due_date":null}`)

		if rec.Code != nethttp.StatusOK {
			t.Fatalf("status = %d, want 200\nbody: %s", rec.Code, rec.Body.String())
		}
		if got := decode[map[string]any](t, rec)["due_date"]; got != nil {
			t.Errorf("due_date = %v, want null after an explicit clear", got)
		}
	})

	t.Run("a value sets it", func(t *testing.T) {
		mux, _ := newTestServer(t)
		id := createTodo(t, mux, "Task")

		rec := do(t, mux, "PATCH", "/api/v1/todos/"+id, `{"due_date":"2099-06-01T00:00:00Z"}`)

		if rec.Code != nethttp.StatusOK {
			t.Fatalf("status = %d, want 200\nbody: %s", rec.Code, rec.Body.String())
		}
		got, _ := decode[map[string]any](t, rec)["due_date"].(string)
		if !strings.HasPrefix(got, "2099-06-01") {
			t.Errorf("due_date = %q, want it set to 2099-06-01", got)
		}
	})
}

func TestHandler_Update_NotFound(t *testing.T) {
	mux, _ := newTestServer(t)

	rec := do(t, mux, "PATCH", "/api/v1/todos/"+domain.NewID().String(), `{"title":"X"}`)

	if rec.Code != nethttp.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Complete / Reopen -- the 409 the spec promises
// ---------------------------------------------------------------------------

func TestHandler_Complete(t *testing.T) {
	mux, _ := newTestServer(t)
	id := createTodo(t, mux, "Task")

	rec := do(t, mux, "POST", "/api/v1/todos/"+id+"/complete", "")

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200\nbody: %s", rec.Code, rec.Body.String())
	}
	if got := decode[map[string]any](t, rec)["completed"]; got != true {
		t.Errorf("completed = %v, want true", got)
	}
}

// The strict state machine, visible at the edge of the system: a second
// complete is a 409, exactly as api/openapi.yaml documents.
func TestHandler_Complete_Twice(t *testing.T) {
	mux, _ := newTestServer(t)
	id := createTodo(t, mux, "Task")

	if rec := do(t, mux, "POST", "/api/v1/todos/"+id+"/complete", ""); rec.Code != nethttp.StatusOK {
		t.Fatalf("first complete: status = %d", rec.Code)
	}

	rec := do(t, mux, "POST", "/api/v1/todos/"+id+"/complete", "")

	if rec.Code != nethttp.StatusConflict {
		t.Fatalf("status = %d, want 409\nbody: %s", rec.Code, rec.Body.String())
	}
	if got := decode[map[string]any](t, rec)["error"]; got != "already_completed" {
		t.Errorf("error = %v, want %q", got, "already_completed")
	}
}

func TestHandler_Reopen_NotCompleted(t *testing.T) {
	mux, _ := newTestServer(t)
	id := createTodo(t, mux, "Task")

	rec := do(t, mux, "POST", "/api/v1/todos/"+id+"/reopen", "")

	if rec.Code != nethttp.StatusConflict {
		t.Fatalf("status = %d, want 409\nbody: %s", rec.Code, rec.Body.String())
	}
	if got := decode[map[string]any](t, rec)["error"]; got != "not_completed" {
		t.Errorf("error = %v, want %q", got, "not_completed")
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestHandler_Delete(t *testing.T) {
	mux, _ := newTestServer(t)
	id := createTodo(t, mux, "Doomed")

	rec := do(t, mux, "DELETE", "/api/v1/todos/"+id, "")

	if rec.Code != nethttp.StatusNoContent {
		t.Fatalf("status = %d, want 204\nbody: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Errorf("204 response has a body: %q", rec.Body.String())
	}

	// Deleting again must report not-found, not succeed silently.
	if rec := do(t, mux, "DELETE", "/api/v1/todos/"+id, ""); rec.Code != nethttp.StatusNotFound {
		t.Errorf("second delete status = %d, want 404", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

func TestHandler_Health(t *testing.T) {
	mux, _ := newTestServer(t)

	rec := do(t, mux, "GET", "/health", "")

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := decode[map[string]any](t, rec)["status"]; got != "ok" {
		t.Errorf("status = %v, want %q", got, "ok")
	}
}

// ---------------------------------------------------------------------------
// The leak test
// ---------------------------------------------------------------------------

// When the repository fails with a database error, the 500 response must NOT
// contain the driver's message. Leaking `pq: relation "todos" does not exist`
// hands an attacker the schema.
//
// This is not a hypothetical: the generated NewStrictHandler's DEFAULT
// ResponseErrorHandlerFunc is http.Error(w, err.Error(), 500), which does
// exactly that. router.go replaces it, and this test is what holds that
// replacement in place.
func TestHandler_InternalErrorDoesNotLeak(t *testing.T) {
	mux, repo := newTestServer(t)
	repo.fail = errors.New(`pq: relation "todos" does not exist`)

	rec := do(t, mux, "POST", "/api/v1/todos", `{"title":"Task"}`)

	if rec.Code != nethttp.StatusInternalServerError {
		t.Fatalf("status = %d, want 500\nbody: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if strings.Contains(body, "relation") || strings.Contains(body, "pq:") {
		t.Errorf("500 response leaked the internal error: %s", body)
	}

	parsed := decode[map[string]any](t, rec)
	if parsed["error"] != "internal_error" {
		t.Errorf("error = %v, want %q", parsed["error"], "internal_error")
	}
}

// Every error path must answer in the same JSON shape, so a client writes one
// error handler rather than one per status code.
func TestHandler_ErrorShapeIsConsistent(t *testing.T) {
	mux, _ := newTestServer(t)
	id := createTodo(t, mux, "Task")
	do(t, mux, "POST", "/api/v1/todos/"+id+"/complete", "")

	cases := []struct {
		name         string
		method, path string
		body         string
		wantStatus   int
	}{
		{"400 invalid title", "POST", "/api/v1/todos", `{"title":""}`, nethttp.StatusBadRequest},
		{"404 unknown id", "GET", "/api/v1/todos/" + domain.NewID().String(), "", nethttp.StatusNotFound},
		{"409 already completed", "POST", "/api/v1/todos/" + id + "/complete", "", nethttp.StatusConflict},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(t, mux, tc.method, tc.path, tc.body)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d\nbody: %s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			body := decode[map[string]any](t, rec)
			for _, field := range []string{"error", "message"} {
				if v, ok := body[field]; !ok || v == "" {
					t.Errorf("error response missing %q: %s", field, rec.Body.String())
				}
			}
		})
	}
}

// Guards against a timestamp regression: the API must emit RFC 3339 in UTC, not
// Go's native time.Time formatting.
func TestHandler_TimestampsAreRFC3339(t *testing.T) {
	mux, _ := newTestServer(t)
	rec := do(t, mux, "POST", "/api/v1/todos", `{"title":"Task"}`)

	created, _ := decode[map[string]any](t, rec)["created_at"].(string)
	if _, err := time.Parse(time.RFC3339, created); err != nil {
		t.Errorf("created_at = %q, which is not RFC 3339: %v", created, err)
	}
}
