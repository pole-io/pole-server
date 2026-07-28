package aia2a

import (
	"net/http"
	"testing"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
)

func TestA2ADefinitionRoutesAreRegistered(t *testing.T) {
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

func TestValidateA2ADefinitionEnvironmentRejectsPoleSystem(t *testing.T) {
	agentCache := &definitionA2AAgentCache{agent: &aitypes.A2AAgent{
		Id: "agent-1", Name: "pole-agent", Namespace: "pole-system",
	}}
	namespaceCache := &definitionA2ANamespaceCache{namespace: &types.Namespace{
		Name: "pole-system", Kind: apimodel.NamespaceKind_NAMESPACE_KIND_BUSINESS,
	}}
	h := &HTTPServer{cacheMgr: definitionA2ACacheManager{
		agent: agentCache, namespace: namespaceCache,
	}}

	response := h.validateA2ADefinitionEnvironment("agent-1")
	require.NotNil(t, response)
	require.Equal(t, uint32(apimodel.Code_NotAllowedAccess), response.GetCode())
}

type definitionA2ACacheManager struct {
	cacheapi.CacheManager
	agent     cacheapi.A2AAgentCache
	namespace cacheapi.NamespaceCache
}

func (m definitionA2ACacheManager) A2AAgent() cacheapi.A2AAgentCache   { return m.agent }
func (m definitionA2ACacheManager) Namespace() cacheapi.NamespaceCache { return m.namespace }

type definitionA2AAgentCache struct {
	cacheapi.A2AAgentCache
	agent *aitypes.A2AAgent
}

func (c *definitionA2AAgentCache) GetA2AAgentByID(string) *aitypes.A2AAgent { return c.agent }

type definitionA2ANamespaceCache struct {
	cacheapi.NamespaceCache
	namespace *types.Namespace
}

func (c *definitionA2ANamespaceCache) GetNamespace(string) *types.Namespace { return c.namespace }
