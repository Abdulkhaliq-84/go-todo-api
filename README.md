# go-todo-api

A to-do REST API in Go, structured with Domain-Driven Design and backed by PostgreSQL.

**Status: working.** All eight endpoints are implemented and covered by 121
tests, from pure domain units up to full end-to-end runs against PostgreSQL.

```bash
make db-local && make run     # → http://localhost:8080/docs
```

| Layer | Tests | Needs a database |
|---|---|---|
| `domain` | 23 | no |
| `app` | 29 | no |
| `http` | 26 | no |
| `platform/config` | 9 | no |
| `postgres` | 14 | yes — `//go:build integration` |
| `test/e2e` | 5 | yes — `//go:build e2e` |

`make test-unit` runs the first four in about a second. `make test-all` runs
everything, one package at a time.

This document is the plan: what every file is for, and why it sits where it does.

---

## Layout

Context-first packaging — the bounded context is the top-level unit, and the DDD
layers live inside it.

```
go-todo-api/
├── cmd/api/                    entry point
├── internal/
│   ├── todo/                   ◄── the Todo bounded context
│   │   ├── domain/             business model — imports stdlib ONLY
│   │   ├── app/                use cases
│   │   ├── postgres/           persistence adapter
│   │   └── http/               transport adapter
│   └── platform/               context-agnostic infrastructure
│       ├── config/
│       ├── database/
│       ├── server/
│       └── docs/
├── api/                        OpenAPI contract + codegen config
├── migrations/                 numbered SQL
├── test/e2e/                   end-to-end tests
└── docs/                       design notes
```

## The dependency rule

```
http  ──▶  app  ──▶  domain  ◀──  postgres
```

Everything points inward. `domain/` imports nothing but the standard library —
no pgx, no net/http, no json tags.

This is not a convention to remember. Reversing an arrow creates an import
cycle, and Go refuses to compile. The architecture is verified on every
`go build`.

---

## File plan

### `cmd/api/`

| File | Responsibility |
|---|---|
| `main.go` | **Composition root.** The only file that knows Postgres exists. Loads config, opens the pool, constructs repository → service → handler, starts the server, handles graceful shutdown on SIGTERM. |

### `internal/todo/domain/` — the business model

Imports the standard library and nothing else. This layer would survive
deleting HTTP and PostgreSQL entirely.

| File | Responsibility |
|---|---|
| `todo.go` | The `Todo` aggregate root. Private fields, exported getters, and the behaviour that enforces invariants: `Complete`, `Reopen`, `UpdateTitle`, `Reschedule`, `IsOverdue`. Also `New` (creation rules) and `Reconstitute` (rebuild from storage, bypassing creation rules). |
| `title.go` | `Title` value object. Immutable, compared by value, constructible only through `NewTitle` — so an invalid title cannot exist anywhere in the system. |
| `id.go` | `ID` value object wrapping a UUID. A distinct type so a `UserID` can never be passed where a `TodoID` belongs. |
| `errors.go` | Sentinel errors that form the domain's vocabulary. Matched with `errors.Is`. Knows nothing about HTTP status codes. |
| `repository.go` | The `Repository` **interface** and the `Filter` type. Declared here, implemented in `postgres/`. This inversion is what makes the dependency rule hold. |
| `todo_test.go` | Entity behaviour and state transitions. No database, no mocks. |
| `title_test.go` | Value object validation, table-driven. |

### `internal/todo/app/` — use cases

Orchestrates. Never decides. If an `if` in this package encodes a business
rule, it escaped from the domain.

| File | Responsibility |
|---|---|
| `service.go` | One method per use case: `Create`, `GetByID`, `List`, `Update`, `Complete`, `Reopen`, `Delete`. Each loads the aggregate, calls a method on it, persists, maps out. Depends on the repository **interface**, never the Postgres type. |
| `commands.go` | Input boundary — `CreateTodoCommand`, `UpdateTodoCommand`, `ListTodosQuery`. Carry primitives, so callers need not know how to build domain value objects. |
| `dto.go` | Output boundary — `TodoDTO` plus the private mapping from entity to DTO. Inert data, so the transport layer cannot mutate the domain. |
| `service_test.go` | Orchestration correctness: does the service call the domain and persist the result, and do domain errors propagate untouched. |
| `fake_repository_test.go` | Hand-written in-memory `Repository`. No mocking library — Go's implicit interfaces make a plain struct sufficient. This is what lets the whole app layer be tested with no database. |

### `internal/todo/postgres/` — persistence adapter

| File | Responsibility |
|---|---|
| `repository.go` | Implements `domain.Repository` with hand-written SQL over `pgxpool`. Translates driver errors into domain errors, so `pgx.ErrNoRows` never escapes this package. Carries a compile-time interface assertion. |
| `mapper.go` | `todoRow` ↔ `domain.Todo`. Keeps the database schema and the domain model free to drift apart. Calls `Reconstitute`, never `New`. |
| `repository_test.go` | Integration tests behind `//go:build integration`. Round-trip fidelity, filter correctness, `ErrNotFound` on missing rows. |

### `internal/todo/http/` — transport adapter

Translates in both directions and does nothing else.

| File | Responsibility |
|---|---|
| `openapi_gen.go` | **Generated — do not edit.** Wire types, `ServerInterface`, `StrictServerInterface` and routing, produced from `api/openapi.yaml` by `make generate`. |
| `handler.go` | `Server`, which implements the generated `StrictServerInterface`. One method per endpoint, each receiving a parsed request object and returning a typed response. A compile-time assertion ties it to the spec. |
| `router.go` | Wires the generated routing onto a `ServeMux`. Implemented, not stubbed — the routes come from the spec, so hand-writing them would only create a second place to disagree. |
| `mapping.go` | `app.TodoDTO` → the generated `Todo` wire type. There are no hand-written request/response structs: the spec owns them. |
| `errors.go` | **The only file that knows both domain errors and HTTP status codes.** Also owns the stable machine-readable error codes, which are part of the public contract. |
| `generate.go` | The `//go:generate` directive driving `make generate`. |

### `internal/platform/` — shared infrastructure

Not about todos. A second bounded context would use all of it unchanged.

| File | Responsibility |
|---|---|
| `config/config.go` | Environment loading with defaults and validation. Built once at startup, passed in — never read from a global. |
| `database/postgres.go` | Creates and verifies the `pgxpool`. Hands out a pool; holds no queries. |
| `server/server.go` | `http.Server` with explicit timeouts (Go ships none by default) and graceful shutdown. |
| `docs/docs.go` | Serves `/openapi.yaml` and the Scalar reference at `/docs`. |

### `api/` — the contract

| File | Responsibility |
|---|---|
| `openapi.yaml` | **Source of truth.** 8 endpoints, 6 schemas. Code is generated from it, so spec/code drift is a compile error. Note it carries no `readOnly` on response-only schemas — that would force optional pointers on required fields. |
| `embed.go` | `//go:embed` of the spec, so the binary serves its own contract. |
| `oapi-codegen.yaml` | Generator config — models, std-http-server, strict-server. |
| `requests.http` | The same calls for poking at by hand, in VS Code / JetBrains HTTP client format. |

### Supporting files

| File | Responsibility |
|---|---|
| `migrations/000001_create_todos.{up,down}.sql` | Schema. `TIMESTAMPTZ` throughout, plus a partial index on `due_date`. |
| `test/e2e/api_test.go` | Full lifecycle through the real stack, behind `//go:build e2e`. |
| `tools.go` | Pins build-time tools (`oapi-codegen`) under a `tools` build tag — Go's `devDependencies`. |
| `go.mod` / `go.sum` | Manifest and lockfile. Packages live in the global `~/go/pkg/mod`, never in the project. |
| `Makefile` | Every command worth running. `make help` lists them. |
| `docker-compose.yml` | Postgres only. The app runs on the host to keep the loop fast. |
| `.vscode/` | gopls configured with `-tags=integration,e2e` so tagged tests aren't greyed out. |

---

## Layer responsibilities at a glance

| Layer | Owns | Must never |
|---|---|---|
| `domain` | Business rules, invariants, state transitions | Know about HTTP, SQL, or JSON |
| `app` | Orchestration: load → call domain → save | Contain business rules |
| `postgres` | SQL, row↔entity mapping | Leak driver types outward |
| `http` | JSON, status codes, routing, shape validation | Contain business rules |

## API surface

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/v1/todos` | Create |
| `GET` | `/api/v1/todos` | List — `?completed=` `?overdue=` `?limit=` `?offset=` |
| `GET` | `/api/v1/todos/{id}` | Fetch one |
| `PATCH` | `/api/v1/todos/{id}` | Partial update |
| `DELETE` | `/api/v1/todos/{id}` | Delete |
| `POST` | `/api/v1/todos/{id}/complete` | Mark done |
| `POST` | `/api/v1/todos/{id}/reopen` | Mark active |
| `GET` | `/health` | Liveness |
| `GET` | `/docs` | Interactive reference |
| `GET` | `/openapi.yaml` | Raw contract |

Completion is a sub-resource action rather than a `PATCH` field, because
completing a todo is a domain operation with its own rules — not an assignment.

---

## Build order

Innermost first. Each layer is fully testable the moment it's finished.

1. `domain/` — pure Go, no dependencies
2. `domain/*_test.go` — prove the rules before anything depends on them
3. `app/` — wire use cases against the fake repository
4. `postgres/` — make the real implementation satisfy the interface
5. `http/` — the thin translation layer
6. `cmd/api/main.go` — wire it together last

## Design decisions

Settled, and each one stated at the site it affects:

| Rule | Where it lives |
|---|---|
| `Complete`/`Reopen` are strict — wrong state is an error, and nothing mutates | `domain/todo.go` |
| PATCH semantics, with an explicit `ClearDueDate` flag | `app/commands.go` |
| One `Save()` doing an upsert | `postgres/repository.go` |
| Coarse error codes; `message` carries the detail | `http/errors.go` |
| Config fails fast on a missing `DATABASE_URL` | `platform/config/config.go` |
| `now` is passed in; the domain never calls `time.Now()` | `domain/todo.go` |
| A completed todo is never overdue | `domain/todo.go` |
| Past due dates are allowed | `domain/errors.go` |

## Further reading

- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) — setup, commands, and the codegen workflow
