package pluginapi

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPlugin struct {
	id int
}

type lifecyclePlugin struct {
	name  string
	order *[]string
}

func (p *lifecyclePlugin) Destroy() error {
	*p.order = append(*p.order, p.name)
	return nil
}

func TestRegistryFactoryCreatesOneRuntimeInstance(t *testing.T) {
	registry := NewRegistry()
	nextID := 0
	require.NoError(t, registry.Register(Descriptor{
		Kind: KindStatis, Name: "test", Origin: OriginBuiltin,
	}, func() (any, error) {
		nextID++
		return &testPlugin{id: nextID}, nil
	}))

	first, err := ResolveAs[*testPlugin](registry, KindStatis, "test")
	require.NoError(t, err)
	second, err := ResolveAs[*testPlugin](registry, KindStatis, "test")
	require.NoError(t, err)

	assert.Equal(t, 1, first.id)
	assert.Same(t, first, second)
	assert.Equal(t, 1, nextID)
}

func TestRegistryRejectsDuplicateAndFrozenRegistration(t *testing.T) {
	registry := NewRegistry()
	descriptor := Descriptor{Kind: KindStore, Name: "test"}
	factory := func() (any, error) { return &testPlugin{}, nil }

	require.NoError(t, registry.Register(descriptor, factory))
	assert.ErrorContains(t, registry.Register(descriptor, factory), "already registered")

	registry.Freeze()
	assert.ErrorContains(t, registry.Register(
		Descriptor{Kind: KindStore, Name: "other"}, factory), "frozen")
}

func TestRegistryCloneCanBeExtendedIndependently(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, registry.Register(
		Descriptor{Kind: KindStore, Name: "builtin"},
		func() (any, error) { return &testPlugin{}, nil },
	))
	registry.Freeze()

	clone := registry.Clone()
	require.False(t, clone.Frozen())
	require.NoError(t, clone.Register(
		Descriptor{Kind: KindStore, Name: "external"},
		func() (any, error) { return &testPlugin{}, nil },
	))

	assert.Len(t, registry.Descriptors(KindStore), 1)
	assert.Len(t, clone.Descriptors(KindStore), 2)

	original, err := ResolveAs[*testPlugin](registry, KindStore, "builtin")
	require.NoError(t, err)
	cloned, err := ResolveAs[*testPlugin](clone, KindStore, "builtin")
	require.NoError(t, err)
	assert.NotSame(t, original, cloned)
}

func TestRegistryResolveReportsFactoryAndTypeErrors(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, registry.Register(
		Descriptor{Kind: KindCMDB, Name: "broken"},
		func() (any, error) { return nil, errors.New("boom") },
	))
	require.NoError(t, registry.Register(
		Descriptor{Kind: KindCMDB, Name: "wrong-type"},
		func() (any, error) { return "value", nil },
	))

	_, err := registry.Resolve(KindCMDB, "missing")
	assert.ErrorContains(t, err, "not found")
	_, err = registry.Resolve(KindCMDB, "broken")
	assert.ErrorContains(t, err, "boom")
	_, err = ResolveAs[*testPlugin](registry, KindCMDB, "wrong-type")
	assert.ErrorContains(t, err, "incompatible type")
}

func TestActivateRequiresFrozenRegistry(t *testing.T) {
	registry := NewRegistry()
	_, err := Activate(registry)
	assert.ErrorContains(t, err, "must be frozen")

	registry.Freeze()
	restore, err := Activate(registry)
	require.NoError(t, err)
	t.Cleanup(restore)
	assert.Same(t, registry, ActiveRegistry())

	other := NewRegistry()
	other.Freeze()
	_, err = Activate(other)
	assert.ErrorContains(t, err, "already active")
}

func TestRegistryConcurrentResolveCreatesInstanceOnce(t *testing.T) {
	registry := NewRegistry()
	created := 0
	var createdMu sync.Mutex
	require.NoError(t, registry.Register(
		Descriptor{Kind: KindHistory, Name: "test"},
		func() (any, error) {
			createdMu.Lock()
			defer createdMu.Unlock()
			created++
			return &testPlugin{id: created}, nil
		},
	))

	const callers = 32
	instances := make(chan *testPlugin, callers)
	errs := make(chan error, callers)
	var waitGroup sync.WaitGroup
	for range callers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			instance, err := ResolveAs[*testPlugin](registry, KindHistory, "test")
			errs <- err
			if err != nil {
				return
			}
			instances <- instance
		}()
	}
	waitGroup.Wait()
	close(instances)
	close(errs)

	var first *testPlugin
	for instance := range instances {
		if first == nil {
			first = instance
			continue
		}
		assert.Same(t, first, instance)
	}
	for err := range errs {
		require.NoError(t, err)
	}
	assert.Equal(t, 1, created)
}

func TestRegistryMergeCopiesFactoriesWithoutRuntimeInstances(t *testing.T) {
	source := NewRegistry()
	created := 0
	require.NoError(t, source.Register(
		Descriptor{Kind: KindCMDB, Name: "source"},
		func() (any, error) {
			created++
			return &testPlugin{id: created}, nil
		},
	))
	sourceInstance, err := ResolveAs[*testPlugin](source, KindCMDB, "source")
	require.NoError(t, err)

	target := NewRegistry()
	require.NoError(t, target.Merge(source))
	targetInstance, err := ResolveAs[*testPlugin](target, KindCMDB, "source")
	require.NoError(t, err)

	assert.NotSame(t, sourceInstance, targetInstance)
	assert.Equal(t, 2, created)
	assert.ErrorContains(t, target.Merge(source), "already registered")
}

func TestRegistryCloseDestroysResolvedPluginsInReverseOrder(t *testing.T) {
	registry := NewRegistry()
	var order []string
	for _, name := range []string{"first", "second", "unused"} {
		name := name
		require.NoError(t, registry.Register(
			Descriptor{Kind: KindHistory, Name: name},
			func() (any, error) {
				return &lifecyclePlugin{name: name, order: &order}, nil
			},
		))
	}

	_, err := registry.Resolve(KindHistory, "first")
	require.NoError(t, err)
	_, err = registry.Resolve(KindHistory, "second")
	require.NoError(t, err)

	require.NoError(t, registry.Close())
	require.NoError(t, registry.Close())
	assert.Equal(t, []string{"second", "first"}, order)
	_, err = registry.Resolve(KindHistory, "first")
	assert.ErrorContains(t, err, "closed")
}
