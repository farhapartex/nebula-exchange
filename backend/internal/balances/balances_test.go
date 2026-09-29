package balances_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/balances"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
)

func TestBalancesSummarizeEveryBucketInSpendingOrder(t *testing.T) {
	pool := databasetest.NewPool(t)
	accessTokens, _ := accesstoken.NewManager("balances-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		balances.NewHandler(balances.NewReader(pool)),
	)

	player := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCard), 2_000_000)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketEarned), 1_500_000)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCrypto), 750_000)
	ledgertest.Hold(t, pool, ledger.PlayerNC(player, ledger.BucketEarned), 500_000)

	unauthenticated := httptest.NewRecorder()
	router.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodGet, "/api/v1/me/balances", nil))
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous status %d", unauthenticated.Code)
	}

	accessToken, _ := accessTokens.Issue(player)
	balanceRequest := httptest.NewRequest(http.MethodGet, "/api/v1/me/balances", nil)
	balanceRequest.Header.Set("Authorization", "Bearer "+accessToken.Value)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, balanceRequest)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body.String())
	}

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	json.Unmarshal(recorder.Body.Bytes(), &envelope)
	expectedTotals := map[string]string{"total": "4250000", "available": "3750000", "held": "500000", "withdrawable": "1750000"}
	for field, expected := range expectedTotals {
		if envelope.Data[field] != expected {
			t.Fatalf("%s = %v, want %s (%s)", field, envelope.Data[field], expected, recorder.Body.String())
		}
	}
	buckets := envelope.Data["buckets"].([]any)
	if len(buckets) != 4 || buckets[0].(map[string]any)["bucket"] != "card" || buckets[3].(map[string]any)["bucket"] != "crypto" {
		t.Fatalf("buckets %v", buckets)
	}
}
