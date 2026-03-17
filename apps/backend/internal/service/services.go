package service

import (
	"github.com/iancenry/jarvis/internal/lib/job"
	"github.com/iancenry/jarvis/internal/repository"
	"github.com/iancenry/jarvis/internal/server"
)

type Services struct {
	Auth *AuthService
	Job  *job.JobService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)

	return &Services{
		Job:  s.Job,
		Auth: authService,
	}, nil
}
