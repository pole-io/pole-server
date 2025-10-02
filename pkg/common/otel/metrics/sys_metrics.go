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
	"context"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/pole-io/pole-server/pkg/common/otel"
)

const (
	labelCacheType        = "cache_type"
	labelCacheUpdateCount = "cache_update_count"
	labelBatchJobLabel    = "batch_label"
)

var (
	lastRedisReadFailureReport  atomic.Value
	lastRedisWriteFailureReport atomic.Value
)

func registerSysMetrics() {
	var err error
	// instanceAsyncRegisCost 实例异步注册任务耗费时间
	instanceAsyncRegisCost, err = otel.Meter().Float64Histogram("instance_regis_cost_time",
		metric.WithDescription("Total time to report the short-term registered task of the reporting instance"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic("register instance async regis cost metric failed: " + err.Error())
	}

	// instanceRegisTaskExpire 实例异步注册任务超时无效事件
	instanceRegisTaskExpire, err = otel.Meter().Int64Counter("instance_regis_task_expire",
		metric.WithDescription("The number of registered tasks discarded due to timeout"),
		metric.WithUnit("1"),
	)
	if err != nil {
		panic("register instance regis task expire metric failed: " + err.Error())
	}

	cacheUpdateCost, err = otel.Meter().Float64Histogram("cache_update_cost",
		metric.WithDescription("The cost of cache updates"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic("register cache update cost metric failed: " + err.Error())
	}

	batchJobUnFinishJobs, err = otel.Meter().Int64Gauge("batch_job_unfinish",
		metric.WithDescription("The number of unfinished batch jobs"),
		metric.WithUnit("1"),
	)
	if err != nil {
		panic("register batch job unfinish metric failed: " + err.Error())
	}
}

// ReportInstanceRegisCost Total time to report the short-term registered task of the reporting instance
func ReportInstanceRegisCost(cost time.Duration) {
	instanceAsyncRegisCost.Record(context.Background(), float64(cost.Milliseconds()))
}

// ReportDropInstanceRegisTask Record the number of registered tasks discarded
func ReportDropInstanceRegisTask() {
	instanceRegisTaskExpire.Add(context.Background(), 1)
}

// RecordCacheUpdateCost record per cache update cost time
func RecordCacheUpdateCost(cost time.Duration, cacheTye string, _ int64) {
	if cacheUpdateCost == nil {
		return
	}
	cacheUpdateCost.Record(context.Background(), float64(cost.Milliseconds()), metric.WithAttributes(
		attribute.String(labelCacheType, cacheTye),
	))
}

// ReportAddBatchJob .
func ReportAddBatchJob(label string, count int64) {
	if batchJobUnFinishJobs == nil {
		return
	}
	batchJobUnFinishJobs.Record(context.Background(), count, metric.WithAttributes(
		attribute.String(labelBatchJobLabel, label),
	))
}

// ReportFinishBatchJob .
func ReportFinishBatchJob(label string, count int64) {
	if batchJobUnFinishJobs == nil {
		return
	}
	batchJobUnFinishJobs.Record(context.Background(), -count, metric.WithAttributes(
		attribute.String(labelBatchJobLabel, label),
	))
}
