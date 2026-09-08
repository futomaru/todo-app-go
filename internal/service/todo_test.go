package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/futomaru/todo-app-go/internal/apperror"
	"github.com/futomaru/todo-app-go/internal/dto"
	"github.com/futomaru/todo-app-go/internal/model"
)

// fakeRepo is a minimal, in-memory stand-in for TodoRepository. Each *Err
// field lets a test force that method to fail without a real DB.
type fakeRepo struct {
	todos  map[int64]model.Todo
	nextID int64

	findAllErr         error
	findByCompletedErr error
	findByIDErr        error
	insertErr          error
	updateErr          error
	deleteByIDErr      error
	deleteCompletedErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{todos: make(map[int64]model.Todo)}
}

func (f *fakeRepo) FindAll(ctx context.Context) ([]model.Todo, error) {
	if f.findAllErr != nil {
		return nil, f.findAllErr
	}
	var out []model.Todo
	for _, t := range f.todos {
		out = append(out, t)
	}
	return out, nil
}

func (f *fakeRepo) FindByCompleted(ctx context.Context, completed bool) ([]model.Todo, error) {
	if f.findByCompletedErr != nil {
		return nil, f.findByCompletedErr
	}
	var out []model.Todo
	for _, t := range f.todos {
		if t.Completed == completed {
			out = append(out, t)
		}
	}
	return out, nil
}

func (f *fakeRepo) FindByID(ctx context.Context, id int64) (model.Todo, error) {
	if f.findByIDErr != nil {
		return model.Todo{}, f.findByIDErr
	}
	t, ok := f.todos[id]
	if !ok {
		return model.Todo{}, sql.ErrNoRows
	}
	return t, nil
}

func (f *fakeRepo) Insert(ctx context.Context, t model.Todo) (int64, error) {
	if f.insertErr != nil {
		return 0, f.insertErr
	}
	f.nextID++
	t.ID = f.nextID
	f.todos[t.ID] = t
	return t.ID, nil
}

func (f *fakeRepo) Update(ctx context.Context, t model.Todo) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.todos[t.ID] = t
	return nil
}

func (f *fakeRepo) DeleteByID(ctx context.Context, id int64) (int64, error) {
	if f.deleteByIDErr != nil {
		return 0, f.deleteByIDErr
	}
	if _, ok := f.todos[id]; !ok {
		return 0, nil
	}
	delete(f.todos, id)
	return 1, nil
}

func (f *fakeRepo) DeleteCompleted(ctx context.Context) (int64, error) {
	if f.deleteCompletedErr != nil {
		return 0, f.deleteCompletedErr
	}
	var n int64
	for id, t := range f.todos {
		if t.Completed {
			delete(f.todos, id)
			n++
		}
	}
	return n, nil
}

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

func strPtr(s string) *string { return &s }

func TestCreate(t *testing.T) {
	repo := newFakeRepo()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := New(repo, fixedClock{now})

	res, err := svc.Create(context.Background(), dto.TodoCreateRequest{Title: "buy milk", Description: strPtr("2%")})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res.ID == 0 {
		t.Error("ID not assigned")
	}
	if res.Completed {
		t.Error("Completed = true, want false on creation")
	}
	if !res.CreatedAt.Equal(now) || !res.UpdatedAt.Equal(now) {
		t.Errorf("CreatedAt/UpdatedAt = %v/%v, want both %v", res.CreatedAt, res.UpdatedAt, now)
	}
	if !res.CreatedAt.Equal(res.UpdatedAt) {
		t.Error("CreatedAt != UpdatedAt on creation")
	}
}

func TestCreate_RepoErrorPropagates(t *testing.T) {
	repo := newFakeRepo()
	repo.insertErr = errors.New("disk full")
	svc := New(repo, fixedClock{time.Now()})

	_, err := svc.Create(context.Background(), dto.TodoCreateRequest{Title: "x"})
	if !errors.Is(err, repo.insertErr) {
		t.Fatalf("err = %v, want repo error to propagate", err)
	}
}

func TestList(t *testing.T) {
	repo := newFakeRepo()
	now := time.Now().UTC()
	repo.todos[1] = model.Todo{ID: 1, Title: "a", Completed: true, CreatedAt: now, UpdatedAt: now}
	repo.todos[2] = model.Todo{ID: 2, Title: "b", Completed: false, CreatedAt: now, UpdatedAt: now}
	svc := New(repo, fixedClock{now})

	all, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List(nil): %v", err)
	}
	if len(all) != 2 {
		t.Errorf("len(all) = %d, want 2", len(all))
	}

	done := true
	filtered, err := svc.List(context.Background(), &done)
	if err != nil {
		t.Fatalf("List(true): %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != 1 {
		t.Errorf("filtered = %+v, want only id=1", filtered)
	}
}

func TestList_EmptyIsNonNilSlice(t *testing.T) {
	repo := newFakeRepo()
	svc := New(repo, fixedClock{time.Now()})

	res, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if res == nil {
		t.Error("List returned nil slice, want non-nil empty slice (so JSON is [] not null)")
	}
	if len(res) != 0 {
		t.Errorf("len(res) = %d, want 0", len(res))
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := New(repo, fixedClock{time.Now()})

	_, err := svc.GetByID(context.Background(), 999)
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("err = %v, want apperror.ErrNotFound", err)
	}
}

func TestGetByID_Found(t *testing.T) {
	repo := newFakeRepo()
	now := time.Now().UTC()
	repo.todos[1] = model.Todo{ID: 1, Title: "a", CreatedAt: now, UpdatedAt: now}
	svc := New(repo, fixedClock{now})

	res, err := svc.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if res.ID != 1 || res.Title != "a" {
		t.Errorf("res = %+v, want id=1 title=a", res)
	}
}

func TestUpdate_PartialFields(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)

	repo := newFakeRepo()
	repo.todos[1] = model.Todo{ID: 1, Title: "old", Description: strPtr("d"), Completed: false, CreatedAt: created, UpdatedAt: created}
	svc := New(repo, fixedClock{updated})

	newTitle := "new"
	res, err := svc.Update(context.Background(), 1, dto.TodoUpdateRequest{Title: &newTitle})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if res.Title != "new" {
		t.Errorf("Title = %q, want %q", res.Title, "new")
	}
	if res.Description == nil || *res.Description != "d" {
		t.Errorf("Description = %v, want unchanged %q", res.Description, "d")
	}
	if res.Completed {
		t.Error("Completed changed, want unchanged (false)")
	}
	if !res.CreatedAt.Equal(created) {
		t.Errorf("CreatedAt = %v, want unchanged %v", res.CreatedAt, created)
	}
	if !res.UpdatedAt.Equal(updated) {
		t.Errorf("UpdatedAt = %v, want %v", res.UpdatedAt, updated)
	}
}

func TestUpdate_CompletedFalseExplicit(t *testing.T) {
	repo := newFakeRepo()
	now := time.Now().UTC()
	repo.todos[1] = model.Todo{ID: 1, Title: "a", Completed: true, CreatedAt: now, UpdatedAt: now}
	svc := New(repo, fixedClock{now})

	completed := false
	res, err := svc.Update(context.Background(), 1, dto.TodoUpdateRequest{Completed: &completed})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if res.Completed {
		t.Error("Completed = true, want false to have been applied explicitly")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := New(repo, fixedClock{time.Now()})

	_, err := svc.Update(context.Background(), 999, dto.TodoUpdateRequest{})
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("err = %v, want apperror.ErrNotFound", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := New(repo, fixedClock{time.Now()})

	err := svc.Delete(context.Background(), 999)
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("err = %v, want apperror.ErrNotFound", err)
	}
}

func TestDelete_Found(t *testing.T) {
	repo := newFakeRepo()
	now := time.Now().UTC()
	repo.todos[1] = model.Todo{ID: 1, Title: "a", CreatedAt: now, UpdatedAt: now}
	svc := New(repo, fixedClock{now})

	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok := repo.todos[1]; ok {
		t.Error("todo still present after Delete")
	}
}

func TestDeleteCompleted_NoneIsSuccess(t *testing.T) {
	repo := newFakeRepo()
	svc := New(repo, fixedClock{time.Now()})

	if err := svc.DeleteCompleted(context.Background()); err != nil {
		t.Fatalf("DeleteCompleted with nothing completed: %v", err)
	}
}

func TestDeleteCompleted_RepoErrorPropagates(t *testing.T) {
	repo := newFakeRepo()
	repo.deleteCompletedErr = errors.New("boom")
	svc := New(repo, fixedClock{time.Now()})

	err := svc.DeleteCompleted(context.Background())
	if !errors.Is(err, repo.deleteCompletedErr) {
		t.Fatalf("err = %v, want repo error to propagate", err)
	}
}
