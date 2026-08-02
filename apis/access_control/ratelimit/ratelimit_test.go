package ratelimit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/pluginapi"
)

type testRatelimit struct {
	initializeCount int
	destroyCount    int
}

func (r *testRatelimit) Name() string {
	return "test-ratelimit"
}

func (r *testRatelimit) Initialize(*apis.ConfigEntry) error {
	r.initializeCount++
	return nil
}

func (r *testRatelimit) Destroy() error {
	r.destroyCount++
	return nil
}

func (r *testRatelimit) Type() apis.PluginType {
	return apis.PluginTypeRateLimit
}

func (r *testRatelimit) Allow(RatelimitType, string) bool {
	return true
}

func TestResolveRatelimitUsesActiveRegistryInstance(t *testing.T) {
	previousConfig := apis.GetPluginConfig()
	apis.SetPluginConfig(&apis.Config{
		RateLimit: apis.ConfigEntry{Name: "test-ratelimit"},
	})
	rateLimitRuntime.Reset()

	registry := pluginapi.NewRegistry()
	created := 0
	var instance *testRatelimit
	require.NoError(t, apis.RegisterPluginFactory(registry, pluginapi.Descriptor{
		Kind: pluginapi.KindRateLimit,
		Name: "test-ratelimit",
	}, func() (apis.Plugin, error) {
		created++
		instance = &testRatelimit{}
		return instance, nil
	}))
	registry.Freeze()
	restore, err := pluginapi.Activate(registry)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = registry.Close()
		restore()
		rateLimitRuntime.Reset()
		apis.SetPluginConfig(previousConfig)
	})

	first, err := ResolveRatelimit()
	require.NoError(t, err)
	second, err := ResolveRatelimit()
	require.NoError(t, err)
	assert.Same(t, first, second)
	assert.Equal(t, 1, created)
	assert.Equal(t, 1, instance.initializeCount)

	require.NoError(t, registry.Close())
	assert.Equal(t, 1, instance.destroyCount)
}

func TestResolveRatelimitAllowsEmptyConfiguration(t *testing.T) {
	previousConfig := apis.GetPluginConfig()
	t.Cleanup(func() {
		rateLimitRuntime.Reset()
		apis.SetPluginConfig(previousConfig)
	})
	apis.SetPluginConfig(&apis.Config{})
	rateLimitRuntime.Reset()

	rateLimit, err := ResolveRatelimit()
	require.NoError(t, err)
	assert.Nil(t, rateLimit)
}
