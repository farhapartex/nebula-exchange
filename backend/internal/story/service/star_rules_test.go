package service

import (
	"testing"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

func TestStarsFollowTheTimeSlots(t *testing.T) {
	starRules, err := ParseStarRules(database.JSONDocument(`[{"stars":3,"within_seconds":45},{"stars":2,"within_seconds":70},{"stars":1,"within_seconds":90}]`), 90)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	expectations := map[float64]int{10: 3, 45: 3, 45.2: 2, 70: 2, 71: 1, 90: 1, 90.5: 0}
	for fightSeconds, expectedStars := range expectations {
		if stars := starRules.StarsForWin(fightSeconds); stars != expectedStars {
			t.Fatalf("%.1fs: got %d stars, want %d", fightSeconds, stars, expectedStars)
		}
	}
}

func TestStarRulesMustBeOrderedAndInsideTheTimeLimit(t *testing.T) {
	brokenRules := []string{
		`[{"stars":2,"within_seconds":45},{"stars":3,"within_seconds":70}]`,
		`[{"stars":3,"within_seconds":45},{"stars":2,"within_seconds":40}]`,
		`[{"stars":1,"within_seconds":120}]`,
		`[{"stars":4,"within_seconds":30}]`,
		`{"stars":3}`,
	}
	for _, brokenRule := range brokenRules {
		if _, err := ParseStarRules(database.JSONDocument(brokenRule), 90); err == nil {
			t.Fatalf("expected %s to be rejected", brokenRule)
		}
	}
}
