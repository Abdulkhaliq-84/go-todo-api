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

## 2. PUT vs PATCH semantics — `internal/todo/app/commands.go`

`UpdateTodoCommand` currently uses pointers, including a `**time.Time`, to
distinguish absent from null. That is genuinely awkward.

- **Keep PATCH** — correct partial-update semantics, awkward types.
- **Switch to PUT** — client sends the whole object every time. Plain fields,
  no pointers, much simpler code, slightly ruder API.

---

## 3. `Save()` vs `Insert()`/`Update()` — `internal/todo/postgres/repository.go`

- **One `Save`** with `INSERT ... ON CONFLICT (id) DO UPDATE` — the repository
  interface stays small and the caller never thinks about it.
- **Separate methods** — clearer SQL, and an insert of an existing ID fails
  loudly instead of silently overwriting.

---

## 4. Error → status mapping — `internal/todo/http/errors.go`

Which domain errors are 400, which are 409, which are 404. Worth deciding
explicitly rather than accumulating cases as you hit them.

Also: what does an unrecognised error return? Never the raw message — log the
real error server-side, return something generic to the client.

---

## 5. Config strictness — `internal/platform/config/config.go`

Missing `DATABASE_URL`: crash at startup, or fall back to a localhost default?

- **Fail fast** — a misconfigured production deploy dies immediately and
  visibly, instead of quietly connecting somewhere wrong.
- **Default** — smoother first-run experience for anyone cloning the repo.

---

## 6. Time handling — `internal/todo/app/dto.go`

`IsOverdue` needs a `now`. Calling `time.Now()` inside the domain makes tests
time-dependent and flaky. Threading a `Clock` interface through is testable but
adds plumbing to every call site.
