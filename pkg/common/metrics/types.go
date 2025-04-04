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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	clientInstanceTotal   prometheus.Gauge
	serviceCount          *prometheus.GaugeVec
	serviceOnlineCount    *prometheus.GaugeVec
	serviceAbnormalCount  *prometheus.GaugeVec
	serviceOfflineCount   *prometheus.GaugeVec
	instanceCount         *prometheus.GaugeVec
	instanceOnlineCount   *prometheus.GaugeVec
	instanceAbnormalCount *prometheus.GaugeVec
	instanceIsolateCount  *prometheus.GaugeVec
)

var (
	configGroupTotal       *prometheus.GaugeVec
	configFileTotal        *prometheus.GaugeVec
	releaseConfigFileTotal *prometheus.GaugeVec
)

// instance astbc registry metrics
var (
	// instanceAsyncRegisCost 实例异步注册任务耗费时间
	instanceAsyncRegisCost prometheus.Histogram
	// instanceRegisTaskExpire 实例异步注册任务超时无效事件
	instanceRegisTaskExpire prometheus.Counter
	redisReadFailure        prometheus.Gauge
	redisWriteFailure       prometheus.Gauge
	redisAliveStatus        prometheus.Gauge
	// discoveryConnTotal 服务发现客户端链接数量
	discoveryConnTotal prometheus.Gauge
	// configurationConnTotal 配置中心客户端链接数量
	configurationConnTotal prometheus.Gauge
	// sdkClientTotal 客户端链接数量
	sdkClientTotal  prometheus.Gauge
	cacheUpdateCost *prometheus.HistogramVec
	// batchJobUnFinishJobs .
	batchJobUnFinishJobs *prometheus.GaugeVec
)
