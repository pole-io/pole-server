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
	metricstypes "github.com/pole-io/pole-server/apis/pkg/types/metrics"
	"github.com/pole-io/pole-server/pkg/common/metrics"
)

func newConfigMetricHandle() *configMetricHandle {
	return &configMetricHandle{}
}

type configMetricHandle struct {
}

func (h *configMetricHandle) handle(ms []metricstypes.ConfigMetrics) {
	for i := range ms {
		m := ms[i]
		switch m.Type {
		case metricstypes.ConfigGroupMetric:
			metrics.GetConfigGroupTotal().With(m.Labels).Set(float64(m.Total))
		case metricstypes.FileMetric:
			metrics.GetConfigFileTotal().With(m.Labels).Set(float64(m.Total))
		case metricstypes.ReleaseFileMetric:
			metrics.GetReleaseConfigFileTotal().With(m.Labels).Set(float64(m.Total))
		}
	}
}
