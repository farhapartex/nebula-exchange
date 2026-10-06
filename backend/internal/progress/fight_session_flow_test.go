package progress_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	identitymodels "github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress"
	progressmodels "github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/story"
	storymodels "github.com/farhapartex/nebula-exchange/backend/internal/story/models"
)

type playerTokens map[string]uuid.UUID

func (tokens playerTokens) Verify(accessToken string) (uuid.UUID, error) {
	userID, isKnown := tokens[accessToken]
	if !isKnown {
		return uuid.Nil, errors.New("invalid token")
	}
	return userID, nil
}

type progressHarness struct {
	t             *testing.T
	database      *gorm.DB
	router        *gin.Engine
	storyProgress interface {
		FurthestStartedLevel(ctx context.Context, userID uuid.UUID) (*string, error)
	}
	playerID uuid.UUID
}

func textPointer(text string) *string {
	return &text
}

func numberPointer(number int) *int {
	return &number
}

func newProgressHarness(t *testing.T) *progressHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	testDatabase := databasetest.Open(t)
	playerID := uuid.New()
	activatedAt := time.Now()
	seedRecords := []any{
		&identitymodels.User{ID: playerID, Email: "boy@streetborn.test", Username: "the_boy", PasswordHash: "x", Status: identitymodels.UserStatusActive, IsActive: true, ActivatedAt: &activatedAt, TermsAcceptedAt: activatedAt},
		&storymodels.Chapter{ID: "1", Number: 1, Title: "The night they came", Summary: "x", IsFree: true, IsPublished: true},
		&storymodels.Chapter{ID: "2", Number: 2, Title: "The club", Summary: "x", IsFree: true, IsPublished: true},
		&storymodels.Arena{ID: "burning-house", Name: "The burning house", Width: 960, FloorY: 470, Stage: database.JSONDocument(`{}`)},
	}
	levels := []struct {
		id          string
		chapterID   string
		number      int
		isPublished bool
	}{{"1-1", "1", 1, true}, {"1-2", "1", 2, true}, {"1-3", "1", 3, false}, {"2-1", "2", 1, true}}
	for _, level := range levels {
		seedRecords = append(seedRecords, &storymodels.Level{ID: level.id, ChapterID: textPointer(level.chapterID), Number: numberPointer(level.number), Kind: storymodels.LevelKindStory, Title: level.id, Teaser: "x", ArenaID: "burning-house", TimeLimitSeconds: 90, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(`[]`), IsPublished: level.isPublished})
	}
	seedRecords = append(seedRecords, &storymodels.Level{ID: "training", Kind: storymodels.LevelKindTraining, Title: "Training", Teaser: "x", ArenaID: "burning-house", TimeLimitSeconds: 90, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(`[]`), IsPublished: true})
	for _, record := range seedRecords {
		if err := testDatabase.Create(record).Error; err != nil {
			t.Fatalf("seed %T: %v", record, err)
		}
	}

	quietLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	storyModule := story.NewModule(story.ModuleDependencies{Database: testDatabase})
	progressModule := progress.NewModule(progress.ModuleDependencies{Database: testDatabase, Levels: storyModule.LevelCatalog, Logger: quietLogger, Now: time.Now})
	router, err := httpserver.NewRouter(httpserver.RouterOptions{Logger: quietLogger, AccessTokens: playerTokens{"player-token": playerID}}, progressModule.RouteRegistrars()...)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return &progressHarness{t: t, database: testDatabase, router: router, storyProgress: progressModule.StoryProgress, playerID: playerID}
}

func (harness *progressHarness) startFight(body string, accessToken string) (int, map[string]any) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/fight-sessions", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Nebula-Client", "web")
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	var decodedBody map[string]any
	_ = json.Unmarshal(recorder.Body.Bytes(), &decodedBody)
	return recorder.Code, decodedBody
}

func (harness *progressHarness) requireStart(levelID string, expectedStatus int) map[string]any {
	harness.t.Helper()
	status, body := harness.startFight(`{"level":"`+levelID+`"}`, "player-token")
	if status != expectedStatus {
		harness.t.Fatalf("start %s: got %d %v, want %d", levelID, status, body, expectedStatus)
	}
	return body
}

func (harness *progressHarness) furthestLevel() *string {
	harness.t.Helper()
	furthestLevelID, err := harness.storyProgress.FurthestStartedLevel(context.Background(), harness.playerID)
	if err != nil {
		harness.t.Fatalf("furthest level: %v", err)
	}
	return furthestLevelID
}

func (harness *progressHarness) winLevel(levelID string) {
	harness.t.Helper()
	err := harness.database.Model(&progressmodels.LevelProgress{}).
		Where(map[string]any{"user_id": harness.playerID, "level_id": levelID}).
		Updates(map[string]any{"status": progressmodels.ProgressStatusCompleted, "first_completed_at": time.Now()}).Error
	if err != nil {
		harness.t.Fatalf("win %s: %v", levelID, err)
	}
}

func (harness *progressHarness) sessionStatuses(levelID string) []string {
	var statuses []string
	harness.database.Model(&progressmodels.FightSession{}).Where(map[string]any{"user_id": harness.playerID, "level_id": levelID}).Order("started_at").Pluck("status", &statuses)
	return statuses
}

func TestStartingTheFirstLevelRecordsAnOpenFightAndTheStoryLevel(t *testing.T) {
	harness := newProgressHarness(t)
	if harness.furthestLevel() != nil {
		t.Fatal("a new player must not have a story level yet")
	}

	startedFight := harness.requireStart("1-1", http.StatusCreated)["data"].(map[string]any)
	if startedFight["level"] != "1-1" || startedFight["status"] != "STARTED" || startedFight["seed"] == "" || startedFight["id"] == "" {
		t.Fatalf("unexpected started fight %v", startedFight)
	}

	var levelProgress progressmodels.LevelProgress
	harness.database.Take(&levelProgress, "user_id = ? AND level_id = ?", harness.playerID, "1-1")
	if levelProgress.Status != progressmodels.ProgressStatusStarted || levelProgress.Attempts != 1 {
		t.Fatalf("unexpected progress %+v", levelProgress)
	}
	if furthestLevel := harness.furthestLevel(); furthestLevel == nil || *furthestLevel != "1-1" {
		t.Fatalf("got story level %v, want 1-1", furthestLevel)
	}
}

func TestPressingPlayAgainAbandonsTheUnfinishedFight(t *testing.T) {
	harness := newProgressHarness(t)
	firstFight := harness.requireStart("1-1", http.StatusCreated)["data"].(map[string]any)
	secondFight := harness.requireStart("1-1", http.StatusCreated)["data"].(map[string]any)

	if firstFight["id"] == secondFight["id"] {
		t.Fatal("a second Play must start a new fight")
	}
	statuses := harness.sessionStatuses("1-1")
	if len(statuses) != 2 || statuses[0] != "ABANDONED" || statuses[1] != "STARTED" {
		t.Fatalf("got statuses %v, want ABANDONED then STARTED", statuses)
	}
	var levelProgress progressmodels.LevelProgress
	harness.database.Take(&levelProgress, "user_id = ? AND level_id = ?", harness.playerID, "1-1")
	if levelProgress.Attempts != 2 {
		t.Fatalf("got %d attempts, want 2", levelProgress.Attempts)
	}
}

func TestLevelsUnlockInOrderAcrossChapters(t *testing.T) {
	harness := newProgressHarness(t)

	lockedBody := harness.requireStart("1-2", http.StatusForbidden)
	if lockedBody["error"].(map[string]any)["code"] != "LEVEL_LOCKED" {
		t.Fatalf("unexpected locked error %v", lockedBody)
	}
	harness.requireStart("1-1", http.StatusCreated)
	harness.requireStart("1-2", http.StatusForbidden)

	harness.winLevel("1-1")
	harness.requireStart("1-2", http.StatusCreated)
	harness.requireStart("2-1", http.StatusForbidden)

	harness.winLevel("1-2")
	harness.requireStart("2-1", http.StatusCreated)
	harness.requireStart("1-1", http.StatusCreated)

	if furthestLevel := harness.furthestLevel(); furthestLevel == nil || *furthestLevel != "2-1" {
		t.Fatalf("got story level %v, want 2-1 even after replaying 1-1", furthestLevel)
	}
}

func TestStartingAFightRejectsBadRequests(t *testing.T) {
	harness := newProgressHarness(t)

	if status, _ := harness.startFight(`{"level":"1-1"}`, ""); status != http.StatusUnauthorized {
		t.Fatalf("got %d without a login, want 401", status)
	}
	if status, _ := harness.startFight(`{}`, "player-token"); status != http.StatusUnprocessableEntity {
		t.Fatalf("got %d without a level, want 422", status)
	}
	for _, unplayableLevel := range []string{"9-9", "1-3", "training"} {
		harness.requireStart(unplayableLevel, http.StatusNotFound)
	}
}
