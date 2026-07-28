package service

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	storemock "github.com/pole-io/pole-server/plugin/store/mock"
)

func TestDeleteServiceResolvesEnvironmentServiceByID(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}
	stored := &svctypes.Service{
		ID:        "service-id",
		Name:      "orders",
		Namespace: "production",
		Token:     "must-not-leak",
		Valid:     true,
	}

	storage.EXPECT().GetServiceByID("service-id").Return(stored, nil)
	expectEnvironmentServiceDeletion(storage, stored)

	response := server.DeleteService(context.Background(), &apiservice.Service{Id: "service-id"})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
	deleted := &apiservice.Service{}
	require.NoError(t, anypb.UnmarshalTo(response.GetData(), deleted, proto.UnmarshalOptions{}))
	require.Equal(t, "service-id", deleted.GetId())
	require.Equal(t, "orders", deleted.GetName())
	require.Equal(t, "production", deleted.GetNamespace())
	require.Empty(t, deleted.GetToken())
}

func TestDeleteServiceResolvesEnvironmentServiceByNamespaceAndName(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}
	stored := &svctypes.Service{
		ID:        "service-id",
		Name:      "orders",
		Namespace: "production",
		Valid:     true,
	}

	storage.EXPECT().GetService("orders", "production").Return(stored, nil)
	expectEnvironmentServiceDeletion(storage, stored)

	response := server.DeleteService(context.Background(), &apiservice.Service{
		Name:      "orders",
		Namespace: "production",
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
	deleted := &apiservice.Service{}
	require.NoError(t, anypb.UnmarshalTo(response.GetData(), deleted, proto.UnmarshalOptions{}))
	require.Equal(t, "service-id", deleted.GetId())
	require.Equal(t, "orders", deleted.GetName())
	require.Equal(t, "production", deleted.GetNamespace())
}

func expectEnvironmentServiceDeletion(storage *storemock.MockStore, service *svctypes.Service) {
	storage.EXPECT().GetExpandInstances(
		map[string]string{"name": service.Name, "namespace": service.Namespace},
		nil,
		uint32(0),
		uint32(1),
	).Return(uint32(0), nil, nil)
	storage.EXPECT().GetServiceAliases(
		map[string]string{"service": service.Name, "namespace": service.Namespace},
		uint32(0),
		uint32(1),
	).Return(uint32(0), nil, nil)
	storage.EXPECT().CountAIBackendReferences(service.ID).Return(uint32(0), nil)
	storage.EXPECT().DeleteService(service.ID, service.Name, service.Namespace).Return(nil)
}
