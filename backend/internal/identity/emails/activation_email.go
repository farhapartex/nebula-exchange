package emails

import (
	"embed"
	"io/fs"
	"net/url"
	"strings"
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email"
)

const (
	ActivationTemplateName = "account_activation"
	activationPagePath     = "/activate"
	activationSubject      = "Activate your Street Born account"
)

//go:embed templates/*.html templates/*.txt
var templateFiles embed.FS

type ActivationEmailComposer struct {
	frontendBaseURL string
	renderer        *email.TemplateRenderer
}

func NewActivationEmailComposer(frontendBaseURL string) (*ActivationEmailComposer, error) {
	templateDirectory, err := fs.Sub(templateFiles, "templates")
	if err != nil {
		return nil, err
	}
	renderer, err := email.NewTemplateRenderer(templateDirectory)
	if err != nil {
		return nil, err
	}
	return &ActivationEmailComposer{frontendBaseURL: strings.TrimRight(frontendBaseURL, "/"), renderer: renderer}, nil
}

func (composer *ActivationEmailComposer) ActivationLink(plaintextToken string) string {
	return composer.frontendBaseURL + activationPagePath + "?token=" + url.QueryEscape(plaintextToken)
}

func (composer *ActivationEmailComposer) Compose(recipient email.Address, plaintextToken string, expiresAt time.Time) (email.Message, error) {
	htmlBody, textBody, err := composer.renderer.Render(ActivationTemplateName, struct {
		Username       string
		ActivationLink string
		ExpiresAt      string
	}{
		Username:       recipient.Name,
		ActivationLink: composer.ActivationLink(plaintextToken),
		ExpiresAt:      expiresAt.UTC().Format(time.RFC1123),
	})
	if err != nil {
		return email.Message{}, err
	}
	return email.Message{To: recipient, Subject: activationSubject, HTMLBody: htmlBody, TextBody: textBody}, nil
}
