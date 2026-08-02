package pluginapi

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type Kind string

const (
	KindStatis        Kind = "plugin/statis"
	KindHistory       Kind = "plugin/history"
	KindDiscoverEvent Kind = "plugin/discover-event"
	KindRateLimit     Kind = "plugin/rate-limit"
	KindWhitelist     Kind = "plugin/whitelist"
	KindResourceAuth  Kind = "plugin/resource-auth"
	KindCMDB          Kind = "plugin/cmdb"
	KindCrypto        Kind = "plugin/crypto"
	KindHealthCheck   Kind = "plugin/health-check"
	KindStore         Kind = "store"
	KindAPIServer     Kind = "api-server"
	KindAuthUser      Kind = "auth/user"
	KindAuthStrategy  Kind = "auth/strategy"
)

type Origin string

const (
	APIVersionV1Alpha1 = "pole.io/plugin/v1alpha1"

	OriginBuiltin  Origin = "builtin"
	OriginExternal Origin = "external"
	OriginLegacy   Origin = "legacy"
)

type Descriptor struct {
	Kind       Kind
	Name       string
	Version    string
	APIVersion string
	Origin     Origin
}

type Factory func() (any, error)

type Destroyer interface {
	Destroy() error
}

type registryKey struct {
	kind Kind
	name string
}

type registration struct {
	descriptor Descriptor
	factory    Factory
	once       sync.Once
	instance   any
	err        error
}

type Registry struct {
	mu             sync.RWMutex
	frozen         bool
	closed         bool
	activeResolves int
	resolveCond    *sync.Cond
	entries        map[registryKey]*registration
	resolved       []*registration
	closeOnce      sync.Once
	closeErr       error
}

func NewRegistry() *Registry {
	registry := &Registry{entries: make(map[registryKey]*registration)}
	registry.resolveCond = sync.NewCond(&registry.mu)
	return registry
}

func (r *Registry) Register(descriptor Descriptor, factory Factory) error {
	if r == nil {
		return errors.New("plugin registry is nil")
	}
	descriptor.Name = strings.TrimSpace(descriptor.Name)
	if descriptor.Kind == "" {
		return errors.New("plugin kind is empty")
	}
	if descriptor.Name == "" {
		return errors.New("plugin name is empty")
	}
	if descriptor.Origin == "" {
		descriptor.Origin = OriginExternal
	}
	if descriptor.APIVersion == "" {
		descriptor.APIVersion = APIVersionV1Alpha1
	}
	if factory == nil {
		return fmt.Errorf("plugin factory is nil: kind=%s name=%s", descriptor.Kind, descriptor.Name)
	}

	key := registryKey{kind: descriptor.Kind, name: descriptor.Name}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.frozen {
		return errors.New("plugin registry is frozen")
	}
	if r.closed {
		return errors.New("plugin registry is closed")
	}
	if current, exists := r.entries[key]; exists {
		return fmt.Errorf("plugin already registered: kind=%s name=%s origin=%s",
			descriptor.Kind, descriptor.Name, current.descriptor.Origin)
	}
	r.entries[key] = &registration{descriptor: descriptor, factory: factory}
	return nil
}

func (r *Registry) MustRegister(descriptor Descriptor, factory Factory) {
	if err := r.Register(descriptor, factory); err != nil {
		panic(err)
	}
}

func (r *Registry) Resolve(kind Kind, name string) (any, error) {
	if r == nil {
		return nil, errors.New("plugin registry is nil")
	}
	key := registryKey{kind: kind, name: strings.TrimSpace(name)}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil, errors.New("plugin registry is closed")
	}
	entry, exists := r.entries[key]
	if !exists {
		r.mu.Unlock()
		return nil, fmt.Errorf("plugin not found: kind=%s name=%s", kind, key.name)
	}
	r.activeResolves++
	r.mu.Unlock()
	defer func() {
		r.mu.Lock()
		r.activeResolves--
		if r.activeResolves == 0 {
			r.resolveCond.Broadcast()
		}
		r.mu.Unlock()
	}()

	entry.once.Do(func() {
		entry.instance, entry.err = entry.factory()
		if entry.err != nil {
			entry.err = fmt.Errorf("create plugin: kind=%s name=%s: %w", kind, key.name, entry.err)
			return
		}
		if entry.instance == nil {
			entry.err = fmt.Errorf("plugin factory returned nil: kind=%s name=%s", kind, key.name)
			return
		}
		r.mu.Lock()
		r.resolved = append(r.resolved, entry)
		r.mu.Unlock()
	})
	if entry.err != nil {
		return nil, entry.err
	}
	return entry.instance, nil
}

func (r *Registry) Close() error {
	if r == nil {
		return nil
	}
	r.closeOnce.Do(func() {
		r.mu.Lock()
		r.closed = true
		for r.activeResolves > 0 {
			r.resolveCond.Wait()
		}
		resolved := append([]*registration(nil), r.resolved...)
		r.mu.Unlock()

		var errs []error
		for i := len(resolved) - 1; i >= 0; i-- {
			destroyer, ok := resolved[i].instance.(Destroyer)
			if !ok {
				continue
			}
			if err := destroyer.Destroy(); err != nil {
				descriptor := resolved[i].descriptor
				errs = append(errs, fmt.Errorf("destroy plugin: kind=%s name=%s: %w",
					descriptor.Kind, descriptor.Name, err))
			}
		}
		r.closeErr = errors.Join(errs...)
	})
	return r.closeErr
}

func (r *Registry) Contains(kind Kind, name string) bool {
	if r == nil {
		return false
	}
	key := registryKey{kind: kind, name: strings.TrimSpace(name)}
	r.mu.RLock()
	_, exists := r.entries[key]
	r.mu.RUnlock()
	return exists
}

func ResolveAs[T any](registry *Registry, kind Kind, name string) (T, error) {
	var zero T
	instance, err := registry.Resolve(kind, name)
	if err != nil {
		return zero, err
	}
	typed, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("plugin has incompatible type: kind=%s name=%s", kind, name)
	}
	return typed, nil
}

func (r *Registry) Descriptors(kind Kind) []Descriptor {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	descriptors := make([]Descriptor, 0, len(r.entries))
	for _, entry := range r.entries {
		if kind == "" || entry.descriptor.Kind == kind {
			descriptors = append(descriptors, entry.descriptor)
		}
	}
	r.mu.RUnlock()
	sort.Slice(descriptors, func(i, j int) bool {
		if descriptors[i].Kind != descriptors[j].Kind {
			return descriptors[i].Kind < descriptors[j].Kind
		}
		return descriptors[i].Name < descriptors[j].Name
	})
	return descriptors
}

func (r *Registry) Clone() *Registry {
	clone := NewRegistry()
	if r == nil {
		return clone
	}
	r.mu.RLock()
	for key, entry := range r.entries {
		clone.entries[key] = &registration{
			descriptor: entry.descriptor,
			factory:    entry.factory,
		}
	}
	r.mu.RUnlock()
	return clone
}

func (r *Registry) Merge(source *Registry) error {
	if r == nil {
		return errors.New("plugin registry is nil")
	}
	if source == nil || source == r {
		return nil
	}
	source.mu.RLock()
	entries := make([]registration, 0, len(source.entries))
	for _, entry := range source.entries {
		entries = append(entries, registration{
			descriptor: entry.descriptor,
			factory:    entry.factory,
		})
	}
	source.mu.RUnlock()
	for _, entry := range entries {
		if err := r.Register(entry.descriptor, entry.factory); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) Freeze() {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.frozen = true
	r.mu.Unlock()
}

func (r *Registry) Frozen() bool {
	if r == nil {
		return false
	}
	r.mu.RLock()
	frozen := r.frozen
	r.mu.RUnlock()
	return frozen
}

var (
	defaultRegistry = NewRegistry()
	activeRegistry  atomic.Pointer[Registry]
	activationMu    sync.Mutex
	activeLease     *activationLease
)

type activationLease struct {
	previous *Registry
	registry *Registry
	once     sync.Once
}

func init() {
	activeRegistry.Store(defaultRegistry)
}

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func ActiveRegistry() *Registry {
	return activeRegistry.Load()
}

func Activate(registry *Registry) (func(), error) {
	if registry == nil {
		return nil, errors.New("plugin registry is nil")
	}
	if !registry.Frozen() {
		return nil, errors.New("plugin registry must be frozen before activation")
	}
	activationMu.Lock()
	defer activationMu.Unlock()
	if activeLease != nil {
		return nil, errors.New("another plugin registry is already active")
	}
	lease := &activationLease{
		previous: activeRegistry.Load(),
		registry: registry,
	}
	activeLease = lease
	activeRegistry.Store(registry)
	return func() {
		lease.once.Do(func() {
			activationMu.Lock()
			defer activationMu.Unlock()
			if activeLease != lease {
				return
			}
			activeRegistry.Store(lease.previous)
			activeLease = nil
		})
	}, nil
}
