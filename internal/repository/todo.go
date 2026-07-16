package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/futomaru/todo-app-go/internal/model"
)

type TodoRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// FindAll returns every todo ordered by id ascending.
//
// TODO: QueryContext a SELECT ... FROM todos ORDER BY id, then loop
// rows.Next()/scanTodo, and finally check rows.Err(). Remember to
// defer rows.Close().
func (r *TodoRepository) FindAll(ctx context.Context) ([]model.Todo, error) {
	panic("not implemented")
}

// FindByCompleted returns todos filtered by their completed flag, ordered
// by id ascending. Structurally identical to FindAll with one extra WHERE
// clause — reuse the same rows.Next()/scanTodo/rows.Err() shape.
//
// TODO: SELECT ... FROM todos WHERE completed = ? ORDER BY id.
func (r *TodoRepository) FindByCompleted(ctx context.Context, completed bool) ([]model.Todo, error) {
	panic("not implemented")
}

// FindByID returns the todo with the given id.
//
// TODO: SELECT ... FROM todos WHERE id = ?, QueryRowContext, scanTodo the
// *sql.Row. Important: do NOT translate sql.ErrNoRows here — let it
// propagate untouched. "Does not exist" is a business-layer concept;
// translating it to apperror.ErrNotFound is internal/service's job
// (STEP 7), not this layer's. Repository only speaks SQL vocabulary.
func (r *TodoRepository) FindByID(ctx context.Context, id int64) (model.Todo, error) {
	panic("not implemented")
}

func (r *TodoRepository) Insert(ctx context.Context, t model.Todo) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO todos (title, description, completed, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`, t.Title, t.Description, t.Completed, formatTime(t.CreatedAt), formatTime(t.UpdatedAt))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Update overwrites an existing todo's mutable fields.
//
// TODO: ExecContext an UPDATE that sets title/description/completed/
// updated_at (bind formatTime(t.UpdatedAt)) WHERE id = ?. Deliberately do
// NOT set created_at — it is immutable after Insert.
func (r *TodoRepository) Update(ctx context.Context, t model.Todo) error {
	panic("not implemented")
}

// DeleteByID deletes a single todo and reports how many rows were removed
// (0 or 1) so the caller (service, STEP 7) can decide whether that means
// "not found".
//
// TODO: ExecContext a DELETE ... WHERE id = ?, return res.RowsAffected().
func (r *TodoRepository) DeleteByID(ctx context.Context, id int64) (int64, error) {
	panic("not implemented")
}

// DeleteCompleted deletes every completed todo and returns how many rows
// were removed. Zero is a valid, non-error result (nothing was completed).
//
// TODO: ExecContext a DELETE ... WHERE completed = 1 (or = true), return
// res.RowsAffected().
func (r *TodoRepository) DeleteCompleted(ctx context.Context) (int64, error) {
	panic("not implemented")
}

func scanTodo(s interface{ Scan(dest ...any) error }) (model.Todo, error) {
	var t model.Todo
	var createdAt, updatedAt string

	err := s.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &createdAt, &updatedAt)
	if err != nil {
		return model.Todo{}, err
	}

	t.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return model.Todo{}, err
	}

	t.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return model.Todo{}, err
	}

	return t, nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}
