package service

import (
	"testing"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
)

func TestNormalizeActiveHealthCheck(t *testing.T) {
	tcp := NormalizeHealthCheck(&apiservice.HealthCheck{Type: apiservice.HealthCheck_TCP})
	if tcp.GetTcp().GetInterval() != DefaultHealthCheckInterval {
		t.Fatalf("tcp interval = %d", tcp.GetTcp().GetInterval())
	}
	httpCheck := NormalizeHealthCheck(&apiservice.HealthCheck{
		Type: apiservice.HealthCheck_HTTP,
		Http: &apiservice.HttpHealthCheck{Interval: 10, Path: "health"},
	})
	if httpCheck.GetHttp().GetInterval() != 10 || httpCheck.GetHttp().GetPath() != "/health" {
		t.Fatalf("unexpected http check: %+v", httpCheck.GetHttp())
	}
	if HealthCheckExpireDuration(tcp) != DefaultHealthCheckInterval {
		t.Fatalf("tcp expire duration = %d", HealthCheckExpireDuration(tcp))
	}
	heartbeat := NormalizeHealthCheck(&apiservice.HealthCheck{
		Heartbeat: &apiservice.HeartbeatHealthCheck{Ttl: 5},
	})
	if HealthCheckExpireDuration(heartbeat) != 15 {
		t.Fatalf("heartbeat expire duration = %d", HealthCheckExpireDuration(heartbeat))
	}
}

func TestActiveHealthCheckStoreRoundTrip(t *testing.T) {
	created := CreateInstanceModel("service-id", &apiservice.Instance{
		Host:              "127.0.0.1",
		Port:              8080,
		EnableHealthCheck: true,
		HealthCheck: &apiservice.HealthCheck{
			Type: apiservice.HealthCheck_HTTP,
			Http: &apiservice.HttpHealthCheck{Interval: 7, Path: "ready"},
		},
	})
	if !created.EnableHealthCheck() || created.HealthCheck().GetType() != apiservice.HealthCheck_HTTP {
		t.Fatalf("unexpected created health check: %+v", created.HealthCheck())
	}
	if created.Proto.Metadata[types.MetadataInternalMetaHealthCheckPath] != "/ready" {
		t.Fatalf("http path metadata = %q", created.Proto.Metadata[types.MetadataInternalMetaHealthCheckPath])
	}
	restored := Store2Instance(&InstanceStore{
		ID:                "instance-id",
		ServiceID:         "service-id",
		Host:              "127.0.0.1",
		Port:              8080,
		EnableHealthCheck: 1,
		CheckType:         int32(apiservice.HealthCheck_HTTP),
		TTL:               7,
		Meta:              created.Proto.Metadata,
	})
	if restored.HealthCheck().GetHttp().GetInterval() != 7 || restored.HealthCheck().GetHttp().GetPath() != "/ready" {
		t.Fatalf("unexpected restored health check: %+v", restored.HealthCheck())
	}
}
