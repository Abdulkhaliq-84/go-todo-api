// Package http adapts HTTP to the application layer.
//
// Its entire job is translation, in both directions:
//   inbound:  a parsed request object  ->  app commands/queries
//   outbound: app DTOs or errors       ->  typed response objects
//
// It imports app. It must NOT import postgres, and it must not contain business
// rules -- those live in the domain.
//
// THE SHAPE OF THIS FILE IS DICTATED BY api/openapi.yaml.
// openapi_gen.go declares StrictServerInterface; the assertion below makes the
// compiler check that this type satisfies it. Add an endpoint to the spec,
// regenerate, and this file stops compiling until you implement it. That is the
// whole point of spec-first: the contract cannot drift from the code.
package http

import (
	"context"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/app"
)

// Server implements the generated StrictServerInterface.
//
// Named Server rather than Handler because the generator already emits a
// package-level func Handler(ServerInterface) http.Handler.
//
// Dependency injection is the struct field -- no container, no decorators.
// Wiring happens once, by hand, in cmd/api/main.go.
type Server struct {
	service *app.Service
}

func NewServer(service *app.Service) *Server {
	return &Server{service: service}
}

// Compile-time contract check. This single line is what turns a spec change
// into a build failure instead of a runtime surprise.
var _ StrictServerInterface = (*Server)(nil)

// ---------------------------------------------------------------------------
// Endpoints
//
// Note the signature style: you receive a PARSED, TYPED request and RETURN a
// typed response. No ResponseWriter, no forgetting to return after writing an
// error. Much closer to an Express handler doing `return res.json(...)`, and it
// makes every handler a pure function you can call directly in a test.
//
// Path params, query params and the decoded JSON body all arrive already
// parsed and validated against the spec -- that part is genuinely FastAPI-like,
// and it is generated rather than reflected at runtime.
// ---------------------------------------------------------------------------

// ListTodos handles GET /api/v1/todos
//
// TODO(you): build an app.ListTodosQuery from request.Params (all pointers --
// nil means the client omitted it, so apply your defaults), call the service,
// return ListTodos200JSONResponse{Data: ..., Count: ...}.
func (s *Server) ListTodos(ctx context.Context, request ListTodosRequestObject) (ListTodosResponseObject, error) {
	return nil, nil // TODO
}

// CreateTodo handles POST /api/v1/todos
//
// request.Body is already decoded into *CreateTodoRequest and checked against
// the spec's constraints. What it has NOT been checked against is your domain
// rules -- those still run in domain.NewTitle. Two layers of validation with
// different jobs: shape here, meaning there.
//
// TODO(you): map body -> app.CreateTodoCommand, call Create, return
// CreateTodo201JSONResponse(toTodo(dto)).
func (s *Server) CreateTodo(ctx context.Context, request CreateTodoRequestObject) (CreateTodoResponseObject, error) {
	return nil, nil // TODO
}

// GetTodoById handles GET /api/v1/todos/{id}
//
// TODO(you): request.Id is already a parsed UUID. On domain.ErrNotFound return
// GetTodoById404JSONResponse -- returning the error instead would produce a 500.
func (s *Server) GetTodoById(ctx context.Context, request GetTodoByIdRequestObject) (GetTodoByIdResponseObject, error) {
	return nil, nil // TODO
}

// UpdateTodo handles PATCH /api/v1/todos/{id}
//
// TODO(you)
func (s *Server) UpdateTodo(ctx context.Context, request UpdateTodoRequestObject) (UpdateTodoResponseObject, error) {
	return nil, nil // TODO
}

// DeleteTodo handles DELETE /api/v1/todos/{id}
//
// TODO(you): return DeleteTodo204Response{} on success.
func (s *Server) DeleteTodo(ctx context.Context, request DeleteTodoRequestObject) (DeleteTodoResponseObject, error) {
	return nil, nil // TODO
}

// CompleteTodo handles POST /api/v1/todos/{id}/complete
//
// The spec promises a 409 here, and the domain delivers one: Complete on an
// already-completed todo returns ErrAlreadyComplete. The contract and the
// domain rule are the same decision, stated twice -- if one ever changes, the
// other has to change with it.
//
// TODO(you)
func (s *Server) CompleteTodo(ctx context.Context, request CompleteTodoRequestObject) (CompleteTodoResponseObject, error) {
	return nil, nil // TODO
}

// ReopenTodo handles POST /api/v1/todos/{id}/reopen
//
// TODO(you)
func (s *Server) ReopenTodo(ctx context.Context, request ReopenTodoRequestObject) (ReopenTodoResponseObject, error) {
	return nil, nil // TODO
}

// GetHealth handles GET /health
//
// TODO(you): return GetHealth200JSONResponse{Status: "ok"}.
// Decide whether this should also ping the database -- a health check that
// reports OK while the DB is unreachable is worse than no health check.
func (s *Server) GetHealth(ctx context.Context, request GetHealthRequestObject) (GetHealthResponseObject, error) {
	return nil, nil // TODO
}
