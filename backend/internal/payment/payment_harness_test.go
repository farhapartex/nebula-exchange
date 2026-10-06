package payment_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v87/webhook"
	"gorm.io/gorm"

	identitymodels "github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment"
	paymentgateway "github.com/farhapartex/nebula-exchange/backend/internal/payment/gateway"
	paymentseeding "github.com/farhapartex/nebula-exchange/backend/internal/payment/seeding"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress"
	"github.com/farhapartex/nebula-exchange/backend/internal/story"
	storymodels "github.com/farhapartex/nebula-exchange/backend/internal/story/models"
)

const (
	playerToken        = "player-token"
	otherPlayerToken   = "other-player-token"
	testWebhookSecret  = "whsec_test_street_born"
	testFrontendURL    = "http://localhost:3000"
	stripeSignatureKey = "Stripe-Signature"
)

type fixedPlayers struct {
	userIDByToken map[string]uuid.UUID
}

func (players fixedPlayers) Verify(accessToken string) (uuid.UUID, error) {
	userID, isKnown := players.userIDByToken[accessToken]
	if !isKnown {
		return uuid.Nil, errors.New("invalid token")
	}
	return userID, nil
}

type fakeCheckoutGateway struct {
	mutex           sync.Mutex
	createdRequests []service.CheckoutRequest
	expiredSessions []string
	sessions        map[string]service.CheckoutSnapshot
	createErr       error
}

func newFakeCheckoutGateway() *fakeCheckoutGateway {
	return &fakeCheckoutGateway{sessions: map[string]service.CheckoutSnapshot{}}
}

func (gateway *fakeCheckoutGateway) CreateCheckout(_ context.Context, checkoutRequest service.CheckoutRequest) (service.CreatedCheckout, error) {
	gateway.mutex.Lock()
	defer gateway.mutex.Unlock()
	if gateway.createErr != nil {
		return service.CreatedCheckout{}, gateway.createErr
	}
	gateway.createdRequests = append(gateway.createdRequests, checkoutRequest)
	sessionID := fmt.Sprintf("cs_test_%d", len(gateway.createdRequests))
	gateway.sessions[sessionID] = service.CheckoutSnapshot{SessionID: sessionID, AmountTotalCents: checkoutRequest.AmountCents, Currency: checkoutRequest.Currency}
	return service.CreatedCheckout{SessionID: sessionID, URL: "https://checkout.stripe.test/" + sessionID}, nil
}

func (gateway *fakeCheckoutGateway) ExpireCheckout(_ context.Context, sessionID string) error {
	gateway.mutex.Lock()
	defer gateway.mutex.Unlock()
	snapshot, isKnown := gateway.sessions[sessionID]
	if !isKnown || snapshot.IsPaid {
		return errors.New("checkout session cannot be expired")
	}
	snapshot.IsExpired = true
	gateway.sessions[sessionID] = snapshot
	gateway.expiredSessions = append(gateway.expiredSessions, sessionID)
	return nil
}

func (gateway *fakeCheckoutGateway) FetchCheckout(_ context.Context, sessionID string) (service.CheckoutSnapshot, error) {
	gateway.mutex.Lock()
	defer gateway.mutex.Unlock()
	snapshot, isKnown := gateway.sessions[sessionID]
	if !isKnown {
		return service.CheckoutSnapshot{}, errors.New("no such checkout session")
	}
	return snapshot, nil
}

func (gateway *fakeCheckoutGateway) payInStripe(sessionID string, paymentIntentID string) {
	gateway.mutex.Lock()
	defer gateway.mutex.Unlock()
	snapshot := gateway.sessions[sessionID]
	snapshot.IsPaid = true
	snapshot.PaymentIntentID = paymentIntentID
	gateway.sessions[sessionID] = snapshot
}

type harnessOptions struct {
	withoutStripe bool
}

type paymentHarness struct {
	t             *testing.T
	database      *gorm.DB
	router        *gin.Engine
	playerID      uuid.UUID
	otherPlayerID uuid.UUID
	gateway       *fakeCheckoutGateway
}

func plansSeedFile() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "..", "..", "seeds", "plans", "plans.json")
}

func newPaymentHarness(t *testing.T, options ...harnessOptions) *paymentHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	testDatabase := databasetest.Open(t)
	quietLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	playerID, otherPlayerID := uuid.New(), uuid.New()
	activatedAt := time.Now()

	records := []any{
		&identitymodels.User{ID: playerID, Email: "boy@streetborn.test", Username: "the_boy", PasswordHash: "x", Status: identitymodels.UserStatusActive, IsActive: true, ActivatedAt: &activatedAt, TermsAcceptedAt: activatedAt},
		&identitymodels.User{ID: otherPlayerID, Email: "girl@streetborn.test", Username: "the_girl", PasswordHash: "x", Status: identitymodels.UserStatusActive, IsActive: true, ActivatedAt: &activatedAt, TermsAcceptedAt: activatedAt},
		&storymodels.Arena{ID: "burning-house", Name: "The burning house", Width: 960, FloorY: 470, Stage: database.JSONDocument(`{}`)},
		&storymodels.Chapter{ID: "0", Number: 1, Title: "Prologue", Summary: "x", IsFree: true, IsPublished: true},
	}
	for chapterNumber := 2; chapterNumber <= 5; chapterNumber++ {
		chapterID := string(rune('0' + chapterNumber))
		levelID := chapterID + "-1"
		chapterPrice := int64(499)
		records = append(records,
			&storymodels.Chapter{ID: chapterID, Number: chapterNumber, Title: "Chapter " + chapterID, Summary: "x", IsFree: false, PriceCents: &chapterPrice, IsPublished: true},
			&storymodels.Level{ID: levelID, ChapterID: &chapterID, Number: func() *int { number := 1; return &number }(), Kind: storymodels.LevelKindStory, Title: levelID, Teaser: "x", ArenaID: "burning-house", TimeLimitSeconds: 90, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(`[]`), IsPublished: true},
		)
	}
	prologueChapterID := "0"
	records = append(records, &storymodels.Level{ID: "0-1", ChapterID: &prologueChapterID, Number: func() *int { number := 1; return &number }(), Kind: storymodels.LevelKindStory, Title: "0-1", Teaser: "x", ArenaID: "burning-house", TimeLimitSeconds: 90, Difficulty: database.JSONDocument(`{}`), StarRules: database.JSONDocument(`[]`), IsPublished: true, IsFree: true})
	for _, record := range records {
		if err := testDatabase.Create(record).Error; err != nil {
			t.Fatalf("seed %T: %v", record, err)
		}
	}
	plans, err := paymentseeding.LoadPlans(plansSeedFile())
	if err != nil {
		t.Fatalf("load plans: %v", err)
	}
	if err := paymentseeding.SeedPlans(context.Background(), testDatabase, plans, quietLogger); err != nil {
		t.Fatalf("seed plans: %v", err)
	}

	storyModule := story.NewModule(story.ModuleDependencies{Database: testDatabase})
	progressModule := progress.NewModule(progress.ModuleDependencies{Database: testDatabase, Levels: storyModule.LevelCatalog, Logger: quietLogger, Now: time.Now})
	fakeGateway := newFakeCheckoutGateway()
	paymentDependencies := payment.ModuleDependencies{
		Database:         testDatabase,
		Chapters:         storyModule.LevelCatalog,
		ChapterOwnership: progressModule.ChapterOwnership,
		ChapterUnlocker:  progressModule.ChapterUnlocks,
		FrontendBaseURL:  testFrontendURL,
		Logger:           quietLogger,
		Now:              time.Now,
	}
	if len(options) == 0 || !options[0].withoutStripe {
		paymentDependencies.CheckoutGateway = fakeGateway
		paymentDependencies.EventVerifier = paymentgateway.NewStripeEventVerifier(testWebhookSecret)
	}
	paymentModule := payment.NewModule(paymentDependencies)
	router, err := httpserver.NewRouter(httpserver.RouterOptions{
		Logger:       quietLogger,
		AccessTokens: fixedPlayers{userIDByToken: map[string]uuid.UUID{playerToken: playerID, otherPlayerToken: otherPlayerID}},
	}, paymentModule.RouteRegistrars()...)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return &paymentHarness{t: t, database: testDatabase, router: router, playerID: playerID, otherPlayerID: otherPlayerID, gateway: fakeGateway}
}

type listBody struct {
	Data       []map[string]any `json:"data"`
	Pagination struct {
		NextCursor *string `json:"next_cursor"`
	} `json:"pagination"`
}

type dataBody struct {
	Data  map[string]any `json:"data"`
	Error struct {
		Code    string            `json:"code"`
		Details map[string]string `json:"details"`
	} `json:"error"`
}

func (harness *paymentHarness) serve(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	return recorder
}

func (harness *paymentHarness) get(path string, accessToken string) (int, listBody) {
	harness.t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/api/v1"+path, nil)
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	recorder := harness.serve(request)
	var body listBody
	_ = json.Unmarshal(recorder.Body.Bytes(), &body)
	return recorder.Code, body
}

func (harness *paymentHarness) getData(path string, accessToken string) (int, dataBody) {
	harness.t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/api/v1"+path, nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	recorder := harness.serve(request)
	var body dataBody
	_ = json.Unmarshal(recorder.Body.Bytes(), &body)
	return recorder.Code, body
}

func (harness *paymentHarness) startCheckout(accessToken string, planID string, chapterCount int) (int, dataBody) {
	harness.t.Helper()
	requestBody, _ := json.Marshal(map[string]any{"plan_id": planID, "chapter_count": chapterCount})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/checkout-sessions", bytes.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Nebula-Client", "web")
	request.Header.Set("Authorization", "Bearer "+accessToken)
	recorder := harness.serve(request)
	var body dataBody
	_ = json.Unmarshal(recorder.Body.Bytes(), &body)
	return recorder.Code, body
}

func (harness *paymentHarness) sendStripeEvent(event map[string]any, secret string) int {
	harness.t.Helper()
	payload, _ := json.Marshal(event)
	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payload, Secret: secret, Timestamp: time.Now()})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(stripeSignatureKey, signedPayload.Header)
	return harness.serve(request).Code
}

func stripeEvent(eventID string, eventType string, object map[string]any) map[string]any {
	return map[string]any{"id": eventID, "object": "event", "type": eventType, "data": map[string]any{"object": object}}
}

func completedCheckoutObject(sessionID string, paymentID string, paymentIntentID string, amountCents int64) map[string]any {
	return map[string]any{
		"id":                  sessionID,
		"object":              "checkout.session",
		"client_reference_id": paymentID,
		"status":              "complete",
		"payment_status":      "paid",
		"payment_intent":      paymentIntentID,
		"amount_total":        amountCents,
		"currency":            "usd",
	}
}

func (harness *paymentHarness) ownedChapterNumbers(accessToken string) []float64 {
	harness.t.Helper()
	_, chapters := harness.get("/chapters", accessToken)
	var ownedNumbers []float64
	for _, chapter := range chapters.Data {
		if chapter["is_owned"] == true {
			ownedNumbers = append(ownedNumbers, chapter["number"].(float64))
		}
	}
	return ownedNumbers
}

func (harness *paymentHarness) paymentStatus(paymentID string) string {
	harness.t.Helper()
	var status string
	if err := harness.database.Table("payments").Where("id = ?", paymentID).Pluck("status", &status).Error; err != nil {
		harness.t.Fatalf("read payment status: %v", err)
	}
	return status
}
