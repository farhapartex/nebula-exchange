package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/emails"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/securetoken"
)

const ActivationLinkLifetime = 24 * time.Hour

type SignupInput struct {
	Email        string
	Username     string
	Password     string
	AcceptsTerms bool
}

type SignedUpAccount struct {
	User                    models.User
	ActivationLinkExpiresAt time.Time
}

type SignupService interface {
	SignUp(ctx context.Context, input SignupInput) (SignedUpAccount, error)
}

type SignupDependencies struct {
	Users            repository.UserRepository
	ActivationTokens repository.ActivationTokenRepository
	PasswordHasher   PasswordHasher
	EmailComposer    ActivationEmailComposer
	EmailEnqueuer    EmailEnqueuer
	Transactions     TransactionRunner
	Now              Clock
}

type signupService struct {
	dependencies SignupDependencies
}

func NewSignupService(dependencies SignupDependencies) SignupService {
	return &signupService{dependencies: dependencies}
}

func (signup *signupService) SignUp(ctx context.Context, input SignupInput) (SignedUpAccount, error) {
	input = input.normalized()
	if err := input.validate(); err != nil {
		return SignedUpAccount{}, err
	}

	takenIdentifiers, err := signup.dependencies.Users.FindTakenIdentifiers(ctx, input.Email, input.Username)
	if err != nil {
		return SignedUpAccount{}, err
	}
	if takenIdentifiers.Any() {
		return SignedUpAccount{}, takenIdentifiersError(takenIdentifiers)
	}

	passwordHash, err := signup.dependencies.PasswordHasher.Hash(ctx, input.Password)
	if err != nil {
		return SignedUpAccount{}, err
	}

	newUser, err := signup.buildUser(input, passwordHash)
	if err != nil {
		return SignedUpAccount{}, err
	}

	var activationLinkExpiresAt time.Time
	err = signup.dependencies.Transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := signup.dependencies.Users.Create(ctx, &newUser); err != nil {
			return err
		}
		activationLinkExpiresAt, err = signup.queueActivationEmail(ctx, newUser)
		return err
	})
	if err != nil {
		return SignedUpAccount{}, translateCreateError(err)
	}
	return SignedUpAccount{User: newUser, ActivationLinkExpiresAt: activationLinkExpiresAt}, nil
}

func (signup *signupService) buildUser(input SignupInput, passwordHash string) (models.User, error) {
	userID, err := uuid.NewV7()
	if err != nil {
		return models.User{}, err
	}
	signedUpAt := signup.dependencies.Now().UTC()
	return models.User{
		ID:              userID,
		Email:           input.Email,
		Username:        input.Username,
		PasswordHash:    passwordHash,
		Status:          models.UserStatusUnverified,
		TermsAcceptedAt: signedUpAt,
		CreatedAt:       signedUpAt,
		UpdatedAt:       signedUpAt,
	}, nil
}

func (signup *signupService) queueActivationEmail(ctx context.Context, newUser models.User) (time.Time, error) {
	generatedToken, err := securetoken.Generate()
	if err != nil {
		return time.Time{}, err
	}
	tokenID, err := uuid.NewV7()
	if err != nil {
		return time.Time{}, err
	}
	issuedAt := signup.dependencies.Now().UTC()
	expiresAt := issuedAt.Add(ActivationLinkLifetime)
	err = signup.dependencies.ActivationTokens.Create(ctx, &models.ActivationToken{
		ID:        tokenID,
		UserID:    newUser.ID,
		TokenHash: generatedToken.Hash,
		ExpiresAt: expiresAt,
		CreatedAt: issuedAt,
	})
	if err != nil {
		return time.Time{}, err
	}

	activationMessage, err := signup.dependencies.EmailComposer.Compose(
		email.Address{Name: newUser.Username, Email: newUser.Email},
		generatedToken.Plaintext,
		expiresAt,
	)
	if err != nil {
		return time.Time{}, err
	}
	return expiresAt, signup.dependencies.EmailEnqueuer.Enqueue(ctx, emails.ActivationTemplateName, activationMessage)
}

func takenIdentifiersError(takenIdentifiers repository.TakenIdentifiers) error {
	fieldErrors := map[string]string{}
	if takenIdentifiers.IsEmailTaken {
		fieldErrors["email"] = "is already registered"
	}
	if takenIdentifiers.IsUsernameTaken {
		fieldErrors["username"] = "is already taken"
	}
	return apierror.ValidationFailed(fieldErrors)
}

func translateCreateError(err error) error {
	switch {
	case errors.Is(err, repository.ErrEmailTaken):
		return takenIdentifiersError(repository.TakenIdentifiers{IsEmailTaken: true})
	case errors.Is(err, repository.ErrUsernameTaken):
		return takenIdentifiersError(repository.TakenIdentifiers{IsUsernameTaken: true})
	default:
		return err
	}
}
