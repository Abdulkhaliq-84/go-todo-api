package http

import (
	"log/slog"
	nethttp "net/http"
)

// RegisterRoutes wires the generated routing onto a mux.
//
// Two layers of generated code are at work:
//
//	NewStrictHandlerWithOptions  adapts the typed Server into the lower-level
//	                             ServerInterface -- decoding bodies, parsing
//	                             params, serialising whatever response object a
//	                             handler returns.
//	HandlerFromMux               registers every path from the spec onto the mux.
//
// The routes themselves are not written here. They come from api/openapi.yaml,
// so hand-writing them would only create a second place for them to disagree.
//
// The options are NOT optional. NewStrictHandler's defaults are:
//
//	RequestErrorHandlerFunc:  http.Error(w, err.Error(), 400)
//	ResponseErrorHandlerFunc: http.Error(w, err.Error(), 500)
//
// Both write err.Error() straight into the response body as plain text. The
// 500 case would hand a client -- and an attacker -- messages like
// `pq: relation "todos" does not exist`, and neither matches the JSON error
// shape the API promises everywhere else. So both are replaced below.
func RegisterRoutes(mux *nethttp.ServeMux, srv *Server, logger *slog.Logger) {
	options := StrictHTTPServerOptions{
		// Fires when the generated code cannot parse the request: malformed
		// JSON, a bad query parameter, an id that is not a UUID. The client's
		// own input is safe to describe back to them.
		RequestErrorHandlerFunc: func(w nethttp.ResponseWriter, r *nethttp.Request, err error) {
			logger.Debug("rejected malformed request",
				"method", r.Method, "path", r.URL.Path, "error", err)

			writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{
				Error:   codeInvalidID,
				Message: "the request could not be parsed",
			})
		},

		// Fires when a handler returns a non-nil error -- by our convention,
		// only for errors with no meaningful HTTP representation.
		//
		// The real error is LOGGED, and the client is told nothing about it.
		// That asymmetry is the point: you get the stack, they get a code.
		ResponseErrorHandlerFunc: func(w nethttp.ResponseWriter, r *nethttp.Request, err error) {
			logger.Error("unhandled error serving request",
				"method", r.Method, "path", r.URL.Path, "error", err)

			writeJSON(w, nethttp.StatusInternalServerError, ErrorResponse{
				Error:   codeInternal,
				Message: "an unexpected error occurred",
			})
		},
	}

	HandlerFromMux(NewStrictHandlerWithOptions(srv, nil, options), mux)
}
