// Package docs serves the OpenAPI document and an interactive UI.
//
// Lives in platform/ because documentation is not about todos -- a second
// bounded context would be described by the same spec and served by the same
// handler.
package docs

import (
	"net/http"

	"github.com/Abdulkhaliq-84/go-todo-api/api"
)

// ScalarCDN is where the docs UI script is fetched from BY THE BROWSER.
//
// Being precise about this: the spec itself is embedded in the binary and
// served locally, but the Scalar renderer is a script tag pointing at jsDelivr.
// So /docs needs internet access in the browser to render. To make it fully
// offline, download scalar.js into this package and embed it the same way the
// spec is embedded.
const ScalarCDN = "https://cdn.jsdelivr.net/npm/@scalar/api-reference"

// RegisterRoutes mounts the documentation endpoints.
//
//	GET /openapi.yaml  the raw contract -- what a client generator consumes
//	GET /docs          the interactive reference
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /openapi.yaml", specHandler)
	mux.HandleFunc("GET /docs", docsHandler)
}

// specHandler serves the embedded OpenAPI document.
//
// Served from the binary rather than from disk, so a deployed container has no
// file to lose and the spec can never drift from the code that shipped with it.
func specHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(api.Spec)
}

// docsHandler serves the Scalar single-page reference.
//
// The whole UI is one script tag pointed at the spec URL -- no build step, no
// node_modules. The closest thing you have used is swagger-ui-express, except
// this is nine lines of HTML instead of a dependency.
func docsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(scalarHTML))
}

const scalarHTML = `<!doctype html>
<html>
  <head>
    <title>Todo API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script id="api-reference" data-url="/openapi.yaml"></script>
    <script src="` + ScalarCDN + `"></script>
  </body>
</html>`
