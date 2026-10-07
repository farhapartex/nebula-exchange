package models_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	identitymodels "github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	progressmodels "github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
	storymodels "github.com/farhapartex/nebula-exchange/backend/internal/story/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/models"
)

type toolFixture struct {
	database       *gorm.DB
	ownerID        uuid.UUID
	ironPipe       models.Tool
	fightSessionID uuid.UUID
}

func newToolFixture(t *testing.T) toolFixture {
	t.Helper()
	testDatabase := databasetest.Open(t)
	ownerID, fightSessionID := uuid.New(), uuid.New()
	activatedAt := time.Now()
	shopPrice := int64(300)
	levelID := "1-1"
	ironPipe := models.Tool{ID: uuid.New(), ToolTypeID: "iron-pipe", OwnerUserID: ownerID, MasteryLevel: 1, Status: models.ToolStatusOwned, AcquiredFrom: models.ToolSourceReward, AcquiredAt: activatedAt}
	records := []any{
		&identitymodels.User{ID: ownerID, Email: "boy@streetborn.test", Username: "the_boy", PasswordHash: "x", Status: identitymodels.UserStatusActive, IsActive: true, ActivatedAt: &activatedAt, TermsAcceptedAt: activatedAt},
		&storymodels.Chapter{ID: "1", Number: 1, Title: "The night they came", Summary: "x", IsFree: true, IsPublished: true},
		&storymodels.Arena{ID: "burning-house", Name: "The burning house", Width: 960, FloorY: 470, Stage: database.JSONDocument(`{}`)},
		&storymodels.Level{ID: levelID, ChapterID: &[]string{"1"}[0], Number: &[]int{1}[0], Kind: storymodels.LevelKindStory, Title: "x", Teaser: "x", ArenaID: "burning-house", TimeLimitSeconds: 90, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(`[]`), IsPublished: true},
		&models.ToolType{ID: "iron-pipe", Name: "Iron pipe", Description: "x", Category: models.ToolCategoryWeapon, Rarity: models.ToolRarityCommon, BaseStats: database.JSONDocument(`{"damage_bonus":4}`), MasteryCurve: database.JSONDocument(`{"points_to_reach_level":[0,100]}`), MaxMasteryLevel: 2, MinimumFighterLevel: 1, ShopPriceCoins: &shopPrice, IsTradeable: true, IntroducedInLevelID: &levelID},
		&models.ToolType{ID: "scrap-shield", Name: "Scrap shield", Description: "x", Category: models.ToolCategoryGuard, Rarity: models.ToolRarityCommon, BaseStats: database.JSONDocument(`{"block_damage_reduction":0.25}`), MasteryCurve: database.JSONDocument(`{"points_to_reach_level":[0,100]}`), MaxMasteryLevel: 2, MinimumFighterLevel: 2, IsTradeable: true},
		&ironPipe,
		&progressmodels.FightSession{ID: fightSessionID, UserID: ownerID, LevelID: levelID, Status: progressmodels.FightStatusStarted, Seed: 1, Loadout: database.JSONDocument(`{}`), StartedAt: activatedAt},
	}
	for _, record := range records {
		if err := testDatabase.Create(record).Error; err != nil {
			t.Fatalf("create %T: %v", record, err)
		}
	}
	return toolFixture{database: testDatabase, ownerID: ownerID, ironPipe: ironPipe, fightSessionID: fightSessionID}
}

func TestAnOwnedToolLoadsWithItsType(t *testing.T) {
	fixture := newToolFixture(t)
	var ownedTool models.Tool
	if err := fixture.database.Preload("ToolType").Take(&ownedTool, "id = ?", fixture.ironPipe.ID).Error; err != nil {
		t.Fatalf("load tool: %v", err)
	}
	if ownedTool.ToolType.Name != "Iron pipe" || ownedTool.ToolType.Category != models.ToolCategoryWeapon || *ownedTool.ToolType.IntroducedInLevelID != "1-1" || ownedTool.MasteryLevel != 1 {
		t.Fatalf("unexpected tool %+v", ownedTool)
	}
}

func TestALoadoutSlotHoldsOneToolAndAToolOnlyOnce(t *testing.T) {
	fixture := newToolFixture(t)
	weaponSlot := models.LoadoutSlot{UserID: fixture.ownerID, Slot: models.LoadoutSlotWeapon, ToolID: &fixture.ironPipe.ID}
	if err := fixture.database.Create(&weaponSlot).Error; err != nil {
		t.Fatalf("equip weapon: %v", err)
	}
	guardSlotWithSameTool := models.LoadoutSlot{UserID: fixture.ownerID, Slot: models.LoadoutSlotGuard, ToolID: &fixture.ironPipe.ID}
	if err := fixture.database.Create(&guardSlotWithSameTool).Error; err == nil {
		t.Fatal("one tool must not sit in two slots")
	}
	var loadedSlot models.LoadoutSlot
	if err := fixture.database.Preload("Tool.ToolType").Take(&loadedSlot, "user_id = ? AND slot = ?", fixture.ownerID, models.LoadoutSlotWeapon).Error; err != nil {
		t.Fatalf("load slot: %v", err)
	}
	if loadedSlot.Tool == nil || loadedSlot.Tool.ToolType.ID != "iron-pipe" {
		t.Fatalf("unexpected slot %+v", loadedSlot)
	}
}

func TestMasteryIsRecordedOncePerFight(t *testing.T) {
	fixture := newToolFixture(t)
	firstEvent := models.ToolMasteryEvent{ID: uuid.New(), ToolID: fixture.ironPipe.ID, FightSessionID: fixture.fightSessionID, PointsGained: 25, LevelAfter: 1}
	if err := fixture.database.Create(&firstEvent).Error; err != nil {
		t.Fatalf("record mastery: %v", err)
	}
	repeatedEvent := models.ToolMasteryEvent{ID: uuid.New(), ToolID: fixture.ironPipe.ID, FightSessionID: fixture.fightSessionID, PointsGained: 25, LevelAfter: 1}
	if err := fixture.database.Create(&repeatedEvent).Error; err == nil {
		t.Fatal("a fight must count toward a tool's mastery only once")
	}
}

func TestToolTypeRulesAreEnforced(t *testing.T) {
	fixture := newToolFixture(t)
	freePrice := int64(0)
	brokenToolTypes := map[string]models.ToolType{
		"zero shop price":    {ID: "free-pipe", Name: "x", Description: "x", Category: models.ToolCategoryWeapon, Rarity: models.ToolRarityCommon, BaseStats: database.JSONDocument(`{"a":1}`), MasteryCurve: database.JSONDocument(`{}`), MaxMasteryLevel: 1, MinimumFighterLevel: 1, ShopPriceCoins: &freePrice},
		"mastery above ten":  {ID: "legend-pipe", Name: "x", Description: "x", Category: models.ToolCategoryWeapon, Rarity: models.ToolRarityLegendary, BaseStats: database.JSONDocument(`{"a":1}`), MasteryCurve: database.JSONDocument(`{}`), MaxMasteryLevel: 11, MinimumFighterLevel: 1},
		"minimum level zero": {ID: "early-pipe", Name: "x", Description: "x", Category: models.ToolCategoryWeapon, Rarity: models.ToolRarityCommon, BaseStats: database.JSONDocument(`{"a":1}`), MasteryCurve: database.JSONDocument(`{}`), MaxMasteryLevel: 1, MinimumFighterLevel: 0},
		"stats are an array": {ID: "array-pipe", Name: "x", Description: "x", Category: models.ToolCategoryWeapon, Rarity: models.ToolRarityCommon, BaseStats: database.JSONDocument(`[1]`), MasteryCurve: database.JSONDocument(`{}`), MaxMasteryLevel: 1, MinimumFighterLevel: 1},
	}
	for caseName, brokenToolType := range brokenToolTypes {
		if err := fixture.database.Create(&brokenToolType).Error; err == nil {
			t.Errorf("%s: expected the database to refuse it", caseName)
		}
	}
}
