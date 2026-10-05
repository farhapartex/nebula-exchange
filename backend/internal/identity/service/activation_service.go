package service

import (
	"context"
	"errors"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/securetoken"
)

type ActivationPreview struct {
	Email       string
	Username    string
	IsActivated bool
}

type ActivationService interface {
	Preview(ctx context.Context, plaintextToken string) (ActivationPreview, error)
	Activate(ctx context.Context, plaintextToken string) (models.User, error)
}

type ActivationDependencies struct {
	Users            repository.UserRepository
	ActivationTokens repository.ActivationTokenRepository
	Transactions     TransactionRunner
	Now              Clock
}

type activationService struct {
	dependencies ActivationDependencies
}

func NewActivationService(dependencies ActivationDependencies) ActivationService {
	return &activationService{dependencies: dependencies}
}

func (activation *activationService) Preview(ctx context.Context, plaintextToken string) (ActivationPreview, error) {
	storedToken, tokenOwner, err := activation.findTokenWithOwner(ctx, plaintextToken)
	if err != nil {
		return ActivationPreview{}, err
	}
	if tokenOwner.IsActive {
		return ActivationPreview{Email: tokenOwner.Email, Username: tokenOwner.Username, IsActivated: true}, nil
	}
	if storedToken.UsedAt != nil || !storedToken.ExpiresAt.After(activation.dependencies.Now()) {
		return ActivationPreview{}, ErrInvalidActivationLink
	}
	return ActivationPreview{Email: tokenOwner.Email, Username: tokenOwner.Username}, nil
}

func (activation *activationService) Activate(ctx context.Context, plaintextToken string) (models.User, error) {
	if !securetoken.HasValidShape(plaintextToken) {
		return models.User{}, ErrInvalidActivationLink
	}
	tokenHash := securetoken.Hash(plaintextToken)
	activatedAt := activation.dependencies.Now().UTC()

	var activatedUser models.User
	err := activation.dependencies.Transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		consumedToken, isConsumed, err := activation.dependencies.ActivationTokens.Consume(ctx, tokenHash, activatedAt)
		if err != nil {
			return err
		}
		if !isConsumed {
			return ErrInvalidActivationLink
		}
		var wasActivated bool
		activatedUser, wasActivated, err = activation.dependencies.Users.Activate(ctx, consumedToken.UserID, activatedAt)
		if err != nil || wasActivated {
			return err
		}
		activatedUser, _, err = activation.dependencies.Users.FindByID(ctx, consumedToken.UserID)
		return err
	})
	if errors.Is(err, ErrInvalidActivationLink) {
		return activation.alreadyActivatedUser(ctx, plaintextToken)
	}
	return activatedUser, err
}

func (activation *activationService) alreadyActivatedUser(ctx context.Context, plaintextToken string) (models.User, error) {
	_, tokenOwner, err := activation.findTokenWithOwner(ctx, plaintextToken)
	if err != nil {
		return models.User{}, err
	}
	if !tokenOwner.IsActive {
		return models.User{}, ErrInvalidActivationLink
	}
	return tokenOwner, nil
}

func (activation *activationService) findTokenWithOwner(ctx context.Context, plaintextToken string) (models.ActivationToken, models.User, error) {
	if !securetoken.HasValidShape(plaintextToken) {
		return models.ActivationToken{}, models.User{}, ErrInvalidActivationLink
	}
	storedToken, isFound, err := activation.dependencies.ActivationTokens.FindByHash(ctx, securetoken.Hash(plaintextToken))
	if err != nil {
		return models.ActivationToken{}, models.User{}, err
	}
	if !isFound {
		return models.ActivationToken{}, models.User{}, ErrInvalidActivationLink
	}
	tokenOwner, isFound, err := activation.dependencies.Users.FindByID(ctx, storedToken.UserID)
	if err != nil {
		return models.ActivationToken{}, models.User{}, err
	}
	if !isFound {
		return models.ActivationToken{}, models.User{}, ErrInvalidActivationLink
	}
	return storedToken, tokenOwner, nil
}
