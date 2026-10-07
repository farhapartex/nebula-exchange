package tools_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	identitymodels "github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/story"
	storymodels "github.com/farhapartex/nebula-exchange/backend/internal/story/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/models"
)

const playerToken = "player-token"

type fixedPlayer struct {
	userID uuid.UUID
}

func (player fixedPlayer) Verify(accessToken string) (uuid.UUID, error) {
	if accessToken != playerToken {
		return uuid.Nil, errors.New("invalid token")
	}
	return player.userID, nil
}

type fakePlayerStanding struct {
	fighterLevel int
	wonLevelIDs  map[string]bool
}

func (standing *fakePlayerStanding) FighterLevel(context.Context, uuid.UUID) (int, error) {
	return standing.fighterLevel, nil
}

func (standing *fakePlayerStanding) WonLevelIDs(context.Context, uuid.UUID) (map[string]bool, error) {
	return standing.wonLevelIDs, nil
}

type shopHarness struct {
	t        *testing.T
	database *gorm.DB
	router   *gin.Engine
	playerID uuid.UUID
	standing *fakePlayerStanding
}

type shopListBody struct {
	Data       []map[string]any `json:"data"`
	Pagination struct {
		NextCursor *string `json:"next_cursor"`
		Limit      int     `json:"limit"`
	} `json:"pagination"`
	Error struct {
		Code    string            `json:"code"`
		Details map[string]string `json:"details"`
	} `json:"error"`
}

func newShopHarness(t *testing.T) *shopHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	testDatabase := databasetest.Open(t)
	playerID := uuid.New()
	activatedAt := time.Now()
	chapterID, levelID, levelNumber := "2", "2-1", 1
	svgMarkup := `<svg xmlns="http://www.w3.org/2000/svg"/>`
	records := []any{
		&identitymodels.User{ID: playerID, Email: "boy@streetborn.test", Username: "the_boy", PasswordHash: "x", Status: identitymodels.UserStatusActive, IsActive: true, ActivatedAt: &activatedAt, TermsAcceptedAt: activatedAt},
		&storymodels.Chapter{ID: chapterID, Number: 2, Title: "Under the bridge", Summary: "x", IsFree: true, IsPublished: true},
		&storymodels.Arena{ID: "bridge", Name: "Bridge", Width: 960, FloorY: 470, Stage: database.JSONDocument(`{}`)},
		&storymodels.Level{ID: levelID, ChapterID: &chapterID, Number: &levelNumber, Kind: storymodels.LevelKindStory, Title: "The toll", Teaser: "x", ArenaID: "bridge", TimeLimitSeconds: 90, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(`[]`), IsPublished: true},
	}
	toolSpecs := []struct {
		id           string
		name         string
		category     models.ToolCategory
		rarity       models.ToolRarity
		minimumLevel int16
		price        *int64
		introducedIn *string
	}{
		{"iron-pipe", "Iron pipe", models.ToolCategoryWeapon, models.ToolRarityCommon, 1, pricePointer(300), nil},
		{"scrap-shield", "Scrap shield", models.ToolCategoryGuard, models.ToolRarityCommon, 2, pricePointer(250), nil},
		{"bike-chain", "Bike chain", models.ToolCategoryWeapon, models.ToolRarityUncommon, 2, pricePointer(650), &levelID},
		{"riot-shield", "Riot shield", models.ToolCategoryGuard, models.ToolRarityRare, 3, pricePointer(1200), nil},
		{"pipe-wrench", "Pipe wrench", models.ToolCategoryWeapon, models.ToolRarityCommon, 3, pricePointer(400), nil},
		{"kings-crowbar", "King's crowbar", models.ToolCategoryWeapon, models.ToolRarityLegendary, 5, nil, nil},
	}
	for _, toolSpec := range toolSpecs {
		records = append(records, &models.ToolType{ID: toolSpec.id, Name: toolSpec.name, Description: "x", Category: toolSpec.category, Rarity: toolSpec.rarity, BaseStats: database.JSONDocument(`{"damage_bonus":4}`), MasteryCurve: database.JSONDocument(`{}`), MaxMasteryLevel: 10, MinimumFighterLevel: toolSpec.minimumLevel, ShopPriceCoins: toolSpec.price, IsTradeable: true, IntroducedInLevelID: toolSpec.introducedIn, ImageSVG: &svgMarkup})
	}
	records = append(records, &models.Tool{ID: uuid.New(), ToolTypeID: "iron-pipe", OwnerUserID: playerID, MasteryLevel: 1, Status: models.ToolStatusOwned, AcquiredFrom: models.ToolSourceReward, AcquiredAt: activatedAt})
	for _, record := range records {
		if err := testDatabase.Create(record).Error; err != nil {
			t.Fatalf("create %T: %v", record, err)
		}
	}

	standing := &fakePlayerStanding{fighterLevel: 2, wonLevelIDs: map[string]bool{}}
	storyModule := story.NewModule(story.ModuleDependencies{Database: testDatabase})
	toolsModule := tools.NewModule(tools.ModuleDependencies{Database: testDatabase, PlayerStanding: standing, Levels: storyModule.LevelCatalog})
	quietLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router, err := httpserver.NewRouter(httpserver.RouterOptions{Logger: quietLogger, AccessTokens: fixedPlayer{userID: playerID}}, toolsModule.RouteRegistrars()...)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return &shopHarness{t: t, database: testDatabase, router: router, playerID: playerID, standing: standing}
}

func pricePointer(price int64) *int64 {
	return &price
}

func (harness *shopHarness) list(query string) (int, shopListBody) {
	harness.t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/shop-items"+query, nil)
	request.Header.Set("Authorization", "Bearer "+playerToken)
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	var body shopListBody
	_ = json.Unmarshal(recorder.Body.Bytes(), &body)
	return recorder.Code, body
}

func shopItemIDs(body shopListBody) []string {
	var toolTypeIDs []string
	for _, shopItem := range body.Data {
		toolTypeIDs = append(toolTypeIDs, shopItem["tool_type"].(map[string]any)["id"].(string))
	}
	return toolTypeIDs
}

func TestTheShopListsToolsByMinimumLevelWithTwentyByDefault(t *testing.T) {
	harness := newShopHarness(t)
	status, body := harness.list("")
	if status != http.StatusOK || body.Pagination.Limit != 20 || body.Pagination.NextCursor != nil {
		t.Fatalf("got %d with pagination %+v", status, body.Pagination)
	}
	expectedOrder := []string{"iron-pipe", "bike-chain", "scrap-shield", "pipe-wrench", "riot-shield"}
	if got := shopItemIDs(body); strings.Join(got, ",") != strings.Join(expectedOrder, ",") {
		t.Fatalf("got %v, want %v; tools without a shop price stay out", got, expectedOrder)
	}
	ironPipe := body.Data[0]
	toolType := ironPipe["tool_type"].(map[string]any)
	if ironPipe["price_coins"] != "300" || ironPipe["owned_count"] != float64(1) || ironPipe["is_unlocked"] != true || ironPipe["is_usable"] != true || toolType["minimum_fighter_level"] != float64(1) || toolType["image_svg"] == nil {
		t.Fatalf("unexpected iron pipe %v", ironPipe)
	}
}

func TestTheShopPagesWithACursor(t *testing.T) {
	harness := newShopHarness(t)
	_, firstPage := harness.list("?limit=2")
	if len(firstPage.Data) != 2 || firstPage.Pagination.NextCursor == nil {
		t.Fatalf("unexpected first page %+v", firstPage.Pagination)
	}
	_, secondPage := harness.list("?limit=2&cursor=" + *firstPage.Pagination.NextCursor)
	_, thirdPage := harness.list("?limit=2&cursor=" + *secondPage.Pagination.NextCursor)
	allIDs := append(append(shopItemIDs(firstPage), shopItemIDs(secondPage)...), shopItemIDs(thirdPage)...)
	if len(allIDs) != 5 || thirdPage.Pagination.NextCursor != nil {
		t.Fatalf("got %v over three pages", allIDs)
	}
	if status, body := harness.list("?cursor=not-a-cursor"); status != http.StatusUnprocessableEntity || body.Error.Details["cursor"] == "" {
		t.Fatalf("a broken cursor got %d %+v", status, body.Error)
	}
	if status, _ := harness.list("?limit=101"); status != http.StatusUnprocessableEntity {
		t.Fatalf("a limit above 100 got %d", status)
	}
}

func TestTheShopSearchesByNameAndFilters(t *testing.T) {
	harness := newShopHarness(t)
	if _, body := harness.list("?q=PIPE"); strings.Join(shopItemIDs(body), ",") != "iron-pipe,pipe-wrench" {
		t.Fatalf("q=PIPE got %v", shopItemIDs(body))
	}
	if _, body := harness.list("?q=%25"); len(body.Data) != 0 {
		t.Fatalf("a percent sign must be searched literally, got %v", shopItemIDs(body))
	}
	if _, body := harness.list("?category=GUARD"); strings.Join(shopItemIDs(body), ",") != "scrap-shield,riot-shield" {
		t.Fatalf("category=GUARD got %v", shopItemIDs(body))
	}
	if _, body := harness.list("?q=shield&rarity=RARE"); strings.Join(shopItemIDs(body), ",") != "riot-shield" {
		t.Fatalf("q and rarity got %v", shopItemIDs(body))
	}
	if status, body := harness.list("?category=HAT"); status != http.StatusUnprocessableEntity || body.Error.Details["category"] == "" {
		t.Fatalf("an unknown category got %d", status)
	}
}

func TestUnlockingAndUsingDependOnThePlayer(t *testing.T) {
	harness := newShopHarness(t)
	_, body := harness.list("?q=bike")
	bikeChain := body.Data[0]
	unlockLevel, _ := bikeChain["unlock_level"].(map[string]any)
	if bikeChain["is_unlocked"] != false || unlockLevel["id"] != "2-1" || unlockLevel["chapter_number"] != float64(2) || unlockLevel["title"] != "The toll" {
		t.Fatalf("the bike chain must be locked behind level 2-1, got %v", bikeChain)
	}
	harness.standing.wonLevelIDs["2-1"] = true
	if _, body := harness.list("?q=bike"); body.Data[0]["is_unlocked"] != true {
		t.Fatal("winning level 2-1 must unlock the bike chain")
	}

	if _, body := harness.list("?q=riot"); body.Data[0]["is_usable"] != false {
		t.Fatal("a fighter at level 2 must not be able to use a level 3 tool")
	}
	harness.standing.fighterLevel = 3
	if _, body := harness.list("?q=riot"); body.Data[0]["is_usable"] != true {
		t.Fatal("a fighter at level 3 can use the riot shield")
	}
}
