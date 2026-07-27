package template

import (
	"fmt"
	"strconv"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
)

// RenderRequestFromSpec converts the language-neutral protobuf contract into the
// strict pole-mustache-v1 renderer input.
func RenderRequestFromSpec(input *apiconfig.ConfigTemplateRenderInput) (PreviewRequest, error) {
	if input == nil {
		return PreviewRequest{}, fmt.Errorf("template render input is required")
	}
	engine := input.GetEngine()
	if engine == nil || engine.GetName() != EnginePoleMustache || engine.GetVersion() != EngineVersionV1 {
		return PreviewRequest{}, fmt.Errorf("unsupported template engine %q version %q",
			engine.GetName(), engine.GetVersion())
	}

	schema := make([]Parameter, 0, len(input.GetParameterSchema()))
	for _, item := range input.GetParameterSchema() {
		if item == nil {
			return PreviewRequest{}, fmt.Errorf("template parameter schema contains nil entry")
		}
		parameter := Parameter{
			Name:     item.GetName(),
			Type:     scalarTypeFromSpec(item.GetType()),
			Required: item.GetRequired(),
		}
		if item.GetDefaultValue() != nil {
			value, err := scalarFromSpec(item.GetDefaultValue())
			if err != nil {
				return PreviewRequest{}, fmt.Errorf("invalid default value for %q: %w", item.GetName(), err)
			}
			parameter.Default = &value
		}
		schema = append(schema, parameter)
	}

	values := make(map[string]Scalar, len(input.GetValues()))
	for name, item := range input.GetValues() {
		value, err := scalarFromSpec(item)
		if err != nil {
			return PreviewRequest{}, fmt.Errorf("invalid value for %q: %w", name, err)
		}
		values[name] = value
	}
	return PreviewRequest{
		RenderRequest: RenderRequest{
			Template: input.GetContent(),
			Schema:   schema,
			Values:   values,
		},
		Format: input.GetFormat(),
	}, nil
}

func scalarTypeFromSpec(value apiconfig.ConfigTemplateParameterType) ScalarType {
	switch value {
	case apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_STRING:
		return ScalarString
	case apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_BOOLEAN:
		return ScalarBoolean
	case apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_INTEGER:
		return ScalarInteger
	case apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_DECIMAL:
		return ScalarDecimal
	default:
		return ScalarType(value.String())
	}
}

func scalarFromSpec(value *apiconfig.ConfigTemplateValue) (Scalar, error) {
	if value == nil {
		return Scalar{}, fmt.Errorf("value is required")
	}
	switch typed := value.GetValue().(type) {
	case *apiconfig.ConfigTemplateValue_StringValue:
		return Scalar{Type: ScalarString, Value: typed.StringValue}, nil
	case *apiconfig.ConfigTemplateValue_BooleanValue:
		return Scalar{Type: ScalarBoolean, Value: strconv.FormatBool(typed.BooleanValue)}, nil
	case *apiconfig.ConfigTemplateValue_IntegerValue:
		return Scalar{Type: ScalarInteger, Value: strconv.FormatInt(typed.IntegerValue, 10)}, nil
	case *apiconfig.ConfigTemplateValue_DecimalValue:
		return Scalar{Type: ScalarDecimal, Value: typed.DecimalValue}, nil
	default:
		return Scalar{}, fmt.Errorf("scalar oneof is not set")
	}
}
