package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/config"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/attachment"
	"github.com/iancenry/jarvis/internal/model/todo"
	"github.com/iancenry/jarvis/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type attachmentRepoStub struct {
	checkTodoExistsFn      func(ctx context.Context, userID string, todoID uuid.UUID) (*todo.Todo, error)
	addAttachmentFn        func(ctx context.Context, todoID uuid.UUID, userID string, s3Key string, fileName string, fileSize int64, mimeType string) (*attachment.Attachment, error)
	getAttachmentsByTodoID func(ctx context.Context, todoID uuid.UUID) ([]attachment.Attachment, error)
	getTodoAttachmentFn    func(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) (*attachment.Attachment, error)
	deleteAttachmentFn     func(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) error
	restoreAttachmentFn    func(ctx context.Context, item *attachment.Attachment) error

	lastAddedMimeType      string
	deleteAttachmentCalls  int
	restoreAttachmentCalls int
	restoredAttachment     *attachment.Attachment
}

func (s *attachmentRepoStub) CheckTodoExists(ctx context.Context, userID string, todoID uuid.UUID) (*todo.Todo, error) {
	if s.checkTodoExistsFn != nil {
		return s.checkTodoExistsFn(ctx, userID, todoID)
	}

	return &todo.Todo{}, nil
}

func (s *attachmentRepoStub) AddAttachment(ctx context.Context, todoID uuid.UUID, userID string, s3Key string, fileName string, fileSize int64, mimeType string) (*attachment.Attachment, error) {
	s.lastAddedMimeType = mimeType
	if s.addAttachmentFn != nil {
		return s.addAttachmentFn(ctx, todoID, userID, s3Key, fileName, fileSize, mimeType)
	}

	return &attachment.Attachment{}, nil
}

func (s *attachmentRepoStub) GetAttachmentsByTodoID(ctx context.Context, todoID uuid.UUID) ([]attachment.Attachment, error) {
	if s.getAttachmentsByTodoID != nil {
		return s.getAttachmentsByTodoID(ctx, todoID)
	}

	return []attachment.Attachment{}, nil
}

func (s *attachmentRepoStub) GetTodoAttachment(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) (*attachment.Attachment, error) {
	if s.getTodoAttachmentFn != nil {
		return s.getTodoAttachmentFn(ctx, todoID, attachmentID)
	}

	return &attachment.Attachment{}, nil
}

func (s *attachmentRepoStub) DeleteAttachment(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) error {
	s.deleteAttachmentCalls++
	if s.deleteAttachmentFn != nil {
		return s.deleteAttachmentFn(ctx, todoID, attachmentID)
	}

	return nil
}

func (s *attachmentRepoStub) RestoreAttachment(ctx context.Context, item *attachment.Attachment) error {
	s.restoreAttachmentCalls++
	s.restoredAttachment = item
	if s.restoreAttachmentFn != nil {
		return s.restoreAttachmentFn(ctx, item)
	}

	return nil
}

type attachmentStoreStub struct {
	uploadFileFn       func(ctx context.Context, bucket, filename string, file io.Reader) (string, error)
	createPresignedURL func(ctx context.Context, bucket, fileKey string) (string, error)
	deleteFileFn       func(ctx context.Context, bucket, fileKey string) error

	uploadedBuckets []string
	uploadedFiles   []string
	deletedBuckets  []string
	deletedKeys     []string
	deleteCalls     int
}

func (s *attachmentStoreStub) UploadFile(ctx context.Context, bucket, filename string, file io.Reader) (string, error) {
	s.uploadedBuckets = append(s.uploadedBuckets, bucket)
	s.uploadedFiles = append(s.uploadedFiles, filename)
	if s.uploadFileFn != nil {
		return s.uploadFileFn(ctx, bucket, filename, file)
	}

	return "uploaded-key", nil
}

func (s *attachmentStoreStub) CreatePresignedURL(ctx context.Context, bucket, fileKey string) (string, error) {
	if s.createPresignedURL != nil {
		return s.createPresignedURL(ctx, bucket, fileKey)
	}

	return "https://example.com/file", nil
}

func (s *attachmentStoreStub) DeleteFile(ctx context.Context, bucket, fileKey string) error {
	s.deleteCalls++
	s.deletedBuckets = append(s.deletedBuckets, bucket)
	s.deletedKeys = append(s.deletedKeys, fileKey)
	if s.deleteFileFn != nil {
		return s.deleteFileFn(ctx, bucket, fileKey)
	}

	return nil
}

func TestUploadTodoAttachmentCleansUpS3WhenMetadataWriteFails(t *testing.T) {
	todoID := uuid.New()
	metadataErr := errors.New("insert attachment metadata failed")
	repo := &attachmentRepoStub{
		addAttachmentFn: func(ctx context.Context, receivedTodoID uuid.UUID, userID string, s3Key string, fileName string, fileSize int64, mimeType string) (*attachment.Attachment, error) {
			assert.Equal(t, todoID, receivedTodoID)
			assert.Equal(t, "user-1", userID)
			assert.Equal(t, "uploaded-key", s3Key)
			assert.Equal(t, "note.txt", fileName)
			assert.EqualValues(t, 4, fileSize)
			assert.NotEmpty(t, mimeType)
			return nil, metadataErr
		},
	}
	store := &attachmentStoreStub{
		uploadFileFn: func(ctx context.Context, bucket, filename string, file io.Reader) (string, error) {
			payload, err := io.ReadAll(file)
			require.NoError(t, err)
			assert.Equal(t, "attachments-bucket", bucket)
			assert.Equal(t, "todos/attachments/note.txt", filename)
			assert.Equal(t, []byte("tiny"), payload)
			return "uploaded-key", nil
		},
	}

	service := newTestTodoService(repo, store)
	fileHeader := newMultipartFileHeader(t, "file", "note.txt", []byte("tiny"))

	attachmentItem, err := service.UploadTodoAttachment(newTestEchoContext(), "user-1", todoID, fileHeader)

	require.Nil(t, attachmentItem)
	require.ErrorIs(t, err, metadataErr)
	assert.Equal(t, "attachments-bucket", store.deletedBuckets[0])
	assert.Equal(t, []string{"uploaded-key"}, store.deletedKeys)
	assert.Equal(t, 1, store.deleteCalls)
	assert.NotEmpty(t, repo.lastAddedMimeType)
}

func TestDeleteTodoAttachmentRestoresMetadataWhenS3DeleteFails(t *testing.T) {
	todoID := uuid.New()
	attachmentID := uuid.New()
	fileSize := int64(42)
	mimeType := "text/plain"
	now := time.Now().UTC()
	attachmentItem := &attachment.Attachment{
		Base: model.Base{
			BaseWithId:        model.BaseWithId{ID: attachmentID},
			BaseWithCreatedAt: model.BaseWithCreatedAt{CreatedAt: now},
			BaseWithUpdatedAt: model.BaseWithUpdatedAt{UpdatedAt: now},
		},
		TodoID:      todoID,
		Name:        "note.txt",
		UploadedBy:  "user-1",
		DownloadKey: "uploaded-key",
		FileSize:    &fileSize,
		MimeType:    &mimeType,
	}
	s3Err := errors.New("s3 delete failed")
	repo := &attachmentRepoStub{
		getTodoAttachmentFn: func(ctx context.Context, receivedTodoID uuid.UUID, receivedAttachmentID uuid.UUID) (*attachment.Attachment, error) {
			assert.Equal(t, todoID, receivedTodoID)
			assert.Equal(t, attachmentID, receivedAttachmentID)
			return attachmentItem, nil
		},
	}
	store := &attachmentStoreStub{
		deleteFileFn: func(ctx context.Context, bucket, fileKey string) error {
			assert.Equal(t, "attachments-bucket", bucket)
			assert.Equal(t, "uploaded-key", fileKey)
			return s3Err
		},
	}

	service := newTestTodoService(repo, store)

	err := service.DeleteTodoAttachment(newTestEchoContext(), "user-1", todoID, attachmentID)

	require.ErrorIs(t, err, s3Err)
	assert.Equal(t, 1, repo.deleteAttachmentCalls)
	assert.Equal(t, 1, repo.restoreAttachmentCalls)
	assert.Same(t, attachmentItem, repo.restoredAttachment)
	assert.Equal(t, 1, store.deleteCalls)
	assert.Equal(t, []string{"uploaded-key"}, store.deletedKeys)
}

func TestDeleteTodoAttachmentReturnsJoinedErrorWhenRestoreFails(t *testing.T) {
	todoID := uuid.New()
	attachmentID := uuid.New()
	attachmentItem := &attachment.Attachment{
		Base: model.Base{
			BaseWithId: model.BaseWithId{ID: attachmentID},
		},
		TodoID:      todoID,
		Name:        "note.txt",
		UploadedBy:  "user-1",
		DownloadKey: "uploaded-key",
	}
	s3Err := errors.New("s3 delete failed")
	restoreErr := errors.New("restore metadata failed")
	repo := &attachmentRepoStub{
		getTodoAttachmentFn: func(ctx context.Context, receivedTodoID uuid.UUID, receivedAttachmentID uuid.UUID) (*attachment.Attachment, error) {
			return attachmentItem, nil
		},
		restoreAttachmentFn: func(ctx context.Context, item *attachment.Attachment) error {
			return restoreErr
		},
	}
	store := &attachmentStoreStub{
		deleteFileFn: func(ctx context.Context, bucket, fileKey string) error {
			return s3Err
		},
	}

	service := newTestTodoService(repo, store)

	err := service.DeleteTodoAttachment(newTestEchoContext(), "user-1", todoID, attachmentID)

	require.Error(t, err)
	require.ErrorIs(t, err, s3Err)
	require.ErrorIs(t, err, restoreErr)
	assert.Equal(t, 1, store.deleteCalls)
	assert.Equal(t, 1, repo.restoreAttachmentCalls)
}

func newTestTodoService(repo todoAttachmentRepository, store attachmentFileStore) *TodoService {
	return &TodoService{
		server: &server.Server{
			Config: &config.Config{
				AWS: config.AWSConfig{
					UploadBucket: "attachments-bucket",
				},
			},
		},
		attachmentRepo: repo,
		fileStore:      store,
	}
}

func newTestEchoContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	logger := zerolog.Nop()
	ctx.Set(middleware.LoggerKey, &logger)

	return ctx
}

func newMultipartFileHeader(t *testing.T, fieldName string, fileName string, content []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile(fieldName, fileName)
	require.NoError(t, err)

	_, err = part.Write(content)
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(int64(len(content)+1024)))
	require.NotNil(t, req.MultipartForm)
	require.Contains(t, req.MultipartForm.File, fieldName)
	require.Len(t, req.MultipartForm.File[fieldName], 1)

	return req.MultipartForm.File[fieldName][0]
}
