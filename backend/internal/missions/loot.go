package missions

import (
	"crypto/rand"
	"math/big"
	"slices"

	"nebula-exchange/backend/internal/catalog"
)

type LootRoll struct {
	ItemID   int `json:"item_id"`
	Quantity int `json:"quantity"`
}

type LootRules struct {
	Loot                       []catalog.LootEntry
	RarityRankByItem           map[int]int
	DrillMultiplierBasisPoints int
	CargoCapacity              int
}

type RandomSource func(exclusiveUpperBound int) int

func CryptoRandom(exclusiveUpperBound int) int {
	randomValue, err := rand.Int(rand.Reader, big.NewInt(int64(exclusiveUpperBound)))
	if err != nil {
		panic("crypto/rand is unavailable: " + err.Error())
	}
	return int(randomValue.Int64())
}

func RollLoot(rules LootRules, random RandomSource) []LootRoll {
	rolls := make([]LootRoll, 0, len(rules.Loot))
	for _, lootEntry := range rules.Loot {
		isChanceDrop := lootEntry.ChanceBasisPoints < catalog.BasisPointsPerWhole
		if isChanceDrop && random(catalog.BasisPointsPerWhole) >= lootEntry.ChanceBasisPoints {
			continue
		}
		quantity := lootEntry.MinimumQuantity + random(lootEntry.MaximumQuantity-lootEntry.MinimumQuantity+1)
		if !isChanceDrop {
			quantity = quantity * rules.DrillMultiplierBasisPoints / catalog.BasisPointsPerWhole
		}
		if quantity > 0 {
			rolls = append(rolls, LootRoll{ItemID: lootEntry.ItemID, Quantity: quantity})
		}
	}
	return capToCargo(rolls, rules.CargoCapacity, rules.RarityRankByItem)
}

func capToCargo(rolls []LootRoll, cargoCapacity int, rarityRankByItem map[int]int) []LootRoll {
	totalUnits := 0
	for _, roll := range rolls {
		totalUnits += roll.Quantity
	}
	if totalUnits <= cargoCapacity {
		return rolls
	}

	capped := make([]LootRoll, len(rolls))
	assignedUnits := 0
	for rollIndex, roll := range rolls {
		scaledQuantity := roll.Quantity * cargoCapacity / totalUnits
		capped[rollIndex] = LootRoll{ItemID: roll.ItemID, Quantity: scaledQuantity}
		assignedUnits += scaledQuantity
	}

	rarestFirst := make([]int, len(capped))
	for rollIndex := range capped {
		rarestFirst[rollIndex] = rollIndex
	}
	slices.SortStableFunc(rarestFirst, func(first, second int) int {
		return rarityRankByItem[capped[second].ItemID] - rarityRankByItem[capped[first].ItemID]
	})
	for leftoverUnits := cargoCapacity - assignedUnits; leftoverUnits > 0; {
		for _, rollIndex := range rarestFirst {
			if leftoverUnits == 0 {
				break
			}
			if capped[rollIndex].Quantity < rolls[rollIndex].Quantity {
				capped[rollIndex].Quantity++
				leftoverUnits--
			}
		}
	}

	return slices.DeleteFunc(capped, func(roll LootRoll) bool { return roll.Quantity == 0 })
}

func MissionDurationSeconds(zoneDurationSeconds int, speedMultiplierBasisPoints int) int {
	return zoneDurationSeconds * speedMultiplierBasisPoints / catalog.BasisPointsPerWhole
}
