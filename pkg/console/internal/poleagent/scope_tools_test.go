package poleagent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScopeRemoteToolsKeepsOnlyNamespaceBoundTools(t *testing.T) {
	tools := []ToolDefinition{
		{
			Name: "get_config_file",
			InputSchema: map[string]any{"properties": map[string]any{
				"namespace": map[string]any{"type": "string"},
			}},
		},
		{Name: "list_namespaces", InputSchema: map[string]any{"properties": map[string]any{}}},
	}

	scoped := namespaceScopedRemoteTools(tools)
	require.Len(t, scoped, 1)
	require.Equal(t, "get_config_file", scoped[0].Name)
}

func TestEnforceToolNamespaceInjectsSingleScope(t *testing.T) {
	arguments := map[string]any{"group": "app"}
	require.NoError(t, enforceToolNamespace(singleScope("default"), arguments))
	require.Equal(t, "default", arguments["namespace"])
}

func TestEnforceToolNamespaceRejectsOutsideScope(t *testing.T) {
	arguments := map[string]any{"namespace": "production"}
	err := enforceToolNamespace(singleScope("development"), arguments)
	require.ErrorContains(t, err, "outside the turn scope")
}

func TestEnforceToolNamespaceRequiresExplicitSelectionForCrossEnvironment(t *testing.T) {
	scope := NamespaceScope{
		Mode:       NamespaceScopeCrossEnvironment,
		Namespaces: []string{"development", "production"},
	}
	err := enforceToolNamespace(scope, map[string]any{})
	require.ErrorContains(t, err, "must select one namespace")
}
