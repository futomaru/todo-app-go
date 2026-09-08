package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/futomaru/todo-app-go/internal/apperror"
	"github.com/futomaru/todo-app-go/internal/clock"
	"github.com/futomaru/todo-app-go/internal/dto"
	"github.com/futomaru/todo-app-go/internal/model"
)

// TodoRepository declares the persistence methods TodoService needs.
// Declared on the consumer side (service) rather than the provider side
// (repository) so tests can substitute a fake without touching a real DB.
// repository.TodoRepository satisfies this interface structurally.
type TodoRepository interface {
	FindAll(ctx context.Context) ([]model.Todo, error)
	FindByCompleted(ctx context.Context, completed bool) ([]model.Todo, error)
	FindByID(ctx context.Context, id int64) (model.Todo, error)
	Insert(ctx context.Context, t model.Todo) (int64, error)
	Update(ctx context.Context, t model.Todo) error
	DeleteByID(ctx context.Context, id int64) (int64, error)
	DeleteCompleted(ctx context.Context) (int64, error)
}

type TodoService struct {
	repo  TodoRepository
	clock clock.Clock
}

func New(repo TodoRepository, clk clock.Clock) *TodoService {
	return &TodoService{repo: repo, clock: clk}
}

// Create inserts a new todo. Completed always starts false, and
// CreatedAt/UpdatedAt are set to the same instant.
func (s *TodoService) Create(ctx context.Context, req dto.TodoCreateRequest) (dto.TodoResponse, error) {
	now := s.clock.Now()
	t := model.Todo{
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	id, err := s.repo.Insert(ctx, t)
	if err != nil {
		return dto.TodoResponse{}, err
	}
	t.ID = id
	return dto.NewTodoResponse(t), nil
}

// List returns all todos, or only those matching completed when non-nil.
// The result slice is always non-nil so the JSON body is [] rather than
// null when there are zero matches.
func (s *TodoService) List(ctx context.Context, completed *bool) ([]dto.TodoResponse, error) {
	var (
		todos []model.Todo
		err   error
	)
	if completed == nil {
		todos, err = s.repo.FindAll(ctx)
	} else {
		todos, err = s.repo.FindByCompleted(ctx, *completed)
	}
	if err != nil {
		return nil, err
	}

	res := make([]dto.TodoResponse, 0, len(todos))
	for _, t := range todos {
		res = append(res, dto.NewTodoResponse(t))
	}
	return res, nil
}

// GetByID translates a missing row (sql.ErrNoRows) into apperror.ErrNotFound
// — the DB-level "no rows" concept becomes the business-level "not found".
func (s *TodoService) GetByID(ctx context.Context, id int64) (dto.TodoResponse, error) {
	t, err := s.repo.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return dto.TodoResponse{}, apperror.NotFoundf("Todo not found: %d", id)
	}
	if err != nil {
		return dto.TodoResponse{}, err
	}
	return dto.NewTodoResponse(t), nil
}

// Update applies PATCH semantics: only non-nil fields in req overwrite the
// existing todo. CreatedAt never changes; UpdatedAt advances to now.
func (s *TodoService) Update(ctx context.Context, id int64, req dto.TodoUpdateRequest) (dto.TodoResponse, error) {
	t, err := s.repo.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return dto.TodoResponse{}, apperror.NotFoundf("Todo not found: %d", id)
	}
	if err != nil {
		return dto.TodoResponse{}, err
	}

	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Description != nil {
		t.Description = req.Description
	}
	if req.Completed != nil {
		t.Completed = *req.Completed
	}
	t.UpdatedAt = s.clock.Now()

	if err := s.repo.Update(ctx, t); err != nil {
		return dto.TodoResponse{}, err
	}
	return dto.NewTodoResponse(t), nil
}

// Delete removes a single todo. Unlike DeleteCompleted, deleting an
// already-absent id is treated as an error (not idempotent) per spec.
func (s *TodoService) Delete(ctx context.Context, id int64) error {
	affected, err := s.repo.DeleteByID(ctx, id)
	if err != nil {
		return err
	}
	if affected == 0 {
		return apperror.NotFoundf("Todo not found: %d", id)
	}
	return nil
}

// DeleteCompleted removes every completed todo. Zero matches is success,
// not an error — this is a collection-level operation.
func (s *TodoService) DeleteCompleted(ctx context.Context) error {
	_, err := s.repo.DeleteCompleted(ctx)
	return err
}
