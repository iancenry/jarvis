package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/errs"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/attachment"
	"github.com/iancenry/jarvis/internal/model/todo"
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
	uploadFileFn       func(ctx context.Context, bucket, fileKey string, file io.Reader, fileSize int64, contentType string) (string, error)
	createPresignedURL func(ctx context.Context, bucket, fileKey string) (string, error)
	deleteFileFn       func(ctx context.Context, bucket, fileKey string) error

	uploadedBuckets []string
	uploadedKeys    []string
	uploadedSizes   []int64
	uploadedTypes   []string
	deletedBuckets  []string
	deletedKeys     []string
	deleteCalls     int
}

func (s *attachmentStoreStub) UploadFile(ctx context.Context, bucket, fileKey string, file io.Reader, fileSize int64, contentType string) (string, error) {
	s.uploadedBuckets = append(s.uploadedBuckets, bucket)
	s.uploadedKeys = append(s.uploadedKeys, fileKey)
	s.uploadedSizes = append(s.uploadedSizes, fileSize)
	s.uploadedTypes = append(s.uploadedTypes, contentType)
	if s.uploadFileFn != nil {
		return s.uploadFileFn(ctx, bucket, fileKey, file, fileSize, contentType)
	}

	return fileKey, nil
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

func TestUploadTodoAttachmentSanitizesNameAndUsesUUIDStorageKey(t *testing.T) {
	todoID := uuid.New()
	uploadedPNG := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x00, 0x00, 0x00, 0x0d}
	repo := &attachmentRepoStub{
		addAttachmentFn: func(ctx context.Context, receivedTodoID uuid.UUID, userID string, s3Key string, fileName string, fileSize int64, mimeType string) (*attachment.Attachment, error) {
			assert.Equal(t, todoID, receivedTodoID)
			assert.Equal(t, "Avatar-Final.png", fileName)
			assert.Equal(t, "image/png", mimeType)
			return &attachment.Attachment{
				Base: model.Base{
					BaseWithId: model.BaseWithId{ID: uuid.New()},
				},
				Name:        fileName,
				DownloadKey: s3Key,
			}, nil
		},
	}
	store := &attachmentStoreStub{
		uploadFileFn: func(ctx context.Context, bucket, fileKey string, file io.Reader, fileSize int64, contentType string) (string, error) {
			payload, err := io.ReadAll(file)
			require.NoError(t, err)
			assert.Equal(t, uploadedPNG, payload)
			assert.Equal(t, "attachments-bucket", bucket)
			assert.EqualValues(t, len(uploadedPNG), fileSize)
			assert.Equal(t, "image/png", contentType)
			return fileKey, nil
		},
	}

	service := newTestTodoService(repo, store)
	fileHeader := newMultipartFileHeader(t, "file", "../Avatar Final!.PNG", uploadedPNG)

	attachmentItem, err := service.UploadTodoAttachment(newTestServiceContext(), "user-1", todoID, fileHeader)

	require.NoError(t, err)
	require.NotNil(t, attachmentItem)
	require.Len(t, store.uploadedKeys, 1)
	assert.Regexp(t, regexp.MustCompile(`^todos/attachments/[0-9a-f-]+\.png$`), store.uploadedKeys[0])
	assert.Equal(t, "image/png", store.uploadedTypes[0])
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
		uploadFileFn: func(ctx context.Context, bucket, fileKey string, file io.Reader, fileSize int64, contentType string) (string, error) {
			payload, err := io.ReadAll(file)
			require.NoError(t, err)
			assert.Equal(t, "attachments-bucket", bucket)
			assert.Regexp(t, regexp.MustCompile(`^todos/attachments/[0-9a-f-]+\.txt$`), fileKey)
			assert.EqualValues(t, 4, fileSize)
			assert.Equal(t, "text/plain", contentType)
			assert.Equal(t, []byte("tiny"), payload)
			return "uploaded-key", nil
		},
	}

	service := newTestTodoService(repo, store)
	fileHeader := newMultipartFileHeader(t, "file", "note.txt", []byte("tiny"))

	attachmentItem, err := service.UploadTodoAttachment(newTestServiceContext(), "user-1", todoID, fileHeader)

	require.Nil(t, attachmentItem)
	require.ErrorIs(t, err, metadataErr)
	assert.Equal(t, "attachments-bucket", store.deletedBuckets[0])
	assert.Equal(t, []string{"uploaded-key"}, store.deletedKeys)
	assert.Equal(t, 1, store.deleteCalls)
	assert.NotEmpty(t, repo.lastAddedMimeType)
}

func TestUploadTodoAttachmentRejectsOversizedFiles(t *testing.T) {
	service := newTestTodoService(&attachmentRepoStub{}, &attachmentStoreStub{})

	attachmentItem, err := service.UploadTodoAttachment(
		newTestServiceContext(),
		"user-1",
		uuid.New(),
		&multipart.FileHeader{
			Filename: "note.txt",
			Size:     MaxAttachmentSizeBytes + 1,
		},
	)

	require.Nil(t, attachmentItem)
	require.Error(t, err)

	var httpErr *errs.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, "ATTACHMENT_FILE_TOO_LARGE", httpErr.Code)
}

func TestUploadTodoAttachmentReturnsServiceUnavailableWhenStorageDisabled(t *testing.T) {
	service := newTestTodoService(&attachmentRepoStub{}, nil)
	fileHeader := newMultipartFileHeader(t, "file", "note.txt", []byte("tiny"))

	attachmentItem, err := service.UploadTodoAttachment(newTestServiceContext(), "user-1", uuid.New(), fileHeader)

	require.Nil(t, attachmentItem)
	require.Error(t, err)

	var httpErr *errs.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 503, httpErr.Status)
	assert.Equal(t, "ATTACHMENT_STORAGE_DISABLED", httpErr.Code)
}

func TestUploadTodoAttachmentRejectsUnsupportedContentTypes(t *testing.T) {
	repo := &attachmentRepoStub{}
	store := &attachmentStoreStub{}
	service := newTestTodoService(repo, store)
	fileHeader := newMultipartFileHeader(t, "file", "payload.bin", []byte{0x00, 0x01, 0x02, 0x03})

	attachmentItem, err := service.UploadTodoAttachment(newTestServiceContext(), "user-1", uuid.New(), fileHeader)

	require.Nil(t, attachmentItem)
	require.Error(t, err)

	var httpErr *errs.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, "ATTACHMENT_FILE_TYPE_NOT_ALLOWED", httpErr.Code)
	assert.Empty(t, store.uploadedKeys)
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

	err := service.DeleteTodoAttachment(newTestServiceContext(), "user-1", todoID, attachmentID)

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

	err := service.DeleteTodoAttachment(newTestServiceContext(), "user-1", todoID, attachmentID)

	require.Error(t, err)
	require.ErrorIs(t, err, s3Err)
	require.ErrorIs(t, err, restoreErr)
	assert.Equal(t, 1, store.deleteCalls)
	assert.Equal(t, 1, repo.restoreAttachmentCalls)
}

func newTestTodoService(repo todoAttachmentRepository, store attachmentFileStore) *TodoService {
	return &TodoService{
		attachmentRepo: repo,
		fileStore:      store,
		uploadBucket:   "attachments-bucket",
	}
}

func newTestServiceContext() context.Context {
	logger := zerolog.Nop()
	return context.WithValue(context.Background(), middleware.LoggerKey, &logger)
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
