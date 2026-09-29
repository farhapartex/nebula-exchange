package login

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/loginchallenge"
	"nebula-exchange/backend/internal/auth/loginlockout"
	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/auth/twofactor"
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

type LoginOutcome struct {
	Session            *EstablishedSession
	TwoFactorChallenge *loginchallenge.IssuedChallenge
}

type TwoFactorRequest struct {
	ChallengeToken string `json:"challenge_token" binding:"required,max=128"`
	Code           string `json:"code" binding:"required,len=6,numeric"`
}

type Dependencies struct {
	Pool                *pgxpool.Pool
	Users               *users.Repository
	AccessTokens        *accesstoken.Manager
	RefreshTokens       *session.RefreshTokens
	PasswordHasher      *passwordhash.Hasher
	PasswordHashOptions passwordhash.Parameters
	LoginLockout        *loginlockout.Guard
	TwoFactor           *twofactor.Service
	LoginChallenges     *loginchallenge.Store
	Logger              *slog.Logger
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
	errChallengeGone  = apierror.Unauthorized("Your login attempt expired. Enter your email and password again.")
)

func (service *Service) LogIn(ctx context.Context, request Request) (LoginOutcome, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(request.Email))
	if err := service.dependencies.LoginLockout.EnsureNotLocked(ctx, normalizedEmail); err != nil {
		return LoginOutcome{}, err
	}

	credentials, isFound, err := service.dependencies.Users.FindCredentialsByEmail(ctx, service.dependencies.Pool, normalizedEmail)
	if err != nil {
		return LoginOutcome{}, err
	}
	passwordHash := service.timingEqualizerHash
	if isFound {
		passwordHash = credentials.PasswordHash
	}
	isPasswordCorrect, err := service.dependencies.PasswordHasher.Verify(ctx, request.Password, passwordHash)
	if err != nil {
		return LoginOutcome{}, err
	}
	if !isFound || !isPasswordCorrect {
		if lockedError := service.dependencies.LoginLockout.RecordFailure(ctx, normalizedEmail); lockedError != nil {
			return LoginOutcome{}, lockedError
		}
		return LoginOutcome{}, errInvalidCredentials
	}
	if err := ensureAccountMayLogIn(credentials.User); err != nil {
		return LoginOutcome{}, err
	}

	if credentials.User.HasTwoFactorEnabled() {
		issuedChallenge, err := service.dependencies.LoginChallenges.Issue(ctx, credentials.User.ID, normalizedEmail)
		if err != nil {
			return LoginOutcome{}, err
		}
		return LoginOutcome{TwoFactorChallenge: &issuedChallenge}, nil
	}

	service.dependencies.LoginLockout.Reset(ctx, normalizedEmail)
	establishedSession, err := service.establishSession(ctx, credentials.User)
	if err != nil {
		return LoginOutcome{}, err
	}
	return LoginOutcome{Session: &establishedSession}, nil
}

func (service *Service) CompleteTwoFactor(ctx context.Context, request TwoFactorRequest) (EstablishedSession, error) {
	pendingChallenge, err := service.dependencies.LoginChallenges.Find(ctx, request.ChallengeToken)
	if errors.Is(err, loginchallenge.ErrChallengeInvalid) {
		return EstablishedSession{}, errChallengeGone
	}
	if err != nil {
		return EstablishedSession{}, err
	}
	if err := service.dependencies.LoginLockout.EnsureNotLocked(ctx, pendingChallenge.Email); err != nil {
		return EstablishedSession{}, err
	}

	if err := service.dependencies.TwoFactor.VerifyCode(ctx, pendingChallenge.UserID, request.Code); err != nil {
		var apiError *apierror.Error
		if !errors.As(err, &apiError) || apiError.Code != apierror.CodeValidationFailed {
			return EstablishedSession{}, err
		}
		if recordErr := service.dependencies.LoginChallenges.RecordFailedAttempt(ctx, request.ChallengeToken); recordErr != nil {
			return EstablishedSession{}, recordErr
		}
		if lockedError := service.dependencies.LoginLockout.RecordFailure(ctx, pendingChallenge.Email); lockedError != nil {
			return EstablishedSession{}, lockedError
		}
		return EstablishedSession{}, err
	}

	if err := service.dependencies.LoginChallenges.Consume(ctx, request.ChallengeToken); err != nil {
		return EstablishedSession{}, err
	}
	service.dependencies.LoginLockout.Reset(ctx, pendingChallenge.Email)

	challengedUser, isFound, err := service.dependencies.Users.FindByID(ctx, service.dependencies.Pool, pendingChallenge.UserID)
	if err != nil {
		return EstablishedSession{}, err
	}
	if !isFound {
		return EstablishedSession{}, errChallengeGone
	}
	if err := ensureAccountMayLogIn(challengedUser); err != nil {
		return EstablishedSession{}, err
	}
	return service.establishSession(ctx, challengedUser)
}

func ensureAccountMayLogIn(accountUser users.User) error {
	if !accountUser.IsActive {
		return errAccountNotActivated
	}
	if accountUser.Status == users.StatusBanned {
		return errAccountBanned
	}
	return nil
}

func (service *Service) establishSession(ctx context.Context, accountUser users.User) (EstablishedSession, error) {
	loggedInAt := service.dependencies.Now().UTC()
	loggedInUser := accountUser
	loggedInUser.LastLoginAt = &loggedInAt

	var refreshToken session.IssuedRefreshToken
	err := database.WithTransaction(ctx, service.dependencies.Pool, func(transaction pgx.Tx) error {
		if err := service.dependencies.Users.RecordLogin(ctx, transaction, loggedInUser.ID, loggedInAt); err != nil {
			return err
		}
		var err error
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
	var reusedError *session.RefreshTokenReusedError
	if errors.As(err, &reusedError) {
		service.revokeAllSessionsAfterReuse(ctx, reusedError.UserID)
		return EstablishedSession{}, errSessionExpired
	}
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

func (service *Service) revokeAllSessionsAfterReuse(ctx context.Context, userID uuid.UUID) {
	revokedCount, err := service.dependencies.RefreshTokens.RevokeAllForUser(context.WithoutCancel(ctx), service.dependencies.Pool, userID)
	if err != nil {
		service.dependencies.Logger.ErrorContext(ctx, "revoke sessions after refresh token reuse",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return
	}
	if revokedCount == 0 {
		service.dependencies.Logger.InfoContext(ctx, "revoked refresh token presented", slog.String("user_id", userID.String()))
		return
	}
	service.dependencies.Logger.WarnContext(ctx, "refresh token reuse detected, all sessions revoked",
		slog.String("user_id", userID.String()),
		slog.Int64("revoked_sessions", revokedCount),
	)
}

func (service *Service) LogOut(ctx context.Context, plaintextRefreshToken string) error {
	if plaintextRefreshToken == "" {
		return nil
	}
	return service.dependencies.RefreshTokens.Revoke(ctx, service.dependencies.Pool, plaintextRefreshToken)
}
