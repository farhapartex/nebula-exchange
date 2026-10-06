package config

import (
	"log/slog"
	"testing"
	"time"
)

const testJWTSecret = "a-test-secret-that-is-long-enough-123"

func TestLoadUsesDefaultsWhenOnlyRequiredValuesAreSet(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://street:street@localhost:5432/street")
	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("STORAGE_ACCESS_KEY", "street-born")
	t.Setenv("STORAGE_SECRET_KEY", "street-born-secret")

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
	if loadedConfig.Session.IsCookieSecure || loadedConfig.Email.SMTPPort != 1025 {
		t.Fatalf("unexpected session or email defaults: %+v %+v", loadedConfig.Session, loadedConfig.Email)
	}
	if loadedConfig.Storage.PublicEndpoint != "localhost:9000" || loadedConfig.Storage.PresignLifetime != time.Hour {
		t.Fatalf("unexpected storage defaults: %+v", loadedConfig.Storage)
	}
	if loadedConfig.ShutdownTimeout != 15*time.Second {
		t.Fatalf("got shutdown timeout %s, want 15s", loadedConfig.ShutdownTimeout)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	invalidSettings := map[string]map[string]string{
		"missing database url":      {"DATABASE_URL": ""},
		"unknown environment":       {"APP_ENV": "staging"},
		"non numeric port":          {"HTTP_PORT": "eighty"},
		"zero connections":          {"DATABASE_MAX_CONNECTIONS": "0"},
		"invalid trusted proxy":     {"TRUSTED_PROXIES": "not-an-ip"},
		"invalid shutdown value":    {"SHUTDOWN_TIMEOUT": "soon"},
		"short jwt secret":          {"JWT_SECRET": "too-short"},
		"invalid cookie flag":       {"COOKIE_SECURE": "maybe"},
		"missing storage key":       {"STORAGE_SECRET_KEY": ""},
		"presign lifetime too long": {"STORAGE_PRESIGN_LIFETIME": "720h"},
	}
	for caseName, settings := range invalidSettings {
		t.Run(caseName, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://street:street@localhost:5432/street")
			t.Setenv("JWT_SECRET", testJWTSecret)
			t.Setenv("STORAGE_ACCESS_KEY", "street-born")
			t.Setenv("STORAGE_SECRET_KEY", "street-born-secret")
			for key, value := range settings {
				t.Setenv(key, value)
			}
			if _, err := Load(); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestCookiesAreSecureOutsideDevelopment(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://street:street@localhost:5432/street")
	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("STORAGE_ACCESS_KEY", "street-born")
	t.Setenv("STORAGE_SECRET_KEY", "street-born-secret")
	t.Setenv("APP_ENV", EnvironmentProduction)

	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !loadedConfig.Session.IsCookieSecure {
		t.Fatal("expected secure cookies in production")
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
