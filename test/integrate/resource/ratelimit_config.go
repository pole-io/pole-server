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

package resource

import (
	"fmt"

	"google.golang.org/protobuf/types/known/durationpb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

/**
 * @brief 创建测试限流规则
 */
func CreateRateLimits(services []*apiservice.Service) []*apitraffic.RateLimit {
	var rateLimits []*apitraffic.RateLimit
	for index := 0; index < 2; index++ {
		rateLimit := &apitraffic.RateLimit{
			Name:      fmt.Sprintf("rlimit-%d", index),
			Service:   services[index].GetName(),
			Namespace: services[index].GetNamespace(),
			Priority:  uint32(index),
			Type:      apitraffic.RateLimit_LOCAL,
			Disable:   false,
			Rules: []*apitraffic.LimitTrigger{
				{
					Name:     fmt.Sprintf("rule-%d", index),
					Resource: apitraffic.LimitTrigger_QPS,
					Action:   "REJECT",
					Disable:  false,
					Arguments: []*apitraffic.MatchArgument{
						{
							Type: apitraffic.MatchArgument_CUSTOM,
							Key:  fmt.Sprintf("name-%d", index),
							Value: &apimodel.MatchString{
								Type:  apimodel.MatchString_REGEX,
								Value: fmt.Sprintf("value-%d", index),
							},
						},
						{
							Type: apitraffic.MatchArgument_CUSTOM,
							Key:  fmt.Sprintf("name-%d", index+1),
							Value: &apimodel.MatchString{
								Type:  apimodel.MatchString_EXACT,
								Value: fmt.Sprintf("value-%d", index+1),
							},
						},
					},
					Amounts: []*apitraffic.Amount{
						{
							MaxAmount: uint32(100 + index),
							ValidDuration: &durationpb.Duration{
								Seconds: int64(1),
							},
						},
					},
					MaxQueueDelay: uint32(1000),
					RegexCombine:  true,
					AmountMode:    apitraffic.LimitTrigger_SHARE_EQUALLY,
					Failover:      apitraffic.LimitTrigger_FAILOVER_LOCAL,
				},
			},
		}
		rateLimits = append(rateLimits, rateLimit)
	}
	return rateLimits
}

/**
 * @brief 更新测试限流规则
 */
func UpdateRateLimits(rateLimits []*apitraffic.RateLimit) {
	for index, rateLimit := range rateLimits {
		rateLimit.Rules = []*apitraffic.LimitTrigger{
			{
				Name:     fmt.Sprintf("updated-rule-%d", index),
				Resource: apitraffic.LimitTrigger_CONCURRENCY,
				Action:   "REJECT",
				Disable:  false,
				Arguments: []*apitraffic.MatchArgument{
					{
						Type: apitraffic.MatchArgument_CUSTOM,
						Key:  "key1",
						Value: &apimodel.MatchString{
							Type:  apimodel.MatchString_REGEX,
							Value: "value-1",
						},
					},
				},
			},
		}
	}
}
