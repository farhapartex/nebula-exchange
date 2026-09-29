package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/activation"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/auth/login"
	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/auth/signup"
	"nebula-exchange/backend/internal/health"
	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/platform/config"
	"nebula-exchange/backend/internal/platform/database"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/idempotency"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/users"
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

	emailTemplates, err := email.NewTemplateRenderer()
	if err != nil {
		return err
	}
	emailSender := email.NewSMTPSender(email.SMTPConfig{
		Host:     appConfig.Email.SMTPHost,
		Port:     appConfig.Email.SMTPPort,
		Username: appConfig.Email.SMTPUsername,
		Password: appConfig.Email.SMTPPassword,
		From:     email.Address{Name: appConfig.Email.FromName, Email: appConfig.Email.FromAddress},
	})

	userRepository := users.NewRepository()
	accessTokens, err := accesstoken.NewManager(appConfig.Session.JWTSecret, accesstoken.DefaultLifetime, time.Now)
	if err != nil {
		return err
	}

	signupService := signup.NewService(signup.Dependencies{
		Pool:                databasePool,
		Users:               userRepository,
		ActivationIssuer:    activation.NewIssuer(activation.DefaultTokenLifetime, time.Now),
		ActivationEmail:     activation.NewEmailComposer(appConfig.FrontendBaseURL, emailTemplates),
		EmailSender:         emailSender,
		PasswordHashOptions: passwordhash.DefaultParameters,
		Logger:              appLogger,
		Now:                 time.Now,
	})

	loginService, err := login.NewService(login.Dependencies{
		Pool:                databasePool,
		Users:               userRepository,
		AccessTokens:        accessTokens,
		RefreshTokens:       session.NewRefreshTokens(session.DefaultRefreshTokenLifetime, time.Now),
		PasswordHashOptions: passwordhash.DefaultParameters,
		Now:                 time.Now,
	})
	if err != nil {
		return err
	}
	cookieSettings := session.CookieSettings{IsSecure: appConfig.Session.IsCookieSecure}

	router := httpserver.NewRouter(httpserver.RouterOptions{
		Logger:           appLogger,
		IsProduction:     appConfig.IsProduction(),
		AllowedOrigins:   appConfig.AllowedOrigins,
		IdentifyUser:     authentication.IdentifyUser(accessTokens),
		IdempotencyStore: idempotency.NewPostgresStore(databasePool),
	},
		health.NewHandler(),
		signup.NewHandler(signupService),
		login.NewHandler(loginService, cookieSettings, time.Now),
		users.NewMeHandler(databasePool, userRepository),
	)

	server := httpserver.New(appConfig.HTTPAddress(), router, appLogger)
	return server.Run(shutdownSignal, appConfig.ShutdownTimeout)
}
