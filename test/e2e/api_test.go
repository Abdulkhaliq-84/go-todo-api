//go:build e2e

// End-to-end tests: the real router, the real service, the real repository, the
// real database. Nothing is faked.
//
// Run with: make test-e2e
//
// These are the slowest and most brittle tests in the project, so they are few
// and they are about WORKFLOWS rather than edge cases. Edge cases belong in the
// layers below, where they run in microseconds. A good rule: if an e2e test
// fails, it should be because a wire came loose BETWEEN layers -- every other
// kind of bug should have been caught earlier and faster.
//
//	domain tests       23, microseconds, no dependencies
//	app tests          29, microseconds, fake repository
//	http tests         26, microseconds, httptest
//	config tests        9, microseconds, environment only
//	integration tests  14, seconds, real Postgres
//	e2e tests           a handful, seconds, everything real
//
// That shape is the testing pyramid, and Go's build tags are what let each
// level run independently.

package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/platform/docs"
	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/app"
	todohttp "github.com/Abdulkhaliq-84/go-todo-api/internal/todo/http"
	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultTestDB = "postgres://localhost:5432/todos_test?sslmode=disable"

// setupAPI boots the whole application against a test database and returns its
// base URL.
//
// This is essentially main.go's wiring, which is a good argument for keeping
// that wiring short and boring: anything clever in main.go would have to be
// duplicated here, and the duplicate would drift.
func setupAPI(t *testing.T) string {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = defaultTestDB
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connect to %s: %v", url, err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping %s: %v (is postgres running? make db-local)", url, err)
	}
	if _, err := pool.Exec(ctx, "TRUNCATE todos"); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mux := http.NewServeMux()
	todohttp.RegisterRoutes(mux, todohttp.NewServer(app.NewService(postgres.NewRepository(pool))), logger)
	docs.RegisterRoutes(mux)

	// httptest.NewServer listens on a real port on localhost, so these requests
	// go over an actual TCP connection -- unlike the handler tests, which call
	// the mux directly in process.
	srv := httptest.NewServer(mux)
	t.Cleanup(func() {
		srv.Close()
		pool.Close()
	})

	return srv.URL
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func request(t *testing.T, method, url, body string) (*http.Response, map[string]any) {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	t.Cleanup(func() { resp.Body.Close() })

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	if len(raw) == 0 {
		return resp, nil
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("%s %s returned non-JSON: %v\nbody: %s", method, url, err, raw)
	}
	return resp, decoded
}

func expectStatus(t *testing.T, resp *http.Response, want int, step string) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("%s: status = %d, want %d", step, resp.StatusCode, want)
	}
}

// ---------------------------------------------------------------------------
// The lifecycle
// ---------------------------------------------------------------------------

// TestTodoLifecycle walks one todo through its entire life.
//
// This single test is worth more than a dozen narrow ones: it is the only place
// that proves every layer actually connects. Each step below crosses all four --
// HTTP, application, domain, Postgres -- and back.
func TestTodoLifecycle(t *testing.T) {
	base := setupAPI(t) + "/api/v1/todos"

	// 1. CREATE
	resp, created := request(t, "POST", base,
		`{"title":"Ship the API","description":"all six stages","due_date":"2099-01-01T00:00:00Z"}`)
	expectStatus(t, resp, http.StatusCreated, "create")

	id, ok := created["id"].(string)
	if !ok || id == "" {
		t.Fatalf("create returned no id: %v", created)
	}
	if created["completed"] != false {
		t.Errorf("new todo completed = %v, want false", created["completed"])
	}

	// 2. READ -- proves the write actually reached Postgres and came back out
	resp, fetched := request(t, "GET", base+"/"+id, "")
	expectStatus(t, resp, http.StatusOK, "read")
	if fetched["title"] != "Ship the API" {
		t.Errorf("title = %v, want %q", fetched["title"], "Ship the API")
	}

	// 3. UPDATE -- title only; the due date must survive a PATCH that omits it
	resp, updated := request(t, "PATCH", base+"/"+id, `{"title":"Ship the API properly"}`)
	expectStatus(t, resp, http.StatusOK, "update")
	if updated["title"] != "Ship the API properly" {
		t.Errorf("title = %v, want the updated value", updated["title"])
	}
	if updated["due_date"] == nil {
		t.Error("an omitted due_date cleared the stored one")
	}

	// 4. COMPLETE
	resp, completed := request(t, "POST", base+"/"+id+"/complete", "")
	expectStatus(t, resp, http.StatusOK, "complete")
	if completed["completed"] != true {
		t.Errorf("completed = %v, want true", completed["completed"])
	}

	// 5. COMPLETE AGAIN -- the strict state machine, seen from outside
	resp, conflict := request(t, "POST", base+"/"+id+"/complete", "")
	expectStatus(t, resp, http.StatusConflict, "double complete")
	if conflict["error"] != "already_completed" {
		t.Errorf("error = %v, want %q", conflict["error"], "already_completed")
	}

	// 6. LIST with a filter -- proves the SQL WHERE clause agrees with the
	//    domain's own notion of completion
	resp, listed := request(t, "GET", base+"?completed=true", "")
	expectStatus(t, resp, http.StatusOK, "list completed")
	data, _ := listed["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("completed list has %d todos, want 1", len(data))
	}
	if first, _ := data[0].(map[string]any); first["id"] != id {
		t.Errorf("list returned the wrong todo: %v", first["id"])
	}

	// 7. REOPEN
	resp, reopened := request(t, "POST", base+"/"+id+"/reopen", "")
	expectStatus(t, resp, http.StatusOK, "reopen")
	if reopened["completed"] != false {
		t.Errorf("completed = %v, want false after reopen", reopened["completed"])
	}

	// 8. DELETE
	resp, _ = request(t, "DELETE", base+"/"+id, "")
	expectStatus(t, resp, http.StatusNoContent, "delete")

	// 9. GONE -- this step matters as much as the delete itself. It proves the
	//    row actually left the database rather than the endpoint merely
	//    returning a cheerful status code.
	resp, _ = request(t, "GET", base+"/"+id, "")
	expectStatus(t, resp, http.StatusNotFound, "read after delete")
}

// A todo past its deadline is overdue; completing it makes it not overdue, even
// though it was finished late. This is the one rule that spans the domain (the
// IsOverdue method) and Postgres (the WHERE clause), so it is worth proving
// once against both at the same time.
func TestOverdueLifecycle(t *testing.T) {
	base := setupAPI(t) + "/api/v1/todos"
	past := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339)

	resp, created := request(t, "POST", base, `{"title":"Late task","due_date":"`+past+`"}`)
	expectStatus(t, resp, http.StatusCreated, "create")
	id := created["id"].(string)

	if created["overdue"] != true {
		t.Errorf("overdue = %v, want true for a past due date", created["overdue"])
	}

	resp, listed := request(t, "GET", base+"?overdue=true", "")
	expectStatus(t, resp, http.StatusOK, "list overdue")
	if data, _ := listed["data"].([]any); len(data) != 1 {
		t.Fatalf("overdue list has %d todos, want 1", len(data))
	}

	resp, completed := request(t, "POST", base+"/"+id+"/complete", "")
	expectStatus(t, resp, http.StatusOK, "complete")
	if completed["overdue"] != false {
		t.Errorf("overdue = %v, want false once completed", completed["overdue"])
	}

	// And the database agrees with the domain: the SQL filter must exclude it
	// too, or the flag and the list would contradict each other.
	resp, afterList := request(t, "GET", base+"?overdue=true", "")
	expectStatus(t, resp, http.StatusOK, "list overdue after completing")
	if data, _ := afterList["data"].([]any); len(data) != 0 {
		t.Errorf("overdue list has %d todos, want 0 -- a completed todo is not overdue", len(data))
	}
}

// Pagination over a real database, through the real handler.
func TestPaginationAcrossPages(t *testing.T) {
	base := setupAPI(t) + "/api/v1/todos"

	for i := 0; i < 5; i++ {
		resp, _ := request(t, "POST", base, `{"title":"Task"}`)
		expectStatus(t, resp, http.StatusCreated, "seeding")
	}

	seen := map[string]bool{}
	for _, url := range []string{base + "?limit=2&offset=0", base + "?limit=2&offset=2", base + "?limit=2&offset=4"} {
		resp, page := request(t, "GET", url, "")
		expectStatus(t, resp, http.StatusOK, "list page")

		data, _ := page["data"].([]any)
		for _, item := range data {
			todo, _ := item.(map[string]any)
			id, _ := todo["id"].(string)
			if seen[id] {
				t.Errorf("todo %s appeared on two pages", id)
			}
			seen[id] = true
		}
	}

	if len(seen) != 5 {
		t.Errorf("paged through %d distinct todos, want 5", len(seen))
	}
}

// ---------------------------------------------------------------------------
// Operational endpoints
// ---------------------------------------------------------------------------

func TestHealthEndpoint(t *testing.T) {
	resp, body := request(t, "GET", setupAPI(t)+"/health", "")

	expectStatus(t, resp, http.StatusOK, "health")
	if body["status"] != "ok" {
		t.Errorf("status = %v, want %q", body["status"], "ok")
	}
}

// The spec is embedded in the binary and served from it, so this also proves
// the //go:embed in api/embed.go worked.
func TestDocsEndpoints(t *testing.T) {
	base := setupAPI(t)

	t.Run("openapi.yaml", func(t *testing.T) {
		resp, err := http.Get(base + "/openapi.yaml")
		if err != nil {
			t.Fatalf("GET /openapi.yaml: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		spec, _ := io.ReadAll(resp.Body)
		if !bytes.Contains(spec, []byte("openapi: 3.1.0")) {
			t.Error("served document does not look like the OpenAPI spec")
		}
	})

	t.Run("docs", func(t *testing.T) {
		resp, err := http.Get(base + "/docs")
		if err != nil {
			t.Fatalf("GET /docs: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		page, _ := io.ReadAll(resp.Body)
		if !bytes.Contains(page, []byte(`data-url="/openapi.yaml"`)) {
			t.Error("docs page does not point at the spec")
		}
	})
}
