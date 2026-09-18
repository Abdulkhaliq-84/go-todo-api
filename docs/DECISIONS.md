# Open design decisions

The scaffold deliberately leaves these unresolved. Each one has more than one
defensible answer, and picking one is the part worth thinking about — the rest
is typing.

---

## 1. Completion state transitions — `internal/todo/domain/todo.go`

`Complete()` on an already-completed todo: error, or silent no-op?

- **Error (`ErrAlreadyComplete`)** — strict. The caller learns something
  unexpected happened. Costs you a 409 path in the HTTP layer.
- **Idempotent no-op** — forgiving. Double-clicking a checkbox does not produce
  an error dialog. Matches how HTTP thinks about idempotency.

Whatever you choose, `Reopen()` must mirror it.

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
