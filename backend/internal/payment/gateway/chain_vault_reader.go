package gateway

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/service"
)

const (
	paymentReceivedTopicCount = 4
	paymentReceivedDataLength = 64
	successfulReceiptStatus   = 1
)

var (
	paymentReceivedTopic    = crypto.Keccak256Hash([]byte("PaymentReceived(bytes32,address,address,uint256,uint256)"))
	paymentTokenSelector    = crypto.Keccak256([]byte("paymentToken()"))[:4]
	errUnexpectedEventShape = errors.New("unexpected PaymentReceived log shape")
)

type ChainVaultReader struct {
	client             *ethclient.Client
	vaultAddress       common.Address
	chainID            int64
	paymentTokenMutex  sync.Mutex
	paymentTokenCached *common.Address
}

func NewChainVaultReader(ctx context.Context, rpcURL string, vaultAddress string, chainID int64) (*ChainVaultReader, error) {
	if !common.IsHexAddress(vaultAddress) {
		return nil, errors.New("CHAPTER_PAYMENT_VAULT_ADDRESS is not a valid address")
	}
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial chain rpc: %w", err)
	}
	return &ChainVaultReader{client: client, vaultAddress: common.HexToAddress(vaultAddress), chainID: chainID}, nil
}

func (reader *ChainVaultReader) ChainID() int64 {
	return reader.chainID
}

func (reader *ChainVaultReader) HeadBlockNumber(ctx context.Context) (uint64, error) {
	return reader.client.BlockNumber(ctx)
}

func (reader *ChainVaultReader) BlockHash(ctx context.Context, blockNumber uint64) (string, error) {
	header, err := reader.client.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNumber))
	if err != nil {
		return "", err
	}
	return strings.ToLower(header.Hash().Hex()), nil
}

func (reader *ChainVaultReader) PaymentsInBlockRange(ctx context.Context, fromBlock uint64, toBlock uint64) ([]service.ObservedVaultPayment, error) {
	vaultLogs, err := reader.client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{reader.vaultAddress},
		Topics:    [][]common.Hash{{paymentReceivedTopic}},
	})
	if err != nil {
		return nil, err
	}
	return reader.decodePaymentLogs(ctx, vaultLogs)
}

func (reader *ChainVaultReader) PaymentsInTransaction(ctx context.Context, transactionHash string) ([]service.ObservedVaultPayment, error) {
	receipt, err := reader.client.TransactionReceipt(ctx, common.HexToHash(transactionHash))
	if errors.Is(err, ethereum.NotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if receipt.Status != successfulReceiptStatus {
		return nil, nil
	}
	receiptLogs := make([]types.Log, 0, len(receipt.Logs))
	for _, receiptLog := range receipt.Logs {
		receiptLogs = append(receiptLogs, *receiptLog)
	}
	return reader.decodePaymentLogs(ctx, receiptLogs)
}

func (reader *ChainVaultReader) decodePaymentLogs(ctx context.Context, vaultLogs []types.Log) ([]service.ObservedVaultPayment, error) {
	var observations []service.ObservedVaultPayment
	for _, vaultLog := range vaultLogs {
		isPaymentLog := vaultLog.Address == reader.vaultAddress && len(vaultLog.Topics) > 0 && vaultLog.Topics[0] == paymentReceivedTopic
		if !isPaymentLog || vaultLog.Removed {
			continue
		}
		observation, err := reader.decodePaymentLog(ctx, vaultLog)
		if err != nil {
			return nil, err
		}
		observations = append(observations, observation)
	}
	return observations, nil
}

func (reader *ChainVaultReader) decodePaymentLog(ctx context.Context, vaultLog types.Log) (service.ObservedVaultPayment, error) {
	if len(vaultLog.Topics) != paymentReceivedTopicCount || len(vaultLog.Data) != paymentReceivedDataLength {
		return service.ObservedVaultPayment{}, errUnexpectedEventShape
	}
	asset, err := reader.assetOf(ctx, common.BytesToAddress(vaultLog.Topics[3].Bytes()))
	if err != nil {
		return service.ObservedVaultPayment{}, err
	}
	usdCents := new(big.Int).SetBytes(vaultLog.Data[32:64])
	if !usdCents.IsInt64() {
		return service.ObservedVaultPayment{}, errUnexpectedEventShape
	}
	return service.ObservedVaultPayment{
		PaymentReference: strings.ToLower(vaultLog.Topics[1].Hex()),
		PayerAddress:     strings.ToLower(common.BytesToAddress(vaultLog.Topics[2].Bytes()).Hex()),
		Asset:            asset,
		AmountUnits:      new(big.Int).SetBytes(vaultLog.Data[0:32]).String(),
		USDCents:         usdCents.Int64(),
		TransactionHash:  strings.ToLower(vaultLog.TxHash.Hex()),
		BlockNumber:      vaultLog.BlockNumber,
	}, nil
}

func (reader *ChainVaultReader) assetOf(ctx context.Context, assetAddress common.Address) (models.WalletPaymentAsset, error) {
	if assetAddress == (common.Address{}) {
		return models.WalletPaymentAssetETH, nil
	}
	paymentToken, err := reader.paymentToken(ctx)
	if err != nil {
		return "", err
	}
	if assetAddress != paymentToken {
		return "", fmt.Errorf("vault payment in unknown asset %s", assetAddress.Hex())
	}
	return models.WalletPaymentAssetUSDC, nil
}

func (reader *ChainVaultReader) paymentToken(ctx context.Context) (common.Address, error) {
	reader.paymentTokenMutex.Lock()
	defer reader.paymentTokenMutex.Unlock()
	if reader.paymentTokenCached != nil {
		return *reader.paymentTokenCached, nil
	}
	result, err := reader.client.CallContract(ctx, ethereum.CallMsg{To: &reader.vaultAddress, Data: paymentTokenSelector}, nil)
	if err != nil {
		return common.Address{}, err
	}
	if len(result) != 32 {
		return common.Address{}, errors.New("vault returned an unexpected payment token")
	}
	paymentToken := common.BytesToAddress(result)
	reader.paymentTokenCached = &paymentToken
	return paymentToken, nil
}
