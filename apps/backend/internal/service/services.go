package service

import (
	"fmt"

	"github.com/iancenry/jarvis/internal/lib/aws"
	"github.com/iancenry/jarvis/internal/repository"
	"github.com/iancenry/jarvis/internal/server"
)

type Services struct {
	Auth     *AuthService
	Todo     *TodoService
	Comment  *CommentService
	Category *CategoryService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	var (
		awsClient    *aws.AWS
		uploadBucket string
		err          error
	)

	if s.Config.S3Enabled() {
		awsClient, err = aws.NewAWS(s.Config.AWS)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS client: %w", err)
		}
		uploadBucket = s.Config.AWS.UploadBucket
	}

	authService := NewAuthService(s.Config.Auth.SecretKey)
	todoService := NewTodoService(repos.Todo, repos.Category, awsClient, uploadBucket)
	commentService := NewCommentService(repos.Comment, repos.Todo)
	categoryService := NewCategoryService(repos.Category)

	return &Services{
		Auth:     authService,
		Category: categoryService,
		Todo:     todoService,
		Comment:  commentService,
	}, nil
}
