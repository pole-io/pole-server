//go:build e2e
// +build e2e

package client

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pole-io/pole-server/test/e2e/internal/e2e"
)

func TestClientAPIDiscoversConsoleWritesAndReleases(t *testing.T) {
	env := e2e.Start(t, e2e.Options{Suite: "client-api", PortBase: 30040, ConsoleOpen: true, ClientOpen: false})
	api := e2e.NewHTTPClient(t, env)
	api.LoginAsMainUser()
	names := e2e.NewNames("client")
	createClientBaseResources(t, api, names)

	t.Run("service and instance changes propagate to Discover", func(t *testing.T) {
		instanceA := e2e.Instance(names.Namespace, names.Payment, "10.1.0.10", 18080)
		instanceB := e2e.Instance(names.Namespace, names.Payment, "10.1.0.11", 18081)
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/instances", []map[string]any{instanceA, instanceB}), "create client instances")
		expectDiscoverSuccess(t, api, names.Namespace, names.Payment, "INSTANCE")
		instanceA["weight"] = 50
		instanceA["isolate"] = true
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/naming/v1/instances", []map[string]any{instanceA}), "update client instance")
		expectDiscoverSuccess(t, api, names.Namespace, names.Payment, "INSTANCE")
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/instances/delete", []map[string]any{instanceB}), "delete client instance")
		expectDiscoverSuccess(t, api, names.Namespace, names.Payment, "INSTANCE")
		expectDiscoverWithoutText(t, api, names.Namespace, names.Payment, "INSTANCE", "10.1.0.11")
	})

	t.Run("governance release state propagates to Discover", func(t *testing.T) {
		for _, tc := range clientGovernanceCases(names) {
			t.Run(tc.name, func(t *testing.T) {
				e2e.RequireSuccess(t, api.Console(http.MethodPost, tc.path, []map[string]any{tc.payload}), "create "+tc.name)
				list := api.ConsoleQuery(http.MethodGet, tc.path, e2e.Query(0, 10, tc.query), nil)
				e2e.RequireSuccess(t, list, "list "+tc.name)
				item, ok := e2e.FindByName(list.Slice(t), tc.ruleName)
				if !ok {
					t.Fatalf("%s rule %s not found", tc.name, tc.ruleName)
				}
				id := e2e.StringField(t, item, "id")
				e2e.RequireSuccess(t, api.Console(http.MethodPost, tc.path+"/releases", []map[string]any{e2e.RuleRelease(id, tc.ruleName, tc.releaseResource)}), "publish "+tc.name)
				expectDiscoverSuccess(t, api, tc.discoverNamespace, tc.discoverService, tc.discoverType)
				e2e.RequireSuccess(t, api.Console(http.MethodPost, tc.path+"/delete", []map[string]any{{"id": id}}), "delete "+tc.name)
				expectDiscoverWithoutText(t, api, tc.discoverNamespace, tc.discoverService, tc.discoverType, tc.ruleName)
			})
		}
	})
}

type clientGovernanceCase struct {
	name              string
	path              string
	ruleName          string
	releaseResource   string
	query             map[string]string
	payload           map[string]any
	discoverNamespace string
	discoverService   string
	discoverType      string
}

func clientGovernanceCases(names e2e.Names) []clientGovernanceCase {
	return []clientGovernanceCase{
		{name: "route", path: "/naming/v1/routings", ruleName: names.Prefix + "-route", releaseResource: "RouteRules", query: map[string]string{"name": names.Prefix + "-route"}, payload: e2e.RouteRule(names.Namespace, names.Prefix+"-route", names.Order, names.Payment), discoverNamespace: names.Namespace, discoverService: names.Payment, discoverType: "CUSTOM_ROUTE_RULE"},
		{name: "ratelimit", path: "/naming/v1/ratelimits", ruleName: names.Prefix + "-ratelimit", releaseResource: "RateLimitRules", query: map[string]string{"name": names.Prefix + "-ratelimit", "namespace": names.Namespace, "service": names.Payment}, payload: e2e.RateLimitRule(names.Namespace, names.Prefix+"-ratelimit", names.Payment), discoverNamespace: names.Namespace, discoverService: names.Payment, discoverType: "RATE_LIMIT"},
		{name: "circuitbreaker", path: "/naming/v1/circuitbreakers", ruleName: names.Prefix + "-circuitbreaker", releaseResource: "CircuitBreakerRules", query: map[string]string{"name": names.Prefix + "-circuitbreaker"}, payload: e2e.CircuitBreakerRule(names.Namespace, names.Prefix+"-circuitbreaker", names.Order, names.Payment), discoverNamespace: names.Namespace, discoverService: names.Payment, discoverType: "CIRCUIT_BREAKER"},
		{name: "faultdetect", path: "/naming/v1/faultdetectors", ruleName: names.Prefix + "-faultdetect", releaseResource: "FaultDetectRules", query: map[string]string{"name": names.Prefix + "-faultdetect"}, payload: e2e.FaultDetectRule(names.Namespace, names.Prefix+"-faultdetect", names.Payment), discoverNamespace: names.Namespace, discoverService: names.Payment, discoverType: "FAULT_DETECTOR"},
		{name: "lossless", path: "/naming/v1/lossless", ruleName: names.Payment, releaseResource: "LosslessRules", query: map[string]string{"namespace": names.Namespace, "service": names.Payment}, payload: e2e.LosslessRule(names.Namespace, names.Payment), discoverNamespace: names.Namespace, discoverService: names.Payment, discoverType: "SERVICES"},
		{name: "lane", path: "/naming/v1/lane/groups", ruleName: names.Prefix + "-lane", releaseResource: "LaneRules", query: map[string]string{"name": names.Prefix + "-lane"}, payload: e2e.LaneGroup(names.Namespace, names.Prefix+"-lane", names.Gateway, names.Payment), discoverNamespace: names.Namespace, discoverService: names.Gateway, discoverType: "CUSTOM_ROUTE_RULE"},
		{name: "traffic-security", path: "/naming/v1/traffic/security", ruleName: names.Prefix + "-security", releaseResource: "TrafficSecurityRules", query: map[string]string{"name": names.Prefix + "-security", "namespace": names.Namespace, "service": names.Gateway}, payload: e2e.TrafficSecurityRule(names.Namespace, names.Prefix+"-security", names.Gateway), discoverNamespace: names.Namespace, discoverService: names.Gateway, discoverType: "TRAFFIC_SECURITY_RULE"},
		{name: "traffic-mirror", path: "/naming/v1/traffic/mirrors", ruleName: names.Prefix + "-mirror", releaseResource: "TrafficMirrorRules", query: map[string]string{"name": names.Prefix + "-mirror", "namespace": names.Namespace, "service": names.Gateway}, payload: e2e.TrafficMirrorRule(names.Namespace, names.Prefix+"-mirror", names.Gateway, names.Shadow), discoverNamespace: names.Namespace, discoverService: names.Gateway, discoverType: "TRAFFIC_MIRROR_RULE"},
		{name: "traffic-mock", path: "/naming/v1/traffic/mocks", ruleName: names.Prefix + "-mock", releaseResource: "TrafficMockRules", query: map[string]string{"name": names.Prefix + "-mock", "namespace": names.Namespace, "service": names.Gateway}, payload: e2e.TrafficMockRule(names.Namespace, names.Prefix+"-mock", names.Gateway), discoverNamespace: names.Namespace, discoverService: names.Gateway, discoverType: "TRAFFIC_MOCK_RULE"},
	}
}

func createClientBaseResources(t *testing.T, api *e2e.HTTPClient, names e2e.Names) {
	t.Helper()
	e2e.RequireSuccess(t, api.Console(http.MethodPost, "/core/v1/namespaces", []map[string]any{e2e.Namespace(names.Namespace)}), "create namespace")
	e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/services", []map[string]any{
		e2e.Service(names.Namespace, names.Gateway),
		e2e.Service(names.Namespace, names.Order),
		e2e.Service(names.Namespace, names.Payment),
		e2e.Service(names.Namespace, names.Shadow),
	}), "create services")
}

func expectDiscoverSuccess(t *testing.T, api *e2e.HTTPClient, namespace, service, typ string) {
	t.Helper()
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := api.Client(http.MethodPost, "/naming/v1/Discover", e2e.DiscoverRequest(namespace, service, typ))
		return resp.Code == e2e.ExecuteSuccess || resp.Code == 200001
	}, "client discover "+typ+" "+namespace+"/"+service)
}

func expectDiscoverWithoutText(t *testing.T, api *e2e.HTTPClient, namespace, service, typ, text string) {
	t.Helper()
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := api.Client(http.MethodPost, "/naming/v1/Discover", e2e.DiscoverRequest(namespace, service, typ))
		if resp.Code != e2e.ExecuteSuccess && resp.Code != 200001 {
			return false
		}
		return !strings.Contains(string(resp.Data), text)
	}, "client discover "+typ+" no longer contains "+text)
}
