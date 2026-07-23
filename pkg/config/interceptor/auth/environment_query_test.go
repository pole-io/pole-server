package config_auth

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	cachemock "github.com/pole-io/pole-server/pkg/cache/mock"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/config"
	authmock "github.com/pole-io/pole-server/plugin/access_control/auth/mock"
)

type environmentQueryServer struct {
	config.ConfigCenterServer
	groups *apimodel.BatchQueryResponse
	files  *apimodel.BatchQueryResponse
}

type environmentCacheManager struct {
	cacheapi.CacheManager
	configGroup cacheapi.ConfigGroupCache
	namespace   cacheapi.NamespaceCache
}

func (m *environmentCacheManager) ConfigGroup() cacheapi.ConfigGroupCache { return m.configGroup }
func (m *environmentCacheManager) Namespace() cacheapi.NamespaceCache     { return m.namespace }

type environmentStrategyServer struct {
	authapi.StrategyServer
	checker authapi.AuthChecker
}

func (s *environmentStrategyServer) GetAuthChecker() authapi.AuthChecker { return s.checker }

type environmentConfigGroupCache struct {
	cacheapi.ConfigGroupCache
	groups map[string]*conftypes.ConfigFileGroup
}

func (c *environmentConfigGroupCache) GetGroupByName(namespace, name string) *conftypes.ConfigFileGroup {
	return c.groups[namespace+"/"+name]
}

func (s *environmentQueryServer) QueryConfigFileGroups(context.Context, map[string]string) *apimodel.BatchQueryResponse {
	return s.groups
}

func (s *environmentQueryServer) SearchConfigFiles(context.Context, map[string]string) *apimodel.BatchQueryResponse {
	return s.files
}

func batchQuery(t *testing.T, messages ...proto.Message) *apimodel.BatchQueryResponse {
	t.Helper()
	resp := api.NewConfigBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	for _, message := range messages {
		if err := api.AddAnyDataIntoBatchQuery(resp, message); err != nil {
			t.Fatalf("add response data: %v", err)
		}
	}
	resp.Amount = uint32(len(messages))
	resp.Size = uint32(len(messages))
	return resp
}

func TestQueryConfigFileGroupsHidesUnauthorizedEnvironments(t *testing.T) {
	ctrl := gomock.NewController(t)
	checker := authmock.NewMockAuthChecker(ctrl)
	namespaceCache := cachemock.NewMockNamespaceCache(ctrl)

	checker.EXPECT().CheckConsolePermission(gomock.Any()).Return(true, nil)
	checker.EXPECT().ResourcePredicate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ *authtypes.AcquireContext, entry *authtypes.ResourceEntry) bool {
			return entry.ID == "group-dev"
		},
	).AnyTimes()
	namespaceCache.EXPECT().GetNamespace("prod").Return(&types.Namespace{Name: "prod"}).AnyTimes()

	next := &environmentQueryServer{groups: batchQuery(t,
		&apiconfig.ConfigFileGroup{Id: "group-dev", Namespace: "dev", Name: "application"},
		&apiconfig.ConfigFileGroup{Id: "group-prod", Namespace: "prod", Name: "application"},
	)}
	server := &Server{
		nextServer: next,
		cacheMgr:   &environmentCacheManager{namespace: namespaceCache},
		policySvr:  &environmentStrategyServer{checker: checker},
	}

	resp := server.QueryConfigFileGroups(context.Background(), map[string]string{"name": "application"})
	if resp.GetAmount() != 1 || resp.GetSize() != 1 || len(resp.GetData()) != 1 {
		t.Fatalf("unauthorized environment leaked through counts or data: amount=%d size=%d data=%d",
			resp.GetAmount(), resp.GetSize(), len(resp.GetData()))
	}
	got := &apiconfig.ConfigFileGroup{}
	if err := resp.GetData()[0].UnmarshalTo(got); err != nil {
		t.Fatalf("unmarshal group: %v", err)
	}
	if got.GetNamespace() != "dev" {
		t.Fatalf("got namespace %q, want dev", got.GetNamespace())
	}
}

func TestSearchConfigFilesHidesUnauthorizedEnvironments(t *testing.T) {
	ctrl := gomock.NewController(t)
	checker := authmock.NewMockAuthChecker(ctrl)
	namespaceCache := cachemock.NewMockNamespaceCache(ctrl)

	checker.EXPECT().CheckConsolePermission(gomock.Any()).Return(true, nil)
	checker.EXPECT().ResourcePredicate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ *authtypes.AcquireContext, entry *authtypes.ResourceEntry) bool {
			return entry.ID == "group-dev"
		},
	).AnyTimes()
	namespaceCache.EXPECT().GetNamespace("prod").Return(&types.Namespace{Name: "prod"}).AnyTimes()

	next := &environmentQueryServer{files: batchQuery(t,
		&apiconfig.ConfigFile{Namespace: "dev", Group: "application", Name: "app.yaml"},
		&apiconfig.ConfigFile{Namespace: "prod", Group: "application", Name: "app.yaml"},
	)}
	server := &Server{
		nextServer: next,
		cacheMgr: &environmentCacheManager{
			configGroup: &environmentConfigGroupCache{groups: map[string]*conftypes.ConfigFileGroup{
				"dev/application":  {Id: "group-dev"},
				"prod/application": {Id: "group-prod"},
			}},
			namespace: namespaceCache,
		},
		policySvr: &environmentStrategyServer{checker: checker},
	}

	resp := server.SearchConfigFiles(context.Background(), map[string]string{"group": "application", "name": "app.yaml"})
	if resp.GetAmount() != 1 || resp.GetSize() != 1 || len(resp.GetData()) != 1 {
		t.Fatalf("unauthorized environment leaked through counts or data: amount=%d size=%d data=%d",
			resp.GetAmount(), resp.GetSize(), len(resp.GetData()))
	}
}
