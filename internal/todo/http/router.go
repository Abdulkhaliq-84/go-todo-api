package http

import nethttp "net/http"

// RegisterRoutes wires the generated routing onto a mux.
//
// Implemented rather than left as a TODO because there is no decision here --
// the routes come from api/openapi.yaml, so writing them by hand would just be
// a second place for them to disagree.
//
// Two layers of generated code are at work:
//
//	NewStrictHandler  adapts your typed Server into the lower-level
//	                  ServerInterface -- it decodes bodies, parses params and
//	                  serialises whatever response object you return.
//	HandlerFromMux    registers every path from the spec onto the mux.
//
// The middlewares slice is where you would add logging, request IDs, auth or
// rate limiting. Go middleware is just a function wrapping a handler -- the
// same shape as Express middleware, minus the implicit next() chain.
func RegisterRoutes(mux *nethttp.ServeMux, srv *Server) {
	HandlerFromMux(NewStrictHandler(srv, nil), mux)
}
