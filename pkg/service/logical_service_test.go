package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	storeapi "github.com/pole-io/pole-server/apis/store"
)

type logicalServiceStoreStub struct {
	storeapi.Store
	created        *svctypes.LogicalService
	logicalService *svctypes.LogicalService
	bindings       []*svctypes.ServiceEnvironmentBinding
	deletedBinding string
	err            error
}

func (s *logicalServiceStoreStub) CreateLogicalService(service *svctypes.LogicalService) error {
	s.created = service
	return s.err
}

func (s *logicalServiceStoreStub) GetLogicalService(_ string) (*svctypes.LogicalService, error) {
	return s.logicalService, s.err
}

func (s *logicalServiceStoreStub) ListServiceEnvironmentBindings(
	_ string) ([]*svctypes.ServiceEnvironmentBinding, error) {
	return s.bindings, s.err
}

func (s *logicalServiceStoreStub) UnbindServiceEnvironment(_, serviceID, _ string) error {
	s.deletedBinding = serviceID
	return s.err
}

type logicalServiceCacheManagerStub struct {
	cacheapi.CacheManager
	serviceCache   cacheapi.ServiceCache
	namespaceCache cacheapi.NamespaceCache
	instanceCache  cacheapi.InstanceCache
}

func (s logicalServiceCacheManagerStub) Service() cacheapi.ServiceCache {
	return s.serviceCache
}

func (s logicalServiceCacheManagerStub) Namespace() cacheapi.NamespaceCache {
	return s.namespaceCache
}

func (s logicalServiceCacheManagerStub) Instance() cacheapi.InstanceCache {
	return s.instanceCache
}

type logicalServiceServiceCacheStub struct {
	cacheapi.ServiceCache
	service  *svctypes.Service
	services []*svctypes.Service
}

func (s logicalServiceServiceCacheStub) GetServiceByID(_ string) *svctypes.Service {
	return s.service
}

func (s logicalServiceServiceCacheStub) IteratorServices(iter cacheapi.ServiceIterProc) error {
	for _, service := range s.services {
		if _, err := iter(service.ID, service); err != nil {
			return err
		}
	}
	return nil
}

type logicalServiceNamespaceCacheStub struct {
	cacheapi.NamespaceCache
	namespace *types.Namespace
}

func (s logicalServiceNamespaceCacheStub) GetNamespace(_ string) *types.Namespace {
	return s.namespace
}

type logicalServiceInstanceCacheStub struct {
	cacheapi.InstanceCache
}

func (logicalServiceInstanceCacheStub) GetInstancesCountByServiceID(_ string) svctypes.InstanceCount {
	return svctypes.InstanceCount{}
}

func (s *logicalServiceStoreStub) ListLogicalServices(
	_ string, _, _ uint32) (uint32, []*svctypes.LogicalService, error) {
	return 0, nil, s.err
}

func TestCreateLogicalServicesAssignsControlPlaneIdentity(t *testing.T) {
	storage := &logicalServiceStoreStub{}
	server := &Server{storage: storage}

	response := server.CreateLogicalServices(context.Background(), []*apiservice.LogicalService{{
		Name: "checkout", Comment: "结算服务", Owners: "team-a",
	}})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
	require.Len(t, response.GetResponses(), 1)
	require.NotNil(t, storage.created)
	require.Len(t, storage.created.ID, 32)
	require.Len(t, storage.created.Revision, 32)
	require.Equal(t, "checkout", storage.created.Name)

	created := &apiservice.LogicalService{}
	require.NoError(t, anypb.UnmarshalTo(response.GetResponses()[0].GetData(), created, proto.UnmarshalOptions{}))
	require.Equal(t, storage.created.ID, created.GetId())
	require.Equal(t, storage.created.Revision, created.GetRevision())
}

func TestBindServiceEnvironmentRejectsSystemNamespace(t *testing.T) {
	storage := &logicalServiceStoreStub{
		logicalService: &svctypes.LogicalService{ID: "logical-1", Name: "pole"},
	}
	server := &Server{
		storage: storage,
		caches: logicalServiceCacheManagerStub{
			serviceCache: logicalServiceServiceCacheStub{
				service: &svctypes.Service{ID: "service-1", Name: "pole-control-plane", Namespace: "pole-system", Valid: true},
			},
			namespaceCache: logicalServiceNamespaceCacheStub{
				namespace: &types.Namespace{
					Name: "pole-system",
					Kind: apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM,
				},
			},
		},
	}

	response := server.BindServiceEnvironment(context.Background(), &apiservice.BindServiceEnvironmentRequest{
		LogicalServiceId: "logical-1",
		ServiceId:        "service-1",
	})

	require.Equal(t, uint32(apimodel.Code_NotAllowedAccess), response.GetCode())
	require.Contains(t, response.GetInfo(), "system namespace")
}

func TestBindServiceEnvironmentRejectsLegacySystemNamespaceByName(t *testing.T) {
	storage := &logicalServiceStoreStub{
		logicalService: &svctypes.LogicalService{ID: "logical-1", Name: "pole"},
	}
	server := &Server{
		storage: storage,
		caches: logicalServiceCacheManagerStub{
			serviceCache: logicalServiceServiceCacheStub{
				service: &svctypes.Service{ID: "service-1", Name: "pole-control-plane", Namespace: SystemNamespace, Valid: true},
			},
			namespaceCache: logicalServiceNamespaceCacheStub{
				namespace: &types.Namespace{
					Name: SystemNamespace,
					Kind: apimodel.NamespaceKind_NAMESPACE_KIND_BUSINESS,
				},
			},
		},
	}

	response := server.BindServiceEnvironment(context.Background(), &apiservice.BindServiceEnvironmentRequest{
		LogicalServiceId: "logical-1",
		ServiceId:        "service-1",
	})

	require.Equal(t, uint32(apimodel.Code_NotAllowedAccess), response.GetCode())
}

func TestHistoricalSystemBindingRemainsDiscoverableAndCanBeUnbound(t *testing.T) {
	storage := &logicalServiceStoreStub{
		bindings: []*svctypes.ServiceEnvironmentBinding{{
			LogicalServiceID: "logical-1",
			ServiceID:        "system-service",
			Namespace:        SystemNamespace,
			ServiceName:      "pole-control-plane",
		}},
	}
	server := &Server{
		storage: storage,
		caches: logicalServiceCacheManagerStub{
			serviceCache: logicalServiceServiceCacheStub{
				service: &svctypes.Service{
					ID:        "system-service",
					Name:      "pole-control-plane",
					Namespace: SystemNamespace,
					Valid:     true,
				},
			},
			namespaceCache: logicalServiceNamespaceCacheStub{
				namespace: &types.Namespace{Name: SystemNamespace},
			},
			instanceCache: logicalServiceInstanceCacheStub{},
		},
	}

	list := server.GetLogicalServiceEnvironments(context.Background(), "logical-1")
	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), list.GetCode())
	require.Len(t, list.GetData(), 1)
	binding := &apiservice.ServiceEnvironmentBinding{}
	require.NoError(t, list.GetData()[0].UnmarshalTo(binding))
	require.Equal(t, "system-service", binding.GetServiceId())

	response := server.UnbindServiceEnvironment(context.Background(), &apiservice.UnbindServiceEnvironmentRequest{
		LogicalServiceId: "logical-1",
		ServiceId:        binding.GetServiceId(),
	})
	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
	require.Equal(t, "system-service", storage.deletedBinding)
}

func TestGetUnboundServiceEnvironmentsExcludesSystemNamespace(t *testing.T) {
	storage := &logicalServiceStoreStub{}
	serviceCache := logicalServiceServiceCacheStub{services: []*svctypes.Service{
		{ID: "system-service", Name: "pole-control-plane", Namespace: "pole-system", Valid: true},
		{ID: "business-service", Name: "checkout", Namespace: "production", Valid: true},
	}}
	namespaceCache := &logicalServiceNamespaceByNameCacheStub{namespaces: map[string]*types.Namespace{
		"pole-system": {
			Name: "pole-system",
			Kind: apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM,
		},
		"production": {
			Name: "production",
			Kind: apimodel.NamespaceKind_NAMESPACE_KIND_BUSINESS,
		},
	}}
	server := &Server{
		storage: storage,
		caches: logicalServiceCacheManagerStub{
			serviceCache:   serviceCache,
			namespaceCache: namespaceCache,
			instanceCache:  logicalServiceInstanceCacheStub{},
		},
	}

	response := server.GetUnboundServiceEnvironments(context.Background(), map[string]string{
		"offset": "0",
		"limit":  "10",
	})

	require.Equal(t, uint32(1), response.GetAmount())
	require.Len(t, response.GetData(), 1)
	service := &apiservice.Service{}
	require.NoError(t, response.GetData()[0].UnmarshalTo(service))
	require.Equal(t, "business-service", service.GetId())
}

type logicalServiceNamespaceByNameCacheStub struct {
	cacheapi.NamespaceCache
	namespaces map[string]*types.Namespace
}

func (s *logicalServiceNamespaceByNameCacheStub) GetNamespace(name string) *types.Namespace {
	return s.namespaces[name]
}
