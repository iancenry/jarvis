package handler

import (
	"github.com/iancenry/jarvis/internal/server"
	"github.com/iancenry/jarvis/internal/service"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
	Todo    *TodoHandler
	Category *CategoryHandler
	Comment  *CommentHandler
}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
		Todo:    NewTodoHandler(s, services.Todo),
		Category: NewCategoryHandler(s, services.Category),
		Comment:  NewCommentHandler(s, services.Comment),
	}
}
