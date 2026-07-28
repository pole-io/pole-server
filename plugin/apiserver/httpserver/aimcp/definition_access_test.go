package aimcp

import (
	"net/http"
	"testing"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/require"

	specai "github.com/pole-io/specification/source/go/api/v1/ai"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
)

func TestMCPDefinitionRoutesAreRegistered(t *testing.T) {
	ws := new(restful.WebService)
	ws.Path(basePath)
	(&HTTPServer{}).addDefaultAccess(ws)

	want := map[string]bool{
		http.MethodGet + " " + basePath + "/definitions":                              false,
		http.MethodPost + " " + basePath + "/definitions":                             false,
		http.MethodPut + " " + basePath + "/definitions":                              false,
		http.MethodPost + " " + basePath + "/definitions/delete":                      false,
		http.MethodGet + " " + basePath + "/definitions/environments":                 false,
		http.MethodGet + " " + basePath + "/definitions/environment-binding":          false,
		http.MethodPost + " " + basePath + "/definitions/environment-bindings":        false,
		http.MethodPost + " " + basePath + "/definitions/environment-bindings/delete": false,
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

func TestValidateMCPDefinitionEnvironmentRejectsSystemNamespace(t *testing.T) {
	serverCache := &definitionMCPServerCache{server: &specai.MCPServer{
		Id: "server-1", Name: "internal-tool", Namespace: "internal",
	}}
	namespaceCache := &definitionNamespaceCache{namespace: &types.Namespace{
		Name: "internal", Kind: apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM,
	}}
	h := &HTTPServer{cacheMgr: definitionCacheManager{
		mcp: serverCache, namespace: namespaceCache,
	}}

	response := h.validateMCPDefinitionEnvironment("server-1")
	require.NotNil(t, response)
	require.Equal(t, uint32(apimodel.Code_NotAllowedAccess), response.GetCode())
}

type definitionCacheManager struct {
	cacheapi.CacheManager
	mcp       cacheapi.MCPServerCache
	namespace cacheapi.NamespaceCache
}

func (m definitionCacheManager) MCPServer() cacheapi.MCPServerCache { return m.mcp }
func (m definitionCacheManager) Namespace() cacheapi.NamespaceCache { return m.namespace }

type definitionMCPServerCache struct {
	cacheapi.MCPServerCache
	server *specai.MCPServer
}

func (c *definitionMCPServerCache) GetMCPServerByID(string) *specai.MCPServer { return c.server }

type definitionNamespaceCache struct {
	cacheapi.NamespaceCache
	namespace *types.Namespace
}

func (c *definitionNamespaceCache) GetNamespace(string) *types.Namespace { return c.namespace }
