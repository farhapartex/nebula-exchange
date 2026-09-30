package inapp_test

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

	"github.com/google/uuid"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/notify/inapp"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/users"
)

func TestNotificationsListCountAndMarkRead(t *testing.T) {
	pool := databasetest.NewPool(t)
	accessTokens, _ := accesstoken.NewManager("notifications-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		inapp.NewHandler(pool, time.Now),
	)
	notifier := inapp.NewNotifier(inapp.Dependencies{Users: users.NewRepository()})
	player := ledgertest.CreatePlayer(t, pool)
	otherPlayer := ledgertest.CreatePlayer(t, pool)
	for noticeIndex := 0; noticeIndex < 3; noticeIndex++ {
		notifier.Notify(context.Background(), pool, inapp.Notice{UserID: player, Kind: inapp.KindMissionCompleted, Title: "Mission back", Body: "collect it", Link: "/missions"})
	}
	notifier.Notify(context.Background(), pool, inapp.Notice{UserID: otherPlayer, Kind: inapp.KindCraftCompleted, Title: "Craft", Body: "done"})

	send := func(playerID uuid.UUID, method, path, body string) (int, map[string]any) {
		accessToken, _ := accessTokens.Issue(playerID)
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
	unreadCount := func(playerID uuid.UUID) float64 {
		_, decoded := send(playerID, http.MethodGet, "/me/notifications/unread-count", "")
		return decoded["data"].(map[string]any)["unread_count"].(float64)
	}

	_, listed := send(player, http.MethodGet, "/me/notifications?limit=2", "")
	notifications := listed["data"].([]any)
	if len(notifications) != 2 || listed["pagination"].(map[string]any)["next_cursor"] == nil || unreadCount(player) != 3 {
		t.Fatalf("list %v", listed)
	}
	firstID := notifications[0].(map[string]any)["id"].(string)
	if _, marked := send(player, http.MethodPost, "/me/notifications/read", `{"ids":["`+firstID+`"]}`); marked["data"].(map[string]any)["marked"].(float64) != 1 {
		t.Fatalf("mark one %v", marked)
	}
	if unreadCount(player) != 2 {
		t.Fatal("one read, two left")
	}
	_, unreadOnly := send(player, http.MethodGet, "/me/notifications?unread_only=true", "")
	if len(unreadOnly["data"].([]any)) != 2 {
		t.Fatalf("unread only %v", unreadOnly)
	}
	send(player, http.MethodPost, "/me/notifications/read", `{"all":true}`)
	if unreadCount(player) != 0 || unreadCount(otherPlayer) != 1 {
		t.Fatal("mark all must only touch the player's own notifications")
	}
	if status, _ := send(player, http.MethodPost, "/me/notifications/read", `{}`); status != http.StatusUnprocessableEntity {
		t.Fatalf("empty mark request %d", status)
	}
}
