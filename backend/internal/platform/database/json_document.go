package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

var errInvalidJSONDocument = errors.New("value is not valid JSON")

type JSONDocument json.RawMessage

func (document JSONDocument) Value() (driver.Value, error) {
	if len(document) == 0 {
		return nil, nil
	}
	if !json.Valid(document) {
		return nil, errInvalidJSONDocument
	}
	return string(document), nil
}

func (document *JSONDocument) Scan(source any) error {
	switch value := source.(type) {
	case nil:
		*document = nil
	case []byte:
		*document = append(JSONDocument(nil), value...)
	case string:
		*document = JSONDocument(value)
	default:
		return fmt.Errorf("scan json document from %T", source)
	}
	return nil
}

func (document JSONDocument) MarshalJSON() ([]byte, error) {
	if len(document) == 0 {
		return []byte("null"), nil
	}
	return document, nil
}

func (document *JSONDocument) UnmarshalJSON(data []byte) error {
	*document = append(JSONDocument(nil), data...)
	return nil
}
