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
//
// TODO(you): register both handlers on the mux.
func RegisterRoutes(mux *http.ServeMux) {
	// TODO
}

// specHandler serves the embedded OpenAPI document.
//
// TODO(you): write api.Spec with Content-Type application/yaml.
func specHandler(w http.ResponseWriter, r *http.Request) {
	_ = api.Spec // TODO(you): remove this line -- it only keeps the import alive
	// TODO
}

// docsHandler serves the Scalar single-page reference.
//
// The whole UI is one script tag pointed at the spec URL -- no build step, no
// node_modules. Closest thing you have used is swagger-ui-express, except this
// is nine lines of HTML instead of a dependency.
//
// TODO(you): write scalarHTML with Content-Type text/html.
func docsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
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
