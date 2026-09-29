package twofactor

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"nebula-exchange/backend/internal/auth/twofactor/twofactorstore"
	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/database"
	"nebula-exchange/backend/internal/platform/secretbox"
)

const (
	issuerName           = "Nebula Exchange"
	codePeriodSeconds    = 30
	allowedStepDeviation = 1
)

var (
	errAlreadyEnabled = apierror.Conflict("Two-factor authentication is already on")
	errNotEnabled     = apierror.Conflict("Two-factor authentication is not on")
	errSetupMissing   = apierror.Conflict("Start the two-factor setup first")
	errIncorrectCode  = apierror.ValidationFailed(map[string]string{"code": "is incorrect or already used"})
	ErrCodeRequired   = apierror.New(http.StatusForbidden, apierror.CodeTwoFactorRequired, "Enter your authenticator code to continue")
)

type SetupDetails struct {
	Secret     string `json:"secret"`
	OtpauthURL string `json:"otpauth_url"`
}

type Dependencies struct {
	Pool           *pgxpool.Pool
	SecretBox      *secretbox.Box
	EmailTemplates *email.TemplateRenderer
	EmailQueue     *outbox.Queue
	Now            func() time.Time
}

type Service struct {
	dependencies Dependencies
}

func NewService(dependencies Dependencies) *Service {
	return &Service{dependencies: dependencies}
}

func (service *Service) BeginSetup(ctx context.Context, userID uuid.UUID) (SetupDetails, error) {
	queries := twofactorstore.New(service.dependencies.Pool)
	twoFactorState, err := queries.FindTwoFactorState(ctx, userID)
	if err != nil {
		return SetupDetails{}, err
	}
	if twoFactorState.TotpEnabledAt != nil {
		return SetupDetails{}, errAlreadyEnabled
	}

	generatedKey, err := totp.Generate(totp.GenerateOpts{Issuer: issuerName, AccountName: twoFactorState.Email})
	if err != nil {
		return SetupDetails{}, err
	}
	encryptedSecret, err := service.dependencies.SecretBox.Seal([]byte(generatedKey.Secret()))
	if err != nil {
		return SetupDetails{}, err
	}
	updatedRows, err := queries.SetPendingTotpSecret(ctx, twofactorstore.SetPendingTotpSecretParams{
		ID:            userID,
		TotpSecretEnc: encryptedSecret,
		Now:           service.dependencies.Now().UTC(),
	})
	if err != nil {
		return SetupDetails{}, err
	}
	if updatedRows == 0 {
		return SetupDetails{}, errAlreadyEnabled
	}
	return SetupDetails{Secret: generatedKey.Secret(), OtpauthURL: generatedKey.URL()}, nil
}

func (service *Service) Enable(ctx context.Context, userID uuid.UUID, code string) error {
	twoFactorState, err := twofactorstore.New(service.dependencies.Pool).FindTwoFactorState(ctx, userID)
	if err != nil {
		return err
	}
	if twoFactorState.TotpEnabledAt != nil {
		return errAlreadyEnabled
	}
	if twoFactorState.TotpSecretEnc == nil {
		return errSetupMissing
	}

	return database.WithTransaction(ctx, service.dependencies.Pool, func(transaction pgx.Tx) error {
		if err := service.verifyAndClaimCode(ctx, transaction, userID, twoFactorState.TotpSecretEnc, code); err != nil {
			return err
		}
		if _, err := twofactorstore.New(transaction).EnableTotp(ctx, twofactorstore.EnableTotpParams{
			ID:  userID,
			Now: service.dependencies.Now().UTC(),
		}); err != nil {
			return err
		}
		return service.queueChangeEmail(ctx, transaction, twoFactorState.Email, twoFactorState.Username, "turned on")
	})
}

func (service *Service) Disable(ctx context.Context, userID uuid.UUID, code string) error {
	twoFactorState, err := twofactorstore.New(service.dependencies.Pool).FindTwoFactorState(ctx, userID)
	if err != nil {
		return err
	}
	if twoFactorState.TotpEnabledAt == nil {
		return errNotEnabled
	}

	return database.WithTransaction(ctx, service.dependencies.Pool, func(transaction pgx.Tx) error {
		if err := service.verifyAndClaimCode(ctx, transaction, userID, twoFactorState.TotpSecretEnc, code); err != nil {
			return err
		}
		if err := twofactorstore.New(transaction).DisableTotp(ctx, twofactorstore.DisableTotpParams{
			ID:  userID,
			Now: service.dependencies.Now().UTC(),
		}); err != nil {
			return err
		}
		return service.queueChangeEmail(ctx, transaction, twoFactorState.Email, twoFactorState.Username, "turned off")
	})
}

func (service *Service) VerifyCode(ctx context.Context, userID uuid.UUID, code string) error {
	twoFactorState, err := twofactorstore.New(service.dependencies.Pool).FindTwoFactorState(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return errIncorrectCode
	}
	if err != nil {
		return err
	}
	if twoFactorState.TotpEnabledAt == nil {
		return errNotEnabled
	}
	return service.verifyAndClaimCode(ctx, service.dependencies.Pool, userID, twoFactorState.TotpSecretEnc, code)
}

func (service *Service) verifyAndClaimCode(ctx context.Context, database twofactorstore.DBTX, userID uuid.UUID, encryptedSecret []byte, code string) error {
	secret, err := service.dependencies.SecretBox.Open(encryptedSecret)
	if err != nil {
		return err
	}
	matchedStep, isMatch := service.matchingStep(string(secret), code)
	if !isMatch {
		return errIncorrectCode
	}
	claimedRows, err := twofactorstore.New(database).ClaimTotpStep(ctx, twofactorstore.ClaimTotpStepParams{ID: userID, Step: matchedStep})
	if err != nil {
		return err
	}
	if claimedRows == 0 {
		return errIncorrectCode
	}
	return nil
}

func (service *Service) matchingStep(secret, code string) (int64, bool) {
	currentStep := service.dependencies.Now().Unix() / codePeriodSeconds
	for stepOffset := -allowedStepDeviation; stepOffset <= allowedStepDeviation; stepOffset++ {
		candidateStep := currentStep + int64(stepOffset)
		expectedCode, err := totp.GenerateCodeCustom(secret, time.Unix(candidateStep*codePeriodSeconds, 0), totp.ValidateOpts{
			Period:    codePeriodSeconds,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		})
		if err == nil && subtle.ConstantTimeCompare([]byte(expectedCode), []byte(code)) == 1 {
			return candidateStep, true
		}
	}
	return 0, false
}

func (service *Service) queueChangeEmail(ctx context.Context, transaction pgx.Tx, recipientEmail, username, changeDescription string) error {
	htmlBody, textBody, err := service.dependencies.EmailTemplates.Render(email.TemplateTwoFactorChanged, struct {
		Username          string
		ChangeDescription string
		ChangedAt         string
	}{
		Username:          username,
		ChangeDescription: changeDescription,
		ChangedAt:         service.dependencies.Now().UTC().Format(time.RFC1123),
	})
	if err != nil {
		return err
	}
	_, err = service.dependencies.EmailQueue.Enqueue(ctx, transaction, email.TemplateTwoFactorChanged, email.Message{
		To:       email.Address{Name: username, Email: recipientEmail},
		Subject:  "Two-factor authentication was " + changeDescription,
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
	return err
}
