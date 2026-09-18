# Open design decisions

The scaffold deliberately leaves these unresolved. Each one has more than one
defensible answer, and picking one is the part worth thinking about — the rest
is typing.

---

## 1. Completion state transitions — DECIDED: strict

`Complete()` on an already-completed todo returns `ErrAlreadyComplete`.
`Reopen()` on an active todo returns `ErrNotCompleted`. Neither is a no-op.

```
        Complete()                  Reopen()
 ACTIVE ──────────▶ COMPLETED   COMPLETED ──────────▶ ACTIVE
```

**Why strict:**

1. `api/openapi.yaml` already promises `409` on both endpoints. Idempotent
   completion makes that response unreachable and the spec untrue — and the
   point of spec-first is that it cannot be.
2. An aggregate exists to refuse invalid transitions. Accepting one silently
   moves the rule into the category of things nothing enforces.
3. It exercises the whole architecture: a domain error propagates untouched
   through the service and is mapped to a status in exactly one file. An
   idempotent version never travels that path, so the layering is never tested.

**Rejected alternative — idempotent no-op.** The argument for it is the
double-clicked checkbox. That is a client-side state problem: disable the
button, or only call `complete` when `completed == false`. Weakening a domain
invariant to compensate for UI state is the wrong direction of fix.

**Consequences, all already in place:**

| Where | What |
|---|---|
| `domain/errors.go` | `ErrAlreadyComplete`, `ErrNotCompleted` exist |
| `app/service.go` | Must propagate them **unwrapped**, or `errors.Is` stops matching and clients get 500s |
| `http/errors.go` | Both map to `409` |
| `api/openapi.yaml` | `409` documented on complete and reopen — no change needed |

**Deliberately excluded from the rule:**

- Due date does not gate completion. An overdue todo can still be completed;
  lateness is an observation, not a permission.
- Reopening does not clear the due date. A todo reopened past its deadline is
  immediately overdue again, which is truthful. Clearing it would be the domain
  inventing a rule nobody asked for — `Reschedule` exists for that, visibly.

---

## 2. Update semantics — DECIDED: PATCH with an explicit clear flag

Only fields the client sent are changed. `UpdateTodoCommand` carries pointers,
plus a `ClearDueDate bool`:

| `DueDate` | `ClearDueDate` | Meaning |
|---|---|---|
| `nil` | `false` | absent — unchanged |
| `&t` | `false` | set to `t` |
| `nil` | `true` | cleared |

**Why not `**time.Time`.** The double pointer encodes the same three states
with no extra field and is correct — and unreadable. Every caller has to reason
about two levels of indirection to answer one question.

**Why not PUT.** Full replacement gives plain non-pointer fields and much
simpler code, but forces every client into read-then-write and makes concurrent
edits clobber each other silently.

**Consequence:** `nullable-type: true` in `api/oapi-codegen.yaml`, so the wire
type is `nullable.Nullable[time.Time]` rather than `*time.Time`. Without it,
"absent" and "null" both arrive as `nil` and a PATCH that omits `due_date`
would silently clear it. `http/mapping.go` resolves the three wire states into
the two command fields.

---

## 3. Persistence strategy — DECIDED: one `Save()` doing an upsert

```sql
INSERT INTO todos (...) VALUES ($1, ...)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, ...
```

The repository interface stays at four methods and the service never tracks
whether an entity is new — that is a persistence concern, and it stays behind
the boundary.

**Trade accepted:** inserting an existing ID overwrites instead of failing
loudly. Tolerable because IDs are domain-minted UUIDs, so a collision implies a
bug that a unique-violation error would not meaningfully mitigate.

**Watch out:** `created_at` must not appear in the `DO UPDATE SET` list.

---

## 4. Error → status mapping — DECIDED: coarse codes

| Domain error | Status | Code |
|---|---|---|
| `ErrNotFound` | 404 | `not_found` |
| `ErrTitleEmpty`, `ErrTitleTooLong` | 400 | `invalid_title` |
| `ErrInvalidID` | 400 | `invalid_id` |
| `ErrAlreadyComplete` | 409 | `already_completed` |
| `ErrNotCompleted` | 409 | `not_completed` |
| `ErrDueDateInPast` | 400 | `invalid_due_date` |
| anything else | 500 | `internal_error` |

The `error` code is part of the public contract — clients branch on it, so
renaming one is a breaking change like renaming a JSON field.

**Coarse chosen.** `invalid_title` covers both the empty and the too-long case;
the `message` carries the specifics for a human. Fewer strings frozen into the
contract forever, at the cost of a client not being able to distinguish the two
title failures without reading prose.

Rejected: one code per domain error (more client-friendly, permanently larger
contract surface), and adding a `field` to `ErrorResponse` (two things to keep
in sync instead of one).

Note `ErrDueDateInPast` was **removed** rather than mapped. A past due date is
allowed — logging a task you already missed is normal, and importing history
would be impossible otherwise. A sentinel the code never returns implies a rule
that does not exist.

---

## 5. Config strictness — DECIDED: fail fast

A missing or unparseable `DATABASE_URL` is an error. No localhost fallback.

A production deploy with the variable unset should die immediately and visibly.
The alternative failure — a service that starts, reports healthy, and is quietly
pointed at the wrong database — is far more expensive to diagnose.

Everything else gets a default. Ports and timeouts have obviously right values;
a database URL does not.

`Load()` returns an error rather than calling `log.Fatal`. It is a library
function; the decision to exit belongs to `main`.

---

## 6. Time handling — DECIDED: pass `now` as a parameter

`IsOverdue(now time.Time)`, and the app layer's mappers take `now` too. The
service calls `time.Now()` once per request and threads it down.

Keeps the domain a pure function of its inputs: a test can ask "is this overdue
as of next Tuesday?" without freezing a clock or sleeping. It also means every
todo in a list is evaluated against the same instant rather than each against a
slightly different one.

**Why not a `Clock` interface.** More ceremony than a two-state domain needs.
Worth revisiting if recurring or scheduled todos are ever added.

**Why not `time.Now()` inside the domain.** Makes the domain impure and every
date-related test time-dependent — the usual source of tests that fail only at
midnight or only in CI.

---

---

## 7. Does a completed todo stay overdue? — DECIDED: no

`IsOverdue` returns `false` once a todo is completed, however late it was
finished.

"Overdue" means **needs attention**, and a finished task needs none. The
practical consequence is that `?overdue=true` returns exactly the list a user
should act on, with no client-side filtering — and no client can forget to
apply it.

**Trade accepted:** lateness becomes unrecoverable from the API. Nothing records
that a task was finished after its deadline.

**Rejected:** a pure date comparison (purest domain, but every client repeats
`overdue && !completed` and one will forget), and carrying both `overdue` and
`completed_late` (most expressive, but a new field across four files for a
feature nobody has asked for).

---

## All seven settled

Nothing is blocking implementation.
