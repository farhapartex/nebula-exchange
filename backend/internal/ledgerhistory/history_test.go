package ledgerhistory_test

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
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/ledgerhistory"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
)

type historyPage struct {
	Data       []ledgerhistory.JournalRecord `json:"data"`
	Pagination struct {
		NextCursor *string `json:"next_cursor"`
	} `json:"pagination"`
}

func TestLedgerHistoryShowsOnlyThePlayersOwnEntriesNewestFirst(t *testing.T) {
	pool := databasetest.NewPool(t)
	accessTokens, _ := accesstoken.NewManager("history-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		ledgerhistory.NewHandler(ledgerhistory.NewReader(pool)),
	)

	player := ledgertest.CreatePlayer(t, pool)
	otherPlayer := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCard), 5_000_000)
	ledgertest.Fund(t, pool, ledger.PlayerItem(player, 401), 10)
	ledgertest.Fund(t, pool, ledger.PlayerNC(otherPlayer, ledger.BucketCard), 9_000_000)

	accessToken, _ := accessTokens.Issue(player)
	fetch := func(path string, expectedStatus int) historyPage {
		historyRequest := httptest.NewRequest(http.MethodGet, "/api/v1"+path, nil)
		historyRequest.Header.Set("Authorization", "Bearer "+accessToken.Value)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, historyRequest)
		if recorder.Code != expectedStatus {
			t.Fatalf("%s status %d: %s", path, recorder.Code, recorder.Body.String())
		}
		var page historyPage
		json.Unmarshal(recorder.Body.Bytes(), &page)
		return page
	}

	firstPage := fetch("/me/ledger?limit=1", http.StatusOK)
	if len(firstPage.Data) != 1 || firstPage.Pagination.NextCursor == nil {
		t.Fatalf("first page %+v", firstPage)
	}
	newestJournal := firstPage.Data[0]
	if len(newestJournal.Entries) != 1 || newestJournal.Entries[0].ItemID == nil || *newestJournal.Entries[0].ItemID != 401 || newestJournal.Entries[0].Amount != "10" {
		t.Fatalf("newest journal should be the fuel grant with only the player's entry: %+v", newestJournal)
	}

	secondPage := fetch("/me/ledger?limit=1&cursor="+*firstPage.Pagination.NextCursor, http.StatusOK)
	if len(secondPage.Data) != 1 || secondPage.Pagination.NextCursor != nil {
		t.Fatalf("second page %+v", secondPage)
	}
	cardEntry := secondPage.Data[0].Entries[0]
	if cardEntry.Bucket == nil || *cardEntry.Bucket != "card" || cardEntry.Amount != "5000000" || cardEntry.ItemID != nil {
		t.Fatalf("card entry %+v", cardEntry)
	}

	if filtered := fetch("/me/ledger?type=TRADE_FILL", http.StatusOK); len(filtered.Data) != 0 {
		t.Fatalf("no trades yet, got %+v", filtered)
	}
	fetch("/me/ledger?type=NOT_A_TYPE", http.StatusUnprocessableEntity)
}
