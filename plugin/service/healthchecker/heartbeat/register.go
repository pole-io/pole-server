package heartbeat

import (
	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/pluginapi"
)

func Register(registry *pluginapi.Registry) error {
	return apis.RegisterPluginFactory(registry, pluginapi.Descriptor{
		Kind:   pluginapi.KindHealthCheck,
		Name:   PluginName,
		Origin: pluginapi.OriginBuiltin,
	}, func() (apis.Plugin, error) {
		return &HeartBeatHealthChecker{}, nil
	})
}
