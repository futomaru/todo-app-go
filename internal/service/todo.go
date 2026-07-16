package service

import (
	"context"

	"github.com/futomaru/todo-app-go/internal/clock"
	"github.com/futomaru/todo-app-go/internal/dto"
	"github.com/futomaru/todo-app-go/internal/model"
)

// TodoRepository declares only the persistence methods the service needs.
// Following the Go convention "the consumer defines the interface", it lives
// here (not in the repository package) so tests can substitute a fake and the
// concrete *repository.TodoRepository satisfies it structurally.
type TodoRepository interface {
	FindAll(ctx context.Context) ([]model.Todo, error)
	FindByCompleted(ctx context.Context, completed bool) ([]model.Todo, error)
	FindByID(ctx context.Context, id int64) (model.Todo, error)
	Insert(ctx context.Context, t model.Todo) (int64, error)
	Update(ctx context.Context, t model.Todo) error
	DeleteByID(ctx context.Context, id int64) (int64, error)
	DeleteCompleted(ctx context.Context) (int64, error)
}

// TodoService holds the use-case logic. It depends on the repository (through
// the interface above) and a Clock so timestamps are deterministic in tests.
type TodoService struct {
	repo  TodoRepository
	clock clock.Clock
}

// New wires a TodoService with its dependencies (constructor injection).
func New(repo TodoRepository, clk clock.Clock) *TodoService {
	return &TodoService{repo: repo, clock: clk}
}

// Create builds a new Todo (completed=false, createdAt==updatedAt==now),
// inserts it, sets the generated id, and returns it as a TodoResponse.
//
// TODO: now := s.clock.Now(); assemble model.Todo from req; id from repo.Insert;
// t.ID = id; return dto.NewTodoResponse(t).
func (s *TodoService) Create(ctx context.Context, req dto.TodoCreateRequest) (dto.TodoResponse, error) {
	panic("not implemented")
}

// List returns all todos, or only those matching completed when it is non-nil.
// The result must be a non-nil (possibly empty) slice so the handler serialises
// it as a JSON [] rather than null.
//
// TODO: pick repo.FindAll vs repo.FindByCompleted on completed==nil; map each
// model.Todo through dto.NewTodoResponse into a make([]dto.TodoResponse, 0, n).
func (s *TodoService) List(ctx context.Context, completed *bool) ([]dto.TodoResponse, error) {
	panic("not implemented")
}

// GetByID fetches one todo. Translating the repository's sql.ErrNoRows into the
// business-level apperror.ErrNotFound is THIS layer's job (repository speaks SQL,
// handler speaks HTTP, service bridges them).
//
// TODO: t, err := repo.FindByID; if errors.Is(err, sql.ErrNoRows) return
// apperror.NotFoundf("Todo not found: %d", id); else propagate; else return
// dto.NewTodoResponse(t).
func (s *TodoService) GetByID(ctx context.Context, id int64) (dto.TodoResponse, error) {
	panic("not implemented")
}

// Update applies PATCH semantics: load the current todo (404 if missing),
// overwrite only the fields whose request pointer is non-nil, bump UpdatedAt
// (CreatedAt stays immutable), persist, and return the result.
//
// TODO: FindByID (+ NotFound translation); if req.Title != nil {...} etc.;
// t.UpdatedAt = s.clock.Now(); repo.Update; return dto.NewTodoResponse(t).
func (s *TodoService) Update(ctx context.Context, id int64, req dto.TodoUpdateRequest) (dto.TodoResponse, error) {
	panic("not implemented")
}

// Delete removes a single todo. Because DeleteByID reports the affected-row
// count, "0 rows" means the todo did not exist and must become a NotFound
// error (single-delete is not idempotent here, by spec §5.5).
//
// TODO: affected, err := repo.DeleteByID; propagate err; if affected == 0
// return apperror.NotFoundf("Todo not found: %d", id); else nil.
func (s *TodoService) Delete(ctx context.Context, id int64) error {
	panic("not implemented")
}

// DeleteCompleted bulk-deletes every completed todo. Deleting zero rows is a
// valid success (collection semantics), so the count is discarded.
//
// TODO: _, err := repo.DeleteCompleted(ctx); return err.
func (s *TodoService) DeleteCompleted(ctx context.Context) error {
	panic("not implemented")
}
