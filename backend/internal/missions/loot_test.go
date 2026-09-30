package missions

import (
	"reflect"
	"testing"

	"nebula-exchange/backend/internal/catalog"
)

func fixedRandom(values ...int) RandomSource {
	nextIndex := 0
	return func(exclusiveUpperBound int) int {
		value := values[nextIndex%len(values)]
		nextIndex++
		return min(value, exclusiveUpperBound-1)
	}
}

var seededRarity = map[int]int{1: 1, 2: 2, 3: 3, 4: 4, 5: 5, 6: 6}

func TestRollLoot(t *testing.T) {
	asteroidBelt := []catalog.LootEntry{
		{ItemID: 1, MinimumQuantity: 8, MaximumQuantity: 14, ChanceBasisPoints: 10_000},
		{ItemID: 2, MinimumQuantity: 3, MaximumQuantity: 6, ChanceBasisPoints: 10_000},
	}
	deepVoid := []catalog.LootEntry{
		{ItemID: 5, MinimumQuantity: 2, MaximumQuantity: 4, ChanceBasisPoints: 10_000},
		{ItemID: 6, MinimumQuantity: 1, MaximumQuantity: 1, ChanceBasisPoints: 500},
	}

	cases := []struct {
		name     string
		rules    LootRules
		random   RandomSource
		expected []LootRoll
	}{
		{
			name:     "minimum rolls with a T1 drill",
			rules:    LootRules{Loot: asteroidBelt, RarityRankByItem: seededRarity, DrillMultiplierBasisPoints: 10_000, CargoCapacity: 60},
			random:   fixedRandom(0),
			expected: []LootRoll{{1, 8}, {2, 3}},
		},
		{
			name:     "maximum rolls with a T2 drill round down",
			rules:    LootRules{Loot: asteroidBelt, RarityRankByItem: seededRarity, DrillMultiplierBasisPoints: 12_500, CargoCapacity: 60},
			random:   fixedRandom(99),
			expected: []LootRoll{{1, 17}, {2, 7}},
		},
		{
			name:     "a Scout's 20-unit hold scales loot down and gives leftovers to the rarer resource",
			rules:    LootRules{Loot: asteroidBelt, RarityRankByItem: seededRarity, DrillMultiplierBasisPoints: 12_500, CargoCapacity: 20},
			random:   fixedRandom(99),
			expected: []LootRoll{{1, 14}, {2, 6}},
		},
		{
			name:     "over the cap with a T5 drill",
			rules:    LootRules{Loot: asteroidBelt, RarityRankByItem: seededRarity, DrillMultiplierBasisPoints: 25_000, CargoCapacity: 20},
			random:   fixedRandom(99),
			expected: []LootRoll{{1, 14}, {2, 6}},
		},
		{
			name:     "void shard drops inside its 5% chance and is not multiplied",
			rules:    LootRules{Loot: deepVoid, RarityRankByItem: seededRarity, DrillMultiplierBasisPoints: 20_000, CargoCapacity: 150},
			random:   fixedRandom(2, 499, 0),
			expected: []LootRoll{{5, 8}, {6, 1}},
		},
		{
			name:     "void shard misses outside its chance",
			rules:    LootRules{Loot: deepVoid, RarityRankByItem: seededRarity, DrillMultiplierBasisPoints: 20_000, CargoCapacity: 150},
			random:   fixedRandom(0, 500),
			expected: []LootRoll{{5, 4}},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if rolled := RollLoot(testCase.rules, testCase.random); !reflect.DeepEqual(rolled, testCase.expected) {
				t.Fatalf("got %v, want %v", rolled, testCase.expected)
			}
		})
	}
}

func TestCargoCapNeverExceededAndRarestKeptLongest(t *testing.T) {
	capped := capToCargo([]LootRoll{{1, 100}, {6, 3}}, 10, seededRarity)
	total := 0
	for _, roll := range capped {
		total += roll.Quantity
	}
	if total != 10 {
		t.Fatalf("capped total %d", total)
	}
	if capped[len(capped)-1].ItemID != 6 || capped[len(capped)-1].Quantity < 1 {
		t.Fatalf("the rarest resource must survive the cap: %v", capped)
	}
}

func TestRollLootWithCryptoRandomStaysInRange(t *testing.T) {
	rules := LootRules{
		Loot:                       []catalog.LootEntry{{ItemID: 1, MinimumQuantity: 8, MaximumQuantity: 14, ChanceBasisPoints: 10_000}},
		RarityRankByItem:           seededRarity,
		DrillMultiplierBasisPoints: 10_000,
		CargoCapacity:              60,
	}
	for attempt := 0; attempt < 500; attempt++ {
		rolled := RollLoot(rules, CryptoRandom)
		if len(rolled) != 1 || rolled[0].Quantity < 8 || rolled[0].Quantity > 14 {
			t.Fatalf("roll out of range: %v", rolled)
		}
	}
}

func TestMissionDurationFollowsShipSpeed(t *testing.T) {
	for speed, expected := range map[int]int{10_000: 900, 6_000: 540, 12_000: 1080} {
		if duration := MissionDurationSeconds(900, speed); duration != expected {
			t.Fatalf("speed %d: %d, want %d", speed, duration, expected)
		}
	}
}
