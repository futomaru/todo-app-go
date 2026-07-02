# Architecture

`todo-app-go` is a small REST API for managing todos, built on the Go standard library alone (the only external dependency is the SQLite driver) — no web framework, ORM, or DI container.

## Layers

Requests flow one-way, top to bottom: `handler → service → repository`.

```mermaid
flowchart TB
    Client([HTTP Client])

    subgraph Server["net/http Server"]
        direction LR
        Recoverer["Recoverer<br/>(panic → 500)"] --> Logger["Request Logger"] --> Mux["ServeMux<br/>(routing)"]
    end

    Handler["handler<br/>HTTP mapping · validation"]
    Service["service<br/>business logic"]
    Repository["repository<br/>raw SQL"]
    DB[("SQLite<br/>todos table")]

    Client -->|JSON over HTTP| Server
    Mux --> Handler --> Service --> Repository --> DB
```

| Layer | Responsibility |
|---|---|
| `handler` | HTTP mapping, request validation, status codes |
| `service` | business logic, existence checks, entity ↔ DTO conversion |
| `repository` | executes SQL |

Each layer defines the interface it needs from the layer below (Go convention: consumer declares the interface, not the implementer). 
All concrete types are created once, in `cmd/todoapp/main.go`, and passed down via constructors — this is also what makes each layer easy to test with a small fake, no mocking library needed.

## Request lifecycle

Example: creating a todo.

```mermaid
sequenceDiagram
    participant C as Client
    participant H as handler
    participant S as service
    participant R as repository
    participant D as SQLite

    C->>H: POST /api/v1/todos
    H->>H: decode + validate (400 on failure)
    H->>S: Create(req)
    S->>S: build Todo (completed=false, createdAt=updatedAt=now)
    S->>R: Insert(todo)
    R->>D: INSERT INTO todos ...
    D-->>R: generated id
    R-->>S: id
    S-->>H: TodoResponse
    H-->>C: 201 Created + Location header
```

Reads follow the same path minus the write step.
`PATCH`/`DELETE` add an existence check in `service`, returning a "not found" error if the todo is missing.

## Error handling

No framework-level exception handler — every error passes through one function, `apperror.WriteError`, which converts it to an `application/problem+json` response (e.g. `{"status": 404, "detail": "Todo not found: 1", ...}`):

```mermaid
flowchart TD
    E["error from any layer"] --> W["apperror.WriteError"]
    W --> Q1{"ErrNotFound?"}
    Q1 -- yes --> R404["404 Not Found"]
    Q1 -- no --> Q2{"*ValidationError?"}
    Q2 -- yes --> R400["400 Bad Request<br/>+ field errors"]
    Q2 -- no --> R500["500 Internal Server Error<br/>(details logged, not exposed)"]
```

## Cross-cutting concerns

| Concern | Mechanism |
|---|---|
| Structured logging | `log/slog`, JSON output, injected (not global) |
| Panic recovery | `Recoverer` middleware — recovers, logs, returns 500 |
| Graceful shutdown | `signal.NotifyContext` + `http.Server.Shutdown` (10s drain) |
| DB concurrency | single connection + SQLite `busy_timeout`/WAL pragmas |

## Package layout

```
todo-app-go/
├── cmd/todoapp/main.go   # composition root: wiring + startup
└── internal/
    ├── handler/          # HTTP mapping
    ├── middleware/       # request logging, panic recovery
    ├── service/          # use cases
    ├── repository/       # SQL execution
    ├── model/            # domain entity (Todo)
    ├── dto/              # request/response structs + validation
    ├── apperror/         # error types + problem+json conversion
    └── clock/            # time source abstraction
```

## Tech stack

| Area | Choice |
|---|---|
| Language | Go 1.26 |
| HTTP | `net/http` (`ServeMux` method + path routing) |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGO) |
| DB access | `database/sql` + raw SQL (no ORM) |
| Validation | hand-written functions |
| Logging | `log/slog` (JSON) |
