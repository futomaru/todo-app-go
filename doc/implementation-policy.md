# Implementation Policy

How the code is written, ordered, and tested.

## Build Bottom-Up

Implementation proceeds from the packages with the fewest dependencies to the ones with the
most — leaves first, entry point last:

| Order | Package | Depends on |
|---|---|---|
| 1 | `model` | nothing |
| 2 | `apperror` | `net/http` only |
| 3 | `clock` | `time` only |
| 4 | `dto` | `model`, `apperror` |
| 5 | `repository` | `model`, `database/sql` |
| 6 | `service` | `model`, `dto`, `apperror`, `clock` (+ its own repository interface) |
| 7 | `handler` | `dto`, `apperror` (+ its own service interface) |
| 8 | `middleware` | `apperror`, `log/slog` |
| 9 | `cmd/todoapp` | everything |

After every step, `go build ./...` must pass and existing tests must stay green.

**Why bottom-up?** Each layer is written against dependencies that already exist and are
already tested, so there is never a pile of unverified code. At any moment it is obvious
which part of the system is done and which is next.

## Coding Conventions

- **Layout.** Entry point in `cmd/<app>/main.go`, implementation in `internal/<layer>/`.
  Package names are short, lowercase, singular. No `utils` or `common` packages.
- **`context.Context` is the first parameter** of every function that performs I/O
  (all repository and service methods). Handlers pass `r.Context()`. A context is never
  stored in a struct field.
- **Absent vs. zero is a pointer.** Optional or nullable values use `*T`, where `nil` means
  "not specified". Dereference only after a `nil` check.
- **Raw SQL only.** Queries are written by hand and rows are mapped to structs with explicit
  `Scan` calls. No ORM or query builder — seeing the SQL is part of the point.
- **JSON rules.** Field names are fixed with struct tags. An empty list is returned as `[]`,
  never `null` (allocate an empty slice; a `nil` slice marshals to `null`). `omitempty` only
  on genuinely optional response fields. Times are UTC, RFC 3339.
- **Error style.** Never discard an error. Wrap with `fmt.Errorf("...: %w", err)` when adding
  context. Error messages are lowercase, no trailing period. `panic` is not used for business
  flow — recovery exists in exactly one place, the outermost middleware.
- **Formatting and checks.** `gofmt`/`goimports` clean, `go vet` clean, and the race detector
  run before committing.

## Testing Policy

Tests use the **standard `testing` package only** — no mocking, assertion, or test-framework
libraries.

| Target | Technique |
|---|---|
| `repository` | real SQLite (`:memory:` DSN, max 1 open connection) executing real SQL |
| `service` | table-driven tests + hand-written fake repository + fixed clock |
| `handler` | `net/http/httptest` + hand-written fake service |
| `dto` / `apperror` | table-driven tests of pure functions |
| integration (optional) | `httptest.NewServer` + temporary SQLite file, a few end-to-end scenarios |

- **Test the lower layers thoroughly, keep integration tests thin.** Most behavior is
  cheapest to pin down close to where it lives.
- **Table-driven tests** with descriptive case names, run as subtests via `t.Run`.
  Boundary values (e.g. a title of exactly the maximum length, in multi-byte characters)
  and error paths are always included.
- **Fakes, not mocks.** Because each interface is declared by its consumer and is minimal,
  a fake is a few lines of struct — small enough that a generation tool or mock library
  would add more machinery than it removes, and the fake itself documents what the consumer
  expects from its dependency.
- Commands:

```bash
go test ./...          # all tests
go test ./... -race    # race detector — required before committing
go test ./... -cover   # coverage overview (a guide, not a target)
```

## Git Conventions

- **Branches:** `<type>/<short-description>` in English kebab-case, where `<type>` is one of
  `feature`, `fix`, `refactor`, `docs`, `test`, `chore`, `hotfix`. Never commit directly to
  `main`.
- **Commits:** Conventional Commits format — `<type>(<scope>): <subject>`, with the same type
  vocabulary. Subject in imperative mood, lowercase, no trailing period. The body explains
  *why* the change was needed; the diff already shows *what* changed.
- One commit = one logical unit. Unrelated changes are never mixed.
- Commit messages are concise English, carry no signatures or trailers.
