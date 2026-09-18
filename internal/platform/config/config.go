// Package config loads settings from the environment.
//
// Lives in platform/, not in todo/, because configuration is not about todos.
// Any second bounded context would share it.
//
// No Pydantic BaseSettings here -- you read env vars and parse them yourself.
// The upside is that there is no magic: every variable the application reads is
// visible in this one file.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

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

// ErrMissingDatabaseURL is returned when DATABASE_URL is unset or empty.
var ErrMissingDatabaseURL = errors.New("DATABASE_URL is required")

// Load reads the environment and returns a Config.
//
// FAIL FAST: a missing or unparseable DATABASE_URL is an error, not a cue to
// fall back to localhost. A production deploy with the variable unset should die
// immediately and visibly. The alternative failure -- a service that starts,
// reports healthy, and is quietly talking to the wrong database, or to nothing
// -- is far more expensive to diagnose.
//
// Everything else gets a sensible default. Ports and timeouts have obviously
// right values; a database URL does not.
//
// Returns an error rather than calling log.Fatal or panicking. Load is a library
// function: the DECISION to exit belongs to main, which is the only place that
// should own the process lifecycle.
func Load() (*Config, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil, ErrMissingDatabaseURL
	}

	maxConns, err := getInt32("DB_MAX_CONNS", 10)
	if err != nil {
		return nil, err
	}
	minConns, err := getInt32("DB_MIN_CONNS", 2)
	if err != nil {
		return nil, err
	}
	if minConns > maxConns {
		return nil, fmt.Errorf("DB_MIN_CONNS (%d) exceeds DB_MAX_CONNS (%d)", minConns, maxConns)
	}

	readTimeout, err := getDuration("READ_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}
	writeTimeout, err := getDuration("WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, err
	}
	idleTimeout, err := getDuration("IDLE_TIMEOUT", 120*time.Second)
	if err != nil {
		return nil, err
	}
	shutdownTimeout, err := getDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	if err != nil {
		return nil, err
	}
	maxConnLifetime, err := getDuration("DB_MAX_CONN_LIFETIME", time.Hour)
	if err != nil {
		return nil, err
	}

	return &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			ReadTimeout:     readTimeout,
			WriteTimeout:    writeTimeout,
			IdleTimeout:     idleTimeout,
			ShutdownTimeout: shutdownTimeout,
		},
		Database: DatabaseConfig{
			URL:             url,
			MaxConns:        maxConns,
			MinConns:        minConns,
			MaxConnLifetime: maxConnLifetime,
		},
	}, nil
}

// ---------------------------------------------------------------------------
// env helpers
//
// Each one fails loudly on a malformed value rather than silently substituting
// the default. PORT=eighty is a typo the operator wants to hear about, not a
// reason to quietly serve on 8080 and leave them wondering.
// ---------------------------------------------------------------------------

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	// time.ParseDuration accepts "5s", "2m30s", "1h" -- the same spellings the
	// .env.example uses, so what an operator writes is what Go parses.
	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s=%q is not a duration (try 5s, 2m, 1h): %w", key, raw, err)
	}
	return v, nil
}

func getInt32(key string, fallback int32) (int32, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s=%q is not a number: %w", key, raw, err)
	}
	if v <= 0 {
		return 0, fmt.Errorf("%s=%d must be positive", key, v)
	}
	return int32(v), nil
}
