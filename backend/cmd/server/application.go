package main

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/activation"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/auth/login"
	"nebula-exchange/backend/internal/auth/loginlockout"
	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/auth/passwordreset"
	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/auth/signup"
	"nebula-exchange/backend/internal/health"
	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/platform/config"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/idempotency"
	"nebula-exchange/backend/internal/platform/ratelimit"
	"nebula-exchange/backend/internal/users"
)

const (
	rateLimitKeyPrefix    = "nebula:ratelimit:"
	loginLockoutKeyPrefix = "nebula:login-lockout:"
)

type application struct {
	router          *gin.Engine
	emailDispatcher *outbox.Dispatcher
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

	router := httpserver.NewRouter(httpserver.RouterOptions{
		Logger:           appLogger,
		IsProduction:     appConfig.IsProduction(),
		AllowedOrigins:   appConfig.AllowedOrigins,
		TrustedProxies:   appConfig.TrustedProxies,
		IdentifyUser:     authentication.IdentifyUser(accessTokens),
		IdempotencyStore: idempotency.NewPostgresStore(databasePool),
	},
		health.NewHandler(),
		signup.NewHandler(signupService, rateLimits.PerClientIP(ratelimit.SignupPolicy)),
		login.NewHandler(loginService, cookieSettings, time.Now, rateLimits.PerClientIP(ratelimit.LoginPolicy)),
		users.NewMeHandler(databasePool, userRepository),
		activation.NewHandler(activation.NewService(databasePool, userRepository, time.Now)),
		activation.NewResendHandler(
			activation.NewResender(databasePool, userRepository, activationIssuer, activationMailer, time.Now),
			rateLimits,
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

	return application{
		router:          router,
		emailDispatcher: outbox.NewDispatcher(databasePool, emailSender, appLogger, outbox.DispatcherOptions{}),
	}, nil
}
