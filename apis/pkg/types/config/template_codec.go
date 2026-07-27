package config

import (
	"encoding/json"
	"fmt"
	"strconv"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
)

type persistedTemplateValue struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type persistedTemplateParameter struct {
	Name      string                  `json:"name"`
	Type      int32                   `json:"type"`
	Required  bool                    `json:"required"`
	Default   *persistedTemplateValue `json:"default,omitempty"`
	Sensitive bool                    `json:"sensitive,omitempty"`
	Comment   string                  `json:"comment,omitempty"`
}

func EncodeTemplateParameterSchema(schema []*apiconfig.ConfigTemplateParameterSchema) (string, error) {
	out := make([]persistedTemplateParameter, 0, len(schema))
	for _, parameter := range schema {
		if parameter == nil {
			return "", fmt.Errorf("template parameter schema contains nil entry")
		}
		item := persistedTemplateParameter{
			Name: parameter.GetName(), Type: int32(parameter.GetType()), Required: parameter.GetRequired(),
			Sensitive: parameter.GetSensitive(), Comment: parameter.GetComment(),
		}
		if parameter.GetDefaultValue() != nil {
			value, err := encodeTemplateValue(parameter.GetDefaultValue())
			if err != nil {
				return "", err
			}
			item.Default = &value
		}
		out = append(out, item)
	}
	data, err := json.Marshal(out)
	return string(data), err
}

func DecodeTemplateParameterSchema(data string) ([]*apiconfig.ConfigTemplateParameterSchema, error) {
	if data == "" {
		return nil, nil
	}
	var persisted []persistedTemplateParameter
	if err := json.Unmarshal([]byte(data), &persisted); err != nil {
		return nil, err
	}
	out := make([]*apiconfig.ConfigTemplateParameterSchema, 0, len(persisted))
	for _, item := range persisted {
		parameter := &apiconfig.ConfigTemplateParameterSchema{
			Name: item.Name, Type: apiconfig.ConfigTemplateParameterType(item.Type),
			Required: item.Required, Sensitive: item.Sensitive, Comment: item.Comment,
		}
		if item.Default != nil {
			value, err := decodeTemplateValue(*item.Default)
			if err != nil {
				return nil, err
			}
			parameter.DefaultValue = value
		}
		out = append(out, parameter)
	}
	return out, nil
}

func EncodeTemplateValues(values map[string]*apiconfig.ConfigTemplateValue) (string, error) {
	out := make(map[string]persistedTemplateValue, len(values))
	for name, value := range values {
		encoded, err := encodeTemplateValue(value)
		if err != nil {
			return "", fmt.Errorf("encode template value %q: %w", name, err)
		}
		out[name] = encoded
	}
	data, err := json.Marshal(out)
	return string(data), err
}

func DecodeTemplateValues(data string) (map[string]*apiconfig.ConfigTemplateValue, error) {
	if data == "" {
		return map[string]*apiconfig.ConfigTemplateValue{}, nil
	}
	var persisted map[string]persistedTemplateValue
	if err := json.Unmarshal([]byte(data), &persisted); err != nil {
		return nil, err
	}
	out := make(map[string]*apiconfig.ConfigTemplateValue, len(persisted))
	for name, value := range persisted {
		decoded, err := decodeTemplateValue(value)
		if err != nil {
			return nil, fmt.Errorf("decode template value %q: %w", name, err)
		}
		out[name] = decoded
	}
	return out, nil
}

func encodeTemplateValue(value *apiconfig.ConfigTemplateValue) (persistedTemplateValue, error) {
	if value == nil {
		return persistedTemplateValue{}, fmt.Errorf("template value is nil")
	}
	switch typed := value.GetValue().(type) {
	case *apiconfig.ConfigTemplateValue_StringValue:
		return persistedTemplateValue{Type: "string", Value: typed.StringValue}, nil
	case *apiconfig.ConfigTemplateValue_BooleanValue:
		return persistedTemplateValue{Type: "boolean", Value: strconv.FormatBool(typed.BooleanValue)}, nil
	case *apiconfig.ConfigTemplateValue_IntegerValue:
		return persistedTemplateValue{Type: "integer", Value: strconv.FormatInt(typed.IntegerValue, 10)}, nil
	case *apiconfig.ConfigTemplateValue_DecimalValue:
		return persistedTemplateValue{Type: "decimal", Value: typed.DecimalValue}, nil
	default:
		return persistedTemplateValue{}, fmt.Errorf("template scalar oneof is not set")
	}
}

func decodeTemplateValue(value persistedTemplateValue) (*apiconfig.ConfigTemplateValue, error) {
	out := &apiconfig.ConfigTemplateValue{}
	switch value.Type {
	case "string":
		out.Value = &apiconfig.ConfigTemplateValue_StringValue{StringValue: value.Value}
	case "boolean":
		parsed, err := strconv.ParseBool(value.Value)
		if err != nil {
			return nil, err
		}
		out.Value = &apiconfig.ConfigTemplateValue_BooleanValue{BooleanValue: parsed}
	case "integer":
		parsed, err := strconv.ParseInt(value.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		out.Value = &apiconfig.ConfigTemplateValue_IntegerValue{IntegerValue: parsed}
	case "decimal":
		out.Value = &apiconfig.ConfigTemplateValue_DecimalValue{DecimalValue: value.Value}
	default:
		return nil, fmt.Errorf("unsupported persisted template scalar type %q", value.Type)
	}
	return out, nil
}
