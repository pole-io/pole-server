package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pole-io/pole-server/console/bootstrap"
	"github.com/stretchr/testify/require"
)

func TestNewRouter_ServesLatestSPAIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)

	webPath := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(webPath, "assets"), 0o755))
	indexPath := filepath.Join(webPath, "index.html")
	require.NoError(t, os.WriteFile(indexPath, []byte("<html>first</html>"), 0o644))

	router := NewRouter(&bootstrap.Config{
		WebServer: bootstrap.WebServer{
			WebPath:    webPath + string(os.PathSeparator),
			CoreURL:    "/core/v1",
			NamingURL:  "/naming/v1",
			AuthURL:    "/auth/v1",
			ConfigURL:  "/config/v1",
			MonitorURL: "/api/v1",
		},
	})

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, first.Code)
	require.Contains(t, first.Body.String(), "first")
	require.Equal(t, "no-cache, no-store, must-revalidate", first.Header().Get("Cache-Control"))
	require.Equal(t, "no-cache", first.Header().Get("Pragma"))
	require.Equal(t, "0", first.Header().Get("Expires"))

	require.NoError(t, os.WriteFile(indexPath, []byte("<html>second</html>"), 0o644))
	deepLink := httptest.NewRecorder()
	router.ServeHTTP(deepLink, httptest.NewRequest(http.MethodGet, "/namespace", nil))
	require.Equal(t, http.StatusOK, deepLink.Code)
	require.Contains(t, deepLink.Body.String(), "second")
	require.Equal(t, "no-cache, no-store, must-revalidate", deepLink.Header().Get("Cache-Control"))
}
