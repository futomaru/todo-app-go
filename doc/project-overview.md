# Project Overview

## Purpose

`todo-app-go` is a Todo management REST API written in Go, built **for learning purposes**.
The goal is not the app itself but what building it teaches: how to structure a layered web
backend in idiomatic Go using almost nothing but the standard library — routing, validation,
error handling, persistence, logging, and server lifecycle, all assembled by hand instead of
being provided by a framework.

Because the project is a learning vehicle, every design choice favors clarity over cleverness,
and the *reason* behind a choice matters as much as the choice itself.

## Scope (MVP)

The API manages a single shared list of todos. It provides exactly six use cases:

1. List todos (optionally filtered by completion state)
2. Get a single todo
3. Create a todo
4. Partially update a todo (PATCH semantics)
5. Delete a single todo
6. Delete all completed todos ("clear completed")

Deliberately **out of scope**: authentication/authorization, users, pagination, sorting
options, keyword search, a frontend, and deployment concerns. Keeping the feature set minimal
keeps the essential structure of the backend visible.

## Tech Stack

| Area | Choice |
|---|---|
| Language | Go 1.26 |
| HTTP | `net/http` standard library (`ServeMux` method + path pattern routing) |
| Database | SQLite, file-based (`./data/todo.db`) |
| DB driver | `modernc.org/sqlite` (pure Go, no CGO) — the **only** external dependency |
| DB access | `database/sql` + raw SQL |
| Validation | hand-written functions |
| Logging | `log/slog` (structured, JSON output) |
| Testing | standard `testing` package + `net/http/httptest` |

## Principles

- **Standard library first.** If the standard library can do the job, no external dependency
  is added. The single exception is the SQLite driver, which cannot be replaced by the
  standard library.
- **No framework, no ORM, no DI container, no mocking library.** These tools hide exactly the
  mechanics this project exists to teach. Writing the routing, SQL, wiring, and test fakes by
  hand makes the underlying model of a web backend explicit.
- **Small and simple over complete.** A minimal feature set and plain, readable code are
  preferred to abstraction built "for the future". An interface with only one conceivable
  implementation is not introduced ahead of need.
- **Explain the why.** Design decisions are documented with their rationale, so the project
  reads as a worked example rather than a black box.
- **Explicit over implicit.** Cross-cutting behavior that frameworks usually provide silently
  (request logging, panic recovery, graceful shutdown, error-to-HTTP conversion) is written
  out as ordinary, visible code.

## Running

```bash
go run ./cmd/todoapp   # start the server (http://localhost:8080)
go build ./...         # build
go test ./...          # run tests
```
