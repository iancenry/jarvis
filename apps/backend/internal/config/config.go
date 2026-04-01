package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Primary       Primary              `koanf:"primary" validate:"required"`
	Server        ServerConfig         `koanf:"server" validate:"required"`
	Database      DatabaseConfig       `koanf:"database" validate:"required"`
	Auth          AuthConfig           `koanf:"auth" validate:"required"`
	Redis         RedisConfig          `koanf:"redis"`
	Worker        *WorkerConfig        `koanf:"worker"`
	Integration   IntegrationConfig    `koanf:"integration" validate:"required"`
	Observability *ObservabilityConfig `koanf:"observability"`
	AWS           AWSConfig            `koanf:"aws" validate:"required"`
	Cron          *CronConfig          `koanf:"cron"`
}

type Primary struct {
	Env string `koanf:"env" validate:"required"`
}

type ServerConfig struct {
	Port               string   `koanf:"port" validate:"required"`
	ReadTimeout        int      `koanf:"read_timeout" validate:"required"`
	WriteTimeout       int      `koanf:"write_timeout" validate:"required"`
	IdleTimeout        int      `koanf:"idle_timeout" validate:"required"`
	CORSAllowedOrigins []string `koanf:"cors_allowed_origins" validate:"required"`
}

type DatabaseConfig struct {
	Host            string `koanf:"host" validate:"required"`
	Port            int    `koanf:"port" validate:"required"`
	User            string `koanf:"user" validate:"required"`
	Password        string `koanf:"password"`
	Name            string `koanf:"name" validate:"required"`
	SSLMode         string `koanf:"ssl_mode" validate:"required"`
	MaxOpenConns    int    `koanf:"max_open_conns" validate:"required"`
	MaxIdleConns    int    `koanf:"max_idle_conns" validate:"required"`
	ConnMaxLifetime int    `koanf:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime int    `koanf:"conn_max_idle_time" validate:"required"`
}
type RedisConfig struct {
	Address  string `koanf:"address"`
	Password string `koanf:"password"`
}

// WorkerConfig defines the configuration for the background worker
type WorkerConfig struct {
	Enabled bool `koanf:"enabled"`
}

type IntegrationConfig struct {
	Enabled      *bool  `koanf:"enabled"`
	ResendAPIKey string `koanf:"resend_api_key" validate:"required"`
}

type AuthConfig struct {
	SecretKey string `koanf:"secret_key" validate:"required"`
}

type AWSConfig struct {
	Enabled         *bool  `koanf:"enabled"`
	Region          string `koanf:"region" validate:"required"`
	AccessKeyID     string `koanf:"access_key_id" validate:"required"`
	SecretAccessKey string `koanf:"secret_access_key" validate:"required"`
	UploadBucket    string `koanf:"upload_bucket" validate:"required"`
	EndpointURL     string `koanf:"endpoint_url"`
}

type CronConfig struct {
	ArchiveDaysThreshold        int `koanf:"archive_days_threshold"`
	BatchSize                   int `koanf:"batch_size"`
	ReminderHours               int `koanf:"reminder_hours"`
	MaxTodosPerUserNotification int `koanf:"max_todos_per_user_notification"`
}

func DefaultCronConfig() *CronConfig {
	return &CronConfig{
		ArchiveDaysThreshold:        30,
		BatchSize:                   100,
		ReminderHours:               24,
		MaxTodosPerUserNotification: 10,
	}
}

func DefaultWorkerConfig() *WorkerConfig {
	return &WorkerConfig{
		Enabled: true,
	}
}

func (c *Config) WorkerEnabled() bool {
	if c == nil || c.Worker == nil {
		return true
	}

	return c.Worker.Enabled
}

func (c *Config) EmailEnabled() bool {
	if c == nil {
		return false
	}

	if c.Integration.Enabled != nil {
		return *c.Integration.Enabled
	}

	return c.Integration.ResendAPIKey != ""
}

func (c *Config) S3Enabled() bool {
	if c == nil {
		return false
	}

	if c.AWS.Enabled != nil {
		return *c.AWS.Enabled
	}

	return c.AWS.Region != "" ||
		c.AWS.AccessKeyID != "" ||
		c.AWS.SecretAccessKey != "" ||
		c.AWS.UploadBucket != "" ||
		c.AWS.EndpointURL != ""
}

func (c *Config) ValidateForAPI() error {
	if c == nil {
		return fmt.Errorf("config is required")
	}

	c.applyDefaults()

	if err := validateConfigSection("primary", c.Primary); err != nil {
		return err
	}
	if err := validateConfigSection("server", c.Server); err != nil {
		return err
	}
	if err := validateConfigSection("database", c.Database); err != nil {
		return err
	}
	if err := validateConfigSection("auth", c.Auth); err != nil {
		return err
	}
	if err := c.validateObservability(); err != nil {
		return err
	}
	if c.S3Enabled() {
		if err := validateConfigSection("aws", c.AWS); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) ValidateForWorker() error {
	if c == nil {
		return fmt.Errorf("config is required")
	}

	c.applyDefaults()

	if err := c.validateObservability(); err != nil {
		return err
	}
	if !c.WorkerEnabled() {
		return nil
	}
	if c.Redis.Address == "" {
		return fmt.Errorf("redis address is required when the background worker is enabled")
	}
	if err := validateConfigSection("auth", c.Auth); err != nil {
		return err
	}
	if !c.EmailEnabled() {
		return fmt.Errorf("email delivery must be enabled when the background worker is enabled")
	}
	if err := validateConfigSection("integration", c.Integration); err != nil {
		return err
	}

	return nil
}

func (c *Config) ValidateForCron() error {
	if c == nil {
		return fmt.Errorf("config is required")
	}

	c.applyDefaults()

	if err := c.validateObservability(); err != nil {
		return err
	}
	if err := validateConfigSection("database", c.Database); err != nil {
		return err
	}
	if c.Redis.Address == "" {
		return fmt.Errorf("redis address is required for cron jobs")
	}

	return nil
}

func LoadConfig() (*Config, error) {
	k := koanf.New(".")

	err := k.Load(env.Provider("JARVIS_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "JARVIS_"))
	}), nil)
	if err != nil {
		return nil, fmt.Errorf("could not load env variables: %w", err)
	}

	mainConfig := &Config{}

	err = k.Unmarshal("", mainConfig)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config: %w", err)
	}

	mainConfig.applyDefaults()
	if err := mainConfig.validateObservability(); err != nil {
		return nil, err
	}

	return mainConfig, nil
}

func (c *Config) applyDefaults() {
	if c == nil {
		return
	}

	observabilityDefaults := DefaultObservabilityConfig()
	if c.Observability == nil {
		c.Observability = observabilityDefaults
	} else {
		if c.Observability.ServiceName == "" {
			c.Observability.ServiceName = observabilityDefaults.ServiceName
		}
		if c.Observability.Environment == "" {
			c.Observability.Environment = observabilityDefaults.Environment
		}
		if c.Observability.Logging.Level == "" {
			c.Observability.Logging.Level = observabilityDefaults.Logging.Level
		}
		if c.Observability.Logging.Format == "" {
			c.Observability.Logging.Format = observabilityDefaults.Logging.Format
		}
		if c.Observability.Logging.SlowQueryThreshold == 0 {
			c.Observability.Logging.SlowQueryThreshold = observabilityDefaults.Logging.SlowQueryThreshold
		}
		if c.Observability.HealthChecks.Interval == 0 {
			c.Observability.HealthChecks.Interval = observabilityDefaults.HealthChecks.Interval
		}
		if c.Observability.HealthChecks.Timeout == 0 {
			c.Observability.HealthChecks.Timeout = observabilityDefaults.HealthChecks.Timeout
		}
		if len(c.Observability.HealthChecks.Checks) == 0 {
			c.Observability.HealthChecks.Checks = observabilityDefaults.HealthChecks.Checks
		}
	}

	if c.Primary.Env != "" {
		c.Observability.Environment = c.Primary.Env
	}
	c.Observability.ServiceName = "jarvis"

	if c.Worker == nil {
		c.Worker = DefaultWorkerConfig()
	}

	cronDefaults := DefaultCronConfig()
	if c.Cron == nil {
		c.Cron = cronDefaults
	} else {
		if c.Cron.ArchiveDaysThreshold == 0 {
			c.Cron.ArchiveDaysThreshold = cronDefaults.ArchiveDaysThreshold
		}
		if c.Cron.BatchSize == 0 {
			c.Cron.BatchSize = cronDefaults.BatchSize
		}
		if c.Cron.ReminderHours == 0 {
			c.Cron.ReminderHours = cronDefaults.ReminderHours
		}
		if c.Cron.MaxTodosPerUserNotification == 0 {
			c.Cron.MaxTodosPerUserNotification = cronDefaults.MaxTodosPerUserNotification
		}
	}
}

func (c *Config) validateObservability() error {
	if c == nil {
		return fmt.Errorf("config is required")
	}
	if c.Observability == nil {
		return fmt.Errorf("invalid observability config: observability settings are required")
	}
	if err := c.Observability.Validate(); err != nil {
		return fmt.Errorf("invalid observability config: %w", err)
	}

	return nil
}

func validateConfigSection(section string, value any) error {
	validate := validator.New()
	if err := validate.Struct(value); err != nil {
		return fmt.Errorf("invalid %s config: %w", section, err)
	}

	return nil
}
