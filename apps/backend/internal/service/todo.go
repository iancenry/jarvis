package service

import (
	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/errs"
	"github.com/iancenry/jarvis/internal/lib/aws"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/todo"
	"github.com/iancenry/jarvis/internal/repository"
	"github.com/iancenry/jarvis/internal/server"
	"github.com/labstack/echo/v4"
)

type TodoService struct {
	server *server.Server
	todoRepo *repository.TodoRepository
	categoryRepo *repository.CategoryRepository
	awsClient *aws.AWS
}

func NewTodoService(s *server.Server, todoRepo *repository.TodoRepository, categoryRepo *repository.CategoryRepository, awsClient *aws.AWS) *TodoService {
	return &TodoService{
		server: s,
		todoRepo: todoRepo,
		categoryRepo: categoryRepo,
		awsClient: awsClient,
	}
}

func (ts *TodoService) CreateTodo(ctx echo.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
	logger := middleware.GetLogger(ctx)

	if payload.ParentTodoID != nil {
		parentTodo, err := ts.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, *payload.ParentTodoID)
		if err != nil {
			logger.Error().Err(err).Msg("parent todo validation failed")
			return nil, err
		}

		if !parentTodo.CanHaveChildren() {
			err := errs.NewBadRequestError("Parent todo cannot have children (subtasks can't have subtasks)", false, nil, nil, nil)
			logger.Warn().Msg("parent todo cannot have children")
			return nil, err
		}
	}

	if payload.CategoryID != nil {
		_, err := ts.categoryRepo.GetCategoryByID(ctx.Request().Context(), userID, *payload.CategoryID)
		if err != nil {
			logger.Error().Err(err).Msg("category validation failed")
			return nil, err
		}
	}

	todoItem, err := ts.todoRepo.CreateTodo(ctx.Request().Context(), userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create todo")
		return nil, err
	}

	// business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "todo_created").
		Str("todo_id", todoItem.ID.String()).
		Str("title", todoItem.Title).
		Str("category_id", func() string {
			if todoItem.CategoryID != nil {
				return todoItem.CategoryID.String()
			}
			return "none"
		}()).
		Str("priority", string(todoItem.Priority)).
		Msg("todo created")

	return todoItem, nil
}

func (ts *TodoService) GetTodoByID(ctx echo.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error) {
	logger := middleware.GetLogger(ctx)

	todoItem, err := ts.todoRepo.GetTodoByID(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get todo by ID")
		return nil, err
	}

	return todoItem, nil
}

func (ts *TodoService) GetTodos(ctx echo.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
	logger := middleware.GetLogger(ctx)

	todos, err := ts.todoRepo.GetTodos(ctx.Request().Context(), userID, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get todos")
		return nil, err
	}

	return todos, nil
}

func (ts *TodoService) UpdateTodo(ctx echo.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
	logger := middleware.GetLogger(ctx)

	if payload.ParentTodoID != nil {
		parentTodo, err := ts.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, *payload.ParentTodoID)
		if err != nil {
			logger.Error().Err(err).Msg("parent todo validation failed")
			return nil, err
		}

		if parentTodo.ID == payload.ID {
			err := errs.NewBadRequestError("Todo cannot be its own parent", false, nil, nil, nil)
			logger.Warn().Msg("todo cannot be its own parent")
			return nil, err
		}

		if !parentTodo.CanHaveChildren() {
			err := errs.NewBadRequestError("Parent todo cannot have children (subtasks can't have subtasks)", false, nil, nil, nil)
			logger.Warn().Msg("parent todo cannot have children")
			return nil, err
		}

		logger.Debug().Msg("parent todo validation passed")
	}

	if payload.CategoryID != nil {
		_, err := ts.categoryRepo.GetCategoryByID(ctx.Request().Context(), userID, *payload.CategoryID)
		if err != nil {
			logger.Error().Err(err).Msg("category validation failed")
			return nil, err
		}
		logger.Debug().Msg("category validation passed")
	}

	updatedTodo, err := ts.todoRepo.UpdateTodo(ctx.Request().Context(), userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update todo")
		return nil, err
	}
	// business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "todo_updated").
		Str("todo_id", updatedTodo.ID.String()).
		Str("title", updatedTodo.Title).
		Str("category_id", func() string {
			if updatedTodo.CategoryID != nil {
				return updatedTodo.CategoryID.String()
			}
			return "none"
		}()).
		Str("priority", string(updatedTodo.Priority)).
		Str("status", string(updatedTodo.Status)).
		Msg("todo updated")

	return updatedTodo, nil
}

func (ts *TodoService) DeleteTodo(ctx echo.Context, userID string, todoID uuid.UUID) error {
	logger := middleware.GetLogger(ctx)

	err := ts.todoRepo.DeleteTodo(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete todo")
		return err
	}

	// business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "todo_deleted").
		Str("todo_id", todoID.String()).
		Msg("todo deleted")

	return nil
}

func (ts *TodoService) GetTodoStats(ctx echo.Context, userID string) (*todo.TodoStats, error) {
	logger := middleware.GetLogger(ctx)

	stats, err := ts.todoRepo.GetTodoStats(ctx.Request().Context(), userID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get todo stats")
		return nil, err
	}

	return stats, nil
}