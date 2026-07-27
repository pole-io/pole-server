package template_test

import (
	"context"
	"errors"
	"testing"

	configtemplate "github.com/pole-io/pole-server/pkg/config/template"
)

func TestPreviewReturnsReferenceContentAndSHA256(t *testing.T) {
	t.Parallel()

	result, err := configtemplate.Preview(context.Background(), configtemplate.PreviewRequest{
		RenderRequest: configtemplate.RenderRequest{
			Template: `{"host":"{{{database.host}}}","port":{{{database.port}}}}`,
			Schema: []configtemplate.Parameter{
				{Name: "database.host", Type: configtemplate.ScalarString, Required: true},
				{Name: "database.port", Type: configtemplate.ScalarInteger, Required: true},
			},
			Values: map[string]configtemplate.Scalar{
				"database.host": {Type: configtemplate.ScalarString, Value: "db"},
				"database.port": {Type: configtemplate.ScalarInteger, Value: "3306"},
			},
		},
		Format: "json",
	})
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	if !result.Valid {
		t.Fatalf("Preview() valid = false, diagnostics = %#v", result.Diagnostics)
	}
	if result.RenderedContent != `{"host":"db","port":3306}` {
		t.Fatalf("Preview() content = %q", result.RenderedContent)
	}
	if result.RenderedSHA256 != "0263db63c6fe6c7b7c10efaecd0f6863655bb8e1a837bade2dc8065b79c0ace5" {
		t.Fatalf("Preview() sha256 = %q", result.RenderedSHA256)
	}
	if result.Engine != configtemplate.EnginePoleMustache || result.EngineVersion != configtemplate.EngineVersionV1 {
		t.Fatalf("Preview() engine = %q/%q", result.Engine, result.EngineVersion)
	}
}

func TestPreviewReportsRenderAndTargetFormatDiagnostics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  configtemplate.PreviewRequest
		code configtemplate.DiagnosticCode
	}{
		{
			name: "render violation",
			req: configtemplate.PreviewRequest{
				RenderRequest: configtemplate.RenderRequest{Template: "{{forbidden}}"},
				Format:        "text",
			},
			code: configtemplate.DiagnosticInvalidSyntax,
		},
		{
			name: "rendered json is invalid",
			req: configtemplate.PreviewRequest{
				RenderRequest: configtemplate.RenderRequest{Template: `{"missing": }`},
				Format:        "json",
			},
			code: configtemplate.DiagnosticInvalidFormat,
		},
		{
			name: "rendered yaml is invalid",
			req: configtemplate.PreviewRequest{
				RenderRequest: configtemplate.RenderRequest{Template: "key: [unterminated"},
				Format:        "yaml",
			},
			code: configtemplate.DiagnosticInvalidFormat,
		},
		{
			name: "unknown format is rejected",
			req: configtemplate.PreviewRequest{
				RenderRequest: configtemplate.RenderRequest{Template: "plain"},
				Format:        "hocon",
			},
			code: configtemplate.DiagnosticUnsupportedFormat,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := configtemplate.Preview(context.Background(), tt.req)
			if err != nil {
				t.Fatalf("Preview() error = %v", err)
			}
			if result.Valid {
				t.Fatal("Preview() valid = true, want false")
			}
			if !hasDiagnostic(result.Diagnostics, tt.code) {
				t.Fatalf("Preview() diagnostics = %#v, want code %q", result.Diagnostics, tt.code)
			}
			if tt.code == configtemplate.DiagnosticInvalidFormat && result.RenderedContent == "" {
				t.Fatal("Preview() content is empty, want rendered content for format troubleshooting")
			}
			if tt.code == configtemplate.DiagnosticInvalidFormat && result.RenderedSHA256 == "" {
				t.Fatal("Preview() sha256 is empty after rendering succeeded")
			}
		})
	}
}

func TestPreviewValidatesEveryExistingConfigFileFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		format  string
		content string
	}{
		{format: "text", content: "anything\n"},
		{format: "json", content: `{"valid":true}`},
		{format: "yaml", content: "valid: true\n"},
		{format: "yml", content: "valid: true\n"},
		{format: "xml", content: `<root><valid>true</valid></root>`},
		{format: "properties", content: "valid=true\n"},
		{format: "html", content: "<!doctype html><title>valid</title>"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.format, func(t *testing.T) {
			t.Parallel()

			result, err := configtemplate.Preview(context.Background(), configtemplate.PreviewRequest{
				RenderRequest: configtemplate.RenderRequest{Template: tt.content},
				Format:        tt.format,
			})
			if err != nil {
				t.Fatalf("Preview() error = %v", err)
			}
			if !result.Valid {
				t.Fatalf("Preview() diagnostics = %#v", result.Diagnostics)
			}
		})
	}
}

func TestPreviewHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := configtemplate.Preview(ctx, configtemplate.PreviewRequest{
		RenderRequest: configtemplate.RenderRequest{Template: "plain"},
		Format:        "text",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Preview() error = %v, want context.Canceled", err)
	}
}
