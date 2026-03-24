package service

import (
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/errs"
	"github.com/iancenry/jarvis/internal/lib/aws"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/attachment"
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

func (ts *TodoService) UploadTodoAttachment(ctx echo.Context, userID string, todoID uuid.UUID, file *multipart.FileHeader) (*attachment.Attachment, error) {
	logger := middleware.GetLogger(ctx)

	// Check if the todo exists and belongs to the user
	_, err := ts.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed")
		return nil, err
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		logger.Error().Err(err).Msg("failed to open uploaded file")
		return nil, errs.NewBadRequestError("failed to open uploaded file", false, nil, nil, nil)
	}
	defer src.Close()

	s3Key, err := ts.awsClient.S3.UploadFile(
		ctx.Request().Context(),
		ts.server.Config.AWS.UploadBucket,
		"todos/attachments/"+file.Filename,
		src,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to upload file to S3")
		return nil, err
	}

	src, err = file.Open()
	if err != nil {
		logger.Error().Err(err).Msg("failed to reopen uploaded file")
		return nil, errs.NewBadRequestError("failed to process uploaded file", false, nil, nil, nil)
	}
	defer src.Close()

	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read uploaded file")
		return nil, errs.NewBadRequestError("failed to process uploaded file", false, nil, nil, nil)
	}
	mimeType := http.DetectContentType(buffer)

	attachmentItem, err := ts.todoRepo.AddAttachment(ctx.Request().Context(), todoID, userID, s3Key, file.Filename, file.Size, mimeType)
	if err != nil {
		logger.Error().Err(err).Msg("failed to save attachment metadata")
		return nil, err
	}

	logger.Info().Str("event", "attachment_uploaded").
		Str("s3_key", s3Key).
		Str("todo_id", todoID.String()).
		Str("attachment_id", attachmentItem.ID.String()).
		Str("filename", file.Filename).
		Int64("size", file.Size).
		Str("mime_type", mimeType).
		Msg("attachment uploaded")

	return attachmentItem, nil


	
}
func (ts *TodoService) GetTodoAttachments(ctx echo.Context, todoID uuid.UUID) ([]attachment.Attachment, error) {
	logger := middleware.GetLogger(ctx)

	attachments, err := ts.todoRepo.GetAttachmentsByTodoID(ctx.Request().Context(), todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get todo attachments")
		return nil, err
	}

	return attachments, nil
}

func (ts *TodoService) DeleteTodoAttachment(ctx echo.Context, userID string, todoID uuid.UUID, attachmentID uuid.UUID) error {
	logger := middleware.GetLogger(ctx)

	_, err := ts.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed")
		return err
	}

	attachmentItem, err := ts.todoRepo.GetTodoAttachment(ctx.Request().Context(), todoID, attachmentID)
	if err != nil {
		logger.Error().Err(err).Msg("attachment validation failed")
		return err
	}

	err = ts.awsClient.S3.DeleteFile(ctx.Request().Context(), ts.server.Config.AWS.UploadBucket, attachmentItem.DownloadKey)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete file from S3")
		return err
	}

	err = ts.todoRepo.DeleteAttachment(ctx.Request().Context(), todoID, attachmentID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete attachment metadata")
		return err
	}

	go func ()  {
		err := ts.awsClient.S3.DeleteFile(ctx.Request().Context(), ts.server.Config.AWS.UploadBucket, attachmentItem.DownloadKey)
		if err != nil {
			logger.Error().Err(err).Msg("failed to delete file from S3 in background")
		}
	}()

	logger.Info().Str("event", "attachment_deleted").
		Str("todo_id", todoID.String()).
		Str("attachment_id", attachmentID.String()).
		Str("filename", attachmentItem.Name).
		Msg("attachment deleted")

	return nil
}

func (s *TodoService) GetAttachmentPresignedURL(ctx echo.Context, userID string, todoID uuid.UUID, attachmentID uuid.UUID) (string, error) {
	logger := middleware.GetLogger(ctx)

	_, err := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed")
		return "", err
	}

	attachmentItem, err := s.todoRepo.GetTodoAttachment(ctx.Request().Context(), todoID, attachmentID)
	if err != nil {
		logger.Error().Err(err).Msg("attachment validation failed")
		return "", err
	}

	presignedURL, err := s.awsClient.S3.CreatePresignedURL(ctx.Request().Context(), s.server.Config.AWS.UploadBucket, attachmentItem.DownloadKey)
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate presigned URL")
		return "", err
	}

	logger.Info().Str("event", "presigned_url_generated").
		Str("todo_id", todoID.String()).
		Str("attachment_id", attachmentID.String()).
		Str("filename", attachmentItem.Name).
		Msg("presigned URL generated")

	return presignedURL, nil
}