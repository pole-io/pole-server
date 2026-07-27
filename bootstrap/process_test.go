package bootstrap

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/apis/apiserver"
	bootconfig "github.com/pole-io/pole-server/bootstrap/config"
	consolebootstrap "github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/limiter"
	limiterapiserver "github.com/pole-io/pole-server/limiter/apiserver"
	apiv2 "github.com/pole-io/specification/source/go/api/v1/traffic_manage/ratelimiter"
)

func TestRunLimiterServerProfileWithoutControlPlaneStore(t *testing.T) {
	port := reserveTCPPort(t)
	loggerPath, err := filepath.Abs("../test/data/bootstrap/pole-log.yaml")
	require.NoError(t, err)
	configPath := filepath.Join(t.TempDir(), "pole-server.yaml")
	configText := fmt.Sprintf(`
bootstrap:
  mode: limiter-server
  logger: %q
limiter:
  registry:
    enable: false
  api-servers:
    - name: grpc
      option:
        ip: 127.0.0.1
        port: %d
  limit:
    node-id: 1
`, loggerPath, port)
	require.NoError(t, os.WriteFile(configPath, []byte(configText), 0o600))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, Options{ConfigPath: configPath})
	}()

	address := fmt.Sprintf("127.0.0.1:%d", port)
	var conn *grpc.ClientConn
	require.Eventually(t, func() bool {
		dialCtx, dialCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer dialCancel()
		var dialErr error
		conn, dialErr = grpc.DialContext(dialCtx, address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock())
		return dialErr == nil
	}, 5*time.Second, 20*time.Millisecond)

	response, err := apiv2.NewRateLimitGRPCV2Client(conn).TimeAdjust(
		context.Background(), &apiv2.TimeAdjustRequest{})
	require.NoError(t, err)
	require.NotZero(t, response.GetServerTimestamp())
	require.NoError(t, conn.Close())

	cancel()
	require.NoError(t, <-errCh)
}

func reserveTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	return port
}

func TestValidateProfileListenersRejectsCrossModuleConflict(t *testing.T) {
	cfg := &bootconfig.Config{
		Bootstrap: bootconfig.Bootstrap{
			Console: consolebootstrap.Config{
				WebServer: consolebootstrap.WebServer{
					ListenIP: "127.0.0.1", ListenPort: 8080,
				},
			},
		},
		Limiter: limiter.Config{
			APIServers: []limiterapiserver.Config{{
				Name: "grpc",
				Option: map[string]interface{}{
					"ip": "127.0.0.1", "port": 8101,
				},
			}},
		},
	}
	controlPlaneEntries := []apiserver.Config{{
		Name: "service-grpc",
		Option: map[string]interface{}{
			"listenIP": "0.0.0.0", "listenPort": 8101,
		},
	}}

	err := validateProfileListeners(bootconfig.StartProfile{
		ControlPlane: true, LimiterServer: true, Console: true,
	}, cfg, controlPlaneEntries)

	require.ErrorContains(t, err, "listener conflict")
	require.ErrorContains(t, err, "control-plane/service-grpc")
	require.ErrorContains(t, err, "limiter-server/grpc")
}

func TestValidateProfileListenersAllowsDistinctPorts(t *testing.T) {
	cfg := &bootconfig.Config{
		Bootstrap: bootconfig.Bootstrap{
			Console: consolebootstrap.Config{
				WebServer: consolebootstrap.WebServer{
					ListenIP: "0.0.0.0", ListenPort: 8080,
				},
			},
		},
		Limiter: limiter.Config{
			APIServers: []limiterapiserver.Config{{
				Name: "grpc",
				Option: map[string]interface{}{
					"ip": "0.0.0.0", "port": 8101,
				},
			}},
		},
	}
	controlPlaneEntries := []apiserver.Config{{
		Name: "service-grpc",
		Option: map[string]interface{}{
			"listenIP": "0.0.0.0", "listenPort": 8091,
		},
	}}

	err := validateProfileListeners(bootconfig.StartProfile{
		ControlPlane: true, LimiterServer: true, Console: true,
	}, cfg, controlPlaneEntries)

	require.NoError(t, err)
}

func TestWaitForTCPListenerReadySupportsLegacyAPIServers(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = listener.Close() }()
	port := uint32(listener.Addr().(*net.TCPAddr).Port)

	accepted := make(chan struct{})
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			_ = conn.Close()
		}
		close(accepted)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, waitForTCPListenerReady(
		ctx, "legacy", "127.0.0.1", port, make(chan error),
	))
	<-accepted
}

func TestReadinessProbeHostPreservesSpecificAddress(t *testing.T) {
	require.Equal(t, "10.0.0.8", readinessProbeHost("10.0.0.8"))
	require.Equal(t, "127.0.0.1", readinessProbeHost("0.0.0.0"))
	require.Equal(t, "::1", readinessProbeHost("[::]"))
}
