package resource

import (
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
)

func TestFilterOutboundRouterRulesUsesCallerAndCalleeAndKeepsDestinations(t *testing.T) {
	custom := &apitraffic.CustomRoute{
		Caller: &apitraffic.CustomRoute_ServiceKey{Namespace: "prod", Service: "checkout"},
		Callee: &apitraffic.CustomRoute_ServiceKey{Namespace: "prod", Service: "payment"},
		Rules: []*apitraffic.CustomRouteRule{{
			Name: "canary",
			Arguments: &apitraffic.TrafficMatchRule{Arguments: []*apitraffic.SourceMatch{{
				Type:  apitraffic.SourceMatch_HEADER,
				Key:   "x-tenant",
				Value: &apimodel.MatchString{Type: apimodel.MatchString_EXACT, Value: "blue"},
			}}},
			Destinations: []*apitraffic.DestinationGroup{
				{Namespace: "prod", Service: "payment", Weight: 80},
				{Namespace: "prod", Service: "payment", Weight: 20},
			},
		}},
	}
	config, err := anypb.New(custom)
	require.NoError(t, err)
	service := &ServiceInfo{
		Name:       "payment",
		Namespace:  "prod",
		ServiceKey: svctypes.ServiceKey{Namespace: "prod", Name: "payment"},
		Routing: []*apitraffic.RouteRule{{
			RoutePolicy:   apitraffic.RoutePolicy_RulePolicy,
			RoutingConfig: config,
		}},
	}

	matched := FilterOutboundRouterRules(service, svctypes.ServiceKey{Namespace: "prod", Name: "checkout"})
	require.Len(t, matched, 1)
	require.Equal(t, uint32(80), matched[0].GetDestinations()[0].GetWeight())
	require.Equal(t, uint32(20), matched[0].GetDestinations()[1].GetWeight())

	unmatched := FilterOutboundRouterRules(service, svctypes.ServiceKey{Namespace: "prod", Name: "catalog"})
	require.Empty(t, unmatched)
}

func TestEnvoyTrafficMatchCapabilityBoundaryAndSampling(t *testing.T) {
	dynamic := &apitraffic.TrafficMatchRule{Arguments: []*apitraffic.SourceMatch{{
		Type: apitraffic.SourceMatch_HEADER,
		Key:  "x-user",
		Value: &apimodel.MatchString{
			Type:      apimodel.MatchString_EXACT,
			ValueType: apimodel.MatchString_PARAMETER,
		},
	}}}
	require.False(t, SupportsEnvoyTrafficMatch(dynamic))
	require.False(t, SupportsEnvoyTrafficMatch(&apitraffic.TrafficMatchRule{
		MatchMode: apitraffic.TrafficMatchRule_OR,
	}))
	require.False(t, SupportsEnvoyTrafficMatch(&apitraffic.TrafficMatchRule{
		Arguments: []*apitraffic.SourceMatch{{
			Type:  apitraffic.SourceMatch_PATH,
			Value: &apimodel.MatchString{Type: apimodel.MatchString_RANGE, Value: "1,10"},
		}},
	}))

	supported := &apitraffic.TrafficMatchRule{
		RandomPercent: 25,
		Arguments: []*apitraffic.SourceMatch{{
			Type:  apitraffic.SourceMatch_METHOD,
			Value: &apimodel.MatchString{Type: apimodel.MatchString_EXACT, Value: "GET"},
		}},
	}
	require.True(t, SupportsEnvoyTrafficMatch(supported))
	routeMatch := &routev3.RouteMatch{
		PathSpecifier: &routev3.RouteMatch_Prefix{Prefix: "/"},
	}
	BuildSidecarRouteMatch(routeMatch, supported)
	require.Equal(t, uint32(25), routeMatch.GetRuntimeFraction().GetDefaultValue().GetNumerator())
}

func TestEnvoyRateLimitCapabilityBoundaryRejectsDynamicAndUnsupportedConditions(t *testing.T) {
	dynamic := &apitraffic.LimitTrigger{
		Name: "dynamic",
		Arguments: []*apitraffic.MatchArgument{{
			Type: apitraffic.MatchArgument_HEADER,
			Key:  "x-user",
			Value: &apimodel.MatchString{
				Type:      apimodel.MatchString_EXACT,
				ValueType: apimodel.MatchString_PARAMETER,
			},
		}},
	}
	require.False(t, SupportsEnvoyLimitTrigger(dynamic))
	require.False(t, SupportsEnvoyLimitTrigger(&apitraffic.LimitTrigger{
		Apis: []*apimodel.API{{Protocol: "grpc", Path: &apimodel.MatchString{
			Type: apimodel.MatchString_EXACT, Value: "/pay",
		}}},
	}))
	require.False(t, SupportsEnvoyLimitTrigger(&apitraffic.LimitTrigger{
		Resource: apitraffic.LimitTrigger_CONCURRENCY,
	}))
	require.False(t, SupportsEnvoyLimitTrigger(&apitraffic.LimitTrigger{
		MaxQueueDelay: 1,
	}))
	require.False(t, SupportsEnvoyRateLimit(&apitraffic.RateLimit{
		Cluster: &apitraffic.RateLimitCluster{},
	}))

	limits, filters, err := MakeSidecarLocalRateLimitFromRules([]*apitraffic.RateLimit{{
		Rules: []*apitraffic.LimitTrigger{dynamic},
	}}, svctypes.ServiceKey{Namespace: "prod", Name: "payment"})
	require.NoError(t, err)
	require.Empty(t, limits)
	require.Empty(t, filters)
}
