package service

import (
	"strings"
	"testing"
	"time"
)

const (
	viemSignedAddress   = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	viemSignedMessage   = "localhost:3000 wants you to sign in with your Ethereum account:\n0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266\n\nLink this wallet to your Street Born account. This does not send a transaction or cost gas.\n\nURI: http://localhost:3000\nVersion: 1\nChain ID: 31337\nNonce: a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6\nIssued At: 2026-10-06T10:00:00.000Z"
	viemSignedSignature = "0x69488fd8c9fea7fb45abaea442f1233b86b19121d2c6fbb47317a7859657dcd544376636fe191959ed5d9f662db23d25cf795c0cafeb81a8df8532ace684bcbb1c"
)

func TestParseSignInMessageReadsAMessageMadeByTheBrowser(t *testing.T) {
	message, err := ParseSignInMessage(viemSignedMessage)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if message.Domain != "localhost:3000" || message.Address != viemSignedAddress || message.URI != "http://localhost:3000" || message.Version != "1" || message.ChainID != 31337 || message.Nonce != "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6" {
		t.Fatalf("unexpected message %+v", message)
	}
	if !message.IssuedAt.Equal(time.Date(2026, 10, 6, 10, 0, 0, 0, time.UTC)) || message.ExpirationTime != nil {
		t.Fatalf("unexpected times %+v", message)
	}
}

func TestParseSignInMessageRejectsBrokenMessages(t *testing.T) {
	brokenMessages := map[string]string{
		"not a sign-in request": "hello\nworld\n",
		"bad address":           strings.Replace(viemSignedMessage, viemSignedAddress, "0xnot-an-address", 1),
		"missing nonce":         strings.Replace(viemSignedMessage, "\nNonce: a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", "", 1),
		"duplicate chain":       viemSignedMessage + "\nChain ID: 1",
		"bad issued at":         strings.Replace(viemSignedMessage, "2026-10-06T10:00:00.000Z", "yesterday", 1),
	}
	for caseName, brokenMessage := range brokenMessages {
		if _, err := ParseSignInMessage(brokenMessage); err == nil {
			t.Errorf("%s: expected an error", caseName)
		}
	}
}

func TestRecoverPersonalSignerMatchesTheBrowserSignature(t *testing.T) {
	signerAddress, err := RecoverPersonalSigner(viemSignedMessage, viemSignedSignature)
	if err != nil || signerAddress != strings.ToLower(viemSignedAddress) {
		t.Fatalf("got %q, %v", signerAddress, err)
	}
	if otherSigner, _ := RecoverPersonalSigner(viemSignedMessage+" ", viemSignedSignature); otherSigner == strings.ToLower(viemSignedAddress) {
		t.Fatal("a changed message must not recover the same signer")
	}
	if _, err := RecoverPersonalSigner(viemSignedMessage, "0x1234"); err == nil {
		t.Fatal("a short signature must be rejected")
	}
}
