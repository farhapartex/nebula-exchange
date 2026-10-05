package models_test

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/models"
)

func textPointer(text string) *string {
	return &text
}

func numberPointer(number int) *int {
	return &number
}

func seedFirstLevel(t *testing.T, testDatabase *gorm.DB) {
	t.Helper()
	records := []any{
		&models.Chapter{ID: "1", Number: 1, Title: "The night they came", Summary: "Bandits burn the village.", IsFree: true, IsPublished: true},
		&models.Arena{ID: "burning-house", Name: "The burning house", Width: 960, FloorY: 470, Stage: database.JSONDocument(`{"world_units_per_pixel":0.0105,"layers":[]}`)},
		&models.Enemy{ID: "hammer-bandit", Name: "Hammer bandit", Title: "First of the gang", Stats: database.JSONDocument(`{"max_health":80}`), Brain: database.JSONDocument(`{"aggression":0.4}`), Look: database.JSONDocument(`{"body_color":"#7f1d1d"}`)},
		&models.Enemy{ID: "knife-bandit", Name: "Knife bandit", Title: "Quick and quiet", Stats: database.JSONDocument(`{"max_health":70}`), Brain: database.JSONDocument(`{"aggression":0.6}`), Look: database.JSONDocument(`{"body_color":"#44403c"}`)},
		&models.Level{ID: "1-1", ChapterID: textPointer("1"), Number: numberPointer(1), Kind: models.LevelKindStory, Title: "The burning house", Teaser: "Get out alive.", ArenaID: "burning-house", TimeLimitSeconds: 90, Difficulty: database.JSONDocument(`{"enemy_health_multiplier":1}`), StarRules: database.JSONDocument(`[{"kind":"win"}]`), FirstClearCoins: 5_000_000, FirstClearExperience: 40, ReplayExperience: 10, IsPublished: true},
		&models.LevelEnemy{LevelID: "1-1", Wave: 1, EnemyID: "hammer-bandit", Modifiers: database.JSONDocument(`{}`), IntroLine: textPointer("Nothing to give? Then burn.")},
		&models.LevelEnemy{LevelID: "1-1", Wave: 2, EnemyID: "knife-bandit", Modifiers: database.JSONDocument(`{"damage_multiplier":1.1}`)},
		&models.StorySlide{ID: uuid.New(), LevelID: "1-1", Position: 1, Kind: models.SlideKindSlide, Eyebrow: textPointer("Chapter 1"), Heading: "At the edge of the village", Body: "A thin boy lives alone.", ImageURL: textPointer("/story/1-1/01-edge-of-the-village.webp"), Palette: database.JSONDocument(`{"background":"#05090a","glow":"#2dd4bf","accent":"#5eead4"}`)},
		&models.StorySlide{ID: uuid.New(), LevelID: "1-1", Position: 2, Kind: models.SlideKindCallToAction, Heading: "Fight to survive", Body: "Win this fight.", Palette: database.JSONDocument(`{"background":"#0a0807","glow":"#f97316","accent":"#fdba74"}`), ButtonLabel: textPointer("Play")},
	}
	for _, record := range records {
		if err := testDatabase.Create(record).Error; err != nil {
			t.Fatalf("create %T: %v", record, err)
		}
	}
}

func TestLevelLoadsWithItsArenaWavesAndSlides(t *testing.T) {
	testDatabase := databasetest.Open(t)
	seedFirstLevel(t, testDatabase)

	var level models.Level
	err := testDatabase.
		Preload("Chapter").
		Preload("Arena").
		Preload("Enemies", func(query *gorm.DB) *gorm.DB { return query.Order("wave") }).
		Preload("Enemies.Enemy").
		Preload("Slides", func(query *gorm.DB) *gorm.DB { return query.Order("position") }).
		Take(&level, "id = ?", "1-1").Error
	if err != nil {
		t.Fatalf("load level: %v", err)
	}

	if level.Chapter == nil || level.Chapter.Title != "The night they came" || level.Arena.Width != 960 {
		t.Fatalf("unexpected chapter or arena: %+v %+v", level.Chapter, level.Arena)
	}
	if len(level.Enemies) != 2 || level.Enemies[0].Enemy.Name != "Hammer bandit" || level.Enemies[1].Wave != 2 {
		t.Fatalf("unexpected waves: %+v", level.Enemies)
	}
	if len(level.Slides) != 2 || level.Slides[1].Kind != models.SlideKindCallToAction || *level.Slides[1].ButtonLabel != "Play" {
		t.Fatalf("unexpected slides: %+v", level.Slides)
	}
	if string(level.Slides[0].Palette) == "" || string(level.StarRules) != `[{"kind": "win"}]` {
		t.Fatalf("json columns did not round trip: palette %s, star rules %s", level.Slides[0].Palette, level.StarRules)
	}
}

func TestContentRulesAreEnforcedByTheDatabase(t *testing.T) {
	testDatabase := databasetest.Open(t)
	seedFirstLevel(t, testDatabase)

	invalidRecords := map[string]any{
		"paid chapter without a price": &models.Chapter{ID: "2", Number: 2, Title: "Paid", Summary: "x", IsFree: false},
		"free chapter with a price":    &models.Chapter{ID: "3", Number: 3, Title: "Free", Summary: "x", IsFree: true, PriceCoins: func() *int64 { price := int64(1); return &price }()},
		"training level in a chapter":  &models.Level{ID: "training", ChapterID: textPointer("1"), Number: numberPointer(9), Kind: models.LevelKindTraining, Title: "Training", Teaser: "x", ArenaID: "burning-house", TimeLimitSeconds: 60, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(`[]`)},
		"duplicate level number":       &models.Level{ID: "1-1b", ChapterID: textPointer("1"), Number: numberPointer(1), Kind: models.LevelKindStory, Title: "Again", Teaser: "x", ArenaID: "burning-house", TimeLimitSeconds: 60, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(`[]`)},
		"second call to action":        &models.StorySlide{ID: uuid.New(), LevelID: "1-1", Position: 3, Kind: models.SlideKindCallToAction, Heading: "Again", Body: "x", Palette: database.JSONDocument(`{}`), ButtonLabel: textPointer("Play")},
		"button on a plain slide":      &models.StorySlide{ID: uuid.New(), LevelID: "1-1", Position: 4, Kind: models.SlideKindSlide, Heading: "Slide", Body: "x", Palette: database.JSONDocument(`{}`), ButtonLabel: textPointer("Play")},
		"wave with an unknown enemy":   &models.LevelEnemy{LevelID: "1-1", Wave: 3, EnemyID: "gun-bandit", Modifiers: database.JSONDocument(`{}`)},
	}
	for caseName, invalidRecord := range invalidRecords {
		t.Run(caseName, func(t *testing.T) {
			if err := testDatabase.Create(invalidRecord).Error; err == nil {
				t.Fatalf("expected %s to be rejected", caseName)
			}
		})
	}

	trainingLevel := &models.Level{ID: "training", Kind: models.LevelKindTraining, Title: "Training", Teaser: "Practice", ArenaID: "burning-house", TimeLimitSeconds: 120, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(`[]`), IsPublished: true}
	if err := testDatabase.Create(trainingLevel).Error; err != nil {
		t.Fatalf("a training level without a chapter must be allowed: %v", err)
	}
}
