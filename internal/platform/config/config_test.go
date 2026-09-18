package config_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/platform/config"
)

// t.Setenv sets an environment variable for the duration of one test and
// restores it afterwards -- including on failure. It also marks the test as
// non-parallel, because the environment is process-wide state and two parallel
// tests writing it would race. Doing this by hand with os.Setenv plus a defer
// is the older pattern and leaks on a panic.

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := config.Load()

	if !errors.Is(err, config.ErrMissingDatabaseURL) {
		t.Errorf("Load error = %v, want config.ErrMissingDatabaseURL", err)
	}
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/todos")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout = %v, want 5s", cfg.Server.ReadTimeout)
	}
	if cfg.Server.ShutdownTimeout != 15*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 15s", cfg.Server.ShutdownTimeout)
	}
	if cfg.Database.MaxConns != 10 {
		t.Errorf("MaxConns = %d, want 10", cfg.Database.MaxConns)
	}

	// Every timeout must be non-zero. Go's zero value for a timeout means NO
	// TIMEOUT, so a default that silently stayed at zero would leave the server
	// willing to hold a connection open forever.
	timeouts := map[string]time.Duration{
		"ReadTimeout":     cfg.Server.ReadTimeout,
		"WriteTimeout":    cfg.Server.WriteTimeout,
		"IdleTimeout":     cfg.Server.IdleTimeout,
		"ShutdownTimeout": cfg.Server.ShutdownTimeout,
	}
	for name, d := range timeouts {
		if d == 0 {
			t.Errorf("%s is zero, which means no timeout at all", name)
		}
	}
}

func TestLoad_ReadsOverrides(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/todos")
	t.Setenv("PORT", "9999")
	t.Setenv("READ_TIMEOUT", "30s")
	t.Setenv("DB_MAX_CONNS", "42")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Server.Port != "9999" {
		t.Errorf("Port = %q, want 9999", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("ReadTimeout = %v, want 30s", cfg.Server.ReadTimeout)
	}
	if cfg.Database.MaxConns != 42 {
		t.Errorf("MaxConns = %d, want 42", cfg.Database.MaxConns)
	}
}

// A malformed value must fail loudly rather than silently falling back to the
// default. PORT=eighty is a typo the operator wants to hear about, not a reason
// to quietly serve on 8080 and leave them wondering why.
func TestLoad_RejectsMalformedValues(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"non-duration timeout", "READ_TIMEOUT", "quickly"},
		{"timeout missing a unit", "WRITE_TIMEOUT", "30"},
		{"non-numeric pool size", "DB_MAX_CONNS", "lots"},
		{"zero pool size", "DB_MAX_CONNS", "0"},
		{"negative pool size", "DB_MIN_CONNS", "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://localhost:5432/todos")
			t.Setenv(tt.key, tt.value)

			_, err := config.Load()

			if err == nil {
				t.Fatalf("Load accepted %s=%q, want an error", tt.key, tt.value)
			}
			// The message must name the offending variable, or an operator has
			// to guess which of a dozen settings is wrong.
			if !strings.Contains(err.Error(), tt.key) {
				t.Errorf("error %q does not mention %s", err, tt.key)
			}
		})
	}
}

// A pool whose minimum exceeds its maximum is nonsense pgx would accept and
// then behave strangely under, so it is rejected here.
func TestLoad_RejectsMinConnsAboveMax(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/todos")
	t.Setenv("DB_MAX_CONNS", "5")
	t.Setenv("DB_MIN_CONNS", "10")

	if _, err := config.Load(); err == nil {
		t.Error("Load accepted DB_MIN_CONNS > DB_MAX_CONNS, want an error")
	}
}
