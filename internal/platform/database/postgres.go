// Package database owns the Postgres connection pool.
//
// Note what is NOT here: no queries, no SQL, no knowledge of todos. This
// package hands out a pool; internal/todo/postgres decides what to do with it.
package database

import (
	"context"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates and verifies a connection pool.
//
// pgxpool is safe for concurrent use and is meant to be created ONCE at startup
// and shared. Every incoming HTTP request runs in its own goroutine and they
// all use this same pool -- which is exactly why it must be concurrency-safe.
//
// TODO(you): parse the config into a pgxpool.Config, create the pool, then Ping
// it so a bad DATABASE_URL fails at startup rather than on the first request.
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	return nil, nil // TODO
}
