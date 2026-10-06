package wallet_test

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	identitymodels "github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/wallet"
)

const (
	firstPlayerToken  = "first-player-token"
	secondPlayerToken = "second-player-token"
	testChainID       = 31337
	frontendBaseURL   = "http://localhost:3000"
)

type fixedPlayers map[string]uuid.UUID

func (players fixedPlayers) Verify(accessToken string) (uuid.UUID, error) {
	userID, isKnown := players[accessToken]
	if !isKnown {
		return uuid.Nil, errors.New("invalid token")
	}
	return userID, nil
}

type walletHarness struct {
	t        *testing.T
	database *gorm.DB
	router   *gin.Engine
	now      time.Time
}

type responseBody struct {
	Data  json.RawMessage `json:"data"`
	Error struct {
		Code    string            `json:"code"`
		Details map[string]string `json:"details"`
	} `json:"error"`
}

func newWalletHarness(t *testing.T) *walletHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	testDatabase := databasetest.Open(t)
	players := fixedPlayers{firstPlayerToken: uuid.New(), secondPlayerToken: uuid.New()}
	activatedAt := time.Now()
	playerNumber := 0
	for _, playerID := range players {
		playerNumber++
		user := &identitymodels.User{ID: playerID, Email: fmt.Sprintf("player%d@streetborn.test", playerNumber), Username: fmt.Sprintf("player_%d", playerNumber), PasswordHash: "x", Status: identitymodels.UserStatusActive, IsActive: true, ActivatedAt: &activatedAt, TermsAcceptedAt: activatedAt}
		if err := testDatabase.Create(user).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}
	harness := &walletHarness{t: t, database: testDatabase, now: time.Now().UTC()}
	walletModule, err := wallet.NewModule(wallet.ModuleDependencies{
		Database:        testDatabase,
		ChainID:         testChainID,
		FrontendBaseURL: frontendBaseURL,
		Now:             func() time.Time { return harness.now },
	})
	if err != nil {
		t.Fatalf("build module: %v", err)
	}
	quietLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	harness.router, err = httpserver.NewRouter(httpserver.RouterOptions{Logger: quietLogger, AccessTokens: players}, walletModule.RouteRegistrars()...)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return harness
}

func (harness *walletHarness) send(method string, path string, accessToken string, body any) (int, responseBody) {
	harness.t.Helper()
	var requestBody io.Reader
	if body != nil {
		encodedBody, _ := json.Marshal(body)
		requestBody = bytes.NewReader(encodedBody)
	}
	request := httptest.NewRequest(method, "/api/v1"+path, requestBody)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Nebula-Client", "web")
	request.Header.Set("Authorization", "Bearer "+accessToken)
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	var decodedBody responseBody
	_ = json.Unmarshal(recorder.Body.Bytes(), &decodedBody)
	return recorder.Code, decodedBody
}

type testWallet struct {
	privateKey *ecdsa.PrivateKey
	address    string
}

func newTestWallet(t *testing.T) testWallet {
	t.Helper()
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return testWallet{privateKey: privateKey, address: crypto.PubkeyToAddress(privateKey.PublicKey).Hex()}
}

func (testWallet testWallet) sign(t *testing.T, message string) string {
	t.Helper()
	signature, err := crypto.Sign(accounts.TextHash([]byte(message)), testWallet.privateKey)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	signature[64] += 27
	return hexutil.Encode(signature)
}

type messageOptions struct {
	domain  string
	chainID int
}

func signInMessage(address string, nonce string, issuedAt time.Time, options ...messageOptions) string {
	domain, chainID := "localhost:3000", testChainID
	if len(options) > 0 {
		domain, chainID = options[0].domain, options[0].chainID
	}
	return fmt.Sprintf("%s wants you to sign in with your Ethereum account:\n%s\n\nLink this wallet to your Street Born account. This does not send a transaction or cost gas.\n\nURI: http://%s\nVersion: 1\nChain ID: %d\nNonce: %s\nIssued At: %s",
		domain, address, domain, chainID, nonce, issuedAt.Format("2006-01-02T15:04:05.000Z"))
}

func (harness *walletHarness) requestNonce(accessToken string, address string) string {
	harness.t.Helper()
	status, body := harness.send(http.MethodPost, "/wallet-challenges", accessToken, map[string]any{"address": address, "chain_id": testChainID})
	if status != http.StatusCreated {
		harness.t.Fatalf("challenge: got %d %+v", status, body.Error)
	}
	var challenge struct {
		Nonce string `json:"nonce"`
	}
	_ = json.Unmarshal(body.Data, &challenge)
	return challenge.Nonce
}

func (harness *walletHarness) link(accessToken string, signer testWallet, message string) (int, responseBody) {
	harness.t.Helper()
	return harness.send(http.MethodPost, "/wallets", accessToken, map[string]any{"message": message, "signature": signer.sign(harness.t, message)})
}

func TestAPlayerLinksTheirWalletWithASignedMessage(t *testing.T) {
	harness := newWalletHarness(t)
	playerWallet := newTestWallet(t)
	message := signInMessage(playerWallet.address, harness.requestNonce(firstPlayerToken, playerWallet.address), harness.now)

	status, body := harness.link(firstPlayerToken, playerWallet, message)
	if status != http.StatusCreated {
		t.Fatalf("got %d %+v", status, body.Error)
	}
	var linkedWallet struct {
		Address string `json:"address"`
		ChainID int64  `json:"chain_id"`
	}
	_ = json.Unmarshal(body.Data, &linkedWallet)
	if linkedWallet.Address != strings.ToLower(playerWallet.address) || linkedWallet.ChainID != testChainID {
		t.Fatalf("unexpected wallet %+v", linkedWallet)
	}

	listStatus, listBody := harness.send(http.MethodGet, "/wallets", firstPlayerToken, nil)
	var wallets []map[string]any
	_ = json.Unmarshal(listBody.Data, &wallets)
	if listStatus != http.StatusOK || len(wallets) != 1 || wallets[0]["address"] != strings.ToLower(playerWallet.address) {
		t.Fatalf("got %d %v", listStatus, wallets)
	}
	_, otherBody := harness.send(http.MethodGet, "/wallets", secondPlayerToken, nil)
	if string(otherBody.Data) != "[]" {
		t.Fatalf("another player saw the wallet: %s", otherBody.Data)
	}

	if status, body := harness.link(firstPlayerToken, playerWallet, message); status != http.StatusUnprocessableEntity || body.Error.Details["message"] == "" {
		t.Fatalf("replaying the message got %d %+v, want 422", status, body.Error)
	}
	freshMessage := signInMessage(playerWallet.address, harness.requestNonce(firstPlayerToken, playerWallet.address), harness.now)
	if status, _ := harness.link(firstPlayerToken, playerWallet, freshMessage); status != http.StatusOK {
		t.Fatalf("linking the same wallet again got %d, want 200", status)
	}
}

func TestAWalletBelongsToOneAccountAndAnAccountToOneWallet(t *testing.T) {
	harness := newWalletHarness(t)
	sharedWallet, secondWallet := newTestWallet(t), newTestWallet(t)
	harness.link(firstPlayerToken, sharedWallet, signInMessage(sharedWallet.address, harness.requestNonce(firstPlayerToken, sharedWallet.address), harness.now))

	takenMessage := signInMessage(sharedWallet.address, harness.requestNonce(secondPlayerToken, sharedWallet.address), harness.now)
	if status, body := harness.link(secondPlayerToken, sharedWallet, takenMessage); status != http.StatusUnprocessableEntity || body.Error.Details["address"] == "" {
		t.Fatalf("got %d %+v, want 422 on address", status, body.Error)
	}
	secondMessage := signInMessage(secondWallet.address, harness.requestNonce(firstPlayerToken, secondWallet.address), harness.now)
	if status, body := harness.link(firstPlayerToken, secondWallet, secondMessage); status != http.StatusConflict {
		t.Fatalf("got %d %+v, want 409", status, body.Error)
	}
}

func TestForgedOrStaleMessagesAreRejected(t *testing.T) {
	harness := newWalletHarness(t)
	playerWallet, attackerWallet := newTestWallet(t), newTestWallet(t)

	forgedMessage := signInMessage(playerWallet.address, harness.requestNonce(firstPlayerToken, playerWallet.address), harness.now)
	if status, body := harness.link(firstPlayerToken, attackerWallet, forgedMessage); status != http.StatusUnprocessableEntity || body.Error.Details["signature"] == "" {
		t.Fatalf("a message signed by another wallet got %d %+v", status, body.Error)
	}

	otherSiteMessage := signInMessage(playerWallet.address, harness.requestNonce(firstPlayerToken, playerWallet.address), harness.now, messageOptions{domain: "evil.test", chainID: testChainID})
	if status, body := harness.link(firstPlayerToken, playerWallet, otherSiteMessage); status != http.StatusUnprocessableEntity || body.Error.Details["message"] != "was made for another site" {
		t.Fatalf("a message for another site got %d %+v", status, body.Error)
	}

	otherChainMessage := signInMessage(playerWallet.address, harness.requestNonce(firstPlayerToken, playerWallet.address), harness.now, messageOptions{domain: "localhost:3000", chainID: 1})
	if status, body := harness.link(firstPlayerToken, playerWallet, otherChainMessage); status != http.StatusUnprocessableEntity || body.Error.Details["message"] != "is for another network" {
		t.Fatalf("a message for another chain got %d %+v", status, body.Error)
	}

	foreignNonceMessage := signInMessage(playerWallet.address, harness.requestNonce(secondPlayerToken, playerWallet.address), harness.now)
	if status, _ := harness.link(firstPlayerToken, playerWallet, foreignNonceMessage); status != http.StatusUnprocessableEntity {
		t.Fatalf("a nonce issued to another player got %d", status)
	}

	expiringMessage := signInMessage(playerWallet.address, harness.requestNonce(firstPlayerToken, playerWallet.address), harness.now)
	harness.now = harness.now.Add(6 * time.Minute)
	if status, _ := harness.link(firstPlayerToken, playerWallet, expiringMessage); status != http.StatusUnprocessableEntity {
		t.Fatalf("an expired challenge got %d", status)
	}

	var walletCount int64
	harness.database.Table("wallets").Count(&walletCount)
	if walletCount != 0 {
		t.Fatalf("got %d wallets, want none", walletCount)
	}
}

func TestChallengesNeedARealAddressOnTheGameChain(t *testing.T) {
	harness := newWalletHarness(t)
	if status, body := harness.send(http.MethodPost, "/wallet-challenges", firstPlayerToken, map[string]any{"address": "0x123", "chain_id": testChainID}); status != http.StatusUnprocessableEntity || body.Error.Details["address"] == "" {
		t.Fatalf("got %d %+v", status, body.Error)
	}
	if status, body := harness.send(http.MethodPost, "/wallet-challenges", firstPlayerToken, map[string]any{"address": newTestWallet(t).address, "chain_id": 1}); status != http.StatusUnprocessableEntity || body.Error.Details["chain_id"] == "" {
		t.Fatalf("got %d %+v", status, body.Error)
	}
	if status, _ := harness.send(http.MethodPost, "/wallet-challenges", "", map[string]any{"address": newTestWallet(t).address, "chain_id": testChainID}); status != http.StatusUnauthorized {
		t.Fatalf("got %d without a login", status)
	}
}
