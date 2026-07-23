package otel

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pole-io/pole-server/pkg/common/localdb"
	"github.com/stretchr/testify/require"
)

func TestLogExporterFlushesOTLPHTTPPayload(t *testing.T) {
	payloadCh := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/logs", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		payloadCh <- payload
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	exporter, err := NewLogExporter(&Config{
		LogsEndpoint: server.URL + "/v1/logs",
		ServiceName:  "pole-test",
		Environment:  "test",
		Cluster:      "local",
		NodeRole:     "control-plane",
		Timeout:      time.Second,
		LogQueueSize: 4,
	})
	require.NoError(t, err)

	exporter.Emit(LogRecord{
		Timestamp:    time.Unix(1, 2),
		SeverityText: "INFO",
		Body:         "operation detail",
		Attributes: map[string]string{
			"event.name":      "pole.audit.operation",
			"pole.event.kind": "audit",
		},
	})
	require.NoError(t, exporter.Shutdown(context.Background()))

	select {
	case payload := <-payloadCh:
		resourceLogs := payload["resourceLogs"].([]any)
		require.Len(t, resourceLogs, 1)
		firstResourceLog := resourceLogs[0].(map[string]any)
		scopeLogs := firstResourceLog["scopeLogs"].([]any)
		logRecords := scopeLogs[0].(map[string]any)["logRecords"].([]any)
		record := logRecords[0].(map[string]any)
		require.Equal(t, "operation detail", record["body"].(map[string]any)["stringValue"])
		require.Equal(t, "INFO", record["severityText"])
	case <-time.After(time.Second):
		t.Fatal("otel log payload was not flushed")
	}
}

func TestNormalizeLogsEndpointDerivesHTTPReceiverFromGRPCReceiver(t *testing.T) {
	require.Equal(t, "http://127.0.0.1:4318/v1/logs", normalizeLogsEndpoint("", "127.0.0.1:4317"))
	require.Equal(t, "http://collector:4318/v1/logs", normalizeLogsEndpoint("", "http://collector:4318"))
	require.Equal(t, "http://custom/v1/logs", normalizeLogsEndpoint("http://custom/v1/logs", "127.0.0.1:4317"))
}

func TestLogExporterSpoolsAndReplaysAfterCollectorRecovery(t *testing.T) {
	var requests atomic.Int64
	payloadCh := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		if requests.Add(1) == 1 {
			http.Error(w, "collector unavailable", http.StatusServiceUnavailable)
			return
		}
		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		payloadCh <- payload
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	spoolDir := t.TempDir()
	exporter, err := NewLogExporter(&Config{
		LogsEndpoint:     server.URL,
		ServiceName:      "pole-test",
		Environment:      "test",
		Cluster:          "local",
		NodeRole:         "control-plane",
		Timeout:          time.Second,
		LogSpoolDir:      spoolDir,
		LogSpoolMaxBytes: 1024 * 1024,
	})
	require.NoError(t, err)
	exporter.Emit(LogRecord{
		Timestamp:    time.Unix(1, 0),
		SeverityText: "INFO",
		Body:         "spooled event",
		Attributes: map[string]string{
			"event.name":      "pole.audit.operation",
			"pole.event.kind": "audit",
		},
	})

	select {
	case payload := <-payloadCh:
		record := payload["resourceLogs"].([]any)[0].(map[string]any)["scopeLogs"].([]any)[0].(map[string]any)["logRecords"].([]any)[0].(map[string]any)
		require.Equal(t, "spooled event", record["body"].(map[string]any)["stringValue"])
	case <-time.After(3 * time.Second):
		t.Fatal("spooled otel log payload was not replayed")
	}
	require.NoError(t, exporter.Shutdown(context.Background()))
	require.GreaterOrEqual(t, requests.Load(), int64(2))

	db, err := localdb.OpenPebble(spoolDir + "/otel-events.pebble")
	require.NoError(t, err)
	defer db.Close()
	count := 0
	require.NoError(t, db.ScanPrefix([]byte("otel-queue/otel_logs/"), defaultLogBatchSize, func(key, value []byte) error {
		count++
		return nil
	}))
	require.Equal(t, 0, count)
}

func TestLogExporterSharesPebbleQueueWithSeparatePrefixes(t *testing.T) {
	requests := make(chan map[string]any, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		requests <- payload
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	spoolDir := t.TempDir()
	first, err := NewLogExporter(&Config{
		LogsEndpoint:     server.URL,
		ServiceName:      "pole-test",
		Timeout:          time.Second,
		LogSpoolDir:      spoolDir,
		LogSpoolBucket:   "history",
		LogSpoolMaxBytes: 1024 * 1024,
	})
	require.NoError(t, err)
	second, err := NewLogExporter(&Config{
		LogsEndpoint:     server.URL,
		ServiceName:      "pole-test",
		Timeout:          time.Second,
		LogSpoolDir:      spoolDir,
		LogSpoolBucket:   "discover_event",
		LogSpoolMaxBytes: 1024 * 1024,
	})
	require.NoError(t, err)

	first.Emit(LogRecord{Body: "history event", Attributes: map[string]string{"pole.event.kind": "audit"}})
	second.Emit(LogRecord{Body: "discover event", Attributes: map[string]string{"pole.event.kind": "service"}})

	seen := map[string]struct{}{}
	for len(seen) < 2 {
		select {
		case payload := <-requests:
			record := payload["resourceLogs"].([]any)[0].(map[string]any)["scopeLogs"].([]any)[0].(map[string]any)["logRecords"].([]any)[0].(map[string]any)
			seen[record["body"].(map[string]any)["stringValue"].(string)] = struct{}{}
		case <-time.After(3 * time.Second):
			t.Fatal("shared pebble spool payloads were not replayed")
		}
	}
	require.Contains(t, seen, "history event")
	require.Contains(t, seen, "discover event")
	require.NoError(t, first.Shutdown(context.Background()))
	require.NoError(t, second.Shutdown(context.Background()))
}
