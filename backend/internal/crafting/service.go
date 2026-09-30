package crafting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/crafting/craftingstore"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/notify/inapp"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/money"
	"nebula-exchange/backend/internal/platform/pagination"
)

type Service struct {
	pool     *pgxpool.Pool
	catalog  *catalog.Service
	notifier *inapp.Notifier
	now      func() time.Time
}

func NewService(pool *pgxpool.Pool, catalogService *catalog.Service, notifier *inapp.Notifier, now func() time.Time) *Service {
	return &Service{pool: pool, catalog: catalogService, notifier: notifier, now: now}
}

func (service *Service) Start(ctx context.Context, userID uuid.UUID, recipeID string, quantity int) (CraftJob, error) {
	if quantity < 1 || quantity > MaximumBatchQuantity {
		return CraftJob{}, apierror.ValidationFailed(map[string]string{"quantity": fmt.Sprintf("must be from 1 to %d", MaximumBatchQuantity)})
	}
	catalogSnapshot, err := service.catalog.Snapshot(ctx)
	if err != nil {
		return CraftJob{}, err
	}
	recipe, isKnownRecipe := catalogSnapshot.RecipeByID(recipeID)
	if !isKnownRecipe {
		return CraftJob{}, apierror.ValidationFailed(map[string]string{"recipe_id": "is not a known recipe"})
	}

	craftJobID, err := uuid.NewV7()
	if err != nil {
		return CraftJob{}, err
	}
	totalInputs := scaledQuantities(recipe.Inputs, quantity)
	encodedInputs, err := json.Marshal(totalInputs)
	if err != nil {
		return CraftJob{}, err
	}
	totalFee := recipe.Fee * money.Micro(quantity)
	startedAt := service.now()

	var insertedJob craftingstore.CraftJob
	err = pgx.BeginFunc(ctx, service.pool, func(tx pgx.Tx) error {
		var err error
		insertedJob, err = craftingstore.New(tx).InsertCraftJob(ctx, craftingstore.InsertCraftJobParams{
			ID:             craftJobID,
			UserID:         userID,
			RecipeID:       recipe.ID,
			Quantity:       int32(quantity),
			OutputItemID:   int32(recipe.OutputItemID),
			OutputQuantity: int32(recipe.OutputQuantity * quantity),
			FeeMicro:       int64(totalFee),
			Inputs:         encodedInputs,
			StartedAt:      startedAt,
			EndsAt:         startedAt.Add(time.Duration(recipe.CraftSeconds*quantity) * time.Second),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.New(http.StatusUnprocessableEntity, apierror.CodeLimitExceeded, "Your workshop is busy. Wait for the current craft to finish")
		}
		if err != nil {
			return fmt.Errorf("insert craft job: %w", err)
		}

		legs := []ledger.Leg{}
		for _, input := range totalInputs {
			legs = append(legs, ledger.BurnFromPlayer(userID, input.ItemID, int64(input.Quantity))...)
		}
		if totalFee > 0 {
			feeLegs, err := ledger.SpendLegs(ctx, tx, userID, int64(totalFee))
			if err != nil {
				return err
			}
			legs = append(legs, feeLegs...)
			legs = append(legs, ledger.Credit(ledger.System(ledger.SystemBurn, ledger.NC), int64(totalFee)))
		}
		_, err = ledger.Post(ctx, tx, ledger.Journal{
			Type:      ledger.JournalCraftStart,
			Reference: ledger.Reference{Type: "craft", ID: craftJobID.String()},
			Metadata:  map[string]any{"recipe_id": recipe.ID, "quantity": quantity},
			Legs:      legs,
		})
		return err
	})
	if err != nil {
		return CraftJob{}, err
	}
	return craftJobFromRow(insertedJob)
}

func (service *Service) Deliver(ctx context.Context, craftJobID uuid.UUID) (bool, error) {
	isDelivered := false
	err := pgx.BeginFunc(ctx, service.pool, func(tx pgx.Tx) error {
		deliveredAt := service.now()
		deliveredJob, err := craftingstore.New(tx).DeliverCraftJob(ctx, craftingstore.DeliverCraftJobParams{ID: craftJobID, DeliveredAt: &deliveredAt})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("deliver craft job: %w", err)
		}
		if _, err := ledger.Post(ctx, tx, ledger.Journal{
			Type:      ledger.JournalCraftOutput,
			Reference: ledger.Reference{Type: "craft", ID: craftJobID.String()},
			Legs:      ledger.MintToPlayer(deliveredJob.UserID, int(deliveredJob.OutputItemID), int64(deliveredJob.OutputQuantity)),
		}); err != nil {
			return err
		}
		isDelivered = true
		return service.notifyDelivery(ctx, tx, deliveredJob)
	})
	return isDelivered, err
}

type ListCursor struct {
	BeforeID uuid.UUID `json:"before_id"`
}

func (service *Service) List(ctx context.Context, userID uuid.UUID, status string, pageRequest pagination.Request, cursor *ListCursor) (pagination.Page[CraftJob], error) {
	parameters := craftingstore.ListCraftJobsForUserParams{UserID: userID, RowLimit: int32(pageRequest.FetchLimit())}
	if status != "" {
		parameters.Status = &status
	}
	if cursor != nil {
		parameters.BeforeID = &cursor.BeforeID
	}
	craftJobRows, err := craftingstore.New(service.pool).ListCraftJobsForUser(ctx, parameters)
	if err != nil {
		return pagination.Page[CraftJob]{}, fmt.Errorf("list craft jobs: %w", err)
	}
	craftJobs := make([]CraftJob, 0, len(craftJobRows))
	for _, craftJobRow := range craftJobRows {
		craftJob, err := craftJobFromRow(craftJobRow)
		if err != nil {
			return pagination.Page[CraftJob]{}, err
		}
		craftJobs = append(craftJobs, craftJob)
	}
	return pagination.BuildPage(craftJobs, pageRequest, func(craftJob CraftJob) ListCursor {
		return ListCursor{BeforeID: craftJob.ID}
	})
}

func (service *Service) notifyDelivery(ctx context.Context, tx pgx.Tx, deliveredJob craftingstore.CraftJob) error {
	if service.notifier == nil {
		return nil
	}
	catalogSnapshot, err := service.catalog.Snapshot(ctx)
	if err != nil {
		return err
	}
	outputItem, _ := catalogSnapshot.ItemByID(int(deliveredJob.OutputItemID))
	return service.notifier.Notify(ctx, tx, inapp.Notice{
		UserID: deliveredJob.UserID,
		Kind:   inapp.KindCraftCompleted,
		Title:  fmt.Sprintf("%d × %s crafted", deliveredJob.OutputQuantity, outputItem.Name),
		Body:   "Your craft is finished and in your inventory. The workshop is free again.",
		Link:   "/workshop",
	})
}
