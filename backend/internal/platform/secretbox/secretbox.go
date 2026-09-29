package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

const keyLength = 32

var ErrCiphertextInvalid = errors.New("ciphertext is invalid or was encrypted with another key")

type Box struct {
	aead cipher.AEAD
}

func NewFromBase64Key(encodedKey string) (*Box, error) {
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil || len(key) != keyLength {
		return nil, fmt.Errorf("encryption key must be %d bytes encoded as base64", keyLength)
	}
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}
	return &Box{aead: aead}, nil
}

func (box *Box) Seal(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, box.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return box.aead.Seal(nonce, nonce, plaintext, nil), nil
}

func (box *Box) Open(ciphertext []byte) ([]byte, error) {
	nonceSize := box.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrCiphertextInvalid
	}
	plaintext, err := box.aead.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], nil)
	if err != nil {
		return nil, ErrCiphertextInvalid
	}
	return plaintext, nil
}
