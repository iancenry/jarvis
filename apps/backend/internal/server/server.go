package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/iancenry/jarvis/internal/config"
	"github.com/iancenry/jarvis/internal/database"
	loggerPkg "github.com/iancenry/jarvis/internal/logger"
	"github.com/newrelic/go-agent/v3/integrations/nrredis-v9"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Server struct {
	Config        *config.Config
	Logger        *zerolog.Logger
	LoggerService *loggerPkg.LoggerService
	DB            *database.Database
	Redis         *redis.Client
	httpServer    *http.Server
}

func New(cfg *config.Config, logger *zerolog.Logger, loggerService *loggerPkg.LoggerService) (*Server, error) {
	db, err := database.New(cfg, logger, loggerService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	var redisClient *redis.Client

	if cfg.Redis.Address != "" {
		// Optional Redis client for API-side integrations and health reporting.
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Address,
			Password: cfg.Redis.Password,
			DB:       0,
		})

		// Add New Relic Redis hooks if available
		if loggerService != nil && loggerService.GetApplication() != nil {
			redisClient.AddHook(nrredis.NewHook(redisClient.Options()))
		}

		// Test Redis connection and fail fast when Redis is explicitly configured.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("failed to connect to configured Redis: %w", err)
		}

		logger.Info().Str("redis_addr", cfg.Redis.Address).Msg("connected to redis")
	} else {
		logger.Info().Msg("redis not configured; skipping redis initialization")
	}

	server := &Server{
		Config:        cfg,
		Logger:        logger,
		LoggerService: loggerService,
		DB:            db,
		Redis:         redisClient,
	}

	// Start metrics collection
	// Runtime metrics are automatically collected by New Relic Go agent

	return server, nil
}

func (s *Server) SetupHTTPServer(handler http.Handler) {
	s.httpServer = &http.Server{
		Addr:         ":" + s.Config.Server.Port,
		Handler:      handler,
		ReadTimeout:  time.Duration(s.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.Config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.Config.Server.IdleTimeout) * time.Second,
	}
}

func (s *Server) Start() error {
	if s.httpServer == nil {
		return fmt.Errorf("HTTP server not initialized")
	}

	s.Logger.Info().
		Str("port", s.Config.Server.Port).
		Str("env", s.Config.Primary.Env).
		Msg("starting server")

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	if s.Redis != nil {
		if err := s.Redis.Close(); err != nil {
			return fmt.Errorf("failed to close redis connection: %w", err)
		}
	}

	if err := s.DB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}
