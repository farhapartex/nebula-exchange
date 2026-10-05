package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
	EnvironmentTest        = "test"
)

type DatabaseConfig struct {
	URL                   string
	MaxOpenConnections    int
	MaxIdleConnections    int
	ConnectionMaxLifetime time.Duration
}

type Config struct {
	Environment     string
	HTTPPort        int
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
	TrustedProxies  []string
	RedisURL        string
	Database        DatabaseConfig
}

func Load() (Config, error) {
	environment, err := loadEnvironment()
	if err != nil {
		return Config{}, err
	}

	httpPort, err := readInt("HTTP_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	logLevel, err := readLogLevel("LOG_LEVEL", slog.LevelInfo)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := readDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}

	trustedProxies := readList("TRUSTED_PROXIES", nil)
	if err := validateTrustedProxies(trustedProxies); err != nil {
		return Config{}, err
	}

	databaseConfig, err := LoadDatabaseConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment:     environment,
		HTTPPort:        httpPort,
		LogLevel:        logLevel,
		ShutdownTimeout: shutdownTimeout,
		AllowedOrigins:  readList("FRONTEND_ORIGINS", []string{"http://localhost:3000"}),
		TrustedProxies:  trustedProxies,
		RedisURL:        readString("REDIS_URL", "redis://localhost:6379/0"),
		Database:        databaseConfig,
	}, nil
}

func LoadDatabaseConfig() (DatabaseConfig, error) {
	databaseURL := readString("DATABASE_URL", "")
	if databaseURL == "" {
		return DatabaseConfig{}, errors.New("DATABASE_URL is required")
	}

	maxOpenConnections, err := readInt("DATABASE_MAX_CONNECTIONS", 20)
	if err != nil {
		return DatabaseConfig{}, err
	}
	if maxOpenConnections < 1 {
		return DatabaseConfig{}, errors.New("DATABASE_MAX_CONNECTIONS must be at least 1")
	}

	connectionMaxLifetime, err := readDuration("DATABASE_CONNECTION_MAX_LIFETIME", 30*time.Minute)
	if err != nil {
		return DatabaseConfig{}, err
	}

	return DatabaseConfig{
		URL:                   databaseURL,
		MaxOpenConnections:    maxOpenConnections,
		MaxIdleConnections:    max(maxOpenConnections/4, 1),
		ConnectionMaxLifetime: connectionMaxLifetime,
	}, nil
}

func (cfg Config) IsProduction() bool {
	return cfg.Environment == EnvironmentProduction
}

func (cfg Config) HTTPAddress() string {
	return fmt.Sprintf(":%d", cfg.HTTPPort)
}

func loadEnvironment() (string, error) {
	environment := readString("APP_ENV", EnvironmentDevelopment)
	switch environment {
	case EnvironmentDevelopment, EnvironmentProduction, EnvironmentTest:
		return environment, nil
	default:
		return "", fmt.Errorf("APP_ENV %q is not one of development, production, test", environment)
	}
}

func validateTrustedProxies(trustedProxies []string) error {
	for _, trustedProxy := range trustedProxies {
		if net.ParseIP(trustedProxy) != nil {
			continue
		}
		if _, _, err := net.ParseCIDR(trustedProxy); err != nil {
			return fmt.Errorf("TRUSTED_PROXIES entry %q is not an IP address or CIDR range", trustedProxy)
		}
	}
	return nil
}
