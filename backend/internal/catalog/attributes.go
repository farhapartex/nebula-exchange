package catalog

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const BasisPointsPerWhole = 10_000

const legendaryDrillTier = 5

type itemAttributes struct {
	DrillMultiplier string `json:"drill_multiplier"`
	CargoCapacity   int    `json:"cargo_capacity"`
	SpeedMultiplier string `json:"speed_multiplier"`
}

type DrillStats struct {
	Tier                       int
	YieldMultiplierBasisPoints int
}

type ShipStats struct {
	CargoCapacity              int
	SpeedMultiplierBasisPoints int
}

func (item Item) parsedAttributes() itemAttributes {
	var attributes itemAttributes
	_ = json.Unmarshal(item.Attributes, &attributes)
	return attributes
}

func (item Item) DrillStats() (DrillStats, bool) {
	attributes := item.parsedAttributes()
	if attributes.DrillMultiplier == "" || (item.Category != CategoryDrill && item.Category != CategoryLegendary) {
		return DrillStats{}, false
	}
	multiplier, err := ParseBasisPoints(attributes.DrillMultiplier)
	if err != nil {
		return DrillStats{}, false
	}
	tier := legendaryDrillTier
	if item.Tier != nil {
		tier = *item.Tier
	}
	return DrillStats{Tier: tier, YieldMultiplierBasisPoints: multiplier}, true
}

func (item Item) ShipStats() (ShipStats, bool) {
	if item.Category != CategoryShip {
		return ShipStats{}, false
	}
	attributes := item.parsedAttributes()
	speed, err := ParseBasisPoints(attributes.SpeedMultiplier)
	if err != nil || attributes.CargoCapacity <= 0 {
		return ShipStats{}, false
	}
	return ShipStats{CargoCapacity: attributes.CargoCapacity, SpeedMultiplierBasisPoints: speed}, true
}

func ParseBasisPoints(decimalText string) (int, error) {
	wholePart, fractionPart, _ := strings.Cut(strings.TrimSpace(decimalText), ".")
	if wholePart == "" || len(fractionPart) > 4 {
		return 0, fmt.Errorf("multiplier %q must have up to 4 decimals", decimalText)
	}
	basisPoints, err := strconv.Atoi(wholePart + fractionPart + strings.Repeat("0", 4-len(fractionPart)))
	if err != nil || basisPoints <= 0 {
		return 0, fmt.Errorf("multiplier %q is not a positive decimal", decimalText)
	}
	return basisPoints, nil
}

func (zone Zone) AllowsShip(shipItemID int) bool {
	if len(zone.AllowedShipItemIDs) == 0 {
		return true
	}
	for _, allowedShipItemID := range zone.AllowedShipItemIDs {
		if allowedShipItemID == shipItemID {
			return true
		}
	}
	return false
}

func (snapshot Snapshot) ZoneByID(zoneID string) (Zone, bool) {
	for _, zone := range snapshot.Zones {
		if zone.ID == zoneID {
			return zone, true
		}
	}
	return Zone{}, false
}
