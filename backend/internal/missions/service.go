package missions

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
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/missions/missionsstore"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/pagination"
)

const fuelCellItemID = 401

type StartRequest struct {
	ZoneID      string
	ShipItemID  int
	DrillItemID int
}

type Service struct {
	pool    *pgxpool.Pool
	catalog *catalog.Service
	now     func() time.Time
}

func NewService(pool *pgxpool.Pool, catalogService *catalog.Service, now func() time.Time) *Service {
	return &Service{pool: pool, catalog: catalogService, now: now}
}

func (service *Service) Start(ctx context.Context, userID uuid.UUID, startRequest StartRequest) (Mission, error) {
	catalogSnapshot, err := service.catalog.Snapshot(ctx)
	if err != nil {
		return Mission{}, err
	}
	zoneSnapshot, speedBasisPoints, err := buildZoneSnapshot(catalogSnapshot, startRequest)
	if err != nil {
		return Mission{}, err
	}
	missionID, err := uuid.NewV7()
	if err != nil {
		return Mission{}, err
	}
	encodedSnapshot, err := json.Marshal(zoneSnapshot)
	if err != nil {
		return Mission{}, err
	}
	missionReference := ledger.Reference{Type: "mission", ID: missionID.String()}
	startedAt := service.now()

	var insertedMission missionsstore.Mission
	err = pgx.BeginFunc(ctx, service.pool, func(tx pgx.Tx) error {
		queries := missionsstore.New(tx)
		if err := queries.LockPlayerMissions(ctx, userID.String()); err != nil {
			return fmt.Errorf("lock player missions: %w", err)
		}
		activeMissionCount, err := queries.CountActiveMissions(ctx, userID)
		if err != nil {
			return fmt.Errorf("count active missions: %w", err)
		}
		if activeMissionCount >= MaximumActiveMissions {
			return apierror.New(http.StatusUnprocessableEntity, apierror.CodeLimitExceeded,
				fmt.Sprintf("You can run up to %d missions at once. Collect one first", MaximumActiveMissions))
		}

		shipHoldID, err := ledger.Hold(ctx, tx, ledger.PlayerItem(userID, startRequest.ShipItemID), 1, missionReference)
		if err != nil {
			return err
		}
		drillHoldID, err := ledger.Hold(ctx, tx, ledger.PlayerItem(userID, startRequest.DrillItemID), 1, missionReference)
		if err != nil {
			return err
		}
		if zoneSnapshot.FuelCost > 0 {
			if _, err := ledger.Post(ctx, tx, ledger.Journal{
				Type:      ledger.JournalMissionFuel,
				Reference: missionReference,
				Legs: []ledger.Leg{
					ledger.Debit(ledger.PlayerItem(userID, fuelCellItemID), int64(zoneSnapshot.FuelCost)),
					ledger.Credit(ledger.System(ledger.SystemBurn, ledger.Item(fuelCellItemID)), int64(zoneSnapshot.FuelCost)),
				},
			}); err != nil {
				return err
			}
		}

		insertedMission, err = queries.InsertMission(ctx, missionsstore.InsertMissionParams{
			ID:           missionID,
			UserID:       userID,
			ZoneID:       zoneSnapshot.ZoneID,
			ShipItemID:   int32(startRequest.ShipItemID),
			DrillItemID:  int32(startRequest.DrillItemID),
			FuelSpent:    int32(zoneSnapshot.FuelCost),
			ShipHoldID:   shipHoldID,
			DrillHoldID:  drillHoldID,
			ZoneSnapshot: encodedSnapshot,
			StartedAt:    startedAt,
			EndsAt:       startedAt.Add(time.Duration(MissionDurationSeconds(zoneSnapshot.DurationSeconds, speedBasisPoints)) * time.Second),
		})
		return err
	})
	if err != nil {
		return Mission{}, err
	}
	return missionFromRow(insertedMission)
}

func buildZoneSnapshot(catalogSnapshot catalog.Snapshot, startRequest StartRequest) (ZoneSnapshot, int, error) {
	zone, isKnownZone := catalogSnapshot.ZoneByID(startRequest.ZoneID)
	if !isKnownZone {
		return ZoneSnapshot{}, 0, apierror.ValidationFailed(map[string]string{"zone_id": "is not a known zone"})
	}
	shipItem, _ := catalogSnapshot.ItemByID(startRequest.ShipItemID)
	shipStats, isShip := shipItem.ShipStats()
	if !isShip {
		return ZoneSnapshot{}, 0, apierror.ValidationFailed(map[string]string{"ship_item_id": "is not a ship"})
	}
	drillItem, _ := catalogSnapshot.ItemByID(startRequest.DrillItemID)
	drillStats, isDrill := drillItem.DrillStats()
	if !isDrill {
		return ZoneSnapshot{}, 0, apierror.ValidationFailed(map[string]string{"drill_item_id": "is not a drill"})
	}
	if drillStats.Tier < zone.MinimumDrillTier {
		return ZoneSnapshot{}, 0, apierror.ValidationFailed(map[string]string{
			"drill_item_id": fmt.Sprintf("%s needs a Tier %d drill or better", zone.Name, zone.MinimumDrillTier),
		})
	}
	if !zone.AllowsShip(startRequest.ShipItemID) {
		return ZoneSnapshot{}, 0, apierror.ValidationFailed(map[string]string{
			"ship_item_id": fmt.Sprintf("%s can't fly to %s", shipItem.Name, zone.Name),
		})
	}

	rarityRankByItem := map[int]int{}
	for _, lootEntry := range zone.Loot {
		lootItem, _ := catalogSnapshot.ItemByID(lootEntry.ItemID)
		rarityRankByItem[lootEntry.ItemID] = lootItem.RarityRank
	}
	return ZoneSnapshot{
		ZoneID:                     zone.ID,
		ZoneName:                   zone.Name,
		DurationSeconds:            zone.DurationSeconds,
		FuelCost:                   zone.FuelCost,
		Loot:                       zone.Loot,
		RarityRankByItem:           rarityRankByItem,
		DrillMultiplierBasisPoints: drillStats.YieldMultiplierBasisPoints,
		CargoCapacity:              shipStats.CargoCapacity,
	}, shipStats.SpeedMultiplierBasisPoints, nil
}

func (service *Service) Collect(ctx context.Context, userID, missionID uuid.UUID) (Mission, error) {
	var collectedMission missionsstore.Mission
	err := pgx.BeginFunc(ctx, service.pool, func(tx pgx.Tx) error {
		var err error
		collectedAt := service.now()
		collectedMission, err = missionsstore.New(tx).CollectMission(ctx, missionsstore.CollectMissionParams{
			ID: missionID, UserID: userID, CollectedAt: &collectedAt,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return service.explainUnavailableMission(ctx, tx, userID, missionID, "collected")
		}
		if err != nil {
			return fmt.Errorf("collect mission: %w", err)
		}
		var loot []LootRoll
		if err := json.Unmarshal(collectedMission.Loot, &loot); err != nil {
			return fmt.Errorf("decode loot: %w", err)
		}
		if err := mintLoot(ctx, tx, userID, missionID, loot); err != nil {
			return err
		}
		return releaseMissionHolds(ctx, tx, collectedMission)
	})
	if err != nil {
		return Mission{}, err
	}
	return missionFromRow(collectedMission)
}

func (service *Service) Abort(ctx context.Context, userID, missionID uuid.UUID) (Mission, error) {
	var abortedMission missionsstore.Mission
	err := pgx.BeginFunc(ctx, service.pool, func(tx pgx.Tx) error {
		var err error
		abortedAt := service.now()
		abortedMission, err = missionsstore.New(tx).AbortMission(ctx, missionsstore.AbortMissionParams{
			ID: missionID, UserID: userID, AbortedAt: &abortedAt,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return service.explainUnavailableMission(ctx, tx, userID, missionID, "aborted")
		}
		if err != nil {
			return fmt.Errorf("abort mission: %w", err)
		}
		return releaseMissionHolds(ctx, tx, abortedMission)
	})
	if err != nil {
		return Mission{}, err
	}
	return missionFromRow(abortedMission)
}

func (service *Service) explainUnavailableMission(ctx context.Context, tx pgx.Tx, userID, missionID uuid.UUID, action string) error {
	existingMission, err := missionsstore.New(tx).GetMissionForUser(ctx, missionsstore.GetMissionForUserParams{ID: missionID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return apierror.NotFound("This mission does not exist")
	}
	if err != nil {
		return fmt.Errorf("load mission: %w", err)
	}
	return apierror.Conflict(fmt.Sprintf("This mission can't be %s while it is %s", action, existingMission.Status)).
		WithDetails(map[string]string{"status": existingMission.Status})
}

func mintLoot(ctx context.Context, tx pgx.Tx, userID, missionID uuid.UUID, loot []LootRoll) error {
	legs := make([]ledger.Leg, 0, len(loot)*2)
	for _, roll := range loot {
		legs = append(legs,
			ledger.Debit(ledger.System(ledger.SystemMint, ledger.Item(roll.ItemID)), int64(roll.Quantity)),
			ledger.Credit(ledger.PlayerItem(userID, roll.ItemID), int64(roll.Quantity)),
		)
	}
	if len(legs) == 0 {
		return nil
	}
	_, err := ledger.Post(ctx, tx, ledger.Journal{
		Type:      ledger.JournalMissionLoot,
		Reference: ledger.Reference{Type: "mission", ID: missionID.String()},
		Legs:      legs,
	})
	return err
}

func releaseMissionHolds(ctx context.Context, tx pgx.Tx, mission missionsstore.Mission) error {
	for _, holdID := range []uuid.UUID{mission.ShipHoldID, mission.DrillHoldID} {
		if _, err := ledger.ReleaseRemaining(ctx, tx, holdID); err != nil {
			return fmt.Errorf("release mission hold: %w", err)
		}
	}
	return nil
}

func (service *Service) Get(ctx context.Context, userID, missionID uuid.UUID) (Mission, error) {
	missionRow, err := missionsstore.New(service.pool).GetMissionForUser(ctx, missionsstore.GetMissionForUserParams{ID: missionID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Mission{}, apierror.NotFound("This mission does not exist")
	}
	if err != nil {
		return Mission{}, fmt.Errorf("load mission: %w", err)
	}
	return missionFromRow(missionRow)
}

type ListCursor struct {
	BeforeID uuid.UUID `json:"before_id"`
}

func (service *Service) List(ctx context.Context, userID uuid.UUID, statuses []string, pageRequest pagination.Request, cursor *ListCursor) (pagination.Page[Mission], error) {
	parameters := missionsstore.ListMissionsForUserParams{UserID: userID, Statuses: statuses, RowLimit: int32(pageRequest.FetchLimit())}
	if parameters.Statuses == nil {
		parameters.Statuses = []string{}
	}
	if cursor != nil {
		parameters.BeforeID = &cursor.BeforeID
	}
	missionRows, err := missionsstore.New(service.pool).ListMissionsForUser(ctx, parameters)
	if err != nil {
		return pagination.Page[Mission]{}, fmt.Errorf("list missions: %w", err)
	}
	missionList := make([]Mission, 0, len(missionRows))
	for _, missionRow := range missionRows {
		mission, err := missionFromRow(missionRow)
		if err != nil {
			return pagination.Page[Mission]{}, err
		}
		missionList = append(missionList, mission)
	}
	return pagination.BuildPage(missionList, pageRequest, func(mission Mission) ListCursor {
		return ListCursor{BeforeID: mission.ID}
	})
}
