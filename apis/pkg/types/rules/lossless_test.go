package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLosslessRuleSpecContextRoundTrip(t *testing.T) {
	original := &LosslessRule{
		ID:          "lossless-1",
		Namespace:   "production",
		Service:     "checkout",
		Description: "checkout graceful lifecycle",
		Metadata:    map[string]string{"owner": "platform"},
	}

	spec := original.ToSpec()
	require.Equal(t, "production", spec.Metadata[losslessNamespaceMetadata])
	require.Equal(t, "checkout", spec.Metadata[losslessServiceMetadata])
	require.Equal(t, "checkout graceful lifecycle", spec.Metadata[losslessDescriptionMetadata])
	require.NotContains(t, original.Metadata, losslessNamespaceMetadata)

	restored := &LosslessRule{}
	restored.FromSpec(spec)
	require.Equal(t, original.Namespace, restored.Namespace)
	require.Equal(t, original.Service, restored.Service)
	require.Equal(t, original.Description, restored.Description)
	require.Equal(t, map[string]string{"owner": "platform"}, restored.Metadata)
	require.NotContains(t, restored.Proto.Metadata, losslessNamespaceMetadata)
}
