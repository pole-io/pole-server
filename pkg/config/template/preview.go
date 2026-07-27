package template

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/magiconair/properties"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v3"
)

const (
	EnginePoleMustache = "pole-mustache"
	EngineVersionV1    = "v1"
)

type PreviewRequest struct {
	RenderRequest
	Format string
}

type PreviewResult struct {
	Valid           bool
	RenderedContent string
	RenderedSHA256  string
	Format          string
	Engine          string
	EngineVersion   string
	Diagnostics     []Diagnostic
}

func Preview(ctx context.Context, req PreviewRequest) (PreviewResult, error) {
	if err := ctx.Err(); err != nil {
		return PreviewResult{}, err
	}

	result := PreviewResult{
		Format:        req.Format,
		Engine:        EnginePoleMustache,
		EngineVersion: EngineVersionV1,
	}
	rendered, err := Render(req.RenderRequest)
	if err != nil {
		renderErr, ok := err.(*RenderError)
		if !ok {
			return result, err
		}
		result.Diagnostics = renderErr.Diagnostics
		return result, nil
	}
	if err := ctx.Err(); err != nil {
		return PreviewResult{}, err
	}

	sum := sha256.Sum256([]byte(rendered.Content))
	result.RenderedContent = rendered.Content
	result.RenderedSHA256 = hex.EncodeToString(sum[:])
	if diagnostic := validateFormat(req.Format, rendered.Content); diagnostic != nil {
		result.Diagnostics = []Diagnostic{*diagnostic}
		return result, nil
	}
	result.Valid = true
	return result, nil
}

func validateFormat(format, content string) *Diagnostic {
	var err error
	switch strings.ToLower(format) {
	case "text":
		return nil
	case "json":
		if !json.Valid([]byte(content)) {
			err = fmt.Errorf("invalid JSON")
		}
	case "yaml", "yml":
		var document any
		err = yaml.Unmarshal([]byte(content), &document)
	case "xml":
		err = validateXML(content)
	case "properties":
		_, err = properties.Load([]byte(content), properties.UTF8)
	case "html":
		_, err = html.Parse(strings.NewReader(content))
	default:
		return &Diagnostic{
			Code:    DiagnosticUnsupportedFormat,
			Message: fmt.Sprintf("unsupported config format %q", format),
		}
	}
	if err == nil {
		return nil
	}
	return &Diagnostic{
		Code:    DiagnosticInvalidFormat,
		Message: fmt.Sprintf("rendered content is not valid %s: %v", format, err),
	}
}

func validateXML(content string) error {
	decoder := xml.NewDecoder(bytes.NewBufferString(content))
	depth := 0
	rootElements := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				rootElements++
				if rootElements > 1 {
					return fmt.Errorf("multiple root elements")
				}
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(typed)) != "" {
				return fmt.Errorf("text outside root element")
			}
		}
	}
	if rootElements != 1 || depth != 0 {
		return fmt.Errorf("expected exactly one root element")
	}
	return nil
}
