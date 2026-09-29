package money

import (
	"encoding/json"
	"testing"
)

func TestMicroTravelsAsAString(t *testing.T) {
	encodedAmount, _ := json.Marshal(struct {
		Price Micro `json:"price"`
	}{Price: 9_223_372_036_854_775_807})
	if string(encodedAmount) != `{"price":"9223372036854775807"}` {
		t.Fatalf("got %s", encodedAmount)
	}

	var decoded struct {
		Price Micro `json:"price"`
	}
	if err := json.Unmarshal([]byte(`{"price":"1284500000"}`), &decoded); err != nil || decoded.Price != 1_284_500_000 {
		t.Fatalf("got %v, %v", decoded.Price, err)
	}
	if err := json.Unmarshal([]byte(`{"price":12.5}`), &decoded); err == nil {
		t.Fatal("numbers must be rejected so no float ever reaches the backend")
	}
	if err := json.Unmarshal([]byte(`{"price":"1.5"}`), &decoded); err == nil {
		t.Fatal("fractional micro-units must be rejected")
	}
}
