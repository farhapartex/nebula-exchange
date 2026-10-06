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
	"github.com/farhapartex/nebula-exchange/backend/internal/payment"
	paymentgateway "github.com/farhapartex/nebula-exchange/backend/internal/payment/gateway"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email/outbox"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/idempotency"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/objectstorage"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress"
	"github.com/farhapartex/nebula-exchange/backend/internal/story"
	"github.com/farhapartex/nebula-exchange/backend/internal/wallet"
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

	assetStorage, err := objectstorage.NewMinioStorage(appConfig.Storage)
	if err != nil {
		return nil, err
	}
	storyModule := story.NewModule(story.ModuleDependencies{Database: gormDatabase, ImageSigner: assetStorage})
	progressModule := progress.NewModule(progress.ModuleDependencies{
		Database: gormDatabase,
		Levels:   storyModule.LevelCatalog,
		Logger:   appLogger,
		Now:      time.Now,
	})

	paymentModule := payment.NewModule(buildPaymentDependencies(appConfig, appLogger, gormDatabase, storyModule, progressModule))

	walletModule, err := wallet.NewModule(wallet.ModuleDependencies{
		Database:        gormDatabase,
		ChainID:         appConfig.ChainID,
		FrontendBaseURL: appConfig.FrontendBaseURL,
		Now:             time.Now,
	})
	if err != nil {
		return nil, err
	}

	identityModule, err := identity.NewModule(ctx, identity.ModuleDependencies{
		Database:        gormDatabase,
		EmailEnqueuer:   outbox.NewEnqueuer(emailOutbox, time.Now),
		Session:         appConfig.Session,
		FrontendBaseURL: appConfig.FrontendBaseURL,
		PlayerProgress:  playerProgressAdapter{playerProgress: progressModule.PlayerProgress},
		Logger:          appLogger,
		Now:             time.Now,
	})
	if err != nil {
		return nil, err
	}

	routeRegistrars := []httpserver.RouteRegistrar{health.NewHandler(buildHealthService(appLogger, gormDatabase, redisClient))}
	routeRegistrars = append(routeRegistrars, identityModule.RouteRegistrars()...)
	routeRegistrars = append(routeRegistrars, storyModule.RouteRegistrars()...)
	routeRegistrars = append(routeRegistrars, progressModule.RouteRegistrars()...)
	routeRegistrars = append(routeRegistrars, paymentModule.RouteRegistrars()...)
	routeRegistrars = append(routeRegistrars, walletModule.RouteRegistrars()...)

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

func buildPaymentDependencies(appConfig config.Config, appLogger *slog.Logger, gormDatabase *gorm.DB, storyModule *story.Module, progressModule *progress.Module) payment.ModuleDependencies {
	paymentDependencies := payment.ModuleDependencies{
		Database:         gormDatabase,
		Chapters:         storyModule.LevelCatalog,
		ChapterOwnership: progressModule.ChapterOwnership,
		ChapterUnlocker:  progressModule.ChapterUnlocks,
		FrontendBaseURL:  appConfig.FrontendBaseURL,
		Logger:           appLogger,
		Now:              time.Now,
	}
	if !appConfig.Stripe.IsConfigured() {
		appLogger.Warn("stripe is not configured, chapter payments are off")
		return paymentDependencies
	}
	paymentDependencies.CheckoutGateway = paymentgateway.NewStripeCheckoutGateway(appConfig.Stripe.SecretKey)
	paymentDependencies.EventVerifier = paymentgateway.NewStripeEventVerifier(appConfig.Stripe.WebhookSecret)
	return paymentDependencies
}
