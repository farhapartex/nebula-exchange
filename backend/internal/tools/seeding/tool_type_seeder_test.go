package seeding_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/seeding"
)

func toolSeedFiles(t *testing.T) []string {
	t.Helper()
	_, currentFile, _, _ := runtime.Caller(0)
	toolFiles, err := filepath.Glob(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "seeds", "tools", "*.json"))
	if err != nil || len(toolFiles) == 0 {
		t.Fatalf("no tool seed files found: %v", err)
	}
	return toolFiles
}

func writeToolFile(t *testing.T, content string) string {
	t.Helper()
	toolFile := filepath.Join(t.TempDir(), "tool.json")
	if err := os.WriteFile(toolFile, []byte(content), 0o600); err != nil {
		t.Fatalf("write tool file: %v", err)
	}
	return toolFile
}

func TestTheToolSeedFilesAreValidAndSeedTwice(t *testing.T) {
	testDatabase := databasetest.Open(t)
	quietLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for round := 1; round <= 2; round++ {
		for _, toolFile := range toolSeedFiles(t) {
			toolType, err := seeding.LoadToolType(toolFile)
			if err != nil {
				t.Fatalf("load %s: %v", toolFile, err)
			}
			if err := seeding.SeedToolType(context.Background(), testDatabase, toolType, quietLogger); err != nil {
				t.Fatalf("round %d seed %s: %v", round, toolFile, err)
			}
		}
	}
	var seededTools []models.ToolType
	testDatabase.Order("id").Find(&seededTools)
	if len(seededTools) != 2 || seededTools[0].ID != "iron-pipe" || seededTools[0].MaxMasteryLevel != 10 || seededTools[1].Category != models.ToolCategoryGuard {
		t.Fatalf("unexpected seeded tools %+v", seededTools)
	}
}

func TestBrokenToolFilesAreRejected(t *testing.T) {
	brokenFiles := map[string]string{
		"unknown category": `{"id":"x","name":"x","description":"x","category":"HAT","rarity":"COMMON","base_stats":{"a":1},"mastery_curve":{"points_to_reach_level":[0],"stat_gain_per_level":{"a":1}}}`,
		"curve not rising": `{"id":"x","name":"x","description":"x","category":"WEAPON","rarity":"COMMON","base_stats":{"a":1},"mastery_curve":{"points_to_reach_level":[0,100,100],"stat_gain_per_level":{"a":1}}}`,
		"curve too long":   `{"id":"x","name":"x","description":"x","category":"WEAPON","rarity":"COMMON","base_stats":{"a":1},"mastery_curve":{"points_to_reach_level":[0,1,2,3,4,5,6,7,8,9,10],"stat_gain_per_level":{"a":1}}}`,
		"no base stats":    `{"id":"x","name":"x","description":"x","category":"GUARD","rarity":"COMMON","base_stats":{},"mastery_curve":{"points_to_reach_level":[0],"stat_gain_per_level":{"a":1}}}`,
		"free in the shop": `{"id":"x","name":"x","description":"x","category":"GUARD","rarity":"COMMON","base_stats":{"a":1},"shop_price_coins":0,"mastery_curve":{"points_to_reach_level":[0],"stat_gain_per_level":{"a":1}}}`,
	}
	for caseName, brokenContent := range brokenFiles {
		if _, err := seeding.LoadToolType(writeToolFile(t, brokenContent)); err == nil {
			t.Errorf("%s: expected an error", caseName)
		}
	}
}
