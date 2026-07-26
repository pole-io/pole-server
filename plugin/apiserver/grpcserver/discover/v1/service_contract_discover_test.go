package v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/service"
)

type serviceContractDiscoverStub struct {
	service.DiscoverServer
	request *apiservice.Service
}

func (s *serviceContractDiscoverStub) DiscoverServiceContracts(
	_ context.Context, request *apiservice.Service,
) *apiservice.DiscoverResponse {
	s.request = request
	response := api.NewDiscoverResponse(apimodel.Code_ExecuteSuccess)
	response.Type = apiservice.DiscoverResponse_SERVICE_CONTRACTS
	return response
}

func TestHandleDiscoverDispatchesServiceContracts(t *testing.T) {
	stub := &serviceContractDiscoverStub{}
	server := &DiscoverGRPCServer{namingServer: stub}
	request := &apiservice.Service{Name: "payments", Namespace: "default"}

	response := server.handleDiscoverRequest(context.Background(), &apiservice.DiscoverRequest{
		Type:    apiservice.DiscoverRequest_SERVICE_CONTRACTS,
		Service: request,
	})

	require.Equal(t, apiservice.DiscoverResponse_SERVICE_CONTRACTS, response.GetType())
	require.Same(t, request, stub.request)
}
