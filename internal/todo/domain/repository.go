package domain

import "context"

// Repository is declared HERE, in the domain, and implemented OUT THERE, in
// internal/todo/postgres. That inversion is the heart of the architecture:
// the domain states what it needs; infrastructure volunteers to provide it.
//
// Because of it, the import arrow points postgres -> domain and never back.
// Go makes that structural: if domain ever imported postgres you would have an
// import cycle, and the build would fail. Your architecture is verified on
// every `go build`, not by code review.
//
// Note there is no `implements` keyword anywhere. Go interfaces are satisfied
// implicitly -- postgres.Repository has the right method set, so it IS a
// domain.Repository. Nothing needs to declare the relationship.
//
// Every method takes context.Context as its first argument. That is the Go
// convention for cancellation and deadlines, and it is how a dropped HTTP
// connection cancels the Postgres query underneath it.
type Repository interface {
	// Save persists a todo, whether it is new or already stored
	// One method rather than Insert + Update, so the
	// service never has to track whether an entity is new -- that is a
	// persistence concern and it stays on this side of the boundary.
	Save(ctx context.Context, todo *Todo) error
	FindByID(ctx context.Context, id ID) (*Todo, error)
	FindAll(ctx context.Context, filter Filter) ([]*Todo, error)
	Delete(ctx context.Context, id ID) error
}

// Filter expresses query intent in DOMAIN terms -- not SQL, not query strings.
// The postgres package turns this into a WHERE clause; the http package builds
// it from ?completed=true. Neither one leaks into the other.
//
// Pointer fields distinguish "not specified" (nil) from "explicitly false".
type Filter struct {
	Completed *bool
	Overdue   *bool
	Limit     int
	Offset    int
}
