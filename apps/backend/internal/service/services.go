package service

import (
	"github.com/iancenry/jarvis/internal/lib/job"
	"github.com/iancenry/jarvis/internal/repository"
	"github.com/iancenry/jarvis/internal/server"
)

type Services struct {
	Auth *AuthService
	Job  *job.JobService
	Todo *TodoService
	Comment *CommentService
	Category *CategoryService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)
	todoService := NewTodoService(s, repos.Todo, repos.Category)
	commentService := NewCommentService(s, repos.Comment, repos.Todo)
	categoryService := NewCategoryService(s, repos.Category)

	return &Services{
		Job:  s.Job,
		Auth: authService,
		Category: categoryService,
		Todo: todoService,
		Comment: commentService,
	}, nil
}
