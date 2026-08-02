package probe

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/service/healthcheck"
	"github.com/pole-io/pole-server/pluginapi"
)

const (
	TCPPluginName  = "tcp"
	HTTPPluginName = "http"
	defaultTimeout = 3 * time.Second
)

var (
	_ healthcheck.HealthChecker = (*Checker)(nil)
)

func Register(registry *pluginapi.Registry) error {
	if err := registerChecker(registry, TCPPluginName, healthcheck.HealthCheckerDetectTCP); err != nil {
		return err
	}
	return registerChecker(registry, HTTPPluginName, healthcheck.HealthCheckerDetectHTTP)
}

func registerChecker(registry *pluginapi.Registry, name string, checkType healthcheck.HealthCheckType) error {
	return apis.RegisterPluginFactory(registry, pluginapi.Descriptor{
		Kind:   pluginapi.KindHealthCheck,
		Name:   name,
		Origin: pluginapi.OriginBuiltin,
	}, func() (apis.Plugin, error) {
		return newChecker(name, checkType), nil
	})
}

// Checker executes a control-plane initiated TCP or HTTP probe.
type Checker struct {
	name           string
	checkType      healthcheck.HealthCheckType
	timeout        time.Duration
	suspendTimeSec int64
}

func newChecker(name string, checkType healthcheck.HealthCheckType) *Checker {
	return &Checker{name: name, checkType: checkType, timeout: defaultTimeout}
}

func (c *Checker) Name() string { return c.name }

func (c *Checker) Initialize(entry *apis.ConfigEntry) error {
	c.timeout = defaultTimeout
	if entry == nil || entry.Option == nil {
		return nil
	}
	raw, ok := entry.Option["timeout"]
	if !ok {
		return nil
	}
	duration, err := time.ParseDuration(fmt.Sprint(raw))
	if err != nil || duration <= 0 {
		return fmt.Errorf("invalid probe timeout %v", raw)
	}
	c.timeout = duration
	return nil
}

func (c *Checker) Destroy() error { return nil }

func (c *Checker) Type() apis.PluginType { return apis.PluginTypeHealthCheck }

func (c *Checker) CheckType() healthcheck.HealthCheckType { return c.checkType }

func (c *Checker) Report(context.Context, *healthcheck.ReportRequest) error { return nil }

func (c *Checker) Check(request *healthcheck.CheckRequest) (*healthcheck.CheckResponse, error) {
	healthy := false
	switch c.checkType {
	case healthcheck.HealthCheckerDetectTCP:
		healthy = c.checkTCP(request.Host, request.Port)
	case healthcheck.HealthCheckerDetectHTTP:
		healthy = c.checkHTTP(request.Host, request.Port, request.Path)
	default:
		return nil, fmt.Errorf("unsupported probe type %d", c.checkType)
	}
	return &healthcheck.CheckResponse{
		Healthy:       healthy,
		StayUnchanged: request.Healthy == healthy,
		Regular:       true,
	}, nil
}

func (c *Checker) checkTCP(host string, port uint32) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(int(port))), c.timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func (c *Checker) checkHTTP(host string, port uint32, path string) bool {
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	target := url.URL{Scheme: "http", Host: net.JoinHostPort(host, strconv.Itoa(int(port))), Path: path}
	client := &http.Client{Timeout: c.timeout}
	response, err := client.Get(target.String())
	if err != nil {
		return false
	}
	_ = response.Body.Close()
	return response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusBadRequest
}

func (c *Checker) Query(context.Context, *healthcheck.QueryRequest) (*healthcheck.QueryResponse, error) {
	return &healthcheck.QueryResponse{}, nil
}

func (c *Checker) BatchQuery(ctx context.Context, request *healthcheck.BatchQueryRequest) (*healthcheck.BatchQueryResponse, error) {
	responses := make([]*healthcheck.QueryResponse, 0, len(request.Requests))
	for _, item := range request.Requests {
		response, err := c.Query(ctx, item)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return &healthcheck.BatchQueryResponse{Responses: responses}, nil
}

func (c *Checker) Suspend() { atomic.StoreInt64(&c.suspendTimeSec, time.Now().Unix()) }

func (c *Checker) SuspendTimeSec() int64 { return atomic.LoadInt64(&c.suspendTimeSec) }

func (c *Checker) Delete(context.Context, string) error { return nil }

func (c *Checker) DebugHandlers() []types.DebugHandler { return nil }
