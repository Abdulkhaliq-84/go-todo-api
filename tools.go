//go:build tools

// Package tools pins build-time tool dependencies.
//
// The problem this solves: oapi-codegen is needed to BUILD the project but is
// never imported by application code, so `go mod tidy` would happily delete it
// from go.mod and everyone's generated code would drift by version.
//
// The fix is this file. The `tools` build tag means it never compiles into your
// binary, but the blank imports are still real imports as far as the module
// graph is concerned -- so the version is pinned in go.mod and go.sum, and
// every developer and CI run generates identical code.
//
// This is Go's answer to devDependencies. Clunkier, and it does guarantee
// version parity in a way a global npm install never does.
package tools

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
