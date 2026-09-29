package config

import (
	"fmt"
	"log/slog"
	"time"
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
	EnvironmentTest        = "test"
)

type Config struct {
	Environment     string
	HTTPPort        int
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	environment := readString("APP_ENV", EnvironmentDevelopment)
	if !isSupportedEnvironment(environment) {
		return Config{}, fmt.Errorf("APP_ENV %q is not one of development, production, test", environment)
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

	return Config{
		Environment:     environment,
		HTTPPort:        httpPort,
		LogLevel:        logLevel,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func (cfg Config) IsProduction() bool {
	return cfg.Environment == EnvironmentProduction
}

func (cfg Config) HTTPAddress() string {
	return fmt.Sprintf(":%d", cfg.HTTPPort)
}

func isSupportedEnvironment(environment string) bool {
	switch environment {
	case EnvironmentDevelopment, EnvironmentProduction, EnvironmentTest:
		return true
	default:
		return false
	}
}
