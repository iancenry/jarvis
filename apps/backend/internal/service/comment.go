package service

import (
	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model/comment"
	"github.com/iancenry/jarvis/internal/repository"
	"github.com/iancenry/jarvis/internal/server"
	"github.com/labstack/echo/v4"
)

type CommentService struct {
	server      *server.Server
	commentRepo *repository.CommentRepository
	todoRepo    *repository.TodoRepository
}

func NewCommentService(s *server.Server, commentRepo *repository.CommentRepository, todoRepo *repository.TodoRepository) *CommentService {
	return &CommentService{
		server:      s,
		commentRepo: commentRepo,
		todoRepo:    todoRepo,
	}
}

func (cs *CommentService) CreateComment(ctx echo.Context, userID string, todoID uuid.UUID, payload *comment.CreateCommentPayload) (*comment.Comment, error) {
	logger := middleware.GetLogger(ctx)

	commentItem, err := cs.commentRepo.CreateComment(ctx.Request().Context(), userID, todoID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create comment")
		return nil, err
	}

	// business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "comment_created").
		Str("comment_id", commentItem.ID.String()).
		Str("todo_id", todoID.String()).
		Msg("Comment created")

	return commentItem, nil
}

func (cs *CommentService) GetCommentsByTodoID(ctx echo.Context, userID string, todoID uuid.UUID) ([]comment.Comment, error) {
	logger := middleware.GetLogger(ctx)

	// Validate the associated todo exists and belongs to the user
	_, err := cs.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed for fetching comments")
		return nil, err
	}

	comments, err := cs.commentRepo.GetCommentsByTodoID(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch comments")
		return nil, err
	}

	return comments, nil
}

func (cs *CommentService) UpdateComment(ctx echo.Context, userID string, commentID uuid.UUID, payload *comment.UpdateCommentPayload) (*comment.Comment, error) {
	logger := middleware.GetLogger(ctx)

	updatedComment, err := cs.commentRepo.UpdateComment(ctx.Request().Context(), userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update comment")
		return nil, err
	}

	// business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "comment_updated").
		Str("comment_id", commentID.String()).
		Msg("Comment updated")

	return updatedComment, nil
}

func (cs *CommentService) DeleteComment(ctx echo.Context, userID string, commentID uuid.UUID) error {
	logger := middleware.GetLogger(ctx)

	deletedComment, err := cs.commentRepo.DeleteComment(ctx.Request().Context(), userID, commentID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete comment")
		return err
	}

	// business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "comment_deleted").
		Str("comment_id", deletedComment.ID.String()).
		Msg("Comment deleted")

	return nil
}
