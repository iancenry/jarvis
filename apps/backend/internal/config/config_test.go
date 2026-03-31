package config

import "testing"

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
