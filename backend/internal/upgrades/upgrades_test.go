package upgrades_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/upgrades"
	"nebula-exchange/backend/internal/users"
)

func TestUpgradesSwapTiersByCraftingOrBuying(t *testing.T) {
	pool := databasetest.NewPool(t)
	accessTokens, _ := accesstoken.NewManager("upgrades-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		upgrades.NewHandler(
			upgrades.NewService(pool, catalog.NewService(catalog.NewLoader(pool), time.Minute, time.Now)),
			users.NewAccountGuard(pool, users.NewRepository()).RequireStatus(users.StatusActive),
		),
	)
	post := func(player uuid.UUID, path string) (int, map[string]any) {
		accessToken, _ := accessTokens.Issue(player)
		apiRequest := httptest.NewRequest(http.MethodPost, "/api/v1"+path, nil)
		apiRequest.Header.Set("Authorization", "Bearer "+accessToken.Value)
		apiRequest.Header.Set("X-Nebula-Client", "web")
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

	player := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerItem(player, 201), 2)
	ledgertest.Fund(t, pool, ledger.PlayerItem(player, 101), 4)
	ledgertest.Fund(t, pool, ledger.PlayerItem(player, 102), 2)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCard), 3_500_000)

	if status, decoded := post(player, "/upgrades/drill-t1-to-t2/crafts"); status != http.StatusConflict || codeOf(decoded) != "INSUFFICIENT_ITEMS" {
		t.Fatalf("missing an alloy plate: %d %v", status, decoded)
	}
	ledgertest.Fund(t, pool, ledger.PlayerItem(player, 101), 1)
	if status, decoded := post(player, "/upgrades/drill-t1-to-t2/crafts"); status != http.StatusCreated || decoded["data"].(map[string]any)["paid"] != "500000" {
		t.Fatalf("craft upgrade: %d %v", status, decoded)
	}
	for itemID, expected := range map[int]int64{201: 1, 202: 1, 101: 0, 102: 0} {
		if balance := ledgertest.BalanceOf(t, pool, ledger.PlayerItem(player, itemID)); balance.Available != expected {
			t.Fatalf("item %d: %d, want %d", itemID, balance.Available, expected)
		}
	}

	ledgertest.Hold(t, pool, ledger.PlayerItem(player, 201), 1)
	if status, decoded := post(player, "/upgrades/drill-t1-to-t2/purchases"); status != http.StatusConflict || codeOf(decoded) != "INSUFFICIENT_ITEMS" {
		t.Fatalf("a drill out on a mission can't be upgraded: %d %v", status, decoded)
	}
	if status, decoded := post(player, "/upgrades/drill-t2-to-t3/purchases"); status != http.StatusConflict || codeOf(decoded) != "INSUFFICIENT_FUNDS" {
		t.Fatalf("7 NC upgrade with 3 NC: %d %v", status, decoded)
	}
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCard), 4_000_000)
	if status, decoded := post(player, "/upgrades/drill-t2-to-t3/purchases"); status != http.StatusCreated {
		t.Fatalf("buy upgrade: %d %v", status, decoded)
	}
	if drill := ledgertest.BalanceOf(t, pool, ledger.PlayerItem(player, 203)); drill.Available != 1 {
		t.Fatalf("drill t3 %d", drill.Available)
	}
	if treasury := ledgertest.BalanceOf(t, pool, ledger.System(ledger.SystemTreasury, ledger.NC)); treasury.Available != 7_000_000 {
		t.Fatalf("treasury %d", treasury.Available)
	}

	if status, _ := post(player, "/upgrades/drill-t4-to-t5/purchases"); status != http.StatusUnprocessableEntity {
		t.Fatalf("T5 is not sold: %d", status)
	}
	if status, _ := post(player, "/upgrades/warp/crafts"); status != http.StatusNotFound {
		t.Fatalf("unknown upgrade: %d", status)
	}
	ledgertest.RequireIntegrity(t, pool)
}
