package http

// The directive below is run by `go generate ./...` (wired up as `make generate`).
//
// It reads ../../../api/openapi.yaml and writes openapi_gen.go into this
// package. That file is COMMITTED to git -- unlike node_modules or a dist/
// folder, Go convention is to commit generated code so the repo builds without
// a generation step and so diffs show exactly what the spec change did.
//
// The workflow, which will feel familiar from openapi-typescript:
//
//   1. edit api/openapi.yaml
//   2. make generate
//   3. go build ./...      <- now FAILS if your handlers no longer match the spec
//   4. fix the handlers
//
// Step 3 is the whole reason for choosing spec-first. The documentation cannot
// silently drift from the implementation, because drift does not compile.

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../../api/oapi-codegen.yaml ../../../api/openapi.yaml
