package service

import (
	"fmt"

	"github.com/iancenry/jarvis/internal/lib/aws"
	"github.com/iancenry/jarvis/internal/lib/job"
	"github.com/iancenry/jarvis/internal/repository"
	"github.com/iancenry/jarvis/internal/server"
)

type Services struct {
	Auth     *AuthService
	Job      *job.JobService
	Todo     *TodoService
	Comment  *CommentService
	Category *CategoryService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	awsClient, err := aws.NewAWS(s.Config.AWS)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client: %w", err)
	}

	authService := NewAuthService(s.Config.Auth.SecretKey)
	todoService := NewTodoService(repos.Todo, repos.Category, awsClient, s.Config.AWS.UploadBucket)
	commentService := NewCommentService(repos.Comment, repos.Todo)
	categoryService := NewCategoryService(repos.Category)

	// Inject AuthService into JobService for email tasks that require user information
	if s.Job != nil {
		s.Job.SetAuthService(authService)
		if err := s.Job.Start(); err != nil {
			return nil, fmt.Errorf("failed to start job service: %w", err)
		}
	}

	return &Services{
		Job:      s.Job,
		Auth:     authService,
		Category: categoryService,
		Todo:     todoService,
		Comment:  commentService,
	}, nil
}
