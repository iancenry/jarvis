package job

import (
	"context"
	"testing"

	"github.com/iancenry/jarvis/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authServiceStub struct{}

func (a authServiceStub) GetUserEmail(ctx context.Context, userID string) (string, error) {
	return "user@example.com", nil
}

func TestJobServiceStartRequiresAuthService(t *testing.T) {
	jobService := NewJobService(&zerolog.Logger{}, testJobConfig())

	err := jobService.Start()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth dependency not configured")
}

func TestJobServiceStartRequiresEmailClient(t *testing.T) {
	jobService := NewJobService(&zerolog.Logger{}, testJobConfig())
	jobService.SetAuthService(authServiceStub{})
	jobService.emailClient = nil

	err := jobService.Start()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email dependency not configured")
}

func testJobConfig() *config.Config {
	return &config.Config{
		Redis: config.RedisConfig{
			Address: "localhost:6379",
		},
		Integration: config.IntegrationConfig{
			ResendAPIKey: "test-api-key",
		},
	}
}
