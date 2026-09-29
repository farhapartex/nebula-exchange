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
	DefaultRefreshTokenLifetime   = 30 * 24 * time.Hour
	DefaultConcurrentRefreshGrace = 30 * time.Second
	refreshTokenByteLength        = 32
)

var ErrRefreshTokenInvalid = errors.New("refresh token is invalid, expired or revoked")

type RefreshTokenReusedError struct {
	UserID uuid.UUID
}

func (reusedError *RefreshTokenReusedError) Error() string {
	return "a revoked refresh token was presented again"
}

const maximumUserAgentLength = 256

type ClientMetadata struct {
	UserAgent string
	IPAddress string
}

func (metadata ClientMetadata) normalized() ClientMetadata {
	if len(metadata.UserAgent) > maximumUserAgentLength {
		metadata.UserAgent = metadata.UserAgent[:maximumUserAgentLength]
	}
	return metadata
}

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
	lifetime               time.Duration
	concurrentRefreshGrace time.Duration
	now                    func() time.Time
}

func NewRefreshTokens(lifetime time.Duration, now func() time.Time) *RefreshTokens {
	return &RefreshTokens{lifetime: lifetime, concurrentRefreshGrace: DefaultConcurrentRefreshGrace, now: now}
}

func (refreshTokens *RefreshTokens) Issue(ctx context.Context, database sessionstore.DBTX, userID uuid.UUID, metadata ClientMetadata) (IssuedRefreshToken, error) {
	return refreshTokens.issueForSession(ctx, database, userID, metadata, refreshTokens.now().UTC())
}

func (refreshTokens *RefreshTokens) issueForSession(ctx context.Context, database sessionstore.DBTX, userID uuid.UUID, metadata ClientMetadata, sessionStartedAt time.Time) (IssuedRefreshToken, error) {
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
	normalizedMetadata := metadata.normalized()
	err = sessionstore.New(database).CreateRefreshToken(ctx, sessionstore.CreateRefreshTokenParams{
		ID:               tokenID,
		UserID:           userID,
		TokenHash:        hashRefreshToken(plaintextToken),
		ExpiresAt:        expiresAt,
		UserAgent:        normalizedMetadata.UserAgent,
		IpAddress:        normalizedMetadata.IPAddress,
		SessionStartedAt: sessionStartedAt,
	})
	if err != nil {
		return IssuedRefreshToken{}, err
	}
	return IssuedRefreshToken{ID: tokenID, Plaintext: plaintextToken, ExpiresAt: expiresAt}, nil
}

func (refreshTokens *RefreshTokens) Rotate(ctx context.Context, transaction pgx.Tx, plaintextToken string, metadata ClientMetadata) (RotatedRefreshToken, error) {
	queries := sessionstore.New(transaction)
	tokenHash := hashRefreshToken(plaintextToken)
	now := refreshTokens.now().UTC()

	consumedToken, err := queries.ConsumeRefreshToken(ctx, sessionstore.ConsumeRefreshTokenParams{TokenHash: tokenHash, Now: now})
	if errors.Is(err, pgx.ErrNoRows) {
		return RotatedRefreshToken{}, refreshTokens.classifyUnusableToken(ctx, queries, tokenHash, now)
	}
	if err != nil {
		return RotatedRefreshToken{}, err
	}

	replacementToken, err := refreshTokens.issueForSession(ctx, transaction, consumedToken.UserID, metadata, consumedToken.SessionStartedAt)
	if err != nil {
		return RotatedRefreshToken{}, err
	}
	err = queries.LinkReplacementRefreshToken(ctx, sessionstore.LinkReplacementRefreshTokenParams{
		ID:         consumedToken.ID,
		ReplacedBy: &replacementToken.ID,
	})
	if err != nil {
		return RotatedRefreshToken{}, err
	}
	return RotatedRefreshToken{UserID: consumedToken.UserID, Issued: replacementToken}, nil
}

func (refreshTokens *RefreshTokens) classifyUnusableToken(ctx context.Context, queries *sessionstore.Queries, tokenHash []byte, now time.Time) error {
	storedToken, err := queries.FindRefreshTokenByHash(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRefreshTokenInvalid
	}
	if err != nil {
		return err
	}
	if storedToken.RevokedAt == nil {
		return ErrRefreshTokenInvalid
	}
	isConcurrentRefresh := storedToken.ReplacedBy != nil && now.Sub(*storedToken.RevokedAt) <= refreshTokens.concurrentRefreshGrace
	if isConcurrentRefresh {
		return ErrRefreshTokenInvalid
	}
	return &RefreshTokenReusedError{UserID: storedToken.UserID}
}

func (refreshTokens *RefreshTokens) Revoke(ctx context.Context, database sessionstore.DBTX, plaintextToken string) error {
	return sessionstore.New(database).RevokeRefreshTokenByHash(ctx, sessionstore.RevokeRefreshTokenByHashParams{
		TokenHash: hashRefreshToken(plaintextToken),
		Now:       refreshTokens.now().UTC(),
	})
}

func (refreshTokens *RefreshTokens) RevokeAllForUser(ctx context.Context, database sessionstore.DBTX, userID uuid.UUID) (int64, error) {
	return sessionstore.New(database).RevokeAllActiveRefreshTokensForUser(ctx, sessionstore.RevokeAllActiveRefreshTokensForUserParams{
		UserID: userID,
		Now:    refreshTokens.now().UTC(),
	})
}

func hashRefreshToken(plaintextToken string) []byte {
	tokenHash := sha256.Sum256([]byte(plaintextToken))
	return tokenHash[:]
}

func HashForCookie(plaintextToken string) []byte {
	return hashRefreshToken(plaintextToken)
}
