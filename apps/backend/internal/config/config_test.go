package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigWorkerEnabledDefaultsToTrue(t *testing.T) {
	cfg := &Config{}

	if !cfg.WorkerEnabled() {
		t.Fatalf("expected worker to default to enabled when config.worker is unset")
	}
}

func TestConfigWorkerEnabledHonorsExplicitFalse(t *testing.T) {
	cfg := &Config{
		Worker: &WorkerConfig{
			Enabled: false,
		},
	}

	if cfg.WorkerEnabled() {
		t.Fatalf("expected worker to be disabled when config.worker.enabled=false")
	}
}

func TestConfigEmailEnabledDefaultsToFalseWithoutAPIKey(t *testing.T) {
	cfg := &Config{}

	if cfg.EmailEnabled() {
		t.Fatalf("expected email delivery to default to disabled when no API key or explicit flag is set")
	}
}

func TestConfigEmailEnabledUsesAPIKeyByDefault(t *testing.T) {
	cfg := &Config{
		Integration: IntegrationConfig{
			ResendAPIKey: "resend-api-key",
		},
	}

	if !cfg.EmailEnabled() {
		t.Fatalf("expected email delivery to auto-enable when a Resend API key is configured")
	}
}

func TestConfigS3EnabledDefaultsToFalseWithoutAWSConfig(t *testing.T) {
	cfg := &Config{}

	if cfg.S3Enabled() {
		t.Fatalf("expected S3 to default to disabled when no AWS config or explicit flag is set")
	}
}

func TestConfigValidateForAPIDoesNotRequireOptionalEmailOrS3(t *testing.T) {
	cfg := &Config{
		Primary: Primary{
			Env: "local",
		},
		Server: ServerConfig{
			Port:               "8080",
			ReadTimeout:        30,
			WriteTimeout:       30,
			IdleTimeout:        30,
			CORSAllowedOrigins: []string{"*"},
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Name:            "jarvis",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    10,
			ConnMaxLifetime: 300,
			ConnMaxIdleTime: 300,
		},
		Auth: AuthConfig{
			SecretKey: "clerk-secret",
		},
	}

	require.NoError(t, cfg.ValidateForAPI())
}

func TestConfigValidateForWorkerRequiresEmailWhenEnabled(t *testing.T) {
	cfg := &Config{
		Worker: &WorkerConfig{
			Enabled: true,
		},
		Redis: RedisConfig{
			Address: "localhost:6379",
		},
		Auth: AuthConfig{
			SecretKey: "clerk-secret",
		},
	}

	err := cfg.ValidateForWorker()
	require.Error(t, err)
	require.Contains(t, err.Error(), "email delivery must be enabled")
}

func TestConfigValidateForCronDoesNotRequireAuthOrS3(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Name:            "jarvis",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    10,
			ConnMaxLifetime: 300,
			ConnMaxIdleTime: 300,
		},
		Redis: RedisConfig{
			Address: "localhost:6379",
		},
	}

	require.NoError(t, cfg.ValidateForCron())
}
