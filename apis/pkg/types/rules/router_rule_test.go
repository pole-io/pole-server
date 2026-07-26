package rules

import (
	"testing"

	"github.com/stretchr/testify/require"

	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func TestParseSubRouteRuleKeepsCallerAndCallee(t *testing.T) {
	wrapper := parseSubRouteRule(&apitraffic.CustomRoute{
		Caller: &apitraffic.CustomRoute_ServiceKey{
			Namespace: "prod",
			Service:   "checkout",
		},
		Callee: &apitraffic.CustomRoute_ServiceKey{
			Namespace: "prod",
			Service:   "payment",
		},
	})

	require.Equal(t, "prod", wrapper.Caller.Namespace)
	require.Equal(t, "checkout", wrapper.Caller.Name)
	require.Equal(t, "prod", wrapper.Callee.Namespace)
	require.Equal(t, "payment", wrapper.Callee.Name)
}
