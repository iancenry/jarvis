package repository

import "github.com/iancenry/jarvis/internal/server"

// Repositories aggregates all individual repositories for easy access
type Repositories struct{
	Todo *TodoRepository
	Comment *CommentRepository
	Category *CategoryRepository
}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		Todo: NewTodoRepository(s),
		Comment: NewCommentRepository(s),
		Category: NewCategoryRepository(s),
	}
}
