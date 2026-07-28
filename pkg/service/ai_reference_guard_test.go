package service

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	storemock "github.com/pole-io/pole-server/plugin/store/mock"
)

func TestDeleteServiceRejectsAIBackendReferences(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}
	service := &svctypes.Service{
		ID:        "service-id",
		Name:      "orders",
		Namespace: "prod",
	}

	storage.EXPECT().GetExpandInstances(
		map[string]string{"name": "orders", "namespace": "prod"}, nil, uint32(0), uint32(1),
	).Return(uint32(0), nil, nil)
	storage.EXPECT().GetServiceAliases(
		map[string]string{"service": "orders", "namespace": "prod"}, uint32(0), uint32(1),
	).Return(uint32(0), nil, nil)
	storage.EXPECT().CountAIBackendReferences("service-id").Return(uint32(1), nil)

	response := server.isServiceExistedResource(context.Background(), service)

	require.Equal(t, uint32(apimodel.Code_ExistedResource), response.GetCode())
	require.Contains(t, response.GetInfo(), "MCP Server or A2A Agent")
}
