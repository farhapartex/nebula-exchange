package identity

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/emails"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/handler"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/security"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
)

type ModuleDependencies struct {
	Database        *gorm.DB
	EmailEnqueuer   service.EmailEnqueuer
	Session         config.SessionConfig
	FrontendBaseURL string
	Logger          *slog.Logger
	Now             func() time.Time
}

type Module struct {
	AccessTokens *security.AccessTokens
	registrars   []httpserver.RouteRegistrar
}

func NewModule(ctx context.Context, dependencies ModuleDependencies) (*Module, error) {
	accessTokens, err := security.NewAccessTokens(dependencies.Session.JWTSecret, security.DefaultAccessTokenLifetime, dependencies.Now)
	if err != nil {
		return nil, err
	}
	activationEmailComposer, err := emails.NewActivationEmailComposer(dependencies.FrontendBaseURL)
	if err != nil {
		return nil, err
	}

	users := repository.NewUserRepository(dependencies.Database)
	activationTokens := repository.NewActivationTokenRepository(dependencies.Database)
	refreshSessions := repository.NewRefreshSessionRepository(dependencies.Database)
	passwordHasher := security.NewPasswordHasher(security.PasswordHasherOptions{Parameters: security.DefaultArgon2idParameters})
	transactions := database.NewTransactionRunner(dependencies.Database)
	clock := service.Clock(dependencies.Now)

	signupService := service.NewSignupService(service.SignupDependencies{
		Users:            users,
		ActivationTokens: activationTokens,
		PasswordHasher:   passwordHasher,
		EmailComposer:    activationEmailComposer,
		EmailEnqueuer:    dependencies.EmailEnqueuer,
		Transactions:     transactions,
		Now:              clock,
	})
	activationService := service.NewActivationService(service.ActivationDependencies{
		Users:            users,
		ActivationTokens: activationTokens,
		Transactions:     transactions,
		Now:              clock,
	})
	sessionService, err := service.NewSessionService(ctx, service.SessionDependencies{
		Users:           users,
		RefreshSessions: refreshSessions,
		PasswordHasher:  passwordHasher,
		AccessTokens:    accessTokens,
		Transactions:    transactions,
		Logger:          dependencies.Logger,
		Now:             clock,
	})
	if err != nil {
		return nil, err
	}

	refreshCookie := handler.RefreshCookie{IsSecure: dependencies.Session.IsCookieSecure, Now: dependencies.Now}
	return &Module{
		AccessTokens: accessTokens,
		registrars: []httpserver.RouteRegistrar{
			handler.NewSignupHandler(signupService),
			handler.NewActivationHandler(activationService),
			handler.NewSessionHandler(sessionService, refreshCookie),
			handler.NewProfileHandler(service.NewProfileService(users)),
		},
	}, nil
}

func (module *Module) RouteRegistrars() []httpserver.RouteRegistrar {
	return module.registrars
}
