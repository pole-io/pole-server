package xdsserverv3

import (
	"context"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	apitypes "github.com/pole-io/pole-server/apis/pkg/types"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/plugin/apiserver/xdsserverv3/resource"
)

type labelSelectingRuleProvider struct{}

func (labelSelectingRuleProvider) response(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	revision := "normal"
	filter, _ := ctx.Value(apitypes.ContextDiscoverFilter).(*apiservice.DiscoverFilter)
	for _, label := range filter.GetCaller().GetLabels() {
		if label.GetKey() == "env" && label.GetValue().GetValue() == "canary" {
			revision = "gray-v2"
		}
	}
	return &apiservice.DiscoverResponse{
		Code:    uint32(apimodel.Code_ExecuteSuccess),
		Service: &apiservice.Service{Name: req.GetName(), Namespace: req.GetNamespace(), Revision: revision},
	}
}

func (p labelSelectingRuleProvider) GetRouterRuleWithCache(
	ctx context.Context, req *apiservice.Service,
) *apiservice.DiscoverResponse {
	resp := p.response(ctx, req)
	resp.CustomRouteRules = []*apitraffic.RouteRule{{Name: resp.GetService().GetRevision()}}
	return resp
}

func (p labelSelectingRuleProvider) GetRateLimitWithCache(
	ctx context.Context, req *apiservice.Service,
) *apiservice.DiscoverResponse {
	return p.response(ctx, req)
}

func (p labelSelectingRuleProvider) GetCircuitBreakerWithCache(
	ctx context.Context, req *apiservice.Service,
) *apiservice.DiscoverResponse {
	return p.response(ctx, req)
}

func (p labelSelectingRuleProvider) GetFaultDetectWithCache(
	ctx context.Context, req *apiservice.Service,
) *apiservice.DiscoverResponse {
	return p.response(ctx, req)
}

func TestGovernanceContextForNodeUsesExplicitMetadataLabels(t *testing.T) {
	node := &resource.XDSClient{Metadata: map[string]string{
		resource.GovernanceLabelPrefix + "env": "canary",
		resource.SidecarTLSModeTag:             "strict",
	}}

	ctx := governanceContextForNode(context.Background(), node)
	filter, ok := ctx.Value(apitypes.ContextDiscoverFilter).(*apiservice.DiscoverFilter)
	require.True(t, ok)
	require.Len(t, filter.GetCaller().GetLabels(), 1)
	require.Equal(t, "env", filter.GetCaller().GetLabels()[0].GetKey())
	require.Equal(t, "canary", filter.GetCaller().GetLabels()[0].GetValue().GetValue())
	require.Equal(t,
		governanceSelectionKey(&resource.XDSClient{Namespace: "prod", Metadata: map[string]string{
			resource.GovernanceLabelPrefix + "env":    "canary",
			resource.GovernanceLabelPrefix + "region": "shanghai",
		}}),
		governanceSelectionKey(&resource.XDSClient{Namespace: "prod", Metadata: map[string]string{
			resource.GovernanceLabelPrefix + "region": "shanghai",
			resource.GovernanceLabelPrefix + "env":    "canary",
		}}),
	)
	require.NotEqual(t,
		governanceSelectionKey(&resource.XDSClient{Namespace: "prod", Metadata: map[string]string{
			resource.GovernanceLabelPrefix + "a": "b=c",
		}}),
		governanceSelectionKey(&resource.XDSClient{Namespace: "prod", Metadata: map[string]string{
			resource.GovernanceLabelPrefix + "a=b": "c",
		}}),
	)
}

func TestServicesForNodeSelectsNormalOrGrayByGovernanceLabels(t *testing.T) {
	serviceKey := svctypes.ServiceKey{Namespace: "prod", Name: "payment"}
	generator := &XdsResourceGenerator{
		ruleServer: labelSelectingRuleProvider{},
		svcInfoProvider: func() ServiceInfos {
			return ServiceInfos{"prod": {
				serviceKey: {Name: serviceKey.Name, Namespace: serviceKey.Namespace, ServiceKey: serviceKey},
			}}
		},
	}

	normal, err := generator.servicesForNode(&resource.XDSClient{Namespace: "prod"})
	require.NoError(t, err)
	require.Equal(t, "normal", normal[serviceKey].Routing[0].GetName())
	require.Equal(t, "normal", normal[serviceKey].SvcRoutingRevision)

	gray, err := generator.servicesForNode(&resource.XDSClient{
		Namespace: "prod",
		Metadata: map[string]string{
			resource.GovernanceLabelPrefix + "env": "canary",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "gray-v2", gray[serviceKey].Routing[0].GetName())
	require.Equal(t, "gray-v2", gray[serviceKey].SvcRoutingRevision)
}

func TestSidecarAndGatewayRoutesUseExplicitCallerCalleeAndDestinationWeights(t *testing.T) {
	caller := svctypes.ServiceKey{Namespace: "prod", Name: "gateway"}
	callee := svctypes.ServiceKey{Namespace: "prod", Name: "payment"}
	custom := &apitraffic.CustomRoute{
		Caller: &apitraffic.CustomRoute_ServiceKey{Namespace: caller.Namespace, Service: caller.Name},
		Callee: &apitraffic.CustomRoute_ServiceKey{Namespace: callee.Namespace, Service: callee.Name},
		Rules: []*apitraffic.CustomRouteRule{{
			Name: "weighted",
			Arguments: &apitraffic.TrafficMatchRule{Arguments: []*apitraffic.SourceMatch{{
				Type:  apitraffic.SourceMatch_PATH,
				Value: &apimodel.MatchString{Type: apimodel.MatchString_EXACT, Value: "/pay"},
			}}},
			Destinations: []*apitraffic.DestinationGroup{
				{
					Namespace: callee.Namespace,
					Service:   callee.Name,
					Weight:    80,
					Labels: map[string]*apimodel.MatchString{
						"version": {Type: apimodel.MatchString_EXACT, Value: "v2"},
					},
				},
				{Namespace: callee.Namespace, Service: callee.Name, Weight: 20},
			},
		}},
	}
	config, err := anypb.New(custom)
	require.NoError(t, err)
	services := map[svctypes.ServiceKey]*resource.ServiceInfo{
		callee: {
			Name:       callee.Name,
			Namespace:  callee.Namespace,
			ServiceKey: callee,
			Routing: []*apitraffic.RouteRule{{
				RoutePolicy:   apitraffic.RoutePolicy_RulePolicy,
				RoutingConfig: config,
			}},
		},
	}

	sidecarRaw, err := (&VHDSBuilder{}).Generate(&resource.BuildOption{
		RunType:          resource.RunTypeSidecar,
		Namespace:        "prod",
		SelfService:      caller,
		Services:         services,
		TrafficDirection: corev3.TrafficDirection_OUTBOUND,
	})
	require.NoError(t, err)
	sidecarHosts := sidecarRaw.([]types.Resource)
	require.Len(t, sidecarHosts, 1)
	sidecarRoute := sidecarHosts[0].(*routev3.VirtualHost).GetRoutes()[0]
	require.Equal(t, "/pay", sidecarRoute.GetMatch().GetPath())
	require.Equal(t, uint32(80), sidecarRoute.GetRoute().GetWeightedClusters().GetClusters()[0].GetWeight().GetValue())
	require.Equal(t, uint32(20), sidecarRoute.GetRoute().GetWeightedClusters().GetClusters()[1].GetWeight().GetValue())
	require.Equal(t, "v2", sidecarRoute.GetRoute().GetWeightedClusters().GetClusters()[0].
		GetMetadataMatch().GetFilterMetadata()["envoy.lb"].GetFields()["version"].GetStringValue())

	gatewayRaw, err := (&RDSBuilder{}).Generate(&resource.BuildOption{
		RunType:          resource.RunTypeGateway,
		Namespace:        "prod",
		SelfService:      caller,
		Services:         services,
		TrafficDirection: corev3.TrafficDirection_OUTBOUND,
	})
	require.NoError(t, err)
	gatewayConfigs := gatewayRaw.([]types.Resource)
	require.Len(t, gatewayConfigs, 1)
	gatewayRoute := gatewayConfigs[0].(*routev3.RouteConfiguration).
		GetVirtualHosts()[0].GetRoutes()[0]
	require.Equal(t, "/pay", gatewayRoute.GetMatch().GetPath())
	require.Equal(t, uint32(80), gatewayRoute.GetRoute().GetWeightedClusters().GetClusters()[0].GetWeight().GetValue())
	require.Equal(t, uint32(20), gatewayRoute.GetRoute().GetWeightedClusters().GetClusters()[1].GetWeight().GetValue())
	require.Equal(t, "v2", gatewayRoute.GetRoute().GetWeightedClusters().GetClusters()[0].
		GetMetadataMatch().GetFilterMetadata()["envoy.lb"].GetFields()["version"].GetStringValue())
}

func TestGatewayRoutesUseStableServiceOrder(t *testing.T) {
	caller := svctypes.ServiceKey{Namespace: "prod", Name: "gateway"}
	makeService := func(name string) *resource.ServiceInfo {
		key := svctypes.ServiceKey{Namespace: "prod", Name: name}
		custom := &apitraffic.CustomRoute{
			Caller: &apitraffic.CustomRoute_ServiceKey{Namespace: caller.Namespace, Service: caller.Name},
			Callee: &apitraffic.CustomRoute_ServiceKey{Namespace: key.Namespace, Service: key.Name},
			Rules: []*apitraffic.CustomRouteRule{{
				Name: name,
				Arguments: &apitraffic.TrafficMatchRule{Arguments: []*apitraffic.SourceMatch{{
					Type:  apitraffic.SourceMatch_PATH,
					Value: &apimodel.MatchString{Type: apimodel.MatchString_EXACT, Value: "/" + name},
				}}},
				Destinations: []*apitraffic.DestinationGroup{{
					Namespace: key.Namespace, Service: key.Name, Weight: 100,
				}},
			}},
		}
		config, err := anypb.New(custom)
		require.NoError(t, err)
		return &resource.ServiceInfo{
			Name: name, Namespace: key.Namespace, ServiceKey: key,
			Routing: []*apitraffic.RouteRule{{
				RoutePolicy: apitraffic.RoutePolicy_RulePolicy, RoutingConfig: config,
			}},
		}
	}
	services := map[svctypes.ServiceKey]*resource.ServiceInfo{
		{Namespace: "prod", Name: "zeta"}:  makeService("zeta"),
		{Namespace: "prod", Name: "alpha"}: makeService("alpha"),
	}

	raw, err := (&RDSBuilder{}).Generate(&resource.BuildOption{
		RunType: resource.RunTypeGateway, Namespace: "prod", SelfService: caller,
		Services: services, TrafficDirection: corev3.TrafficDirection_OUTBOUND,
	})
	require.NoError(t, err)
	routes := raw.([]types.Resource)[0].(*routev3.RouteConfiguration).
		GetVirtualHosts()[0].GetRoutes()
	require.Len(t, routes, 3)
	require.Equal(t, "/alpha", routes[0].GetMatch().GetPath())
	require.Equal(t, "/zeta", routes[1].GetMatch().GetPath())
}
