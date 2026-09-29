package email

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
)

//go:embed templates/*.html templates/*.txt
var templateFiles embed.FS

type TemplateName string

const (
	TemplateAccountActivation TemplateName = "account_activation"
	TemplatePasswordReset     TemplateName = "password_reset"
	TemplateTwoFactorChanged  TemplateName = "two_factor_changed"
)

type TemplateRenderer struct {
	htmlTemplates *htmltemplate.Template
	textTemplates *texttemplate.Template
}

func NewTemplateRenderer() (*TemplateRenderer, error) {
	htmlTemplates, err := htmltemplate.ParseFS(templateFiles, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse html email templates: %w", err)
	}
	textTemplates, err := texttemplate.ParseFS(templateFiles, "templates/*.txt")
	if err != nil {
		return nil, fmt.Errorf("parse text email templates: %w", err)
	}
	return &TemplateRenderer{htmlTemplates: htmlTemplates, textTemplates: textTemplates}, nil
}

func (renderer *TemplateRenderer) Render(templateName TemplateName, templateData any) (htmlBody, textBody string, err error) {
	var htmlBuffer, textBuffer bytes.Buffer
	if err := renderer.htmlTemplates.ExecuteTemplate(&htmlBuffer, string(templateName)+".html", templateData); err != nil {
		return "", "", fmt.Errorf("render html template %s: %w", templateName, err)
	}
	if err := renderer.textTemplates.ExecuteTemplate(&textBuffer, string(templateName)+".txt", templateData); err != nil {
		return "", "", fmt.Errorf("render text template %s: %w", templateName, err)
	}
	return htmlBuffer.String(), textBuffer.String(), nil
}
