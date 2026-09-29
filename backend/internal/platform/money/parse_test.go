package money

import "testing"

func TestParseNC(t *testing.T) {
	validAmounts := map[string]Micro{"25": 25_000_000, "0.02": 20_000, "1.234567": 1_234_567, " 3.5 ": 3_500_000}
	for amountText, expected := range validAmounts {
		if parsed, err := ParseNC(amountText); err != nil || parsed != expected {
			t.Fatalf("ParseNC(%q) = %d, %v", amountText, parsed, err)
		}
	}
	for _, invalidText := range []string{"", ".5", "1.2345678", "-1", "+1", "1e3", "abc", "1.2.3"} {
		if _, err := ParseNC(invalidText); err == nil {
			t.Fatalf("ParseNC(%q) should fail", invalidText)
		}
	}
}
