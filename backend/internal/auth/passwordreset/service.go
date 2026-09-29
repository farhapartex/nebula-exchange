package passwordreset

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/loginlockout"
	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/auth/passwordreset/passwordresetstore"
	"nebula-exchange/backend/internal/auth/securetoken"
	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/database"
	"nebula-exchange/backend/internal/users"
)

const (
	DefaultTokenLifetime = 30 * time.Minute
	resetPagePath        = "/reset"
)

var errInvalidResetLink = apierror.NotFound("This password reset link is invalid or has expired")

type Preview struct {
	EmailHint string `json:"email_hint"`
}

type Dependencies struct {
	Pool            *pgxpool.Pool
	Users           *users.Repository
	PasswordHasher  *passwordhash.Hasher
	RefreshTokens   *session.RefreshTokens
	LoginLockout    *loginlockout.Guard
	EmailTemplates  *email.TemplateRenderer
	EmailQueue      *outbox.Queue
	FrontendBaseURL string
	TokenLifetime   time.Duration
	Now             func() time.Time
}

type Service struct {
	dependencies Dependencies
}

func NewService(dependencies Dependencies) *Service {
	dependencies.FrontendBaseURL = strings.TrimRight(dependencies.FrontendBaseURL, "/")
	if dependencies.TokenLifetime == 0 {
		dependencies.TokenLifetime = DefaultTokenLifetime
	}
	return &Service{dependencies: dependencies}
}

func (service *Service) RequestReset(ctx context.Context, emailAddress string) error {
	credentials, isFound, err := service.dependencies.Users.FindCredentialsByEmail(ctx, service.dependencies.Pool, normalizeEmail(emailAddress))
	if err != nil {
		return err
	}
	if !isFound || !credentials.User.IsActive || credentials.User.Status == users.StatusBanned {
		return nil
	}

	accountOwner := credentials.User
	now := service.dependencies.Now().UTC()
	return database.WithTransaction(ctx, service.dependencies.Pool, func(transaction pgx.Tx) error {
		queries := passwordresetstore.New(transaction)
		err := queries.ExpireUnusedPasswordResetTokens(ctx, passwordresetstore.ExpireUnusedPasswordResetTokensParams{
			UserID: accountOwner.ID,
			Now:    now,
		})
		if err != nil {
			return err
		}

		generatedToken, err := securetoken.Generate()
		if err != nil {
			return err
		}
		tokenID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		expiresAt := now.Add(service.dependencies.TokenLifetime)
		err = queries.CreatePasswordResetToken(ctx, passwordresetstore.CreatePasswordResetTokenParams{
			ID:        tokenID,
			UserID:    accountOwner.ID,
			TokenHash: generatedToken.Hash,
			ExpiresAt: expiresAt,
		})
		if err != nil {
			return err
		}
		return service.queueResetEmail(ctx, transaction, accountOwner, generatedToken.Plaintext, expiresAt)
	})
}

func (service *Service) queueResetEmail(ctx context.Context, transaction pgx.Tx, accountOwner users.User, plaintextToken string, expiresAt time.Time) error {
	htmlBody, textBody, err := service.dependencies.EmailTemplates.Render(email.TemplatePasswordReset, struct {
		Username  string
		ResetLink string
		ExpiresAt string
	}{
		Username:  accountOwner.Username,
		ResetLink: service.dependencies.FrontendBaseURL + resetPagePath + "?token=" + url.QueryEscape(plaintextToken),
		ExpiresAt: expiresAt.Format(time.RFC1123),
	})
	if err != nil {
		return err
	}
	_, err = service.dependencies.EmailQueue.Enqueue(ctx, transaction, email.TemplatePasswordReset, email.Message{
		To:       email.Address{Name: accountOwner.Username, Email: accountOwner.Email},
		Subject:  "Reset your Nebula Exchange password",
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
	return err
}

func (service *Service) Preview(ctx context.Context, plaintextToken string) (Preview, error) {
	if !securetoken.HasValidShape(plaintextToken) {
		return Preview{}, errInvalidResetLink
	}
	usableToken, err := passwordresetstore.New(service.dependencies.Pool).FindUsablePasswordResetToken(ctx, passwordresetstore.FindUsablePasswordResetTokenParams{
		TokenHash: securetoken.Hash(plaintextToken),
		Now:       service.dependencies.Now().UTC(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Preview{}, errInvalidResetLink
	}
	if err != nil {
		return Preview{}, err
	}
	return Preview{EmailHint: MaskEmail(usableToken.Email)}, nil
}

func (service *Service) ResetPassword(ctx context.Context, plaintextToken, newPassword string) error {
	if _, err := service.Preview(ctx, plaintextToken); err != nil {
		return err
	}
	newPasswordHash, err := service.dependencies.PasswordHasher.Hash(ctx, newPassword)
	if err != nil {
		return err
	}

	now := service.dependencies.Now().UTC()
	var accountEmail string
	err = database.WithTransaction(ctx, service.dependencies.Pool, func(transaction pgx.Tx) error {
		consumedUserID, err := passwordresetstore.New(transaction).ConsumePasswordResetToken(ctx, passwordresetstore.ConsumePasswordResetTokenParams{
			TokenHash: securetoken.Hash(plaintextToken),
			Now:       now,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return errInvalidResetLink
		}
		if err != nil {
			return err
		}
		accountEmail, err = service.dependencies.Users.UpdatePasswordHash(ctx, transaction, consumedUserID, newPasswordHash, now)
		if err != nil {
			return err
		}
		_, err = service.dependencies.RefreshTokens.RevokeAllForUser(ctx, transaction, consumedUserID)
		return err
	})
	if err != nil {
		return err
	}

	service.dependencies.LoginLockout.Unlock(ctx, accountEmail)
	return nil
}

func normalizeEmail(emailAddress string) string {
	return strings.ToLower(strings.TrimSpace(emailAddress))
}
