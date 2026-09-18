package http

import "time"

// Request DTOs. THIS is where json tags belong -- not on the domain entity.
//
// Coming from FastAPI, this is the file that feels like the biggest loss:
// there is no Pydantic, so no automatic parsing, coercion, or 422 response.
// You decode, then you validate, by hand. The upside is that the validation is
// ordinary Go you can read, step through in a debugger, and unit test.

type createTodoRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

// validate performs TRANSPORT-level checks only: is the JSON well-formed, are
// required fields present, do the types make sense.
//
// It must NOT check business rules. "Is the title under 200 characters?" is a
// domain rule and lives in domain.NewTitle -- duplicating it here means two
// places to change and two chances to disagree.
//
// TODO(you): check that required fields are present, nothing more.
func (r createTodoRequest) validate() error {
	return nil // TODO
}

// updateTodoRequest uses pointers so an omitted field is distinguishable from
// a field explicitly set to its zero value. Without the pointer, a JSON body of
// {"description": ""} and a body with no description at all both arrive as "".
type updateTodoRequest struct {
	Title       *string     `json:"title"`
	Description *string     `json:"description"`
	DueDate     *time.Time  `json:"due_date"`
}

// TODO(you)
func (r updateTodoRequest) validate() error {
	return nil // TODO
}

// decodeJSON reads and parses a request body.
//
// TODO(you): use json.NewDecoder. Consider DisallowUnknownFields() -- it
// rejects typos like {"titel": "..."} instead of silently ignoring them.
// Also consider http.MaxBytesReader so a huge body cannot exhaust memory.
func decodeJSON(r any, body any) error {
	return nil // TODO
}
