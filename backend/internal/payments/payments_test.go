package payments_test

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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/notify/inapp"
	"nebula-exchange/backend/internal/payments"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/platform/money"
	"nebula-exchange/backend/internal/purpose"
	"nebula-exchange/backend/internal/users"
)

type paymentsHarness struct {
	pool         *pgxpool.Pool
	router       http.Handler
	accessTokens *accesstoken.Manager
	settler      *payments.Settler
}

func allowEverything(context *gin.Context) { context.Next() }

func newTestNotifier(t *testing.T, userRepository *users.Repository) *inapp.Notifier {
	t.Helper()
	emailTemplates, err := email.NewTemplateRenderer()
	if err != nil {
		t.Fatalf("templates: %v", err)
	}
	return inapp.NewNotifier(inapp.Dependencies{
		Users:           userRepository,
		EmailTemplates:  emailTemplates,
		EmailQueue:      outbox.NewQueue(),
		FrontendBaseURL: "http://localhost:3000",
	})
}

func newPaymentsHarness(t *testing.T) *paymentsHarness {
	t.Helper()
	pool := databasetest.NewPool(t)
	accessTokens, _ := accesstoken.NewManager("payments-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	userRepository := users.NewRepository()
	catalogService := catalog.NewService(catalog.NewLoader(pool), time.Minute, time.Now)
	service := payments.NewService(payments.Dependencies{
		Pool:     pool,
		Users:    userRepository,
		Catalog:  catalogService,
		Checkout: payments.NewDevelopmentCheckout("http://localhost:3000"),
		Now:      time.Now,
	})
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		payments.NewHandler(service, users.NewAccountGuard(pool, userRepository).RequireStatus(users.StatusActive, users.StatusPendingPayment), allowEverything),
	)
	return &paymentsHarness{
		pool:         pool,
		router:       router,
		accessTokens: accessTokens,
		settler:      payments.NewSettler(pool, purpose.NewDefaultRunner(userRepository, catalogService, time.Now), newTestNotifier(t, userRepository), testLogger, time.Now),
	}
}

func (harness *paymentsHarness) createPlayer(t *testing.T, status users.Status) uuid.UUID {
	t.Helper()
	playerID := ledgertest.CreatePlayer(t, harness.pool)
	if _, err := harness.pool.Exec(context.Background(), "UPDATE users SET status = $2 WHERE id = $1", playerID, string(status)); err != nil {
		t.Fatalf("set status: %v", err)
	}
	return playerID
}

func (harness *paymentsHarness) send(t *testing.T, playerID uuid.UUID, method, path, body string) (int, map[string]any) {
	t.Helper()
	accessToken, _ := harness.accessTokens.Issue(playerID)
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

func (harness *paymentsHarness) createPayment(t *testing.T, playerID uuid.UUID, body string) payments.Payment {
	t.Helper()
	status, decoded := harness.send(t, playerID, http.MethodPost, "/payments", body)
	if status != http.StatusCreated {
		t.Fatalf("create payment status %d: %v", status, decoded)
	}
	encoded, _ := json.Marshal(decoded["data"])
	var payment payments.Payment
	json.Unmarshal(encoded, &payment)
	return payment
}

func (harness *paymentsHarness) settle(t *testing.T, payment payments.Payment, eventID string) payments.SettlementOutcome {
	t.Helper()
	outcome, err := harness.settler.Settle(context.Background(), payments.SettlementRequest{
		PaymentID: payment.ID,
		Credited:  payment.Amount,
		Event:     payments.ExternalEvent{ID: eventID, Provider: "development", Type: "checkout.completed"},
	})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	return outcome
}

func accountStatus(t *testing.T, pool *pgxpool.Pool, playerID uuid.UUID) string {
	t.Helper()
	var status string
	pool.QueryRow(context.Background(), "SELECT status FROM users WHERE id = $1", playerID).Scan(&status)
	return status
}

func errorCode(decoded map[string]any) string {
	errorBody, _ := decoded["error"].(map[string]any)
	code, _ := errorBody["code"].(string)
	return code
}

func TestEntryFeeActivatesTheAccountAndGrantsTheStarterPackOnce(t *testing.T) {
	harness := newPaymentsHarness(t)
	player := harness.createPlayer(t, users.StatusPendingPayment)

	entryFee := harness.createPayment(t, player, `{"purpose":"ENTRY_FEE","method":"card"}`)
	if entryFee.Status != payments.StatusPending || entryFee.Amount != 5*money.MicroPerNC || entryFee.CheckoutURL == nil ||
		!strings.Contains(*entryFee.CheckoutURL, "/payment/result?id="+entryFee.ID.String()) {
		t.Fatalf("unexpected payment %+v", entryFee)
	}

	outcome := harness.settle(t, entryFee, "evt_entry_1")
	if !outcome.WasSettled || !outcome.PurposeApplied {
		t.Fatalf("outcome %+v", outcome)
	}
	if duplicate := harness.settle(t, entryFee, "evt_entry_1"); !duplicate.WasDuplicateEvent || duplicate.WasSettled {
		t.Fatalf("a replayed event must be ignored: %+v", duplicate)
	}
	if again := harness.settle(t, entryFee, "evt_entry_2"); again.WasSettled {
		t.Fatalf("a settled payment must not settle twice: %+v", again)
	}

	if status := accountStatus(t, harness.pool, player); status != "ACTIVE" {
		t.Fatalf("account status %s", status)
	}
	if card := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerNC(player, ledger.BucketCard)); card.Available != 2*int64(money.MicroPerNC) {
		t.Fatalf("card should hold only the 2 NC starter bonus, got %d", card.Available)
	}
	for itemID, expected := range map[int]int64{301: 1, 201: 1, 401: 10} {
		if holding := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, itemID)); holding.Available != expected {
			t.Fatalf("item %d: %d, want %d", itemID, holding.Available, expected)
		}
	}

	var paymentNotices, paymentEmails int
	harness.pool.QueryRow(context.Background(), "SELECT count(*) FROM notifications WHERE user_id = $1 AND kind = 'payment_succeeded'", player).Scan(&paymentNotices)
	harness.pool.QueryRow(context.Background(), "SELECT count(*) FROM email_outbox WHERE template = 'game_notice' AND subject LIKE 'Welcome aboard%'").Scan(&paymentEmails)
	if paymentNotices != 1 || paymentEmails != 1 {
		t.Fatalf("payment notices %d, emails %d, want 1 each", paymentNotices, paymentEmails)
	}

	_, fetched := harness.send(t, player, http.MethodGet, "/payments/"+entryFee.ID.String(), "")
	fetchedPayment := fetched["data"].(map[string]any)
	if fetchedPayment["status"] != "SUCCEEDED" || fetchedPayment["purpose_status"] != "APPLIED" || fetchedPayment["credited"] != "5000000" {
		t.Fatalf("fetched payment %v", fetchedPayment)
	}
	ledgertest.RequireIntegrity(t, harness.pool)
}

func TestFailedPurposeKeepsTheCreditedNC(t *testing.T) {
	harness := newPaymentsHarness(t)
	player := harness.createPlayer(t, users.StatusPendingPayment)
	entryFee := harness.createPayment(t, player, `{"purpose":"ENTRY_FEE","method":"card"}`)
	harness.pool.Exec(context.Background(), "UPDATE users SET status = 'FROZEN' WHERE id = $1", player)

	outcome := harness.settle(t, entryFee, "evt_frozen")
	if !outcome.WasSettled || outcome.PurposeApplied || outcome.PurposeFailure != "ACCOUNT_NOT_PENDING_PAYMENT" {
		t.Fatalf("outcome %+v", outcome)
	}
	if card := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerNC(player, ledger.BucketCard)); card.Available != 5*int64(money.MicroPerNC) {
		t.Fatalf("the paid NC must stay in the balance, got %d", card.Available)
	}
	if scouts := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, 301)); scouts.Available != 0 {
		t.Fatal("no starter pack for a failed activation")
	}
	ledgertest.RequireIntegrity(t, harness.pool)
}

func TestPaymentRulesPerPurposeAndAccountState(t *testing.T) {
	harness := newPaymentsHarness(t)
	pendingPlayer := harness.createPlayer(t, users.StatusPendingPayment)
	activePlayer := harness.createPlayer(t, users.StatusActive)

	cases := []struct {
		name           string
		player         uuid.UUID
		body           string
		expectedStatus int
		expectedCode   string
	}{
		{"entry fee twice", activePlayer, `{"purpose":"ENTRY_FEE","method":"card"}`, http.StatusConflict, "CONFLICT"},
		{"top-up before activation", pendingPlayer, `{"purpose":"TOPUP","method":"card","amount_nc":"10"}`, http.StatusForbidden, "ACCOUNT_NOT_ACTIVE"},
		{"odd top-up amount", activePlayer, `{"purpose":"TOPUP","method":"card","amount_nc":"7"}`, http.StatusUnprocessableEntity, "VALIDATION_FAILED"},
		{"top-up without amount", activePlayer, `{"purpose":"TOPUP","method":"card"}`, http.StatusUnprocessableEntity, "VALIDATION_FAILED"},
		{"unknown sku", activePlayer, `{"purpose":"SHOP_PURCHASE","method":"card","sku":"warp-drive"}`, http.StatusUnprocessableEntity, "VALIDATION_FAILED"},
		{"crypto without wallet", activePlayer, `{"purpose":"TOPUP","method":"crypto","amount_nc":"10"}`, http.StatusUnprocessableEntity, "WALLET_REQUIRED"},
		{"unknown purpose", activePlayer, `{"purpose":"GIFT","method":"card"}`, http.StatusUnprocessableEntity, "VALIDATION_FAILED"},
	}
	for _, testCase := range cases {
		status, decoded := harness.send(t, testCase.player, http.MethodPost, "/payments", testCase.body)
		if status != testCase.expectedStatus || errorCode(decoded) != testCase.expectedCode {
			t.Fatalf("%s: got %d %s, want %d %s", testCase.name, status, errorCode(decoded), testCase.expectedStatus, testCase.expectedCode)
		}
	}

	for topupIndex := 0; topupIndex < 5; topupIndex++ {
		harness.createPayment(t, activePlayer, `{"purpose":"TOPUP","method":"card","amount_nc":"100"}`)
	}
	status, decoded := harness.send(t, activePlayer, http.MethodPost, "/payments", `{"purpose":"TOPUP","method":"card","amount_nc":"5"}`)
	if status != http.StatusUnprocessableEntity || errorCode(decoded) != "LIMIT_EXCEEDED" {
		t.Fatalf("daily limit: %d %v", status, decoded)
	}

	otherPlayer := harness.createPlayer(t, users.StatusActive)
	listStatus, listed := harness.send(t, activePlayer, http.MethodGet, "/payments?limit=3", "")
	if listStatus != http.StatusOK || len(listed["data"].([]any)) != 3 || listed["pagination"].(map[string]any)["next_cursor"] == nil {
		t.Fatalf("list payments %v", listed)
	}
	someonesPayment := listed["data"].([]any)[0].(map[string]any)["id"].(string)
	if status, _ := harness.send(t, otherPlayer, http.MethodGet, "/payments/"+someonesPayment, ""); status != http.StatusNotFound {
		t.Fatalf("another player's payment must look missing, got %d", status)
	}
}

func TestShopPurchaseByCardGrantsTheBundle(t *testing.T) {
	harness := newPaymentsHarness(t)
	player := harness.createPlayer(t, users.StatusActive)
	bundle := harness.createPayment(t, player, `{"purpose":"SHOP_PURCHASE","method":"card","sku":"starter-bundle"}`)
	if bundle.Amount != 3_500_000 {
		t.Fatalf("bundle price %d", bundle.Amount)
	}
	if outcome := harness.settle(t, bundle, "evt_bundle"); !outcome.PurposeApplied {
		t.Fatalf("outcome %+v", outcome)
	}
	if fuel := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, 401)); fuel.Available != 20 {
		t.Fatalf("fuel %d", fuel.Available)
	}
	if card := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerNC(player, ledger.BucketCard)); card.Available != 0 {
		t.Fatalf("card NC should be spent on the bundle, left %d", card.Available)
	}
	ledgertest.RequireIntegrity(t, harness.pool)
}

func TestCardAndBalancePurchasesProduceTheSameItems(t *testing.T) {
	harness := newPaymentsHarness(t)
	player := harness.createPlayer(t, users.StatusActive)
	fuelOrder := harness.createPayment(t, player, `{"purpose":"SHOP_PURCHASE","method":"card","sku":"fuel-cell","quantity":3}`)
	if fuelOrder.Amount != 600_000 || fuelOrder.Quantity != 3 {
		t.Fatalf("card order %+v", fuelOrder)
	}
	if outcome := harness.settle(t, fuelOrder, "evt_fuel"); !outcome.PurposeApplied {
		t.Fatalf("outcome %+v", outcome)
	}
	if fuel := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, 401)); fuel.Available != 3 {
		t.Fatalf("fuel %d", fuel.Available)
	}
	if card := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerNC(player, ledger.BucketCard)); card.Available != 0 {
		t.Fatalf("card left %d", card.Available)
	}
	ledgertest.RequireIntegrity(t, harness.pool)
}

func TestUpgradePurchaseByCardSwapsTheTier(t *testing.T) {
	harness := newPaymentsHarness(t)
	player := harness.createPlayer(t, users.StatusActive)
	ledgertest.Fund(t, harness.pool, ledger.PlayerItem(player, 301), 1)

	upgradeOrder := harness.createPayment(t, player, `{"purpose":"UPGRADE_PURCHASE","method":"card","upgrade_id":"scout-to-hauler"}`)
	if upgradeOrder.Amount != 6*money.MicroPerNC || upgradeOrder.UpgradeID == nil {
		t.Fatalf("upgrade order %+v", upgradeOrder)
	}
	if outcome := harness.settle(t, upgradeOrder, "evt_upgrade"); !outcome.PurposeApplied {
		t.Fatalf("outcome %+v", outcome)
	}
	if hauler := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerItem(player, 302)); hauler.Available != 1 {
		t.Fatalf("hauler %d", hauler.Available)
	}

	shiplessPlayer := harness.createPlayer(t, users.StatusActive)
	shiplessOrder := harness.createPayment(t, shiplessPlayer, `{"purpose":"UPGRADE_PURCHASE","method":"card","upgrade_id":"scout-to-hauler"}`)
	if outcome := harness.settle(t, shiplessOrder, "evt_shipless"); outcome.PurposeApplied || outcome.PurposeFailure != "INSUFFICIENT_ITEMS" {
		t.Fatalf("no scout to upgrade: %+v", outcome)
	}
	if card := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerNC(shiplessPlayer, ledger.BucketCard)); card.Available != 6*int64(money.MicroPerNC) {
		t.Fatalf("NC must stay after a failed upgrade, got %d", card.Available)
	}
	if status, _ := harness.send(t, player, http.MethodPost, "/payments", `{"purpose":"UPGRADE_PURCHASE","method":"card","upgrade_id":"drill-t4-to-t5"}`); status != http.StatusUnprocessableEntity {
		t.Fatalf("T5 can't be bought: %d", status)
	}
	ledgertest.RequireIntegrity(t, harness.pool)
}
