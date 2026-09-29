package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"nebula-exchange/backend/internal/platform/config"
	"nebula-exchange/backend/internal/platform/database"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/platform/redisclient"
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

	redisClient, err := redisclient.New(shutdownSignal, appConfig.RedisURL)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer redisClient.Close()

	application, err := buildApplication(appConfig, appLogger, databasePool, redisClient)
	if err != nil {
		return err
	}

	var backgroundWorkers sync.WaitGroup
	backgroundWorkers.Add(2)
	go func() {
		defer backgroundWorkers.Done()
		application.emailDispatcher.Run(shutdownSignal)
	}()
	go func() {
		defer backgroundWorkers.Done()
		application.jobScheduler.Run(shutdownSignal)
	}()

	server := httpserver.New(appConfig.HTTPAddress(), application.router, appLogger)
	serverError := server.Run(shutdownSignal, appConfig.ShutdownTimeout)
	stopListening()
	backgroundWorkers.Wait()
	return serverError
}
