package whitelist

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/pluginapi"
)

type testWhitelist struct {
	initializeCount int
	destroyCount    int
}

func (w *testWhitelist) Name() string {
	return "test-whitelist"
}

func (w *testWhitelist) Initialize(*apis.ConfigEntry) error {
	w.initializeCount++
	return nil
}

func (w *testWhitelist) Destroy() error {
	w.destroyCount++
	return nil
}

func (w *testWhitelist) Type() apis.PluginType {
	return apis.PluginTypeWhitelist
}

func (w *testWhitelist) Contain(interface{}) bool {
	return true
}

func TestResolveWhitelistUsesActiveRegistryInstance(t *testing.T) {
	previousConfig := apis.GetPluginConfig()
	apis.SetPluginConfig(&apis.Config{
		Whitelist: apis.ConfigEntry{Name: "test-whitelist"},
	})
	whitelistRuntime.Reset()

	registry := pluginapi.NewRegistry()
	created := 0
	var instance *testWhitelist
	require.NoError(t, apis.RegisterPluginFactory(registry, pluginapi.Descriptor{
		Kind: pluginapi.KindWhitelist,
		Name: "test-whitelist",
	}, func() (apis.Plugin, error) {
		created++
		instance = &testWhitelist{}
		return instance, nil
	}))
	registry.Freeze()
	restore, err := pluginapi.Activate(registry)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = registry.Close()
		restore()
		whitelistRuntime.Reset()
		apis.SetPluginConfig(previousConfig)
	})

	first, err := ResolveWhitelist()
	require.NoError(t, err)
	second, err := ResolveWhitelist()
	require.NoError(t, err)
	assert.Same(t, first, second)
	assert.Equal(t, 1, created)
	assert.Equal(t, 1, instance.initializeCount)

	require.NoError(t, registry.Close())
	assert.Equal(t, 1, instance.destroyCount)
}

func TestResolveWhitelistAllowsEmptyConfiguration(t *testing.T) {
	previousConfig := apis.GetPluginConfig()
	t.Cleanup(func() {
		whitelistRuntime.Reset()
		apis.SetPluginConfig(previousConfig)
	})
	apis.SetPluginConfig(&apis.Config{})
	whitelistRuntime.Reset()

	whitelist, err := ResolveWhitelist()
	require.NoError(t, err)
	assert.Nil(t, whitelist)
}
