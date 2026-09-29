package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"nebula-exchange/backend/internal/health"
	"nebula-exchange/backend/internal/platform/config"
	"nebula-exchange/backend/internal/platform/database"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	appConfig, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	appLogger := logger.New(appConfig.LogLevel, appConfig.IsProduction())

	shutdownSignal, stopListening := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopListening()

	databasePool, err := database.NewPool(shutdownSignal, appConfig.Database)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer databasePool.Close()

	router := httpserver.NewRouter(appLogger, appConfig.IsProduction(),
		health.NewHandler(),
	)

	server := httpserver.New(appConfig.HTTPAddress(), router, appLogger)
	return server.Run(shutdownSignal, appConfig.ShutdownTimeout)
}
