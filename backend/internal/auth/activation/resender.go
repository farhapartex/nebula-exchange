package activation

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/activation/activationstore"
	"nebula-exchange/backend/internal/platform/database"
	"nebula-exchange/backend/internal/users"
)

type Resender struct {
	pool   *pgxpool.Pool
	users  *users.Repository
	issuer *Issuer
	mailer *Mailer
	now    func() time.Time
}

func NewResender(pool *pgxpool.Pool, userRepository *users.Repository, issuer *Issuer, mailer *Mailer, now func() time.Time) *Resender {
	return &Resender{pool: pool, users: userRepository, issuer: issuer, mailer: mailer, now: now}
}

func NormalizeEmail(emailAddress string) string {
	return strings.ToLower(strings.TrimSpace(emailAddress))
}

func (resender *Resender) ResendActivationEmail(ctx context.Context, emailAddress string) error {
	credentials, isFound, err := resender.users.FindCredentialsByEmail(ctx, resender.pool, NormalizeEmail(emailAddress))
	if err != nil {
		return err
	}
	if !isFound || credentials.User.IsActive {
		return nil
	}

	pendingUser := credentials.User
	return database.WithTransaction(ctx, resender.pool, func(transaction pgx.Tx) error {
		err := activationstore.New(transaction).ExpireUnusedActivationTokens(ctx, activationstore.ExpireUnusedActivationTokensParams{
			UserID: pendingUser.ID,
			Now:    resender.now().UTC(),
		})
		if err != nil {
			return err
		}
		issuedToken, err := resender.issuer.Issue(ctx, transaction, pendingUser.ID)
		if err != nil {
			return err
		}
		return resender.mailer.QueueActivationEmail(ctx, transaction, pendingUser.Email, pendingUser.Username, issuedToken)
	})
}
