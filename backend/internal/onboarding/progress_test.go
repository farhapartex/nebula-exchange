package onboarding_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/onboarding"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
)

func TestOnboardingProgressFollowsTheLedger(t *testing.T) {
	pool := databasetest.NewPool(t)
	accessTokens, _ := accesstoken.NewManager("onboarding-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		onboarding.NewHandler(pool),
	)
	player := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCard), 5_000_000)
	if err := ledgertest.InTransaction(t, pool, func(tx pgx.Tx) error {
		_, err := ledger.Post(context.Background(), tx, ledger.Journal{
			Type:      ledger.JournalEntryFee,
			Reference: ledger.Reference{Type: "payment", ID: "entry"},
			Legs: []ledger.Leg{
				ledger.Debit(ledger.PlayerNC(player, ledger.BucketCard), 5_000_000),
				ledger.Credit(ledger.System(ledger.SystemTreasury, ledger.NC), 5_000_000),
			},
		})
		return err
	}); err != nil {
		t.Fatalf("post entry fee: %v", err)
	}

	accessToken, _ := accessTokens.Issue(player)
	progressRequest := httptest.NewRequest(http.MethodGet, "/api/v1/me/onboarding", nil)
	progressRequest.Header.Set("Authorization", "Bearer "+accessToken.Value)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, progressRequest)

	var envelope struct {
		Data onboarding.Progress `json:"data"`
	}
	json.Unmarshal(recorder.Body.Bytes(), &envelope)
	if recorder.Code != http.StatusOK || envelope.Data.CompletedCount != 1 || !envelope.Data.Steps[0].IsCompleted || envelope.Data.Steps[1].IsCompleted {
		t.Fatalf("progress %d %+v", recorder.Code, envelope.Data)
	}
}
