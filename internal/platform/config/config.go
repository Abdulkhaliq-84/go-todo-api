// Package config loads settings from the environment.
//
// Lives in platform/, not in todo/, because configuration is not about todos.
// Any second bounded context you add would share it.
//
// No Pydantic BaseSettings here -- you read env vars and parse them yourself.
// The upside is there is no magic: you can see exactly which variables the app
// reads by reading this one file.
package config

import "time"

// Config is the fully-resolved configuration. It is built once at startup and
// then treated as immutable -- passed into constructors, never read from a
// global. Globals make code untestable and hide dependencies.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
}

// Load reads the environment and returns a Config.
//
// FAIL FAST (docs/DECISIONS.md #5): a missing or unparseable DATABASE_URL is an
// error, not a cue to fall back to localhost. A production deploy with the
// variable unset should die immediately and visibly. The alternative failure --
// a service that starts, reports healthy, and is quietly talking to the wrong
// database, or to nothing -- is far more expensive to diagnose.
//
// Everything else gets a sensible default. Ports and timeouts have obviously
// right values; a database URL does not.
//
// Return an error rather than calling log.Fatal or panicking. Load is a
// library function: the DECISION to exit belongs to main, which is the only
// place that should own the process lifecycle.
//
// TODO(you): read the env vars, apply defaults, validate, return.
func Load() (*Config, error) {
	return nil, nil // TODO
}

// getEnv returns an env var or a fallback.
//
// TODO(you)
func getEnv(key, fallback string) string {
	return "" // TODO
}
