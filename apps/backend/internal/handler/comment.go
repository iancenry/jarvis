package handler

import (
	"net/http"

	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model/comment"
	"github.com/iancenry/jarvis/internal/service"
	"github.com/labstack/echo/v4"
)

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

func (h *CommentHandler) AddComment(c echo.Context) error {
	return Handle(
		func(c echo.Context, payload *comment.CreateCommentPayload) (*comment.Comment, error) {
			userID := middleware.GetUserID(c)
			return h.commentService.CreateComment(c.Request().Context(), userID, payload.TodoID, payload)
		},
		http.StatusCreated,
		&comment.CreateCommentPayload{},
	)(c)
}

func (h *CommentHandler) GetCommentsByTodoID(c echo.Context) error {
	return Handle(
		func(c echo.Context, payload *comment.GetCommentsByTodoIDQuery) ([]comment.Comment, error) {
			userID := middleware.GetUserID(c)
			return h.commentService.GetCommentsByTodoID(c.Request().Context(), userID, payload.TodoID)
		},
		http.StatusOK,
		&comment.GetCommentsByTodoIDQuery{},
	)(c)
}

func (h *CommentHandler) UpdateComment(c echo.Context) error {
	return Handle(
		func(c echo.Context, payload *comment.UpdateCommentPayload) (*comment.Comment, error) {
			userID := middleware.GetUserID(c)
			return h.commentService.UpdateComment(c.Request().Context(), userID, payload.ID, payload)
		},
		http.StatusOK,
		&comment.UpdateCommentPayload{},
	)(c)
}

func (h *CommentHandler) DeleteComment(c echo.Context) error {
	return HandleNoContent(
		func(c echo.Context, payload *comment.DeleteCommentByIDQuery) error {
			userID := middleware.GetUserID(c)
			return h.commentService.DeleteComment(c.Request().Context(), userID, payload.ID)
		},
		http.StatusNoContent,
		&comment.DeleteCommentByIDQuery{},
	)(c)
}
