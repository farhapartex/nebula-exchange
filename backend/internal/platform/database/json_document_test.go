package database

import (
	"encoding/json"
	"testing"
)

func TestJSONDocumentWritesTextAndReadsBothDriverForms(t *testing.T) {
	document := JSONDocument(`{"accent":"#f97316"}`)
	storedValue, err := document.Value()
	if err != nil || storedValue != `{"accent":"#f97316"}` {
		t.Fatalf("got %v, %v", storedValue, err)
	}

	var fromBytes, fromText JSONDocument
	if err := fromBytes.Scan([]byte(`{"a":1}`)); err != nil || string(fromBytes) != `{"a":1}` {
		t.Fatalf("scan bytes: %s, %v", fromBytes, err)
	}
	if err := fromText.Scan(`[1,2]`); err != nil || string(fromText) != `[1,2]` {
		t.Fatalf("scan text: %s, %v", fromText, err)
	}
}

func TestJSONDocumentRejectsInvalidJSONAndEmbedsAsRawJSON(t *testing.T) {
	if _, err := JSONDocument(`{broken`).Value(); err == nil {
		t.Fatal("expected invalid JSON to be rejected")
	}
	encoded, err := json.Marshal(struct {
		Palette JSONDocument `json:"palette"`
	}{Palette: JSONDocument(`{"glow":"#fdba74"}`)})
	if err != nil || string(encoded) != `{"palette":{"glow":"#fdba74"}}` {
		t.Fatalf("got %s, %v", encoded, err)
	}
}
