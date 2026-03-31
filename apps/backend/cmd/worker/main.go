package main

import (
	"context"
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
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	loggerService := logger.NewLoggerService(cfg.Observability)
	defer loggerService.Shutdown()

	log := logger.NewLoggerWithService(cfg.Observability, loggerService)

	runtime, err := worker.New(cfg, &log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize worker")
	}

	if !runtime.Enabled() {
		log.Info().Msg("background worker disabled; exiting")
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runtime.Start(); err != nil {
		log.Fatal().Err(err).Msg("failed to start worker")
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout*time.Second)
	defer cancel()

	if err := runtime.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("worker forced to shutdown")
	}

	log.Info().Msg("worker exited properly")
}
