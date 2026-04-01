package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunReturnsFailureForInvalidAPIConfig(t *testing.T) {
	clearJarvisEnv(t)

	if exitCode := run(); exitCode != 1 {
		t.Fatalf("expected exit code 1 for invalid API config, got %d", exitCode)
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
