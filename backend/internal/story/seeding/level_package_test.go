package seeding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePackage(t *testing.T, levelJSON string, imageFileNames ...string) string {
	t.Helper()
	directory := t.TempDir()
	if err := os.MkdirAll(filepath.Join(directory, imagesDirectoryName), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, imageFileName := range imageFileNames {
		if err := os.WriteFile(filepath.Join(directory, imagesDirectoryName, imageFileName), []byte("webp"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(directory, levelFileName), []byte(levelJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	return directory
}

const packageHeader = `"chapter":{"id":"1","number":1},"arena":{"id":"a","stage":{}},"level":{"id":"1-1","number":1}`

func TestLoadLevelPackageRejectsBrokenContent(t *testing.T) {
	brokenPackages := map[string]struct {
		levelJSON      string
		images         []string
		expectedReason string
	}{
		"missing image file": {
			levelJSON:      `{` + packageHeader + `,"slides":[{"kind":"SLIDE","image":"01.webp"},{"kind":"CALL_TO_ACTION"}]}`,
			expectedReason: "no such file",
		},
		"call to action not last": {
			levelJSON:      `{` + packageHeader + `,"slides":[{"kind":"CALL_TO_ACTION"},{"kind":"SLIDE"}]}`,
			expectedReason: "only the last slide",
		},
		"image outside the images folder": {
			levelJSON:      `{` + packageHeader + `,"slides":[{"kind":"SLIDE","image":"../secret.webp"},{"kind":"CALL_TO_ACTION"}]}`,
			expectedReason: "must be a file name",
		},
		"no slides": {
			levelJSON:      `{` + packageHeader + `,"slides":[]}`,
			expectedReason: "at least one slide",
		},
	}
	for caseName, brokenPackage := range brokenPackages {
		t.Run(caseName, func(t *testing.T) {
			_, err := LoadLevelPackage(writePackage(t, brokenPackage.levelJSON, brokenPackage.images...))
			if err == nil || !strings.Contains(err.Error(), brokenPackage.expectedReason) {
				t.Fatalf("got %v, want an error about %q", err, brokenPackage.expectedReason)
			}
		})
	}
}

func TestLoadLevelPackageAcceptsTheLevelOnePackage(t *testing.T) {
	levelPackage, err := LoadLevelPackage(filepath.Join("..", "..", "..", "seeds", "story", "level-1-1"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(levelPackage.Slides) != 6 || levelPackage.ImageObjectKey("05-fire.webp") != "story/1-1/05-fire.webp" {
		t.Fatalf("unexpected package %d slides", len(levelPackage.Slides))
	}
}
