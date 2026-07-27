package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	storeapi "github.com/pole-io/pole-server/apis/store"
)

type logicalServiceStoreStub struct {
	storeapi.Store
	created *svctypes.LogicalService
	err     error
}

func (s *logicalServiceStoreStub) CreateLogicalService(service *svctypes.LogicalService) error {
	s.created = service
	return s.err
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
