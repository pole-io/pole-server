package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
)

func TestConfigFileTemplateTypeRoundTripsWithoutLeakingInternalMetadata(t *testing.T) {
	input := &apiconfig.ConfigFile{
		Name:       "application.yaml",
		Namespace:  "prod",
		Group:      "app",
		ConfigType: apiconfig.ConfigFile_CONFIG_TEMPLATE,
		Labels:     map[string]string{"owner": "platform"},
		TemplateBinding: &apiconfig.ConfigTemplateBinding{
			TemplateId: 7, TemplateReleaseId: "template-r1", BindingReleaseId: "binding-r1",
		},
	}
	stored := ToConfigFileStore(input)

	apiFile := ToConfigFileAPI(stored)
	require.Equal(t, apiconfig.ConfigFile_CONFIG_TEMPLATE, apiFile.GetConfigType())
	require.Equal(t, "platform", apiFile.GetLabels()["owner"])
	require.NotContains(t, apiFile.GetLabels(), metadataKeyConfigFileType)
	require.Equal(t, "binding-r1", apiFile.GetTemplateBinding().GetBindingReleaseId())
	require.NotContains(t, apiFile.GetLabels(), metadataKeyBindingReleaseID)
	require.Equal(t, map[string]string{"owner": "platform"}, input.GetLabels(),
		"persistence conversion must not write internal metadata into the API request")
}

func TestConfigFilePlainTypeIsTheCompatibleDefault(t *testing.T) {
	apiFile := ToConfigFileAPI(&ConfigFile{
		Name:     "application.yaml",
		Metadata: map[string]string{},
	})
	require.Equal(t, apiconfig.ConfigFile_CONFIG_FILE, apiFile.GetConfigType())
}

func TestConfigFileTemplateDraftRoundTripsEngineAndSchema(t *testing.T) {
	input := &apiconfig.ConfigFileTemplate{
		Id: 7, Name: "application", Content: "port={{{server.port}}}", Format: "text",
		Engine: &apiconfig.ConfigTemplateEngine{Name: "pole-mustache", Version: "v1"},
		ParameterSchema: []*apiconfig.ConfigTemplateParameterSchema{{
			Name: "server.port", Type: apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_INTEGER, Required: true,
		}},
	}
	stored := ToConfigFileTemplateStore(input)
	require.NotEmpty(t, stored.Revision)

	output := ToConfigFileTemplateAPI(stored)
	require.Equal(t, "pole-mustache", output.GetEngine().GetName())
	require.Equal(t, "v1", output.GetEngine().GetVersion())
	require.Len(t, output.GetParameterSchema(), 1)
	require.Equal(t, "server.port", output.GetParameterSchema()[0].GetName())
}
