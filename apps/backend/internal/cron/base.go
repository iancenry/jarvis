package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/iancenry/jarvis/internal/config"
	"github.com/iancenry/jarvis/internal/database"
	joblib "github.com/iancenry/jarvis/internal/lib/job"
	"github.com/iancenry/jarvis/internal/logger"
	"github.com/iancenry/jarvis/internal/repository"
	"github.com/iancenry/jarvis/internal/server"
	"github.com/redis/go-redis/v9"
)

// JobContext holds the dependencies and resources needed to run cron jobs
type JobContext struct {
	Config        *config.Config
	Server        *server.Server
	JobClient     *asynq.Client
	Repositories  *repository.Repositories
	LoggerService *logger.LoggerService
}

// NewJobContext initializes the JobContext by loading configuration, setting up database and Redis connections, and initializing repositories
func NewJobContext() (*JobContext, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	if err := cfg.ValidateForCron(); err != nil {
		return nil, fmt.Errorf("invalid cron config: %w", err)
	}

	loggerService := logger.NewLoggerService(cfg.Observability)
	loggerInstance := logger.NewLoggerWithService(cfg.Observability, loggerService)

	db, err := database.New(cfg, &loggerInstance, loggerService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis for cron jobs: %w", err)
	}

	srv := &server.Server{
		Config:        cfg,
		Logger:        &loggerInstance,
		LoggerService: loggerService,
		DB:            db,
		Redis:         redisClient,
	}

	jobClient, err := initJobClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize job client: %w", err)
	}

	repositories := repository.NewRepositories(srv.DB)

	return &JobContext{
		Config:        cfg,
		Server:        srv,
		JobClient:     jobClient,
		Repositories:  repositories,
		LoggerService: loggerService,
	}, nil
}

func (c *JobContext) Close() {
	if c.Server != nil && c.Server.DB != nil {
		c.Server.DB.Pool.Close()
	}
	if c.Server != nil && c.Server.Redis != nil {
		c.Server.Redis.Close()
	}
	if c.JobClient != nil {
		c.JobClient.Close()
	}
	if c.LoggerService != nil {
		c.LoggerService.Shutdown()
	}
}

// initJobClient initializes the Asynq client for enqueuing background jobs
func initJobClient(cfg *config.Config) (*asynq.Client, error) {
	client := asynq.NewClient(joblib.NewRedisClientOpt(cfg))
	return client, nil
}

// Job defines the interface that all cron jobs must implement
type Job interface {
	Name() string
	Description() string
	Run(ctx context.Context, jobCtx *JobContext) error
}

type JobRunner struct {
	job Job
	ctx *JobContext
}

// NewJobRunner creates a new JobRunner with the given Job and initializes the JobContext
func NewJobRunner(job Job) (*JobRunner, error) {
	ctx, err := NewJobContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create job context: %w", err)
	}

	return &JobRunner{
		job: job,
		ctx: ctx,
	}, nil
}

// Run executes the cron job and handles logging and error reporting
func (r *JobRunner) Run() error {
	defer r.ctx.Close()

	r.ctx.Server.Logger.Info().
		Str("job", r.job.Name()).
		Msg("Starting cron job")

	ctx := context.Background()
	err := r.job.Run(ctx, r.ctx)
	if err != nil {
		r.ctx.Server.Logger.Error().
			Err(err).
			Str("job", r.job.Name()).
			Msg("Failed to run cron job")
		return err
	}

	r.ctx.Server.Logger.Info().
		Str("job", r.job.Name()).
		Msg("Cron job completed successfully")
	return nil
}
