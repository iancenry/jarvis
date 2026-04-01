package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/model/attachment"
	"github.com/jackc/pgx/v5"
)

func (r *TodoRepository) AddAttachment(ctx context.Context, todoID uuid.UUID, userID string, s3Key string, fileName string, fileSize int64, mimeType string) (*attachment.Attachment, error) {
	stmt := `
	INSERT INTO attachments (
		todo_id,
		name,
		uploaded_by,
		download_key,
		file_size,
		mime_type
	)
	VALUES (
		@todo_id,
		@name,
		@uploaded_by,
		@download_key,
		@file_size,
		@mime_type
	)
	RETURNING *
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"todo_id":      todoID,
		"name":         fileName,
		"uploaded_by":  userID,
		"download_key": s3Key,
		"file_size":    fileSize,
		"mime_type":    mimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute add attachment query for todo_id=%s filename=%s: %w", todoID, fileName, err)
	}

	attachmentItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[attachment.Attachment])
	if err != nil {
		return nil, fmt.Errorf("failed to collect added attachment for todo_id=%s filename=%s: %w", todoID, fileName, err)
	}

	return &attachmentItem, nil
}

func (r *TodoRepository) GetTodoAttachment(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) (*attachment.Attachment, error) {
	stmt := `
	SELECT * FROM attachments
	WHERE id = @id AND todo_id = @todo_id
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      attachmentID,
		"todo_id": todoID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get todo attachment query for todo_id=%s attachment_id=%s: %w", todoID, attachmentID, err)
	}

	attachmentItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[attachment.Attachment])
	if err != nil {
		if isNoRowsError(err) {
			return nil, newDomainNotFoundError("ATTACHMENT")
		}
		return nil, fmt.Errorf("failed to collect todo attachment for todo_id=%s attachment_id=%s: %w", todoID, attachmentID, err)
	}

	return &attachmentItem, nil
}

func (r *TodoRepository) GetAttachmentsByTodoID(ctx context.Context, todoID uuid.UUID) ([]attachment.Attachment, error) {
	stmt := `
	SELECT * FROM attachments
	WHERE todo_id = @todo_id
	ORDER BY created_at ASC
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"todo_id": todoID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get attachments by todo id query for todo_id=%s: %w", todoID, err)
	}

	attachments, err := pgx.CollectRows(rows, pgx.RowToStructByName[attachment.Attachment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []attachment.Attachment{}, nil
		}
		return nil, fmt.Errorf("failed to collect attachments by todo id for todo_id=%s: %w", todoID, err)
	}

	return attachments, nil
}

func (r *TodoRepository) DeleteAttachment(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) error {
	stmt := `DELETE FROM attachments WHERE id = @id AND todo_id = @todo_id`

	result, err := r.db.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":      attachmentID,
		"todo_id": todoID,
	})
	if err != nil {
		return fmt.Errorf("failed to execute delete attachment query for todo_id=%s attachment_id=%s: %w", todoID, attachmentID, err)
	}

	if result.RowsAffected() == 0 {
		return newDomainNotFoundError("ATTACHMENT")
	}

	return nil
}

// RestoreAttachment is used to restore attachment metadata when an attachment is deleted and then restored from S3.
//
// Since we perform a hard delete on the attachment record in the database when an attachment is deleted, we need to restore the metadata when the attachment is restored from S3.
func (r *TodoRepository) RestoreAttachment(ctx context.Context, item *attachment.Attachment) error {
	stmt := `
	INSERT INTO attachments (
		id,
		created_at,
		updated_at,
		todo_id,
		name,
		uploaded_by,
		download_key,
		file_size,
		mime_type
	)
	VALUES (
		@id,
		@created_at,
		@updated_at,
		@todo_id,
		@name,
		@uploaded_by,
		@download_key,
		@file_size,
		@mime_type
	)
	`

	_, err := r.db.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":           item.ID,
		"created_at":   item.CreatedAt,
		"updated_at":   item.UpdatedAt,
		"todo_id":      item.TodoID,
		"name":         item.Name,
		"uploaded_by":  item.UploadedBy,
		"download_key": item.DownloadKey,
		"file_size":    item.FileSize,
		"mime_type":    item.MimeType,
	})
	if err != nil {
		return fmt.Errorf("failed to restore attachment metadata for attachment_id=%s todo_id=%s: %w", item.ID, item.TodoID, err)
	}

	return nil
}
