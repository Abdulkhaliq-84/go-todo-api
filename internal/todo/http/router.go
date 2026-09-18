package http

import nethttp "net/http"

// RegisterRoutes wires URL patterns to handler methods.
//
// Since Go 1.22 the standard library's ServeMux understands method prefixes and
// {path} wildcards, so a third-party router is genuinely optional now. This is
// your app.get('/todos/:id', ...) equivalent -- just declared as strings.
//
// TODO(you): register the routes. Suggested shape:
//
//   mux.HandleFunc("POST   /api/v1/todos",               h.Create)
//   mux.HandleFunc("GET    /api/v1/todos",               h.List)
//   mux.HandleFunc("GET    /api/v1/todos/{id}",          h.GetByID)
//   mux.HandleFunc("PATCH  /api/v1/todos/{id}",          h.Update)
//   mux.HandleFunc("DELETE /api/v1/todos/{id}",          h.Delete)
//   mux.HandleFunc("POST   /api/v1/todos/{id}/complete", h.Complete)
//   mux.HandleFunc("POST   /api/v1/todos/{id}/reopen",   h.Reopen)
func RegisterRoutes(mux *nethttp.ServeMux, h *Handler) {
	// TODO
}
