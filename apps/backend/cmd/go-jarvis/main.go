package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/iancenry/jarvis/internal/config"
	"github.com/iancenry/jarvis/internal/database"
	"github.com/iancenry/jarvis/internal/handler"
	"github.com/iancenry/jarvis/internal/logger"
	"github.com/iancenry/jarvis/internal/repository"
	"github.com/iancenry/jarvis/internal/router"
	"github.com/iancenry/jarvis/internal/server"
	"github.com/iancenry/jarvis/internal/service"
	"github.com/rs/zerolog"
)

const DefaultContextTimeout = 30

// API: serves HTTP, talks to Postgres, optionally knows about Redis for health/integrations.
func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.LoadConfig()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}

	if err := cfg.ValidateForAPI(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid API config: %v\n", err)
		return 1
	}

	loggerService := logger.NewLoggerService(cfg.Observability)
	defer loggerService.Shutdown()

	log := logger.NewLoggerWithService(cfg.Observability, loggerService)
	if err := serve(cfg, &log, loggerService); err != nil {
		log.Error().Err(err).Msg("api server exited with error")
		return 1
	}

	return 0
}

func serve(cfg *config.Config, log *zerolog.Logger, loggerService *logger.LoggerService) error {
	if cfg.Primary.Env != "local" {
		if err := database.Migrate(context.Background(), log, cfg); err != nil {
			return fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	srv, err := server.New(cfg, log, loggerService)
	if err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	repos := repository.NewRepositories(srv.DB)
	services, err := service.NewServices(srv, repos)
	if err != nil {
		return fmt.Errorf("could not create services: %w", err)
	}
	handlers := handler.NewHandlers(srv, services)
	r := router.NewRouter(srv, handlers, services)
	srv.SetupHTTPServer(r)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- srv.Start()
	}()

	select {
	case err := <-serverErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed to start server: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), DefaultContextTimeout*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	if err := <-serverErrCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server exited with error: %w", err)
	}

	log.Info().Msg("server exited properly")
	return nil
}
