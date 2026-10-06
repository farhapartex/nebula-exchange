package service

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	signatureLength        = 65
	recoveryIDIndex        = 64
	legacyRecoveryIDOffset = 27
)

var errInvalidSignature = errors.New("invalid wallet signature")

func RecoverPersonalSigner(messageText string, signatureHex string) (string, error) {
	signature, err := hexutil.Decode(signatureHex)
	if err != nil || len(signature) != signatureLength {
		return "", errInvalidSignature
	}
	normalizedSignature := append([]byte(nil), signature...)
	if normalizedSignature[recoveryIDIndex] >= legacyRecoveryIDOffset {
		normalizedSignature[recoveryIDIndex] -= legacyRecoveryIDOffset
	}
	publicKey, err := crypto.SigToPub(accounts.TextHash([]byte(messageText)), normalizedSignature)
	if err != nil {
		return "", errInvalidSignature
	}
	return strings.ToLower(crypto.PubkeyToAddress(*publicKey).Hex()), nil
}
