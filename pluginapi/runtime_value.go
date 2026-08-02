package pluginapi

import "sync"

type RuntimeValue[T any] struct {
	mu       sync.Mutex
	registry *Registry
	value    T
	err      error
	loaded   bool
}

func (v *RuntimeValue[T]) Load(factory func() (T, error)) (T, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	registry := ActiveRegistry()
	if v.loaded && v.registry == registry {
		return v.value, v.err
	}
	v.registry = registry
	v.value, v.err = factory()
	v.loaded = true
	return v.value, v.err
}

func (v *RuntimeValue[T]) Reset() {
	v.mu.Lock()
	var zero T
	v.registry = nil
	v.value = zero
	v.err = nil
	v.loaded = false
	v.mu.Unlock()
}
