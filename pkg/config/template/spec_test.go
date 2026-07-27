package template_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"

	configtemplate "github.com/pole-io/pole-server/pkg/config/template"
)

func TestRenderRequestFromSpecPreservesTypedScalars(t *testing.T) {
	req, err := configtemplate.RenderRequestFromSpec(&apiconfig.ConfigTemplateRenderInput{
		Content: `port={{{service.port}}}, enabled={{{service.enabled}}}`,
		Format:  "text",
		Engine: &apiconfig.ConfigTemplateEngine{
			Name: configtemplate.EnginePoleMustache, Version: configtemplate.EngineVersionV1,
		},
		ParameterSchema: []*apiconfig.ConfigTemplateParameterSchema{
			{Name: "service.port", Type: apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_INTEGER, Required: true},
			{Name: "service.enabled", Type: apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_BOOLEAN, Required: true},
		},
		Values: map[string]*apiconfig.ConfigTemplateValue{
			"service.port": {
				Value: &apiconfig.ConfigTemplateValue_IntegerValue{IntegerValue: 8080},
			},
			"service.enabled": {
				Value: &apiconfig.ConfigTemplateValue_BooleanValue{BooleanValue: true},
			},
		},
	})
	require.NoError(t, err)

	result, err := configtemplate.Preview(t.Context(), req)
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Equal(t, "port=8080, enabled=true", result.RenderedContent)
}

func TestRenderRequestFromSpecRejectsUnsupportedEngine(t *testing.T) {
	_, err := configtemplate.RenderRequestFromSpec(&apiconfig.ConfigTemplateRenderInput{
		Engine: &apiconfig.ConfigTemplateEngine{Name: "go-template", Version: "v1"},
	})
	require.ErrorContains(t, err, "unsupported template engine")
}
