package missions

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/missions/missionsstore"
)

const MaximumActiveMissions = 3

type Status string

const (
	StatusRunning   Status = "RUNNING"
	StatusCompleted Status = "COMPLETED"
	StatusCollected Status = "COLLECTED"
	StatusAborted   Status = "ABORTED"
)

type ZoneSnapshot struct {
	ZoneID                     string              `json:"zone_id"`
	ZoneName                   string              `json:"zone_name"`
	DurationSeconds            int                 `json:"duration_seconds"`
	FuelCost                   int                 `json:"fuel_cost"`
	Loot                       []catalog.LootEntry `json:"loot"`
	RarityRankByItem           map[int]int         `json:"rarity_rank_by_item"`
	DrillMultiplierBasisPoints int                 `json:"drill_multiplier_basis_points"`
	CargoCapacity              int                 `json:"cargo_capacity"`
}

func (snapshot ZoneSnapshot) lootRules() LootRules {
	return LootRules{
		Loot:                       snapshot.Loot,
		RarityRankByItem:           snapshot.RarityRankByItem,
		DrillMultiplierBasisPoints: snapshot.DrillMultiplierBasisPoints,
		CargoCapacity:              snapshot.CargoCapacity,
	}
}

type ExpectedLoot struct {
	ItemID            int `json:"item_id"`
	MinimumQuantity   int `json:"minimum_quantity"`
	MaximumQuantity   int `json:"maximum_quantity"`
	ChanceBasisPoints int `json:"chance_basis_points"`
}

type Mission struct {
	ID            uuid.UUID      `json:"id"`
	ZoneID        string         `json:"zone_id"`
	ZoneName      string         `json:"zone_name"`
	ShipItemID    int            `json:"ship_item_id"`
	DrillItemID   int            `json:"drill_item_id"`
	Status        Status         `json:"status"`
	FuelSpent     int            `json:"fuel_spent"`
	CargoCapacity int            `json:"cargo_capacity"`
	ExpectedLoot  []ExpectedLoot `json:"expected_loot"`
	Loot          []LootRoll     `json:"loot"`
	StartedAt     time.Time      `json:"started_at"`
	EndsAt        time.Time      `json:"ends_at"`
	ResolvedAt    *time.Time     `json:"resolved_at"`
	CollectedAt   *time.Time     `json:"collected_at"`
	AbortedAt     *time.Time     `json:"aborted_at"`
}

func expectedLootFor(snapshot ZoneSnapshot) []ExpectedLoot {
	expected := make([]ExpectedLoot, 0, len(snapshot.Loot))
	for _, lootEntry := range snapshot.Loot {
		minimumQuantity, maximumQuantity := lootEntry.MinimumQuantity, lootEntry.MaximumQuantity
		if lootEntry.ChanceBasisPoints >= catalog.BasisPointsPerWhole {
			minimumQuantity = minimumQuantity * snapshot.DrillMultiplierBasisPoints / catalog.BasisPointsPerWhole
			maximumQuantity = maximumQuantity * snapshot.DrillMultiplierBasisPoints / catalog.BasisPointsPerWhole
		}
		expected = append(expected, ExpectedLoot{
			ItemID:            lootEntry.ItemID,
			MinimumQuantity:   minimumQuantity,
			MaximumQuantity:   maximumQuantity,
			ChanceBasisPoints: lootEntry.ChanceBasisPoints,
		})
	}
	return expected
}

func missionFromRow(missionRow missionsstore.Mission) (Mission, error) {
	var snapshot ZoneSnapshot
	if err := json.Unmarshal(missionRow.ZoneSnapshot, &snapshot); err != nil {
		return Mission{}, err
	}
	loot := []LootRoll{}
	if missionRow.Loot != nil {
		if err := json.Unmarshal(missionRow.Loot, &loot); err != nil {
			return Mission{}, err
		}
	}
	return Mission{
		ID:            missionRow.ID,
		ZoneID:        missionRow.ZoneID,
		ZoneName:      snapshot.ZoneName,
		ShipItemID:    int(missionRow.ShipItemID),
		DrillItemID:   int(missionRow.DrillItemID),
		Status:        Status(missionRow.Status),
		FuelSpent:     int(missionRow.FuelSpent),
		CargoCapacity: snapshot.CargoCapacity,
		ExpectedLoot:  expectedLootFor(snapshot),
		Loot:          loot,
		StartedAt:     missionRow.StartedAt,
		EndsAt:        missionRow.EndsAt,
		ResolvedAt:    missionRow.ResolvedAt,
		CollectedAt:   missionRow.CollectedAt,
		AbortedAt:     missionRow.AbortedAt,
	}, nil
}
