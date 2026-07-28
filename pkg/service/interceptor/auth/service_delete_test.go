package service_auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	serviceapi "github.com/pole-io/pole-server/pkg/service"
)

type deleteServiceServerStub struct {
	serviceapi.DiscoverServer
	cache cacheapi.CacheManager
}

func (s deleteServiceServerStub) Cache() cacheapi.CacheManager {
	return s.cache
}

type deleteServiceCacheManagerStub struct {
	cacheapi.CacheManager
	service   cacheapi.ServiceCache
	namespace cacheapi.NamespaceCache
}

func (s deleteServiceCacheManagerStub) Service() cacheapi.ServiceCache {
	return s.service
}

func (s deleteServiceCacheManagerStub) Namespace() cacheapi.NamespaceCache {
	return s.namespace
}

type deleteServiceCacheStub struct {
	cacheapi.ServiceCache
	service *svctypes.Service
}

func (s deleteServiceCacheStub) GetServiceByID(id string) *svctypes.Service {
	if s.service != nil && s.service.ID == id {
		return s.service
	}
	return nil
}

type deleteServiceNamespaceCacheStub struct {
	cacheapi.NamespaceCache
}

func (deleteServiceNamespaceCacheStub) GetNamespacesByName(_ []string) []*types.Namespace {
	return nil
}

func TestDeleteServiceAuthorizationResolvesResourceByID(t *testing.T) {
	stored := &svctypes.Service{ID: "service-id", Name: "orders", Namespace: "production"}
	server := &Server{nextSvr: deleteServiceServerStub{
		cache: deleteServiceCacheManagerStub{
			service:   deleteServiceCacheStub{service: stored},
			namespace: deleteServiceNamespaceCacheStub{},
		},
	}}

	resources := server.queryServiceResource([]*apiservice.Service{{Id: "service-id"}})

	require.Empty(t, resources[apisecurity.ResourceType_Namespaces])
	require.Equal(t, "service-id", resources[apisecurity.ResourceType_Services][0].ID)
}
