package limiter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigValidateRequiresGRPCServer(t *testing.T) {
	cfg := Config{
		APIServers: []APIServerConfig{{
			Name: "http",
			Option: map[string]interface{}{
				"ip":   "127.0.0.1",
				"port": 0,
			},
		}},
		Limit: validLimitConfig(),
	}

	require.EqualError(t, cfg.Validate(), "limiter grpc api server is required")
}

func TestConfigValidateRejectsIncompleteRegistry(t *testing.T) {
	cfg := testConfig()
	cfg.Registry.Enable = true

	require.EqualError(t, cfg.Validate(),
		"limiter registry control-plane-address is required")
}

func TestConfigValidateRejectsDuplicateServer(t *testing.T) {
	cfg := testConfig()
	cfg.APIServers = append(cfg.APIServers, cfg.APIServers[0])

	require.EqualError(t, cfg.Validate(), `limiter api server "grpc" is duplicated`)
}
