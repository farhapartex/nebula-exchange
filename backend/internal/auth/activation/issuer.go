package activation

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nebula-exchange/backend/internal/auth/activation/activationstore"
	"nebula-exchange/backend/internal/auth/securetoken"
)

const DefaultTokenLifetime = 24 * time.Hour

type IssuedToken struct {
	Plaintext string
	ExpiresAt time.Time
}

type Issuer struct {
	tokenLifetime time.Duration
	now           func() time.Time
}

func NewIssuer(tokenLifetime time.Duration, now func() time.Time) *Issuer {
	return &Issuer{tokenLifetime: tokenLifetime, now: now}
}

func (issuer *Issuer) Issue(ctx context.Context, database activationstore.DBTX, userID uuid.UUID) (IssuedToken, error) {
	generatedToken, err := securetoken.Generate()
	if err != nil {
		return IssuedToken{}, err
	}
	tokenID, err := uuid.NewV7()
	if err != nil {
		return IssuedToken{}, err
	}

	expiresAt := issuer.now().Add(issuer.tokenLifetime).UTC()
	err = activationstore.New(database).CreateActivationToken(ctx, activationstore.CreateActivationTokenParams{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: generatedToken.Hash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return IssuedToken{}, err
	}
	return IssuedToken{Plaintext: generatedToken.Plaintext, ExpiresAt: expiresAt}, nil
}
