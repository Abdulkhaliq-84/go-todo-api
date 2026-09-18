.PHONY: help generate docs-open validate-spec run build db-local test test-unit test-integration test-e2e test-all test-cover test-race lint fmt vet tidy db-up db-down migrate-up migrate-down

DB_URL ?= postgres://postgres:postgres@localhost:5432/todos?sslmode=disable

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

# --- OpenAPI ----------------------------------------------------------------
# api/openapi.yaml is the source of truth. Edit it, regenerate, and let the
# compiler tell you which handlers no longer satisfy the contract.

generate: ## Regenerate Go types + ServerInterface from api/openapi.yaml
	go generate ./...

validate-spec: ## Lint the OpenAPI document (npx, no install needed)
	npx --yes @redocly/cli lint api/openapi.yaml

docs-open: ## Open the interactive API reference (requires make run)
	open http://localhost:8080/docs

run: ## Run the API
	go run ./cmd/api

build: ## Compile to bin/todo-api
	go build -o bin/todo-api ./cmd/api

# --- Testing pyramid ---------------------------------------------------------
# Unit tests run on every save. Integration and e2e are gated behind build tags
# so they only run when you ask, and never slow down the inner loop.

test: test-unit ## Alias for test-unit

test-unit: ## Fast tests: domain, app, http. No database required.
	go test ./... -v

test-integration: ## Real Postgres required (make db-up OR make db-local)
	go test -tags=integration ./internal/todo/postgres/... -v

test-e2e: ## Whole app against a real database
	go test -tags=e2e ./test/e2e/... -v

test-all: ## Everything
	go test -tags="integration e2e" ./... -v

test-race: ## Unit tests under the race detector
	go test -race ./...

test-cover: ## Coverage report -> coverage.html
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

vet: ## Run go vet (built-in static analysis)
	go vet ./...

fmt: ## Format all code
	go fmt ./...

tidy: ## Sync go.mod/go.sum with actual imports
	go mod tidy

lint: ## Run golangci-lint (brew install golangci-lint)
	golangci-lint run

db-local: ## Create + migrate todos_test on a Postgres already running locally
	createdb todos_test 2>/dev/null || true
	psql -q -d todos_test -f migrations/000001_create_todos.up.sql
	@echo "todos_test ready"

db-up: ## Start Postgres
	docker compose up -d postgres

db-down: ## Stop Postgres
	docker compose down

migrate-up: ## Apply migrations (brew install golang-migrate)
	migrate -path migrations -database "$(DB_URL)" up

migrate-down: ## Roll back the last migration
	migrate -path migrations -database "$(DB_URL)" down 1
