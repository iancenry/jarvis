package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/iancenry/jarvis/internal/errs"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/attachment"
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
		Handler:     NewHandler(s),
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

func (th *TodoHandler) AddAttachment(c echo.Context) error {
	return Handle(
		th.Handler,
		func(c echo.Context, payload *attachment.UploadTodoAttachmentPayload) (*attachment.Attachment, error) {
			userID := middleware.GetUserID(c)

			c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, service.MaxAttachmentUploadRequestSizeBytes)
			form, err := c.MultipartForm()
			if err != nil {
				var maxBytesErr *http.MaxBytesError
				if errors.As(err, &maxBytesErr) {
					code := "ATTACHMENT_FILE_TOO_LARGE"
					return nil, errs.NewBadRequestError(
						fmt.Sprintf("attachment file exceeds %d MB limit", service.MaxAttachmentSizeBytes/(1<<20)),
						false,
						&code,
						nil,
						nil,
					)
				}
				return nil, errs.NewBadRequestError("multipart form not found", false, nil, nil, nil)
			}

			files := form.File["file"]
			if len(files) == 0 {
				return nil, errs.NewBadRequestError("file not found in multipart form", false, nil, nil, nil)
			}

			if len(files) > 1 {
				return nil, errs.NewBadRequestError("multiple files found in multipart form, only one file is allowed", false, nil, nil, nil)
			}

			return th.todoService.UploadTodoAttachment(c, userID, payload.TodoID, files[0])
		},
		http.StatusCreated,
		&attachment.UploadTodoAttachmentPayload{},
	)(c)
}

type presignedURLResponse struct {
	URL string `json:"url"`
}

func (th *TodoHandler) GetAttachmentPresignedURL(c echo.Context) error {
	return Handle(
		th.Handler,
		func(c echo.Context, payload *attachment.GetAttachmentPresignedURLPayload) (*presignedURLResponse, error) {
			userID := middleware.GetUserID(c)
			url, err := th.todoService.GetAttachmentPresignedURL(c, userID, payload.TodoID, payload.AttachmentID)
			if err != nil {
				return nil, err
			}
			return &presignedURLResponse{URL: url}, nil
		},
		http.StatusOK,
		&attachment.GetAttachmentPresignedURLPayload{},
	)(c)
}

func (th *TodoHandler) DeleteAttachment(c echo.Context) error {
	return HandleNoContent(
		th.Handler,
		func(c echo.Context, payload *attachment.DeleteTodoAttachmentPayload) error {
			userID := middleware.GetUserID(c)
			return th.todoService.DeleteTodoAttachment(c, userID, payload.TodoID, payload.ID)
		},
		http.StatusNoContent,
		&attachment.DeleteTodoAttachmentPayload{},
	)(c)
}
