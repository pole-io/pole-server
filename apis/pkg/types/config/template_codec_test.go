package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
)

func TestTemplatePersistenceCodecStoresOnlySchemaAndValues(t *testing.T) {
	schemaJSON, err := EncodeTemplateParameterSchema([]*apiconfig.ConfigTemplateParameterSchema{{
		Name: "server.port", Type: apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_INTEGER,
		Required: true,
	}})
	require.NoError(t, err)
	require.JSONEq(t, `[{"name":"server.port","type":3,"required":true}]`, schemaJSON)

	valuesJSON, err := EncodeTemplateValues(map[string]*apiconfig.ConfigTemplateValue{
		"server.port": {Value: &apiconfig.ConfigTemplateValue_IntegerValue{IntegerValue: 8080}},
	})
	require.NoError(t, err)
	require.JSONEq(t, `{"server.port":{"type":"integer","value":"8080"}}`, valuesJSON)

	schema, err := DecodeTemplateParameterSchema(schemaJSON)
	require.NoError(t, err)
	require.Equal(t, "server.port", schema[0].GetName())
	values, err := DecodeTemplateValues(valuesJSON)
	require.NoError(t, err)
	require.Equal(t, int64(8080), values["server.port"].GetIntegerValue())
}
