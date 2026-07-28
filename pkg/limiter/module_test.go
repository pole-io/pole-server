package limiter

import (
	"context"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	apiv2 "github.com/pole-io/specification/source/go/api/v1/traffic_manage/ratelimiter"
)

func TestRunningCanStartStopAndStartAgain(t *testing.T) {
	for range 2 {
		running, err := Start(context.Background(), testConfig())
		require.NoError(t, err)
		require.Len(t, running.Endpoints(), 1)

		address := strings.TrimSuffix(running.Endpoints()[0], "/grpc")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		conn, err := grpc.DialContext(ctx, address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock())
		cancel()
		require.NoError(t, err)

		response, err := apiv2.NewRateLimitGRPCV2Client(conn).TimeAdjust(
			context.Background(), &apiv2.TimeAdjustRequest{})
		require.NoError(t, err)
		require.NotZero(t, response.GetServerTimestamp())
		require.NoError(t, conn.Close())

		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		require.NoError(t, running.Stop(stopCtx))
		stopCancel()
		require.NoError(t, running.Wait())
	}
}

func TestStartReturnsListenerBindError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = listener.Close() }()
	port := listener.Addr().(*net.TCPAddr).Port

	cfg := testConfig()
	cfg.APIServers[0].Option["port"] = port
	_, err = Start(context.Background(), cfg)

	require.ErrorContains(t, err, "listen limiter grpc")
	require.ErrorContains(t, err, strconv.Itoa(port))
}

func testConfig() Config {
	return Config{
		APIServers: []APIServerConfig{{
			Name: "grpc",
			Option: map[string]interface{}{
				"ip":   "127.0.0.1",
				"port": 0,
			},
		}},
		Limit: validLimitConfig(),
	}
}

func validLimitConfig() LimitConfig {
	return LimitConfig{
		NodeID: 1,
	}
}
