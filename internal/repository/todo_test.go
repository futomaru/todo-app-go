package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/futomaru/todo-app-go/internal/model"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// :memory: databases are per-connection; force everything onto one
	// connection so all queries in a test see the same schema and data.
	db.SetMaxOpenConns(1)

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return db
}

func strPtr(s string) *string { return &s }

func TestInsertAndFindByID(t *testing.T) {
	db := newTestDB(t)
	repo := New(db)
	ctx := context.Background()

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	cases := []struct {
		name string
		todo model.Todo
	}{
		{
			name: "with description",
			todo: model.Todo{Title: "buy milk", Description: strPtr("2%"), Completed: false, CreatedAt: now, UpdatedAt: now},
		},
		{
			name: "nil description",
			todo: model.Todo{Title: "walk dog", Description: nil, Completed: true, CreatedAt: now, UpdatedAt: now},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := repo.Insert(ctx, tc.todo)
			if err != nil {
				t.Fatalf("Insert: %v", err)
			}

			got, err := repo.FindByID(ctx, id)
			if err != nil {
				t.Fatalf("FindByID: %v", err)
			}

			if got.ID != id {
				t.Errorf("ID = %d, want %d", got.ID, id)
			}
			if got.Title != tc.todo.Title {
				t.Errorf("Title = %q, want %q", got.Title, tc.todo.Title)
			}
			if (got.Description == nil) != (tc.todo.Description == nil) {
				t.Errorf("Description nil-ness mismatch: got %v, want %v", got.Description, tc.todo.Description)
			} else if got.Description != nil && *got.Description != *tc.todo.Description {
				t.Errorf("Description = %q, want %q", *got.Description, *tc.todo.Description)
			}
			if got.Completed != tc.todo.Completed {
				t.Errorf("Completed = %v, want %v", got.Completed, tc.todo.Completed)
			}
			if !got.CreatedAt.Equal(tc.todo.CreatedAt) {
				t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, tc.todo.CreatedAt)
			}
			if !got.UpdatedAt.Equal(tc.todo.UpdatedAt) {
				t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, tc.todo.UpdatedAt)
			}
			if got.CreatedAt.Location() != time.UTC {
				t.Errorf("CreatedAt not UTC: %v", got.CreatedAt.Location())
			}
		})
	}
}

func TestFindByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := New(db)

	_, err := repo.FindByID(context.Background(), 999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestFindAll_OrderedByID(t *testing.T) {
	db := newTestDB(t)
	repo := New(db)
	ctx := context.Background()
	now := time.Now().UTC()

	var ids []int64
	for _, title := range []string{"c", "a", "b"} {
		id, err := repo.Insert(ctx, model.Todo{Title: title, CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatalf("Insert: %v", err)
		}
		ids = append(ids, id)
	}

	got, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(got) != len(ids) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(ids))
	}
	for i, id := range ids {
		if got[i].ID != id {
			t.Errorf("got[%d].ID = %d, want %d (order not ascending by id)", i, got[i].ID, id)
		}
	}
}

func TestFindByCompleted(t *testing.T) {
	db := newTestDB(t)
	repo := New(db)
	ctx := context.Background()
	now := time.Now().UTC()

	doneID, err := repo.Insert(ctx, model.Todo{Title: "done", Completed: true, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if _, err := repo.Insert(ctx, model.Todo{Title: "pending", Completed: false, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	got, err := repo.FindByCompleted(ctx, true)
	if err != nil {
		t.Fatalf("FindByCompleted: %v", err)
	}
	if len(got) != 1 || got[0].ID != doneID {
		t.Fatalf("FindByCompleted(true) = %+v, want only id=%d", got, doneID)
	}
}

func TestUpdate(t *testing.T) {
	db := newTestDB(t)
	repo := New(db)
	ctx := context.Background()
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)

	id, err := repo.Insert(ctx, model.Todo{Title: "old", Completed: false, CreatedAt: created, UpdatedAt: created})
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}

	err = repo.Update(ctx, model.Todo{ID: id, Title: "new", Description: strPtr("d"), Completed: true, CreatedAt: created, UpdatedAt: updated})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Title != "new" || !got.Completed || got.Description == nil || *got.Description != "d" {
		t.Errorf("Update did not apply: %+v", got)
	}
	if !got.UpdatedAt.Equal(updated) {
		t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, updated)
	}
	if !got.CreatedAt.Equal(created) {
		t.Errorf("CreatedAt changed: got %v, want %v (must stay immutable)", got.CreatedAt, created)
	}
}

func TestDeleteByID(t *testing.T) {
	db := newTestDB(t)
	repo := New(db)
	ctx := context.Background()
	now := time.Now().UTC()

	id, err := repo.Insert(ctx, model.Todo{Title: "x", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}

	affected, err := repo.DeleteByID(ctx, id)
	if err != nil {
		t.Fatalf("DeleteByID: %v", err)
	}
	if affected != 1 {
		t.Errorf("affected = %d, want 1", affected)
	}

	affected, err = repo.DeleteByID(ctx, id)
	if err != nil {
		t.Fatalf("DeleteByID (second): %v", err)
	}
	if affected != 0 {
		t.Errorf("affected = %d, want 0 for already-deleted id", affected)
	}
}

func TestDeleteCompleted(t *testing.T) {
	db := newTestDB(t)
	repo := New(db)
	ctx := context.Background()
	now := time.Now().UTC()

	if _, err := repo.Insert(ctx, model.Todo{Title: "done1", Completed: true, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if _, err := repo.Insert(ctx, model.Todo{Title: "done2", Completed: true, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	pendingID, err := repo.Insert(ctx, model.Todo{Title: "pending", Completed: false, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}

	affected, err := repo.DeleteCompleted(ctx)
	if err != nil {
		t.Fatalf("DeleteCompleted: %v", err)
	}
	if affected != 2 {
		t.Errorf("affected = %d, want 2", affected)
	}

	remaining, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != pendingID {
		t.Fatalf("remaining = %+v, want only id=%d", remaining, pendingID)
	}

	affected, err = repo.DeleteCompleted(ctx)
	if err != nil {
		t.Fatalf("DeleteCompleted (empty): %v", err)
	}
	if affected != 0 {
		t.Errorf("affected = %d, want 0 when nothing completed", affected)
	}
}
