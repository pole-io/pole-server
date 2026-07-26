package cache

import (
	"context"
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pole-io/pole-server/plugin/apiserver/xdsserverv3/resource"
)

func TestNodeScopedPolicyResourcesAreIsolatedAndVersionedByContent(t *testing.T) {
	cache := NewResourceCache(nil)
	stable := &routev3.RouteConfiguration{Name: "routes", ValidateClusters: wrapperspb.Bool(true)}
	canary := &routev3.RouteConfiguration{Name: "routes", ValidateClusters: wrapperspb.Bool(false)}
	err := cache.UpdateResources(context.Background(), &UpdateResourcesRequest{
		NodeResources: map[string]map[resource.XDSType]map[string]types.Resource{
			"stable": {resource.RDS: {"routes": stable}},
			"canary": {resource.RDS: {"routes": canary}},
		},
	})
	require.NoError(t, err)

	stableContainer, ok := cache.loadResourceContainer(&resource.XDSClient{ID: "stable"}, resource.RDS)
	require.True(t, ok)
	canaryContainer, ok := cache.loadResourceContainer(&resource.XDSClient{ID: "canary"}, resource.RDS)
	require.True(t, ok)
	require.NotEqual(t, stableContainer.GlobalVersion, canaryContainer.GlobalVersion)
	require.True(t, stableContainer.Resources["routes"].(*routev3.RouteConfiguration).
		GetValidateClusters().GetValue())
	require.False(t, canaryContainer.Resources["routes"].(*routev3.RouteConfiguration).
		GetValidateClusters().GetValue())

	firstVersion := stableContainer.GlobalVersion
	err = cache.UpdateResources(context.Background(), &UpdateResourcesRequest{
		NodeResources: map[string]map[resource.XDSType]map[string]types.Resource{
			"stable": {resource.RDS: {"routes": stable}},
		},
	})
	require.NoError(t, err)
	stableContainer, ok = cache.loadResourceContainer(&resource.XDSClient{ID: "stable"}, resource.RDS)
	require.True(t, ok)
	require.Equal(t, firstVersion, stableContainer.GlobalVersion)
}

func TestNodeScopedPolicyUpdateRejectsInvalidResourceAtomically(t *testing.T) {
	cache := NewResourceCache(nil)
	valid := &routev3.RouteConfiguration{Name: "routes"}
	err := cache.UpdateResources(context.Background(), &UpdateResourcesRequest{
		Lds: map[string]map[string]types.Resource{
			"node": {"listener-routes": valid},
		},
		NodeResources: map[string]map[resource.XDSType]map[string]types.Resource{
			"node": {resource.RDS: {"routes": valid}},
		},
	})
	require.NoError(t, err)
	beforeLDS := cache.ldsResources["node"]
	before := cache.nodeResources["node"][resource.RDS]

	invalid := &routev3.RouteConfiguration{Name: string([]byte{0xff})}
	err = cache.UpdateResources(context.Background(), &UpdateResourcesRequest{
		Lds: map[string]map[string]types.Resource{
			"node": {"listener-routes": &routev3.RouteConfiguration{Name: "changed"}},
		},
		NodeResources: map[string]map[resource.XDSType]map[string]types.Resource{
			"node": {resource.RDS: {"routes": invalid}},
		},
	})
	require.Error(t, err)
	require.Same(t, beforeLDS, cache.ldsResources["node"])
	require.Same(t, before, cache.nodeResources["node"][resource.RDS])
}
