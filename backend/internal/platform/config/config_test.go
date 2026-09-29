package config

import (
	"log/slog"
	"reflect"
	"testing"
	"time"
)

const testDatabaseURL = "postgres://nebula:nebula@localhost:5432/nebula_exchange?sslmode=disable"

func clearOptionalEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{"APP_ENV", "HTTP_PORT", "LOG_LEVEL", "SHUTDOWN_TIMEOUT", "DATABASE_MAX_CONNECTIONS", "FRONTEND_ORIGINS", "FRONTEND_BASE_URL", "SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD", "EMAIL_FROM_ADDRESS", "EMAIL_FROM_NAME"} {
		t.Setenv(key, "")
	}
	t.Setenv("DATABASE_URL", testDatabaseURL)
}

func TestLoadUsesDefaultsWhenEnvironmentIsEmpty(t *testing.T) {
	clearOptionalEnvironment(t)

	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedConfig := Config{
		Environment:     EnvironmentDevelopment,
		HTTPPort:        8080,
		LogLevel:        slog.LevelInfo,
		ShutdownTimeout: 15 * time.Second,
		AllowedOrigins:  []string{"http://localhost:3000"},
		FrontendBaseURL: "http://localhost:3000",
		Database:        DatabaseConfig{URL: testDatabaseURL, MaxConnections: 20},
		Email: EmailConfig{
			SMTPHost:    "localhost",
			SMTPPort:    1025,
			FromAddress: "no-reply@nebula.test",
			FromName:    "Nebula Exchange",
		},
	}
	if !reflect.DeepEqual(loadedConfig, expectedConfig) {
		t.Fatalf("got %+v, want %+v", loadedConfig, expectedConfig)
	}
}

func TestLoadReadsValuesFromEnvironment(t *testing.T) {
	clearOptionalEnvironment(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_PORT", "9000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("SHUTDOWN_TIMEOUT", "5s")
	t.Setenv("DATABASE_MAX_CONNECTIONS", "7")
	t.Setenv("FRONTEND_ORIGINS", "https://play.nebula.test, https://admin.nebula.test")

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
	if !reflect.DeepEqual(loadedConfig.AllowedOrigins, []string{"https://play.nebula.test", "https://admin.nebula.test"}) {
		t.Fatalf("got origins %v", loadedConfig.AllowedOrigins)
	}
	if loadedConfig.Database.MaxConnections != 7 {
		t.Fatalf("got max connections %d, want 7", loadedConfig.Database.MaxConnections)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	invalidEnvironments := map[string]map[string]string{
		"unknown environment":      {"APP_ENV": "staging"},
		"non numeric port":         {"HTTP_PORT": "eighty"},
		"unknown log level":        {"LOG_LEVEL": "verbose"},
		"malformed timeout":        {"SHUTDOWN_TIMEOUT": "fifteen"},
		"missing database url":     {"DATABASE_URL": ""},
		"zero database connection": {"DATABASE_MAX_CONNECTIONS": "0"},
		"non numeric smtp port":    {"SMTP_PORT": "mail"},
	}

	for caseName, environmentValues := range invalidEnvironments {
		t.Run(caseName, func(t *testing.T) {
			clearOptionalEnvironment(t)
			for key, value := range environmentValues {
				t.Setenv(key, value)
			}
			if _, err := Load(); err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}
