package loginchallenge

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"nebula-exchange/backend/internal/auth/securetoken"
)

const (
	DefaultLifetime     = 5 * time.Minute
	MaximumCodeAttempts = 5
	userIDField         = "user_id"
	emailField          = "email"
	attemptsField       = "attempts"
)

var ErrChallengeInvalid = errors.New("login challenge is invalid, expired or used up")

type IssuedChallenge struct {
	Token     string
	ExpiresAt time.Time
}

type Challenge struct {
	UserID uuid.UUID
	Email  string
}

type Store struct {
	client    redis.Cmdable
	keyPrefix string
	lifetime  time.Duration
	now       func() time.Time
}

func NewStore(client redis.Cmdable, keyPrefix string, lifetime time.Duration, now func() time.Time) *Store {
	return &Store{client: client, keyPrefix: keyPrefix, lifetime: lifetime, now: now}
}

func (store *Store) Issue(ctx context.Context, userID uuid.UUID, emailAddress string) (IssuedChallenge, error) {
	generatedToken, err := securetoken.Generate()
	if err != nil {
		return IssuedChallenge{}, err
	}
	challengeKey := store.key(generatedToken.Plaintext)
	pipeline := store.client.TxPipeline()
	pipeline.HSet(ctx, challengeKey, userIDField, userID.String(), emailField, emailAddress, attemptsField, 0)
	pipeline.PExpire(ctx, challengeKey, store.lifetime)
	if _, err := pipeline.Exec(ctx); err != nil {
		return IssuedChallenge{}, err
	}
	return IssuedChallenge{Token: generatedToken.Plaintext, ExpiresAt: store.now().UTC().Add(store.lifetime)}, nil
}

func (store *Store) Find(ctx context.Context, plaintextToken string) (Challenge, error) {
	if !securetoken.HasValidShape(plaintextToken) {
		return Challenge{}, ErrChallengeInvalid
	}
	storedFields, err := store.client.HGetAll(ctx, store.key(plaintextToken)).Result()
	if err != nil {
		return Challenge{}, err
	}
	userID, err := uuid.Parse(storedFields[userIDField])
	if err != nil {
		return Challenge{}, ErrChallengeInvalid
	}
	return Challenge{UserID: userID, Email: storedFields[emailField]}, nil
}

func (store *Store) RecordFailedAttempt(ctx context.Context, plaintextToken string) error {
	challengeKey := store.key(plaintextToken)
	attemptCount, err := store.client.HIncrBy(ctx, challengeKey, attemptsField, 1).Result()
	if err != nil {
		return err
	}
	if attemptCount >= MaximumCodeAttempts {
		return store.client.Del(ctx, challengeKey).Err()
	}
	return nil
}

func (store *Store) Consume(ctx context.Context, plaintextToken string) error {
	return store.client.Del(ctx, store.key(plaintextToken)).Err()
}

func (store *Store) key(plaintextToken string) string {
	tokenHash := sha256.Sum256([]byte(plaintextToken))
	return store.keyPrefix + hex.EncodeToString(tokenHash[:])
}
