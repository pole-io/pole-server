package observabilityquery

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProviderPlatformOverviewQueriesGreptimeDB(t *testing.T) {
	queries := make([]string, 0, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/sql", r.URL.Path)
		require.Equal(t, "public", r.URL.Query().Get("db"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		queries = append(queries, string(body))

		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(string(body), "pole_control_plane_request_count_total"):
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[[1784386237437,3,"pole-control-plane","ListServices","apiserver","HTTP","success"],[1784386237436,1,"pole-control-plane","ListServices","apiserver","HTTP","failure"]]}}]}`))
		case strings.Contains(string(body), "pole_control_plane_request_duration_seconds_sum"):
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[[1784386237437,0.12,"pole-control-plane","ListServices","apiserver","HTTP","success"]]}}]}`))
		case strings.Contains(string(body), "pole_control_plane_request_duration_seconds_count"):
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[[1784386237437,4,"pole-control-plane","ListServices","apiserver","HTTP","success"]]}}]}`))
		case strings.Contains(string(body), "process_runtime_go_goroutines"):
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[[1784386237436,128],[1784386237437,144]]}}]}`))
		case strings.Contains(string(body), "process_runtime_go_mem_heap_alloc_bytes"):
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[[1784386237436,10485760],[1784386237437,12582912]]}}]}`))
		case strings.Contains(string(body), "process_runtime_go_gc_count_total"):
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[[1784386237436,5],[1784386237437,8]]}}]}`))
		case strings.Contains(string(body), "process_runtime_go_gc_pause_ns_sum"):
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[[1784386237436,2000000],[1784386237437,8000000]]}}]}`))
		case strings.Contains(string(body), "process_runtime_go_gc_pause_ns_count"):
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[[1784386237436,1],[1784386237437,3]]}}]}`))
		default:
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[]}}]}`))
		}
	}))
	t.Cleanup(server.Close)

	provider := NewProvider(Config{Endpoint: server.URL})
	overview, err := provider.PlatformOverview(context.Background(), PlatformOverviewQuery{
		Category: "control-plane",
		API:      "ListServices",
	})
	require.NoError(t, err)

	require.True(t, overview.Provider.Configured)
	require.Equal(t, "public", overview.Provider.Database)
	require.Len(t, overview.Components, 1)
	component := overview.Components[0]
	require.Equal(t, "pole-control-plane", component.Component)
	require.Equal(t, "ListServices", component.API)
	require.Equal(t, "control-plane", component.Category)
	require.Equal(t, 4.0, component.QPS)
	require.Equal(t, 25.0, component.ErrorRate)
	require.Equal(t, 30.0, component.P95)
	require.Equal(t, 30.0, component.P99)
	require.NotEmpty(t, component.Series)
	require.NotEmpty(t, overview.Runtime)
	require.Equal(t, "process.runtime.go.goroutines", overview.Runtime[0].Name)
	require.Equal(t, 144.0, overview.Runtime[0].Value)
	require.Empty(t, overview.Resources)
	require.Len(t, overview.Stats, 4)
	require.Len(t, overview.Series, 1)
	require.GreaterOrEqual(t, len(queries), 8)
	require.Contains(t, queries[0], "pole_api_name+%3D+%27ListServices%27")
}

func TestProviderPlatformOverviewReturnsEmptyWhenMetricTableMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":[{"error":"Table not found: pole_control_plane_request_count_total"}]}`))
	}))
	t.Cleanup(server.Close)

	provider := NewProvider(Config{Endpoint: server.URL})
	overview, err := provider.PlatformOverview(context.Background(), PlatformOverviewQuery{})
	require.NoError(t, err)
	require.True(t, overview.Provider.Configured)
	require.Empty(t, overview.Components)
	require.Empty(t, overview.Stats)
}

func TestProviderServiceEventsQueriesPoleEvents(t *testing.T) {
	var queryBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/sql", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		queryBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[
			["2026-07-19 09:05:00","INFO","default/checkout/10.0.0.1:8080","{\"pole.event.kind\":\"service\",\"event.name\":\"InstanceOffline\",\"pole.discovery.event\":\"InstanceOffline\",\"pole.namespace\":\"default\",\"pole.service.name\":\"checkout\",\"pole.resource.name\":\"default/checkout/10.0.0.1:8080\",\"pole.server.address\":\"node-0\"}","{\"service.name\":\"pole-control-plane\"}"],
			["2026-07-19 09:04:00","INFO","audit body","{\"pole.event.kind\":\"audit\",\"event.name\":\"pole.audit.operation\"}","{}"]
		]}}]}`))
	}))
	t.Cleanup(server.Close)

	provider := NewProvider(Config{Endpoint: server.URL})
	events, err := provider.ServiceEvents(context.Background(), EventLogQuery{
		Namespace: "default",
		Service:   "checkout",
		EventType: "InstanceOffline",
		Limit:     "10",
	})
	require.NoError(t, err)
	require.Contains(t, queryBody, "pole_events")
	require.Len(t, events.Data, 1)
	require.Equal(t, "InstanceOffline", events.Data[0].EventType)
	require.Equal(t, "default", events.Data[0].Namespace)
	require.Equal(t, "checkout", events.Data[0].Service)
	require.Equal(t, "node-0", events.Data[0].Server)
	require.Equal(t, uint32(1), events.Size)
}

func TestProviderOperationsQueriesPoleEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/sql", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[
			["2026-07-19 09:07:00","INFO","release route rule","{\"pole.event.kind\":\"audit\",\"event.name\":\"pole.audit.operation\",\"pole.audit.operation\":\"Update\",\"pole.audit.operator\":\"admin\",\"pole.namespace\":\"mall-prod\",\"pole.resource.type\":\"Routing\",\"pole.resource.name\":\"checkout-gray-route\",\"pole.server.address\":\"node-1\"}","{}"],
			["2026-07-19 09:06:00","INFO","default/checkout","{\"pole.event.kind\":\"service\",\"event.name\":\"InstanceOnline\"}","{}"]
		]}}]}`))
	}))
	t.Cleanup(server.Close)

	provider := NewProvider(Config{Endpoint: server.URL})
	operations, err := provider.Operations(context.Background(), OperationLogQuery{
		ResourceType:  "Routing",
		OperationType: "Update",
		Operator:      "admin",
		Limit:         "10",
	})
	require.NoError(t, err)
	require.Len(t, operations.Data, 1)
	require.Equal(t, "Routing", operations.Data[0].ResourceType)
	require.Equal(t, "checkout-gray-route", operations.Data[0].ResourceName)
	require.Equal(t, "Update", operations.Data[0].OperationType)
	require.Equal(t, "release route rule", operations.Data[0].OperationDetail)
	require.Equal(t, "node-1", operations.Data[0].Server)
	require.Equal(t, uint32(1), operations.Size)
}
