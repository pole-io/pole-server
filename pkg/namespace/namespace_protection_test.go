package namespace

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	cachemock "github.com/pole-io/pole-server/pkg/cache/mock"
	storemock "github.com/pole-io/pole-server/plugin/store/mock"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

func TestDeleteNamespaceRejectsProtectedNamespaces(t *testing.T) {
	t.Parallel()

	server := &Server{}
	for _, name := range []string{DefaultNamespace, SystemNamespace} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := server.DeleteNamespace(context.Background(), &apimodel.Namespace{Name: name})

			require.Equal(t, uint32(apimodel.Code_InvalidParameter), response.GetCode())
			require.Contains(t, response.GetInfo(), "built-in namespace cannot be deleted")
		})
	}
}

func TestIsProtectedNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		protected bool
	}{
		{name: DefaultNamespace, protected: true},
		{name: SystemNamespace, protected: true},
		{name: ProductionNamespace, protected: false},
		{name: "DEFAULT", protected: false},
		{name: "", protected: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.protected, isProtectedNamespace(test.name))
		})
	}
}

func TestDeleteNamespacesRejectsProtectedNamespaces(t *testing.T) {
	t.Parallel()

	server := &Server{}
	response := server.DeleteNamespaces(context.Background(), []*apimodel.Namespace{
		{Name: DefaultNamespace},
		{Name: SystemNamespace},
	})

	require.Equal(t, uint32(apimodel.Code_InvalidParameter), response.GetCode())
	require.Len(t, response.GetResponses(), 2)
	for _, item := range response.GetResponses() {
		require.Equal(t, uint32(apimodel.Code_InvalidParameter), item.GetCode())
		require.Contains(t, item.GetInfo(), "built-in namespace cannot be deleted")
	}
}

func TestDeleteNamespaceRejectsOwnedGovernanceRules(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	tx := storemock.NewMockTransaction(controller)
	server := &Server{storage: storage}

	storage.EXPECT().CreateTransaction().Return(tx, nil)
	tx.EXPECT().Commit().Return(nil)
	tx.EXPECT().LockNamespace("prod").Return(&types.Namespace{Name: "prod"}, nil)
	storage.EXPECT().GetServices(map[string]string{"namespace": "prod"}, nil, nil, uint32(0), uint32(1)).Return(uint32(0), nil, nil)
	storage.EXPECT().CountConfigGroups("prod").Return(uint64(0), nil)
	storage.EXPECT().CountGovernanceRules("prod").Return(uint64(2), nil)

	response := server.DeleteNamespace(context.Background(), &apimodel.Namespace{Name: "prod"})

	require.Equal(t, uint32(apimodel.Code_NamespaceExistedGovernanceRules), response.GetCode())
	require.Contains(t, response.GetInfo(), "governance rules")
}

func TestCreateNamespaceRejectsSystemKindAndReservedName(t *testing.T) {
	t.Parallel()

	server := &Server{}
	for _, request := range []*apimodel.Namespace{
		{Name: "business-name", Kind: apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM},
		{Name: "unknown-kind", Kind: apimodel.NamespaceKind(2)},
		{Name: SystemNamespace, Kind: apimodel.NamespaceKind_NAMESPACE_KIND_BUSINESS},
	} {
		response := server.CreateNamespace(context.Background(), request)

		require.Equal(t, uint32(apimodel.Code_InvalidParameter), response.GetCode())
		require.Contains(t, response.GetInfo(), "system namespace")
	}
}

func TestCreateNamespacePersistsBusinessKindByDefault(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}

	storage.EXPECT().GetNamespace("development").Return(nil, nil)
	storage.EXPECT().AddNamespace(gomock.Any()).DoAndReturn(func(namespace *types.Namespace) error {
		require.Equal(t, apimodel.NamespaceKind_NAMESPACE_KIND_BUSINESS, namespace.Kind)
		return nil
	})

	response := server.CreateNamespace(context.Background(), &apimodel.Namespace{Name: "development"})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}

func TestGetNamespacesReturnsKindsAndAppliesKindFilter(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	cacheManager := cachemock.NewMockCacheManager(controller)
	namespaceCache := cachemock.NewMockNamespaceCache(controller)
	serviceCache := cachemock.NewMockServiceCache(controller)
	server := &Server{storage: storage, caches: cacheManager}
	filter := map[string][]string{"kind": {"system"}}

	cacheManager.EXPECT().Namespace().Return(namespaceCache)
	namespaceCache.EXPECT().Query(gomock.Any(), &cacheapi.NamespaceArgs{
		Filter: filter,
		Offset: 0,
		Limit:  10,
	}).Return(uint32(1), []*types.Namespace{{
		Name: SystemNamespace,
		Kind: apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM,
	}}, nil)
	storage.EXPECT().CountConfigFileEachGroup().Return(map[string]map[string]int64{}, nil)
	cacheManager.EXPECT().Service().Return(serviceCache)
	serviceCache.EXPECT().GetNamespaceCntInfo(SystemNamespace).Return(svctypes.NamespaceServiceCount{
		InstanceCnt: &svctypes.InstanceCount{},
	})

	response := server.GetNamespaces(context.Background(), map[string][]string{
		"kind":   {"system"},
		"offset": {"0"},
		"limit":  {"10"},
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
	require.Len(t, response.GetData(), 1)
	namespace := &apimodel.Namespace{}
	require.NoError(t, response.GetData()[0].UnmarshalTo(namespace))
	require.Equal(t, apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM, namespace.GetKind())
	require.False(t, namespace.GetDeleteable())
}
