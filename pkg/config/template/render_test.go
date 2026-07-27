package template_test

import (
	"errors"
	"testing"

	configtemplate "github.com/pole-io/pole-server/pkg/config/template"
)

func TestRenderReplacesTypedDottedScalarsWithoutChangingWhitespace(t *testing.T) {
	t.Parallel()

	result, err := configtemplate.Render(configtemplate.RenderRequest{
		Template: "server:\r\n  host: {{{database.host}}}\r\n  port: {{{database.port}}}\r\n  enabled: {{{database.enabled}}}\r\n  ratio: {{{traffic.ratio}}}\r\n",
		Schema: []configtemplate.Parameter{
			{Name: "database.host", Type: configtemplate.ScalarString, Required: true},
			{Name: "database.port", Type: configtemplate.ScalarInteger, Required: true},
			{Name: "database.enabled", Type: configtemplate.ScalarBoolean, Required: true},
			{Name: "traffic.ratio", Type: configtemplate.ScalarDecimal, Required: true},
		},
		Values: map[string]configtemplate.Scalar{
			"database.host":    {Type: configtemplate.ScalarString, Value: "数据库.internal"},
			"database.port":    {Type: configtemplate.ScalarInteger, Value: "3306"},
			"database.enabled": {Type: configtemplate.ScalarBoolean, Value: "true"},
			"traffic.ratio":    {Type: configtemplate.ScalarDecimal, Value: "1.23"},
		},
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	const want = "server:\r\n  host: 数据库.internal\r\n  port: 3306\r\n  enabled: true\r\n  ratio: 1.23\r\n"
	if result.Content != want {
		t.Fatalf("Render() content = %q, want %q", result.Content, want)
	}
}

func TestRenderRejectsSchemaValueAndSyntaxViolations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  configtemplate.RenderRequest
		code configtemplate.DiagnosticCode
	}{
		{
			name: "noncanonical integer is forbidden",
			req: configtemplate.RenderRequest{
				Template: "{{{port}}}",
				Schema:   []configtemplate.Parameter{{Name: "port", Type: configtemplate.ScalarInteger}},
				Values:   map[string]configtemplate.Scalar{"port": {Type: configtemplate.ScalarInteger, Value: "08080"}},
			},
			code: configtemplate.DiagnosticInvalidValue,
		},
		{
			name: "noncanonical decimal is forbidden",
			req: configtemplate.RenderRequest{
				Template: "{{{ratio}}}",
				Schema:   []configtemplate.Parameter{{Name: "ratio", Type: configtemplate.ScalarDecimal}},
				Values:   map[string]configtemplate.Scalar{"ratio": {Type: configtemplate.ScalarDecimal, Value: "1.250"}},
			},
			code: configtemplate.DiagnosticInvalidValue,
		},
		{
			name: "negative decimal zero is forbidden",
			req: configtemplate.RenderRequest{
				Template: "{{{ratio}}}",
				Schema:   []configtemplate.Parameter{{Name: "ratio", Type: configtemplate.ScalarDecimal}},
				Values:   map[string]configtemplate.Scalar{"ratio": {Type: configtemplate.ScalarDecimal, Value: "-0"}},
			},
			code: configtemplate.DiagnosticInvalidValue,
		},
		{
			name: "required value is missing even when parameter is unused",
			req: configtemplate.RenderRequest{
				Template: "plain",
				Schema:   []configtemplate.Parameter{{Name: "required.value", Type: configtemplate.ScalarString, Required: true}},
			},
			code: configtemplate.DiagnosticMissingValue,
		},
		{
			name: "value is not declared by schema",
			req: configtemplate.RenderRequest{
				Template: "plain",
				Values:   map[string]configtemplate.Scalar{"unknown": {Type: configtemplate.ScalarString, Value: "value"}},
			},
			code: configtemplate.DiagnosticUnknownValue,
		},
		{
			name: "placeholder is not declared by schema",
			req:  configtemplate.RenderRequest{Template: "{{{unknown.value}}}"},
			code: configtemplate.DiagnosticUnknownParameter,
		},
		{
			name: "value type differs from schema",
			req: configtemplate.RenderRequest{
				Template: "{{{port}}}",
				Schema:   []configtemplate.Parameter{{Name: "port", Type: configtemplate.ScalarInteger}},
				Values:   map[string]configtemplate.Scalar{"port": {Type: configtemplate.ScalarString, Value: "8080"}},
			},
			code: configtemplate.DiagnosticTypeMismatch,
		},
		{
			name: "double mustache is forbidden",
			req:  configtemplate.RenderRequest{Template: "{{name}}"},
			code: configtemplate.DiagnosticInvalidSyntax,
		},
		{
			name: "section is forbidden",
			req:  configtemplate.RenderRequest{Template: "{{{#items}}}"},
			code: configtemplate.DiagnosticInvalidSyntax,
		},
		{
			name: "whitespace in placeholder is forbidden",
			req:  configtemplate.RenderRequest{Template: "{{{ database.host }}}"},
			code: configtemplate.DiagnosticInvalidSyntax,
		},
		{
			name: "unclosed placeholder is forbidden",
			req:  configtemplate.RenderRequest{Template: "{{{database.host}}"},
			code: configtemplate.DiagnosticInvalidSyntax,
		},
		{
			name: "duplicate schema parameter is forbidden",
			req: configtemplate.RenderRequest{
				Template: "plain",
				Schema: []configtemplate.Parameter{
					{Name: "database.host", Type: configtemplate.ScalarString},
					{Name: "database.host", Type: configtemplate.ScalarString},
				},
			},
			code: configtemplate.DiagnosticDuplicateParameter,
		},
		{
			name: "unsupported schema scalar type is forbidden even when unused",
			req: configtemplate.RenderRequest{
				Template: "plain",
				Schema:   []configtemplate.Parameter{{Name: "database.options", Type: "object"}},
			},
			code: configtemplate.DiagnosticInvalidValue,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := configtemplate.Render(tt.req)
			var renderErr *configtemplate.RenderError
			if !errors.As(err, &renderErr) {
				t.Fatalf("Render() error = %T %v, want *RenderError", err, err)
			}
			if !hasDiagnostic(renderErr.Diagnostics, tt.code) {
				t.Fatalf("Render() diagnostics = %#v, want code %q", renderErr.Diagnostics, tt.code)
			}
		})
	}
}

func TestRenderUsesTypedDefaultAndSupportsArbitraryPrecisionNumbers(t *testing.T) {
	t.Parallel()

	result, err := configtemplate.Render(configtemplate.RenderRequest{
		Template: "{{{huge.integer}}}\n{{{tiny.decimal}}}\n{{{negative.decimal}}}",
		Schema: []configtemplate.Parameter{
			{
				Name:     "huge.integer",
				Type:     configtemplate.ScalarInteger,
				Default:  &configtemplate.Scalar{Type: configtemplate.ScalarInteger, Value: "999999999999999999999999999999"},
				Required: false,
			},
			{Name: "tiny.decimal", Type: configtemplate.ScalarDecimal, Required: true},
			{Name: "negative.decimal", Type: configtemplate.ScalarDecimal, Required: true},
		},
		Values: map[string]configtemplate.Scalar{
			"tiny.decimal":     {Type: configtemplate.ScalarDecimal, Value: "0.00000000000000000000123"},
			"negative.decimal": {Type: configtemplate.ScalarDecimal, Value: "-100.25"},
		},
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	const want = "999999999999999999999999999999\n0.00000000000000000000123\n-100.25"
	if result.Content != want {
		t.Fatalf("Render() content = %q, want %q", result.Content, want)
	}
}

func TestRenderDoesNotRecursivelyInterpretStringValues(t *testing.T) {
	t.Parallel()

	result, err := configtemplate.Render(configtemplate.RenderRequest{
		Template: "{{{literal}}}",
		Schema:   []configtemplate.Parameter{{Name: "literal", Type: configtemplate.ScalarString, Required: true}},
		Values: map[string]configtemplate.Scalar{
			"literal": {Type: configtemplate.ScalarString, Value: "{{{not.a.placeholder}}}"},
		},
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if result.Content != "{{{not.a.placeholder}}}" {
		t.Fatalf("Render() content = %q", result.Content)
	}
}

func hasDiagnostic(diagnostics []configtemplate.Diagnostic, code configtemplate.DiagnosticCode) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}
