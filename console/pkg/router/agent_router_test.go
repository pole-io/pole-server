package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/console/bootstrap"
)

func TestAgentRouterPreviewsAndStagesConfigDraftWithoutPublishing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const (
		userID = "alice"
		token  = "pole-token"
		secret = "agent-test-secret"
	)
	var mu sync.Mutex
	content := "timeout: 3s\n"
	putCalls := 0
	var upstreamPaths []string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		upstreamPaths = append(upstreamPaths, r.Method+" "+r.URL.Path)
		assert.Equal(t, token, r.Header.Get("Authorization"))
		assert.Equal(t, userID, r.Header.Get("X-Pole-User"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "upstream-request")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/config/v1/files/detail":
			_, _ = w.Write([]byte(`{"code":200000,"info":"execute success","data":{"@type":"type.googleapis.com/v1.ConfigFile","id":"10","name":"application.yaml","namespace":"default","group":"orders","content":` + mustJSON(t, content) + `,"format":"yaml","comment":"current","labels":{},"encrypted":false}}`))
		case r.Method == http.MethodPut && r.URL.Path == "/config/v1/files":
			var payload []map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			content = payload[0]["content"].(string)
			putCalls++
			_, _ = w.Write([]byte(`{"code":200000,"info":"execute success","responses":[{"code":200000,"info":"execute success"}]}`))
		default:
			t.Fatalf("unexpected upstream request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer backend.Close()

	router := newAgentTestRouter(t, backend.URL, secret)
	console := httptest.NewServer(router)
	defer console.Close()

	prepareBody := `{"namespace":"default","group":"orders","name":"application.yaml","desiredContent":"timeout: 5s\n","comment":"agent draft"}`
	prepareReq, err := http.NewRequest(http.MethodPost, console.URL+"/ai/agent/v1/proposals/config-file", strings.NewReader(prepareBody))
	require.NoError(t, err)
	prepareReq.Header.Set("Content-Type", "application/json")
	prepareReq.Header.Set("X-Pole-User", userID)
	prepareReq.Header.Set("Authorization", "untrusted-browser-token")
	prepareReq.Header.Set("X-Request-Id", "console-request")
	prepareReq.AddCookie(newTestJWTCookie(t, userID, token, secret))
	prepareResp, err := http.DefaultClient.Do(prepareReq)
	require.NoError(t, err)
	defer prepareResp.Body.Close()
	require.Equal(t, http.StatusOK, prepareResp.StatusCode)
	var prepared struct {
		Data struct {
			ID          string `json:"id"`
			Version     uint64 `json:"version"`
			PreviewHash string `json:"previewHash"`
			Status      string `json:"status"`
			After       struct {
				Content string `json:"content"`
			} `json:"after"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(prepareResp.Body).Decode(&prepared))
	require.Equal(t, "preview_ready", prepared.Data.Status)
	require.Equal(t, "timeout: 5s\n", prepared.Data.After.Content)
	require.NotEmpty(t, prepared.Data.PreviewHash)
	require.Zero(t, putCalls)

	confirmBody := `{"proposalVersion":` + fmt.Sprint(prepared.Data.Version) + `,"previewHash":` + mustJSON(t, prepared.Data.PreviewHash) + `,"idempotencyKey":"confirm-1"}`
	confirmURL := console.URL + "/ai/agent/v1/proposals/" + prepared.Data.ID + "/confirm"
	firstReceipt := confirmAgentProposal(t, confirmURL, userID, token, secret, confirmBody)
	secondReceipt := confirmAgentProposal(t, confirmURL, userID, token, secret, confirmBody)
	require.Equal(t, firstReceipt, secondReceipt)
	require.Equal(t, "waiting_for_publish", firstReceipt.Status)
	require.Equal(t, "/configuration/group/files?namespace=default&group=orders", firstReceipt.DetailURL)
	require.Equal(t, "upstream-request", firstReceipt.RequestID)
	require.Equal(t, 1, putCalls)
	require.Equal(t, "timeout: 5s\n", content)
	for _, path := range upstreamPaths {
		assert.NotContains(t, path, "release")
	}
}

func TestAgentUnknownAPIRouteReturnsJSON404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	backend := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer backend.Close()
	router := newAgentTestRouter(t, backend.URL, "secret")
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ai/agent/v1/unknown", nil)
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.Contains(t, recorder.Header().Get("Content-Type"), "application/json")
	require.NotContains(t, recorder.Body.String(), "<html")
}

func TestAgentRouterAcceptsValidatedHeaderCredentialsForAPIClients(t *testing.T) {
	gin.SetMode(gin.TestMode)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "api-user", r.Header.Get("X-Pole-User"))
		require.Equal(t, "api-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200000,"info":"execute success","data":{"name":"app.yaml","namespace":"default","group":"orders","content":"before","encrypted":false}}`))
	}))
	defer backend.Close()

	router := newAgentTestRouter(t, backend.URL, "secret")
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ai/agent/v1/proposals/config-file", strings.NewReader(
		`{"namespace":"default","group":"orders","name":"app.yaml","desiredContent":"after"}`,
	))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pole-User", "api-user")
	req.Header.Set("Authorization", "api-token")
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"status":"preview_ready"`)
}

func TestAgentRuntimeAndTurnFailClosedWhenLLMGatewayIsNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newAgentTestRouter(t, "127.0.0.1:1", "secret")

	runtimeRecorder := httptest.NewRecorder()
	runtimeRequest := httptest.NewRequest(http.MethodGet, "/ai/agent/v1/runtime", nil)
	runtimeRequest.Header.Set("X-Pole-User", "admin")
	runtimeRequest.AddCookie(newTestJWTCookie(t, "admin", "pole-token", "secret"))
	router.ServeHTTP(runtimeRecorder, runtimeRequest)
	require.Equal(t, http.StatusOK, runtimeRecorder.Code, runtimeRecorder.Body.String())
	require.Contains(t, runtimeRecorder.Body.String(), `"configured":false`)
	require.Contains(t, runtimeRecorder.Body.String(), `"mcpConnected":false`)
	require.Contains(t, runtimeRecorder.Body.String(), `"mcpTools":[]`)
	require.Contains(t, runtimeRecorder.Body.String(), `"tools":[]`)
	require.NotContains(t, runtimeRecorder.Body.String(), "apiKey")
	require.NotContains(t, runtimeRecorder.Body.String(), "baseURL")

	turnRecorder := httptest.NewRecorder()
	turnRequest := httptest.NewRequest(http.MethodPost, "/ai/agent/v1/turns", strings.NewReader(
		`{"sessionId":"local-session","message":"列出命名空间","history":[]}`,
	))
	turnRequest.Header.Set("Content-Type", "application/json")
	turnRequest.Header.Set("X-Pole-User", "admin")
	turnRequest.AddCookie(newTestJWTCookie(t, "admin", "pole-token", "secret"))
	router.ServeHTTP(turnRecorder, turnRequest)
	require.Equal(t, http.StatusServiceUnavailable, turnRecorder.Code, turnRecorder.Body.String())
	require.Contains(t, turnRecorder.Body.String(), `"category":"RUNTIME_UNAVAILABLE"`)
}

type agentReceipt struct {
	ProposalID string `json:"proposalId"`
	Status     string `json:"status"`
	RequestID  string `json:"requestId"`
	DetailURL  string `json:"detailUrl"`
}

func confirmAgentProposal(t *testing.T, target, userID, token, secret, body string) agentReceipt {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, target, bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pole-User", userID)
	req.Header.Set("X-Request-Id", "confirm-request")
	req.AddCookie(newTestJWTCookie(t, userID, token, secret))
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(bodyBytes))
	var envelope struct {
		Data agentReceipt `json:"data"`
	}
	require.NoError(t, json.Unmarshal(bodyBytes, &envelope))
	return envelope.Data
}

func newAgentTestRouter(t *testing.T, backendURL, secret string) *gin.Engine {
	t.Helper()
	webPath := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(webPath, "assets"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(webPath, "index.html"), []byte("<html>agent test</html>"), 0o644))
	return NewRouter(&bootstrap.Config{
		WebServer: bootstrap.WebServer{
			WebPath: webPath + string(os.PathSeparator), CoreURL: "/core/v1", NamingURL: "/naming/v1",
			AuthURL: "/auth/v1", ConfigURL: "/config/v1", MonitorURL: "/api/v1",
			JWT: bootstrap.JWT{SecretKey: secret, Expired: 1800},
		},
		PoleServer: bootstrap.PoleServer{Address: strings.TrimPrefix(backendURL, "http://")},
	})
}

func mustJSON(t *testing.T, value string) string {
	t.Helper()
	content, err := json.Marshal(value)
	require.NoError(t, err)
	return string(content)
}
