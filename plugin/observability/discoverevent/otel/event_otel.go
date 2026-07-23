package otel

import (
	"context"
	"time"

	"github.com/pole-io/pole-server/apis"
	obsevent "github.com/pole-io/pole-server/apis/observability/event"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	commonlog "github.com/pole-io/pole-server/pkg/common/log"
	commonotel "github.com/pole-io/pole-server/pkg/common/otel"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

const PluginName = "otel"

var log = commonlog.RegisterScope("DiscoverEventOtel", "", 0)

func init() {
	worker := &DiscoverEventWorker{}
	apis.RegisterPlugin(worker.Name(), worker)
}

type DiscoverEventWorker struct {
	exporter *commonotel.LogExporter
}

func (d *DiscoverEventWorker) Name() string {
	return PluginName
}

func (d *DiscoverEventWorker) Initialize(conf *apis.ConfigEntry) error {
	config, err := commonotel.ConfigFromOptions(confOption(conf))
	if err != nil {
		log.Errorf("load otel discover event config failed, disable otel discover event exporter: %v", err)
		return nil
	}
	exporter, err := commonotel.NewLogExporter(&config)
	if err != nil {
		log.Errorf("initialize otel discover event exporter failed, disable otel discover event exporter: %v", err)
		return nil
	}
	d.exporter = exporter
	return nil
}

func (d *DiscoverEventWorker) Type() apis.PluginType {
	return apis.PluginTypeDiscoverEvent
}

func (d *DiscoverEventWorker) Destroy() error {
	if d.exporter == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return d.exporter.Shutdown(ctx)
}

func (d *DiscoverEventWorker) PublishEvent(discoverEvent obsevent.DiscoverEvent) {
	if d.exporter == nil || discoverEvent == nil {
		return
	}
	if _, ok := subscribedEvents[discoverEvent.Event()]; !ok {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("emit otel discover event failed: %v", err)
		}
	}()

	namespace, service := eventNamespaceService(discoverEvent)
	resource := safeResource(discoverEvent)
	happenTime := discoverEvent.HappenTime()
	if happenTime.IsZero() {
		happenTime = time.Now()
	}
	attrs := map[string]string{
		"event.name":                discoverEvent.Event(),
		"pole.event.kind":           "service",
		"pole.discovery.event":      discoverEvent.Event(),
		"pole.namespace":            namespace,
		"pole.service.name":         service,
		"pole.resource.id":          discoverEvent.ID(),
		"pole.resource.name":        resource,
		"pole.server.address":       utils.LocalHost,
		"pole.control_plane.signal": "discover_event",
	}
	if instanceID := eventInstanceID(discoverEvent); instanceID != "" {
		attrs["pole.service.instance.id"] = instanceID
	}
	d.exporter.Emit(commonotel.LogRecord{
		Timestamp:    happenTime,
		SeverityText: "INFO",
		Body:         resource,
		Attributes:   attrs,
	})
}

var subscribedEvents = map[string]struct{}{
	string(svctypes.EventInstanceCloseIsolate):          {},
	string(svctypes.EventInstanceOpenIsolate):           {},
	string(svctypes.EventInstanceOffline):               {},
	string(svctypes.EventInstanceOnline):                {},
	string(svctypes.EventInstanceTurnHealth):            {},
	string(svctypes.EventInstanceTurnUnHealth):          {},
	string(svctypes.EventServiceOpenEmptyPushProtect):   {},
	string(svctypes.EventServiceCloseEmptyPushProtect):  {},
	string(svctypes.EventServiceExpireEmptyPushProtect): {},
}

func eventNamespaceService(discoverEvent obsevent.DiscoverEvent) (string, string) {
	switch typed := discoverEvent.(type) {
	case *svctypes.ServiceEvent:
		return typed.Namespace, typed.Service
	case *svctypes.InstanceEvent:
		return typed.Namespace, typed.Service
	default:
		return "", ""
	}
}

func eventInstanceID(discoverEvent obsevent.DiscoverEvent) string {
	if typed, ok := discoverEvent.(*svctypes.InstanceEvent); ok && typed.Instance != nil {
		return typed.Instance.GetId()
	}
	return ""
}

func safeResource(discoverEvent obsevent.DiscoverEvent) (resource string) {
	defer func() {
		if recover() != nil {
			resource = discoverEvent.ID()
		}
	}()
	return discoverEvent.Resource()
}

func confOption(conf *apis.ConfigEntry) map[string]interface{} {
	if conf == nil || conf.Option == nil {
		return map[string]interface{}{}
	}
	return conf.Option
}
