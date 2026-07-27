package discover

import (
	"net/http"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/require"
)

func TestLogicalServiceManagementRoutesAreRegistered(t *testing.T) {
	ws := new(restful.WebService)
	ws.Path("/naming/v1")
	(&HTTPServer{}).addLogicalServiceAccess(ws)

	want := map[string]bool{
		http.MethodGet + " /naming/v1/logical-services":                              false,
		http.MethodPost + " /naming/v1/logical-services":                             false,
		http.MethodGet + " /naming/v1/logical-services/environments":                 false,
		http.MethodGet + " /naming/v1/logical-services/unbound-environments":         false,
		http.MethodPost + " /naming/v1/logical-services/environment-bindings":        false,
		http.MethodPost + " /naming/v1/logical-services/environment-bindings/delete": false,
	}
	for _, route := range ws.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for route, found := range want {
		require.True(t, found, route)
	}
}
