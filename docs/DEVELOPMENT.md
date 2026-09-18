# Development

Setup, commands, and workflows. The structural plan lives in the
[README](../README.md).

## Prerequisites

Go 1.25+ (1.27.1 installed here), Docker, and `golang-migrate`.

```bash
brew install golang-migrate
```

`~/go/bin` must be on your `PATH` for `go install`-ed tools to resolve:

```bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
```

## First run

```bash
cp .env.example .env
go mod tidy
make db-up
make migrate-up
make run
```

Then open http://localhost:8080/docs.

## Commands

`make help` lists everything. The ones that matter:

| Command | Does |
|---|---|
| `make run` | Run the API |
| `make generate` | Regenerate Go types from `api/openapi.yaml` |
| `make test-unit` | Fast tests — domain, app, http. No database. |
| `make test-integration` | Repository against real Postgres (needs `make db-up`) |
| `make test-e2e` | Whole app, wired |
| `make test-race` | Unit tests under the race detector |
| `make test-cover` | Coverage → `coverage.html` |
| `make migrate-up` / `migrate-down` | Apply / roll back schema |
| `make validate-spec` | Lint the OpenAPI document |

## The OpenAPI workflow

`api/openapi.yaml` is the source of truth, not documentation written after the
fact. `oapi-codegen` generates Go models and a `ServerInterface` from it, so a
handler that no longer matches the spec is a **compile error** — the same
discipline `openapi-typescript` gives a React client, applied to the server.

```
edit api/openapi.yaml  →  make generate  →  go build ./...  →  fix what broke
```

Generated code is committed, per Go convention: the repo builds without a
generation step, and diffs show exactly what a spec change did.

The same file is embedded in the binary, served at `/openapi.yaml`, and rendered
at `/docs`. Spec, docs and code cannot drift apart — they are one file.

## The testing pyramid

Build tags keep the layers separate, so `go test ./...` never touches a database
and stays fast enough to run on every save.

| Layer | Count | Speed | Needs |
|---|---|---|---|
| `domain` | many | µs | nothing |
| `app` | dozens | µs | fake repository |
| `http` | dozens | µs | `httptest` |
| `postgres` | ~12 | seconds | real Postgres, `//go:build integration` |
| `e2e` | a handful | seconds | everything real, `//go:build e2e` |

Every test currently starts with `t.Skip("TODO(you)")`. Delete the skip as you
implement each piece. The tables and assertions are already written as comments.

`api/requests.http` holds the same calls for manual exploration.

## Editor

`.vscode/settings.json` is committed and configures gopls with
`-tags=integration,e2e`. Without it, the tagged test files show as excluded by
build constraints — no autocomplete, no go-to-definition — which reads as a
broken editor rather than a config gap.

`organizeImports` runs on save. That is close to mandatory in Go, since an
unused import is a compile error rather than a lint warning.
