// Package http adapts HTTP to the application layer.
//
// Its entire job is translation, in both directions:
//
//	inbound:  a parsed request object  ->  app commands/queries
//	outbound: app DTOs or errors       ->  typed response objects
//
// It imports app. It must NOT import postgres, and it must contain no business
// rules -- those live in the domain.
//
// THE SHAPE OF THIS FILE IS DICTATED BY api/openapi.yaml. openapi_gen.go
// declares StrictServerInterface; the assertion below makes the compiler check
// that this type satisfies it. Add an endpoint to the spec, regenerate, and
// this file stops compiling until it is implemented.
package http

import (
	"context"
	nethttp "net/http"

	"github.com/Abdulkhaliq-84/go-todo-api/internal/todo/app"
)

// Server implements the generated StrictServerInterface.
//
// Named Server rather than Handler because the generator already emits a
// package-level func Handler(ServerInterface) http.Handler.
//
// Dependency injection is the struct field -- no container, no decorators.
type Server struct {
	service *app.Service
}

func NewServer(service *app.Service) *Server {
	return &Server{service: service}
}

var _ StrictServerInterface = (*Server)(nil)

// ---------------------------------------------------------------------------
// Endpoints
//
// Signature style: receive a PARSED, TYPED request and RETURN a typed response.
// No ResponseWriter, so forgetting to return after writing an error is not
// possible. Much closer to `return res.json(...)` than Go usually gets, and it
// makes each handler a pure function a test can call directly.
//
// Error handling follows one rule everywhere: ask classify() what this error
// means. If it maps to a status this endpoint documents, return that typed
// response; if it is internal, return it as an `error` so the strict runtime
// logs it and answers a generic 500 (see router.go).
// ---------------------------------------------------------------------------

// ListTodos handles GET /api/v1/todos
func (s *Server) ListTodos(ctx context.Context, request ListTodosRequestObject) (ListTodosResponseObject, error) {
	dtos, err := s.service.List(ctx, toListQuery(request.Params))
	if err != nil {
		if isInternal(err) {
			return nil, err
		}
		_, body := classify(err)
		return ListTodos400JSONResponse{BadRequestJSONResponse(body)}, nil
	}

	todos := toTodos(dtos)
	return ListTodos200JSONResponse{Data: todos, Count: len(todos)}, nil
}

// CreateTodo handles POST /api/v1/todos
//
// request.Body is already decoded and checked against the spec's constraints.
// What it has NOT been checked against is the domain rules -- those still run
// in domain.NewTitle, reached through the service. Two layers of validation
// with different jobs: shape here, meaning there.
func (s *Server) CreateTodo(ctx context.Context, request CreateTodoRequestObject) (CreateTodoResponseObject, error) {
	if request.Body == nil {
		return CreateTodo400JSONResponse{BadRequestJSONResponse{
			Error:   codeInvalidTitle,
			Message: "a request body is required",
		}}, nil
	}

	dto, err := s.service.Create(ctx, toCreateCommand(*request.Body))
	if err != nil {
		if isInternal(err) {
			return nil, err
		}
		_, body := classify(err)
		return CreateTodo400JSONResponse{BadRequestJSONResponse(body)}, nil
	}
	return CreateTodo201JSONResponse(toTodo(dto)), nil
}

// GetTodoById handles GET /api/v1/todos/{id}
//
// request.Id is already a parsed UUID -- the generated code rejected anything
// else before this method ran.
func (s *Server) GetTodoById(ctx context.Context, request GetTodoByIdRequestObject) (GetTodoByIdResponseObject, error) {
	dto, err := s.service.GetByID(ctx, request.Id.String())
	if err != nil {
		if isInternal(err) {
			return nil, err
		}
		status, body := classify(err)
		if status == nethttp.StatusNotFound {
			return GetTodoById404JSONResponse{NotFoundJSONResponse(body)}, nil
		}
		return GetTodoById400JSONResponse{BadRequestJSONResponse(body)}, nil
	}
	return GetTodoById200JSONResponse(toTodo(dto)), nil
}

// UpdateTodo handles PATCH /api/v1/todos/{id}
func (s *Server) UpdateTodo(ctx context.Context, request UpdateTodoRequestObject) (UpdateTodoResponseObject, error) {
	if request.Body == nil {
		return UpdateTodo400JSONResponse{BadRequestJSONResponse{
			Error:   codeInvalidTitle,
			Message: "a request body is required",
		}}, nil
	}

	dto, err := s.service.Update(ctx, request.Id.String(), toUpdateCommand(*request.Body))
	if err != nil {
		if isInternal(err) {
			return nil, err
		}
		status, body := classify(err)
		if status == nethttp.StatusNotFound {
			return UpdateTodo404JSONResponse{NotFoundJSONResponse(body)}, nil
		}
		return UpdateTodo400JSONResponse{BadRequestJSONResponse(body)}, nil
	}
	return UpdateTodo200JSONResponse(toTodo(dto)), nil
}

// DeleteTodo handles DELETE /api/v1/todos/{id}
func (s *Server) DeleteTodo(ctx context.Context, request DeleteTodoRequestObject) (DeleteTodoResponseObject, error) {
	if err := s.service.Delete(ctx, request.Id.String()); err != nil {
		if isInternal(err) {
			return nil, err
		}
		status, body := classify(err)
		if status == nethttp.StatusNotFound {
			return DeleteTodo404JSONResponse{NotFoundJSONResponse(body)}, nil
		}
		return DeleteTodo400JSONResponse{BadRequestJSONResponse(body)}, nil
	}
	return DeleteTodo204Response{}, nil
}

// CompleteTodo handles POST /api/v1/todos/{id}/complete
//
// The spec promises a 409 here, and the domain delivers one: Complete on an
// already-completed todo returns ErrAlreadyComplete. The contract and the
// domain rule are the same decision stated twice -- change one and the other
// has to change with it.
func (s *Server) CompleteTodo(ctx context.Context, request CompleteTodoRequestObject) (CompleteTodoResponseObject, error) {
	dto, err := s.service.Complete(ctx, request.Id.String())
	if err != nil {
		if isInternal(err) {
			return nil, err
		}
		status, body := classify(err)
		if status == nethttp.StatusConflict {
			return CompleteTodo409JSONResponse{ConflictJSONResponse(body)}, nil
		}
		return CompleteTodo404JSONResponse{NotFoundJSONResponse(body)}, nil
	}
	return CompleteTodo200JSONResponse(toTodo(dto)), nil
}

// ReopenTodo handles POST /api/v1/todos/{id}/reopen
func (s *Server) ReopenTodo(ctx context.Context, request ReopenTodoRequestObject) (ReopenTodoResponseObject, error) {
	dto, err := s.service.Reopen(ctx, request.Id.String())
	if err != nil {
		if isInternal(err) {
			return nil, err
		}
		status, body := classify(err)
		if status == nethttp.StatusConflict {
			return ReopenTodo409JSONResponse{ConflictJSONResponse(body)}, nil
		}
		return ReopenTodo404JSONResponse{NotFoundJSONResponse(body)}, nil
	}
	return ReopenTodo200JSONResponse(toTodo(dto)), nil
}

// GetHealth handles GET /health
//
// Reports that the PROCESS is alive, deliberately without touching the
// database. A liveness probe that fails on a database blip gets the container
// killed and restarted, which fixes nothing and turns a brief outage into a
// crash loop. A readiness probe that does ping the database is a separate
// endpoint worth adding when there is something to orchestrate.
func (s *Server) GetHealth(ctx context.Context, request GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200JSONResponse{Status: "ok"}, nil
}
