package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iancenry/jarvis/internal/config"
	"github.com/iancenry/jarvis/internal/logger"
	"github.com/iancenry/jarvis/internal/worker"
)

const DefaultShutdownTimeout = 30

// Worker: consumes Asynq jobs from Redis.
func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.LoadConfig()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}

	if err := cfg.ValidateForWorker(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid worker config: %v\n", err)
		return 1
	}

	loggerService := logger.NewLoggerService(cfg.Observability)
	defer loggerService.Shutdown()

	log := logger.NewLoggerWithService(cfg.Observability, loggerService)

	runtime, err := worker.New(cfg, &log)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize worker")
		return 1
	}

	if !runtime.Enabled() {
		log.Info().Msg("background worker disabled; exiting")
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runtime.Start(); err != nil {
		log.Error().Err(err).Msg("failed to start worker")
		return 1
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout*time.Second)
	defer cancel()

	if err := runtime.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("worker forced to shutdown")
		return 1
	}

	log.Info().Msg("worker exited properly")
	return 0
}
