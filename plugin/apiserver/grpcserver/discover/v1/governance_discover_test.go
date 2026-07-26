package v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/goverrule"
)

type governanceDiscoverStub struct {
	goverrule.GoverRuleServer
	filter *apiservice.DiscoverFilter
}

func (s *governanceDiscoverStub) GetLaneRuleWithCache(
	ctx context.Context, _ *apiservice.Service,
) *apiservice.DiscoverResponse {
	s.filter, _ = ctx.Value(types.ContextDiscoverFilter).(*apiservice.DiscoverFilter)
	resp := api.NewDiscoverResponse(apimodel.Code_ExecuteSuccess)
	resp.Type = apiservice.DiscoverResponse_LANE
	return resp
}

func TestHandleDiscoverDispatchesLaneAndForwardsFilter(t *testing.T) {
	stub := &governanceDiscoverStub{}
	server := &DiscoverGRPCServer{ruleServer: stub}
	filter := &apiservice.DiscoverFilter{Caller: &apimodel.Caller{
		Labels: []*apimodel.ClientLabel{{
			Key: "env", Value: &apimodel.MatchString{Type: apimodel.MatchString_EXACT, Value: "canary"},
		}},
	}}

	resp := server.handleDiscoverRequest(context.Background(), &apiservice.DiscoverRequest{
		Type:    apiservice.DiscoverRequest_LANE,
		Service: &apiservice.Service{Name: "orders", Namespace: "default"},
		Filter:  filter,
	})

	require.Equal(t, apiservice.DiscoverResponse_LANE, resp.GetType())
	require.Same(t, filter, stub.filter)
}
