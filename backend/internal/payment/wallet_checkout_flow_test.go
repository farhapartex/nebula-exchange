package payment_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/service"
)

type walletCheckoutBody struct {
	ID               string `json:"id"`
	PaymentReference string `json:"payment_reference"`
	USDCents         string `json:"usd_cents"`
	Deadline         string `json:"deadline"`
	Signature        string `json:"signature"`
	VaultAddress     string `json:"vault_address"`
	ChainID          int64  `json:"chain_id"`
}

func (harness *paymentHarness) postJSON(path string, accessToken string, body any) (int, []byte) {
	harness.t.Helper()
	requestBody, _ := json.Marshal(body)
	request := httptest.NewRequest(http.MethodPost, "/api/v1"+path, bytes.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Nebula-Client", "web")
	request.Header.Set("Authorization", "Bearer "+accessToken)
	recorder := harness.serve(request)
	return recorder.Code, recorder.Body.Bytes()
}

func (harness *paymentHarness) startWalletCheckout(accessToken string, planID string, chapterCount int) (int, walletCheckoutBody, dataBody) {
	harness.t.Helper()
	status, rawBody := harness.postJSON("/checkout-sessions", accessToken, map[string]any{"plan_id": planID, "chapter_count": chapterCount, "payment_method": "WALLET"})
	var envelope struct {
		Data walletCheckoutBody `json:"data"`
	}
	_ = json.Unmarshal(rawBody, &envelope)
	var errorBody dataBody
	_ = json.Unmarshal(rawBody, &errorBody)
	return status, envelope.Data, errorBody
}

func (harness *paymentHarness) reportTransaction(accessToken string, checkoutID string, transactionHash string) (int, dataBody) {
	harness.t.Helper()
	status, rawBody := harness.postJSON("/checkout-sessions/"+checkoutID+"/transactions", accessToken, map[string]any{"transaction_hash": transactionHash})
	var body dataBody
	_ = json.Unmarshal(rawBody, &body)
	return status, body
}

func (harness *paymentHarness) linkPlayerWallet() {
	harness.linkedWallets[harness.playerID] = testPlayerWallet
}

func ethPaymentFor(checkout walletCheckoutBody, amountUnits string) service.ObservedVaultPayment {
	usdCents, _ := strconv.ParseInt(checkout.USDCents, 10, 64)
	return service.ObservedVaultPayment{
		PaymentReference: checkout.PaymentReference,
		PayerAddress:     testPlayerWallet,
		Asset:            models.WalletPaymentAssetETH,
		AmountUnits:      amountUnits,
		USDCents:         usdCents,
	}
}

func TestAWalletCheckoutIsSignedForTheLinkedWallet(t *testing.T) {
	harness := newPaymentHarness(t)
	if status, _, body := harness.startWalletCheckout(playerToken, "single-chapter", 1); status != http.StatusForbidden || body.Error.Code != "WALLET_REQUIRED" {
		t.Fatalf("without a linked wallet got %d %+v", status, body.Error)
	}
	harness.linkPlayerWallet()

	status, checkout, body := harness.startWalletCheckout(playerToken, "chapter-bundle", 3)
	if status != http.StatusCreated {
		t.Fatalf("got %d %+v", status, body.Error)
	}
	expectedReference := "0x" + hex.EncodeToString(crypto.Keccak256([]byte(checkout.ID)))
	if checkout.PaymentReference != expectedReference || checkout.USDCents != "1347" || checkout.ChainID != testChainID || !strings.EqualFold(checkout.VaultAddress, testVaultAddress) {
		t.Fatalf("unexpected checkout %+v", checkout)
	}
	deadlineSeconds, _ := strconv.ParseInt(checkout.Deadline, 10, 64)
	digest, _ := harness.authorizer.PaymentAuthorizationDigest(checkout.PaymentReference, testPlayerWallet, 1347, time.Unix(deadlineSeconds, 0))
	signature := hexutil.MustDecode(checkout.Signature)
	signature[64] -= 27
	publicKey, err := crypto.SigToPub(digest, signature)
	if err != nil || !strings.EqualFold(crypto.PubkeyToAddress(*publicKey).Hex(), harness.authorizer.SignerAddress()) {
		t.Fatalf("the signature is not from the payment signer for the linked wallet: %v", err)
	}
	if time.Until(time.Unix(deadlineSeconds, 0)) < 29*time.Minute {
		t.Fatalf("deadline is too short: %s", checkout.Deadline)
	}
}

func TestTheSameWalletCheckoutIsReusedAndADifferentOneReplacesIt(t *testing.T) {
	harness := newPaymentHarness(t)
	harness.linkPlayerWallet()
	_, firstCheckout, _ := harness.startWalletCheckout(playerToken, "single-chapter", 1)

	status, sameCheckout, _ := harness.startWalletCheckout(playerToken, "single-chapter", 1)
	if status != http.StatusOK || sameCheckout.ID != firstCheckout.ID || sameCheckout.Signature != firstCheckout.Signature {
		t.Fatalf("got %d %+v, want the first checkout reused", status, sameCheckout)
	}
	status, otherCheckout, _ := harness.startWalletCheckout(playerToken, "all-chapters", 4)
	if status != http.StatusCreated || otherCheckout.ID == firstCheckout.ID {
		t.Fatalf("got %d %+v, want a new checkout", status, otherCheckout)
	}
	if harness.paymentStatus(firstCheckout.ID) != "EXPIRED" || harness.paymentStatus(otherCheckout.ID) != "OPEN" {
		t.Fatal("the replaced checkout must be expired and the new one open")
	}
}

func TestAReportedTransactionPaysForTheChapters(t *testing.T) {
	harness := newPaymentHarness(t)
	harness.linkPlayerWallet()
	_, checkout, _ := harness.startWalletCheckout(playerToken, "chapter-bundle", 2)
	minedPayment := harness.walletChain.minePayment(ethPaymentFor(checkout, "315960000000000"))

	if status, body := harness.reportTransaction(playerToken, checkout.ID, "0x1234"); status != http.StatusUnprocessableEntity || body.Error.Details["transaction_hash"] == "" {
		t.Fatalf("a broken hash got %d %+v", status, body.Error)
	}
	if status, _ := harness.reportTransaction(otherPlayerToken, checkout.ID, minedPayment.TransactionHash); status != http.StatusNotFound {
		t.Fatalf("another player reporting got %d", status)
	}
	status, body := harness.reportTransaction(playerToken, checkout.ID, strings.ToUpper(minedPayment.TransactionHash[2:3])+minedPayment.TransactionHash[3:])
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("a hash without 0x got %d", status)
	}
	status, body = harness.reportTransaction(playerToken, checkout.ID, minedPayment.TransactionHash)
	if status != http.StatusOK || body.Data["status"] != "PAID" || body.Data["confirmations"] != float64(1) || body.Data["required_confirmations"] != float64(1) {
		t.Fatalf("got %d %v", status, body)
	}
	if owned := harness.ownedChapterNumbers(playerToken); !reflect.DeepEqual(owned, []float64{2, 3}) {
		t.Fatalf("got owned chapters %v", owned)
	}

	_, subscriptions := harness.get("/subscriptions", playerToken)
	if len(subscriptions.Data) != 1 {
		t.Fatalf("got %v", subscriptions.Data)
	}
	subscription := subscriptions.Data[0]
	walletPayment, _ := subscription["wallet_payment"].(map[string]any)
	if subscription["payment_method"] != "WALLET" || walletPayment["asset"] != "ETH" || walletPayment["amount_units"] != "315960000000000" || walletPayment["payer_address"] != testPlayerWallet || walletPayment["transaction_hash"] != minedPayment.TransactionHash {
		t.Fatalf("unexpected wallet subscription %v", subscription)
	}
}

func TestTheChainWatcherPaysWithoutAReportedTransaction(t *testing.T) {
	harness := newPaymentHarness(t)
	harness.linkPlayerWallet()
	ctx := context.Background()
	if _, err := harness.watcher.SyncOnce(ctx); err != nil {
		t.Fatalf("first sync: %v", err)
	}
	_, checkout, _ := harness.startWalletCheckout(playerToken, "single-chapter", 1)
	usdcPayment := ethPaymentFor(checkout, "4990000")
	usdcPayment.Asset = models.WalletPaymentAssetUSDC
	harness.walletChain.minePayment(usdcPayment)
	harness.walletChain.minePayment(service.ObservedVaultPayment{PaymentReference: "0x" + strings.Repeat("ab", 32), Asset: models.WalletPaymentAssetETH, AmountUnits: "1", USDCents: 499})

	settledCount, err := harness.watcher.SyncOnce(ctx)
	if err != nil || settledCount != 2 {
		t.Fatalf("got %d observations, %v", settledCount, err)
	}
	if harness.paymentStatus(checkout.ID) != "PAID" {
		t.Fatal("the watcher must pay the checkout")
	}
	if _, body := harness.getData("/checkout-sessions/"+checkout.ID, playerToken); body.Data["status"] != "PAID" {
		t.Fatalf("got %v", body.Data)
	}
}

func TestALatePaymentForAReplacedCheckoutStillUnlocksItsChapters(t *testing.T) {
	harness := newPaymentHarness(t)
	harness.linkPlayerWallet()
	ctx := context.Background()
	harness.watcher.SyncOnce(ctx)
	_, replacedCheckout, _ := harness.startWalletCheckout(playerToken, "single-chapter", 1)
	harness.startWalletCheckout(playerToken, "chapter-bundle", 2)
	if harness.paymentStatus(replacedCheckout.ID) != "EXPIRED" {
		t.Fatal("the first checkout must be replaced")
	}

	harness.walletChain.minePayment(ethPaymentFor(replacedCheckout, "1663333333333334"))
	harness.watcher.SyncOnce(ctx)
	if harness.paymentStatus(replacedCheckout.ID) != "PAID" {
		t.Fatal("money that reached the vault must unlock the chapters it was signed for")
	}
	if owned := harness.ownedChapterNumbers(playerToken); !reflect.DeepEqual(owned, []float64{2}) {
		t.Fatalf("got %v", owned)
	}
}

func TestAPaymentBelowThePriceIsNotAccepted(t *testing.T) {
	harness := newPaymentHarness(t)
	harness.linkPlayerWallet()
	_, checkout, _ := harness.startWalletCheckout(playerToken, "single-chapter", 1)
	underpaid := ethPaymentFor(checkout, "1")
	underpaid.USDCents = 1
	minedPayment := harness.walletChain.minePayment(underpaid)

	harness.reportTransaction(playerToken, checkout.ID, minedPayment.TransactionHash)
	if harness.paymentStatus(checkout.ID) != "OPEN" || len(harness.ownedChapterNumbers(playerToken)) != 0 {
		t.Fatal("an underpaid payment must not unlock chapters")
	}
}

func TestAPaymentWaitsForTheRequiredConfirmations(t *testing.T) {
	harness := newPaymentHarness(t, harnessOptions{requiredConfirmations: 3})
	harness.linkPlayerWallet()
	_, checkout, _ := harness.startWalletCheckout(playerToken, "single-chapter", 1)
	minedPayment := harness.walletChain.minePayment(ethPaymentFor(checkout, "1663333333333334"))

	_, body := harness.reportTransaction(playerToken, checkout.ID, minedPayment.TransactionHash)
	if body.Data["status"] != "OPEN" || body.Data["confirmations"] != float64(1) || body.Data["required_confirmations"] != float64(3) {
		t.Fatalf("got %v, want 1 of 3 confirmations", body.Data)
	}
	harness.walletChain.mineEmptyBlock()
	if _, body := harness.getData("/checkout-sessions/"+checkout.ID, playerToken); body.Data["confirmations"] != float64(2) || body.Data["status"] != "OPEN" {
		t.Fatalf("got %v, want 2 of 3 confirmations", body.Data)
	}
	harness.walletChain.mineEmptyBlock()
	if _, body := harness.getData("/checkout-sessions/"+checkout.ID, playerToken); body.Data["status"] != "PAID" {
		t.Fatalf("got %v, want paid after 3 confirmations", body.Data)
	}
}

func TestTheWatcherStartsOverWhenTheChainIsReset(t *testing.T) {
	harness := newPaymentHarness(t)
	ctx := context.Background()
	harness.watcher.SyncOnce(ctx)
	harness.walletChain.restart()
	harness.walletChain.mineEmptyBlock()
	if _, err := harness.watcher.SyncOnce(ctx); err != nil {
		t.Fatalf("sync after a chain reset: %v", err)
	}
	var cursor models.ChainSyncCursor
	harness.database.Where("name = ?", "chapter_payment_vault").Take(&cursor)
	head, _ := harness.walletChain.HeadBlockNumber(ctx)
	if uint64(cursor.BlockNumber) != head {
		t.Fatalf("cursor at %d, want the new head %d", cursor.BlockNumber, head)
	}
}

func TestACardCheckoutReplacesAnOpenWalletCheckout(t *testing.T) {
	harness := newPaymentHarness(t)
	harness.linkPlayerWallet()
	_, walletCheckout, _ := harness.startWalletCheckout(playerToken, "single-chapter", 1)
	if status, _ := harness.startCheckout(playerToken, "single-chapter", 1); status != http.StatusCreated {
		t.Fatalf("card checkout got %d", status)
	}
	if harness.paymentStatus(walletCheckout.ID) != "EXPIRED" {
		t.Fatal("the open wallet checkout must be replaced by the card checkout")
	}
}

func TestWalletPaymentsAreOffWithoutChainSettings(t *testing.T) {
	harness := newPaymentHarness(t, harnessOptions{withoutWalletPayments: true})
	if status, _, _ := harness.startWalletCheckout(playerToken, "single-chapter", 1); status != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503", status)
	}
}
