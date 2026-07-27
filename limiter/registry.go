package limiter

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/pole-io/pole-server/limiter/apiserver"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/version"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	namingapi "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

const (
	registryHeartbeatInterval = 5 * time.Second
	registryRequestTimeout    = time.Second
	registryRequestIDHeader   = "request-id"
)

type registryClient struct {
	cfg       RegistryConfig
	instances []*namingapi.Instance
	rid       atomic.Uint64
	mu        sync.Mutex
}

func newRegistryClient(cfg RegistryConfig) *registryClient {
	return &registryClient{cfg: cfg}
}

func (r *registryClient) Register(ctx context.Context, servers []apiserver.APIServer) error {
	host, err := r.resolveAdvertisedHost(ctx)
	if err != nil {
		return err
	}
	instances := make([]*namingapi.Instance, 0, len(servers))
	for _, server := range servers {
		instance := &namingapi.Instance{
			Namespace: r.cfg.Namespace,
			Service:   r.cfg.Name,
			Host:      host,
			Port:      server.GetPort(),
			Protocol:  server.GetProtocol(),
			Version:   version.Get(),
			Metadata:  map[string]string{"build-revision": version.GetRevision()},
		}
		if r.cfg.HealthCheckEnable {
			instance.EnableHealthCheck = true
			instance.HealthCheck = &namingapi.HealthCheck{
				Type: namingapi.HealthCheck_HEARTBEAT,
				Heartbeat: &namingapi.HeartbeatHealthCheck{
					Ttl: uint32(registryHeartbeatInterval / time.Second),
				},
			}
		}
		instances = append(instances, instance)
	}

	registered := make([]*namingapi.Instance, 0, len(instances))
	err = r.withClient(ctx, func(client namingapi.DiscoverGRPCClient) error {
		for _, instance := range instances {
			requestCtx, cancel := r.requestContext(ctx, registryRequestTimeout)
			response, registerErr := client.RegisterInstance(requestCtx, instance)
			cancel()
			if registerErr != nil {
				return fmt.Errorf("register limiter instance %s:%d: %w",
					instance.GetHost(), instance.GetPort(), registerErr)
			}
			registeredInstance, decodeErr := decodeRegisteredInstance(response)
			if decodeErr != nil {
				return decodeErr
			}
			registered = append(registered, registeredInstance)
		}
		return nil
	})
	if err != nil {
		r.instances = registered
		_ = r.Deregister(ctx)
		return err
	}
	r.mu.Lock()
	r.instances = registered
	r.mu.Unlock()
	go r.heartbeatLoop(ctx)
	return nil
}

func (r *registryClient) Deregister(ctx context.Context) error {
	r.mu.Lock()
	instances := append([]*namingapi.Instance(nil), r.instances...)
	r.instances = nil
	r.mu.Unlock()
	if len(instances) == 0 {
		return nil
	}
	return r.withClient(ctx, func(client namingapi.DiscoverGRPCClient) error {
		for _, instance := range instances {
			requestCtx, cancel := r.requestContext(ctx, registryRequestTimeout)
			response, err := client.DeregisterInstance(requestCtx, instance)
			cancel()
			if err != nil {
				return fmt.Errorf("deregister limiter instance %s:%d: %w",
					instance.GetHost(), instance.GetPort(), err)
			}
			if err := checkPoleResponse("deregister limiter instance", response); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *registryClient) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(registryHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.sendHeartbeats(ctx); err != nil {
				log.Warnf("send limiter heartbeat: %v", err)
			}
		}
	}
}

func (r *registryClient) sendHeartbeats(ctx context.Context) error {
	r.mu.Lock()
	instances := append([]*namingapi.Instance(nil), r.instances...)
	r.mu.Unlock()
	if len(instances) == 0 {
		return nil
	}
	return r.withClient(ctx, func(client namingapi.DiscoverGRPCClient) error {
		requestCtx, cancel := r.requestContext(ctx, registryRequestTimeout)
		defer cancel()
		stream, err := client.Heartbeat(requestCtx)
		if err != nil {
			return err
		}
		defer func() { _ = stream.CloseSend() }()
		heartbeats := make([]*namingapi.InstanceHeartbeat, 0, len(instances))
		for _, instance := range instances {
			heartbeats = append(heartbeats, &namingapi.InstanceHeartbeat{
				InstanceId: instance.GetId(),
				Service:    instance.GetService(),
				Namespace:  instance.GetNamespace(),
				Host:       instance.GetHost(),
				Port:       instance.GetPort(),
			})
		}
		if err := stream.Send(&namingapi.HeartbeatsRequest{Heartbeats: heartbeats}); err != nil {
			return err
		}
		_, err = stream.Recv()
		return err
	})
}

func (r *registryClient) withClient(ctx context.Context,
	handle func(namingapi.DiscoverGRPCClient) error) error {
	conn, err := grpc.DialContext(ctx, r.cfg.ControlPlaneAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	return handle(namingapi.NewDiscoverGRPCClient(conn))
}

func (r *registryClient) requestContext(parent context.Context,
	timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	headers := map[string]string{
		registryRequestIDHeader: fmt.Sprintf("%s_%d", r.cfg.Name, r.rid.Add(1)),
	}
	if r.cfg.Token != "" {
		headers["authorization"] = r.cfg.Token
	}
	return metadata.NewOutgoingContext(ctx, metadata.New(headers)), cancel
}

func (r *registryClient) resolveAdvertisedHost(ctx context.Context) (string, error) {
	if host := strings.TrimSpace(r.cfg.AdvertisedHost); host != "" {
		return host, nil
	}
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", r.cfg.ControlPlaneAddress)
	if err != nil {
		return "", fmt.Errorf("derive limiter advertised host via %s: %w",
			r.cfg.ControlPlaneAddress, err)
	}
	defer func() { _ = conn.Close() }()
	host, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return "", fmt.Errorf("parse limiter local address %s: %w", conn.LocalAddr(), err)
	}
	return host, nil
}

func decodeRegisteredInstance(response *apimodel.Response) (*namingapi.Instance, error) {
	if err := checkPoleResponse("register limiter instance", response); err != nil {
		return nil, err
	}
	if response.GetData() == nil {
		return nil, fmt.Errorf("register limiter instance response does not contain instance data")
	}
	instance := &namingapi.Instance{}
	if err := response.GetData().UnmarshalTo(instance); err != nil {
		return nil, fmt.Errorf("decode registered limiter instance: %w", err)
	}
	return instance, nil
}

func checkPoleResponse(action string, response *apimodel.Response) error {
	if response == nil {
		return fmt.Errorf("%s failed: empty response", action)
	}
	if response.GetCode() != uint32(apimodel.Code_ExecuteSuccess) {
		return fmt.Errorf("%s failed: code %d, info %s",
			action, response.GetCode(), response.GetInfo())
	}
	return nil
}

func endpointAddress(host string, port uint32) string {
	return net.JoinHostPort(host, strconv.FormatUint(uint64(port), 10))
}
