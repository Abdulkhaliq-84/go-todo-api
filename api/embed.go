// Package api embeds the OpenAPI document so the running binary can serve it.
//
// go:embed can only reach files in its OWN directory or below -- no "../".
// That constraint is why this tiny package lives next to the spec rather than
// the docs package reaching across for it.
//
// The result: one openapi.yaml feeds three consumers -- the code generator, the
// /docs UI, and the raw spec endpoint. They cannot disagree, because there is
// only one file.
package api

import _ "embed"

//go:embed openapi.yaml
var Spec []byte
