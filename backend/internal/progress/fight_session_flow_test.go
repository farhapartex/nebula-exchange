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
	progressservice "github.com/farhapartex/nebula-exchange/backend/internal/progress/service"
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

const (
	boyStats          = `{"max_health":100,"punch":{"damage":8,"windup_ms":90,"active_ms":80,"recovery_ms":170},"kick":{"damage":14,"windup_ms":200,"active_ms":100,"recovery_ms":280}}`
	torchBanditStats  = `{"max_health":90,"punch":{"damage":9,"windup_ms":260,"active_ms":90,"recovery_ms":260},"kick":{"damage":15,"windup_ms":380,"active_ms":110,"recovery_ms":340}}`
	levelOneStarRules = `[{"stars":3,"within_seconds":45},{"stars":2,"within_seconds":70},{"stars":1,"within_seconds":90}]`
)

type testClock struct {
	currentTime time.Time
}

func (clock *testClock) now() time.Time {
	return clock.currentTime
}

func (clock *testClock) advance(duration time.Duration) {
	clock.currentTime = clock.currentTime.Add(duration)
}

type progressHarness struct {
	clock          *testClock
	t              *testing.T
	database       *gorm.DB
	router         *gin.Engine
	playerProgress interface {
		PlayerProgress(ctx context.Context, userID uuid.UUID) (progressservice.PlayerProgress, error)
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
		seedRecords = append(seedRecords, &storymodels.Level{ID: level.id, ChapterID: textPointer(level.chapterID), Number: numberPointer(level.number), Kind: storymodels.LevelKindStory, Title: level.id, Teaser: "x", ArenaID: "burning-house", TimeLimitSeconds: 90, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(levelOneStarRules), IsPublished: level.isPublished})
	}
	seedRecords = append(seedRecords,
		&storymodels.Enemy{ID: "torch-bandit", Name: "Torch bandit", Title: "First of the gang", Stats: database.JSONDocument(torchBanditStats), Brain: database.JSONDocument(`{"aggression":0.5}`), Look: database.JSONDocument(`{"body_color":"#7f1d1d"}`)},
		&storymodels.LevelEnemy{LevelID: "1-1", Wave: 1, EnemyID: "torch-bandit", Modifiers: database.JSONDocument(`{}`)},
		&storymodels.LevelEnemy{LevelID: "1-2", Wave: 1, EnemyID: "torch-bandit", Modifiers: database.JSONDocument(`{}`)},
		&storymodels.LevelEnemy{LevelID: "2-1", Wave: 1, EnemyID: "torch-bandit", Modifiers: database.JSONDocument(`{}`)},
		&progressmodels.FighterTemplate{ID: "the-boy", Name: "The boy", Title: "Nobody", StartingLevel: 1, Stats: database.JSONDocument(boyStats), Look: database.JSONDocument(`{"body_color":"#e7d3c1"}`), IsDefault: true},
	)
	seedRecords = append(seedRecords, &storymodels.Level{ID: "training", Kind: storymodels.LevelKindTraining, Title: "Training", Teaser: "x", ArenaID: "burning-house", TimeLimitSeconds: 90, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(`[]`), IsPublished: true})
	for _, record := range seedRecords {
		if err := testDatabase.Create(record).Error; err != nil {
			t.Fatalf("seed %T: %v", record, err)
		}
	}

	quietLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	clock := &testClock{currentTime: time.Now().UTC()}
	storyModule := story.NewModule(story.ModuleDependencies{Database: testDatabase})
	progressModule := progress.NewModule(progress.ModuleDependencies{Database: testDatabase, Levels: storyModule.LevelCatalog, Logger: quietLogger, Now: clock.now})
	router, err := httpserver.NewRouter(httpserver.RouterOptions{Logger: quietLogger, AccessTokens: playerTokens{"player-token": playerID, "stranger-token": uuid.New()}}, progressModule.RouteRegistrars()...)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return &progressHarness{clock: clock, t: t, database: testDatabase, router: router, playerProgress: progressModule.PlayerProgress, playerID: playerID}
}

func (harness *progressHarness) getFightSetup(levelID string) (int, map[string]any) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/levels/"+levelID+"/fight-setup", nil)
	request.Header.Set("Authorization", "Bearer player-token")
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	var decodedBody map[string]any
	_ = json.Unmarshal(recorder.Body.Bytes(), &decodedBody)
	return recorder.Code, decodedBody
}

func (harness *progressHarness) getNextLevel() map[string]any {
	harness.t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me/next-level", nil)
	request.Header.Set("Authorization", "Bearer player-token")
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	var decodedBody map[string]any
	_ = json.Unmarshal(recorder.Body.Bytes(), &decodedBody)
	if recorder.Code != http.StatusOK {
		harness.t.Fatalf("next level: got %d %v", recorder.Code, decodedBody)
	}
	return decodedBody["data"].(map[string]any)
}

func (harness *progressHarness) winWithAResult(levelID string) {
	harness.t.Helper()
	fightID := harness.startedFightID(levelID)
	harness.clock.advance(42 * time.Second)
	if status, body := harness.submitResult(fightID, `{"outcome":"WON","duration_ms":40000,"damage_dealt":96,"damage_taken":30}`, "player-token"); status != http.StatusCreated {
		harness.t.Fatalf("win %s: got %d %v", levelID, status, body)
	}
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

func (harness *progressHarness) submitResult(fightSessionID string, body string, accessToken string) (int, map[string]any) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/fight-sessions/"+fightSessionID+"/results", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Nebula-Client", "web")
	request.Header.Set("Authorization", "Bearer "+accessToken)
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	var decodedBody map[string]any
	_ = json.Unmarshal(recorder.Body.Bytes(), &decodedBody)
	return recorder.Code, decodedBody
}

func (harness *progressHarness) startedFightID(levelID string) string {
	harness.t.Helper()
	return harness.requireStart(levelID, http.StatusCreated)["data"].(map[string]any)["id"].(string)
}

func (harness *progressHarness) progressOf() progressservice.PlayerProgress {
	harness.t.Helper()
	playerProgress, err := harness.playerProgress.PlayerProgress(context.Background(), harness.playerID)
	if err != nil {
		harness.t.Fatalf("player progress: %v", err)
	}
	return playerProgress
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
	if newPlayer := harness.progressOf(); newPlayer.CurrentLevel != 1 || newPlayer.StoryLevel != 0 || newPlayer.NextLevel.Level.LevelID != "1-1" {
		t.Fatalf("a new player must start in chapter 1 with no level won, got %+v", newPlayer)
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
	if startedPlayer := harness.progressOf(); startedPlayer.StoryLevel != 0 || startedPlayer.NextLevel.Attempts != 1 {
		t.Fatalf("starting a level must not count as winning it, got %+v", startedPlayer)
	}
}

func TestPressingPlayAgainAbandonsTheUnfinishedFight(t *testing.T) {
	harness := newProgressHarness(t)
	firstFight := harness.requireStart("1-1", http.StatusCreated)["data"].(map[string]any)
	harness.clock.advance(time.Minute)
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

	if playerProgress := harness.progressOf(); playerProgress.CurrentLevel != 2 || playerProgress.StoryLevel != 0 || playerProgress.NextLevel.Level.LevelID != "2-1" {
		t.Fatalf("after chapter 1 the player must be in chapter 2, got %+v", playerProgress)
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

func TestFightSetupUsesTheStarterFighterUntilTheFirstFightCreatesAProfile(t *testing.T) {
	harness := newProgressHarness(t)

	status, body := harness.getFightSetup("1-1")
	if status != http.StatusOK {
		t.Fatalf("got %d %v", status, body)
	}
	fightSetup := body["data"].(map[string]any)
	player := fightSetup["player"].(map[string]any)
	enemy := fightSetup["enemy"].(map[string]any)
	arena := fightSetup["arena"].(map[string]any)
	if fightSetup["level_id"] != "1-1" || fightSetup["time_limit_seconds"] != float64(90) || arena["width"] != float64(960) {
		t.Fatalf("unexpected fight setup %v", fightSetup)
	}
	if player["id"] != "the-boy" || player["stats"].(map[string]any)["max_health"] != float64(100) || enemy["id"] != "torch-bandit" || enemy["brain"] == nil {
		t.Fatalf("unexpected fighters %v / %v", player, enemy)
	}
	var profileCount int64
	harness.database.Model(&progressmodels.FighterProfile{}).Count(&profileCount)
	if profileCount != 0 {
		t.Fatal("reading the fight setup must not create a profile")
	}

	harness.requireStart("1-1", http.StatusCreated)
	var profile progressmodels.FighterProfile
	if err := harness.database.Take(&profile, "user_id = ?", harness.playerID).Error; err != nil {
		t.Fatalf("expected a profile after the first fight: %v", err)
	}
	if profile.TemplateID != "the-boy" || profile.FighterLevel != 1 || len(profile.Stats) == 0 {
		t.Fatalf("unexpected profile %+v (stats %s)", profile, profile.Stats)
	}
	playerProgress, err := harness.playerProgress.PlayerProgress(context.Background(), harness.playerID)
	if err != nil || playerProgress.CurrentLevel != 1 || playerProgress.TotalWins != 0 || playerProgress.TotalLosses != 0 {
		t.Fatalf("unexpected player progress %+v, %v", playerProgress, err)
	}
}

func TestFightSetupNeedsAPlayableLevelWithAnEnemy(t *testing.T) {
	harness := newProgressHarness(t)
	enemylessLevel := &storymodels.Level{ID: "2-2", ChapterID: textPointer("2"), Number: numberPointer(2), Kind: storymodels.LevelKindStory, Title: "2-2", Teaser: "x", ArenaID: "burning-house", TimeLimitSeconds: 90, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(levelOneStarRules), IsPublished: true}
	if err := harness.database.Create(enemylessLevel).Error; err != nil {
		t.Fatalf("create enemy-less level: %v", err)
	}
	for _, levelID := range []string{"9-9", "1-3", "training", "2-2"} {
		if status, _ := harness.getFightSetup(levelID); status != http.StatusNotFound {
			t.Fatalf("%s: got %d, want 404", levelID, status)
		}
	}
}

func TestStartingAFightWithoutAStarterFighterIsRefused(t *testing.T) {
	harness := newProgressHarness(t)
	harness.database.Where("1 = 1").Delete(&progressmodels.FighterTemplate{})

	status, body := harness.startFight(`{"level":"1-1"}`, "player-token")
	if status != http.StatusServiceUnavailable || len(harness.sessionStatuses("1-1")) != 0 {
		t.Fatalf("got %d %v with sessions %v, want 503 and nothing saved", status, body, harness.sessionStatuses("1-1"))
	}
}

func TestAWinCountsStarsCompletesTheLevelAndUnlocksTheNextOne(t *testing.T) {
	harness := newProgressHarness(t)
	fightID := harness.startedFightID("1-1")
	harness.clock.advance(42 * time.Second)

	status, body := harness.submitResult(fightID, `{"outcome":"WON","duration_ms":40000,"damage_dealt":96,"damage_taken":30}`, "player-token")
	if status != http.StatusCreated {
		t.Fatalf("got %d %v", status, body)
	}
	result := body["data"].(map[string]any)
	if result["status"] != "FINISHED" || result["outcome"] != "WON" || result["stars"] != float64(3) || result["is_level_completed"] != true || result["reward_coins"] != "0" {
		t.Fatalf("unexpected result %v", result)
	}

	var levelProgress progressmodels.LevelProgress
	harness.database.Take(&levelProgress, "user_id = ? AND level_id = ?", harness.playerID, "1-1")
	if levelProgress.Status != progressmodels.ProgressStatusCompleted || levelProgress.BestStars == nil || *levelProgress.BestStars != 3 || levelProgress.FirstCompletedAt == nil {
		t.Fatalf("unexpected progress %+v", levelProgress)
	}
	if playerProgress := harness.progressOf(); playerProgress.TotalWins != 1 || playerProgress.TotalLosses != 0 || playerProgress.StoryLevel != 1 {
		t.Fatalf("unexpected progress after a win %+v", playerProgress)
	}
	harness.requireStart("1-2", http.StatusCreated)

	if status, _ := harness.submitResult(fightID, `{"outcome":"WON","duration_ms":40000,"damage_dealt":96,"damage_taken":30}`, "player-token"); status != http.StatusConflict {
		t.Fatalf("got %d for a second result, want 409", status)
	}
}

func TestLossesCountAndKeepTheBestStars(t *testing.T) {
	harness := newProgressHarness(t)
	firstFightID := harness.startedFightID("1-1")
	harness.clock.advance(62 * time.Second)
	harness.submitResult(firstFightID, `{"outcome":"WON","duration_ms":60000,"damage_dealt":91,"damage_taken":50}`, "player-token")

	timeUpFightID := harness.startedFightID("1-1")
	harness.clock.advance(95 * time.Second)
	status, body := harness.submitResult(timeUpFightID, `{"outcome":"LOST","duration_ms":90000,"damage_dealt":70,"damage_taken":60}`, "player-token")
	if status != http.StatusCreated || body["data"].(map[string]any)["stars"] != float64(0) || body["data"].(map[string]any)["is_level_completed"] != false {
		t.Fatalf("got %d %v for a time up loss", status, body)
	}

	knockoutFightID := harness.startedFightID("1-1")
	harness.clock.advance(31 * time.Second)
	harness.submitResult(knockoutFightID, `{"outcome":"LOST","duration_ms":30000,"damage_dealt":20,"damage_taken":100}`, "player-token")

	var levelProgress progressmodels.LevelProgress
	harness.database.Take(&levelProgress, "user_id = ? AND level_id = ?", harness.playerID, "1-1")
	if playerProgress := harness.progressOf(); playerProgress.TotalWins != 1 || playerProgress.TotalLosses != 2 || playerProgress.CurrentLevelWins != 1 || playerProgress.CurrentLevelLosses != 2 {
		t.Fatalf("unexpected win and loss counts %+v", playerProgress)
	}
	if levelProgress.Status != progressmodels.ProgressStatusCompleted || *levelProgress.BestStars != 2 || levelProgress.Attempts != 3 {
		t.Fatalf("losses must not undo the win: %+v", levelProgress)
	}
}

func TestImpossibleResultsAreRejectedAndNotCounted(t *testing.T) {
	harness := newProgressHarness(t)
	fightID := harness.startedFightID("1-1")
	harness.clock.advance(5 * time.Second)

	status, body := harness.submitResult(fightID, `{"outcome":"WON","duration_ms":2000,"damage_dealt":90,"damage_taken":0}`, "player-token")
	errorBody := body["error"].(map[string]any)
	if status != http.StatusUnprocessableEntity || errorBody["code"] != "FIGHT_RESULT_REJECTED" {
		t.Fatalf("got %d %v, want 422 FIGHT_RESULT_REJECTED", status, body)
	}
	var fightSession progressmodels.FightSession
	harness.database.Take(&fightSession, "id = ?", fightID)
	if fightSession.Status != progressmodels.FightStatusRejected || fightSession.RejectionReason == nil {
		t.Fatalf("unexpected rejected session %+v", fightSession)
	}
	if playerProgress := harness.progressOf(); playerProgress.TotalWins != 0 || playerProgress.TotalLosses != 0 {
		t.Fatalf("a rejected result must not count, got %+v", playerProgress)
	}
	harness.requireStart("1-2", http.StatusForbidden)
}

func TestResultsBelongToTheirOwnPlayerAndNeedAValidBody(t *testing.T) {
	harness := newProgressHarness(t)
	fightID := harness.startedFightID("1-1")
	harness.clock.advance(41 * time.Second)

	if status, _ := harness.submitResult(fightID, `{"outcome":"WON","duration_ms":40000,"damage_dealt":96,"damage_taken":30}`, "stranger-token"); status != http.StatusNotFound {
		t.Fatalf("got %d for another player's fight, want 404", status)
	}
	if status, _ := harness.submitResult("not-a-uuid", `{"outcome":"WON","duration_ms":1,"damage_dealt":1,"damage_taken":1}`, "player-token"); status != http.StatusNotFound {
		t.Fatalf("got %d for a malformed fight id, want 404", status)
	}
	for _, invalidBody := range []string{`{"outcome":"DRAW","duration_ms":40000,"damage_dealt":96,"damage_taken":30}`, `{"outcome":"WON"}`} {
		if status, _ := harness.submitResult(fightID, invalidBody, "player-token"); status != http.StatusUnprocessableEntity {
			t.Fatalf("got %d for %s, want 422", status, invalidBody)
		}
	}
}

func TestNextLevelFollowsTheChapterAndCountsWinsPerChapter(t *testing.T) {
	harness := newProgressHarness(t)

	firstLevel := harness.getNextLevel()
	if firstLevel["status"] != "AVAILABLE" || firstLevel["chapter"].(map[string]any)["number"] != float64(1) || firstLevel["chapter"].(map[string]any)["title"] != "The night they came" {
		t.Fatalf("unexpected first next level %v", firstLevel)
	}
	if level := firstLevel["level"].(map[string]any); level["id"] != "1-1" || level["number"] != float64(1) || level["time_limit_seconds"] != float64(90) || level["best_stars"] != nil {
		t.Fatalf("unexpected first level details %v", level)
	}

	harness.winWithAResult("1-1")
	afterFirstWin := harness.progressOf()
	if afterFirstWin.CurrentLevel != 1 || afterFirstWin.StoryLevel != 1 || afterFirstWin.CurrentLevelWins != 1 || harness.getNextLevel()["level"].(map[string]any)["id"] != "1-2" {
		t.Fatalf("after winning 1-1 the next level must be 1-2, got %+v", afterFirstWin)
	}

	harness.winWithAResult("1-2")
	inChapterTwo := harness.progressOf()
	if inChapterTwo.CurrentLevel != 2 || inChapterTwo.StoryLevel != 0 || inChapterTwo.CurrentLevelWins != 0 || inChapterTwo.TotalWins != 2 {
		t.Fatalf("chapter 2 must start with its own win count, got %+v", inChapterTwo)
	}

	harness.winWithAResult("2-1")
	comingSoon := harness.getNextLevel()
	if comingSoon["status"] != "COMING_SOON" || comingSoon["level"] != nil || comingSoon["chapter"].(map[string]any)["number"] != float64(3) || comingSoon["chapter"].(map[string]any)["title"] != nil {
		t.Fatalf("after the last level the next chapter must be coming soon, got %v", comingSoon)
	}
	if harness.progressOf().CurrentLevel != 3 {
		t.Fatal("finishing every chapter must move the player to the next chapter number")
	}
}

func TestAPaidChapterIsShownAsLockedAndCannotBeStarted(t *testing.T) {
	harness := newProgressHarness(t)
	err := harness.database.Model(&storymodels.Chapter{}).Where(map[string]any{"id": "2"}).
		Updates(map[string]any{"is_free": false, "price_coins": int64(50_000_000)}).Error
	if err != nil {
		t.Fatalf("make chapter 2 paid: %v", err)
	}
	harness.winWithAResult("1-1")
	harness.winWithAResult("1-2")

	lockedLevel := harness.getNextLevel()
	if lockedLevel["status"] != "LOCKED" || lockedLevel["level"].(map[string]any)["id"] != "2-1" || harness.progressOf().CurrentLevel != 2 {
		t.Fatalf("a paid chapter must show its first level as locked, got %v", lockedLevel)
	}
	status, body := harness.startFight(`{"level":"2-1"}`, "player-token")
	if status != http.StatusForbidden || body["error"].(map[string]any)["code"] != "CHAPTER_LOCKED" {
		t.Fatalf("got %d %v for a paid chapter, want 403 CHAPTER_LOCKED", status, body)
	}
}

func TestOnlyTheFreeLevelOfAPaidChapterCanBePlayedUntilTheChapterIsUnlocked(t *testing.T) {
	harness := newProgressHarness(t)
	if err := harness.database.Model(&storymodels.Chapter{}).Where(map[string]any{"id": "1"}).
		Updates(map[string]any{"is_free": false, "price_coins": int64(499_000_000)}).Error; err != nil {
		t.Fatalf("make chapter 1 paid: %v", err)
	}
	if err := harness.database.Model(&storymodels.Level{}).Where(map[string]any{"id": "1-1"}).Update("is_free", true).Error; err != nil {
		t.Fatalf("make level 1-1 free: %v", err)
	}

	newPlayer := harness.progressOf()
	if newPlayer.CurrentLevelPrice == nil || *newPlayer.CurrentLevelPrice != 499_000_000 || newPlayer.IsCurrentLevelPaid {
		t.Fatalf("a new player must see chapter 1 as unpaid with its price, got %+v", newPlayer)
	}
	firstLevel := harness.getNextLevel()
	firstChapter := firstLevel["chapter"].(map[string]any)
	if firstLevel["status"] != "AVAILABLE" || firstChapter["price"] != "499000000" || firstChapter["is_paid"] != false {
		t.Fatalf("the free first level must be playable in an unpaid chapter, got %v", firstLevel)
	}

	harness.winWithAResult("1-1")
	if lockedLevel := harness.getNextLevel(); lockedLevel["status"] != "LOCKED" || lockedLevel["level"].(map[string]any)["id"] != "1-2" {
		t.Fatalf("level 1-2 must be locked until chapter 1 is bought, got %v", lockedLevel)
	}
	status, body := harness.startFight(`{"level":"1-2"}`, "player-token")
	if status != http.StatusForbidden || body["error"].(map[string]any)["code"] != "CHAPTER_LOCKED" {
		t.Fatalf("got %d %v, want 403 CHAPTER_LOCKED", status, body)
	}

	unlock := &progressmodels.ChapterUnlock{UserID: harness.playerID, ChapterID: "1", Source: progressmodels.ChapterUnlockSourceGrant, UnlockedAt: harness.clock.now()}
	if err := harness.database.Create(unlock).Error; err != nil {
		t.Fatalf("unlock chapter 1: %v", err)
	}
	if paidPlayer := harness.progressOf(); !paidPlayer.IsCurrentLevelPaid {
		t.Fatalf("an unlocked chapter must count as paid, got %+v", paidPlayer)
	}
	if unlockedLevel := harness.getNextLevel(); unlockedLevel["status"] != "AVAILABLE" || unlockedLevel["chapter"].(map[string]any)["is_paid"] != true {
		t.Fatalf("level 1-2 must open after the unlock, got %v", unlockedLevel)
	}
	harness.requireStart("1-2", http.StatusCreated)
}
