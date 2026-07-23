package router

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/observabilityquery"
	"github.com/stretchr/testify/require"
)

func TestNewRouter_ServesObservabilityV1FromConsoleModule(t *testing.T) {
	gin.SetMode(gin.TestMode)

	webPath := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(webPath, "assets"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(webPath, "index.html"), []byte("<html></html>"), 0o644))

	greptime := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/sql", r.URL.Path)
		require.Equal(t, "public", r.URL.Query().Get("db"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		sql := string(body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(sql, "pole_control_plane_request_count_total"):
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[[1784386237437,1,"pole-control-plane","ListServices","apiserver","HTTP","success"]]}}]}`))
		case strings.Contains(sql, "pole_control_plane_request_duration_seconds_sum"):
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[[1784386237437,0.02,"pole-control-plane","ListServices","apiserver","HTTP","success"]]}}]}`))
		case strings.Contains(sql, "pole_control_plane_request_duration_seconds_count"):
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[[1784386237437,1,"pole-control-plane","ListServices","apiserver","HTTP","success"]]}}]}`))
		case strings.Contains(sql, "pole_events"):
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[["2026-07-19 09:05:00","INFO","default/checkout/10.0.0.1:8080","{\"pole.event.kind\":\"service\",\"event.name\":\"InstanceOffline\",\"pole.discovery.event\":\"InstanceOffline\",\"pole.namespace\":\"default\",\"pole.service.name\":\"checkout\",\"pole.resource.name\":\"default/checkout/10.0.0.1:8080\",\"pole.server.address\":\"node-0\"}","{}"]]}}]}`))
		default:
			_, _ = w.Write([]byte(`{"output":[{"records":{"rows":[]}}]}`))
		}
	}))
	t.Cleanup(greptime.Close)

	router := NewRouter(&bootstrap.Config{
		WebServer: bootstrap.WebServer{
			WebPath:    webPath + string(os.PathSeparator),
			CoreURL:    "/core/v1",
			NamingURL:  "/naming/v1",
			AuthURL:    "/auth/v1",
			ConfigURL:  "/config/v1",
			MonitorURL: "/api/v1",
		},
		ObservabilityQuery: observabilityquery.Config{
			Provider: "greptimedb",
			Endpoint: greptime.URL,
			Database: "public",
			Timeout:  "10s",
		},
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/observability/v1/platform/overview?category=control-plane", nil)
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)

	var body struct {
		Code int `json:"code"`
		Data struct {
			Provider struct {
				Provider   string `json:"provider"`
				Endpoint   string `json:"endpoint"`
				Database   string `json:"database"`
				Configured bool   `json:"configured"`
			} `json:"provider"`
			Components []struct {
				Component string  `json:"component"`
				API       string  `json:"api"`
				QPS       float64 `json:"qps"`
			} `json:"components"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 200000, body.Code)
	require.Equal(t, "greptimedb", body.Data.Provider.Provider)
	require.Equal(t, greptime.URL, body.Data.Provider.Endpoint)
	require.Equal(t, "public", body.Data.Provider.Database)
	require.True(t, body.Data.Provider.Configured)
	require.Len(t, body.Data.Components, 1)
	require.Equal(t, "pole-control-plane", body.Data.Components[0].Component)
	require.Equal(t, "ListServices", body.Data.Components[0].API)
	require.Equal(t, 1.0, body.Data.Components[0].QPS)

	eventRecorder := httptest.NewRecorder()
	eventReq := httptest.NewRequest(http.MethodGet, "/observability/v1/events?namespace=default&service=checkout&event_type=InstanceOffline", nil)
	router.ServeHTTP(eventRecorder, eventReq)

	require.Equal(t, http.StatusOK, eventRecorder.Code)
	var eventBody struct {
		Code uint32 `json:"code"`
		Data []struct {
			EventType string `json:"event_type"`
			Namespace string `json:"namespace"`
			Service   string `json:"service"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(eventRecorder.Body.Bytes(), &eventBody))
	require.Equal(t, uint32(200000), eventBody.Code)
	require.Len(t, eventBody.Data, 1)
	require.Equal(t, "InstanceOffline", eventBody.Data[0].EventType)
	require.Equal(t, "default", eventBody.Data[0].Namespace)
	require.Equal(t, "checkout", eventBody.Data[0].Service)
}
