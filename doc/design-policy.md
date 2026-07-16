# Design Policy

How the backend is structured, and why.

## Layered Architecture

Requests flow one way, top to bottom. No layer ever calls back up.

```
HTTP client
   │
   ▼
middleware   — request logging, panic recovery (cross-cutting only)
   │
   ▼
handler      — HTTP mapping, decoding, validation calls, status codes
   │            calls service through the TodoService interface
   ▼
service      — use cases, existence checks, timestamps, entity ↔ DTO conversion
   │            calls repository through the TodoRepository interface
   ▼
repository   — executes raw SQL (database/sql)
   │
   ▼
SQLite       — single `todos` table
```

| Layer | Responsibility | Explicitly not its job |
|---|---|---|
| handler | HTTP mapping, request decoding, validation calls, status codes, `Location` header | business logic, SQL |
| service | use cases, not-found checks, setting timestamps, entity ↔ DTO conversion | anything HTTP-related |
| repository | executing SQL, mapping rows to structs | branching or transforming data in Go |

**Why layers at all, in an app this small?** Each concern (HTTP, business rules, SQL) changes
for different reasons and is tested with different techniques. Separating them keeps every
file single-purpose and makes each layer testable in isolation.

## Package Layout

```
cmd/todoapp/main.go   # composition root: builds every object, wires them, starts the server
internal/
├── handler/          # HTTP handlers + TodoService interface definition
├── middleware/       # request logging, panic recovery
├── service/          # business logic + TodoRepository interface definition
├── repository/       # SQL execution
├── model/            # domain entity (Todo)
├── dto/              # request/response structs + validation functions
├── apperror/         # error types + problem+json conversion
└── clock/            # time source abstraction
```

Everything lives under `internal/`, so no other module can import it — the conventional Go
way to separate a program's implementation from any public API surface.

## Interfaces Belong to the Consumer

The key Go convention this project practices: **an interface is declared by the code that
uses it, not by the code that implements it**, and contains only the methods the consumer
actually needs.

| Interface | Declared in (consumer) | Implemented in (provider) |
|---|---|---|
| `TodoService` | `internal/handler` | `internal/service` |
| `TodoRepository` | `internal/service` | `internal/repository` |
| `Clock` | `internal/clock` | `internal/clock` (system), tests (fixed) |

Because concrete types satisfy interfaces implicitly, this arrangement means lower packages
never import higher ones — the dependency arrow points strictly downward. It also makes
testing trivial: a test replaces a dependency by writing a tiny struct that satisfies the
interface, with no mocking library.

## Dependency Injection by Hand

All concrete types are created once, in `cmd/todoapp/main.go`, and passed down through
constructors (`repository.New`, `service.New`, `handler.New`). No DI framework, no globals,
no `init()` side effects. The full object graph of the application is readable in one place.

## Data Model

One table, no relations:

```sql
CREATE TABLE IF NOT EXISTS todos (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    title       TEXT NOT NULL,
    description TEXT,
    completed   INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);
```

```go
type Todo struct {
    ID          int64
    Title       string
    Description *string   // nil represents SQL NULL
    Completed   bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

- **NULL / "not specified" is a pointer.** `*string` (and in update requests `*bool`) uses
  `nil` to mean "absent", cleanly distinguishing it from an explicit zero value.
  `database/sql` converts NULL ↔ `nil` automatically when scanning into `*string`.
- **Timestamps are set in the service layer**, not by DB defaults, via an injected `Clock`
  interface. This keeps time a controllable dependency: tests inject a fixed clock and can
  assert exact values.
- **Times are UTC, serialized as RFC 3339** with timezone offset (e.g.
  `2026-06-02T10:00:00Z`) — the default `encoding/json` behavior for `time.Time`.
- The schema is created at startup with `CREATE TABLE IF NOT EXISTS` (idempotent), so the app
  needs no separate migration step.

## Error Handling

Errors travel up the layers as ordinary `error` values and are converted to HTTP exactly
once, at the boundary, by a single function: `apperror.WriteError`.

- **Sentinel error for classification only:** `apperror.ErrNotFound`, checked with
  `errors.Is` → `404`.
- **Typed error when data must travel with it:** `*apperror.ValidationError` carrying
  per-field messages, checked with `errors.As` → `400` with an `errors` array.
- **Everything else** → `500`. Internal details are written to the log only — the response
  body carries a generic message, never stack traces or driver errors.

All error responses use the `application/problem+json` format (RFC 7807):

```json
{
  "type": "about:blank",
  "title": "Not Found",
  "status": 404,
  "detail": "Todo not found: 1",
  "instance": "/api/v1/todos/1"
}
```

**Why one conversion point?** Without a framework's global exception handler, scattering
status-code decisions across handlers would drift apart over time. Centralizing the mapping
keeps the error contract consistent and testable in one place.

## Cross-Cutting Concerns

Middleware uses the plain `func(http.Handler) http.Handler` shape and is chained in `main`,
outermost first: `Recoverer → RequestLogger → ServeMux`.

| Concern | Mechanism |
|---|---|
| Request logging | `RequestLogger` middleware logs method, path, status, duration via `slog` |
| Panic recovery | `Recoverer` middleware recovers, logs the stack, returns 500; placed outermost so nothing escapes |
| Structured logging | `log/slog` with a JSON handler, created in `main` and injected — not accessed as a global |
| Graceful shutdown | `signal.NotifyContext` catches SIGINT/SIGTERM; `http.Server.Shutdown` drains in-flight requests (10s limit) |
| Time | `Clock` interface abstracts `time.Now()` so services never read the wall clock directly |
