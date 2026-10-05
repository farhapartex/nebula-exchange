package securetoken

import (
	"bytes"
	"testing"
)

func TestGenerateProducesUniqueTokensWithMatchingHashes(t *testing.T) {
	firstToken, err := Generate()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	secondToken, _ := Generate()

	if firstToken.Plaintext == secondToken.Plaintext {
		t.Fatal("two generated tokens must differ")
	}
	if !HasValidShape(firstToken.Plaintext) {
		t.Fatalf("token %q has an unexpected length", firstToken.Plaintext)
	}
	if !bytes.Equal(firstToken.Hash, Hash(firstToken.Plaintext)) {
		t.Fatal("stored hash must match the plaintext token")
	}
}
