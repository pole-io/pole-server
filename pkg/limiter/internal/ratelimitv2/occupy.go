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

package ratelimitv2

import (
	apiv2 "github.com/pole-io/specification/source/go/api/v1/traffic_manage/ratelimiter"

	"github.com/pole-io/pole-server/pkg/limiter/internal/utils"
)

// OccupyAllocator 抢占式分配器
type OccupyAllocator struct {
	slidingWindow *utils.SlidingWindow
	mode          apiv2.Mode
}

// NewOccupyAllocator 创建抢占式分配器
func NewOccupyAllocator(slideCount int, intervalMs int) QuotaAllocator {
	return &OccupyAllocator{
		slidingWindow: utils.NewSlidingWindow(slideCount, intervalMs),
		mode:          apiv2.Mode_BATCH_OCCUPY,
	}
}

// Mode 返回分配器所属的模式
func (o *OccupyAllocator) Mode() apiv2.Mode {
	return o.mode
}

// Current 返回当前窗口已提交配额。
func (o *OccupyAllocator) Current(timestampMs int64) uint32 {
	return o.slidingWindow.AddAndGetCurrent(timestampMs, timestampMs, 0)
}

// Commit 将消费量提交到当前窗口。
func (o *OccupyAllocator) Commit(timestampMs int64, amount uint32) uint32 {
	return o.slidingWindow.AddAndGetCurrent(timestampMs, timestampMs, amount)
}
