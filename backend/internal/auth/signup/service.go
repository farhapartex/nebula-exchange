package signup

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/activation"
	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/database"
	"nebula-exchange/backend/internal/users"
)

type SignedUpAccount struct {
	ID                      uuid.UUID    `json:"id"`
	Email                   string       `json:"email"`
	Username                string       `json:"username"`
	Status                  users.Status `json:"status"`
	ActivationLinkExpiresAt time.Time    `json:"activation_link_expires_at"`
}

type Dependencies struct {
	Pool             *pgxpool.Pool
	Users            *users.Repository
	ActivationIssuer *activation.Issuer
	ActivationEmail  *activation.EmailComposer
	EmailQueue       *outbox.Queue
	PasswordHasher   *passwordhash.Hasher
	Now              func() time.Time
}

type Service struct {
	dependencies Dependencies
}

func NewService(dependencies Dependencies) *Service {
	return &Service{dependencies: dependencies}
}

func (service *Service) SignUp(ctx context.Context, request Request) (SignedUpAccount, error) {
	request = request.normalized()
	if err := request.validateBusinessRules(); err != nil {
		return SignedUpAccount{}, err
	}

	takenIdentifiers, err := service.dependencies.Users.FindTakenIdentifiers(ctx, service.dependencies.Pool, request.Email, request.Username)
	if err != nil {
		return SignedUpAccount{}, err
	}
	if takenIdentifiers.Any() {
		return SignedUpAccount{}, takenIdentifiersError(takenIdentifiers)
	}

	passwordHash, err := service.dependencies.PasswordHasher.Hash(ctx, request.Password)
	if err != nil {
		return SignedUpAccount{}, err
	}

	var createdUser users.User
	var issuedToken activation.IssuedToken
	err = database.WithTransaction(ctx, service.dependencies.Pool, func(transaction pgx.Tx) error {
		createdUser, err = service.dependencies.Users.Create(ctx, transaction, users.NewUser{
			Email:           request.Email,
			Username:        request.Username,
			PasswordHash:    passwordHash,
			TermsAcceptedAt: service.dependencies.Now().UTC(),
		})
		if err != nil {
			return err
		}
		issuedToken, err = service.dependencies.ActivationIssuer.Issue(ctx, transaction, createdUser.ID)
		if err != nil {
			return err
		}
		return service.queueActivationEmail(ctx, transaction, createdUser, issuedToken)
	})
	if err != nil {
		return SignedUpAccount{}, translateCreateError(err)
	}

	return SignedUpAccount{
		ID:                      createdUser.ID,
		Email:                   createdUser.Email,
		Username:                createdUser.Username,
		Status:                  createdUser.Status,
		ActivationLinkExpiresAt: issuedToken.ExpiresAt,
	}, nil
}

func (service *Service) queueActivationEmail(ctx context.Context, transaction pgx.Tx, createdUser users.User, issuedToken activation.IssuedToken) error {
	activationMessage, err := service.dependencies.ActivationEmail.Compose(
		email.Address{Name: createdUser.Username, Email: createdUser.Email},
		createdUser.Username,
		issuedToken,
	)
	if err != nil {
		return err
	}
	_, err = service.dependencies.EmailQueue.Enqueue(ctx, transaction, email.TemplateAccountActivation, activationMessage)
	return err
}

func takenIdentifiersError(takenIdentifiers users.TakenIdentifiers) error {
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
	case errors.Is(err, users.ErrEmailTaken):
		return takenIdentifiersError(users.TakenIdentifiers{IsEmailTaken: true})
	case errors.Is(err, users.ErrUsernameTaken):
		return takenIdentifiersError(users.TakenIdentifiers{IsUsernameTaken: true})
	default:
		return err
	}
}
