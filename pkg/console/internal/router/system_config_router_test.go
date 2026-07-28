package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"

	bootstrap "github.com/pole-io/pole-server/pkg/console/config"
	"github.com/pole-io/pole-server/pkg/systemconfig"
)

func TestSystemConfigurationRouterCombinesRuntimeSnapshotsWithoutLeakingSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/admin/v1/system/configuration", r.URL.Path)
		require.Equal(t, "api-user", r.Header.Get("X-Pole-User"))
		require.Equal(t, "api-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"component\":\"pole-server\",\"settings\":[{\"key\":\"server.store.dsn\",\"component\":\"pole-server\",\"domain\":\"storage\",\"label\":\"数据库连接\",\"value_type\":\"secret\",\"apply_mode\":\"BootstrapOnly\",\"sensitivity\":\"secret\",\"owner\":\"pole-server\",\"display_value\":\"已配置\",\"configured\":true,\"redacted\":true,\"source\":{\"kind\":\"environment\",\"reference\":\"env:MYSQL_DSN\"}}]}"))
	}))
	defer backend.Close()

	cfg := &bootstrap.Config{
		WebServer: bootstrap.WebServer{
			CoreURL: "/core/v1", NamingURL: "/naming/v1",
			AuthURL: "/auth/v1", ConfigURL: "/config/v1", MonitorURL: "/api/v1",
			Mode: "release", JWT: bootstrap.JWT{SecretKey: "must-not-leak-console-secret", Expired: 1800},
		},
		PoleServer: bootstrap.PoleServer{Address: strings.TrimPrefix(backend.URL, "http://")},
		Agent: bootstrap.AgentConfig{
			Model: bootstrap.AgentModelConfig{
				BaseURL: "https://llm-gateway.example.test",
				APIKey:  "must-not-leak-agent-secret",
			},
		},
	}
	cfg.SystemConfigSources = systemconfig.SourceIndex{
		"webServer.jwt.secretKey": {Kind: systemconfig.SourceEnvironment, Reference: "env:CONSOLE_JWT_SECRET"},
		"agent.model.apiKey":      {Kind: systemconfig.SourceEnvironment, Reference: "env:POLE_AGENT_LLM_API_KEY"},
	}
	router := NewRouter(cfg, testAssets("<html>system config</html>"))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/system-config/v1/settings", nil)
	req.Header.Set("X-Pole-User", "api-user")
	req.AddCookie(newSystemConfigTestJWTCookie(t, "api-user", "api-token", "main", cfg.WebServer.JWT.SecretKey))
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.NotContains(t, recorder.Body.String(), "must-not-leak-console-secret")
	require.NotContains(t, recorder.Body.String(), "must-not-leak-agent-secret")
	var body struct {
		Code int `json:"code"`
		Data struct {
			Settings []systemconfig.EffectiveSetting `json:"settings"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 200000, body.Code)
	require.Greater(t, len(body.Data.Settings), 2)
	keys := map[string]bool{}
	for _, setting := range body.Data.Settings {
		keys[setting.Key] = true
	}
	require.True(t, keys["server.store.dsn"])
	require.True(t, keys["console.web.jwt.secret_key"])
	require.True(t, keys["console.agent.model_base_url"])
	require.True(t, keys["console.agent.model_api_key"])
}

func TestUnknownSystemConfigurationRouteReturnsJSON404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newAgentTestRouter(t, "127.0.0.1:1", "secret")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/system-config/v1/unknown", nil))
	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.Contains(t, recorder.Header().Get("Content-Type"), "application/json")
	require.NotContains(t, recorder.Body.String(), "<html")
}

func TestSystemConfigurationRouterRejectsAnonymousRequestWith401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newAgentTestRouter(t, "127.0.0.1:1", "secret")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/system-config/v1/settings", nil))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Contains(t, recorder.Header().Get("Content-Type"), "application/json")
}

func TestSystemConfigurationRouterRejectsNonAdminWith403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newAgentTestRouter(t, "127.0.0.1:1", "secret")
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/system-config/v1/settings", nil)
	req.Header.Set("X-Pole-User", "sub-user")
	req.AddCookie(newSystemConfigTestJWTCookie(t, "sub-user", "token", "sub", "secret"))
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Contains(t, recorder.Body.String(), "requires admin role")
}

func newSystemConfigTestJWTCookie(t *testing.T, userID, token, role, secret string) *http.Cookie {
	t.Helper()
	now := time.Now()
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"UserID": userID,
		"Token":  token,
		"Role":   role,
		"exp":    now.Add(time.Hour).Unix(),
		"nbf":    now.Unix(),
		"iat":    now.Unix(),
	}).SignedString([]byte(secret))
	require.NoError(t, err)
	return &http.Cookie{Name: "jwt", Value: signed}
}
