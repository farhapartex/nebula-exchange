package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/auth/session/sessionstore"
)

const (
	DefaultRefreshTokenLifetime = 30 * 24 * time.Hour
	refreshTokenByteLength      = 32
)

var ErrRefreshTokenInvalid = errors.New("refresh token is invalid, expired or revoked")

type IssuedRefreshToken struct {
	ID        uuid.UUID
	Plaintext string
	ExpiresAt time.Time
}

type RotatedRefreshToken struct {
	UserID uuid.UUID
	Issued IssuedRefreshToken
}

type RefreshTokens struct {
	lifetime time.Duration
	now      func() time.Time
}

func NewRefreshTokens(lifetime time.Duration, now func() time.Time) *RefreshTokens {
	return &RefreshTokens{lifetime: lifetime, now: now}
}

func (refreshTokens *RefreshTokens) Issue(ctx context.Context, database sessionstore.DBTX, userID uuid.UUID) (IssuedRefreshToken, error) {
	randomBytes := make([]byte, refreshTokenByteLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return IssuedRefreshToken{}, fmt.Errorf("generate refresh token: %w", err)
	}
	tokenID, err := uuid.NewV7()
	if err != nil {
		return IssuedRefreshToken{}, err
	}

	plaintextToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	expiresAt := refreshTokens.now().UTC().Add(refreshTokens.lifetime)
	err = sessionstore.New(database).CreateRefreshToken(ctx, sessionstore.CreateRefreshTokenParams{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: hashRefreshToken(plaintextToken),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return IssuedRefreshToken{}, err
	}
	return IssuedRefreshToken{ID: tokenID, Plaintext: plaintextToken, ExpiresAt: expiresAt}, nil
}

func (refreshTokens *RefreshTokens) Rotate(ctx context.Context, transaction pgx.Tx, plaintextToken string) (RotatedRefreshToken, error) {
	queries := sessionstore.New(transaction)
	storedToken, err := queries.FindRefreshTokenForUpdate(ctx, hashRefreshToken(plaintextToken))
	if errors.Is(err, pgx.ErrNoRows) {
		return RotatedRefreshToken{}, ErrRefreshTokenInvalid
	}
	if err != nil {
		return RotatedRefreshToken{}, err
	}

	now := refreshTokens.now().UTC()
	if storedToken.RevokedAt != nil || !storedToken.ExpiresAt.After(now) {
		return RotatedRefreshToken{}, ErrRefreshTokenInvalid
	}

	replacementToken, err := refreshTokens.Issue(ctx, transaction, storedToken.UserID)
	if err != nil {
		return RotatedRefreshToken{}, err
	}
	err = queries.RevokeRefreshToken(ctx, sessionstore.RevokeRefreshTokenParams{
		ID:         storedToken.ID,
		RevokedAt:  &now,
		ReplacedBy: &replacementToken.ID,
	})
	if err != nil {
		return RotatedRefreshToken{}, err
	}
	return RotatedRefreshToken{UserID: storedToken.UserID, Issued: replacementToken}, nil
}

func hashRefreshToken(plaintextToken string) []byte {
	tokenHash := sha256.Sum256([]byte(plaintextToken))
	return tokenHash[:]
}
