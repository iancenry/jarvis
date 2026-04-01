package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iancenry/jarvis/internal/config"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandlerHealthyProbeSkipsUnconfiguredRedisWithoutLoggingSuccess(t *testing.T) {
	handler, logOutput := newTestHealthHandler(t, func(h *HealthHandler) {
		h.dbCheck = func(ctx context.Context) error { return nil }
		h.redisConfigured = false
		h.redisCheck = nil
	})

	rec := performHealthRequest(t, handler, context.Background(), logOutput)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, logOutput.String())

	var response healthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, healthStatusHealthy, response.Status)
	assert.Equal(t, healthStatusHealthy, response.Checks.Database.Status)
	require.NotNil(t, response.Checks.Redis)
	assert.Equal(t, healthStatusSkipped, response.Checks.Redis.Status)
	assert.Equal(t, "redis not configured", response.Checks.Redis.Error)
}

func TestHealthHandlerRedisFailureReturnsServiceUnavailable(t *testing.T) {
	handler, logOutput := newTestHealthHandler(t, func(h *HealthHandler) {
		h.dbCheck = func(ctx context.Context) error { return nil }
		h.redisConfigured = true
		h.redisCheck = func(ctx context.Context) error { return errors.New("redis unavailable") }
	})

	rec := performHealthRequest(t, handler, context.Background(), logOutput)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, logOutput.String(), `"failed_checks":["redis"]`)

	var response healthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "unhealthy", response.Status)
	require.NotNil(t, response.Checks.Redis)
	assert.Equal(t, "unhealthy", response.Checks.Redis.Status)
	assert.Equal(t, "redis unavailable", response.Checks.Redis.Error)
}

func TestHealthHandlerUsesRequestContextForChecks(t *testing.T) {
	handler, logOutput := newTestHealthHandler(t, func(h *HealthHandler) {
		h.dbCheck = func(ctx context.Context) error { return ctx.Err() }
		h.redisConfigured = false
		h.redisCheck = nil
	})

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	rec := performHealthRequest(t, handler, canceledCtx, logOutput)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response healthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "unhealthy", response.Status)
	assert.Equal(t, "unhealthy", response.Checks.Database.Status)
	assert.Equal(t, context.Canceled.Error(), response.Checks.Database.Error)
}

func TestHealthHandlerDisabledChecksReturnHealthySkippedResults(t *testing.T) {
	handler, logOutput := newTestHealthHandler(t, func(h *HealthHandler) {
		h.enabled = false
		h.dbCheck = func(ctx context.Context) error {
			t.Fatal("database check should not run when health checks are disabled")
			return nil
		}
		h.redisConfigured = true
		h.redisCheck = func(ctx context.Context) error {
			t.Fatal("redis check should not run when health checks are disabled")
			return nil
		}
	})

	rec := performHealthRequest(t, handler, context.Background(), logOutput)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, logOutput.String())

	var response healthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, healthStatusHealthy, response.Status)
	assert.Equal(t, healthStatusSkipped, response.Checks.Database.Status)
	require.NotNil(t, response.Checks.Redis)
	assert.Equal(t, healthStatusSkipped, response.Checks.Redis.Status)
}

func newTestHealthHandler(t *testing.T, mutate func(h *HealthHandler)) (*HealthHandler, *bytes.Buffer) {
	t.Helper()

	logOutput := &bytes.Buffer{}
	logger := zerolog.New(logOutput)
	cfg := &config.Config{
		Primary: config.Primary{
			Env: "test",
		},
		Observability: &config.ObservabilityConfig{
			ServiceName: "jarvis-test",
			Environment: "test",
			Logging: config.LoggingConfig{
				Level:              "info",
				Format:             "json",
				SlowQueryThreshold: 100 * time.Millisecond,
			},
			NewRelic: config.NewRelicConfig{},
			HealthChecks: config.HealthChecksConfig{
				Enabled: true,
				Timeout: 25 * time.Millisecond,
				Checks:  []string{"database", "redis"},
			},
		},
	}

	h := NewHealthHandler(&server.Server{
		Config: cfg,
		Logger: &logger,
	})
	if mutate != nil {
		mutate(h)
	}

	return h, logOutput
}

func performHealthRequest(t *testing.T, handler *HealthHandler, reqCtx context.Context, logOutput *bytes.Buffer) *httptest.ResponseRecorder {
	t.Helper()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/status", nil).WithContext(reqCtx)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	logger := zerolog.New(logOutput)
	ctx.Set(middleware.LoggerKey, &logger)

	err := handler.CheckHealth(ctx)
	require.NoError(t, err)

	return rec
}
