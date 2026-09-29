package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestLoadUsesDefaultsWhenEnvironmentIsEmpty(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")

	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedConfig := Config{
		Environment:     EnvironmentDevelopment,
		HTTPPort:        8080,
		LogLevel:        slog.LevelInfo,
		ShutdownTimeout: 15 * time.Second,
	}
	if loadedConfig != expectedConfig {
		t.Fatalf("got %+v, want %+v", loadedConfig, expectedConfig)
	}
}

func TestLoadReadsValuesFromEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_PORT", "9000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("SHUTDOWN_TIMEOUT", "5s")

	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !loadedConfig.IsProduction() {
		t.Fatalf("expected production environment, got %q", loadedConfig.Environment)
	}
	if loadedConfig.HTTPAddress() != ":9000" {
		t.Fatalf("got address %q, want :9000", loadedConfig.HTTPAddress())
	}
	if loadedConfig.LogLevel != slog.LevelDebug {
		t.Fatalf("got log level %v, want debug", loadedConfig.LogLevel)
	}
	if loadedConfig.ShutdownTimeout != 5*time.Second {
		t.Fatalf("got shutdown timeout %v, want 5s", loadedConfig.ShutdownTimeout)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	invalidEnvironments := map[string]map[string]string{
		"unknown environment": {"APP_ENV": "staging"},
		"non numeric port":    {"HTTP_PORT": "eighty"},
		"unknown log level":   {"LOG_LEVEL": "verbose"},
		"malformed timeout":   {"SHUTDOWN_TIMEOUT": "fifteen"},
	}

	for caseName, environmentValues := range invalidEnvironments {
		t.Run(caseName, func(t *testing.T) {
			for key, value := range environmentValues {
				t.Setenv(key, value)
			}
			if _, err := Load(); err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}
