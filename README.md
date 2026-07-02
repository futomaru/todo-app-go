# todo-app-go

A project that implements a Todo management REST API in Go.

> [!NOTE]
> This is a repository for learning purposes.

## Tech Stack

| Area | Technology |
|---|---|
| Language | Go 1.26 |
| Web | `net/http` standard library |
| Database | SQLite (`modernc.org/sqlite`) |
| DB access | `database/sql` + raw SQL |
| Validation | Hand-written |

The only external dependency is the database driver; everything else relies on the standard library.

## Usage

```bash
go run ./cmd/todoapp   # Start the server (http://localhost:8080)
go build ./...         # Build
go test ./...          # Run tests
```

