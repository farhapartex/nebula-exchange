package securetoken

import (
	"bytes"
	"testing"
)

func TestGenerateProducesUniqueUrlSafeTokens(t *testing.T) {
	firstToken, err := Generate()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	secondToken, _ := Generate()

	if firstToken.Plaintext == secondToken.Plaintext {
		t.Fatal("tokens must be unique")
	}
	if !HasValidShape(firstToken.Plaintext) {
		t.Fatalf("unexpected token length %d", len(firstToken.Plaintext))
	}
	if !bytes.Equal(firstToken.Hash, Hash(firstToken.Plaintext)) {
		t.Fatal("hash must be derived from the plaintext")
	}
}
