# go-todo-api

A to-do REST API in Go, structured with Domain-Driven Design and backed by PostgreSQL.

Built as a deliberate exercise in thinking in Go rather than in Node/Express or FastAPI.

## Architecture

Context-first packaging: the bounded context is the top-level unit, and the DDD
layers live inside it.

```
cmd/api/                     composition root -- the ONLY file that wires concrete to abstract
internal/
├── todo/                    the Todo bounded context
│   ├── domain/              entity, value objects, errors, Repository interface
│   ├── app/                 use cases, commands/queries, output DTOs
│   ├── postgres/            Repository implementation + row mapping
│   └── http/                handlers, routes, request/response DTOs, error mapping
└── platform/                context-agnostic infrastructure
    ├── config/              env loading
    ├── database/            pgx connection pool
    └── server/              http.Server + graceful shutdown
migrations/                  numbered SQL, run by golang-migrate
```

### The dependency rule

```
http  ──▶  app  ──▶  domain  ◀──  postgres
```

Everything points inward. `domain/` imports only the standard library — no pgx,
no net/http, no json tags. This is not a convention to remember: reversing an
arrow creates an import cycle, and Go refuses to build. The architecture is
verified on every `go build`.

### Layer responsibilities

| Layer | Owns | Must never |
|---|---|---|
| `domain` | Business rules, invariants, entity state transitions | Know about HTTP, SQL, or JSON |
| `app` | Orchestration: load → call domain → save | Contain business rules |
| `postgres` | SQL, row↔entity mapping, driver errors | Leak pgx types outward |
| `http` | JSON, status codes, routing, validation of *shape* | Contain business rules |

## API

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/v1/todos` | Create |
| `GET` | `/api/v1/todos` | List (`?completed=`, `?overdue=`, `?limit=`, `?offset=`) |
| `GET` | `/api/v1/todos/{id}` | Fetch one |
| `PATCH` | `/api/v1/todos/{id}` | Partial update |
| `DELETE` | `/api/v1/todos/{id}` | Delete |
| `POST` | `/api/v1/todos/{id}/complete` | Mark done |
| `POST` | `/api/v1/todos/{id}/reopen` | Mark active |
| `GET` | `/health` | Liveness |

Completion is a sub-resource action, not a `PATCH` field, because completing a
todo is a domain operation with its own rules — not a field assignment.

## Getting started

Go is not yet installed on this machine:

```bash
brew install go golang-migrate
```

Then:

```bash
cp .env.example .env
go get github.com/jackc/pgx/v5 github.com/google/uuid
go mod tidy
make db-up
make migrate-up
make run
```

`make help` lists everything else.

## Status

Scaffold only — every function is a signature with a `TODO(you)` marker.
Nothing is implemented yet. See [docs/DECISIONS.md](docs/DECISIONS.md) for the
design decisions deliberately left open.

Suggested build order, innermost first:

1. `domain/` — pure Go, no dependencies, fully testable with no database
2. `domain/todo_test.go` — prove the rules before anything can depend on them
3. `app/` — wire use cases against a fake repository
4. `postgres/` — make the real implementation satisfy the interface
5. `http/` — the thin translation layer
6. `cmd/api/main.go` — wire it together last

Building inward-out means each layer is testable the moment you finish it.
