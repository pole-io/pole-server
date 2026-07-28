package observabilityquery

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigNormalizeDefaultsProviderAndTimeout(t *testing.T) {
	config := Config{}.Normalize()

	require.Equal(t, DefaultProvider, config.Provider)
	require.Equal(t, "public", config.Database)
	require.Equal(t, "10s", config.Timeout)
	require.False(t, config.Configured())
}
