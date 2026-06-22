package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	cachemock "github.com/pole-io/pole-server/pkg/cache/mock"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
)

// TestServer_GetServiceWithCache 测试服务查询的缓存逻辑
func TestServer_GetServiceWithCache(t *testing.T) {
	// 跳过：需要完整的测试套件初始化
	// 该测试需要在 DiscoverTestSuit 中运行
	// 追踪 Issue: POLE-TEST-SERVICE-CACHE-001
	t.Skip("需要测试套件初始化，请在集成测试中验证")

	// 测试场景：
	// 1. 首次查询，缓存未命中，从存储层加载
	// 2. 二次查询，缓存命中，直接返回
	// 3. 缓存过期后，重新加载
	// 4. 服务更新后，缓存刷新

	_ = context.Background()
	_ = time.Second
	_ = assert.Equal

	// 示例测试结构（需要在完整测试套件中实现）：
	// testSuit := &DiscoverTestSuit{}
	// err := testSuit.Initialize()
	// assert.NoError(t, err)
	// defer testSuit.Destroy()
}

func TestServer_ServiceInstancesCacheDoesNotMergeVisibleServicesFromOtherNamespaces(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceCache := cachemock.NewMockServiceCache(ctrl)
	instanceCache := cachemock.NewMockInstanceCache(ctrl)
	revisionWorker := cachemock.NewMockServiceRevisionWorker(ctrl)

	server := &Server{
		caches:                &testServiceDiscoveryCacheManager{service: serviceCache, instance: instanceCache},
		emptyPushProtectSvs:   container.NewSyncMap[string, time.Time](),
		discoverResponseCache: newDiscoverResponseCache(defaultDiscoverResponseCacheSize),
	}

	currentSvc := &svctypes.Service{
		ID:        "svc-current",
		Name:      "orders",
		Namespace: "default",
		Valid:     true,
	}
	currentInstance := svctypes.CreateInstanceModel(currentSvc.ID, &apiservice.Instance{
		Id:        "ins-current",
		Service:   currentSvc.Name,
		Namespace: currentSvc.Namespace,
		Host:      "10.0.0.1",
		Port:      8080,
		Healthy:   true,
		Location:  &apimodel.Location{Region: "local"},
	})

	serviceCache.EXPECT().GetServiceByName("orders", "default").Return(currentSvc)
	serviceCache.EXPECT().GetRevisionWorker().Return(revisionWorker).AnyTimes()
	revisionWorker.EXPECT().GetServiceInstanceRevision(currentSvc.ID).Return("rev-current").AnyTimes()
	instanceCache.EXPECT().DiscoverServiceInstances(currentSvc.ID, false, gomock.Any()).
		Do(func(_ string, _ bool, consume func(*svctypes.Instance)) {
			consume(currentInstance)
		})

	resp := server.ServiceInstancesCache(context.Background(), &apiservice.DiscoverFilter{}, &apiservice.Service{
		Name:      "orders",
		Namespace: "default",
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), resp.GetCode())
	require.Len(t, resp.GetInstances(), 1)
	assert.Equal(t, "ins-current", resp.GetInstances()[0].GetId())
	assert.Equal(t, "default", resp.GetInstances()[0].GetNamespace())
}

func TestServer_ServiceInstancesCacheReusesFullResponseForSameRevision(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceCache := cachemock.NewMockServiceCache(ctrl)
	instanceCache := cachemock.NewMockInstanceCache(ctrl)
	revisionWorker := cachemock.NewMockServiceRevisionWorker(ctrl)

	server := &Server{
		caches:                &testServiceDiscoveryCacheManager{service: serviceCache, instance: instanceCache},
		emptyPushProtectSvs:   container.NewSyncMap[string, time.Time](),
		discoverResponseCache: newDiscoverResponseCache(defaultDiscoverResponseCacheSize),
	}

	currentSvc := &svctypes.Service{
		ID:        "svc-current",
		Name:      "orders",
		Namespace: "default",
		Valid:     true,
	}
	currentInstance := svctypes.CreateInstanceModel(currentSvc.ID, &apiservice.Instance{
		Id:        "ins-current",
		Service:   currentSvc.Name,
		Namespace: currentSvc.Namespace,
		Host:      "10.0.0.1",
		Port:      8080,
		Healthy:   true,
		Location:  &apimodel.Location{Region: "local"},
	})

	serviceCache.EXPECT().GetServiceByName("orders", "default").Return(currentSvc).Times(2)
	serviceCache.EXPECT().GetRevisionWorker().Return(revisionWorker).AnyTimes()
	revisionWorker.EXPECT().GetServiceInstanceRevision(currentSvc.ID).Return("rev-current").Times(2)
	instanceCache.EXPECT().DiscoverServiceInstances(currentSvc.ID, false, gomock.Any()).
		Do(func(_ string, _ bool, consume func(*svctypes.Instance)) {
			consume(currentInstance)
		}).
		Times(1)

	first := server.ServiceInstancesCache(context.Background(), &apiservice.DiscoverFilter{}, &apiservice.Service{
		Name:      "orders",
		Namespace: "default",
	})
	second := server.ServiceInstancesCache(context.Background(), &apiservice.DiscoverFilter{}, &apiservice.Service{
		Name:      "orders",
		Namespace: "default",
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), first.GetCode())
	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), second.GetCode())
	require.Len(t, second.GetInstances(), 1)
	assert.Equal(t, "ins-current", second.GetInstances()[0].GetId())
}

type testServiceDiscoveryCacheManager struct {
	service  cacheapi.ServiceCache
	instance cacheapi.InstanceCache
}

func (m *testServiceDiscoveryCacheManager) GetUpdateCacheInterval() time.Duration { return time.Second }
func (m *testServiceDiscoveryCacheManager) GetReportInterval() time.Duration      { return time.Second }
func (m *testServiceDiscoveryCacheManager) GetTimeDiff() time.Duration            { return 0 }
func (m *testServiceDiscoveryCacheManager) GetCacher(cacheapi.CacheIndex) cacheapi.Cache {
	return nil
}
func (m *testServiceDiscoveryCacheManager) RegisterCacher(cacheapi.CacheIndex, cacheapi.Cache) {}
func (m *testServiceDiscoveryCacheManager) OpenResourceCache(...cacheapi.ConfigEntry) error {
	return nil
}
func (m *testServiceDiscoveryCacheManager) Service() cacheapi.ServiceCache               { return m.service }
func (m *testServiceDiscoveryCacheManager) Instance() cacheapi.InstanceCache             { return m.instance }
func (m *testServiceDiscoveryCacheManager) RoutingConfig() cacheapi.RouterRuleCache      { return nil }
func (m *testServiceDiscoveryCacheManager) RateLimit() cacheapi.RateLimitCache           { return nil }
func (m *testServiceDiscoveryCacheManager) CircuitBreaker() cacheapi.CircuitBreakerCache { return nil }
func (m *testServiceDiscoveryCacheManager) FaultDetector() cacheapi.FaultDetectCache     { return nil }
func (m *testServiceDiscoveryCacheManager) Lossless() cacheapi.LosslessCache             { return nil }
func (m *testServiceDiscoveryCacheManager) TrafficSecurity() cacheapi.TrafficGovernanceCache {
	return nil
}
func (m *testServiceDiscoveryCacheManager) TrafficMirror() cacheapi.TrafficGovernanceCache {
	return nil
}
func (m *testServiceDiscoveryCacheManager) TrafficMock() cacheapi.TrafficGovernanceCache {
	return nil
}
func (m *testServiceDiscoveryCacheManager) ServiceContract() cacheapi.ServiceContractCache {
	return nil
}
func (m *testServiceDiscoveryCacheManager) LaneRule() cacheapi.LaneCache           { return nil }
func (m *testServiceDiscoveryCacheManager) User() cacheapi.UserCache               { return nil }
func (m *testServiceDiscoveryCacheManager) AuthStrategy() cacheapi.StrategyCache   { return nil }
func (m *testServiceDiscoveryCacheManager) Namespace() cacheapi.NamespaceCache     { return nil }
func (m *testServiceDiscoveryCacheManager) Client() cacheapi.ClientCache           { return nil }
func (m *testServiceDiscoveryCacheManager) ConfigFile() cacheapi.ConfigFileCache   { return nil }
func (m *testServiceDiscoveryCacheManager) ConfigGroup() cacheapi.ConfigGroupCache { return nil }
func (m *testServiceDiscoveryCacheManager) Gray() cacheapi.GrayCache               { return nil }
func (m *testServiceDiscoveryCacheManager) Role() cacheapi.RoleCache               { return nil }
func (m *testServiceDiscoveryCacheManager) MCPServer() cacheapi.MCPServerCache     { return nil }
func (m *testServiceDiscoveryCacheManager) A2AAgent() cacheapi.A2AAgentCache       { return nil }
