package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/errs"
	"github.com/iancenry/jarvis/internal/lib/aws"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/attachment"
	"github.com/iancenry/jarvis/internal/model/todo"
	"github.com/iancenry/jarvis/internal/repository"
	"github.com/rs/zerolog"
)

const attachmentCompensationTimeout = 5 * time.Second
const attachmentStorageDisabledCode = "ATTACHMENT_STORAGE_DISABLED"

type todoAttachmentRepository interface {
	CheckTodoExists(ctx context.Context, userID string, todoID uuid.UUID) (*todo.Todo, error)
	AddAttachment(ctx context.Context, todoID uuid.UUID, userID string, s3Key string, fileName string, fileSize int64, mimeType string) (*attachment.Attachment, error)
	GetAttachmentsByTodoID(ctx context.Context, todoID uuid.UUID) ([]attachment.Attachment, error)
	GetTodoAttachment(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) (*attachment.Attachment, error)
	DeleteAttachment(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) error
	RestoreAttachment(ctx context.Context, item *attachment.Attachment) error
}

type attachmentFileStore interface {
	UploadFile(ctx context.Context, bucket, fileKey string, file io.Reader, fileSize int64, contentType string) (string, error)
	CreatePresignedURL(ctx context.Context, bucket, fileKey string) (string, error)
	DeleteFile(ctx context.Context, bucket, fileKey string) error
}

type TodoService struct {
	todoRepo       *repository.TodoRepository
	attachmentRepo todoAttachmentRepository
	categoryRepo   *repository.CategoryRepository
	fileStore      attachmentFileStore
	uploadBucket   string
}

func NewTodoService(todoRepo *repository.TodoRepository, categoryRepo *repository.CategoryRepository, awsClient *aws.AWS, uploadBucket string) *TodoService {
	var fileStore attachmentFileStore
	if awsClient != nil {
		fileStore = awsClient.S3
	}

	return &TodoService{
		todoRepo:       todoRepo,
		attachmentRepo: todoRepo,
		categoryRepo:   categoryRepo,
		fileStore:      fileStore,
		uploadBucket:   uploadBucket,
	}
}

func (ts *TodoService) CreateTodo(ctx context.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
	logger := middleware.LoggerFromContext(ctx)

	if payload.ParentTodoID != nil {
		parentTodo, err := ts.todoRepo.CheckTodoExists(ctx, userID, *payload.ParentTodoID)
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
		_, err := ts.categoryRepo.GetCategoryByID(ctx, userID, *payload.CategoryID)
		if err != nil {
			logger.Error().Err(err).Msg("category validation failed")
			return nil, err
		}
	}

	todoItem, err := ts.todoRepo.CreateTodo(ctx, userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create todo")
		return nil, err
	}

	// business event log
	eventLogger := middleware.LoggerFromContext(ctx)
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

func (ts *TodoService) GetTodoByID(ctx context.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error) {
	logger := middleware.LoggerFromContext(ctx)

	todoItem, err := ts.todoRepo.GetTodoByID(ctx, userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get todo by ID")
		return nil, err
	}

	return todoItem, nil
}

func (ts *TodoService) GetTodos(ctx context.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
	logger := middleware.LoggerFromContext(ctx)

	todos, err := ts.todoRepo.GetTodos(ctx, userID, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get todos")
		return nil, err
	}

	return todos, nil
}

func (ts *TodoService) UpdateTodo(ctx context.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
	logger := middleware.LoggerFromContext(ctx)

	if payload.ParentTodoID.IsSet() && !payload.ParentTodoID.IsNull() {
		parentTodoID := payload.ParentTodoID.Value()
		parentTodo, err := ts.todoRepo.CheckTodoExists(ctx, userID, parentTodoID)
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

	if payload.CategoryID.IsSet() && !payload.CategoryID.IsNull() {
		categoryID := payload.CategoryID.Value()
		_, err := ts.categoryRepo.GetCategoryByID(ctx, userID, categoryID)
		if err != nil {
			logger.Error().Err(err).Msg("category validation failed")
			return nil, err
		}
		logger.Debug().Msg("category validation passed")
	}

	updatedTodo, err := ts.todoRepo.UpdateTodo(ctx, userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update todo")
		return nil, err
	}
	// business event log
	eventLogger := middleware.LoggerFromContext(ctx)
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

func (ts *TodoService) DeleteTodo(ctx context.Context, userID string, todoID uuid.UUID) error {
	logger := middleware.LoggerFromContext(ctx)

	deletedTodo, err := ts.todoRepo.DeleteTodo(ctx, userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete todo")
		return err
	}

	// business event log
	eventLogger := middleware.LoggerFromContext(ctx)
	eventLogger.Info().
		Str("event", "todo_deleted").
		Str("todo_id", deletedTodo.ID.String()).
		Str("title", deletedTodo.Title).
		Msg("todo deleted")

	return nil
}

func (ts *TodoService) GetTodoStats(ctx context.Context, userID string) (*todo.TodoStats, error) {
	logger := middleware.LoggerFromContext(ctx)

	stats, err := ts.todoRepo.GetTodoStats(ctx, userID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get todo stats")
		return nil, err
	}

	return stats, nil
}

func (ts *TodoService) UploadTodoAttachment(ctx context.Context, userID string, todoID uuid.UUID, file *multipart.FileHeader) (*attachment.Attachment, error) {
	logger := middleware.LoggerFromContext(ctx)

	if err := ts.requireAttachmentStorage(); err != nil {
		logger.Warn().Err(err).Msg("attachment upload requested while storage is disabled")
		return nil, err
	}

	// Check if the todo exists and belongs to the user
	_, err := ts.attachmentRepo.CheckTodoExists(ctx, userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed")
		return nil, err
	}

	sanitizedFileName, err := sanitizeAttachmentFileName(file.Filename)
	if err != nil {
		logger.Warn().Err(err).Str("filename", file.Filename).Msg("attachment filename validation failed")
		return nil, err
	}

	if err := validateAttachmentSize(file.Size); err != nil {
		logger.Warn().
			Err(err).
			Str("filename", sanitizedFileName).
			Int64("size", file.Size).
			Msg("attachment size validation failed")
		return nil, err
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		logger.Error().Err(err).Msg("failed to open uploaded file")
		return nil, errs.NewBadRequestError("failed to open uploaded file", false, nil, nil, nil)
	}
	defer src.Close()

	contentType, uploadReader, err := detectAttachmentContentType(src)
	if err != nil {
		logger.Warn().
			Err(err).
			Str("filename", sanitizedFileName).
			Int64("size", file.Size).
			Msg("attachment content validation failed")
		return nil, err
	}

	storageKey := buildAttachmentStorageKey(sanitizedFileName)
	s3Key, err := ts.fileStore.UploadFile(
		ctx,
		ts.uploadBucket,
		storageKey,
		uploadReader,
		file.Size,
		contentType,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to upload file to S3")
		return nil, err
	}

	attachmentItem, err := ts.attachmentRepo.AddAttachment(ctx, todoID, userID, s3Key, sanitizedFileName, file.Size, contentType)
	if err != nil {
		logger.Error().Err(err).Msg("failed to save attachment metadata")
		return nil, ts.compensateAttachmentUpload(logger, s3Key, err)
	}

	logger.Info().Str("event", "attachment_uploaded").
		Str("s3_key", s3Key).
		Str("todo_id", todoID.String()).
		Str("attachment_id", attachmentItem.ID.String()).
		Str("filename", sanitizedFileName).
		Int64("size", file.Size).
		Str("mime_type", contentType).
		Msg("attachment uploaded")

	return attachmentItem, nil
}

func (ts *TodoService) GetTodoAttachments(ctx context.Context, todoID uuid.UUID) ([]attachment.Attachment, error) {
	logger := middleware.LoggerFromContext(ctx)

	attachments, err := ts.attachmentRepo.GetAttachmentsByTodoID(ctx, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get todo attachments")
		return nil, err
	}

	return attachments, nil
}

func (ts *TodoService) DeleteTodoAttachment(ctx context.Context, userID string, todoID uuid.UUID, attachmentID uuid.UUID) error {
	logger := middleware.LoggerFromContext(ctx)

	if err := ts.requireAttachmentStorage(); err != nil {
		logger.Warn().Err(err).Msg("attachment delete requested while storage is disabled")
		return err
	}

	_, err := ts.attachmentRepo.CheckTodoExists(ctx, userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed")
		return err
	}

	attachmentItem, err := ts.attachmentRepo.GetTodoAttachment(ctx, todoID, attachmentID)
	if err != nil {
		logger.Error().Err(err).Msg("attachment validation failed")
		return err
	}

	err = ts.attachmentRepo.DeleteAttachment(ctx, todoID, attachmentID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete attachment metadata")
		return err
	}

	err = ts.deleteAttachmentFile(attachmentItem.DownloadKey)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete file from S3")

		restoreErr := ts.restoreAttachmentMetadata(attachmentItem)
		if restoreErr != nil {
			logger.Error().Err(restoreErr).Msg("failed to restore attachment metadata after S3 delete failure")
			return errors.Join(err, fmt.Errorf("failed to restore attachment metadata after S3 deletion error: %w", restoreErr))
		}

		logger.Warn().
			Err(err).
			Str("attachment_id", attachmentItem.ID.String()).
			Str("todo_id", todoID.String()).
			Msg("restored attachment metadata after S3 delete failure")
		return err
	}

	logger.Info().Str("event", "attachment_deleted").
		Str("todo_id", todoID.String()).
		Str("attachment_id", attachmentID.String()).
		Str("filename", attachmentItem.Name).
		Msg("attachment deleted")

	return nil
}

func (s *TodoService) GetAttachmentPresignedURL(ctx context.Context, userID string, todoID uuid.UUID, attachmentID uuid.UUID) (string, error) {
	logger := middleware.LoggerFromContext(ctx)

	if err := s.requireAttachmentStorage(); err != nil {
		logger.Warn().Err(err).Msg("presigned url requested while storage is disabled")
		return "", err
	}

	_, err := s.attachmentRepo.CheckTodoExists(ctx, userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed")
		return "", err
	}

	attachmentItem, err := s.attachmentRepo.GetTodoAttachment(ctx, todoID, attachmentID)
	if err != nil {
		logger.Error().Err(err).Msg("attachment validation failed")
		return "", err
	}

	presignedURL, err := s.fileStore.CreatePresignedURL(ctx, s.uploadBucket, attachmentItem.DownloadKey)
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

func (ts *TodoService) compensateAttachmentUpload(logger *zerolog.Logger, s3Key string, originalErr error) error {
	cleanupErr := ts.deleteAttachmentFile(s3Key)
	if cleanupErr == nil {
		logger.Warn().
			Err(originalErr).
			Str("s3_key", s3Key).
			Msg("cleaned up uploaded attachment after metadata failure")
		return originalErr
	}

	logger.Error().
		Err(cleanupErr).
		Str("s3_key", s3Key).
		Msg("failed to clean up uploaded attachment after metadata failure")

	return errors.Join(originalErr, fmt.Errorf("failed to clean up uploaded attachment %s: %w", s3Key, cleanupErr))
}

func (ts *TodoService) deleteAttachmentFile(s3Key string) error {
	if err := ts.requireAttachmentStorage(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), attachmentCompensationTimeout)
	defer cancel()

	return ts.fileStore.DeleteFile(ctx, ts.uploadBucket, s3Key)
}

func (ts *TodoService) restoreAttachmentMetadata(item *attachment.Attachment) error {
	ctx, cancel := context.WithTimeout(context.Background(), attachmentCompensationTimeout)
	defer cancel()

	return ts.attachmentRepo.RestoreAttachment(ctx, item)
}

func (ts *TodoService) requireAttachmentStorage() error {
	if ts != nil && ts.fileStore != nil && ts.uploadBucket != "" {
		return nil
	}

	return errs.NewServiceUnavailableError(
		"attachment storage is disabled",
		false,
		stringPointer(attachmentStorageDisabledCode),
	)
}

func stringPointer(value string) *string {
	return &value
}
