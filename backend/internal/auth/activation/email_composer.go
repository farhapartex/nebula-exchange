package activation

import (
	"net/url"
	"strings"
	"time"

	"nebula-exchange/backend/internal/notify/email"
)

const activationPagePath = "/activate"

type EmailComposer struct {
	frontendBaseURL string
	renderer        *email.TemplateRenderer
}

func NewEmailComposer(frontendBaseURL string, renderer *email.TemplateRenderer) *EmailComposer {
	return &EmailComposer{frontendBaseURL: strings.TrimRight(frontendBaseURL, "/"), renderer: renderer}
}

func (composer *EmailComposer) ActivationLink(plaintextToken string) string {
	return composer.frontendBaseURL + activationPagePath + "?token=" + url.QueryEscape(plaintextToken)
}

func (composer *EmailComposer) Compose(recipient email.Address, username string, issuedToken IssuedToken) (email.Message, error) {
	htmlBody, textBody, err := composer.renderer.Render(email.TemplateAccountActivation, struct {
		Username       string
		ActivationLink string
		ExpiresAt      string
	}{
		Username:       username,
		ActivationLink: composer.ActivationLink(issuedToken.Plaintext),
		ExpiresAt:      issuedToken.ExpiresAt.Format(time.RFC1123),
	})
	if err != nil {
		return email.Message{}, err
	}
	return email.Message{
		To:       recipient,
		Subject:  "Activate your Nebula Exchange account",
		HTMLBody: htmlBody,
		TextBody: textBody,
	}, nil
}
