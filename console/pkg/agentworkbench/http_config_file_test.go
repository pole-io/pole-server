package agentworkbench

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPConfigFilePortReadsAndUpdatesDraftWithoutRelease(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		assert.Equal(t, "pole-token", r.Header.Get("Authorization"))
		assert.Equal(t, "alice", r.Header.Get("X-Pole-User"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "upstream-request")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/config/v1/files/detail":
			assert.Equal(t, "default", r.URL.Query().Get("namespace"))
			assert.Equal(t, "orders", r.URL.Query().Get("group"))
			assert.Equal(t, "application.yaml", r.URL.Query().Get("name"))
			_, _ = w.Write([]byte(`{"code":200000,"info":"execute success","data":{"@type":"type.googleapis.com/v1.ConfigFile","id":"10","name":"application.yaml","namespace":"default","group":"orders","content":"timeout: 3s\n","format":"yaml","comment":"current","labels":{"env":"test"},"encrypted":false,"encryptAlgo":""}}`))
		case r.Method == http.MethodPut && r.URL.Path == "/config/v1/files":
			var payload []map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			require.Len(t, payload, 1)
			assert.Equal(t, "timeout: 5s\n", payload[0]["content"])
			_, _ = w.Write([]byte(`{"code":200000,"info":"execute success","size":1,"responses":[{"code":200000,"info":"execute success"}]}`))
		default:
			t.Fatalf("unexpected upstream request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	port := NewHTTPConfigFilePort(strings.TrimPrefix(server.URL, "http://"), server.Client())
	actor := Actor{UserID: "alice", Token: "pole-token", RequestID: "console-request"}
	file, requestID, err := port.GetConfigFile(context.Background(), actor, ResourceRef{
		Kind: "config.file", Namespace: "default", Group: "orders", Name: "application.yaml",
	})
	require.NoError(t, err)
	require.Equal(t, "upstream-request", requestID)
	require.Equal(t, "timeout: 3s\n", file.Content)
	require.Equal(t, map[string]string{"env": "test"}, file.Labels)

	file.Content = "timeout: 5s\n"
	requestID, err = port.UpdateConfigFile(context.Background(), actor, file)
	require.NoError(t, err)
	require.Equal(t, "upstream-request", requestID)
	require.Equal(t, []string{"GET /config/v1/files/detail", "PUT /config/v1/files"}, paths)
	for _, path := range paths {
		assert.NotContains(t, path, "release")
	}
}

func TestHTTPConfigFilePortPreservesUpstreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-Id", "request-42")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":401001,"info":"not allowed access"}`))
	}))
	defer server.Close()

	port := NewHTTPConfigFilePort(strings.TrimPrefix(server.URL, "http://"), server.Client())
	_, _, err := port.GetConfigFile(context.Background(), Actor{UserID: "alice", Token: "token"}, ResourceRef{
		Kind: "config.file", Namespace: "default", Group: "orders", Name: "application.yaml",
	})
	var upstream *UpstreamError
	require.ErrorAs(t, err, &upstream)
	require.Equal(t, uint32(401001), upstream.Code)
	require.Equal(t, "not allowed access", upstream.Info)
	require.Equal(t, "request-42", upstream.RequestID)
}

func TestHTTPConfigFilePortClassifiesMalformedResponseAsDownstreamFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-Id", "request-43")
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer server.Close()

	port := NewHTTPConfigFilePort(strings.TrimPrefix(server.URL, "http://"), server.Client())
	_, _, err := port.GetConfigFile(context.Background(), Actor{UserID: "alice", Token: "token"}, ResourceRef{
		Kind: "config.file", Namespace: "default", Group: "orders", Name: "application.yaml",
	})
	var upstream *UpstreamError
	require.ErrorAs(t, err, &upstream)
	require.Equal(t, uint32(502001), upstream.Code)
	require.Equal(t, http.StatusBadGateway, upstream.Status)
	require.Equal(t, "request-43", upstream.RequestID)
}

func TestHTTPConfigFilePortRejectsEmptyBatchResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-Id", "request-44")
		_, _ = w.Write([]byte(`{"code":200000,"info":"execute success","responses":[]}`))
	}))
	defer server.Close()

	port := NewHTTPConfigFilePort(strings.TrimPrefix(server.URL, "http://"), server.Client())
	_, err := port.UpdateConfigFile(context.Background(), Actor{UserID: "alice", Token: "token"}, ConfigFile{
		Namespace: "default", Group: "orders", Name: "application.yaml", Content: "a: 2\n",
	})
	var upstream *UpstreamError
	require.ErrorAs(t, err, &upstream)
	require.Equal(t, uint32(502001), upstream.Code)
	require.Equal(t, http.StatusBadGateway, upstream.Status)
}
