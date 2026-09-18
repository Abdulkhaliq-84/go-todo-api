// Command api is the entry point.
//
// THIS FILE IS THE COMPOSITION ROOT -- the single place where abstract meets
// concrete. It is the only file that knows Postgres is the database, and it
// knows that for exactly one line.
//
// Read the wiring below and you can see the dependency rule as executable code:
//
//     postgres.NewRepository(pool)   <- concrete infrastructure
//              |
//              v  (satisfies domain.Repository)
//     app.NewService(repo)           <- knows only the interface
//              |
//              v
//     http.NewHandler(service)       <- knows only the app layer
//
// Swapping Postgres for an in-memory fake is a one-line change, here, and
// nothing else in the codebase notices. That is the payoff for all the
// indirection.
package main

func main() {
	// TODO(you): implement startup. The sequence:
	//
	//  1. cfg, err := config.Load()
	//  2. ctx, stop := signal.NotifyContext(context.Background(),
	//                    os.Interrupt, syscall.SIGTERM)
	//     defer stop()
	//  3. pool, err := database.NewPool(ctx, cfg.Database)
	//     defer pool.Close()
	//  4. repo    := postgres.NewRepository(pool)
	//     service := app.NewService(repo)
	//     srv     := todohttp.NewServer(service)
	//  5. mux := http.NewServeMux()
	//     todohttp.RegisterRoutes(mux, srv)   // routes come from the spec,
	//                                         // /health included
	//     docs.RegisterRoutes(mux)            // /docs and /openapi.yaml
	//  6. srv := server.New(cfg.Server, mux)
	//     go srv.Start()
	//  7. <-ctx.Done()            // blocks until SIGINT/SIGTERM
	//     srv.Shutdown(shutdownCtx)
	//
	// Step 7 is the Go idiom for graceful shutdown: a context that cancels on
	// signal, and a goroutine serving in the background. No equivalent in
	// Express -- there you would wire process.on('SIGTERM') by hand.
}
