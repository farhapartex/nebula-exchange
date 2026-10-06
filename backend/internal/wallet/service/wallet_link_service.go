package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
	"github.com/farhapartex/nebula-exchange/backend/internal/wallet/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/wallet/repository"
)

const (
	challengeLifetime       = 5 * time.Minute
	nonceByteLength         = 16
	supportedMessageVersion = "1"
	allowedClockSkew        = time.Minute
)

var (
	ErrUnsupportedChain       = apierror.ValidationFailed(map[string]string{"chain_id": "is not supported"})
	ErrMalformedMessage       = apierror.ValidationFailed(map[string]string{"message": "is not a valid sign-in message"})
	ErrMessageForAnotherSite  = apierror.ValidationFailed(map[string]string{"message": "was made for another site"})
	ErrMessageForAnotherChain = apierror.ValidationFailed(map[string]string{"message": "is for another network"})
	ErrMessageOutOfTime       = apierror.ValidationFailed(map[string]string{"message": "is not valid at this time"})
	ErrChallengeUnavailable   = apierror.ValidationFailed(map[string]string{"message": "has expired or was already used. Request a new one."})
	ErrSignatureMismatch      = apierror.ValidationFailed(map[string]string{"signature": "does not match the wallet"})
	ErrAddressLinkedElsewhere = apierror.ValidationFailed(map[string]string{"address": "is linked to another account"})
	ErrAnotherWalletLinked    = apierror.Conflict("Your account already has a linked wallet")
)

type Clock func() time.Time

type TransactionRunner interface {
	WithinTransaction(ctx context.Context, work func(ctx context.Context) error) error
}

type IssuedChallenge struct {
	Nonce     string
	ExpiresAt time.Time
}

type LinkedWallet struct {
	Wallet  models.Wallet
	IsNewly bool
}

type WalletCursor struct {
	LinkedAt time.Time `json:"linked_at"`
	ID       uuid.UUID `json:"id"`
}

type WalletLinkService interface {
	IssueChallenge(ctx context.Context, userID uuid.UUID, address string, chainID int64) (IssuedChallenge, error)
	Link(ctx context.Context, userID uuid.UUID, messageText string, signatureHex string) (LinkedWallet, error)
	List(ctx context.Context, userID uuid.UUID, after *WalletCursor, pageRequest pagination.Request) (pagination.Page[models.Wallet], error)
}

type WalletLinkDependencies struct {
	Challenges      repository.WalletChallengeRepository
	Wallets         repository.WalletRepository
	Transactions    TransactionRunner
	ChainID         int64
	FrontendBaseURL string
	Now             Clock
}

type walletLinkService struct {
	dependencies   WalletLinkDependencies
	expectedDomain string
	expectedOrigin string
}

func NewWalletLinkService(dependencies WalletLinkDependencies) (WalletLinkService, error) {
	frontendURL, err := url.Parse(dependencies.FrontendBaseURL)
	if err != nil || frontendURL.Host == "" {
		return nil, errors.New("wallet linking needs a valid frontend base URL")
	}
	return &walletLinkService{
		dependencies:   dependencies,
		expectedDomain: frontendURL.Host,
		expectedOrigin: frontendURL.Scheme + "://" + frontendURL.Host,
	}, nil
}

func (links *walletLinkService) IssueChallenge(ctx context.Context, userID uuid.UUID, address string, chainID int64) (IssuedChallenge, error) {
	if chainID != links.dependencies.ChainID {
		return IssuedChallenge{}, ErrUnsupportedChain
	}
	challengeID, err := uuid.NewV7()
	if err != nil {
		return IssuedChallenge{}, err
	}
	nonce, err := randomNonce()
	if err != nil {
		return IssuedChallenge{}, err
	}
	issuedAt := links.dependencies.Now().UTC()
	challenge := models.WalletChallenge{
		ID:        challengeID,
		UserID:    userID,
		Address:   strings.ToLower(address),
		ChainID:   chainID,
		Nonce:     nonce,
		ExpiresAt: issuedAt.Add(challengeLifetime),
		CreatedAt: issuedAt,
	}
	if err := links.dependencies.Challenges.Create(ctx, &challenge); err != nil {
		return IssuedChallenge{}, err
	}
	return IssuedChallenge{Nonce: nonce, ExpiresAt: challenge.ExpiresAt}, nil
}

func (links *walletLinkService) Link(ctx context.Context, userID uuid.UUID, messageText string, signatureHex string) (LinkedWallet, error) {
	message, err := ParseSignInMessage(messageText)
	if err != nil {
		return LinkedWallet{}, ErrMalformedMessage
	}
	now := links.dependencies.Now().UTC()
	if err := links.checkMessage(message, now); err != nil {
		return LinkedWallet{}, err
	}
	signerAddress, err := RecoverPersonalSigner(messageText, signatureHex)
	if err != nil || signerAddress != strings.ToLower(message.Address) {
		return LinkedWallet{}, ErrSignatureMismatch
	}

	var linkedWallet LinkedWallet
	err = links.dependencies.Transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		isClaimed, err := links.dependencies.Challenges.Consume(ctx, repository.ChallengeClaim{
			UserID:  userID,
			Address: signerAddress,
			ChainID: message.ChainID,
			Nonce:   message.Nonce,
			UsedAt:  now,
		})
		if err != nil {
			return err
		}
		if !isClaimed {
			return ErrChallengeUnavailable
		}
		linkedWallet, err = links.saveWallet(ctx, userID, signerAddress, message.ChainID, now)
		return err
	})
	return linkedWallet, err
}

func (links *walletLinkService) checkMessage(message SignInMessage, now time.Time) error {
	messageURL, err := url.Parse(message.URI)
	if err != nil || message.Domain != links.expectedDomain || messageURL.Scheme+"://"+messageURL.Host != links.expectedOrigin {
		return ErrMessageForAnotherSite
	}
	if message.Version != supportedMessageVersion {
		return ErrMalformedMessage
	}
	if message.ChainID != links.dependencies.ChainID {
		return ErrMessageForAnotherChain
	}
	isIssuedInFuture := message.IssuedAt.After(now.Add(allowedClockSkew))
	isExpired := message.ExpirationTime != nil && !message.ExpirationTime.After(now)
	isNotYetValid := message.NotBefore != nil && message.NotBefore.After(now.Add(allowedClockSkew))
	if isIssuedInFuture || isExpired || isNotYetValid {
		return ErrMessageOutOfTime
	}
	return nil
}

func (links *walletLinkService) saveWallet(ctx context.Context, userID uuid.UUID, address string, chainID int64, linkedAt time.Time) (LinkedWallet, error) {
	existingWallet, err := links.dependencies.Wallets.FindByUser(ctx, userID)
	if err == nil {
		if existingWallet.Address == address {
			return LinkedWallet{Wallet: existingWallet}, nil
		}
		return LinkedWallet{}, ErrAnotherWalletLinked
	}
	if !errors.Is(err, repository.ErrWalletNotFound) {
		return LinkedWallet{}, err
	}
	walletID, err := uuid.NewV7()
	if err != nil {
		return LinkedWallet{}, err
	}
	wallet := models.Wallet{ID: walletID, UserID: userID, Address: address, ChainID: chainID, LinkedAt: linkedAt, CreatedAt: linkedAt, UpdatedAt: linkedAt}
	switch err := links.dependencies.Wallets.Create(ctx, &wallet); {
	case errors.Is(err, repository.ErrAddressLinkedElsewhere):
		return LinkedWallet{}, ErrAddressLinkedElsewhere
	case errors.Is(err, repository.ErrUserAlreadyHasWallet):
		return LinkedWallet{}, ErrAnotherWalletLinked
	case err != nil:
		return LinkedWallet{}, err
	}
	return LinkedWallet{Wallet: wallet, IsNewly: true}, nil
}

func (links *walletLinkService) List(ctx context.Context, userID uuid.UUID, after *WalletCursor, pageRequest pagination.Request) (pagination.Page[models.Wallet], error) {
	var afterPosition *repository.WalletPosition
	if after != nil {
		afterPosition = &repository.WalletPosition{LinkedAt: after.LinkedAt, ID: after.ID}
	}
	wallets, err := links.dependencies.Wallets.ListForUser(ctx, userID, afterPosition, pageRequest.FetchLimit())
	if err != nil {
		return pagination.Page[models.Wallet]{}, err
	}
	return pagination.BuildPage(wallets, pageRequest, func(wallet models.Wallet) WalletCursor {
		return WalletCursor{LinkedAt: wallet.LinkedAt, ID: wallet.ID}
	})
}

func randomNonce() (string, error) {
	nonceBytes := make([]byte, nonceByteLength)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(nonceBytes), nil
}
