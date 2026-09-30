package shop_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/shop"
	"nebula-exchange/backend/internal/users"
)

func TestBuyingWithBalanceChargesSpendingOrderAndMintsItems(t *testing.T) {
	pool := databasetest.NewPool(t)
	accessTokens, _ := accesstoken.NewManager("shop-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	userRepository := users.NewRepository()
	catalogService := catalog.NewService(catalog.NewLoader(pool), time.Minute, time.Now)
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		shop.NewHandler(shop.NewService(pool, catalogService), users.NewAccountGuard(pool, userRepository).RequireStatus(users.StatusActive)),
	)

	player := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCard), 300_000)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketEarned), 1_000_000)
	accessToken, _ := accessTokens.Issue(player)

	buy := func(sku, body string) (int, map[string]any) {
		purchaseRequest := httptest.NewRequest(http.MethodPost, "/api/v1/shop-items/"+sku+"/purchases", strings.NewReader(body))
		purchaseRequest.Header.Set("Authorization", "Bearer "+accessToken.Value)
		purchaseRequest.Header.Set("X-Nebula-Client", "web")
		purchaseRequest.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, purchaseRequest)
		var decoded map[string]any
		json.Unmarshal(recorder.Body.Bytes(), &decoded)
		return recorder.Code, decoded
	}

	status, decoded := buy("fuel-cell", `{"quantity":5}`)
	if status != http.StatusCreated || decoded["data"].(map[string]any)["total"] != "1000000" {
		t.Fatalf("buy fuel: %d %v", status, decoded)
	}
	if fuel := ledgertest.BalanceOf(t, pool, ledger.PlayerItem(player, 401)); fuel.Available != 5 {
		t.Fatalf("fuel %d", fuel.Available)
	}
	if card := ledgertest.BalanceOf(t, pool, ledger.PlayerNC(player, ledger.BucketCard)); card.Available != 0 {
		t.Fatalf("card is spent first, left %d", card.Available)
	}
	if earned := ledgertest.BalanceOf(t, pool, ledger.PlayerNC(player, ledger.BucketEarned)); earned.Available != 300_000 {
		t.Fatalf("earned left %d", earned.Available)
	}

	if status, decoded := buy("scout", `{"quantity":1}`); status != http.StatusConflict || decoded["error"].(map[string]any)["code"] != "INSUFFICIENT_FUNDS" {
		t.Fatalf("overspend: %d %v", status, decoded)
	}
	if status, _ := buy("warp-drive", `{"quantity":1}`); status != http.StatusNotFound {
		t.Fatalf("unknown sku: %d", status)
	}
	if status, _ := buy("fuel-cell", `{"quantity":101}`); status != http.StatusUnprocessableEntity {
		t.Fatalf("too many: %d", status)
	}

	pool.Exec(context.Background(), "UPDATE users SET status = 'FROZEN' WHERE id = $1", player)
	if status, _ := buy("fuel-cell", `{"quantity":1}`); status != http.StatusForbidden {
		t.Fatalf("frozen players can't buy: %d", status)
	}
	ledgertest.RequireIntegrity(t, pool)
}
