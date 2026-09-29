package accesstoken

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	DefaultLifetime     = 15 * time.Minute
	MinimumSecretLength = 32
	issuerName          = "nebula-exchange"
	allowedClockSkew    = 30 * time.Second
)

var (
	ErrSecretTooShort = fmt.Errorf("JWT secret must be at least %d bytes", MinimumSecretLength)
	ErrInvalidToken   = errors.New("access token is invalid or expired")
)

type IssuedToken struct {
	Value     string
	ExpiresAt time.Time
}

type Manager struct {
	secret   []byte
	lifetime time.Duration
	now      func() time.Time
}

func NewManager(secret string, lifetime time.Duration, now func() time.Time) (*Manager, error) {
	if len(secret) < MinimumSecretLength {
		return nil, ErrSecretTooShort
	}
	return &Manager{secret: []byte(secret), lifetime: lifetime, now: now}, nil
}

func (manager *Manager) Issue(userID uuid.UUID) (IssuedToken, error) {
	issuedAt := manager.now().UTC()
	expiresAt := issuedAt.Add(manager.lifetime)
	tokenID, err := uuid.NewV7()
	if err != nil {
		return IssuedToken{}, err
	}

	signedToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuerName,
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		ID:        tokenID.String(),
	}).SignedString(manager.secret)
	if err != nil {
		return IssuedToken{}, fmt.Errorf("sign access token: %w", err)
	}
	return IssuedToken{Value: signedToken, ExpiresAt: expiresAt}, nil
}

func (manager *Manager) Verify(tokenValue string) (uuid.UUID, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(tokenValue, &claims, func(*jwt.Token) (any, error) {
		return manager.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuerName),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(allowedClockSkew),
		jwt.WithTimeFunc(manager.now),
	)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	return userID, nil
}
