package bootstrap

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/apis/apiserver"
	"github.com/pole-io/pole-server/pluginapi"
)

func TestRunRestoresPluginRegistryWhenConfigurationLoadFails(t *testing.T) {
	previous := pluginapi.ActiveRegistry()
	registry := pluginapi.NewRegistry()

	err := Run(context.Background(), Options{
		ConfigPath:     filepath.Join(t.TempDir(), "missing.yaml"),
		PluginRegistry: registry,
	})

	assert.ErrorContains(t, err, "load config")
	assert.Same(t, previous, pluginapi.ActiveRegistry())
}

func TestPreparePluginRegistryIncludesLegacyDefaultRegistrations(t *testing.T) {
	name := "bootstrap-legacy-test"
	require.NoError(t, pluginapi.DefaultRegistry().Register(pluginapi.Descriptor{
		Kind: pluginapi.KindHistory, Name: name, Origin: pluginapi.OriginLegacy,
	}, func() (any, error) {
		return name, nil
	}))
	explicit := pluginapi.NewRegistry()
	require.NoError(t, explicit.Register(pluginapi.Descriptor{
		Kind: pluginapi.KindHistory, Name: "explicit", Origin: pluginapi.OriginExternal,
	}, func() (any, error) {
		return "explicit", nil
	}))

	runtimeRegistry, err := preparePluginRegistry(explicit)
	require.NoError(t, err)

	assert.True(t, runtimeRegistry.Frozen())
	assert.True(t, runtimeRegistry.Contains(pluginapi.KindHistory, name))
	assert.True(t, runtimeRegistry.Contains(pluginapi.KindHistory, "explicit"))
}

func TestStartServersDoesNotInstantiateUnselectedPlugin(t *testing.T) {
	catalog := pluginapi.NewRegistry()
	created := 0
	require.NoError(t, apiserver.RegisterFactory(catalog, pluginapi.Descriptor{
		Name: "broken", Origin: pluginapi.OriginExternal,
	}, func() (apiserver.Apiserver, error) {
		created++
		return nil, errors.New("boom")
	}))
	runtimeRegistry, err := preparePluginRegistry(catalog)
	require.NoError(t, err)
	restore, err := pluginapi.Activate(runtimeRegistry)
	require.NoError(t, err)
	t.Cleanup(restore)

	servers, err := StartServers(context.Background(), nil, make(chan error, 1))

	require.NoError(t, err)
	assert.Empty(t, servers)
	assert.Zero(t, created)
}
