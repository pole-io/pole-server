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

package v2

import (
	apiv2 "github.com/pole-io/specification/source/go/api/v1/traffic_manage/ratelimiter"

	"github.com/pole-io/pole-server/pkg/limiter/internal/utils"
)

// 新建一个初始化回复结构体
func NewRateLimitInitResponse(code Code, target *apiv2.LimitTarget) *apiv2.RateLimitInitResponse {
	return &apiv2.RateLimitInitResponse{Code: uint32(code), Target: target, Timestamp: utils.CurrentMillisecond()}
}

// 新建一个初始化回复结构体
func NewRateLimitBatchInitResponse(code Code) *apiv2.RateLimitBatchInitResponse {
	return &apiv2.RateLimitBatchInitResponse{Code: uint32(code), Timestamp: utils.CurrentMillisecond()}
}
