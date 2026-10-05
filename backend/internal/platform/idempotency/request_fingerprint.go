package idempotency

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"net/http"
	"regexp"
)

var (
	errRequestBodyTooLarge = errors.New("request body is too large")
	acceptedKeyPattern     = regexp.MustCompile(`^[A-Za-z0-9_\-:.]{8,255}$`)
)

func isStateChangingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func isAcceptedKey(key string) bool {
	return acceptedKeyPattern.MatchString(key)
}

func readAndRestoreBody(request *http.Request, maximumBodyBytes int64) ([]byte, error) {
	if request.Body == nil {
		return nil, nil
	}
	requestBody, err := io.ReadAll(io.LimitReader(request.Body, maximumBodyBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(requestBody)) > maximumBodyBytes {
		return nil, errRequestBodyTooLarge
	}
	request.Body = io.NopCloser(bytes.NewReader(requestBody))
	return requestBody, nil
}

func fingerprintRequest(request *http.Request, requestBody []byte) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(request.Method))
	hasher.Write([]byte{0})
	hasher.Write([]byte(request.URL.RequestURI()))
	hasher.Write([]byte{0})
	hasher.Write(requestBody)
	return hasher.Sum(nil)
}
