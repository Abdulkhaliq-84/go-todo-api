// Command api is the entry point.
//
// THIS FILE IS THE COMPOSITION ROOT -- the single place where abstract meets
// concrete. It is the only file in the codebase that knows Postgres is the
// database, and it knows that for exactly one line.
//
// The wiring below is the dependency rule as executable code:
//
//	postgres.NewRepository(pool)   <- concrete infrastructure
//	         |
//	         v  (satisfies domain.Repository)
//	app.NewService(repo)           <- knows only the interface
//	         |
//	         v
//	todohttp.NewServer(service)    <- knows only the application layer
//
// Swapping Postgres for an in-memory fake is a one-line change, here, and
// nothing else in the codebase notices. That is the payoff for the indirection.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/platform/config"
	"github.com/Abdulkhaliq-84/go-todo-api/internal/platform/database"
	"github.com/Abdulkhaliq-84/go-todo-api/internal/platform/docs"
	"github.com/Abdulkhaliq-84/go-todo-api/internal/platform/server"
	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/app"
	todohttp "github.com/Abdulkhaliq-84/go-todo-api/internal/todo/http"
	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/postgres"
)

func main() {
	// main does one thing: decide the exit code. Everything else lives in run,
	// so that every other function can return an error instead of killing the
	// process -- and so deferred cleanup actually runs. os.Exit skips defers,
	// which is why calling it anywhere but here would silently leak the pool.
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// ---- configuration -----------------------------------------------------
	cfg, err := config.Load()
	if err != nil {
		return err // fail fast: no DATABASE_URL, no server
	}

	// ---- signals -----------------------------------------------------------
	//
	// NotifyContext gives a context that cancels on SIGINT or SIGTERM. That one
	// context is both the startup deadline and the shutdown trigger, so Ctrl+C
	// during a slow database connect aborts it rather than hanging.
	//
	// No Express equivalent -- there you wire process.on('SIGTERM') by hand and
	// hope nothing else registered a competing handler.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ---- infrastructure ----------------------------------------------------
	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer pool.Close()
	logger.Info("connected to database", "max_conns", cfg.Database.MaxConns)

	// ---- the three lines that are the architecture -------------------------
	repo := postgres.NewRepository(pool) // the ONLY mention of Postgres
	service := app.NewService(repo)      // sees domain.Repository, nothing more
	todoServer := todohttp.NewServer(service)

	// ---- routing -----------------------------------------------------------
	mux := http.NewServeMux()
	todohttp.RegisterRoutes(mux, todoServer, logger) // every path from the spec
	docs.RegisterRoutes(mux)                         // /docs and /openapi.yaml

	srv := server.New(cfg.Server, mux)

	// ---- serve -------------------------------------------------------------
	//
	// Start blocks, so it runs in its own goroutine and reports failure over a
	// channel. A buffered channel of one means the goroutine can send and exit
	// even if nobody is receiving yet -- with an unbuffered channel it would
	// block forever on a startup failure, and the process would hang instead of
	// reporting the error.
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server listening",
			"addr", srv.Addr(),
			"docs", "http://localhost:"+cfg.Server.Port+"/docs")
		serverErr <- srv.Start()
	}()

	// Wait for whichever comes first: the server falling over, or a signal.
	select {
	case err := <-serverErr:
		if err != nil {
			return err
		}
		return nil

	case <-ctx.Done():
		logger.Info("shutdown signal received", "timeout", cfg.Server.ShutdownTimeout)
	}

	// ---- graceful shutdown -------------------------------------------------
	//
	// A FRESH context with its own deadline. ctx is already cancelled at this
	// point -- that is why we are here -- so passing it to Shutdown would abort
	// instantly and cut off the very requests we are trying to let finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// Requests were still running when time ran out. Worth saying
			// plainly: the timeout is probably too short, or something is stuck.
			logger.Warn("shutdown deadline exceeded; some requests were cut off")
			return nil
		}
		return err
	}

	logger.Info("shutdown complete")
	return nil
}
