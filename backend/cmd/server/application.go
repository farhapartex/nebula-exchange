package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"nebula-exchange/backend/internal/account"
	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/activation"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/auth/login"
	"nebula-exchange/backend/internal/auth/loginchallenge"
	"nebula-exchange/backend/internal/auth/loginlockout"
	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/auth/passwordreset"
	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/auth/signup"
	"nebula-exchange/backend/internal/auth/twofactor"
	"nebula-exchange/backend/internal/balances"
	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/health"
	"nebula-exchange/backend/internal/inventory"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledgerhistory"
	"nebula-exchange/backend/internal/maintenance"
	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/payments"
	"nebula-exchange/backend/internal/platform/config"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/idempotency"
	"nebula-exchange/backend/internal/platform/ratelimit"
	"nebula-exchange/backend/internal/platform/scheduler"
	"nebula-exchange/backend/internal/platform/secretbox"
	"nebula-exchange/backend/internal/users"
)

const (
	rateLimitKeyPrefix      = "nebula:ratelimit:"
	loginLockoutKeyPrefix   = "nebula:login-lockout:"
	loginChallengeKeyPrefix = "nebula:login-challenge:"
	catalogCacheLifetime    = time.Minute
)

type application struct {
	router          *gin.Engine
	emailDispatcher *outbox.Dispatcher
	jobScheduler    *scheduler.Scheduler
}

func buildApplication(appConfig config.Config, appLogger *slog.Logger, databasePool *pgxpool.Pool, redisClient *redis.Client) (application, error) {
	emailTemplates, err := email.NewTemplateRenderer()
	if err != nil {
		return application{}, err
	}
	emailSender := email.NewSMTPSender(email.SMTPConfig{
		Host:     appConfig.Email.SMTPHost,
		Port:     appConfig.Email.SMTPPort,
		Username: appConfig.Email.SMTPUsername,
		Password: appConfig.Email.SMTPPassword,
		From:     email.Address{Name: appConfig.Email.FromName, Email: appConfig.Email.FromAddress},
	})

	accessTokens, err := accesstoken.NewManager(appConfig.Session.JWTSecret, accesstoken.DefaultLifetime, time.Now)
	if err != nil {
		return application{}, err
	}
	userRepository := users.NewRepository()
	passwordHasher := passwordhash.NewHasher(passwordhash.HasherOptions{Parameters: passwordhash.DefaultParameters})

	activationIssuer := activation.NewIssuer(activation.DefaultTokenLifetime, time.Now)
	activationMailer := activation.NewMailer(activation.NewEmailComposer(appConfig.FrontendBaseURL, emailTemplates), outbox.NewQueue())

	signupService := signup.NewService(signup.Dependencies{
		Pool:             databasePool,
		Users:            userRepository,
		ActivationIssuer: activationIssuer,
		ActivationMailer: activationMailer,
		PasswordHasher:   passwordHasher,
		Now:              time.Now,
	})

	totpSecretBox, err := secretbox.NewFromBase64Key(appConfig.Session.TOTPEncryptionKey)
	if err != nil {
		return application{}, fmt.Errorf("TOTP_ENCRYPTION_KEY: %w", err)
	}
	twoFactorService := twofactor.NewService(twofactor.Dependencies{
		Pool:           databasePool,
		SecretBox:      totpSecretBox,
		EmailTemplates: emailTemplates,
		EmailQueue:     outbox.NewQueue(),
		Now:            time.Now,
	})

	refreshTokens := session.NewRefreshTokens(session.DefaultRefreshTokenLifetime, time.Now)
	loginLockout := loginlockout.NewGuard(redisClient, loginLockoutKeyPrefix, loginlockout.DefaultPolicy, appLogger)

	loginService, err := login.NewService(login.Dependencies{
		Pool:                databasePool,
		Users:               userRepository,
		AccessTokens:        accessTokens,
		RefreshTokens:       refreshTokens,
		PasswordHasher:      passwordHasher,
		PasswordHashOptions: passwordhash.DefaultParameters,
		LoginLockout:        loginLockout,
		TwoFactor:           twoFactorService,
		LoginChallenges:     loginchallenge.NewStore(redisClient, loginChallengeKeyPrefix, loginchallenge.DefaultLifetime, time.Now),
		Logger:              appLogger,
		Now:                 time.Now,
	})
	if err != nil {
		return application{}, err
	}

	rateLimits := ratelimit.NewMiddlewareFactory(
		ratelimit.NewSlidingWindowLimiter(redisClient, rateLimitKeyPrefix, time.Now),
		appLogger,
	)
	cookieSettings := session.CookieSettings{IsSecure: appConfig.Session.IsCookieSecure}
	activeSessions := session.NewSessions(databasePool, time.Now)
	accountGuard := users.NewAccountGuard(databasePool, userRepository)

	catalogService := catalog.NewService(catalog.NewLoader(databasePool), catalogCacheLifetime, time.Now)
	paymentService := payments.NewService(payments.Dependencies{
		Pool:     databasePool,
		Users:    userRepository,
		Catalog:  catalogService,
		Checkout: cardCheckoutFor(appConfig),
		Now:      time.Now,
	})

	router := httpserver.NewRouter(httpserver.RouterOptions{
		Logger:           appLogger,
		IsProduction:     appConfig.IsProduction(),
		AllowedOrigins:   appConfig.AllowedOrigins,
		TrustedProxies:   appConfig.TrustedProxies,
		IdentifyUser:     authentication.IdentifyUser(accessTokens),
		IdempotencyStore: idempotency.NewPostgresStore(databasePool),
	},
		health.NewHandler(),
		catalog.NewHandler(catalogService),
		signup.NewHandler(signupService, rateLimits.PerClientIP(ratelimit.SignupPolicy)),
		login.NewHandler(loginService, cookieSettings, time.Now, rateLimits.PerClientIP(ratelimit.LoginPolicy)),
		users.NewMeHandler(databasePool, userRepository),
		balances.NewHandler(balances.NewReader(databasePool)),
		inventory.NewHandler(inventory.NewReader(databasePool)),
		ledgerhistory.NewHandler(ledgerhistory.NewReader(databasePool)),
		payments.NewHandler(
			paymentService,
			accountGuard.RequireStatus(users.StatusActive, users.StatusPendingPayment),
			rateLimits.PerSubject(ratelimit.PaymentCreationPolicy, authenticatedSubject),
		),
		activation.NewHandler(activation.NewService(databasePool, userRepository, time.Now)),
		activation.NewResendHandler(
			activation.NewResender(databasePool, userRepository, activationIssuer, activationMailer, time.Now),
			rateLimits,
		),
		twofactor.NewHandler(twoFactorService, rateLimits.PerClientIP(ratelimit.TwoFactorChangePolicy)),
		session.NewSessionsHandler(activeSessions, cookieSettings),
		account.NewHandler(account.NewService(account.Dependencies{
			Pool:           databasePool,
			Users:          userRepository,
			PasswordHasher: passwordHasher,
			Sessions:       activeSessions,
			Now:            time.Now,
		}),
			rateLimits.PerClientIP(ratelimit.PasswordChangePolicy),
			accountGuard.RequireStatus(users.StatusActive, users.StatusPendingPayment),
		),
		passwordreset.NewHandler(passwordreset.NewService(passwordreset.Dependencies{
			Pool:            databasePool,
			Users:           userRepository,
			PasswordHasher:  passwordHasher,
			RefreshTokens:   refreshTokens,
			LoginLockout:    loginLockout,
			EmailTemplates:  emailTemplates,
			EmailQueue:      outbox.NewQueue(),
			FrontendBaseURL: appConfig.FrontendBaseURL,
			Now:             time.Now,
		}), rateLimits),
	)

	jobScheduler := scheduler.New(databasePool, appLogger, scheduler.Options{InitialDelay: time.Minute})
	jobScheduler.Register(maintenance.NewCleanupJob(databasePool, appLogger, time.Now))
	jobScheduler.Register(ledger.NewCheckJob(databasePool, appLogger))
	jobScheduler.Register(payments.NewExpiryJob(databasePool, appLogger, time.Now))

	return application{
		router:          router,
		emailDispatcher: outbox.NewDispatcher(databasePool, emailSender, appLogger, outbox.DispatcherOptions{}),
		jobScheduler:    jobScheduler,
	}, nil
}

func cardCheckoutFor(appConfig config.Config) payments.CardCheckout {
	if appConfig.Environment == config.EnvironmentDevelopment {
		return payments.NewDevelopmentCheckout(appConfig.FrontendBaseURL)
	}
	return payments.UnavailableCheckout{}
}

func authenticatedSubject(context *gin.Context) string {
	userID, _ := authentication.UserIDFrom(context)
	return userID.String()
}
