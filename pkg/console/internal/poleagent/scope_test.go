package poleagent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNamespaceScopeValidate(t *testing.T) {
	tests := []struct {
		name    string
		scope   NamespaceScope
		wantErr string
	}{
		{
			name:  "single",
			scope: NamespaceScope{Mode: NamespaceScopeSingle, Namespaces: []string{"default"}},
		},
		{
			name: "cross environment",
			scope: NamespaceScope{
				Mode:       NamespaceScopeCrossEnvironment,
				Namespaces: []string{"development", "production"},
			},
		},
		{
			name:    "missing mode",
			scope:   NamespaceScope{},
			wantErr: "unsupported namespace scope mode",
		},
		{
			name: "single with multiple namespaces",
			scope: NamespaceScope{
				Mode:       NamespaceScopeSingle,
				Namespaces: []string{"development", "production"},
			},
			wantErr: "requires exactly one namespace",
		},
		{
			name: "cross environment with one namespace",
			scope: NamespaceScope{
				Mode:       NamespaceScopeCrossEnvironment,
				Namespaces: []string{"development"},
			},
			wantErr: "requires at least two namespaces",
		},
		{
			name: "empty namespace",
			scope: NamespaceScope{
				Mode:       NamespaceScopeSingle,
				Namespaces: []string{" "},
			},
			wantErr: "empty namespace",
		},
		{
			name: "duplicate namespace",
			scope: NamespaceScope{
				Mode:       NamespaceScopeCrossEnvironment,
				Namespaces: []string{"default", " default "},
			},
			wantErr: "duplicate namespace",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.scope.Validate()
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestNamespaceScopeContainsAndSingleNamespace(t *testing.T) {
	scope := NamespaceScope{Mode: NamespaceScopeSingle, Namespaces: []string{" default "}}
	require.True(t, scope.Contains("default"))
	require.True(t, scope.IsSingle())
	require.Equal(t, "default", scope.SingleNamespace())

	cross := NamespaceScope{
		Mode:       NamespaceScopeCrossEnvironment,
		Namespaces: []string{"development", "production"},
	}
	require.True(t, cross.Contains("production"))
	require.False(t, cross.IsSingle())
	require.Empty(t, cross.SingleNamespace())
}
