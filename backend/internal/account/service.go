package account

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/database"
	"nebula-exchange/backend/internal/users"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,20}$`)

var errCurrentPasswordIncorrect = apierror.ValidationFailed(map[string]string{"current_password": "is incorrect"})

type ProfileUpdate struct {
	Username string `json:"username" binding:"required"`
}

type PasswordChange struct {
	CurrentPassword string `json:"current_password" binding:"required,max=128"`
	NewPassword     string `json:"new_password" binding:"required,min=10,max=128"`
}

type PasswordChangeResult struct {
	PasswordChanged bool  `json:"password_changed"`
	SignedOutOthers int64 `json:"signed_out_other_sessions"`
}

type Dependencies struct {
	Pool           *pgxpool.Pool
	Users          *users.Repository
	PasswordHasher *passwordhash.Hasher
	Sessions       *session.Sessions
	Now            func() time.Time
}

type Service struct {
	dependencies Dependencies
}

func NewService(dependencies Dependencies) *Service {
	return &Service{dependencies: dependencies}
}

func (service *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, profileUpdate ProfileUpdate) (users.Profile, error) {
	requestedUsername := strings.TrimSpace(profileUpdate.Username)
	if !usernamePattern.MatchString(requestedUsername) {
		return users.Profile{}, apierror.ValidationFailed(map[string]string{"username": "must be 3 to 20 letters, numbers or underscores"})
	}

	err := service.dependencies.Users.UpdateUsername(ctx, service.dependencies.Pool, userID, requestedUsername, service.dependencies.Now().UTC())
	if errors.Is(err, users.ErrUsernameTaken) {
		return users.Profile{}, apierror.ValidationFailed(map[string]string{"username": "is already taken"})
	}
	if err != nil {
		return users.Profile{}, err
	}

	updatedUser, _, err := service.dependencies.Users.FindByID(ctx, service.dependencies.Pool, userID)
	if err != nil {
		return users.Profile{}, err
	}
	return updatedUser.Profile(), nil
}

func (service *Service) ChangePassword(ctx context.Context, userID uuid.UUID, passwordChange PasswordChange, currentTokenHash []byte) (PasswordChangeResult, error) {
	storedPasswordHash, err := service.dependencies.Users.FindPasswordHash(ctx, service.dependencies.Pool, userID)
	if err != nil {
		return PasswordChangeResult{}, err
	}
	isCurrentPasswordCorrect, err := service.dependencies.PasswordHasher.Verify(ctx, passwordChange.CurrentPassword, storedPasswordHash)
	if err != nil {
		return PasswordChangeResult{}, err
	}
	if !isCurrentPasswordCorrect {
		return PasswordChangeResult{}, errCurrentPasswordIncorrect
	}
	if passwordChange.CurrentPassword == passwordChange.NewPassword {
		return PasswordChangeResult{}, apierror.ValidationFailed(map[string]string{"new_password": "must be different from your current password"})
	}

	newPasswordHash, err := service.dependencies.PasswordHasher.Hash(ctx, passwordChange.NewPassword)
	if err != nil {
		return PasswordChangeResult{}, err
	}

	var signedOutCount int64
	err = database.WithTransaction(ctx, service.dependencies.Pool, func(transaction pgx.Tx) error {
		if _, err := service.dependencies.Users.UpdatePasswordHash(ctx, transaction, userID, newPasswordHash, service.dependencies.Now().UTC()); err != nil {
			return err
		}
		signedOutCount, err = service.dependencies.Sessions.RevokeOthers(ctx, transaction, userID, currentTokenHash)
		return err
	})
	if err != nil {
		return PasswordChangeResult{}, err
	}
	return PasswordChangeResult{PasswordChanged: true, SignedOutOthers: signedOutCount}, nil
}
