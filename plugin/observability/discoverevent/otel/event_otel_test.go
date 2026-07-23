package otel

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pole-io/pole-server/apis"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/stretchr/testify/require"
)

func TestDiscoverEventWorkerEmitsServiceLogRecord(t *testing.T) {
	payloadCh := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		payloadCh <- payload
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	worker := &DiscoverEventWorker{}
	require.NoError(t, worker.Initialize(&apis.ConfigEntry{
		Option: map[string]interface{}{
			"logsEndpoint": server.URL,
			"serviceName":  "pole-control-plane",
			"logQueueSize": 4,
		},
	}))
	worker.PublishEvent(&svctypes.ServiceEvent{
		Id:         "evt-1",
		Namespace:  "default",
		Service:    "checkout",
		EType:      svctypes.EventServiceOpenEmptyPushProtect,
		CreateTime: time.Unix(1, 0),
	})
	require.NoError(t, worker.Destroy())

	select {
	case payload := <-payloadCh:
		record := firstLogRecord(t, payload)
		attrs := attrMap(record["attributes"].([]any))
		require.Equal(t, "ServiceOpenEmptyPushProtect", attrs["event.name"])
		require.Equal(t, "service", attrs["pole.event.kind"])
		require.Equal(t, "default", attrs["pole.namespace"])
		require.Equal(t, "checkout", attrs["pole.service.name"])
		require.Equal(t, "evt-1", attrs["pole.resource.id"])
	case <-time.After(time.Second):
		t.Fatal("discover event otel payload was not emitted")
	}
}

func firstLogRecord(t *testing.T, payload map[string]any) map[string]any {
	t.Helper()
	resourceLogs := payload["resourceLogs"].([]any)
	scopeLogs := resourceLogs[0].(map[string]any)["scopeLogs"].([]any)
	records := scopeLogs[0].(map[string]any)["logRecords"].([]any)
	return records[0].(map[string]any)
}

func attrMap(attrs []any) map[string]string {
	out := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		item := attr.(map[string]any)
		out[item["key"].(string)] = item["value"].(map[string]any)["stringValue"].(string)
	}
	return out
}
