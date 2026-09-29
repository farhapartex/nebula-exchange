package activation

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/activation/activationstore"
	"nebula-exchange/backend/internal/auth/securetoken"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/database"
	"nebula-exchange/backend/internal/users"
)

var errInvalidActivationLink = apierror.NotFound("This activation link is invalid or has expired")

type Preview struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	IsActivated bool   `json:"is_activated"`
}

type ActivatedAccount struct {
	Email       string       `json:"email"`
	Username    string       `json:"username"`
	Status      users.Status `json:"status"`
	ActivatedAt time.Time    `json:"activated_at"`
}

type Service struct {
	pool  *pgxpool.Pool
	users *users.Repository
	now   func() time.Time
}

func NewService(pool *pgxpool.Pool, userRepository *users.Repository, now func() time.Time) *Service {
	return &Service{pool: pool, users: userRepository, now: now}
}

func (service *Service) Preview(ctx context.Context, plaintextToken string) (Preview, error) {
	if !securetoken.HasValidShape(plaintextToken) {
		return Preview{}, errInvalidActivationLink
	}
	storedToken, err := activationstore.New(service.pool).FindActivationTokenWithUser(ctx, securetoken.Hash(plaintextToken))
	if errors.Is(err, pgx.ErrNoRows) {
		return Preview{}, errInvalidActivationLink
	}
	if err != nil {
		return Preview{}, err
	}

	if storedToken.IsActive {
		return Preview{Email: storedToken.Email, Username: storedToken.Username, IsActivated: true}, nil
	}
	if storedToken.UsedAt != nil || !storedToken.ExpiresAt.After(service.now()) {
		return Preview{}, errInvalidActivationLink
	}
	return Preview{Email: storedToken.Email, Username: storedToken.Username, IsActivated: false}, nil
}

func (service *Service) Activate(ctx context.Context, plaintextToken string) (ActivatedAccount, error) {
	if !securetoken.HasValidShape(plaintextToken) {
		return ActivatedAccount{}, errInvalidActivationLink
	}
	tokenHash := securetoken.Hash(plaintextToken)
	activatedAt := service.now().UTC()

	var activatedUser users.User
	err := database.WithTransaction(ctx, service.pool, func(transaction pgx.Tx) error {
		consumedUserID, err := activationstore.New(transaction).ConsumeActivationToken(ctx, activationstore.ConsumeActivationTokenParams{
			TokenHash: tokenHash,
			Now:       activatedAt,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return errInvalidActivationLink
		}
		if err != nil {
			return err
		}

		var wasActivated bool
		activatedUser, wasActivated, err = service.users.Activate(ctx, transaction, consumedUserID, activatedAt)
		if err != nil {
			return err
		}
		if !wasActivated {
			activatedUser, _, err = service.users.FindByID(ctx, transaction, consumedUserID)
		}
		return err
	})

	if errors.Is(err, errInvalidActivationLink) {
		return service.alreadyActivatedAccount(ctx, tokenHash)
	}
	if err != nil {
		return ActivatedAccount{}, err
	}
	return toActivatedAccount(activatedUser, activatedAt), nil
}

func (service *Service) alreadyActivatedAccount(ctx context.Context, tokenHash []byte) (ActivatedAccount, error) {
	storedToken, err := activationstore.New(service.pool).FindActivationTokenWithUser(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !storedToken.IsActive) {
		return ActivatedAccount{}, errInvalidActivationLink
	}
	if err != nil {
		return ActivatedAccount{}, err
	}
	activatedUser, isFound, err := service.users.FindByID(ctx, service.pool, storedToken.UserID)
	if err != nil {
		return ActivatedAccount{}, err
	}
	if !isFound {
		return ActivatedAccount{}, errInvalidActivationLink
	}
	return toActivatedAccount(activatedUser, service.now().UTC()), nil
}

func toActivatedAccount(activatedUser users.User, fallbackActivatedAt time.Time) ActivatedAccount {
	activatedAt := fallbackActivatedAt
	if activatedUser.ActivatedAt != nil {
		activatedAt = *activatedUser.ActivatedAt
	}
	return ActivatedAccount{
		Email:       activatedUser.Email,
		Username:    activatedUser.Username,
		Status:      activatedUser.Status,
		ActivatedAt: activatedAt,
	}
}
