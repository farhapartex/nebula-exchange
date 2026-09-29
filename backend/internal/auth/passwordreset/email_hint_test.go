package passwordreset

import "testing"

func TestMaskEmailKeepsOnlyTheEdgesOfTheLocalPart(t *testing.T) {
	expectedHints := map[string]string{
		"pilot@nebula.test": "p•••t@nebula.test",
		"ab@nebula.test":    "a•••@nebula.test",
		"a@nebula.test":     "a•••@nebula.test",
		"broken":            "•••",
	}
	for emailAddress, expectedHint := range expectedHints {
		if actualHint := MaskEmail(emailAddress); actualHint != expectedHint {
			t.Fatalf("%s: got %s, want %s", emailAddress, actualHint, expectedHint)
		}
	}
}
