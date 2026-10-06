package gateway

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const (
	signingDomainName      = "StreetBornChapterPaymentVault"
	signingDomainVersion   = "1"
	chapterPaymentType     = "ChapterPayment"
	eip712DomainType       = "EIP712Domain"
	recoveryIDIndex        = 64
	legacyRecoveryIDOffset = 27
)

var errInvalidSignerKey = errors.New("PAYMENT_SIGNER_PRIVATE_KEY is not a valid private key")

type EIP712PaymentAuthorizer struct {
	signerKey    *ecdsa.PrivateKey
	vaultAddress common.Address
	chainID      int64
}

func NewEIP712PaymentAuthorizer(signerPrivateKeyHex string, vaultAddress string, chainID int64) (*EIP712PaymentAuthorizer, error) {
	signerKey, err := crypto.HexToECDSA(strings.TrimPrefix(signerPrivateKeyHex, "0x"))
	if err != nil {
		return nil, errInvalidSignerKey
	}
	if !common.IsHexAddress(vaultAddress) {
		return nil, errors.New("CHAPTER_PAYMENT_VAULT_ADDRESS is not a valid address")
	}
	return &EIP712PaymentAuthorizer{signerKey: signerKey, vaultAddress: common.HexToAddress(vaultAddress), chainID: chainID}, nil
}

func (authorizer *EIP712PaymentAuthorizer) VaultAddress() string {
	return authorizer.vaultAddress.Hex()
}

func (authorizer *EIP712PaymentAuthorizer) ChainID() int64 {
	return authorizer.chainID
}

func (authorizer *EIP712PaymentAuthorizer) SignerAddress() string {
	return crypto.PubkeyToAddress(authorizer.signerKey.PublicKey).Hex()
}

func (authorizer *EIP712PaymentAuthorizer) PaymentAuthorizationDigest(paymentReference string, payerAddress string, usdCents int64, deadline time.Time) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			eip712DomainType: {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			chapterPaymentType: {
				{Name: "paymentReference", Type: "bytes32"},
				{Name: "payer", Type: "address"},
				{Name: "usdCents", Type: "uint256"},
				{Name: "deadline", Type: "uint256"},
			},
		},
		PrimaryType: chapterPaymentType,
		Domain: apitypes.TypedDataDomain{
			Name:              signingDomainName,
			Version:           signingDomainVersion,
			ChainId:           (*math.HexOrDecimal256)(big.NewInt(authorizer.chainID)),
			VerifyingContract: authorizer.vaultAddress.Hex(),
		},
		Message: apitypes.TypedDataMessage{
			"paymentReference": paymentReference,
			"payer":            common.HexToAddress(payerAddress).Hex(),
			"usdCents":         strconv.FormatInt(usdCents, 10),
			"deadline":         strconv.FormatInt(deadline.Unix(), 10),
		},
	}
	digest, _, err := apitypes.TypedDataAndHash(typedData)
	return digest, err
}

func (authorizer *EIP712PaymentAuthorizer) SignPaymentAuthorization(paymentReference string, payerAddress string, usdCents int64, deadline time.Time) (string, error) {
	digest, err := authorizer.PaymentAuthorizationDigest(paymentReference, payerAddress, usdCents, deadline)
	if err != nil {
		return "", err
	}
	signature, err := crypto.Sign(digest, authorizer.signerKey)
	if err != nil {
		return "", err
	}
	signature[recoveryIDIndex] += legacyRecoveryIDOffset
	return hexutil.Encode(signature), nil
}
