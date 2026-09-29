package secretbox

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"testing"
)

func newTestBox(t *testing.T) *Box {
	t.Helper()
	key := make([]byte, keyLength)
	_, _ = rand.Read(key)
	box, err := NewFromBase64Key(base64.StdEncoding.EncodeToString(key))
	if err != nil {
		t.Fatalf("create box: %v", err)
	}
	return box
}

func TestSealAndOpenRoundTrip(t *testing.T) {
	box := newTestBox(t)
	ciphertext, err := box.Seal([]byte("JBSWY3DPEHPK3PXP"))
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	if bytes.Contains(ciphertext, []byte("JBSWY3DPEHPK3PXP")) {
		t.Fatal("ciphertext must not contain the plaintext")
	}
	plaintext, err := box.Open(ciphertext)
	if err != nil || string(plaintext) != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("open: got %q, %v", plaintext, err)
	}
}

func TestOpenRejectsTamperedOrForeignCiphertext(t *testing.T) {
	box := newTestBox(t)
	ciphertext, _ := box.Seal([]byte("secret"))
	ciphertext[len(ciphertext)-1] ^= 0xFF
	if _, err := box.Open(ciphertext); !errors.Is(err, ErrCiphertextInvalid) {
		t.Fatalf("tampered: got %v", err)
	}
	otherCiphertext, _ := newTestBox(t).Seal([]byte("secret"))
	if _, err := box.Open(otherCiphertext); !errors.Is(err, ErrCiphertextInvalid) {
		t.Fatalf("foreign key: got %v", err)
	}
}

func TestNewRejectsBadKeys(t *testing.T) {
	for _, badKey := range []string{"", "not-base64!!", base64.StdEncoding.EncodeToString([]byte("too-short"))} {
		if _, err := NewFromBase64Key(badKey); err == nil {
			t.Fatalf("expected an error for %q", badKey)
		}
	}
}
