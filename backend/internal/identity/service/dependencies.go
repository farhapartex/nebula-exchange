package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/security"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email"
)

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Verify(ctx context.Context, password, encodedHash string) (bool, error)
}

type AccessTokenIssuer interface {
	Issue(userID uuid.UUID) (security.IssuedAccessToken, error)
}

type ActivationEmailComposer interface {
	Compose(recipient email.Address, plaintextToken string, expiresAt time.Time) (email.Message, error)
}

type EmailEnqueuer interface {
	Enqueue(ctx context.Context, templateName string, message email.Message) error
}

type TransactionRunner interface {
	WithinTransaction(ctx context.Context, work func(ctx context.Context) error) error
}

type Clock func() time.Time
