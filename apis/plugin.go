package apis

import (
	"fmt"
	"sync"
)

var (
	pluginSet = make(map[PluginType]map[string]Plugin)
	config    = &Config{}
	once      sync.Once
)

// RegisterPlugin 注册插件
func RegisterPlugin(name string, plugin Plugin) {
	if _, exist := pluginSet[plugin.Type()]; !exist {
		pluginSet[plugin.Type()] = make(map[string]Plugin)
	}
	if _, exist := pluginSet[plugin.Type()][name]; exist {
		panic(fmt.Sprintf("existed plugin: name=%v", name))
	}
	pluginSet[plugin.Type()][plugin.Name()] = plugin
}

func GetPlugin(t PluginType, name string) (Plugin, bool) {
	if _, exist := pluginSet[t]; !exist {
		return nil, false
	}
	if plugin, exist := pluginSet[t][name]; exist {
		return plugin, true
	}
	return nil, false
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
