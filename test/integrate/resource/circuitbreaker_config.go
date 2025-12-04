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

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

const (
	circuitBreakerName      = "test-name-%v"
	circuitBreakerNamespace = "test-namespace-%v"
)

/**
 * @brief 创建测试熔断规则
 */
func CreateCircuitBreakers(namespace *apimodel.Namespace) []*apifault.CircuitBreakerRule {
	var circuitBreakers []*apifault.CircuitBreakerRule
	for index := 0; index < 2; index++ {
		circuitBreaker := &apifault.CircuitBreakerRule{
			Name:        fmt.Sprintf(circuitBreakerName, index),
			Namespace:   namespace.GetName(),
			Description: "test",
			Enable:      true,
			Level:       apifault.Level_INSTANCE,
			BlockConfigs: []*apifault.BlockConfig{
				{
					Name: fmt.Sprintf("block-config-%d", index),
					Api: &apimodel.API{
						Path: &apimodel.MatchString{
							Type:  apimodel.MatchString_EXACT,
							Value: fmt.Sprintf("/api/test-%d", index),
						},
						Method: "GET",
					},
					ErrorConditions: []*apifault.ErrorCondition{
						{
							InputType: apifault.ErrorCondition_RET_CODE,
							Condition: &apimodel.MatchString{
								Type:  apimodel.MatchString_EXACT,
								Value: "500",
							},
						},
						{
							InputType: apifault.ErrorCondition_RET_CODE,
							Condition: &apimodel.MatchString{
								Type:  apimodel.MatchString_REGEX,
								Value: "5[0-9]{2}",
							},
						},
					},
					TriggerConditions: []*apifault.TriggerCondition{
						{
							TriggerType:    apifault.TriggerCondition_ERROR_RATE,
							ErrorPercent:   50,
							Interval:       60,
							MinimumRequest: 10,
						},
						{
							TriggerType: apifault.TriggerCondition_CONSECUTIVE_ERROR,
							ErrorCount:  5,
							Interval:    60,
						},
					},
				},
			},
			RuleMatcher: &apifault.RuleMatcher{
				Source: &apifault.RuleMatcher_SourceService{
					Service:   fmt.Sprintf("service-test-%d", index),
					Namespace: fmt.Sprintf("namespace-test-%d", index),
				},
				Destination: &apifault.RuleMatcher_DestinationService{
					Service:   fmt.Sprintf("service-test-%d", index),
					Namespace: fmt.Sprintf("namespace-test-%d", index),
					Method: &apimodel.MatchString{
						Type:  apimodel.MatchString_EXACT,
						Value: "GET",
					},
				},
			},
			MaxEjectionPercent: 50,
			RecoverCondition: &apifault.RecoverCondition{
				SleepWindow:        60,
				ConsecutiveSuccess: 3,
			},
		}
		circuitBreakers = append(circuitBreakers, circuitBreaker)
	}
	return circuitBreakers
}

/**
 * @brief 更新测试熔断规则
 */
func UpdateCircuitBreakers(circuitBreakers []*apifault.CircuitBreakerRule) {
	for index, item := range circuitBreakers {
		item.Description = fmt.Sprintf("updated-test-%d", index)
		item.BlockConfigs = []*apifault.BlockConfig{
			{
				Name: fmt.Sprintf("updated-block-config-%d", index),
				Api: &apimodel.API{
					Path: &apimodel.MatchString{
						Type:  apimodel.MatchString_EXACT,
						Value: "/api/updated",
					},
					Method: "POST",
				},
				ErrorConditions: []*apifault.ErrorCondition{
					{
						InputType: apifault.ErrorCondition_RET_CODE,
						Condition: &apimodel.MatchString{
							Type:  apimodel.MatchString_EXACT,
							Value: "503",
						},
					},
				},
				TriggerConditions: []*apifault.TriggerCondition{
					{
						TriggerType:    apifault.TriggerCondition_ERROR_RATE,
						ErrorPercent:   80,
						Interval:       30,
						MinimumRequest: 5,
					},
				},
			},
		}
		item.RuleMatcher = &apifault.RuleMatcher{
			Source: &apifault.RuleMatcher_SourceService{
				Service:   "testSvc",
				Namespace: "test",
			},
			Destination: &apifault.RuleMatcher_DestinationService{
				Service:   "testSvc1",
				Namespace: "test",
			},
		}
	}
}

/**
 * @brief 创建熔断规则版本
 */
func CreateCircuitBreakerVersions(circuitBreakers []*apifault.CircuitBreakerRule) []*apifault.CircuitBreakerRule {
	var newCircuitBreakers []*apifault.CircuitBreakerRule

	for index, item := range circuitBreakers {
		newCircuitBreaker := &apifault.CircuitBreakerRule{
			Id:        item.GetId(),
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
			Revision:  fmt.Sprintf("test-version-%d", index),
		}
		newCircuitBreakers = append(newCircuitBreakers, newCircuitBreaker)
	}

	return newCircuitBreakers
}

/**
 * @brief 创建测试发布熔断规则
 */
func CreateConfigRelease(services []*apiservice.Service, circuitBreakers []*apifault.CircuitBreakerRule) []*apiconfig.ConfigFileRelease {
	var configReleases []*apiconfig.ConfigFileRelease
	for index := 0; index < 2; index++ {
		configRelease := &apiconfig.ConfigFileRelease{
			Namespace: services[index].GetNamespace(),
			Group:     "circuit-breaker",
			FileName:  fmt.Sprintf("cb-%s-%s", services[index].GetNamespace(), services[index].GetName()),
			Name:      fmt.Sprintf("release-%d", index),
			Content:   fmt.Sprintf("cb-content-%d", index),
		}
		configReleases = append(configReleases, configRelease)
	}
	return configReleases
}
