package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/logger"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/redisclient"
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

	gormDatabase, err := database.Open(shutdownSignal, appConfig.Database, appLogger)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer database.Close(gormDatabase)

	redisClient, err := redisclient.New(shutdownSignal, appConfig.RedisURL)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer redisClient.Close()

	app, err := buildApplication(shutdownSignal, appConfig, appLogger, gormDatabase, redisClient)
	if err != nil {
		return fmt.Errorf("build application: %w", err)
	}

	var backgroundWorkers sync.WaitGroup
	backgroundWorkers.Go(func() { app.emailDispatcher.Run(shutdownSignal) })

	server := httpserver.New(appConfig.HTTPAddress(), app.router, appLogger)
	serverError := server.Run(shutdownSignal, appConfig.ShutdownTimeout)
	stopListening()
	backgroundWorkers.Wait()
	return serverError
}
