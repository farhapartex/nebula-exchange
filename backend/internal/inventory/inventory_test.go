package inventory_test

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
	"nebula-exchange/backend/internal/inventory"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
)

type inventoryPage struct {
	Data       []inventory.Holding `json:"data"`
	Pagination struct {
		NextCursor *string `json:"next_cursor"`
	} `json:"pagination"`
}

func TestInventoryListsOwnedItemsWithHeldQuantities(t *testing.T) {
	pool := databasetest.NewPool(t)
	accessTokens, _ := accesstoken.NewManager("inventory-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		inventory.NewHandler(inventory.NewReader(pool)),
	)

	player := ledgertest.CreatePlayer(t, pool)
	otherPlayer := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerItem(player, 401), 10)
	ledgertest.Fund(t, pool, ledger.PlayerItem(player, 1), 25)
	ledgertest.Fund(t, pool, ledger.PlayerItem(player, 301), 1)
	ledgertest.Fund(t, pool, ledger.PlayerItem(otherPlayer, 6), 3)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCard), 5_000_000)
	ledgertest.Hold(t, pool, ledger.PlayerItem(player, 401), 4)

	accessToken, _ := accessTokens.Issue(player)
	fetch := func(path string) inventoryPage {
		inventoryRequest := httptest.NewRequest(http.MethodGet, "/api/v1"+path, nil)
		inventoryRequest.Header.Set("Authorization", "Bearer "+accessToken.Value)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, inventoryRequest)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status %d: %s", path, recorder.Code, recorder.Body.String())
		}
		var page inventoryPage
		json.Unmarshal(recorder.Body.Bytes(), &page)
		return page
	}

	firstPage := fetch("/me/inventory?limit=2")
	if len(firstPage.Data) != 2 || firstPage.Data[0].ItemID != 1 || firstPage.Data[1].ItemID != 301 || firstPage.Pagination.NextCursor == nil {
		t.Fatalf("first page %+v", firstPage)
	}
	secondPage := fetch("/me/inventory?limit=2&cursor=" + *firstPage.Pagination.NextCursor)
	if len(secondPage.Data) != 1 || secondPage.Data[0] != (inventory.Holding{ItemID: 401, Available: 6, Held: 4}) || secondPage.Pagination.NextCursor != nil {
		t.Fatalf("second page %+v", secondPage)
	}
}
