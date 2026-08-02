package logger

import (
	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/pluginapi"
)

func Register(registry *pluginapi.Registry) error {
	return apis.RegisterPluginFactory(registry, pluginapi.Descriptor{
		Kind:   pluginapi.KindHistory,
		Name:   PluginName,
		Origin: pluginapi.OriginBuiltin,
	}, func() (apis.Plugin, error) {
		return &HistoryLogger{}, nil
	})
}
