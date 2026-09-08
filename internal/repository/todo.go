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

func (r *TodoRepository) FindAll(ctx context.Context) ([]model.Todo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, description, completed, created_at, updated_at FROM todos ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []model.Todo
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return todos, nil
}

// FindByCompleted returns todos filtered by their completed flag, ordered
// by id ascending.
func (r *TodoRepository) FindByCompleted(ctx context.Context, completed bool) ([]model.Todo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, description, completed, created_at, updated_at FROM todos WHERE completed = ? ORDER BY id`,
		completed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []model.Todo
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return todos, nil
}

// FindByID returns the todo with the given id. sql.ErrNoRows propagates
// untouched — translating "does not exist" to apperror.ErrNotFound is
// internal/service's job (STEP 7), not this layer's.
func (r *TodoRepository) FindByID(ctx context.Context, id int64) (model.Todo, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, title, description, completed, created_at, updated_at FROM todos WHERE id = ?`, id)
	return scanTodo(row)
}

// Insert creates a new todo row and returns its generated id.
func (r *TodoRepository) Insert(ctx context.Context, t model.Todo) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO todos (title, description, completed, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		t.Title, t.Description, t.Completed, formatTime(t.CreatedAt), formatTime(t.UpdatedAt))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Update overwrites an existing todo's mutable fields. created_at is
// deliberately left untouched — it is immutable after Insert.
func (r *TodoRepository) Update(ctx context.Context, t model.Todo) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todos SET title = ?, description = ?, completed = ?, updated_at = ? WHERE id = ?`,
		t.Title, t.Description, t.Completed, formatTime(t.UpdatedAt), t.ID)
	return err
}

// DeleteByID deletes a single todo and reports how many rows were removed
// (0 or 1) so the caller (service) can decide whether that means
// "not found".
func (r *TodoRepository) DeleteByID(ctx context.Context, id int64) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM todos WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteCompleted deletes every completed todo and returns how many rows
// were removed. Zero is a valid, non-error result (nothing was completed).
func (r *TodoRepository) DeleteCompleted(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM todos WHERE completed = 1`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
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
