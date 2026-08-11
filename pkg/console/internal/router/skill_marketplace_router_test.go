package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	bootstrap "github.com/pole-io/pole-server/pkg/console/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouterProxiesSkillMarketplaceRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const (
		userID = "user1"
		token  = "token1"
		secret = "polarismesh@2021"
	)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/v1/user/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":200000,"data":{"auth_token":"token1","token_enable":true}}`))
			return
		}
		assert.Equal(t, "/api/skill-marketplace/v1/skills", r.URL.Path)
		assert.Equal(t, token, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"total":0}`))
	}))
	defer backend.Close()

	router := NewRouter(&bootstrap.Config{
		WebServer: bootstrap.WebServer{
			CoreURL: "/core/v1", NamingURL: "/naming/v1", AuthURL: "/auth/v1",
			ConfigURL: "/config/v1", MonitorURL: "/api/v1",
			JWT: bootstrap.JWT{SecretKey: secret, Expired: 1800},
		},
		PoleServer: bootstrap.PoleServer{Address: strings.TrimPrefix(backend.URL, "http://")},
	}, testAssets("<html></html>"))

	console := httptest.NewServer(router)
	defer console.Close()
	req, err := http.NewRequest(http.MethodGet, console.URL+"/api/skill-marketplace/v1/skills", nil)
	require.NoError(t, err)
	req.AddCookie(newTestJWTCookie(t, userID, token, secret))
	req.Header.Set("x-pole-user", userID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"items":[],"total":0}`, string(body))
}

func TestSkillMarketplaceUnknownRouteDoesNotFallBackToSPA(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(&bootstrap.Config{WebServer: bootstrap.WebServer{
		CoreURL: "/core/v1", NamingURL: "/naming/v1", AuthURL: "/auth/v1",
		ConfigURL: "/config/v1", MonitorURL: "/api/v1",
	}}, testAssets("<html>embedded</html>"))

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/skill-marketplace", nil))
	require.NotContains(t, response.Body.String(), "embedded")
}
