package payment_test

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/service"
)

const (
	testPaymentSignerKey = "0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	testVaultAddress     = "0x9EA86D9885d6344663f48A85286e2744695cD53F"
	testChainID          = 31337
	testPlayerWallet     = "0x70997970c51812dc3a010c7d01b50e0d17dc79c8"
)

type fakeLinkedWallets map[uuid.UUID]string

func (linkedWallets fakeLinkedWallets) LinkedWalletAddress(_ context.Context, userID uuid.UUID) (string, bool, error) {
	address, isLinked := linkedWallets[userID]
	return address, isLinked, nil
}

type fakeVaultChain struct {
	mutex           sync.Mutex
	headBlock       uint64
	chainGeneration int
	minedPayments   []service.ObservedVaultPayment
}

func newFakeVaultChain() *fakeVaultChain {
	return &fakeVaultChain{headBlock: 100}
}

func (chain *fakeVaultChain) ChainID() int64 {
	return testChainID
}

func (chain *fakeVaultChain) HeadBlockNumber(context.Context) (uint64, error) {
	chain.mutex.Lock()
	defer chain.mutex.Unlock()
	return chain.headBlock, nil
}

func (chain *fakeVaultChain) BlockHash(_ context.Context, blockNumber uint64) (string, error) {
	chain.mutex.Lock()
	defer chain.mutex.Unlock()
	return fmt.Sprintf("0x%062x%02x", blockNumber, chain.chainGeneration), nil
}

func (chain *fakeVaultChain) PaymentsInBlockRange(_ context.Context, fromBlock uint64, toBlock uint64) ([]service.ObservedVaultPayment, error) {
	chain.mutex.Lock()
	defer chain.mutex.Unlock()
	var observations []service.ObservedVaultPayment
	for _, minedPayment := range chain.minedPayments {
		if minedPayment.BlockNumber >= fromBlock && minedPayment.BlockNumber <= toBlock {
			observations = append(observations, minedPayment)
		}
	}
	return observations, nil
}

func (chain *fakeVaultChain) PaymentsInTransaction(_ context.Context, transactionHash string) ([]service.ObservedVaultPayment, error) {
	chain.mutex.Lock()
	defer chain.mutex.Unlock()
	var observations []service.ObservedVaultPayment
	for _, minedPayment := range chain.minedPayments {
		if minedPayment.TransactionHash == transactionHash {
			observations = append(observations, minedPayment)
		}
	}
	return observations, nil
}

func (chain *fakeVaultChain) minePayment(observation service.ObservedVaultPayment) service.ObservedVaultPayment {
	chain.mutex.Lock()
	defer chain.mutex.Unlock()
	chain.headBlock++
	observation.BlockNumber = chain.headBlock
	if observation.TransactionHash == "" {
		observation.TransactionHash = fmt.Sprintf("0x%064x", chain.headBlock)
	}
	chain.minedPayments = append(chain.minedPayments, observation)
	return observation
}

func (chain *fakeVaultChain) mineEmptyBlock() {
	chain.mutex.Lock()
	defer chain.mutex.Unlock()
	chain.headBlock++
}

func (chain *fakeVaultChain) restart() {
	chain.mutex.Lock()
	defer chain.mutex.Unlock()
	chain.chainGeneration++
	chain.minedPayments = nil
}
