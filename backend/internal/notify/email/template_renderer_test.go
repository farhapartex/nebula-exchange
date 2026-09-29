package email

import (
	"strings"
	"testing"
)

func TestActivationTemplateRendersAndEscapes(t *testing.T) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}

	htmlBody, textBody, err := renderer.Render(TemplateAccountActivation, struct {
		Username       string
		ActivationLink string
		ExpiresAt      string
	}{
		Username:       "<script>pilot</script>",
		ActivationLink: "http://localhost:3000/activate?token=abc",
		ExpiresAt:      "Thu, 01 Oct 2026 10:00:00 UTC",
	})
	if err != nil {
		t.Fatalf("render template: %v", err)
	}

	if strings.Contains(htmlBody, "<script>pilot</script>") {
		t.Fatal("html body must escape the username")
	}
	for _, expectedText := range []string{"http://localhost:3000/activate?token=abc", "Thu, 01 Oct 2026 10:00:00 UTC"} {
		if !strings.Contains(htmlBody, expectedText) || !strings.Contains(textBody, expectedText) {
			t.Fatalf("both bodies must contain %q", expectedText)
		}
	}
}

func TestMIMEMessageContainsBothParts(t *testing.T) {
	encodedMessage, err := buildMIMEMessage(
		Address{Name: "Nebula Exchange", Email: "no-reply@nebula.test"},
		Message{To: Address{Email: "pilot@nebula.test"}, Subject: "Activate", HTMLBody: "<p>hi</p>", TextBody: "hi"},
	)
	if err != nil {
		t.Fatalf("build message: %v", err)
	}
	for _, expectedPart := range []string{"To: <pilot@nebula.test>", "multipart/alternative", "text/plain", "text/html", "@nebula.test>"} {
		if !strings.Contains(string(encodedMessage), expectedPart) {
			t.Fatalf("message is missing %q:\n%s", expectedPart, encodedMessage)
		}
	}
}
