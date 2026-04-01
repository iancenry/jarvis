package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iancenry/jarvis/internal/config"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/server"

	"github.com/labstack/echo/v4"
)

const (
	healthCheckDatabase = "database"
	healthCheckRedis    = "redis"
	healthStatusHealthy = "healthy"
	healthStatusSkipped = "skipped"
)

type healthCheckFunc func(ctx context.Context) error

type healthCheckResult struct {
	Status       string `json:"status"`
	ResponseTime string `json:"response_time"`
	Error        string `json:"error,omitempty"`
}

type healthChecksResponse struct {
	Database healthCheckResult  `json:"database"`
	Redis    *healthCheckResult `json:"redis,omitempty"`
}

type healthResponse struct {
	Status      string               `json:"status"`
	Timestamp   time.Time            `json:"timestamp"`
	Environment string               `json:"environment"`
	Checks      healthChecksResponse `json:"checks"`
}

type HealthHandler struct {
	Handler
	environment string
	timeout     time.Duration
	enabled     bool
	checks      map[string]struct{}

	dbCheck         healthCheckFunc
	redisCheck      healthCheckFunc
	redisConfigured bool
}

func NewHealthHandler(s *server.Server) *HealthHandler {
	defaultChecks := config.DefaultObservabilityConfig().HealthChecks

	handler := &HealthHandler{
		Handler: NewHandler(s),
		timeout: defaultChecks.Timeout,
		enabled: defaultChecks.Enabled,
		checks:  make(map[string]struct{}, len(defaultChecks.Checks)),
	}

	for _, check := range defaultChecks.Checks {
		handler.checks[strings.ToLower(check)] = struct{}{}
	}

	if s == nil {
		return handler
	}

	if s.Config != nil {
		handler.environment = s.Config.Primary.Env
		if s.Config.Observability != nil {
			healthCfg := s.Config.Observability.HealthChecks
			handler.enabled = healthCfg.Enabled
			if healthCfg.Timeout > 0 {
				handler.timeout = healthCfg.Timeout
			}
			handler.checks = make(map[string]struct{}, len(healthCfg.Checks))
			for _, check := range healthCfg.Checks {
				handler.checks[strings.ToLower(strings.TrimSpace(check))] = struct{}{}
			}
		}
	}

	if s.DB != nil && s.DB.Pool != nil {
		handler.dbCheck = s.DB.Pool.Ping
	}

	if s.Redis != nil {
		handler.redisConfigured = true
		handler.redisCheck = func(ctx context.Context) error {
			return s.Redis.Ping(ctx).Err()
		}
	}

	return handler
}

func (h *HealthHandler) CheckHealth(c echo.Context) error {
	start := time.Now()
	logger := middleware.GetLogger(c).With().
		Str("operation", "health_check").
		Logger()

	response := healthResponse{
		Status:      healthStatusHealthy,
		Timestamp:   time.Now().UTC(),
		Environment: h.environment,
		Checks: healthChecksResponse{
			Database: h.evaluateDatabaseCheck(c.Request().Context()),
		},
	}

	if h.shouldExposeRedisCheck() {
		redisCheck := h.evaluateRedisCheck(c.Request().Context())
		response.Checks.Redis = &redisCheck
	}

	failedChecks := h.failedChecks(response.Checks)
	if len(failedChecks) > 0 {
		response.Status = "unhealthy"
		totalDuration := time.Since(start)
		logger.Warn().
			Strs("failed_checks", failedChecks).
			Dur("total_duration", totalDuration).
			Msg("health check failed")
		h.recordHealthFailure(failedChecks, totalDuration)
		return c.JSON(http.StatusServiceUnavailable, response)
	}

	if err := c.JSON(http.StatusOK, response); err != nil {
		logger.Error().Err(err).Msg("failed to write health check response")
		h.recordHealthFailure([]string{"response"}, time.Since(start))
		return fmt.Errorf("failed to write JSON response: %w", err)
	}

	return nil
}

func (h *HealthHandler) evaluateDatabaseCheck(ctx context.Context) healthCheckResult {
	if !h.isCheckEnabled(healthCheckDatabase) {
		return skippedHealthCheckResult("database health check disabled")
	}

	if h.dbCheck == nil {
		return failedHealthCheckResult("database not configured")
	}

	return h.runCheck(ctx, h.dbCheck)
}

func (h *HealthHandler) evaluateRedisCheck(ctx context.Context) healthCheckResult {
	if !h.isCheckEnabled(healthCheckRedis) {
		return skippedHealthCheckResult("redis health check disabled")
	}

	if !h.redisConfigured || h.redisCheck == nil {
		return skippedHealthCheckResult("redis not configured")
	}

	return h.runCheck(ctx, h.redisCheck)
}

func (h *HealthHandler) runCheck(parent context.Context, check healthCheckFunc) healthCheckResult {
	checkCtx, cancel := context.WithTimeout(parent, h.timeout)
	defer cancel()

	start := time.Now()
	if err := check(checkCtx); err != nil {
		return healthCheckResult{
			Status:       "unhealthy",
			ResponseTime: time.Since(start).String(),
			Error:        err.Error(),
		}
	}

	return healthCheckResult{
		Status:       healthStatusHealthy,
		ResponseTime: time.Since(start).String(),
	}
}

func (h *HealthHandler) isCheckEnabled(name string) bool {
	if !h.enabled {
		return false
	}

	_, ok := h.checks[name]
	return ok
}

func (h *HealthHandler) shouldExposeRedisCheck() bool {
	if h.redisConfigured {
		return true
	}

	_, listed := h.checks[healthCheckRedis]
	return listed
}

func (h *HealthHandler) failedChecks(checks healthChecksResponse) []string {
	failed := make([]string, 0, 2)
	if checks.Database.Status == "unhealthy" {
		failed = append(failed, healthCheckDatabase)
	}
	if checks.Redis != nil && checks.Redis.Status == "unhealthy" {
		failed = append(failed, healthCheckRedis)
	}

	return failed
}

func (h *HealthHandler) recordHealthFailure(failedChecks []string, duration time.Duration) {
	if h.server == nil || h.server.LoggerService == nil || h.server.LoggerService.GetApplication() == nil {
		return
	}

	h.server.LoggerService.GetApplication().RecordCustomEvent(
		"HealthCheckError", map[string]interface{}{
			"operation":         "health_check",
			"failed_checks":     strings.Join(failedChecks, ","),
			"total_duration_ms": duration.Milliseconds(),
		},
	)
}

func skippedHealthCheckResult(reason string) healthCheckResult {
	return healthCheckResult{
		Status:       healthStatusSkipped,
		ResponseTime: "0s",
		Error:        reason,
	}
}

func failedHealthCheckResult(reason string) healthCheckResult {
	return healthCheckResult{
		Status:       "unhealthy",
		ResponseTime: "0s",
		Error:        reason,
	}
}
