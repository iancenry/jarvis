package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model/comment"
	"github.com/iancenry/jarvis/internal/repository"
)

type CommentService struct {
	commentRepo *repository.CommentRepository
	todoRepo    *repository.TodoRepository
}

func NewCommentService(commentRepo *repository.CommentRepository, todoRepo *repository.TodoRepository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		todoRepo:    todoRepo,
	}
}

func (cs *CommentService) CreateComment(ctx context.Context, userID string, todoID uuid.UUID, payload *comment.CreateCommentPayload) (*comment.Comment, error) {
	logger := middleware.LoggerFromContext(ctx)

	commentItem, err := cs.commentRepo.CreateComment(ctx, userID, todoID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create comment")
		return nil, err
	}

	// business event log
	eventLogger := middleware.LoggerFromContext(ctx)
	eventLogger.Info().
		Str("event", "comment_created").
		Str("comment_id", commentItem.ID.String()).
		Str("todo_id", todoID.String()).
		Msg("Comment created")

	return commentItem, nil
}

func (cs *CommentService) GetCommentsByTodoID(ctx context.Context, userID string, todoID uuid.UUID) ([]comment.Comment, error) {
	logger := middleware.LoggerFromContext(ctx)

	// Validate the associated todo exists and belongs to the user
	_, err := cs.todoRepo.CheckTodoExists(ctx, userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed for fetching comments")
		return nil, err
	}

	comments, err := cs.commentRepo.GetCommentsByTodoID(ctx, userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch comments")
		return nil, err
	}

	return comments, nil
}

func (cs *CommentService) UpdateComment(ctx context.Context, userID string, commentID uuid.UUID, payload *comment.UpdateCommentPayload) (*comment.Comment, error) {
	logger := middleware.LoggerFromContext(ctx)

	updatedComment, err := cs.commentRepo.UpdateComment(ctx, userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update comment")
		return nil, err
	}

	// business event log
	eventLogger := middleware.LoggerFromContext(ctx)
	eventLogger.Info().
		Str("event", "comment_updated").
		Str("comment_id", commentID.String()).
		Msg("Comment updated")

	return updatedComment, nil
}

func (cs *CommentService) DeleteComment(ctx context.Context, userID string, commentID uuid.UUID) error {
	logger := middleware.LoggerFromContext(ctx)

	deletedComment, err := cs.commentRepo.DeleteComment(ctx, userID, commentID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete comment")
		return err
	}

	// business event log
	eventLogger := middleware.LoggerFromContext(ctx)
	eventLogger.Info().
		Str("event", "comment_deleted").
		Str("comment_id", deletedComment.ID.String()).
		Msg("Comment deleted")

	return nil
}
