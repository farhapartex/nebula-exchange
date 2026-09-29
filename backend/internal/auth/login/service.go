package login

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/database"
	"nebula-exchange/backend/internal/users"
)

type Request struct {
	Email    string `json:"email" binding:"required,max=254"`
	Password string `json:"password" binding:"required,max=128"`
}

type EstablishedSession struct {
	AccessToken  accesstoken.IssuedToken
	RefreshToken session.IssuedRefreshToken
	User         users.User
}

type Dependencies struct {
	Pool                *pgxpool.Pool
	Users               *users.Repository
	AccessTokens        *accesstoken.Manager
	RefreshTokens       *session.RefreshTokens
	PasswordHashOptions passwordhash.Parameters
	Now                 func() time.Time
}

type Service struct {
	dependencies        Dependencies
	timingEqualizerHash string
}

func NewService(dependencies Dependencies) (*Service, error) {
	timingEqualizerHash, err := passwordhash.Hash("timing-equalizer-password", dependencies.PasswordHashOptions)
	if err != nil {
		return nil, err
	}
	return &Service{dependencies: dependencies, timingEqualizerHash: timingEqualizerHash}, nil
}

var (
	errInvalidCredentials  = apierror.Unauthorized("Invalid email or password")
	errAccountNotActivated = apierror.New(http.StatusForbidden, apierror.CodeAccountNotActivated,
		"Your account is not activated yet. Check your email for the activation link.")
	errAccountBanned  = apierror.Forbidden("This account is banned")
	errSessionExpired = apierror.Unauthorized("Your session has expired. Log in again.")
)

func (service *Service) LogIn(ctx context.Context, request Request) (EstablishedSession, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(request.Email))
	credentials, isFound, err := service.dependencies.Users.FindCredentialsByEmail(ctx, service.dependencies.Pool, normalizedEmail)
	if err != nil {
		return EstablishedSession{}, err
	}
	if !isFound {
		_, _ = passwordhash.Verify(request.Password, service.timingEqualizerHash)
		return EstablishedSession{}, errInvalidCredentials
	}

	isPasswordCorrect, err := passwordhash.Verify(request.Password, credentials.PasswordHash)
	if err != nil {
		return EstablishedSession{}, err
	}
	if !isPasswordCorrect {
		return EstablishedSession{}, errInvalidCredentials
	}
	if !credentials.User.IsActive {
		return EstablishedSession{}, errAccountNotActivated
	}
	if credentials.User.Status == users.StatusBanned {
		return EstablishedSession{}, errAccountBanned
	}

	loggedInAt := service.dependencies.Now().UTC()
	loggedInUser := credentials.User
	loggedInUser.LastLoginAt = &loggedInAt

	var refreshToken session.IssuedRefreshToken
	err = database.WithTransaction(ctx, service.dependencies.Pool, func(transaction pgx.Tx) error {
		if err := service.dependencies.Users.RecordLogin(ctx, transaction, loggedInUser.ID, loggedInAt); err != nil {
			return err
		}
		refreshToken, err = service.dependencies.RefreshTokens.Issue(ctx, transaction, loggedInUser.ID)
		return err
	})
	if err != nil {
		return EstablishedSession{}, err
	}

	accessToken, err := service.dependencies.AccessTokens.Issue(loggedInUser.ID)
	if err != nil {
		return EstablishedSession{}, err
	}
	return EstablishedSession{AccessToken: accessToken, RefreshToken: refreshToken, User: loggedInUser}, nil
}

func (service *Service) Refresh(ctx context.Context, plaintextRefreshToken string) (EstablishedSession, error) {
	if plaintextRefreshToken == "" {
		return EstablishedSession{}, errSessionExpired
	}

	var rotatedToken session.RotatedRefreshToken
	var sessionUser users.User
	err := database.WithTransaction(ctx, service.dependencies.Pool, func(transaction pgx.Tx) error {
		var err error
		rotatedToken, err = service.dependencies.RefreshTokens.Rotate(ctx, transaction, plaintextRefreshToken)
		if err != nil {
			return err
		}
		var isFound bool
		sessionUser, isFound, err = service.dependencies.Users.FindByID(ctx, transaction, rotatedToken.UserID)
		if err != nil {
			return err
		}
		if !isFound || !sessionUser.IsActive || sessionUser.Status == users.StatusBanned {
			return session.ErrRefreshTokenInvalid
		}
		return nil
	})
	if errors.Is(err, session.ErrRefreshTokenInvalid) {
		return EstablishedSession{}, errSessionExpired
	}
	if err != nil {
		return EstablishedSession{}, err
	}

	accessToken, err := service.dependencies.AccessTokens.Issue(sessionUser.ID)
	if err != nil {
		return EstablishedSession{}, err
	}
	return EstablishedSession{AccessToken: accessToken, RefreshToken: rotatedToken.Issued, User: sessionUser}, nil
}
