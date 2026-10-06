package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/health"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email/outbox"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/idempotency"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/objectstorage"
	"github.com/farhapartex/nebula-exchange/backend/internal/story"
)

type application struct {
	router          *gin.Engine
	emailDispatcher *outbox.Dispatcher
}

func buildApplication(ctx context.Context, appConfig config.Config, appLogger *slog.Logger, gormDatabase *gorm.DB, redisClient *redis.Client) (*application, error) {
	emailOutbox := outbox.NewGormRepository(gormDatabase)
	smtpSender := email.NewSMTPSender(email.SMTPConfig{
		Host:     appConfig.Email.SMTPHost,
		Port:     appConfig.Email.SMTPPort,
		Username: appConfig.Email.SMTPUsername,
		Password: appConfig.Email.SMTPPassword,
		From:     email.Address{Name: appConfig.Email.FromName, Email: appConfig.Email.FromAddress},
	})

	identityModule, err := identity.NewModule(ctx, identity.ModuleDependencies{
		Database:        gormDatabase,
		EmailEnqueuer:   outbox.NewEnqueuer(emailOutbox, time.Now),
		Session:         appConfig.Session,
		FrontendBaseURL: appConfig.FrontendBaseURL,
		Logger:          appLogger,
		Now:             time.Now,
	})
	if err != nil {
		return nil, err
	}

	assetStorage, err := objectstorage.NewMinioStorage(appConfig.Storage)
	if err != nil {
		return nil, err
	}
	storyModule := story.NewModule(story.ModuleDependencies{Database: gormDatabase, ImageSigner: assetStorage})

	routeRegistrars := []httpserver.RouteRegistrar{health.NewHandler(buildHealthService(appLogger, gormDatabase, redisClient))}
	routeRegistrars = append(routeRegistrars, identityModule.RouteRegistrars()...)
	routeRegistrars = append(routeRegistrars, storyModule.RouteRegistrars()...)

	router, err := httpserver.NewRouter(
		httpserver.RouterOptions{
			Logger:         appLogger,
			IsProduction:   appConfig.IsProduction(),
			AllowedOrigins: appConfig.AllowedOrigins,
			TrustedProxies: appConfig.TrustedProxies,
			AccessTokens:   identityModule.AccessTokens,
			Idempotency:    idempotency.NewGormStore(gormDatabase),
			NonReplayable:  identityModule.NonReplayableRoutes(),
		},
		routeRegistrars...,
	)
	if err != nil {
		return nil, err
	}

	return &application{
		router:          router,
		emailDispatcher: outbox.NewDispatcher(emailOutbox, smtpSender, appLogger, outbox.DispatcherOptions{}),
	}, nil
}

func buildHealthService(appLogger *slog.Logger, gormDatabase *gorm.DB, redisClient *redis.Client) health.Service {
	return health.NewService(appLogger,
		health.NewProbe("database", func(ctx context.Context) error {
			sqlDatabase, err := gormDatabase.DB()
			if err != nil {
				return err
			}
			return sqlDatabase.PingContext(ctx)
		}),
		health.NewProbe("redis", func(ctx context.Context) error {
			return redisClient.Ping(ctx).Err()
		}),
	)
}
