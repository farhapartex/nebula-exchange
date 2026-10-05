package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestLoadUsesDefaultsWhenOnlyTheDatabaseIsSet(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://street:street@localhost:5432/street")

	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if loadedConfig.Environment != EnvironmentDevelopment || loadedConfig.HTTPPort != 8080 || loadedConfig.LogLevel != slog.LevelInfo {
		t.Fatalf("unexpected defaults: %+v", loadedConfig)
	}
	if loadedConfig.Database.MaxOpenConnections != 20 || loadedConfig.Database.MaxIdleConnections != 5 {
		t.Fatalf("unexpected pool sizes: %+v", loadedConfig.Database)
	}
	if loadedConfig.ShutdownTimeout != 15*time.Second {
		t.Fatalf("got shutdown timeout %s, want 15s", loadedConfig.ShutdownTimeout)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	invalidSettings := map[string]map[string]string{
		"missing database url":   {"DATABASE_URL": ""},
		"unknown environment":    {"APP_ENV": "staging"},
		"non numeric port":       {"HTTP_PORT": "eighty"},
		"zero connections":       {"DATABASE_MAX_CONNECTIONS": "0"},
		"invalid trusted proxy":  {"TRUSTED_PROXIES": "not-an-ip"},
		"invalid shutdown value": {"SHUTDOWN_TIMEOUT": "soon"},
	}
	for caseName, settings := range invalidSettings {
		t.Run(caseName, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://street:street@localhost:5432/street")
			for key, value := range settings {
				t.Setenv(key, value)
			}
			if _, err := Load(); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestLoadMigrationConfigBuildsAFileSource(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://street:street@localhost:5432/street")
	t.Setenv("MIGRATIONS_PATH", "/app/migrations")

	migrationConfig, err := LoadMigrationConfig()
	if err != nil {
		t.Fatalf("load migration config: %v", err)
	}
	if migrationConfig.MigrationsSource != "file:///app/migrations" {
		t.Fatalf("got source %q", migrationConfig.MigrationsSource)
	}
}
