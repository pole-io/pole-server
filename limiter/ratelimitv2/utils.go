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
	"time"

	limiterapi "github.com/pole-io/pole-server/limiter/pkg/api/v2"
	"github.com/pole-io/pole-server/limiter/pkg/config"
	apiv2 "github.com/pole-io/specification/source/go/api/v1/traffic_manage/ratelimiter"
)

// 默认滑窗数量
const (
	// MaxSlideCount 最大滑窗
	MaxSlideCount = config.MaxSlideCount
)

// CheckRateLimitReportRequest 检查限流上报请求参数
func CheckRateLimitReportRequest(req *apiv2.RateLimitReportRequest) *limiterapi.TimedRateLimitReportResponse {
	if req.GetClientKey() == 0 {
		return limiterapi.NewRateLimitReportResponse(limiterapi.InvalidClientKey)
	}
	if req.GetTimestamp() == 0 {
		return limiterapi.NewRateLimitReportResponse(limiterapi.InvalidTimestamp)
	}
	if len(req.GetQuotaUses()) == 0 {
		return limiterapi.NewRateLimitReportResponse(limiterapi.InvalidUsedLimit)
	}
	for _, quotaUsed := range req.GetQuotaUses() {
		if quotaUsed.GetCounterKey() == 0 {
			return limiterapi.NewRateLimitReportResponse(limiterapi.InvalidCounterKey)
		}
	}
	return nil
}

// 通用检查限流请求的参数
func checkInitRequest(
	req *apiv2.RateLimitInitRequest, defaultSlideCount uint32) (*apiv2.RateLimitInitResponse, time.Duration) {
	if len(req.GetTarget().GetService()) == 0 {
		return limiterapi.NewRateLimitInitResponse(limiterapi.InvalidServiceName, req.GetTarget()), 0
	}
	if len(req.GetTarget().GetNamespace()) == 0 {
		return limiterapi.NewRateLimitInitResponse(limiterapi.InvalidNamespace, req.GetTarget()), 0
	}
	if len(req.GetTotals()) == 0 {
		return limiterapi.NewRateLimitInitResponse(limiterapi.InvalidTotalLimit, req.GetTarget()), 0
	}
	var maxDuration time.Duration
	for _, total := range req.GetTotals() {
		if total.GetDuration() == 0 {
			return limiterapi.NewRateLimitInitResponse(limiterapi.InvalidDuration, req.GetTarget()), 0
		}
		timeDuration := time.Duration(total.GetDuration()) * time.Second
		if maxDuration < timeDuration {
			maxDuration = timeDuration
		}
	}
	if req.GetSlideCount() == 0 {
		req.SlideCount = defaultSlideCount
	} else if req.GetSlideCount() > MaxSlideCount {
		return limiterapi.NewRateLimitInitResponse(limiterapi.InvalidSlideCount, req.GetTarget()), 0
	}
	if req.GetMode() != apiv2.Mode_ADAPTIVE && req.GetMode() != apiv2.Mode_BATCH_OCCUPY {
		return limiterapi.NewRateLimitInitResponse(limiterapi.InvalidMode, req.GetTarget()), 0
	}
	return nil, maxDuration
}

// CheckRateLimitInitRequest 检查限流初始化请求参数
func CheckRateLimitInitRequest(
	req *apiv2.RateLimitInitRequest, defaultSlideCount uint32) (*apiv2.RateLimitInitResponse, time.Duration) {
	if len(req.GetClientId()) == 0 {
		return limiterapi.NewRateLimitInitResponse(limiterapi.InvalidClientId, req.GetTarget()), 0
	}
	return checkInitRequest(req, defaultSlideCount)
}

// CheckRateLimitBatchInitRequest 检查限流初始化请求参数
func CheckRateLimitBatchInitRequest(
	req *apiv2.RateLimitInitRequest, defaultSlideCount uint32) (*apiv2.RateLimitInitResponse, time.Duration) {
	if len(req.GetTarget().GetLabelsList()) == 0 {
		return limiterapi.NewRateLimitInitResponse(limiterapi.InvalidLabels, req.GetTarget()), 0
	}
	return checkInitRequest(req, defaultSlideCount)
}
