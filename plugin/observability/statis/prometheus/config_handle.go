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
	configGroupTotal       otelmetric.Int64Gauge
	configFileTotal        otelmetric.Int64Gauge
	releaseConfigFileTotal otelmetric.Int64Gauge
)

func newConfigMetricHandle() *configMetricHandle {
	registerConfigFileMetrics()
	return &configMetricHandle{}
}

type configMetricHandle struct {
}

func (h *configMetricHandle) handle(ms []metricstypes.ConfigMetrics) {
	for i := range ms {
		m := ms[i]
		switch m.Type {
		case metricstypes.ConfigGroupMetric:
			ReportConfigGroupTotal(m.Total, m.Labels)
		case metricstypes.FileMetric:
			ReportConfigFileTotal(m.Total, m.Labels)
		case metricstypes.ReleaseFileMetric:
			ReportReleaseConfigFileTotal(m.Total, m.Labels)
		}
	}
}

func registerConfigFileMetrics() {
	var err error

	configGroupTotal, err = otel.Meter().Int64Gauge("config_group_count",
		metric.WithDescription("polaris config group total number"),
	)
	if err != nil {
		panic("register config group total metric failed: " + err.Error())
	}

	configFileTotal, err = otel.Meter().Int64Gauge("config_file_count",
		metric.WithDescription("total number of config_file each config group"),
	)
	if err != nil {
		panic("register config file total metric failed: " + err.Error())
	}

	releaseConfigFileTotal, err = otel.Meter().Int64Gauge("config_release_file_count",
		metric.WithDescription("total number of config_release_file each config group"),
	)
	if err != nil {
		panic("register config release file total metric failed: " + err.Error())
	}
}

func ReportConfigGroupTotal(v int64, labels []attribute.KeyValue) {
	if configGroupTotal != nil {
		configGroupTotal.Record(context.Background(), v, metric.WithAttributes(labels...))
	}
}

func ReportConfigFileTotal(v int64, labels []attribute.KeyValue) {
	if configFileTotal != nil {
		configFileTotal.Record(context.Background(), v, metric.WithAttributes(labels...))
	}
}

func ReportReleaseConfigFileTotal(v int64, labels []attribute.KeyValue) {
	if releaseConfigFileTotal != nil {
		releaseConfigFileTotal.Record(context.Background(), v, metric.WithAttributes(labels...))
	}
}
