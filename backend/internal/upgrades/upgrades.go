package upgrades

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/money"
)

type Path string

const (
	PathCraft Path = "craft"
	PathBuy   Path = "buy"
)

type CompletedUpgrade struct {
	ID         uuid.UUID   `json:"id"`
	UpgradeID  string      `json:"upgrade_id"`
	Path       Path        `json:"path"`
	FromItemID int         `json:"from_item_id"`
	ToItemID   int         `json:"to_item_id"`
	Paid       money.Micro `json:"paid"`
	JournalID  uuid.UUID   `json:"journal_id"`
}

var ErrNotForSale = apierror.New(http.StatusUnprocessableEntity, apierror.CodeValidationFailed, "This upgrade can only be crafted").
	WithDetails(map[string]string{"upgrade_id": "has no buy path"})

func swapLegs(userID uuid.UUID, upgrade catalog.Upgrade) []ledger.Leg {
	legs := ledger.BurnFromPlayer(userID, upgrade.FromItemID, 1)
	return append(legs, ledger.MintToPlayer(userID, upgrade.ToItemID, 1)...)
}

func ApplyCraft(ctx context.Context, tx pgx.Tx, userID uuid.UUID, upgrade catalog.Upgrade, reference ledger.Reference) (uuid.UUID, error) {
	legs := swapLegs(userID, upgrade)
	for _, input := range upgrade.Inputs {
		legs = append(legs, ledger.BurnFromPlayer(userID, input.ItemID, int64(input.Quantity))...)
	}
	if upgrade.CraftFee > 0 {
		feeLegs, err := ledger.SpendLegs(ctx, tx, userID, int64(upgrade.CraftFee))
		if err != nil {
			return uuid.Nil, err
		}
		legs = append(legs, feeLegs...)
		legs = append(legs, ledger.Credit(ledger.System(ledger.SystemBurn, ledger.NC), int64(upgrade.CraftFee)))
	}
	return ledger.Post(ctx, tx, ledger.Journal{
		Type:      ledger.JournalUpgrade,
		Reference: reference,
		Metadata:  map[string]any{"upgrade_id": upgrade.ID, "path": string(PathCraft)},
		Legs:      legs,
	})
}

func ApplyPurchase(ctx context.Context, tx pgx.Tx, userID uuid.UUID, upgrade catalog.Upgrade, reference ledger.Reference) (uuid.UUID, error) {
	if upgrade.BuyPrice == nil {
		return uuid.Nil, ErrNotForSale
	}
	priceLegs, err := ledger.SpendLegs(ctx, tx, userID, int64(*upgrade.BuyPrice))
	if err != nil {
		return uuid.Nil, err
	}
	legs := append(swapLegs(userID, upgrade), priceLegs...)
	legs = append(legs, ledger.Credit(ledger.System(ledger.SystemTreasury, ledger.NC), int64(*upgrade.BuyPrice)))
	return ledger.Post(ctx, tx, ledger.Journal{
		Type:      ledger.JournalUpgrade,
		Reference: reference,
		Metadata:  map[string]any{"upgrade_id": upgrade.ID, "path": string(PathBuy)},
		Legs:      legs,
	})
}

type Service struct {
	pool    *pgxpool.Pool
	catalog *catalog.Service
}

func NewService(pool *pgxpool.Pool, catalogService *catalog.Service) *Service {
	return &Service{pool: pool, catalog: catalogService}
}

func (service *Service) Perform(ctx context.Context, userID uuid.UUID, upgradeID string, path Path) (CompletedUpgrade, error) {
	catalogSnapshot, err := service.catalog.Snapshot(ctx)
	if err != nil {
		return CompletedUpgrade{}, err
	}
	upgrade, isKnownUpgrade := catalogSnapshot.UpgradeByID(upgradeID)
	if !isKnownUpgrade {
		return CompletedUpgrade{}, apierror.NotFound("This upgrade does not exist")
	}
	upgradeRunID, err := uuid.NewV7()
	if err != nil {
		return CompletedUpgrade{}, err
	}
	completed := CompletedUpgrade{ID: upgradeRunID, UpgradeID: upgrade.ID, Path: path, FromItemID: upgrade.FromItemID, ToItemID: upgrade.ToItemID}
	err = pgx.BeginFunc(ctx, service.pool, func(tx pgx.Tx) error {
		var applyErr error
		if path == PathCraft {
			completed.Paid = upgrade.CraftFee
			completed.JournalID, applyErr = ApplyCraft(ctx, tx, userID, upgrade, ledger.Reference{Type: "upgrade_craft", ID: upgradeRunID.String()})
		} else {
			if upgrade.BuyPrice != nil {
				completed.Paid = *upgrade.BuyPrice
			}
			completed.JournalID, applyErr = ApplyPurchase(ctx, tx, userID, upgrade, ledger.Reference{Type: "upgrade_purchase", ID: upgradeRunID.String()})
		}
		return applyErr
	})
	if err != nil {
		return CompletedUpgrade{}, fmt.Errorf("upgrade %s: %w", upgradeID, err)
	}
	return completed, nil
}
