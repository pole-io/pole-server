package discover

import (
	"net/http"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/apis/apiserver"
)

func TestClientAccessRegistersServiceContractReportRoute(t *testing.T) {
	ws := new(restful.WebService)
	ws.Path("/naming/v1").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	(&HTTPServer{}).GetClientAccessServer(ws, []string{apiserver.DiscoverAccess})

	found := false
	for _, route := range ws.Routes() {
		if route.Method == http.MethodPost && route.Path == "/naming/v1/ReportServiceContract" {
			found = true
			break
		}
	}
	require.True(t, found)
}
