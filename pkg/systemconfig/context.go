package systemconfig

import "context"

type providerContextKey struct {
	component Component
}

func WithProvider(ctx context.Context, component Component, provider EffectiveProvider) context.Context {
	return context.WithValue(ctx, providerContextKey{component: component}, provider)
}

func ProviderFromContext(ctx context.Context, component Component) (EffectiveProvider, bool) {
	provider, ok := ctx.Value(providerContextKey{component: component}).(EffectiveProvider)
	return provider, ok
}
