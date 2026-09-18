package http

import (
	"encoding/json"
	"log/slog"
	nethttp "net/http"
)

// writeJSON serialises a payload with the right headers and status.
//
// Only the two error-handler paths in router.go need this -- every successful
// response is serialised by the generated code. It exists because those
// handlers receive a raw ResponseWriter rather than returning a typed value.
//
// ORDER MATTERS: Content-Type must be set BEFORE WriteHeader. Headers written
// after the status line are silently discarded, and the response goes out as
// text/plain with a JSON body. A classic Go gotcha with no warning attached.
func writeJSON(w nethttp.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// The status line is already sent, so the response cannot be corrected.
		// All that is left is to record it.
		slog.Error("failed to encode response body", "error", err)
	}
}
