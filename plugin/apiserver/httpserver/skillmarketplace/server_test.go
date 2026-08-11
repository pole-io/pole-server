package skillmarketplace

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/require"
)

func TestAccessServerRegistersConsoleAndCLIRoutes(t *testing.T) {
	ws := (&HTTPServer{}).GetAccessServer()
	routes := map[string]bool{}
	for _, route := range ws.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, expected := range []string{
		"GET /api/skill-marketplace/v1/skills",
		"GET /api/skill-marketplace/v1/skills/{publisher}/{name}",
		"GET /api/skill-marketplace/v1/skills/{publisher}/{name}/releases/{version}/bundle",
		"GET /api/skill-marketplace/v1/skills/{publisher}/{name}/releases/{version}/bundle-manifest",
		"POST /api/skill-marketplace/v1/skills/releases/upload",
		"POST /api/skill-marketplace/v1/skills/releases/import-git/discover",
		"POST /api/skill-marketplace/v1/skills/releases/import-git",
		"GET /api/skill-marketplace/v1/reviews",
		"PUT /api/skill-marketplace/v1/reviews/{releaseId}",
		"GET /api/skill-marketplace/v1/skills/{publisher}/{name}/grants",
		"PUT /api/skill-marketplace/v1/skills/{publisher}/{name}/grants",
		"GET /api/skill-marketplace/v1/registry-sources",
		"POST /api/skill-marketplace/v1/registry-sources/{id}/sync",
	} {
		require.Truef(t, routes[expected], "missing route %s", expected)
	}
}

func TestParseSignatureEnvelopeFile(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("signature", "release.sig.json")
	require.NoError(t, err)
	signature := bytes.Repeat([]byte{7}, 64)
	signedAt := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	_, err = fmt.Fprintf(part, `{"algorithm":"Ed25519","signedAt":%q,"signature":%q}`,
		signedAt.Format(time.RFC3339), base64.StdEncoding.EncodeToString(signature))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	httpRequest := httptest.NewRequest(http.MethodPost, "/", &body)
	httpRequest.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, httpRequest.ParseMultipartForm(1<<20))
	parsed, parsedAt, err := parseSignatureEnvelope(restful.NewRequest(httpRequest))
	require.NoError(t, err)
	require.Equal(t, signature, parsed)
	require.Equal(t, signedAt, parsedAt)
}

func TestRegistryIndexCursorContinuesPastFirstHundredSkills(t *testing.T) {
	require.Equal(t, "100", nextIndexCursor(0, 100, 101))
	require.Empty(t, nextIndexCursor(100, 1, 101))
}
