package gateway

import (
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	testSignerPrivateKey   = "0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	vaultUsedForVector     = "0x9EA86D9885d6344663f48A85286e2744695cD53F"
	referenceUsedForVector = "0x1111111111111111111111111111111111111111111111111111111111111111"
	payerUsedForVector     = "0x70997970c51812dc3a010c7d01b50e0d17dc79c8"
	digestFromTheContract  = "0x9061d71fc1be40b644baa30298d26617718ae118c7fc5c95f7738d1f253ad76d"
)

func TestTheDigestMatchesTheDeployedVault(t *testing.T) {
	authorizer, err := NewEIP712PaymentAuthorizer(testSignerPrivateKey, vaultUsedForVector, 31337)
	if err != nil {
		t.Fatalf("build authorizer: %v", err)
	}
	digest, err := authorizer.PaymentAuthorizationDigest(referenceUsedForVector, payerUsedForVector, 499, time.Unix(1791300000, 0))
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	if hexutil.Encode(digest) != digestFromTheContract {
		t.Fatalf("got %s, want the vault's %s", hexutil.Encode(digest), digestFromTheContract)
	}
}

func TestTheSignatureRecoversToThePaymentSigner(t *testing.T) {
	authorizer, _ := NewEIP712PaymentAuthorizer(testSignerPrivateKey, vaultUsedForVector, 31337)
	deadline := time.Unix(1791300000, 0)
	signatureHex, err := authorizer.SignPaymentAuthorization(referenceUsedForVector, payerUsedForVector, 499, deadline)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	signature := hexutil.MustDecode(signatureHex)
	if len(signature) != 65 || signature[64] < 27 {
		t.Fatalf("unexpected signature %s", signatureHex)
	}
	signature[64] -= 27
	digest, _ := authorizer.PaymentAuthorizationDigest(referenceUsedForVector, payerUsedForVector, 499, deadline)
	publicKey, err := crypto.SigToPub(digest, signature)
	if err != nil || !strings.EqualFold(crypto.PubkeyToAddress(*publicKey).Hex(), authorizer.SignerAddress()) {
		t.Fatalf("signature does not recover to the signer: %v", err)
	}
}

func TestAnInvalidSignerKeyIsRejected(t *testing.T) {
	if _, err := NewEIP712PaymentAuthorizer("not-a-key", vaultUsedForVector, 31337); err == nil {
		t.Fatal("expected an error for a broken key")
	}
}
