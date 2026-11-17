package service_test

import (
	"context"
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/economy/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/economy/service"
)

type inMemoryCoinStore struct {
	balances map[uuid.UUID]int64
	entries  []models.CoinLedgerEntry
	flags    []models.CoinClawbackFlag
}

func newInMemoryCoinStore() *inMemoryCoinStore {
	return &inMemoryCoinStore{balances: map[uuid.UUID]int64{}}
}

func (store *inMemoryCoinStore) CoinsOf(_ context.Context, userID uuid.UUID) (int64, error) {
	return store.balances[userID], nil
}

func (store *inMemoryCoinStore) LockOrCreate(_ context.Context, userID uuid.UUID, createdAt time.Time) (models.CoinBalance, error) {
	return models.CoinBalance{UserID: userID, Coins: store.balances[userID], UpdatedAt: createdAt}, nil
}

func (store *inMemoryCoinStore) SetCoins(_ context.Context, userID uuid.UUID, coins int64, _ time.Time) error {
	store.balances[userID] = coins
	return nil
}

func (store *inMemoryCoinStore) Append(_ context.Context, entry *models.CoinLedgerEntry) error {
	store.entries = append(store.entries, *entry)
	return nil
}

func (store *inMemoryCoinStore) HasWelcomeGift(_ context.Context, userID uuid.UUID) (bool, error) {
	for _, entry := range store.entries {
		if entry.UserID == userID && entry.Reason == models.CoinEntryReasonWelcomeGift {
			return true, nil
		}
	}
	return false, nil
}

func (store *inMemoryCoinStore) HasPaymentEntry(_ context.Context, paymentID uuid.UUID, reason models.CoinEntryReason) (bool, error) {
	for _, entry := range store.entries {
		if entry.PaymentID != nil && *entry.PaymentID == paymentID && entry.Reason == reason {
			return true, nil
		}
	}
	return false, nil
}

func (store *inMemoryCoinStore) Create(_ context.Context, flag *models.CoinClawbackFlag) error {
	store.flags = append(store.flags, *flag)
	return nil
}

func (store *inMemoryCoinStore) ExistsForPayment(_ context.Context, paymentID uuid.UUID) (bool, error) {
	for _, flag := range store.flags {
		if flag.PaymentID == paymentID {
			return true, nil
		}
	}
	return false, nil
}

type immediateTransactions struct{}

func (immediateTransactions) WithinTransaction(ctx context.Context, work func(ctx context.Context) error) error {
	return work(ctx)
}

func newCoinAccounts(store *inMemoryCoinStore) service.CoinAccountService {
	return service.NewCoinAccountService(service.CoinAccountDependencies{
		Balances:      store,
		Ledger:        store,
		ClawbackFlags: store,
		Transactions:  immediateTransactions{},
		Logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		Now:           time.Now,
	})
}

func ledgerAmounts(store *inMemoryCoinStore) []int64 {
	amounts := make([]int64, 0, len(store.entries))
	for _, entry := range store.entries {
		amounts = append(amounts, entry.Amount)
	}
	return amounts
}

func TestWelcomeGiftIsGrantedOnlyOnce(t *testing.T) {
	store := newInMemoryCoinStore()
	coinAccounts := newCoinAccounts(store)
	playerID := uuid.New()

	for range 2 {
		if err := coinAccounts.GrantWelcomeGift(context.Background(), playerID); err != nil {
			t.Fatalf("grant welcome gift: %v", err)
		}
	}
	if balance, _ := coinAccounts.Balance(context.Background(), playerID); balance != service.WelcomeGiftCoins {
		t.Fatalf("got %d coins, want %d", balance, service.WelcomeGiftCoins)
	}
	if amounts := ledgerAmounts(store); !reflect.DeepEqual(amounts, []int64{service.WelcomeGiftCoins}) {
		t.Fatalf("got ledger %v", amounts)
	}
}

func TestCoinPackIsCreditedOncePerPaymentWithTheBalanceAfterEachEntry(t *testing.T) {
	store := newInMemoryCoinStore()
	coinAccounts := newCoinAccounts(store)
	playerID, firstPaymentID, secondPaymentID := uuid.New(), uuid.New(), uuid.New()

	_ = coinAccounts.CreditCoinPack(context.Background(), playerID, firstPaymentID, 500)
	_ = coinAccounts.CreditCoinPack(context.Background(), playerID, firstPaymentID, 500)
	_ = coinAccounts.CreditCoinPack(context.Background(), playerID, secondPaymentID, 1100)

	if store.balances[playerID] != 1600 {
		t.Fatalf("got %d coins, want 1600", store.balances[playerID])
	}
	if store.entries[1].BalanceAfter != 1600 || len(store.entries) != 2 {
		t.Fatalf("unexpected ledger %+v", store.entries)
	}
	if err := coinAccounts.CreditCoinPack(context.Background(), playerID, uuid.New(), 0); err == nil {
		t.Fatal("a pack without coins must be rejected")
	}
}

func TestClawbackTakesBackUnspentCoinsWithoutAFlag(t *testing.T) {
	store := newInMemoryCoinStore()
	coinAccounts := newCoinAccounts(store)
	playerID, paymentID := uuid.New(), uuid.New()
	store.balances[playerID] = 700

	_ = coinAccounts.ClawBackCoinPack(context.Background(), playerID, paymentID, 500)
	_ = coinAccounts.ClawBackCoinPack(context.Background(), playerID, paymentID, 500)

	if store.balances[playerID] != 200 || len(store.flags) != 0 {
		t.Fatalf("got %d coins and %d flags", store.balances[playerID], len(store.flags))
	}
	if amounts := ledgerAmounts(store); !reflect.DeepEqual(amounts, []int64{-500}) {
		t.Fatalf("got ledger %v", amounts)
	}
}

func TestClawbackNeverGoesBelowZeroAndFlagsTheMissingCoins(t *testing.T) {
	store := newInMemoryCoinStore()
	coinAccounts := newCoinAccounts(store)
	playerID, paymentID := uuid.New(), uuid.New()
	store.balances[playerID] = 200

	_ = coinAccounts.ClawBackCoinPack(context.Background(), playerID, paymentID, 500)
	_ = coinAccounts.ClawBackCoinPack(context.Background(), playerID, paymentID, 500)

	if store.balances[playerID] != 0 {
		t.Fatalf("got %d coins, want 0", store.balances[playerID])
	}
	if len(store.flags) != 1 || store.flags[0].ClawedBackCoins != 200 || store.flags[0].MissingCoins != 300 {
		t.Fatalf("got flags %+v", store.flags)
	}
}

func TestClawbackOfAnEmptyBalanceOnlyFlags(t *testing.T) {
	store := newInMemoryCoinStore()
	coinAccounts := newCoinAccounts(store)
	playerID, paymentID := uuid.New(), uuid.New()

	_ = coinAccounts.ClawBackCoinPack(context.Background(), playerID, paymentID, 500)
	_ = coinAccounts.ClawBackCoinPack(context.Background(), playerID, paymentID, 500)

	if len(store.entries) != 0 || len(store.flags) != 1 || store.flags[0].MissingCoins != 500 {
		t.Fatalf("got entries %+v and flags %+v", store.entries, store.flags)
	}
}
