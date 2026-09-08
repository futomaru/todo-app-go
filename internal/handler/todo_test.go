package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/futomaru/todo-app-go/internal/apperror"
	"github.com/futomaru/todo-app-go/internal/dto"
)

// fakeService is a minimal, per-test-configurable stand-in for TodoService.
// Each field defaults to nil; a test only needs to set the method(s) it
// exercises, and calling an unset one panics loudly (a real bug, not a
// silently-passing test).
type fakeService struct {
	listFunc            func(ctx context.Context, completed *bool) ([]dto.TodoResponse, error)
	getByIDFunc         func(ctx context.Context, id int64) (dto.TodoResponse, error)
	createFunc          func(ctx context.Context, req dto.TodoCreateRequest) (dto.TodoResponse, error)
	updateFunc          func(ctx context.Context, id int64, req dto.TodoUpdateRequest) (dto.TodoResponse, error)
	deleteFunc          func(ctx context.Context, id int64) error
	deleteCompletedFunc func(ctx context.Context) error
}

func (f *fakeService) List(ctx context.Context, completed *bool) ([]dto.TodoResponse, error) {
	return f.listFunc(ctx, completed)
}
func (f *fakeService) GetByID(ctx context.Context, id int64) (dto.TodoResponse, error) {
	return f.getByIDFunc(ctx, id)
}
func (f *fakeService) Create(ctx context.Context, req dto.TodoCreateRequest) (dto.TodoResponse, error) {
	return f.createFunc(ctx, req)
}
func (f *fakeService) Update(ctx context.Context, id int64, req dto.TodoUpdateRequest) (dto.TodoResponse, error) {
	return f.updateFunc(ctx, id, req)
}
func (f *fakeService) Delete(ctx context.Context, id int64) error {
	return f.deleteFunc(ctx, id)
}
func (f *fakeService) DeleteCompleted(ctx context.Context) error {
	return f.deleteCompletedFunc(ctx)
}

func decodeProblem(t *testing.T, rec *httptest.ResponseRecorder) apperror.ProblemDetail {
	t.Helper()
	var pd apperror.ProblemDetail
	if err := json.NewDecoder(rec.Body).Decode(&pd); err != nil {
		t.Fatalf("decode ProblemDetail: %v (body=%q)", err, rec.Body.String())
	}
	return pd
}

func TestCreate_Success(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := &fakeService{
		createFunc: func(ctx context.Context, req dto.TodoCreateRequest) (dto.TodoResponse, error) {
			return dto.TodoResponse{ID: 5, Title: req.Title, CreatedAt: now, UpdatedAt: now}, nil
		},
	}
	h := New(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(`{"title":"buy milk"}`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body=%s)", rec.Code, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); loc != "/api/v1/todos/5" {
		t.Errorf("Location = %q, want /api/v1/todos/5", loc)
	}
	var got dto.TodoResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.ID != 5 || got.Title != "buy milk" {
		t.Errorf("got = %+v", got)
	}
}

func TestCreate_ValidationError(t *testing.T) {
	h := New(&fakeService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(`{"title":""}`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	pd := decodeProblem(t, rec)
	if len(pd.Errors) != 1 || pd.Errors[0].Field != "title" {
		t.Errorf("Errors = %+v, want one error on title", pd.Errors)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	h := New(&fakeService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(`{`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGet_Success(t *testing.T) {
	svc := &fakeService{
		getByIDFunc: func(ctx context.Context, id int64) (dto.TodoResponse, error) {
			return dto.TodoResponse{ID: id, Title: "x"}, nil
		},
	}
	h := New(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos/7", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestGet_InvalidID_ShortCircuitsBeforeService(t *testing.T) {
	svc := &fakeService{
		getByIDFunc: func(ctx context.Context, id int64) (dto.TodoResponse, error) {
			t.Fatal("service should not be called for an invalid id")
			return dto.TodoResponse{}, nil
		},
	}
	h := New(svc)

	for _, path := range []string{"/api/v1/todos/0", "/api/v1/todos/abc"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("path %s: status = %d, want 400", path, rec.Code)
		}
		pd := decodeProblem(t, rec)
		if len(pd.Errors) != 1 || pd.Errors[0].Field != "id" {
			t.Errorf("path %s: Errors = %+v, want one error on id", path, pd.Errors)
		}
	}
}

func TestGet_NotFound(t *testing.T) {
	svc := &fakeService{
		getByIDFunc: func(ctx context.Context, id int64) (dto.TodoResponse, error) {
			return dto.TodoResponse{}, apperror.NotFoundf("Todo not found: %d", id)
		},
	}
	h := New(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos/999", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestList_EmptyIsJSONArrayNotNull(t *testing.T) {
	svc := &fakeService{
		listFunc: func(ctx context.Context, completed *bool) ([]dto.TodoResponse, error) {
			return []dto.TodoResponse{}, nil
		},
	}
	h := New(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("body = %q, want []", body)
	}
}

func TestList_CompletedFilterParsed(t *testing.T) {
	var gotCompleted *bool
	svc := &fakeService{
		listFunc: func(ctx context.Context, completed *bool) ([]dto.TodoResponse, error) {
			gotCompleted = completed
			return []dto.TodoResponse{}, nil
		},
	}
	h := New(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos?completed=true", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if gotCompleted == nil || !*gotCompleted {
		t.Errorf("gotCompleted = %v, want true", gotCompleted)
	}
}

func TestUpdate_Success(t *testing.T) {
	svc := &fakeService{
		updateFunc: func(ctx context.Context, id int64, req dto.TodoUpdateRequest) (dto.TodoResponse, error) {
			return dto.TodoResponse{ID: id, Title: *req.Title}, nil
		},
	}
	h := New(svc)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/todos/1", strings.NewReader(`{"title":"new"}`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestUpdate_NotFound(t *testing.T) {
	svc := &fakeService{
		updateFunc: func(ctx context.Context, id int64, req dto.TodoUpdateRequest) (dto.TodoResponse, error) {
			return dto.TodoResponse{}, apperror.NotFoundf("Todo not found: %d", id)
		},
	}
	h := New(svc)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/todos/999", strings.NewReader(`{"title":"new"}`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestDelete_Success(t *testing.T) {
	svc := &fakeService{
		deleteFunc: func(ctx context.Context, id int64) error { return nil },
	}
	h := New(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/todos/1", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

func TestDelete_NotFound(t *testing.T) {
	svc := &fakeService{
		deleteFunc: func(ctx context.Context, id int64) error {
			return apperror.NotFoundf("Todo not found: %d", id)
		},
	}
	h := New(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/todos/999", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestDeleteCompleted_MissingParam(t *testing.T) {
	h := New(&fakeService{})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/todos", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	pd := decodeProblem(t, rec)
	if pd.Detail != "Required parameter 'completed' is not present" {
		t.Errorf("Detail = %q", pd.Detail)
	}
	if len(pd.Errors) != 0 {
		t.Errorf("Errors = %+v, want none for a missing param", pd.Errors)
	}
}

func TestDeleteCompleted_False(t *testing.T) {
	h := New(&fakeService{})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/todos?completed=false", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	pd := decodeProblem(t, rec)
	if len(pd.Errors) != 1 || pd.Errors[0].Field != "completed" || pd.Errors[0].Message != "must be true" {
		t.Errorf("Errors = %+v, want one 'must be true' error on completed", pd.Errors)
	}
}

func TestDeleteCompleted_True(t *testing.T) {
	called := false
	svc := &fakeService{
		deleteCompletedFunc: func(ctx context.Context) error {
			called = true
			return nil
		},
	}
	h := New(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/todos?completed=true", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if !called {
		t.Error("service.DeleteCompleted was not called")
	}
}
