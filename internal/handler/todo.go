package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/futomaru/todo-app-go/internal/apperror"
	"github.com/futomaru/todo-app-go/internal/dto"
)

// TodoService declares the use-case methods the handler needs.
// Implemented by service.TodoService in production, by a fake in tests.
type TodoService interface {
	List(ctx context.Context, completed *bool) ([]dto.TodoResponse, error)
	GetByID(ctx context.Context, id int64) (dto.TodoResponse, error)
	Create(ctx context.Context, req dto.TodoCreateRequest) (dto.TodoResponse, error)
	Update(ctx context.Context, id int64, req dto.TodoUpdateRequest) (dto.TodoResponse, error)
	Delete(ctx context.Context, id int64) error
	DeleteCompleted(ctx context.Context) error
}

type TodoHandler struct {
	svc TodoService
}

func New(svc TodoService) *TodoHandler {
	return &TodoHandler{svc: svc}
}

func (h *TodoHandler) list(w http.ResponseWriter, r *http.Request) {
	var completed *bool
	if raw := r.URL.Query().Get("completed"); raw != "" {
		b, err := strconv.ParseBool(raw)
		if err != nil {
			apperror.WriteError(w, r, &apperror.ValidationError{Fields: []apperror.FieldError{{
				Field: "completed", Message: "must be a boolean",
			}}})
			return
		}
		completed = &b
	}

	res, err := h.svc.List(r.Context(), completed)
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *TodoHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}

	res, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *TodoHandler) create(w http.ResponseWriter, r *http.Request) {
	var req dto.TodoCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		apperror.WriteError(w, r, err)
		return
	}
	if fe := dto.ValidateCreateRequest(req); len(fe) > 0 {
		apperror.WriteError(w, r, &apperror.ValidationError{Fields: fe})
		return
	}

	res, err := h.svc.Create(r.Context(), req)
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("%s/%d", r.URL.Path, res.ID))
	writeJSON(w, http.StatusCreated, res)
}

func (h *TodoHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}

	var req dto.TodoUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		apperror.WriteError(w, r, err)
		return
	}
	if fe := dto.ValidateUpdateRequest(req); len(fe) > 0 {
		apperror.WriteError(w, r, &apperror.ValidationError{Fields: fe})
		return
	}

	res, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *TodoHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		apperror.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// deleteCompleted requires ?completed=true. Missing and "wrong value" are
// deliberately different errors (api.yaml): missing has no Fields (just a
// Detail override), an explicit non-true value gets a field-level error.
func (h *TodoHandler) deleteCompleted(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("completed")
	if raw == "" {
		apperror.WriteError(w, r, &apperror.ValidationError{
			Detail: "Required parameter 'completed' is not present",
		})
		return
	}
	if b, err := strconv.ParseBool(raw); err != nil || !b {
		apperror.WriteError(w, r, &apperror.ValidationError{Fields: []apperror.FieldError{{
			Field: "completed", Message: "must be true",
		}}})
		return
	}

	if err := h.svc.DeleteCompleted(r.Context()); err != nil {
		apperror.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseID short-circuits before the service layer is ever called —
// an invalid id is a request-shape problem, not a business-logic one.
func parseID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, &apperror.ValidationError{Fields: []apperror.FieldError{{
			Field: "id", Message: "must be greater than or equal to 1",
		}}}
	}
	return id, nil
}
