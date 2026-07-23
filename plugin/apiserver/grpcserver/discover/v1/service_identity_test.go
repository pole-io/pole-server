package v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/service"
)

type serviceIdentityDiscoverStub struct {
	service.DiscoverServer
	token string
}

func (s *serviceIdentityDiscoverStub) GetServiceIdentity(
	ctx context.Context, _ *apiservice.Service,
) *apiservice.DiscoverResponse {
	s.token = utils.ParseAuthToken(ctx)
	return api.NewDiscoverServiceIdentityResponse(apimodel.Code_ExecuteSuccess)
}

func TestHandleServiceIdentityDoesNotAllowBodyTokenOverride(t *testing.T) {
	stub := &serviceIdentityDiscoverStub{}
	server := &DiscoverGRPCServer{namingServer: stub}
	ctx := context.WithValue(context.Background(), types.ContextAuthTokenKey, "metadata-token")

	resp := server.handleDiscoverRequest(ctx, &apiservice.DiscoverRequest{
		Type: apiservice.DiscoverRequest_SERVICE_IDENTITY,
		Service: &apiservice.Service{
			Name: "orders", Namespace: "default", Token: "body-token",
		},
	})

	assert.Equal(t, "metadata-token", stub.token)
	assert.Equal(t, apiservice.DiscoverResponse_SERVICE_IDENTITY, resp.GetType())
}
