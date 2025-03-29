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

package config

import (
	"github.com/pole-io/pole-server/common/metrics"
	"github.com/pole-io/pole-server/common/model"
	"github.com/pole-io/pole-server/common/utils"
	"github.com/pole-io/pole-server/plugin"
)

func (fc *configGroupCache) reportMetricsInfo() {
	fc.name2groups.Range(func(ns string, val *utils.SyncMap[string, *model.ConfigFileGroup]) {
		count := val.Len()
		reportValue := metrics.ConfigMetrics{
			Type:    metrics.ConfigGroupMetric,
			Total:   int64(count),
			Release: 0,
			Labels: map[string]string{
				metrics.LabelNamespace: ns,
			},
		}
		plugin.GetStatis().ReportConfigMetrics(reportValue)
	})

}
