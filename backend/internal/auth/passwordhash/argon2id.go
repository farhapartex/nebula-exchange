package passwordhash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Parameters struct {
	MemoryInKibibytes uint32
	Iterations        uint32
	Parallelism       uint8
	SaltLength        uint32
	KeyLength         uint32
}

var DefaultParameters = Parameters{
	MemoryInKibibytes: 19 * 1024,
	Iterations:        2,
	Parallelism:       1,
	SaltLength:        16,
	KeyLength:         32,
}

var ErrMalformedHash = errors.New("password hash is malformed")

func Hash(password string, parameters Parameters) (string, error) {
	salt := make([]byte, parameters.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	derivedKey := argon2.IDKey([]byte(password), salt, parameters.Iterations, parameters.MemoryInKibibytes, parameters.Parallelism, parameters.KeyLength)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		parameters.MemoryInKibibytes,
		parameters.Iterations,
		parameters.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(derivedKey),
	), nil
}

func Verify(password, encodedHash string) (bool, error) {
	parameters, salt, expectedKey, err := decode(encodedHash)
	if err != nil {
		return false, err
	}
	derivedKey := argon2.IDKey([]byte(password), salt, parameters.Iterations, parameters.MemoryInKibibytes, parameters.Parallelism, uint32(len(expectedKey)))
	return subtle.ConstantTimeCompare(derivedKey, expectedKey) == 1, nil
}

func decode(encodedHash string) (Parameters, []byte, []byte, error) {
	hashParts := strings.Split(encodedHash, "$")
	if len(hashParts) != 6 || hashParts[1] != "argon2id" {
		return Parameters{}, nil, nil, ErrMalformedHash
	}

	var algorithmVersion int
	if _, err := fmt.Sscanf(hashParts[2], "v=%d", &algorithmVersion); err != nil || algorithmVersion != argon2.Version {
		return Parameters{}, nil, nil, ErrMalformedHash
	}

	var parameters Parameters
	if _, err := fmt.Sscanf(hashParts[3], "m=%d,t=%d,p=%d", &parameters.MemoryInKibibytes, &parameters.Iterations, &parameters.Parallelism); err != nil {
		return Parameters{}, nil, nil, ErrMalformedHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(hashParts[4])
	if err != nil {
		return Parameters{}, nil, nil, ErrMalformedHash
	}
	derivedKey, err := base64.RawStdEncoding.DecodeString(hashParts[5])
	if err != nil {
		return Parameters{}, nil, nil, ErrMalformedHash
	}
	return parameters, salt, derivedKey, nil
}
