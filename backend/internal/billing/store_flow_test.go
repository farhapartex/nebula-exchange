package billing_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/billing"
	billingseeding "github.com/farhapartex/nebula-exchange/backend/internal/billing/seeding"
	identitymodels "github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress"
	progressmodels "github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/story"
	storymodels "github.com/farhapartex/nebula-exchange/backend/internal/story/models"
)

type fixedPlayer struct {
	userID uuid.UUID
}

func (player fixedPlayer) Verify(accessToken string) (uuid.UUID, error) {
	if accessToken != "player-token" {
		return uuid.Nil, errors.New("invalid token")
	}
	return player.userID, nil
}

type storeHarness struct {
	t        *testing.T
	database *gorm.DB
	router   *gin.Engine
	playerID uuid.UUID
}

func plansSeedFile() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "..", "..", "seeds", "plans", "plans.json")
}

func newStoreHarness(t *testing.T) *storeHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	testDatabase := databasetest.Open(t)
	quietLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	playerID := uuid.New()
	activatedAt := time.Now()
	price := int64(499)

	records := []any{
		&identitymodels.User{ID: playerID, Email: "boy@streetborn.test", Username: "the_boy", PasswordHash: "x", Status: identitymodels.UserStatusActive, IsActive: true, ActivatedAt: &activatedAt, TermsAcceptedAt: activatedAt},
		&storymodels.Arena{ID: "burning-house", Name: "The burning house", Width: 960, FloorY: 470, Stage: database.JSONDocument(`{}`)},
		&storymodels.Chapter{ID: "0", Number: 1, Title: "Prologue", Summary: "x", IsFree: true, IsPublished: true},
	}
	for chapterNumber := 2; chapterNumber <= 5; chapterNumber++ {
		chapterID := string(rune('0' + chapterNumber))
		levelID := chapterID + "-1"
		chapterPrice := price
		records = append(records,
			&storymodels.Chapter{ID: chapterID, Number: chapterNumber, Title: "Chapter " + chapterID, Summary: "x", IsFree: false, PriceCents: &chapterPrice, IsPublished: true},
			&storymodels.Level{ID: levelID, ChapterID: &chapterID, Number: func() *int { number := 1; return &number }(), Kind: storymodels.LevelKindStory, Title: levelID, Teaser: "x", ArenaID: "burning-house", TimeLimitSeconds: 90, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(`[]`), IsPublished: true},
		)
	}
	prologueChapterID := "0"
	records = append(records, &storymodels.Level{ID: "0-1", ChapterID: &prologueChapterID, Number: func() *int { number := 1; return &number }(), Kind: storymodels.LevelKindStory, Title: "0-1", Teaser: "x", ArenaID: "burning-house", TimeLimitSeconds: 90, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(`[]`), IsPublished: true, IsFree: true})
	for _, record := range records {
		if err := testDatabase.Create(record).Error; err != nil {
			t.Fatalf("seed %T: %v", record, err)
		}
	}
	plans, err := billingseeding.LoadPlans(plansSeedFile())
	if err != nil {
		t.Fatalf("load plans: %v", err)
	}
	if err := billingseeding.SeedPlans(context.Background(), testDatabase, plans, quietLogger); err != nil {
		t.Fatalf("seed plans: %v", err)
	}

	storyModule := story.NewModule(story.ModuleDependencies{Database: testDatabase})
	progressModule := progress.NewModule(progress.ModuleDependencies{Database: testDatabase, Levels: storyModule.LevelCatalog, Logger: quietLogger, Now: time.Now})
	billingModule := billing.NewModule(billing.ModuleDependencies{Database: testDatabase, Chapters: storyModule.LevelCatalog, ChapterOwnership: progressModule.ChapterOwnership})
	router, err := httpserver.NewRouter(httpserver.RouterOptions{Logger: quietLogger, AccessTokens: fixedPlayer{userID: playerID}}, billingModule.RouteRegistrars()...)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return &storeHarness{t: t, database: testDatabase, router: router, playerID: playerID}
}

type listBody struct {
	Data       []map[string]any `json:"data"`
	Pagination struct {
		NextCursor *string `json:"next_cursor"`
	} `json:"pagination"`
}

func (harness *storeHarness) get(path string, accessToken string) (int, listBody) {
	harness.t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/api/v1"+path, nil)
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	var body listBody
	_ = json.Unmarshal(recorder.Body.Bytes(), &body)
	return recorder.Code, body
}

func (harness *storeHarness) plansByID() map[string]map[string]any {
	harness.t.Helper()
	status, body := harness.get("/plans", "player-token")
	if status != http.StatusOK {
		harness.t.Fatalf("plans: got %d", status)
	}
	plans := map[string]map[string]any{}
	for _, plan := range body.Data {
		plans[plan["id"].(string)] = plan
	}
	return plans
}

func optionTotals(plan map[string]any) map[float64]string {
	totals := map[float64]string{}
	for _, option := range plan["options"].([]any) {
		optionData := option.(map[string]any)
		totals[optionData["chapter_count"].(float64)] = optionData["total_cents"].(string)
	}
	return totals
}

func TestChaptersListPricesAndOwnershipWithPagination(t *testing.T) {
	harness := newStoreHarness(t)

	status, firstPage := harness.get("/chapters?limit=3", "player-token")
	if status != http.StatusOK || len(firstPage.Data) != 3 || firstPage.Pagination.NextCursor == nil {
		t.Fatalf("got %d with %d chapters and cursor %v", status, len(firstPage.Data), firstPage.Pagination.NextCursor)
	}
	prologue, secondChapter := firstPage.Data[0], firstPage.Data[1]
	if prologue["is_free"] != true || prologue["price_cents"] != nil || secondChapter["price_cents"] != "499" || secondChapter["is_owned"] != false || secondChapter["level_count"] != float64(1) {
		t.Fatalf("unexpected chapters %v", firstPage.Data)
	}
	_, secondPage := harness.get("/chapters?limit=3&cursor="+*firstPage.Pagination.NextCursor, "player-token")
	if len(secondPage.Data) != 2 || secondPage.Data[0]["number"] != float64(4) || secondPage.Pagination.NextCursor != nil {
		t.Fatalf("unexpected second page %v", secondPage)
	}
	if status, _ := harness.get("/chapters", ""); status != http.StatusUnauthorized {
		t.Fatalf("got %d without a login, want 401", status)
	}
}

func TestPlansArePricedOnTheServerForTheChaptersStillToBuy(t *testing.T) {
	harness := newStoreHarness(t)
	plans := harness.plansByID()

	if len(plans) != 3 {
		t.Fatalf("got %d plans, want 3", len(plans))
	}
	if totals := optionTotals(plans["single-chapter"]); len(totals) != 1 || totals[1] != "499" {
		t.Fatalf("unexpected single chapter totals %v", totals)
	}
	if totals := optionTotals(plans["chapter-bundle"]); len(totals) != 2 || totals[2] != "948" || totals[3] != "1347" {
		t.Fatalf("unexpected bundle totals %v", totals)
	}
	allOption := plans["all-chapters"]["options"].([]any)[0].(map[string]any)
	if allOption["chapter_count"] != float64(4) || allOption["subtotal_cents"] != "1996" || allOption["discount_cents"] != "399" || allOption["total_cents"] != "1597" {
		t.Fatalf("unexpected all chapters option %v", allOption)
	}

	ownedChapter := &progressmodels.ChapterUnlock{UserID: harness.playerID, ChapterID: "2", Source: progressmodels.ChapterUnlockSourceGrant, UnlockedAt: time.Now()}
	if err := harness.database.Create(ownedChapter).Error; err != nil {
		t.Fatalf("own chapter 2: %v", err)
	}
	afterOwning := harness.plansByID()
	singleChapters := afterOwning["single-chapter"]["options"].([]any)[0].(map[string]any)["chapters"].([]any)
	if singleChapters[0].(map[string]any)["number"] != float64(3) {
		t.Fatalf("the single plan must now offer chapter 3, got %v", singleChapters)
	}
	if totals := optionTotals(afterOwning["chapter-bundle"]); len(totals) != 1 || totals[2] != "948" {
		t.Fatalf("with 3 chapters left the bundle can only be 2 chapters, got %v", totals)
	}
	if allOption := afterOwning["all-chapters"]["options"].([]any)[0].(map[string]any); allOption["chapter_count"] != float64(3) {
		t.Fatalf("all chapters must now cover 3 chapters, got %v", allOption)
	}

	for _, chapterID := range []string{"3", "4"} {
		harness.database.Create(&progressmodels.ChapterUnlock{UserID: harness.playerID, ChapterID: chapterID, Source: progressmodels.ChapterUnlockSourceGrant, UnlockedAt: time.Now()})
	}
	oneLeft := harness.plansByID()
	if oneLeft["single-chapter"]["is_available"] != true || oneLeft["chapter-bundle"]["is_available"] != false || oneLeft["all-chapters"]["is_available"] != false {
		t.Fatalf("with one chapter left only the single plan makes sense, got %v", oneLeft)
	}
}

func TestThePlansSeedFileIsValid(t *testing.T) {
	plans, err := billingseeding.LoadPlans(plansSeedFile())
	if err != nil || len(plans) != 3 {
		t.Fatalf("got %d plans, %v", len(plans), err)
	}
}
