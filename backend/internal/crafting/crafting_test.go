package crafting_test

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

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/crafting"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/notify/inapp"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/users"
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

func TestCraftingBurnsInputsNowAndDeliversOutputsLater(t *testing.T) {
	pool := databasetest.NewPool(t)
	clock := &testClock{currentTime: time.Now()}
	accessTokens, _ := accesstoken.NewManager("crafting-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	service := crafting.NewService(pool, catalog.NewService(catalog.NewLoader(pool), time.Minute, time.Now), inapp.NewNotifier(inapp.Dependencies{Users: users.NewRepository()}), clock.now)
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		crafting.NewHandler(service, users.NewAccountGuard(pool, users.NewRepository()).RequireStatus(users.StatusActive)),
	)
	resolver := crafting.NewResolverJob(service, testLogger)

	send := func(player uuid.UUID, method, path, body string) (int, map[string]any) {
		accessToken, _ := accessTokens.Issue(player)
		apiRequest := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
		apiRequest.Header.Set("Authorization", "Bearer "+accessToken.Value)
		apiRequest.Header.Set("X-Nebula-Client", "web")
		apiRequest.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, apiRequest)
		var decoded map[string]any
		json.Unmarshal(recorder.Body.Bytes(), &decoded)
		return recorder.Code, decoded
	}
	codeOf := func(decoded map[string]any) string {
		errorBody, _ := decoded["error"].(map[string]any)
		code, _ := errorBody["code"].(string)
		return code
	}

	crafter := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerItem(crafter, 1), 11)
	ledgertest.Fund(t, pool, ledger.PlayerItem(crafter, 2), 4)
	ledgertest.Fund(t, pool, ledger.PlayerNC(crafter, ledger.BucketCard), 100_000)

	status, started := send(crafter, http.MethodPost, "/crafts", `{"recipe_id":"alloy-plate","quantity":2}`)
	if status != http.StatusCreated {
		t.Fatalf("start craft: %d %v", status, started)
	}
	craftJob := started["data"].(map[string]any)
	if craftJob["fee"] != "40000" || craftJob["output_quantity"].(float64) != 2 {
		t.Fatalf("craft job %v", craftJob)
	}
	for itemID, expected := range map[int]int64{1: 1, 2: 0} {
		if balance := ledgertest.BalanceOf(t, pool, ledger.PlayerItem(crafter, itemID)); balance.Available != expected {
			t.Fatalf("item %d left %d, want %d", itemID, balance.Available, expected)
		}
	}
	if card := ledgertest.BalanceOf(t, pool, ledger.PlayerNC(crafter, ledger.BucketCard)); card.Available != 60_000 {
		t.Fatalf("fee not charged, card %d", card.Available)
	}

	if status, decoded := send(crafter, http.MethodPost, "/crafts", `{"recipe_id":"alloy-plate","quantity":1}`); status != http.StatusUnprocessableEntity || codeOf(decoded) != "LIMIT_EXCEEDED" {
		t.Fatalf("second craft while busy: %d %v", status, decoded)
	}

	resolver.Run(context.Background())
	if alloy := ledgertest.BalanceOf(t, pool, ledger.PlayerItem(crafter, 101)); alloy.Available != 0 {
		t.Fatal("outputs must wait for the timer")
	}
	clock.advance(2 * time.Minute)
	resolver.Run(context.Background())
	if alloy := ledgertest.BalanceOf(t, pool, ledger.PlayerItem(crafter, 101)); alloy.Available != 2 {
		t.Fatalf("alloy plates %d", alloy.Available)
	}
	resolver.Run(context.Background())
	if alloy := ledgertest.BalanceOf(t, pool, ledger.PlayerItem(crafter, 101)); alloy.Available != 2 {
		t.Fatalf("delivered twice: %d", alloy.Available)
	}

	_, listed := send(crafter, http.MethodGet, "/crafts?status=DELIVERED", "")
	if len(listed["data"].([]any)) != 1 {
		t.Fatalf("list %v", listed)
	}

	poorCrafter := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerItem(poorCrafter, 2), 3)
	ledgertest.Fund(t, pool, ledger.PlayerNC(poorCrafter, ledger.BucketCard), 1_000_000)
	if status, decoded := send(poorCrafter, http.MethodPost, "/crafts", `{"recipe_id":"circuit","quantity":1}`); status != http.StatusConflict || codeOf(decoded) != "INSUFFICIENT_ITEMS" {
		t.Fatalf("missing crystal: %d %v", status, decoded)
	}
	if status, _ := send(poorCrafter, http.MethodGet, "/crafts?status=CRAFTING", ""); status != http.StatusOK {
		t.Fatalf("list status %d", status)
	}
	if copper := ledgertest.BalanceOf(t, pool, ledger.PlayerItem(poorCrafter, 2)); copper.Available != 3 {
		t.Fatal("a failed craft must not burn anything")
	}
	if status, decoded := send(poorCrafter, http.MethodPost, "/crafts", `{"recipe_id":"warp-core","quantity":1}`); status != http.StatusUnprocessableEntity {
		t.Fatalf("unknown recipe: %d %v", status, decoded)
	}
	var craftNotices int
	pool.QueryRow(context.Background(), "SELECT count(*) FROM notifications WHERE user_id = $1 AND kind = 'craft_completed'", crafter).Scan(&craftNotices)
	if craftNotices != 1 {
		t.Fatalf("craft notifications %d, want 1", craftNotices)
	}
	ledgertest.RequireIntegrity(t, pool)
}
