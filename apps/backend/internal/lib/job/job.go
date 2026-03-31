package job

import (
	"context"
	"errors"

	"github.com/hibiken/asynq"
	"github.com/iancenry/jarvis/internal/config"
	"github.com/iancenry/jarvis/internal/lib/email"
	"github.com/rs/zerolog"
)

type JobService struct {
	Client      *asynq.Client
	server      *asynq.Server
	logger      *zerolog.Logger
	authService AuthServiceInterface
	emailClient *email.Client
	started     bool
}

type AuthServiceInterface interface {
	GetUserEmail(ctx context.Context, userID string) (string, error)
}

func NewRedisClientOpt(cfg *config.Config) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       0,
	}
}

func NewJobService(logger *zerolog.Logger, cfg *config.Config) *JobService {
	redisOpt := NewRedisClientOpt(cfg)

	client := asynq.NewClient(redisOpt)

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6, // Higher priority queue for important emails
				"default":  3, // Default priority for most emails
				"low":      1, // Lower priority for non-urgent emails
			},
		},
	)

	return &JobService{
		Client:      client,
		server:      server,
		logger:      logger,
		emailClient: email.NewClient(cfg, logger),
	}
}

func (j *JobService) SetAuthService(authService AuthServiceInterface) {
	j.authService = authService
}

func (j *JobService) validateDependencies() error {
	if j.authService == nil {
		return errors.New("job service auth dependency not configured")
	}

	if j.emailClient == nil {
		return errors.New("job service email dependency not configured")
	}

	return nil
}

func (j *JobService) Start() error {
	if j.started {
		return nil
	}

	if err := j.validateDependencies(); err != nil {
		return err
	}

	// Register task handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskWelcome, j.handleWelcomeEmailTask)
	mux.HandleFunc(TaskReminderEmail, j.handleReminderEmailTask)
	mux.HandleFunc(TaskWeeklyReportEmail, j.handleWeeklyReportEmailTask)

	j.logger.Info().Msg("Starting background job server")
	if err := j.server.Start(mux); err != nil {
		return err
	}

	j.started = true
	return nil
}

func (j *JobService) Stop() {
	if j == nil {
		return
	}

	j.logger.Info().Msg("Stopping background job server")
	if j.started {
		j.server.Shutdown()
		j.started = false
	}
	if j.Client != nil {
		j.Client.Close()
	}
}
