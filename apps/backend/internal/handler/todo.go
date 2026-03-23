package handler

import (
	"net/http"

	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/todo"
	"github.com/iancenry/jarvis/internal/server"
	"github.com/iancenry/jarvis/internal/service"
	"github.com/labstack/echo/v4"
)

type TodoHandler struct {
	Handler
	todoService *service.TodoService
}

func NewTodoHandler(s *server.Server, todoService *service.TodoService) *TodoHandler {
	return &TodoHandler{
		Handler: NewHandler(s),
		todoService: todoService,
	}
}

func (th *TodoHandler) CreateTodo(c echo.Context) error {
	return Handle(
		th.Handler,
		func(c echo.Context, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
			userID := middleware.GetUserID(c)
			return th.todoService.CreateTodo(c, userID, payload)
		},
		http.StatusCreated,
		&todo.CreateTodoPayload{},
	)(c)
}

func (th *TodoHandler) GetTodoByID(c echo.Context) error {
	return Handle(
		th.Handler,
		func(c echo.Context, payload *todo.GetTodoByIDQuery) (*todo.PopulatedTodo, error) {
			userID := middleware.GetUserID(c)
			return th.todoService.GetTodoByID(c, userID, payload.ID)
		},
		http.StatusOK,
		&todo.GetTodoByIDQuery{},
	)(c)
}

func (th *TodoHandler) GetTodos(c echo.Context) error {
	return Handle(
		th.Handler,
		func(c echo.Context, payload *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
			userID := middleware.GetUserID(c)
			return th.todoService.GetTodos(c, userID, payload)
		},
		http.StatusOK,
		&todo.GetTodosQuery{},
	)(c)
}

func (th *TodoHandler) UpdateTodo(c echo.Context) error {
	return Handle(
		th.Handler,
		func(c echo.Context, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
			userID := middleware.GetUserID(c)
			return th.todoService.UpdateTodo(c, userID, payload)
		},
		http.StatusOK,
		&todo.UpdateTodoPayload{},
	)(c)
}

func (th *TodoHandler) DeleteTodo(c echo.Context) error {
	return HandleNoContent(
		th.Handler,
		func(c echo.Context, payload *todo.DeleteTodoByIDQuery) error {
			userID := middleware.GetUserID(c)
			return th.todoService.DeleteTodo(c, userID, payload.ID)
		},
		http.StatusNoContent,
		&todo.DeleteTodoByIDQuery{},
	)(c)
}


func (th *TodoHandler) GetTodoStats(c echo.Context) error {
	return Handle(
		th.Handler,
		func(c echo.Context, payload *todo.GetTodoStatsQuery) (*todo.TodoStats, error) {
			userID := middleware.GetUserID(c)
			return th.todoService.GetTodoStats(c, userID)
		},
		http.StatusOK,
		&todo.GetTodoStatsQuery{},
	)(c)
}