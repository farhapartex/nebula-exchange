package seeding_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/seeding"
)

const (
	validSVG        = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><defs><linearGradient id="g"><stop offset="0" stop-color="#fff"/></linearGradient></defs><rect width="10" height="10" fill="url(#g)"/></svg>`
	validToolFields = `"id":"x","name":"x","description":"x","category":"WEAPON","rarity":"COMMON","base_stats":{"a":1},"minimum_fighter_level":1,"image_svg_file":"tool.svg","mastery_curve":{"points_to_reach_level":[0,100],"stat_gain_per_level":{"a":1}}`
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

func writeTool(t *testing.T, toolJSON string, svgMarkup string) string {
	t.Helper()
	toolDirectory := t.TempDir()
	if err := os.WriteFile(filepath.Join(toolDirectory, "tool.svg"), []byte(svgMarkup), 0o600); err != nil {
		t.Fatalf("write svg: %v", err)
	}
	toolFile := filepath.Join(toolDirectory, "tool.json")
	if err := os.WriteFile(toolFile, []byte(toolJSON), 0o600); err != nil {
		t.Fatalf("write tool: %v", err)
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
	if len(seededTools) != len(toolSeedFiles(t)) || len(seededTools) < 2 {
		t.Fatalf("unexpected seeded tools %+v", seededTools)
	}
	for _, seededTool := range seededTools {
		if seededTool.ImageSVG == nil || !strings.HasPrefix(*seededTool.ImageSVG, "<svg") || seededTool.MinimumFighterLevel < 1 {
			t.Fatalf("tool %s is missing its SVG or minimum level", seededTool.ID)
		}
	}
}

func TestAValidToolFileLoads(t *testing.T) {
	toolType, err := seeding.LoadToolType(writeTool(t, "{"+validToolFields+"}", validSVG))
	if err != nil || toolType.ImageSVG != validSVG {
		t.Fatalf("got %v", err)
	}
}

func TestBrokenToolFilesAreRejected(t *testing.T) {
	brokenTools := map[string]string{
		"unknown category":   strings.Replace(validToolFields, `"WEAPON"`, `"HAT"`, 1),
		"curve not rising":   strings.Replace(validToolFields, `[0,100]`, `[0,100,100]`, 1),
		"curve too long":     strings.Replace(validToolFields, `[0,100]`, `[0,1,2,3,4,5,6,7,8,9,10]`, 1),
		"no base stats":      strings.Replace(validToolFields, `{"a":1},"minimum`, `{},"minimum`, 1),
		"minimum level zero": strings.Replace(validToolFields, `"minimum_fighter_level":1`, `"minimum_fighter_level":0`, 1),
		"image outside":      strings.Replace(validToolFields, `"tool.svg"`, `"../tool.svg"`, 1),
	}
	for caseName, brokenFields := range brokenTools {
		if _, err := seeding.LoadToolType(writeTool(t, "{"+brokenFields+"}", validSVG)); err == nil {
			t.Errorf("%s: expected an error", caseName)
		}
	}
}

func TestUnsafeSVGImagesAreRejected(t *testing.T) {
	unsafeSVGs := map[string]string{
		"script element":  `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`,
		"event attribute": `<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"><rect width="1" height="1"/></svg>`,
		"outside link":    `<svg xmlns="http://www.w3.org/2000/svg"><use href="https://evil.test/a.svg#x"/></svg>`,
		"outside image":   `<svg xmlns="http://www.w3.org/2000/svg"><image href="#x"/></svg>`,
		"outside url":     `<svg xmlns="http://www.w3.org/2000/svg"><rect fill="url(https://evil.test/x)"/></svg>`,
		"doctype":         `<!DOCTYPE svg [<!ENTITY x "y">]><svg xmlns="http://www.w3.org/2000/svg"/>`,
		"not an svg":      `<html><body/></html>`,
		"broken xml":      `<svg xmlns="http://www.w3.org/2000/svg"><rect></svg>`,
	}
	for caseName, unsafeSVG := range unsafeSVGs {
		if _, err := seeding.LoadToolType(writeTool(t, "{"+validToolFields+"}", unsafeSVG)); err == nil {
			t.Errorf("%s: expected the SVG to be rejected", caseName)
		}
	}
}
