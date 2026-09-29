package activation_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/activation"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/users"
)

type activationTestHarness struct {
	pool        *pgxpool.Pool
	router      http.Handler
	issuer      *activation.Issuer
	repository  *users.Repository
	currentTime atomic.Pointer[time.Time]
}

func newActivationTestHarness(t *testing.T) *activationTestHarness {
	t.Helper()
	harness := &activationTestHarness{pool: databasetest.NewPool(t), repository: users.NewRepository()}
	startTime := time.Now().UTC()
	harness.currentTime.Store(&startTime)
	clock := func() time.Time { return *harness.currentTime.Load() }

	harness.issuer = activation.NewIssuer(activation.DefaultTokenLifetime, clock)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	harness.router = httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger},
		activation.NewHandler(activation.NewService(harness.pool, harness.repository, clock)),
	)
	return harness
}

func (harness *activationTestHarness) advanceClock(duration time.Duration) {
	advancedTime := harness.currentTime.Load().Add(duration)
	harness.currentTime.Store(&advancedTime)
}

func (harness *activationTestHarness) signedUpUserWithToken(t *testing.T) (uuid.UUID, string) {
	t.Helper()
	createdUser, err := harness.repository.Create(context.Background(), harness.pool, users.NewUser{
		Email:           "pilot@nebula.test",
		Username:        "pilot_nova",
		PasswordHash:    "unused-in-this-test",
		TermsAcceptedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	issuedToken, err := harness.issuer.Issue(context.Background(), harness.pool, createdUser.ID)
	if err != nil {
		t.Fatalf("issue activation token: %v", err)
	}
	return createdUser.ID, issuedToken.Plaintext
}

type activationEnvelope struct {
	Data struct {
		Email       string    `json:"email"`
		Username    string    `json:"username"`
		IsActivated bool      `json:"is_activated"`
		Status      string    `json:"status"`
		ActivatedAt time.Time `json:"activated_at"`
	} `json:"data"`
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func (harness *activationTestHarness) preview(t *testing.T, plaintextToken string) (int, activationEnvelope) {
	t.Helper()
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/auth/activations/"+plaintextToken, nil))
	var decodedBody activationEnvelope
	_ = json.Unmarshal(recorder.Body.Bytes(), &decodedBody)
	return recorder.Code, decodedBody
}

func (harness *activationTestHarness) activate(t *testing.T, plaintextToken string) (int, activationEnvelope) {
	t.Helper()
	encodedBody, _ := json.Marshal(map[string]string{"token": plaintextToken})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/activations", bytes.NewReader(encodedBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Nebula-Client", "web")
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	var decodedBody activationEnvelope
	_ = json.Unmarshal(recorder.Body.Bytes(), &decodedBody)
	return recorder.Code, decodedBody
}

type storedUserState struct {
	isActive    bool
	status      string
	activatedAt *time.Time
}

func (harness *activationTestHarness) userState(t *testing.T, userID uuid.UUID) storedUserState {
	t.Helper()
	var state storedUserState
	err := harness.pool.QueryRow(context.Background(), "SELECT is_active, status, activated_at FROM users WHERE id = $1", userID).
		Scan(&state.isActive, &state.status, &state.activatedAt)
	if err != nil {
		t.Fatalf("load user: %v", err)
	}
	return state
}

func TestPreviewShowsTheAccountBehindAValidLink(t *testing.T) {
	harness := newActivationTestHarness(t)
	_, plaintextToken := harness.signedUpUserWithToken(t)

	statusCode, preview := harness.preview(t, plaintextToken)

	if statusCode != http.StatusOK || preview.Data.Email != "pilot@nebula.test" || preview.Data.Username != "pilot_nova" || preview.Data.IsActivated {
		t.Fatalf("got %d %+v", statusCode, preview.Data)
	}
}

func TestActivationActivatesTheAccountOnce(t *testing.T) {
	harness := newActivationTestHarness(t)
	userID, plaintextToken := harness.signedUpUserWithToken(t)

	statusCode, activated := harness.activate(t, plaintextToken)

	if statusCode != http.StatusOK || activated.Data.Status != "PENDING_PAYMENT" || activated.Data.ActivatedAt.IsZero() {
		t.Fatalf("got %d %+v %s", statusCode, activated.Data, activated.Error.Code)
	}
	if state := harness.userState(t, userID); !state.isActive || state.status != "PENDING_PAYMENT" || state.activatedAt == nil {
		t.Fatalf("user was not activated: %+v", state)
	}

	repeatStatusCode, repeated := harness.activate(t, plaintextToken)
	if repeatStatusCode != http.StatusOK || !repeated.Data.ActivatedAt.Equal(activated.Data.ActivatedAt) {
		t.Fatalf("activating again should report the original activation, got %d %+v", repeatStatusCode, repeated.Data)
	}

	previewStatusCode, preview := harness.preview(t, plaintextToken)
	if previewStatusCode != http.StatusOK || !preview.Data.IsActivated {
		t.Fatalf("a used link for an active account must say it is activated, got %d %+v", previewStatusCode, preview.Data)
	}
}

func TestExpiredAndUnknownLinksAreRejected(t *testing.T) {
	harness := newActivationTestHarness(t)
	userID, plaintextToken := harness.signedUpUserWithToken(t)
	harness.advanceClock(activation.DefaultTokenLifetime + time.Second)

	previewStatusCode, previewBody := harness.preview(t, plaintextToken)
	activateStatusCode, activateBody := harness.activate(t, plaintextToken)
	unknownStatusCode, _ := harness.preview(t, "not-a-real-token")
	randomStatusCode, _ := harness.activate(t, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	if previewStatusCode != http.StatusNotFound || previewBody.Error.Code != "NOT_FOUND" {
		t.Fatalf("expired preview: got %d %s", previewStatusCode, previewBody.Error.Code)
	}
	if activateStatusCode != http.StatusNotFound || activateBody.Error.Code != "NOT_FOUND" {
		t.Fatalf("expired activation: got %d %s", activateStatusCode, activateBody.Error.Code)
	}
	if unknownStatusCode != http.StatusNotFound || randomStatusCode != http.StatusNotFound {
		t.Fatalf("unknown tokens: got %d and %d", unknownStatusCode, randomStatusCode)
	}
	if state := harness.userState(t, userID); state.isActive {
		t.Fatal("an expired link must not activate the account")
	}
}

func TestConcurrentActivationsActivateExactlyOnce(t *testing.T) {
	harness := newActivationTestHarness(t)
	userID, plaintextToken := harness.signedUpUserWithToken(t)

	var waitGroup sync.WaitGroup
	activatedAtValues := make(chan time.Time, 8)
	for attemptIndex := 0; attemptIndex < 8; attemptIndex++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			statusCode, activated := harness.activate(t, plaintextToken)
			if statusCode != http.StatusOK {
				t.Errorf("concurrent activation got %d", statusCode)
				return
			}
			activatedAtValues <- activated.Data.ActivatedAt
		}()
	}
	waitGroup.Wait()
	close(activatedAtValues)

	storedActivatedAt := harness.userState(t, userID).activatedAt
	for activatedAt := range activatedAtValues {
		if storedActivatedAt == nil || !activatedAt.Equal(*storedActivatedAt) {
			t.Fatalf("every response must report the single stored activation time, got %v want %v", activatedAt, storedActivatedAt)
		}
	}
}
