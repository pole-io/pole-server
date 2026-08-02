package pluginapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntimeValueIsScopedToActiveRegistry(t *testing.T) {
	var value RuntimeValue[*testPlugin]
	created := 0
	load := func() (*testPlugin, error) {
		created++
		return &testPlugin{id: created}, nil
	}

	firstRegistry := NewRegistry()
	firstRegistry.Freeze()
	restoreFirst, err := Activate(firstRegistry)
	require.NoError(t, err)
	first, err := value.Load(load)
	require.NoError(t, err)
	same, err := value.Load(load)
	require.NoError(t, err)
	assert.Same(t, first, same)
	assert.Equal(t, 1, created)
	restoreFirst()

	secondRegistry := NewRegistry()
	secondRegistry.Freeze()
	restoreSecond, err := Activate(secondRegistry)
	require.NoError(t, err)
	t.Cleanup(restoreSecond)
	second, err := value.Load(load)
	require.NoError(t, err)
	assert.NotSame(t, first, second)
	assert.Equal(t, 2, created)
}
