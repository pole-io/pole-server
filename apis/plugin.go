package apis

import (
	"fmt"
	"sync"
)

var (
	pluginSet = make(map[string]map[string]Plugin)
	config    = &Config{}
	once      sync.Once
)

// RegisterPlugin 注册插件
func RegisterPlugin(name string, plugin Plugin) {
	if _, exist := pluginSet[name]; exist {
		panic(fmt.Sprintf("existed plugin: name=%v", name))
	}
	pluginSet[name][plugin.Name()] = plugin
}

// SetPluginConfig 设置插件配置
func SetPluginConfig(c *Config) {
	config = c
}

// Plugin 通用插件接口
type Plugin interface {
	Name() string
	Initialize(c *ConfigEntry) error
	Destroy() error
}

// ConfigEntry 单个插件配置
type ConfigEntry struct {
	Name   string                 `yaml:"name"`
	Option map[string]interface{} `yaml:"option"`
}

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
