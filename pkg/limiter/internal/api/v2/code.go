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

	"github.com/pole-io/pole-server/pkg/limiter/internal/statistics"
)

type Code uint32

// 对于400的错误，一般是规则错误或者客户端BUG导致，直接前台返回失败，无需重试
// 对于404错误，一般是淘汰可以重新init
// 对于500错误，属于服务端错误，需要告警处理，同时客户端重试

const (
	// 接口调用成功
	ExecuteSuccess Code = 200000
	// 接口调用成功
	ExecuteContinue Code = 200100
	// 针对HTTP接口，json转pb失败
	ParseException = 400001
	// 初始化接口，服务名不合法
	InvalidServiceName = 400201
	// 初始化接口，命名空间不合法
	InvalidNamespace = 400202
	// 初始化接口，客户端ID不合法
	InvalidClientId = 400203
	// 初始化接口，总配额数不合法
	InvalidTotalLimit = 400204
	// 初始化接口，总配额数不合法
	InvalidDuration = 400205
	// 初始化接口，滑窗数不合法
	InvalidSlideCount = 400206
	// 初始化接口，限流模式不合法
	InvalidMode = 400207
	// 初始化接口，labels不合法
	InvalidLabels = 400208
	// 初始化接口，批量初始化请求不合法
	InvalidBatchInitReq = 400209
	// 限流接口，使用的配额不合法
	InvalidUsedLimit = 400211
	// 上报接口，时间戳不合法
	InvalidTimestamp = 400212
	// 上报接口，客户端标识不合法
	InvalidClientKey = 400213
	// 上报接口，计数器标识不合法
	InvalidCounterKey = 400214
	// 租约预留参数不合法
	InvalidReservation = 400215
	// 租约消费明细不合法
	InvalidConsumption = 400216
	// 累计消费量不合法
	InvalidConsumedTotal = 400217
	// 租约消息序号不合法
	InvalidSequence = 400218
	// 幂等键不合法
	InvalidIdempotencyKey = 400219
	// 相同幂等键对应了不同请求
	IdempotencyConflict = 409001
	// 配额不足
	QuotaExceeded = 429001
	// 超过最大的counter限制
	ExceedMaxCounter = 401101
	// 超过最大的client限制
	ExceedMaxClient = 401102
	// 多个客户端使用同一个stream上报
	ExceedMaxClientOneStream = 401103
	// 上报接口，找不到计数器
	NotFoundLimiter = 404001
	// 上报接口，找不到已注册的客户端
	NotFoundClient = 404002
	// 找不到租约
	LeaseNotFound = 404003
	// 租约已过期
	LeaseExpired = 410001
	// 租约已结算
	LeaseSettled = 410002
)

// 返回接口名
func GetAPIKey(resp *apiv2.RateLimitResponse) statistics.APIKey {
	switch resp.GetCmd() {
	case apiv2.RateLimitCmd_INIT:
		return statistics.InitQuota
	case apiv2.RateLimitCmd_RESERVE:
		return statistics.ReserveQuota
	case apiv2.RateLimitCmd_BATCH_INIT:
		return statistics.BatchInitQuota
	case apiv2.RateLimitCmd_UPDATE:
		return statistics.UpdateQuota
	default:
		return statistics.SettleQuota
	}
}

// 返回错误码
func GetErrorCode(resp *apiv2.RateLimitResponse) uint32 {
	switch resp.GetCmd() {
	case apiv2.RateLimitCmd_INIT:
		return resp.GetRateLimitInitResponse().GetCode()
	case apiv2.RateLimitCmd_RESERVE:
		return resp.GetQuotaReserveResponse().GetCode()
	case apiv2.RateLimitCmd_BATCH_INIT:
		return resp.GetRateLimitBatchInitResponse().GetCode()
	case apiv2.RateLimitCmd_UPDATE:
		return resp.GetQuotaUpdateResponse().GetCode()
	default:
		return resp.GetQuotaSettleResponse().GetCode()
	}
}

// 转为http status
func Code2HTTPStatus(code Code) int {
	return int(code / 1000)
}
