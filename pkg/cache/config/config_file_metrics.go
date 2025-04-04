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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/pole-io/pole-server/apis/observability/statis"
	metrictypes "github.com/pole-io/pole-server/apis/pkg/types/metrics"
	"github.com/pole-io/pole-server/pkg/common/metrics"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
)

func (fc *fileCache) reportMetricsInfo() {
	lastReportTime := fc.lastReportTime.Load()
	if time.Since(lastReportTime) <= time.Minute {
		return
	}
	defer func() {
		fc.lastReportTime.Store(time.Now())
	}()

	metricValues := make([]metrictypes.ConfigMetrics, 0, 64)

	configFiles, err := fc.storage.CountConfigFileEachGroup()
	if err != nil {
		log.Error("[Cache][ConfigFile] report metrics for config_file each group", zap.Error(err))
		return
	}
	tmpGroup := map[string]map[string]struct{}{}
	for ns, groups := range configFiles {
		if _, ok := tmpGroup[ns]; !ok {
			tmpGroup[ns] = map[string]struct{}{}
		}
		for group := range groups {
			tmpGroup[ns][group] = struct{}{}
		}
	}
	_, _ = cleanExpireConfigFileMetricLabel(fc.preMetricsFiles.Load(), tmpGroup)
	fc.preMetricsFiles.Store(tmpGroup)

	for ns, groups := range configFiles {
		for group, total := range groups {
			metricValues = append(metricValues, metrictypes.ConfigMetrics{
				Type:  metrictypes.FileMetric,
				Total: total,
				Labels: map[string]string{
					metrictypes.LabelNamespace: ns,
					metrictypes.LabelGroup:     group,
				},
			})
		}
	}

	fc.metricsReleaseCount.ReadRange(func(namespace string, groups *container.SyncMap[string, uint64]) {
		groups.ReadRange(func(groupName string, count uint64) {
			metricValues = append(metricValues, metrictypes.ConfigMetrics{
				Type:  metrictypes.ReleaseFileMetric,
				Total: int64(count),
				Labels: map[string]string{
					metrictypes.LabelNamespace: namespace,
					metrictypes.LabelGroup:     groupName,
				},
			})
		})
	})

	statis.GetStatis().ReportConfigMetrics(metricValues...)
}

func cleanExpireConfigFileMetricLabel(pre, curr map[string]map[string]struct{}) (map[string]struct{}, map[string]map[string]struct{}) {
	if len(pre) == 0 {
		return map[string]struct{}{}, map[string]map[string]struct{}{}
	}

	var (
		removeNs     = map[string]struct{}{}
		removeGroups = map[string]map[string]struct{}{}
	)

	for ns, groups := range pre {
		if _, ok := curr[ns]; !ok {
			removeNs[ns] = struct{}{}
		}
		if _, ok := removeGroups[ns]; !ok {
			removeGroups[ns] = map[string]struct{}{}
		}
		for group := range groups {
			if _, ok := curr[ns][group]; !ok {
				removeGroups[ns][group] = struct{}{}
			}
		}
	}

	for ns := range removeNs {
		metrics.GetConfigGroupTotal().Delete(prometheus.Labels{
			metrictypes.LabelNamespace: ns,
		})
	}

	for ns, groups := range removeGroups {
		for group := range groups {
			metrics.GetConfigFileTotal().Delete(prometheus.Labels{
				metrictypes.LabelNamespace: ns,
				metrictypes.LabelGroup:     group,
			})
			metrics.GetReleaseConfigFileTotal().Delete(prometheus.Labels{
				metrictypes.LabelNamespace: ns,
				metrictypes.LabelGroup:     group,
			})
			metrics.GetConfigFileTotal().Delete(prometheus.Labels{
				metrictypes.LabelNamespace: ns,
				metrictypes.LabelGroup:     group,
			})
		}
	}
	return removeNs, removeGroups
}
