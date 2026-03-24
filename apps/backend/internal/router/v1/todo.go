package v1

import (
	"github.com/iancenry/jarvis/internal/handler"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/labstack/echo/v4"
)

func registerTodoRoutes(r *echo.Group, h *handler.TodoHandler, ch *handler.CommentHandler, auth *middleware.AuthMiddleware) {
	// Todo operations
	todos := r.Group("/todos")
	todos.Use(auth.RequireAuth)

	// Collection operations
	todos.POST("", h.CreateTodo)
	todos.GET("", h.GetTodos)
	todos.GET("/stats", h.GetTodoStats)

	// Individual todo operations
	dynamicTodo := todos.Group("/:id")
	dynamicTodo.GET("", h.GetTodoByID)
	dynamicTodo.PATCH("", h.UpdateTodo)
	dynamicTodo.DELETE("", h.DeleteTodo)

	// Todo comments
	todoComments := dynamicTodo.Group("/comments")
	todoComments.POST("", ch.AddComment)
	todoComments.GET("", ch.GetCommentsByTodoID)
 
	// Todo attachments
	todoAttachments := dynamicTodo.Group("/attachments")
	todoAttachments.POST("", h.AddAttachment)
	todoAttachments.DELETE("/:attachmentId", h.DeleteAttachment)
	todoAttachments.GET("/:attachmentId/download", h.GetAttachmentPresignedURL)
}