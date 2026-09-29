package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

var ErrInvalidCursor = errors.New("cursor is invalid")

func EncodeCursor(position any) (string, error) {
	encodedPosition, err := json.Marshal(position)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encodedPosition), nil
}

func DecodeCursor[Position any](cursor string) (Position, error) {
	var position Position
	decodedBytes, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return position, ErrInvalidCursor
	}
	if err := json.Unmarshal(decodedBytes, &position); err != nil {
		return position, ErrInvalidCursor
	}
	return position, nil
}
