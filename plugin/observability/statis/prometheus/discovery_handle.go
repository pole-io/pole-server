/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package prometheus

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	otelmetric "go.opentelemetry.io/otel/metric"

	metricstypes "github.com/pole-io/pole-server/apis/pkg/types/metrics"
	"github.com/pole-io/pole-server/pkg/common/otel"
)

var (
	clientInstanceTotal   otelmetric.Int64Gauge
	serviceCount          otelmetric.Int64Gauge
	serviceOnlineCount    otelmetric.Int64Gauge
	serviceAbnormalCount  otelmetric.Int64Gauge
	serviceOfflineCount   otelmetric.Int64Gauge
	instanceCount         otelmetric.Int64Gauge
	instanceOnlineCount   otelmetric.Int64Gauge
	instanceAbnormalCount otelmetric.Int64Gauge
	instanceIsolateCount  otelmetric.Int64Gauge
)

func newDiscoveryMetricHandle() *discoveryMetricHandle {
	registerDiscoveryMetrics()
	return &discoveryMetricHandle{}
}

type discoveryMetricHandle struct {
}

func (h *discoveryMetricHandle) handle(ms []metricstypes.DiscoveryMetric) {
	for i := range ms {
		m := ms[i]
		switch m.Type {
		case metricstypes.ServiceMetrics:
			ReportServiceCount(m.Total, m.Labels)
			ReportServiceAbnormalCount(m.Abnormal, m.Labels)
			ReportServiceOfflineCount(m.Offline, m.Labels)
			ReportServiceOnlineCount(m.Online, m.Labels)
		case metricstypes.InstanceMetrics:
			ReportInstanceCount(m.Total, m.Labels)
			ReportInstanceAbnormalCount(m.Abnormal, m.Labels)
			ReportInstanceIsolateCount(m.Isolate, m.Labels)
			ReportInstanceOnlineCount(m.Online, m.Labels)
		case metricstypes.ClientMetrics:
			ReportClientInstanceTotal(m.Total)
		}
	}
}

func registerDiscoveryMetrics() {
	var err error

	clientInstanceTotal, err = otel.Meter().Int64Gauge("client_total",
		metric.WithDescription("polaris client instance total number"),
	)
	if err != nil {
		panic("register client instance total metric failed: " + err.Error())
	}

	serviceCount, err = otel.Meter().Int64Gauge("service_count",
		metric.WithDescription("service total number"),
	)
	if err != nil {
		panic("register service count metric failed: " + err.Error())
	}

	serviceOnlineCount, err = otel.Meter().Int64Gauge("service_online_count",
		metric.WithDescription("total number of service status is online"),
	)
	if err != nil {
		panic("register service online count metric failed: " + err.Error())
	}

	serviceAbnormalCount, err = otel.Meter().Int64Gauge("service_abnormal_count",
		metric.WithDescription("total number of service status is abnormal"),
	)
	if err != nil {
		panic("register service abnormal count metric failed: " + err.Error())
	}

	serviceOfflineCount, err = otel.Meter().Int64Gauge("service_offline_count",
		metric.WithDescription("total number of service status is offline"),
	)
	if err != nil {
		panic("register service offline count metric failed: " + err.Error())
	}

	instanceCount, err = otel.Meter().Int64Gauge("instance_count",
		metric.WithDescription("instance total number"),
	)
	if err != nil {
		panic("register instance count metric failed: " + err.Error())
	}

	instanceOnlineCount, err = otel.Meter().Int64Gauge("instance_online_count",
		metric.WithDescription("total number of instance status is health"),
	)

	instanceAbnormalCount, err = otel.Meter().Int64Gauge("instance_abnormal_count",
		metric.WithDescription("total number of instance status is unhealth"),
	)
	if err != nil {
		panic("register instance abnormal count metric failed: " + err.Error())
	}

	instanceIsolateCount, err = otel.Meter().Int64Gauge("instance_isolate_count",
		metric.WithDescription("total number of instance status is isolate"),
	)
	if err != nil {
		panic("register instance isolate count metric failed: " + err.Error())
	}
}

func ReportClientInstanceTotal(v int64) {
	if clientInstanceTotal == nil {
		return
	}
	clientInstanceTotal.Record(context.Background(), v)
}

func ReportServiceCount(v int64, labels []attribute.KeyValue) {
	if serviceCount == nil {
		return
	}
	serviceCount.Record(context.Background(), v, metric.WithAttributes(labels...))
}

func ReportServiceOnlineCount(v int64, labels []attribute.KeyValue) {
	if serviceOnlineCount == nil {
		return
	}
	serviceOnlineCount.Record(context.Background(), v, metric.WithAttributes(labels...))
}

func ReportServiceOfflineCount(v int64, labels []attribute.KeyValue) {
	if serviceOfflineCount == nil {
		return
	}
	serviceOfflineCount.Record(context.Background(), v, metric.WithAttributes(labels...))
}

func ReportServiceAbnormalCount(v int64, labels []attribute.KeyValue) {
	if serviceAbnormalCount == nil {
		return
	}
	serviceAbnormalCount.Record(context.Background(), v, metric.WithAttributes(labels...))
}

func ReportInstanceCount(v int64, labels []attribute.KeyValue) {
	if instanceCount == nil {
		return
	}
	instanceCount.Record(context.Background(), v, metric.WithAttributes(labels...))
}

func ReportInstanceOnlineCount(v int64, labels []attribute.KeyValue) {
	if instanceOnlineCount == nil {
		return
	}
	instanceOnlineCount.Record(context.Background(), v, metric.WithAttributes(labels...))
}

func ReportInstanceIsolateCount(v int64, labels []attribute.KeyValue) {
	if instanceIsolateCount == nil {
		return
	}
	instanceIsolateCount.Record(context.Background(), v, metric.WithAttributes(labels...))
}

func ReportInstanceAbnormalCount(v int64, labels []attribute.KeyValue) {
	if instanceAbnormalCount == nil {
		return
	}
	instanceAbnormalCount.Record(context.Background(), v, metric.WithAttributes(labels...))
}
