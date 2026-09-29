package ledger

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/ledger/ledgerstore"
)

type Asset int32

const NC Asset = 0

func Item(itemID int) Asset {
	return Asset(itemID)
}

func (asset Asset) IsNC() bool {
	return asset == NC
}

type Bucket string

const (
	BucketCard          Bucket = "card"
	BucketCrypto        Bucket = "crypto"
	BucketEarnedPending Bucket = "earned_pending"
	BucketEarned        Bucket = "earned"
)

var SpendingOrder = []Bucket{BucketCard, BucketEarnedPending, BucketEarned, BucketCrypto}

type SystemAccount string

const (
	SystemStripeClearing SystemAccount = "stripe_clearing"
	SystemCryptoClearing SystemAccount = "crypto_clearing"
	SystemFees           SystemAccount = "fees"
	SystemTreasury       SystemAccount = "treasury"
	SystemMint           SystemAccount = "mint"
	SystemBurn           SystemAccount = "burn"
	SystemWithdrawn      SystemAccount = "withdrawn"
	SystemUnclaimed      SystemAccount = "unclaimed"
)

type negativeAvailablePolicy string

const (
	negativeNever       negativeAvailablePolicy = "never"
	negativeDisputeOnly negativeAvailablePolicy = "dispute_only"
	negativeAlways      negativeAvailablePolicy = "always"
)

type AccountKey struct {
	UserID uuid.UUID
	System SystemAccount
	Asset  Asset
	Bucket Bucket
}

func PlayerNC(userID uuid.UUID, bucket Bucket) AccountKey {
	return AccountKey{UserID: userID, Asset: NC, Bucket: bucket}
}

func PlayerItem(userID uuid.UUID, itemID int) AccountKey {
	return AccountKey{UserID: userID, Asset: Item(itemID)}
}

func System(systemAccount SystemAccount, asset Asset) AccountKey {
	return AccountKey{System: systemAccount, Asset: asset}
}

func (key AccountKey) isPlayer() bool {
	return key.UserID != uuid.Nil
}

func (key AccountKey) validate() error {
	if key.isPlayer() == (key.System != "") {
		return fmt.Errorf("account key needs exactly one owner: %+v", key)
	}
	if key.isPlayer() && key.Asset.IsNC() && key.Bucket == "" {
		return fmt.Errorf("player NC accounts need a bucket: %+v", key)
	}
	if (!key.isPlayer() || !key.Asset.IsNC()) && key.Bucket != "" {
		return fmt.Errorf("only player NC accounts have a bucket: %+v", key)
	}
	return nil
}

func (key AccountKey) negativePolicy() negativeAvailablePolicy {
	switch {
	case key.System == SystemStripeClearing, key.System == SystemCryptoClearing, key.System == SystemMint:
		return negativeAlways
	case key.isPlayer() && key.Bucket == BucketCard:
		return negativeDisputeOnly
	default:
		return negativeNever
	}
}

func (key AccountKey) storeParameters() ledgerstore.FindAccountParams {
	parameters := ledgerstore.FindAccountParams{}
	if key.isPlayer() {
		userID := key.UserID
		parameters.UserID = &userID
	} else {
		systemAccount := string(key.System)
		parameters.SystemAccount = &systemAccount
	}
	if !key.Asset.IsNC() {
		itemID := int32(key.Asset)
		parameters.ItemID = &itemID
	}
	if key.Bucket != "" {
		bucket := string(key.Bucket)
		parameters.Bucket = &bucket
	}
	return parameters
}

func resolveAccount(ctx context.Context, queries *ledgerstore.Queries, key AccountKey) (int64, error) {
	if err := key.validate(); err != nil {
		return 0, err
	}
	lookup := key.storeParameters()
	accountID, err := queries.FindAccount(ctx, lookup)
	if err == nil {
		return accountID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("find ledger account: %w", err)
	}

	accountID, err = queries.InsertAccount(ctx, ledgerstore.InsertAccountParams{
		UserID:                  lookup.UserID,
		SystemAccount:           lookup.SystemAccount,
		ItemID:                  lookup.ItemID,
		Bucket:                  lookup.Bucket,
		NegativeAvailablePolicy: string(key.negativePolicy()),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		accountID, err = queries.FindAccount(ctx, lookup)
	}
	if err != nil {
		return 0, fmt.Errorf("create ledger account: %w", err)
	}
	return accountID, nil
}
