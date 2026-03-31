package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/database"
	"github.com/iancenry/jarvis/internal/model/comment"
	"github.com/jackc/pgx/v5"
)

type CommentRepository struct {
	db *database.Database
}

func NewCommentRepository(db *database.Database) *CommentRepository {
	return &CommentRepository{
		db: db,
	}
}

func (r *CommentRepository) CreateComment(ctx context.Context, userID string, todoID uuid.UUID, payload *comment.CreateCommentPayload) (*comment.Comment, error) {
	stmt := `
		INSERT INTO todo_comments (todo_id, user_id, content)
		SELECT t.id, @user_id, @content
		FROM todos t
		WHERE t.id = @todo_id AND t.user_id = @user_id
		RETURNING *
		`
	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"todo_id": todoID,
		"user_id": userID,
		"content": payload.Content,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute create comment query for user_id=%s todo_id=%s: %w", userID, todoID, err)
	}
	comment, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[comment.Comment])
	if err != nil {
		if isNoRowsError(err) {
			return nil, newDomainNotFoundError("TODO")
		}
		return nil, fmt.Errorf("failed to collect created comment for user_id=%s todo_id=%s: %w", userID, todoID, err)
	}

	return &comment, nil
}

func (r *CommentRepository) GetCommentsByTodoID(ctx context.Context, userID string, todoID uuid.UUID) ([]comment.Comment, error) {
	stmt := `
		SELECT * FROM todo_comments
		WHERE todo_id = @todo_id AND user_id = @user_id
		ORDER BY created_at ASC
	`
	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"todo_id": todoID,
		"user_id": userID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute get comments by todo id query for user_id=%s todo_id=%s: %w", userID, todoID, err)
	}

	comments, err := pgx.CollectRows(rows, pgx.RowToStructByName[comment.Comment])
	if err != nil {
		return nil, fmt.Errorf("failed to collect comments by todo id for user_id=%s todo_id=%s: %w", userID, todoID, err)
	}

	return comments, nil
}

func (r *CommentRepository) GetCommentByID(ctx context.Context, userID string, commentID uuid.UUID) (*comment.Comment, error) {
	stmt := `
		SELECT * FROM todo_comments
		WHERE id = @id AND user_id = @user_id
	`
	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      commentID,
		"user_id": userID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute get comment by id query for user_id=%s comment_id=%s: %w", userID, commentID, err)
	}

	comment, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[comment.Comment])
	if err != nil {
		if isNoRowsError(err) {
			return nil, newDomainNotFoundError("COMMENT")
		}
		return nil, fmt.Errorf("failed to collect comment by id for user_id=%s comment_id=%s: %w", userID, commentID, err)
	}

	return &comment, nil
}

func (r *CommentRepository) UpdateComment(ctx context.Context, userID string, payload *comment.UpdateCommentPayload) (*comment.Comment, error) {
	// updated at is automatically set to current timestamp by the database trigger
	stmt := `
		UPDATE todo_comments
		SET content = @content
		WHERE id = @id AND user_id = @user_id
		RETURNING *
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      payload.ID,
		"user_id": userID,
		"content": payload.Content,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute update comment query for user_id=%s comment_id=%s: %w", userID, payload.ID, err)
	}

	updatedComment, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[comment.Comment])
	if err != nil {
		if isNoRowsError(err) {
			return nil, newDomainNotFoundError("COMMENT")
		}
		return nil, fmt.Errorf("failed to collect updated comment for user_id=%s comment_id=%s: %w", userID, payload.ID, err)
	}

	return &updatedComment, nil
}

func (r *CommentRepository) DeleteComment(ctx context.Context, userID string, commentID uuid.UUID) (*comment.Comment, error) {
	stmt := `
		DELETE FROM todo_comments
		WHERE id = @id AND user_id = @user_id
		RETURNING *
	`
	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      commentID,
		"user_id": userID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute delete comment query for user_id=%s comment_id=%s: %w", userID, commentID, err)
	}

	deletedComment, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[comment.Comment])
	if err != nil {
		if isNoRowsError(err) {
			return nil, newDomainNotFoundError("COMMENT")
		}
		return nil, fmt.Errorf("failed to collect deleted comment for user_id=%s comment_id=%s: %w", userID, commentID, err)
	}

	return &deletedComment, nil
}
