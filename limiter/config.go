package limiter

import (
	"fmt"
	"strings"

	"github.com/pole-io/pole-server/limiter/apiserver"
	limitconfig "github.com/pole-io/pole-server/limiter/pkg/config"
	"github.com/pole-io/pole-server/limiter/plugin"
)

// Config 描述 Limiter 运行模块配置。
type Config struct {
	Registry   RegistryConfig     `yaml:"registry"`
	APIServers []apiserver.Config `yaml:"api-servers"`
	Limit      limitconfig.Config `yaml:"limit"`
	Plugin     plugin.Config      `yaml:"plugin"`
}

// RegistryConfig 描述 Limiter 在 Control Plane 中的服务身份。
type RegistryConfig struct {
	Enable              bool   `yaml:"enable"`
	ControlPlaneAddress string `yaml:"control-plane-address"`
	Name                string `yaml:"name"`
	Namespace           string `yaml:"namespace"`
	AdvertisedHost      string `yaml:"advertised-host"`
	Token               string `yaml:"token"`
	HealthCheckEnable   bool   `yaml:"health-check-enable"`
}

// Validate 在获得任何运行资源前验证配置。
func (c *Config) Validate() error {
	if len(c.APIServers) == 0 {
		return fmt.Errorf("limiter api-servers is required")
	}
	seen := make(map[string]struct{}, len(c.APIServers))
	hasGRPC := false
	for i, entry := range c.APIServers {
		name := strings.ToLower(strings.TrimSpace(entry.Name))
		switch name {
		case "grpc":
			hasGRPC = true
		case "http":
		default:
			return fmt.Errorf("limiter api-servers[%d].name %q is unsupported", i, entry.Name)
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("limiter api server %q is duplicated", name)
		}
		seen[name] = struct{}{}
		if _, _, err := apiserver.ParseListenOption(entry.Option); err != nil {
			return fmt.Errorf("limiter api server %q: %w", name, err)
		}
	}
	if !hasGRPC {
		return fmt.Errorf("limiter grpc api server is required")
	}
	if _, err := limitconfig.ParseConfig(&c.Limit); err != nil {
		return fmt.Errorf("limiter runtime config: %w", err)
	}
	if c.Registry.Enable {
		if strings.TrimSpace(c.Registry.ControlPlaneAddress) == "" {
			return fmt.Errorf("limiter registry control-plane-address is required")
		}
		if strings.TrimSpace(c.Registry.Namespace) == "" {
			return fmt.Errorf("limiter registry namespace is required")
		}
		if strings.TrimSpace(c.Registry.Name) == "" {
			return fmt.Errorf("limiter registry name is required")
		}
	}
	if c.Plugin.Statistics != nil {
		switch strings.ToLower(strings.TrimSpace(c.Plugin.Statistics.Name)) {
		case "echo", "file":
		default:
			return fmt.Errorf("limiter statistics plugin %q is unsupported", c.Plugin.Statistics.Name)
		}
	}
	return nil
}
