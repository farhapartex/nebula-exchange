package catalog

import (
	"context"

	"github.com/google/uuid"
)

type ZoneUnlock struct {
	IsUnlocked       bool `json:"is_unlocked"`
	HasRequiredDrill bool `json:"has_required_drill"`
	HasAllowedShip   bool `json:"has_allowed_ship"`
	HasEnoughFuel    bool `json:"has_enough_fuel"`
}

type PlayerZone struct {
	Zone
	Unlock *ZoneUnlock `json:"unlock"`
}

type ZoneAvailability interface {
	ForPlayer(ctx context.Context, userID uuid.UUID, zones []Zone) (map[string]ZoneUnlock, error)
}
