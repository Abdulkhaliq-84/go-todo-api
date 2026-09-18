.PHONY: help run build test test-cover lint fmt vet tidy db-up db-down migrate-up migrate-down

DB_URL ?= postgres://postgres:postgres@localhost:5432/todos?sslmode=disable

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

run: ## Run the API
	go run ./cmd/api

build: ## Compile to bin/todo-api
	go build -o bin/todo-api ./cmd/api

test: ## Run all tests
	go test ./... -v

test-cover: ## Run tests with a coverage report
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

db-up: ## Start Postgres
	docker compose up -d postgres

db-down: ## Stop Postgres
	docker compose down

migrate-up: ## Apply migrations (brew install golang-migrate)
	migrate -path migrations -database "$(DB_URL)" up

migrate-down: ## Roll back the last migration
	migrate -path migrations -database "$(DB_URL)" down 1
