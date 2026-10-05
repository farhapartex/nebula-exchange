package service

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/security"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/securetoken"
)

const (
	RefreshTokenLifetime   = 30 * 24 * time.Hour
	concurrentRefreshGrace = 30 * time.Second
	maximumUserAgentLength = 256
	timingEqualizerSecret  = "timing-equalizer-password"
)

type LoginInput struct {
	Email    string
	Password string
}

type ClientMetadata struct {
	UserAgent string
	IPAddress string
}

type IssuedRefreshToken struct {
	Plaintext string
	ExpiresAt time.Time
}

type EstablishedSession struct {
	AccessToken  security.IssuedAccessToken
	RefreshToken IssuedRefreshToken
	User         models.User
}

type SessionService interface {
	LogIn(ctx context.Context, input LoginInput, metadata ClientMetadata) (EstablishedSession, error)
	Refresh(ctx context.Context, plaintextRefreshToken string, metadata ClientMetadata) (EstablishedSession, error)
	LogOut(ctx context.Context, plaintextRefreshToken string) error
}

type SessionDependencies struct {
	Users           repository.UserRepository
	RefreshSessions repository.RefreshSessionRepository
	PasswordHasher  PasswordHasher
	AccessTokens    AccessTokenIssuer
	Transactions    TransactionRunner
	Logger          *slog.Logger
	Now             Clock
}

type sessionService struct {
	dependencies        SessionDependencies
	timingEqualizerHash string
}

func NewSessionService(ctx context.Context, dependencies SessionDependencies) (SessionService, error) {
	timingEqualizerHash, err := dependencies.PasswordHasher.Hash(ctx, timingEqualizerSecret)
	if err != nil {
		return nil, err
	}
	return &sessionService{dependencies: dependencies, timingEqualizerHash: timingEqualizerHash}, nil
}

func (sessions *sessionService) LogIn(ctx context.Context, input LoginInput, metadata ClientMetadata) (EstablishedSession, error) {
	accountUser, isFound, err := sessions.dependencies.Users.FindByEmail(ctx, normalizeEmail(input.Email))
	if err != nil {
		return EstablishedSession{}, err
	}

	passwordHash := sessions.timingEqualizerHash
	if isFound {
		passwordHash = accountUser.PasswordHash
	}
	isPasswordCorrect, err := sessions.dependencies.PasswordHasher.Verify(ctx, input.Password, passwordHash)
	if err != nil {
		return EstablishedSession{}, err
	}
	if !isFound || !isPasswordCorrect {
		return EstablishedSession{}, ErrInvalidCredentials
	}
	if err := ensureAccountMayLogIn(accountUser); err != nil {
		return EstablishedSession{}, err
	}

	loggedInAt := sessions.dependencies.Now().UTC()
	var refreshToken IssuedRefreshToken
	err = sessions.dependencies.Transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := sessions.dependencies.Users.RecordLogin(ctx, accountUser.ID, loggedInAt); err != nil {
			return err
		}
		newSessionID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		refreshToken, err = sessions.issueRefreshToken(ctx, newSessionID, accountUser.ID, newSessionID, metadata)
		return err
	})
	if err != nil {
		return EstablishedSession{}, err
	}
	accountUser.LastLoginAt = &loggedInAt
	return sessions.withAccessToken(accountUser, refreshToken)
}

func (sessions *sessionService) Refresh(ctx context.Context, plaintextRefreshToken string, metadata ClientMetadata) (EstablishedSession, error) {
	if !securetoken.HasValidShape(plaintextRefreshToken) {
		return EstablishedSession{}, ErrSessionExpired
	}
	tokenHash := securetoken.Hash(plaintextRefreshToken)
	now := sessions.dependencies.Now().UTC()

	var sessionUser models.User
	var replacementToken IssuedRefreshToken
	err := sessions.dependencies.Transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		consumedSession, isConsumed, err := sessions.dependencies.RefreshSessions.Consume(ctx, tokenHash, now)
		if err != nil {
			return err
		}
		if !isConsumed {
			return ErrSessionExpired
		}
		var isFound bool
		sessionUser, isFound, err = sessions.dependencies.Users.FindByID(ctx, consumedSession.UserID)
		if err != nil {
			return err
		}
		if !isFound || !sessionUser.IsActive || sessionUser.Status == models.UserStatusBanned {
			return ErrSessionExpired
		}
		replacementID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		replacementToken, err = sessions.issueRefreshToken(ctx, replacementID, sessionUser.ID, consumedSession.FamilyID, metadata)
		if err != nil {
			return err
		}
		return sessions.dependencies.RefreshSessions.LinkReplacement(ctx, consumedSession.ID, replacementID)
	})
	if errors.Is(err, ErrSessionExpired) {
		sessions.revokeFamilyIfTokenWasReused(ctx, tokenHash, now)
		return EstablishedSession{}, ErrSessionExpired
	}
	if err != nil {
		return EstablishedSession{}, err
	}
	return sessions.withAccessToken(sessionUser, replacementToken)
}

func (sessions *sessionService) LogOut(ctx context.Context, plaintextRefreshToken string) error {
	if !securetoken.HasValidShape(plaintextRefreshToken) {
		return nil
	}
	return sessions.dependencies.RefreshSessions.RevokeByHash(ctx, securetoken.Hash(plaintextRefreshToken), sessions.dependencies.Now().UTC())
}

func (sessions *sessionService) issueRefreshToken(ctx context.Context, sessionID, userID, familyID uuid.UUID, metadata ClientMetadata) (IssuedRefreshToken, error) {
	generatedToken, err := securetoken.Generate()
	if err != nil {
		return IssuedRefreshToken{}, err
	}
	issuedAt := sessions.dependencies.Now().UTC()
	expiresAt := issuedAt.Add(RefreshTokenLifetime)
	err = sessions.dependencies.RefreshSessions.Create(ctx, &models.RefreshSession{
		ID:        sessionID,
		UserID:    userID,
		FamilyID:  familyID,
		TokenHash: generatedToken.Hash,
		UserAgent: truncateUserAgent(metadata.UserAgent),
		IPAddress: validIPAddress(metadata.IPAddress),
		ExpiresAt: expiresAt,
		CreatedAt: issuedAt,
	})
	if err != nil {
		return IssuedRefreshToken{}, err
	}
	return IssuedRefreshToken{Plaintext: generatedToken.Plaintext, ExpiresAt: expiresAt}, nil
}

func (sessions *sessionService) withAccessToken(sessionUser models.User, refreshToken IssuedRefreshToken) (EstablishedSession, error) {
	accessToken, err := sessions.dependencies.AccessTokens.Issue(sessionUser.ID)
	if err != nil {
		return EstablishedSession{}, err
	}
	return EstablishedSession{AccessToken: accessToken, RefreshToken: refreshToken, User: sessionUser}, nil
}

func (sessions *sessionService) revokeFamilyIfTokenWasReused(ctx context.Context, tokenHash []byte, now time.Time) {
	persistContext := context.WithoutCancel(ctx)
	storedSession, isFound, err := sessions.dependencies.RefreshSessions.FindByHash(persistContext, tokenHash)
	if err != nil || !isFound || storedSession.RevokedAt == nil {
		return
	}
	isConcurrentRefresh := storedSession.ReplacedByID != nil && now.Sub(*storedSession.RevokedAt) <= concurrentRefreshGrace
	if isConcurrentRefresh {
		return
	}
	revokedCount, err := sessions.dependencies.RefreshSessions.RevokeFamily(persistContext, storedSession.FamilyID, now)
	if err != nil {
		sessions.dependencies.Logger.ErrorContext(ctx, "revoke session family after token reuse", slog.String("family_id", storedSession.FamilyID.String()), slog.Any("error", err))
		return
	}
	if revokedCount > 0 {
		sessions.dependencies.Logger.WarnContext(ctx, "refresh token reuse detected, session family revoked",
			slog.String("user_id", storedSession.UserID.String()),
			slog.Int64("revoked_sessions", revokedCount),
		)
	}
}

func ensureAccountMayLogIn(accountUser models.User) error {
	if !accountUser.IsActive {
		return ErrAccountNotActivated
	}
	if accountUser.Status == models.UserStatusBanned {
		return ErrAccountBanned
	}
	return nil
}

func truncateUserAgent(userAgent string) string {
	if len(userAgent) > maximumUserAgentLength {
		return userAgent[:maximumUserAgentLength]
	}
	return userAgent
}

func validIPAddress(ipAddress string) *string {
	if net.ParseIP(ipAddress) == nil {
		return nil
	}
	return &ipAddress
}
