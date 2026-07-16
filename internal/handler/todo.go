package handler

import (
	"context"
	"net/http"

	"github.com/futomaru/todo-app-go/internal/dto"
)

// TodoService declares only the use-case methods the handler needs (consumer
// defines the interface). The concrete *service.TodoService satisfies it, and
// tests can pass a fake.
type TodoService interface {
	List(ctx context.Context, completed *bool) ([]dto.TodoResponse, error)
	GetByID(ctx context.Context, id int64) (dto.TodoResponse, error)
	Create(ctx context.Context, req dto.TodoCreateRequest) (dto.TodoResponse, error)
	Update(ctx context.Context, id int64, req dto.TodoUpdateRequest) (dto.TodoResponse, error)
	Delete(ctx context.Context, id int64) error
	DeleteCompleted(ctx context.Context) error
}

// TodoHandler adapts HTTP requests to TodoService calls.
type TodoHandler struct {
	svc TodoService
}

// New wires a TodoHandler with the service it delegates to.
func New(svc TodoService) *TodoHandler {
	return &TodoHandler{svc: svc}
}

// Every handler shares the same shape: parse/decode input, validate, call the
// service, then write success — routing all failures through
// apperror.WriteError so error-to-HTTP mapping lives in exactly one place.

// list handles GET /api/v1/todos. Reads the optional ?completed query
// (absent -> nil = all; present -> strconv.ParseBool) and returns the slice.
//
// TODO: derive *bool from r.URL.Query().Get("completed"); svc.List; writeJSON 200.
func (h *TodoHandler) list(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// get handles GET /api/v1/todos/{id}.
//
// TODO: parseID (400 on bad id); svc.GetByID; on error WriteError (404 when
// NotFound); writeJSON 200.
func (h *TodoHandler) get(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// create handles POST /api/v1/todos.
//
// TODO: decodeJSON into dto.TodoCreateRequest; dto.ValidateCreateRequest (wrap
// non-empty result in *apperror.ValidationError); svc.Create; set Location
// header to r.URL.Path + "/" + id; writeJSON 201.
func (h *TodoHandler) create(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// update handles PATCH /api/v1/todos/{id}.
//
// TODO: parseID; decodeJSON into dto.TodoUpdateRequest; dto.ValidateUpdateRequest;
// svc.Update; writeJSON 200.
func (h *TodoHandler) update(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// delete handles DELETE /api/v1/todos/{id}.
//
// TODO: parseID; svc.Delete (404 when NotFound); w.WriteHeader(204).
func (h *TodoHandler) delete(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// deleteCompleted handles DELETE /api/v1/todos (bulk). The ?completed query is
// REQUIRED and must equal true.
//
// TODO: if completed param missing -> &apperror.ValidationError{Detail:
// "Required parameter 'completed' is not present"}; if not true ->
// ValidationError with a {field:"completed", message:"must be true"};
// otherwise svc.DeleteCompleted; w.WriteHeader(204).
func (h *TodoHandler) deleteCompleted(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// parseID extracts and validates the {id} path segment shared by get/update/delete.
// A non-numeric or < 1 id is a 400 (a *apperror.ValidationError on "id"), short-
// circuiting before the service is ever called.
//
// TODO: strconv.ParseInt(r.PathValue("id"), 10, 64); reject err != nil || id < 1.
func parseID(r *http.Request) (int64, error) {
	panic("not implemented")
}
