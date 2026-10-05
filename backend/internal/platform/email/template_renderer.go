package email

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"io/fs"
	texttemplate "text/template"
)

type TemplateRenderer struct {
	htmlTemplates *htmltemplate.Template
	textTemplates *texttemplate.Template
}

func NewTemplateRenderer(templateFiles fs.FS) (*TemplateRenderer, error) {
	htmlTemplates, err := htmltemplate.ParseFS(templateFiles, "*.html")
	if err != nil {
		return nil, fmt.Errorf("parse html email templates: %w", err)
	}
	textTemplates, err := texttemplate.ParseFS(templateFiles, "*.txt")
	if err != nil {
		return nil, fmt.Errorf("parse text email templates: %w", err)
	}
	return &TemplateRenderer{htmlTemplates: htmlTemplates, textTemplates: textTemplates}, nil
}

func (renderer *TemplateRenderer) Render(templateName string, templateData any) (htmlBody, textBody string, err error) {
	var htmlBuffer, textBuffer bytes.Buffer
	if err := renderer.htmlTemplates.ExecuteTemplate(&htmlBuffer, templateName+".html", templateData); err != nil {
		return "", "", fmt.Errorf("render html template %s: %w", templateName, err)
	}
	if err := renderer.textTemplates.ExecuteTemplate(&textBuffer, templateName+".txt", templateData); err != nil {
		return "", "", fmt.Errorf("render text template %s: %w", templateName, err)
	}
	return htmlBuffer.String(), textBuffer.String(), nil
}
