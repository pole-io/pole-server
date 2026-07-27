package template

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

type ScalarType string

const (
	ScalarString  ScalarType = "string"
	ScalarBoolean ScalarType = "boolean"
	ScalarInteger ScalarType = "integer"
	ScalarDecimal ScalarType = "decimal"
)

type Scalar struct {
	Type  ScalarType
	Value string
}

type Parameter struct {
	Name     string
	Type     ScalarType
	Required bool
	Default  *Scalar
}

type RenderRequest struct {
	Template string
	Schema   []Parameter
	Values   map[string]Scalar
}

type RenderResult struct {
	Content string
}

type DiagnosticCode string

const (
	DiagnosticInvalidSyntax      DiagnosticCode = "INVALID_SYNTAX"
	DiagnosticDuplicateParameter DiagnosticCode = "DUPLICATE_PARAMETER"
	DiagnosticUnknownParameter   DiagnosticCode = "UNKNOWN_PARAMETER"
	DiagnosticUnknownValue       DiagnosticCode = "UNKNOWN_VALUE"
	DiagnosticMissingValue       DiagnosticCode = "MISSING_VALUE"
	DiagnosticTypeMismatch       DiagnosticCode = "TYPE_MISMATCH"
	DiagnosticInvalidValue       DiagnosticCode = "INVALID_VALUE"
	DiagnosticInvalidFormat      DiagnosticCode = "INVALID_FORMAT"
	DiagnosticUnsupportedFormat  DiagnosticCode = "UNSUPPORTED_FORMAT"
)

type Diagnostic struct {
	Code      DiagnosticCode
	Parameter string
	Offset    int
	Message   string
}

type RenderError struct {
	Diagnostics []Diagnostic
}

func (e *RenderError) Error() string {
	if len(e.Diagnostics) == 0 {
		return "template rendering failed"
	}
	return e.Diagnostics[0].Message
}

type segment struct {
	literal   string
	parameter string
	offset    int
}

var parameterNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*(?:\.[A-Za-z_][A-Za-z0-9_-]*)*$`)
var canonicalIntegerPattern = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)$`)
var canonicalDecimalPattern = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)(?:\.[0-9]*[1-9])?$`)

func Render(req RenderRequest) (RenderResult, error) {
	parameters, diagnostics := validateInputs(req)
	segments, syntaxDiagnostics := parse(req.Template)
	diagnostics = append(diagnostics, syntaxDiagnostics...)

	var content strings.Builder
	for _, item := range segments {
		if item.parameter == "" {
			content.WriteString(item.literal)
			continue
		}

		parameter, ok := parameters[item.parameter]
		if !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Code:      DiagnosticUnknownParameter,
				Parameter: item.parameter,
				Offset:    item.offset,
				Message:   fmt.Sprintf("unknown parameter %q", item.parameter),
			})
			continue
		}
		value, ok := req.Values[item.parameter]
		if !ok && parameter.Default != nil {
			value, ok = *parameter.Default, true
		}
		if !ok {
			if !parameter.Required {
				diagnostics = append(diagnostics, Diagnostic{
					Code:      DiagnosticMissingValue,
					Parameter: item.parameter,
					Offset:    item.offset,
					Message:   fmt.Sprintf("parameter %q has no value", item.parameter),
				})
			}
			continue
		}
		if value.Type != parameter.Type {
			continue
		}
		normalized, err := normalize(value)
		if err == nil {
			content.WriteString(normalized)
		}
	}
	if len(diagnostics) != 0 {
		return RenderResult{}, &RenderError{Diagnostics: diagnostics}
	}
	return RenderResult{Content: content.String()}, nil
}

func validateInputs(req RenderRequest) (map[string]Parameter, []Diagnostic) {
	parameters := make(map[string]Parameter, len(req.Schema))
	diagnostics := make([]Diagnostic, 0)
	for _, parameter := range req.Schema {
		if !parameterNamePattern.MatchString(parameter.Name) {
			diagnostics = append(diagnostics, Diagnostic{
				Code:      DiagnosticInvalidSyntax,
				Parameter: parameter.Name,
				Message:   fmt.Sprintf("invalid parameter name %q", parameter.Name),
			})
			continue
		}
		if _, exists := parameters[parameter.Name]; exists {
			diagnostics = append(diagnostics, Diagnostic{
				Code:      DiagnosticDuplicateParameter,
				Parameter: parameter.Name,
				Message:   fmt.Sprintf("duplicate parameter %q", parameter.Name),
			})
			continue
		}
		parameters[parameter.Name] = parameter
		if !isSupportedScalarType(parameter.Type) {
			diagnostics = append(diagnostics, Diagnostic{
				Code:      DiagnosticInvalidValue,
				Parameter: parameter.Name,
				Message:   fmt.Sprintf("parameter %q has unsupported scalar type %q", parameter.Name, parameter.Type),
			})
			continue
		}
		if parameter.Default != nil {
			diagnostics = append(diagnostics, validateScalar(parameter.Name, parameter.Type, *parameter.Default)...)
		}
		if parameter.Required {
			if _, exists := req.Values[parameter.Name]; !exists && parameter.Default == nil {
				diagnostics = append(diagnostics, Diagnostic{
					Code:      DiagnosticMissingValue,
					Parameter: parameter.Name,
					Message:   fmt.Sprintf("required parameter %q has no value", parameter.Name),
				})
			}
		}
	}

	valueNames := make([]string, 0, len(req.Values))
	for name := range req.Values {
		valueNames = append(valueNames, name)
	}
	sort.Strings(valueNames)
	for _, name := range valueNames {
		parameter, exists := parameters[name]
		if !exists {
			diagnostics = append(diagnostics, Diagnostic{
				Code:      DiagnosticUnknownValue,
				Parameter: name,
				Message:   fmt.Sprintf("value %q is not declared by the schema", name),
			})
			continue
		}
		diagnostics = append(diagnostics, validateScalar(name, parameter.Type, req.Values[name])...)
	}
	return parameters, diagnostics
}

func isSupportedScalarType(scalarType ScalarType) bool {
	switch scalarType {
	case ScalarString, ScalarBoolean, ScalarInteger, ScalarDecimal:
		return true
	default:
		return false
	}
}

func validateScalar(name string, expected ScalarType, value Scalar) []Diagnostic {
	if value.Type != expected {
		return []Diagnostic{{
			Code:      DiagnosticTypeMismatch,
			Parameter: name,
			Message:   fmt.Sprintf("parameter %q expects %s, got %s", name, expected, value.Type),
		}}
	}
	if _, err := normalize(value); err != nil {
		return []Diagnostic{{
			Code:      DiagnosticInvalidValue,
			Parameter: name,
			Message:   fmt.Sprintf("parameter %q: %v", name, err),
		}}
	}
	return nil
}

func parse(input string) ([]segment, []Diagnostic) {
	if !utf8.ValidString(input) {
		return nil, []Diagnostic{{
			Code:    DiagnosticInvalidSyntax,
			Message: "template is not valid UTF-8",
		}}
	}

	segments := make([]segment, 0)
	diagnostics := make([]Diagnostic, 0)
	literalStart := 0
	for offset := 0; offset < len(input); {
		if strings.HasPrefix(input[offset:], "{{{") {
			if literalStart < offset {
				segments = append(segments, segment{literal: input[literalStart:offset]})
			}
			closeAt := strings.Index(input[offset+3:], "}}}")
			if closeAt < 0 {
				diagnostics = append(diagnostics, Diagnostic{
					Code:    DiagnosticInvalidSyntax,
					Offset:  offset,
					Message: fmt.Sprintf("unclosed placeholder at byte %d", offset),
				})
				return segments, diagnostics
			}
			closeAt += offset + 3
			name := input[offset+3 : closeAt]
			if !parameterNamePattern.MatchString(name) {
				diagnostics = append(diagnostics, Diagnostic{
					Code:      DiagnosticInvalidSyntax,
					Parameter: name,
					Offset:    offset,
					Message:   fmt.Sprintf("invalid placeholder %q at byte %d", name, offset),
				})
			} else {
				segments = append(segments, segment{parameter: name, offset: offset})
			}
			offset = closeAt + 3
			literalStart = offset
			continue
		}
		if strings.HasPrefix(input[offset:], "{{") || strings.HasPrefix(input[offset:], "}}") {
			token := input[offset : offset+2]
			diagnostics = append(diagnostics, Diagnostic{
				Code:    DiagnosticInvalidSyntax,
				Offset:  offset,
				Message: fmt.Sprintf("unsupported template token %q at byte %d", token, offset),
			})
			offset += 2
			continue
		}
		_, width := utf8.DecodeRuneInString(input[offset:])
		offset += width
	}
	if literalStart < len(input) {
		segments = append(segments, segment{literal: input[literalStart:]})
	}
	return segments, diagnostics
}

func normalize(value Scalar) (string, error) {
	switch value.Type {
	case ScalarString:
		if !utf8.ValidString(value.Value) {
			return "", fmt.Errorf("string is not valid UTF-8")
		}
		return value.Value, nil
	case ScalarBoolean:
		if value.Value != "true" && value.Value != "false" {
			return "", fmt.Errorf("invalid boolean %q", value.Value)
		}
		return value.Value, nil
	case ScalarInteger:
		if !canonicalIntegerPattern.MatchString(value.Value) || value.Value == "-0" {
			return "", fmt.Errorf("invalid integer %q", value.Value)
		}
		return value.Value, nil
	case ScalarDecimal:
		return normalizeDecimal(value.Value)
	default:
		return "", fmt.Errorf("unsupported scalar type %q", value.Type)
	}
}

func normalizeDecimal(input string) (string, error) {
	if !canonicalDecimalPattern.MatchString(input) || input == "-0" {
		return "", fmt.Errorf("invalid decimal %q", input)
	}
	return input, nil
}

func splitSign(input string) (string, string, error) {
	if input == "" {
		return "", "", fmt.Errorf("empty number")
	}
	sign := ""
	if input[0] == '+' || input[0] == '-' {
		sign, input = input[:1], input[1:]
	}
	if input == "" {
		return "", "", fmt.Errorf("number has no digits")
	}
	return sign, input, nil
}

func isDigits(input string) bool {
	if input == "" {
		return false
	}
	for _, char := range input {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func trimIntegerZeros(input string) string {
	input = strings.TrimLeft(input, "0")
	if input == "" {
		return "0"
	}
	return input
}
