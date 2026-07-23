package namespace

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/apis/pkg/types"
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
