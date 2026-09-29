package activation

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const tokenByteLength = 32

type GeneratedToken struct {
	Plaintext string
	Hash      []byte
}

func generateToken() (GeneratedToken, error) {
	randomBytes := make([]byte, tokenByteLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return GeneratedToken{}, fmt.Errorf("generate activation token: %w", err)
	}
	plaintextToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	return GeneratedToken{Plaintext: plaintextToken, Hash: HashToken(plaintextToken)}, nil
}

func HashToken(plaintextToken string) []byte {
	tokenHash := sha256.Sum256([]byte(plaintextToken))
	return tokenHash[:]
}
