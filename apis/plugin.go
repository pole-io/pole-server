package apis

import (
	"fmt"

	"github.com/pole-io/pole-server/pluginapi"
)

var (
	config = &Config{}
)

// RegisterPlugin 注册插件
func RegisterPlugin(name string, plugin Plugin) {
	if plugin == nil {
		panic("plugin is nil")
	}
	if name != plugin.Name() {
		panic(fmt.Sprintf("plugin registration name mismatch: registered=%s actual=%s", name, plugin.Name()))
	}
	err := RegisterPluginFactory(pluginapi.DefaultRegistry(), pluginapi.Descriptor{
		Kind:   PluginKind(plugin.Type()),
		Name:   name,
		Origin: pluginapi.OriginLegacy,
	}, func() (Plugin, error) {
		return plugin, nil
	})
	if err != nil {
		panic(err)
	}
}

type PluginFactory func() (Plugin, error)

func RegisterPluginFactory(registry *pluginapi.Registry, descriptor pluginapi.Descriptor,
	factory PluginFactory) error {
	if factory == nil {
		return fmt.Errorf("plugin factory is nil: kind=%s name=%s", descriptor.Kind, descriptor.Name)
	}
	return registry.Register(descriptor, func() (any, error) {
		plugin, err := factory()
		if err != nil {
			return nil, err
		}
		if plugin == nil {
			return nil, fmt.Errorf("plugin factory returned nil: kind=%s name=%s", descriptor.Kind, descriptor.Name)
		}
		if descriptor.Kind != PluginKind(plugin.Type()) {
			return nil, fmt.Errorf("plugin kind mismatch: registered=%s actual=%s",
				descriptor.Kind, PluginKind(plugin.Type()))
		}
		if descriptor.Name != plugin.Name() {
			return nil, fmt.Errorf("plugin name mismatch: registered=%s actual=%s",
				descriptor.Name, plugin.Name())
		}
		return plugin, nil
	})
}

func GetPlugin(t PluginType, name string) (Plugin, bool) {
	plugin, err := ResolvePlugin(t, name)
	if err != nil {
		return nil, false
	}
	return plugin, true
}

func ResolvePlugin(pluginType PluginType, name string) (Plugin, error) {
	return pluginapi.ResolveAs[Plugin](pluginapi.ActiveRegistry(), PluginKind(pluginType), name)
}

// SetPluginConfig 设置插件配置
func SetPluginConfig(c *Config) {
	config = c
}

func GetPluginConfig() *Config {
	return config
}

// Plugin 通用插件接口
type Plugin interface {
	Name() string
	Initialize(c *ConfigEntry) error
	Destroy() error
	Type() PluginType
}

// ConfigEntry 单个插件配置
type ConfigEntry struct {
	Name   string                 `yaml:"name"`
	Option map[string]interface{} `yaml:"option"`
}

type PluginType int32

func PluginKind(pluginType PluginType) pluginapi.Kind {
	switch pluginType {
	case PluginTypeStatis:
		return pluginapi.KindStatis
	case PluginTypeHistory:
		return pluginapi.KindHistory
	case PluginTypeDiscoverEvent:
		return pluginapi.KindDiscoverEvent
	case PluginTypeRateLimit:
		return pluginapi.KindRateLimit
	case PluginTypeWhitelist:
		return pluginapi.KindWhitelist
	case PluginTypeResourceAuth:
		return pluginapi.KindResourceAuth
	case PluginTypeCMDB:
		return pluginapi.KindCMDB
	case PluginTypeApiServer:
		return pluginapi.KindAPIServer
	case PluginTypeCrypto:
		return pluginapi.KindCrypto
	case PluginTypeStore:
		return pluginapi.KindStore
	case PluginTypeHealthCheck:
		return pluginapi.KindHealthCheck
	default:
		return pluginapi.Kind(fmt.Sprintf("plugin/%d", pluginType))
	}
}

const (
	_ PluginType = iota
	//  -------- observability plugins --------
	// PluginTypeStatis 统计插件
	PluginTypeStatis
	// PluginTypeHistory 历史记录插件
	PluginTypeHistory
	// PluginTypeDiscoverEvent 发现事件插件
	PluginTypeDiscoverEvent
	// -------- observability plugins --------

	// -------- access_control plugins --------
	// PluginTypeRateLimit 限流插件
	PluginTypeRateLimit
	// PluginTypeWhitelist 白名单插件
	PluginTypeWhitelist
	// PluginTypeResourceAuth 资源鉴权插件
	PluginTypeResourceAuth
	// -------- access_control plugins --------

	// PluginTypeCMDB CMDB插件
	PluginTypeCMDB
	// PluginTypeApiServer API服务插件
	PluginTypeApiServer
	// PluginTypeCrypto 加密插件
	PluginTypeCrypto
	// PluginTypeStore 存储插件
	PluginTypeStore

	//
	PluginTypeHealthCheck
)

// Config 插件配置
type Config struct {
	CMDB                 ConfigEntry      `yaml:"cmdb"`
	RateLimit            ConfigEntry      `yaml:"ratelimit"`
	History              PluginChanConfig `yaml:"history"`
	Statis               PluginChanConfig `yaml:"statis"`
	DiscoverStatis       ConfigEntry      `yaml:"discoverStatis"`
	ParsePassword        ConfigEntry      `yaml:"parsePassword"`
	Whitelist            ConfigEntry      `yaml:"whitelist"`
	MeshResourceValidate ConfigEntry      `yaml:"meshResourceValidate"`
	DiscoverEvent        PluginChanConfig `yaml:"discoverEvent"`
	Crypto               PluginChanConfig `yaml:"crypto"`
}

// PluginChanConfig 插件执行链配置
type PluginChanConfig struct {
	Name    string                 `yaml:"name"`
	Option  map[string]interface{} `yaml:"option"`
	Entries []ConfigEntry          `yaml:"entries"`
}
