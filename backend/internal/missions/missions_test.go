package missions_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/missions"
	"nebula-exchange/backend/internal/notify/inapp"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/users"
)

const (
	scoutID       = 301
	interceptorID = 304
	drillT1ID     = 201
	fuelID        = 401
)

type testClock struct {
	mutex       sync.Mutex
	currentTime time.Time
}

func (clock *testClock) now() time.Time {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()
	return clock.currentTime
}

func (clock *testClock) advance(duration time.Duration) {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()
	clock.currentTime = clock.currentTime.Add(duration)
}

type missionsHarness struct {
	pool         *pgxpool.Pool
	router       http.Handler
	accessTokens *accesstoken.Manager
	clock        *testClock
	resolver     *missions.ResolverJob
}

func newMissionsHarness(t *testing.T) *missionsHarness {
	t.Helper()
	pool := databasetest.NewPool(t)
	clock := &testClock{currentTime: time.Now()}
	accessTokens, _ := accesstoken.NewManager("missions-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	catalogService := catalog.NewService(catalog.NewLoader(pool), time.Minute, time.Now)
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		missions.NewHandler(missions.NewService(pool, catalogService, clock.now), users.NewAccountGuard(pool, users.NewRepository()).RequireStatus(users.StatusActive)),
		catalog.NewHandler(catalogService, missions.NewZoneAvailability(pool, catalogService)),
	)
	return &missionsHarness{
		pool:         pool,
		router:       router,
		accessTokens: accessTokens,
		clock:        clock,
		resolver:     missions.NewResolverJob(pool, inapp.NewNotifier(inapp.Dependencies{Users: users.NewRepository()}), testLogger, func(int) int { return 0 }, clock.now),
	}
}

func (harness *missionsHarness) equippedPlayer(t *testing.T, ships, fuel int64) uuid.UUID {
	t.Helper()
	player := ledgertest.CreatePlayer(t, harness.pool)
	ledgertest.Fund(t, harness.pool, ledger.PlayerItem(player, scoutID), ships)
	ledgertest.Fund(t, harness.pool, ledger.PlayerItem(player, drillT1ID), ships)
	ledgertest.Fund(t, harness.pool, ledger.PlayerItem(player, fuelID), fuel)
	return player
}

func (harness *missionsHarness) send(t *testing.T, player uuid.UUID, method, path, body string) (int, map[string]any) {
	t.Helper()
	accessToken, _ := harness.accessTokens.Issue(player)
	apiRequest := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
	apiRequest.Header.Set("Authorization", "Bearer "+accessToken.Value)
	apiRequest.Header.Set("X-Nebula-Client", "web")
	apiRequest.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, apiRequest)
	var decoded map[string]any
	json.Unmarshal(recorder.Body.Bytes(), &decoded)
	return recorder.Code, decoded
}

func (harness *missionsHarness) start(t *testing.T, player uuid.UUID, body string) (int, map[string]any) {
	t.Helper()
	return harness.send(t, player, http.MethodPost, "/missions", body)
}

func codeOf(decoded map[string]any) string {
	errorBody, _ := decoded["error"].(map[string]any)
	code, _ := errorBody["code"].(string)
	return code
}

func dataOf(decoded map[string]any) map[string]any {
	data, _ := decoded["data"].(map[string]any)
	return data
}

const asteroidRun = `{"zone_id":"asteroid-belt","ship_item_id":301,"drill_item_id":201}`

func TestMissionLoopFromLaunchToCollect(t *testing.T) {
	harness := newMissionsHarness(t)
	player := harness.equippedPlayer(t, 1, 10)

	status, started := harness.start(t, player, asteroidRun)
	if status != http.StatusCreated {
		t.Fatalf("start: %d %v", status, started)
	}
	mission := dataOf(started)
	missionID := mission["id"].(string)
	if mission["status"] != "RUNNING" || mission["fuel_spent"].(float64) != 2 {
		t.Fatalf("mission %v", mission)
	}
	if ship := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, scoutID)); ship != (ledgertest.Balance{Available: 0, Held: 1}) {
		t.Fatalf("ship should be held: %+v", ship)
	}
	if fuel := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, fuelID)); fuel.Available != 8 {
		t.Fatalf("fuel %d", fuel.Available)
	}

	if status, decoded := harness.start(t, player, asteroidRun); status != http.StatusConflict || codeOf(decoded) != "INSUFFICIENT_ITEMS" {
		t.Fatalf("a held ship can't fly twice: %d %v", status, decoded)
	}
	if status, decoded := harness.send(t, player, http.MethodPost, "/missions/"+missionID+"/collect", ""); status != http.StatusConflict {
		t.Fatalf("collect while running: %d %v", status, decoded)
	}

	harness.resolver.Run(context.Background())
	if _, fetched := harness.send(t, player, http.MethodGet, "/missions/"+missionID, ""); dataOf(fetched)["status"] != "RUNNING" {
		t.Fatal("the resolver must wait for the end time")
	}
	harness.clock.advance(15 * time.Minute)
	if err := harness.resolver.Run(context.Background()); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	_, resolved := harness.send(t, player, http.MethodGet, "/missions/"+missionID, "")
	resolvedLoot := dataOf(resolved)["loot"].([]any)
	if dataOf(resolved)["status"] != "COMPLETED" || len(resolvedLoot) != 2 {
		t.Fatalf("resolved %v", resolved)
	}

	status, collected := harness.send(t, player, http.MethodPost, "/missions/"+missionID+"/collect", "")
	if status != http.StatusOK || dataOf(collected)["status"] != "COLLECTED" {
		t.Fatalf("collect: %d %v", status, collected)
	}
	if iron := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, 1)); iron.Available != 8 {
		t.Fatalf("iron %d", iron.Available)
	}
	if copper := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, 2)); copper.Available != 3 {
		t.Fatalf("copper %d", copper.Available)
	}
	if ship := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, scoutID)); ship != (ledgertest.Balance{Available: 1, Held: 0}) {
		t.Fatalf("ship should be free again: %+v", ship)
	}
	if status, _ := harness.send(t, player, http.MethodPost, "/missions/"+missionID+"/collect", ""); status != http.StatusConflict {
		t.Fatalf("second collect: %d", status)
	}
	ledgertest.RequireIntegrity(t, harness.pool)
}

func TestAbortKeepsFuelBurnedAndFreesTheShip(t *testing.T) {
	harness := newMissionsHarness(t)
	player := harness.equippedPlayer(t, 1, 10)
	_, started := harness.start(t, player, asteroidRun)
	missionID := dataOf(started)["id"].(string)

	status, aborted := harness.send(t, player, http.MethodPost, "/missions/"+missionID+"/abort", "")
	if status != http.StatusOK || dataOf(aborted)["status"] != "ABORTED" {
		t.Fatalf("abort: %d %v", status, aborted)
	}
	if fuel := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, fuelID)); fuel.Available != 8 {
		t.Fatalf("fuel is not refunded, got %d", fuel.Available)
	}
	if ship := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, scoutID)); ship.Available != 1 {
		t.Fatalf("ship %+v", ship)
	}

	_, second := harness.start(t, player, asteroidRun)
	harness.clock.advance(16 * time.Minute)
	if status, _ := harness.send(t, player, http.MethodPost, "/missions/"+dataOf(second)["id"].(string)+"/abort", ""); status != http.StatusConflict {
		t.Fatalf("a finished mission can't be aborted: %d", status)
	}
	ledgertest.RequireIntegrity(t, harness.pool)
}

func TestMissionRequirementsAndLimits(t *testing.T) {
	harness := newMissionsHarness(t)
	player := harness.equippedPlayer(t, 4, 40)

	cases := []struct {
		body         string
		expectedCode string
	}{
		{`{"zone_id":"crystal-moon","ship_item_id":301,"drill_item_id":201}`, "VALIDATION_FAILED"},
		{`{"zone_id":"nowhere","ship_item_id":301,"drill_item_id":201}`, "VALIDATION_FAILED"},
		{`{"zone_id":"asteroid-belt","ship_item_id":201,"drill_item_id":301}`, "VALIDATION_FAILED"},
		{`{"zone_id":"asteroid-belt","ship_item_id":302,"drill_item_id":201}`, "INSUFFICIENT_ITEMS"},
	}
	for _, testCase := range cases {
		if _, decoded := harness.start(t, player, testCase.body); codeOf(decoded) != testCase.expectedCode {
			t.Fatalf("%s: got %v, want %s", testCase.body, decoded, testCase.expectedCode)
		}
	}

	for missionIndex := 0; missionIndex < 3; missionIndex++ {
		if status, decoded := harness.start(t, player, asteroidRun); status != http.StatusCreated {
			t.Fatalf("mission %d: %d %v", missionIndex, status, decoded)
		}
	}
	if status, decoded := harness.start(t, player, asteroidRun); status != http.StatusUnprocessableEntity || codeOf(decoded) != "LIMIT_EXCEEDED" {
		t.Fatalf("fourth mission: %d %v", status, decoded)
	}

	_, listed := harness.send(t, player, http.MethodGet, "/missions?status=RUNNING&limit=2", "")
	if len(listed["data"].([]any)) != 2 || listed["pagination"].(map[string]any)["next_cursor"] == nil {
		t.Fatalf("list %v", listed)
	}
	if status, _ := harness.send(t, player, http.MethodGet, "/missions?status=FLYING", ""); status != http.StatusUnprocessableEntity {
		t.Fatalf("bad status filter: %d", status)
	}
}

func TestInterceptorFlightsAreFaster(t *testing.T) {
	harness := newMissionsHarness(t)
	player := ledgertest.CreatePlayer(t, harness.pool)
	ledgertest.Fund(t, harness.pool, ledger.PlayerItem(player, interceptorID), 1)
	ledgertest.Fund(t, harness.pool, ledger.PlayerItem(player, drillT1ID), 1)
	ledgertest.Fund(t, harness.pool, ledger.PlayerItem(player, fuelID), 10)

	_, started := harness.start(t, player, `{"zone_id":"asteroid-belt","ship_item_id":304,"drill_item_id":201}`)
	mission := dataOf(started)
	startedAt, _ := time.Parse(time.RFC3339Nano, mission["started_at"].(string))
	endsAt, _ := time.Parse(time.RFC3339Nano, mission["ends_at"].(string))
	if flight := endsAt.Sub(startedAt); flight != 9*time.Minute {
		t.Fatalf("interceptor flight %s, want 9m", flight)
	}
}

func TestZonesShowWhatThePlayerCanRun(t *testing.T) {
	harness := newMissionsHarness(t)
	player := harness.equippedPlayer(t, 1, 3)

	_, zones := harness.send(t, player, http.MethodGet, "/zones", "")
	unlockByZone := map[string]map[string]any{}
	for _, zone := range zones["data"].([]any) {
		zoneData := zone.(map[string]any)
		unlockByZone[zoneData["id"].(string)] = zoneData["unlock"].(map[string]any)
	}
	if unlockByZone["asteroid-belt"]["is_unlocked"] != true {
		t.Fatalf("asteroid belt %v", unlockByZone["asteroid-belt"])
	}
	crystalMoon := unlockByZone["crystal-moon"]
	if crystalMoon["is_unlocked"] != false || crystalMoon["has_required_drill"] != false || crystalMoon["has_enough_fuel"] != false {
		t.Fatalf("crystal moon %v", crystalMoon)
	}
	if unlockByZone["plasma-nebula"]["has_allowed_ship"] != false {
		t.Fatal("a Scout can't fly to the Plasma Nebula")
	}

	guestRecorder := httptest.NewRecorder()
	harness.router.ServeHTTP(guestRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/zones", nil))
	if !strings.Contains(guestRecorder.Body.String(), `"unlock":null`) {
		t.Fatalf("guests get no unlock info: %s", guestRecorder.Body.String())
	}
}

func TestParallelLaunchesRespectTheMissionLimit(t *testing.T) {
	harness := newMissionsHarness(t)
	player := harness.equippedPlayer(t, 8, 40)

	var waitGroup sync.WaitGroup
	statusCodes := make(chan int, 8)
	for attempt := 0; attempt < 8; attempt++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			status, _ := harness.start(t, player, asteroidRun)
			statusCodes <- status
		}()
	}
	waitGroup.Wait()
	close(statusCodes)

	launched := 0
	for status := range statusCodes {
		if status == http.StatusCreated {
			launched++
		} else if status != http.StatusUnprocessableEntity {
			t.Fatalf("unexpected status %d", status)
		}
	}
	if launched != missions.MaximumActiveMissions {
		t.Fatalf("%d missions launched, want %d", launched, missions.MaximumActiveMissions)
	}
	if fuel := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, fuelID)); fuel.Available != 40-2*int64(missions.MaximumActiveMissions) {
		t.Fatalf("fuel burned only for launched missions, left %d", fuel.Available)
	}
	ledgertest.RequireIntegrity(t, harness.pool)
}
