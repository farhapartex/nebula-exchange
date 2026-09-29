package securetoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const (
	byteLength      = 32
	PlaintextLength = 43
)

type Generated struct {
	Plaintext string
	Hash      []byte
}

func Generate() (Generated, error) {
	randomBytes := make([]byte, byteLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return Generated{}, fmt.Errorf("generate secure token: %w", err)
	}
	plaintextToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	return Generated{Plaintext: plaintextToken, Hash: Hash(plaintextToken)}, nil
}

func Hash(plaintextToken string) []byte {
	tokenHash := sha256.Sum256([]byte(plaintextToken))
	return tokenHash[:]
}

func HasValidShape(plaintextToken string) bool {
	return len(plaintextToken) == PlaintextLength
}
