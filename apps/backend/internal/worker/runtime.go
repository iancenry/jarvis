package worker

import (
	"context"
	"errors"

	"github.com/iancenry/jarvis/internal/config"
	"github.com/iancenry/jarvis/internal/lib/job"
	"github.com/iancenry/jarvis/internal/service"
	"github.com/rs/zerolog"
)

type Runtime struct {
	logger     *zerolog.Logger
	jobService *job.JobService
	enabled    bool
}

func New(cfg *config.Config, logger *zerolog.Logger) (*Runtime, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	if err := cfg.ValidateForWorker(); err != nil {
		return nil, err
	}

	if !cfg.WorkerEnabled() {
		return &Runtime{
			logger:  logger,
			enabled: false,
		}, nil
	}

	jobService := job.NewJobService(logger, cfg)
	jobService.SetAuthService(service.NewAuthService(cfg.Auth.SecretKey))

	return &Runtime{
		logger:     logger,
		jobService: jobService,
		enabled:    true,
	}, nil
}

func (r *Runtime) Enabled() bool {
	return r != nil && r.enabled
}

func (r *Runtime) Start() error {
	if !r.Enabled() {
		if r != nil && r.logger != nil {
			r.logger.Info().Msg("background worker disabled; skipping worker startup")
		}
		return nil
	}

	r.logger.Info().Msg("starting background worker")
	return r.jobService.Start()
}

func (r *Runtime) Shutdown(ctx context.Context) error {
	_ = ctx

	if !r.Enabled() {
		return nil
	}

	r.logger.Info().Msg("stopping background worker")
	r.jobService.Stop()
	return nil
}
