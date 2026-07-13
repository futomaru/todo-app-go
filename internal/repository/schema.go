package repository

import (
	"context"
	"database/sql"
)

// schema is the DDL for the todos table.
//
// TODO: Fill in the CREATE TABLE statement. Columns should mirror
// model.Todo 1:1 (id, title, description, completed, created_at, updated_at).
// - description is nullable (model.Todo.Description is *string)
// - created_at / updated_at are TEXT, not a native timestamp type — SQLite
//   has none, and this project stores time.Time as RFC3339Nano strings
//   (see formatTime/parseTime in todo.go).
// - Use "CREATE TABLE IF NOT EXISTS" so Migrate is safe to call on every
//   process startup (idempotent).
const schema = ``

// Migrate applies the schema DDL. It is idempotent and safe to call every
// time the process starts.
//
// TODO: Run `schema` via db.ExecContext(ctx, schema) and return its error.
// Use ExecContext (not Exec) so request/startup cancellation propagates.
func Migrate(ctx context.Context, db *sql.DB) error {
	panic("not implemented")
}
