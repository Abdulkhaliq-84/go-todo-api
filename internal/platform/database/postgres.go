// Package database owns the Postgres connection pool.
//
// Note what is NOT here: no queries, no SQL, no knowledge of todos. This
// package hands out a pool; internal/todo/postgres decides what to do with it.
package database

import (
	"context"
	"fmt"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates and verifies a connection pool.
//
// pgxpool is safe for concurrent use and is meant to be created ONCE at startup
// and shared. Every incoming HTTP request runs in its own goroutine and they
// all use this same pool -- which is exactly why it must be concurrency-safe.
//
// Coming from Node: there is no per-request client to acquire and release.
// Hand the pool to the repository and it borrows a connection per query.
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		// The URL may embed a password, so the raw value never reaches the
		// error message.
		return nil, fmt.Errorf("parsing DATABASE_URL: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	// NewWithConfig is LAZY -- it validates the config but opens no connection,
	// so a wrong host or bad credentials would not surface until the first
	// request. Pinging here moves that failure to startup, where it belongs:
	// the process dies before it can accept traffic it cannot serve.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	return pool, nil
}
