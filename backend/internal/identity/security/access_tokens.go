package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	DefaultAccessTokenLifetime     = 15 * time.Minute
	MinimumAccessTokenSecretLength = 32
	issuerName                     = "street-born"
	allowedClockSkew               = 30 * time.Second
)

var (
	ErrAccessTokenSecretTooShort = fmt.Errorf("JWT secret must be at least %d bytes", MinimumAccessTokenSecretLength)
	ErrAccessTokenInvalid        = errors.New("access token is invalid or expired")
)

type IssuedAccessToken struct {
	Value     string
	ExpiresAt time.Time
}

type AccessTokens struct {
	secret   []byte
	lifetime time.Duration
	now      func() time.Time
}

func NewAccessTokens(secret string, lifetime time.Duration, now func() time.Time) (*AccessTokens, error) {
	if len(secret) < MinimumAccessTokenSecretLength {
		return nil, ErrAccessTokenSecretTooShort
	}
	return &AccessTokens{secret: []byte(secret), lifetime: lifetime, now: now}, nil
}

func (manager *AccessTokens) Issue(userID uuid.UUID) (IssuedAccessToken, error) {
	issuedAt := manager.now().UTC()
	expiresAt := issuedAt.Add(manager.lifetime)
	tokenID, err := uuid.NewV7()
	if err != nil {
		return IssuedAccessToken{}, err
	}

	signedToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuerName,
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		ID:        tokenID.String(),
	}).SignedString(manager.secret)
	if err != nil {
		return IssuedAccessToken{}, fmt.Errorf("sign access token: %w", err)
	}
	return IssuedAccessToken{Value: signedToken, ExpiresAt: expiresAt}, nil
}

func (manager *AccessTokens) Verify(tokenValue string) (uuid.UUID, error) {
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
		return uuid.Nil, ErrAccessTokenInvalid
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrAccessTokenInvalid
	}
	return userID, nil
}
