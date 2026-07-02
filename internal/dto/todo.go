package dto

import (
	"time"

	"github.com/futomaru/todo-app-go/internal/model"
)

type TodoResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func NewTodoResponse(t model.Todo) TodoResponse {
	return TodoResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

type TodoCreateRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
}

type TodoUpdateRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}
