package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const signInRequestSuffix = " wants you to sign in with your Ethereum account:"

const (
	uriField            = "URI"
	versionField        = "Version"
	chainIDField        = "Chain ID"
	nonceField          = "Nonce"
	issuedAtField       = "Issued At"
	expirationTimeField = "Expiration Time"
	notBeforeField      = "Not Before"
)

var errMalformedSignInMessage = errors.New("malformed sign-in message")

type SignInMessage struct {
	Domain         string
	Address        string
	URI            string
	Version        string
	ChainID        int64
	Nonce          string
	IssuedAt       time.Time
	ExpirationTime *time.Time
	NotBefore      *time.Time
}

func ParseSignInMessage(messageText string) (SignInMessage, error) {
	lines := strings.Split(messageText, "\n")
	if len(lines) < 3 || !strings.HasSuffix(lines[0], signInRequestSuffix) {
		return SignInMessage{}, errMalformedSignInMessage
	}
	message := SignInMessage{
		Domain:  strings.TrimSuffix(lines[0], signInRequestSuffix),
		Address: lines[1],
	}
	if message.Domain == "" || !isWalletAddress(message.Address) {
		return SignInMessage{}, errMalformedSignInMessage
	}

	fields := map[string]string{}
	for _, line := range lines[2:] {
		fieldName, fieldValue, hasSeparator := strings.Cut(line, ": ")
		if hasSeparator && isKnownField(fieldName) {
			if _, isDuplicate := fields[fieldName]; isDuplicate {
				return SignInMessage{}, errMalformedSignInMessage
			}
			fields[fieldName] = fieldValue
		}
	}
	for _, requiredField := range []string{uriField, versionField, chainIDField, nonceField, issuedAtField} {
		if fields[requiredField] == "" {
			return SignInMessage{}, fmt.Errorf("%w: missing %s", errMalformedSignInMessage, requiredField)
		}
	}

	chainID, err := strconv.ParseInt(fields[chainIDField], 10, 64)
	if err != nil {
		return SignInMessage{}, errMalformedSignInMessage
	}
	issuedAt, err := time.Parse(time.RFC3339Nano, fields[issuedAtField])
	if err != nil {
		return SignInMessage{}, errMalformedSignInMessage
	}
	message.URI = fields[uriField]
	message.Version = fields[versionField]
	message.ChainID = chainID
	message.Nonce = fields[nonceField]
	message.IssuedAt = issuedAt
	if message.ExpirationTime, err = parseOptionalTime(fields[expirationTimeField]); err != nil {
		return SignInMessage{}, err
	}
	if message.NotBefore, err = parseOptionalTime(fields[notBeforeField]); err != nil {
		return SignInMessage{}, err
	}
	return message, nil
}

func isKnownField(fieldName string) bool {
	switch fieldName {
	case uriField, versionField, chainIDField, nonceField, issuedAtField, expirationTimeField, notBeforeField:
		return true
	}
	return false
}

func parseOptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsedTime, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, errMalformedSignInMessage
	}
	return &parsedTime, nil
}

func isWalletAddress(address string) bool {
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return false
	}
	for _, character := range address[2:] {
		isHexDigit := (character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')
		if !isHexDigit {
			return false
		}
	}
	return true
}
