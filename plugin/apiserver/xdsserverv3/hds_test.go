package xdsserverv3

import (
	"context"
	"io"
	"testing"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	healthservice "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFetchHealthCheckRejectsEmptyRequest(t *testing.T) {
	resp, err := (&XDSServer{}).FetchHealthCheck(context.Background(), nil)

	require.Nil(t, resp)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestFetchHealthCheckAcceptsEndpointHealthResponse(t *testing.T) {
	resp, err := (&XDSServer{}).FetchHealthCheck(context.Background(), &healthservice.HealthCheckRequestOrEndpointHealthResponse{
		RequestType: &healthservice.HealthCheckRequestOrEndpointHealthResponse_EndpointHealthResponse{
			EndpointHealthResponse: &healthservice.EndpointHealthResponse{},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.ClusterHealthChecks)
}

func TestStreamHealthCheckKeepsClientForEndpointResponses(t *testing.T) {
	stream := &fakeHealthCheckStream{
		ctx: context.Background(),
		requests: []*healthservice.HealthCheckRequestOrEndpointHealthResponse{
			{
				RequestType: &healthservice.HealthCheckRequestOrEndpointHealthResponse_HealthCheckRequest{
					HealthCheckRequest: &healthservice.HealthCheckRequest{
						Node: &core.Node{Id: "default/test-node~127.0.0.1"},
					},
				},
			},
			{
				RequestType: &healthservice.HealthCheckRequestOrEndpointHealthResponse_EndpointHealthResponse{
					EndpointHealthResponse: &healthservice.EndpointHealthResponse{},
				},
			},
		},
	}

	err := (&XDSServer{}).StreamHealthCheck(stream)

	require.NoError(t, err)
	require.Len(t, stream.sent, 1)
}

type fakeHealthCheckStream struct {
	grpc.ServerStream
	ctx      context.Context
	requests []*healthservice.HealthCheckRequestOrEndpointHealthResponse
	sent     []*healthservice.HealthCheckSpecifier
}

func (f *fakeHealthCheckStream) Context() context.Context {
	return f.ctx
}

func (f *fakeHealthCheckStream) Recv() (*healthservice.HealthCheckRequestOrEndpointHealthResponse, error) {
	if len(f.requests) == 0 {
		return nil, io.EOF
	}
	req := f.requests[0]
	f.requests = f.requests[1:]
	return req, nil
}

func (f *fakeHealthCheckStream) Send(resp *healthservice.HealthCheckSpecifier) error {
	f.sent = append(f.sent, resp)
	return nil
}
