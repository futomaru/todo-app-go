# todo-app-go

A Todo management REST API written in Go.

> [!NOTE]
> 📚 This repository is for learning purposes.

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

