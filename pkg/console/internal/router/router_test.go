package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	bootstrap "github.com/pole-io/pole-server/pkg/console/config"
	"github.com/stretchr/testify/require"
)

func testAssets(index string) fstest.MapFS {
	return fstest.MapFS{
		"index.html":    &fstest.MapFile{Data: []byte(index)},
		"assets/app.js": &fstest.MapFile{Data: []byte("console asset")},
	}
}

func TestNewRouterServesEmbeddedSPAAndDeepLinks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(&bootstrap.Config{
		WebServer: bootstrap.WebServer{
			CoreURL:    "/core/v1",
			NamingURL:  "/naming/v1",
			AuthURL:    "/auth/v1",
			ConfigURL:  "/config/v1",
			MonitorURL: "/api/v1",
		},
	}, testAssets("<html>embedded</html>"))

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, first.Code)
	require.Contains(t, first.Body.String(), "embedded")
	require.Equal(t, "no-cache, no-store, must-revalidate", first.Header().Get("Cache-Control"))
	require.Equal(t, "no-cache", first.Header().Get("Pragma"))
	require.Equal(t, "0", first.Header().Get("Expires"))

	deepLink := httptest.NewRecorder()
	router.ServeHTTP(deepLink, httptest.NewRequest(http.MethodGet, "/namespace", nil))
	require.Equal(t, http.StatusOK, deepLink.Code)
	require.Contains(t, deepLink.Body.String(), "embedded")
	require.Equal(t, "no-cache, no-store, must-revalidate", deepLink.Header().Get("Cache-Control"))

	asset := httptest.NewRecorder()
	router.ServeHTTP(asset, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	require.Equal(t, http.StatusOK, asset.Code)
	require.Equal(t, "console asset", asset.Body.String())
}
