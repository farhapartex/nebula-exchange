package seeding

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

const maximumSVGBytes = 32 * 1024

var forbiddenSVGElements = map[string]bool{
	"script":        true,
	"style":         true,
	"foreignobject": true,
	"iframe":        true,
	"object":        true,
	"embed":         true,
	"image":         true,
	"audio":         true,
	"video":         true,
}

func validateSVGImage(svgMarkup string) error {
	if len(svgMarkup) == 0 || len(svgMarkup) > maximumSVGBytes {
		return fmt.Errorf("the SVG must be between 1 and %d bytes", maximumSVGBytes)
	}
	decoder := xml.NewDecoder(strings.NewReader(svgMarkup))
	decoder.Strict = true
	hasRootElement := false
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("the SVG is not valid XML: %w", err)
		}
		switch typedToken := token.(type) {
		case xml.Directive:
			return errors.New("the SVG must not contain a DOCTYPE or other directives")
		case xml.StartElement:
			elementName := strings.ToLower(typedToken.Name.Local)
			if !hasRootElement && elementName != "svg" {
				return errors.New("the SVG must start with an svg element")
			}
			hasRootElement = true
			if forbiddenSVGElements[elementName] {
				return fmt.Errorf("the SVG must not contain a %s element", elementName)
			}
			if err := validateSVGAttributes(elementName, typedToken.Attr); err != nil {
				return err
			}
		}
	}
	if !hasRootElement {
		return errors.New("the SVG has no svg element")
	}
	return nil
}

func validateSVGAttributes(elementName string, attributes []xml.Attr) error {
	for _, attribute := range attributes {
		attributeName := strings.ToLower(attribute.Name.Local)
		attributeValue := strings.ToLower(strings.TrimSpace(attribute.Value))
		if strings.HasPrefix(attributeName, "on") {
			return fmt.Errorf("the SVG must not use the %s event attribute on %s", attributeName, elementName)
		}
		if attributeName == "href" && !strings.HasPrefix(attributeValue, "#") {
			return fmt.Errorf("the SVG may only link to its own elements, not %q", attribute.Value)
		}
		if strings.Contains(attributeValue, "javascript:") || strings.Contains(attributeValue, "@import") {
			return errors.New("the SVG must not contain scripts or imports")
		}
		if strings.Contains(attributeValue, "url(") && !strings.Contains(attributeValue, "url(#") {
			return errors.New("the SVG may only use url() with its own elements")
		}
	}
	return nil
}
