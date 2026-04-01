package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunReturnsSuccessWhenWorkerDisabled(t *testing.T) {
	clearJarvisEnv(t)
	t.Setenv("JARVIS_WORKER.ENABLED", "false")

	if exitCode := run(); exitCode != 0 {
		t.Fatalf("expected exit code 0 when worker is disabled, got %d", exitCode)
	}
}

func TestRunReturnsFailureWhenWorkerEnabledWithoutRedis(t *testing.T) {
	clearJarvisEnv(t)
	t.Setenv("JARVIS_WORKER.ENABLED", "true")

	if exitCode := run(); exitCode != 1 {
		t.Fatalf("expected exit code 1 when worker is enabled without redis config, got %d", exitCode)
	}
}

func clearJarvisEnv(t *testing.T) {
	t.Helper()

	for _, entry := range os.Environ() {
		key, _, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if strings.HasPrefix(key, "JARVIS_") {
			t.Setenv(key, "")
		}
	}
}
