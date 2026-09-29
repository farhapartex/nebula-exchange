package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

func readString(key, fallback string) string {
	value, isSet := os.LookupEnv(key)
	if !isSet || strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func readInt(key string, fallback int) (int, error) {
	rawValue := readString(key, "")
	if rawValue == "" {
		return fallback, nil
	}
	parsedValue, err := strconv.Atoi(rawValue)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return parsedValue, nil
}

func readDuration(key string, fallback time.Duration) (time.Duration, error) {
	rawValue := readString(key, "")
	if rawValue == "" {
		return fallback, nil
	}
	parsedValue, err := time.ParseDuration(rawValue)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration such as 15s: %w", key, err)
	}
	return parsedValue, nil
}

func readLogLevel(key string, fallback slog.Level) (slog.Level, error) {
	rawValue := readString(key, "")
	if rawValue == "" {
		return fallback, nil
	}
	var parsedLevel slog.Level
	if err := parsedLevel.UnmarshalText([]byte(rawValue)); err != nil {
		return fallback, fmt.Errorf("%s must be one of debug, info, warn, error: %w", key, err)
	}
	return parsedLevel, nil
}
