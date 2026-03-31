package worker

import (
	"testing"

	"github.com/iancenry/jarvis/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkerRuntimeDisabled(t *testing.T) {
	runtime, err := New(&config.Config{
		Worker: &config.WorkerConfig{
			Enabled: false,
		},
	}, &zerolog.Logger{})

	require.NoError(t, err)
	require.NotNil(t, runtime)
	assert.False(t, runtime.Enabled())
	assert.Nil(t, runtime.jobService)
}

func TestNewWorkerRuntimeRequiresRedisWhenEnabled(t *testing.T) {
	runtime, err := New(&config.Config{
		Worker: &config.WorkerConfig{
			Enabled: true,
		},
		Auth: config.AuthConfig{
			SecretKey: "clerk-secret",
		},
		Integration: config.IntegrationConfig{
			ResendAPIKey: "resend-api-key",
		},
	}, &zerolog.Logger{})

	require.Error(t, err)
	assert.Nil(t, runtime)
	assert.Contains(t, err.Error(), "redis address is required")
}

func TestNewWorkerRuntimeConfiguresJobService(t *testing.T) {
	runtime, err := New(&config.Config{
		Worker: &config.WorkerConfig{
			Enabled: true,
		},
		Redis: config.RedisConfig{
			Address: "localhost:6379",
		},
		Auth: config.AuthConfig{
			SecretKey: "clerk-secret",
		},
		Integration: config.IntegrationConfig{
			ResendAPIKey: "resend-api-key",
		},
	}, &zerolog.Logger{})

	require.NoError(t, err)
	require.NotNil(t, runtime)
	assert.True(t, runtime.Enabled())
	assert.NotNil(t, runtime.jobService)
	assert.NotNil(t, runtime.jobService.Client)
}
