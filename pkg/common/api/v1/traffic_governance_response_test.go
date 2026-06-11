package v1

import (
	"testing"

	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

func TestNewDiscoverTrafficGovernanceResponses(t *testing.T) {
	service := &apiservice.Service{Name: "svc-a", Namespace: "default"}

	require.Equal(t, apiservice.DiscoverResponse_TRAFFIC_SECURITY_RULE,
		NewDiscoverTrafficSecurityResponse(apimodel.Code_DataNoChange, service).GetType())
	require.Equal(t, apiservice.DiscoverResponse_TRAFFIC_MIRROR_RULE,
		NewDiscoverTrafficMirrorResponse(apimodel.Code_DataNoChange, service).GetType())
	require.Equal(t, apiservice.DiscoverResponse_TRAFFIC_MOCK_RULE,
		NewDiscoverTrafficMockResponse(apimodel.Code_DataNoChange, service).GetType())
}
