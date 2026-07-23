package healthcheck

import (
	"testing"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

func TestGetExpireDurationSecByHealthCheckType(t *testing.T) {
	tests := []struct {
		name  string
		check *apiservice.HealthCheck
		want  uint32
	}{
		{
			name: "heartbeat",
			check: &apiservice.HealthCheck{Type: apiservice.HealthCheck_HEARTBEAT,
				Heartbeat: &apiservice.HeartbeatHealthCheck{Ttl: 5}},
			want: 15,
		},
		{
			name: "tcp",
			check: &apiservice.HealthCheck{Type: apiservice.HealthCheck_TCP,
				Tcp: &apiservice.TcpHealthCheck{Interval: 5}},
			want: 5,
		},
		{
			name: "http",
			check: &apiservice.HealthCheck{Type: apiservice.HealthCheck_HTTP,
				Http: &apiservice.HttpHealthCheck{Interval: 8, Path: "/health"}},
			want: 8,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			instance := &apiservice.Instance{HealthCheck: test.check}
			if got := getExpireDurationSec(instance); got != test.want {
				t.Fatalf("expire duration = %d, want %d", got, test.want)
			}
		})
	}
}
