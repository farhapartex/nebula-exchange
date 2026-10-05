package emails

import (
	"strings"
	"testing"
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email"
)

func TestActivationEmailContainsTheLinkAndEscapesTheUsername(t *testing.T) {
	composer, err := NewActivationEmailComposer("http://localhost:3000/")
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}

	message, err := composer.Compose(email.Address{Name: "<b>boy</b>", Email: "boy@streetborn.test"}, "token-123", time.Date(2026, 10, 6, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("compose: %v", err)
	}

	expectedLink := "http://localhost:3000/activate?token=token-123"
	if !strings.Contains(message.TextBody, expectedLink) || !strings.Contains(message.HTMLBody, expectedLink) {
		t.Fatalf("expected both bodies to contain %s", expectedLink)
	}
	if strings.Contains(message.HTMLBody, "<b>boy</b>") {
		t.Fatal("the html body must escape the username")
	}
	if message.To.Email != "boy@streetborn.test" || message.Subject == "" {
		t.Fatalf("unexpected message header %+v", message)
	}
}
