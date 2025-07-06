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

package store

import (
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
)

// GovernanceStore Service discovery, governance center module storage interface
type GovernanceStore interface {
	// RateLimitStore 限流规则接口
	RateLimitStore
	// CircuitBreakerStore 熔断规则接口
	CircuitBreakerStore
	// ToolStore 函数及工具接口
	ToolStore
	// RouterRuleConfigStore 路由策略接口
	RouterRuleConfigStore
	// FaultDetectRuleStore fault detect rule interface
	FaultDetectRuleStore
	// ServiceContractStore 服务契约操作接口
	ServiceContractStore
	// LaneStore 泳道规则存储操作接口
	LaneStore
}

// RateLimitStore 限流规则的存储接口
type RateLimitStore interface {
	// CreateRateLimit 新增限流规则
	CreateRateLimit(limiting *rules.RateLimit) error
	// UpdateRateLimit 更新限流规则
	UpdateRateLimit(limiting *rules.RateLimit) error
	// EnableRateLimit 启用限流规则
	EnableRateLimit(limit *rules.RateLimit) error
	// DeleteRateLimit 删除限流规则
	DeleteRateLimit(limiting *rules.RateLimit) error
	// GetRateLimitWithID 根据限流ID拉取限流规则
	GetRateLimitWithID(id string) (*rules.RateLimit, error)
	// GetRateLimitsForCache 根据修改时间拉取增量限流规则及最新版本号
	// 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
	GetRateLimitsForCache(mtime time.Time, firstUpdate bool) ([]*rules.RateLimit, error)
	// LockRateLimitRule 锁住一个限流规则
	LockRateLimitRule(tx Tx, name string) (*rules.RateLimit, error)

	// 关于规则发布
	// GetActiveRateLimitRule 获取处于使用状态的限流规则
	GetActiveRateLimitRule(tx Tx, release *rules.RateLimitRelease) (*rules.RateLimitRelease, error)
	// ActiveRateLimitRule 设置某个限流规则发布为使用状态
	ActiveRateLimitRule(tx Tx, release *rules.RateLimitRelease) error
	// InactiveRateLimitRule 设置某个限流规则的发布为不使用状态
	InactiveRateLimitRule(tx Tx, release *rules.RateLimitRelease) error
	// PublishRateLimitRule 发布限流规则
	PublishRateLimitRule(tx Tx, rule *rules.RateLimitRelease) error
	// 获取已发布的限流规则
	GetMoreRateLimitReleases(mtime time.Time, firstUpdate bool) ([]*rules.RateLimitRelease, error)
}

// CircuitBreakerStore 熔断规则的存储接口
type CircuitBreakerStore interface {
	// CreateCircuitBreakerRule create general circuitbreaker rule
	CreateCircuitBreakerRule(cbRule *rules.CircuitBreakerRule) error
	// UpdateCircuitBreakerRule update general circuitbreaker rule
	UpdateCircuitBreakerRule(cbRule *rules.CircuitBreakerRule) error
	// DeleteCircuitBreakerRule delete general circuitbreaker rule
	DeleteCircuitBreakerRule(id string) error
	// HasCircuitBreakerRule check circuitbreaker rule exists
	HasCircuitBreakerRule(id string) (bool, error)
	// HasCircuitBreakerRuleByName check circuitbreaker rule exists for name
	HasCircuitBreakerRuleByName(name string, namespace string) (bool, error)
	// HasCircuitBreakerRuleByNameExcludeId check circuitbreaker rule exists for name not this id
	HasCircuitBreakerRuleByNameExcludeId(name string, namespace string, id string) (bool, error)
	// GetCircuitBreakerRules get all circuitbreaker rules by query and limit
	GetCircuitBreakerRules(
		filter map[string]string, offset uint32, limit uint32) (uint32, []*rules.CircuitBreakerRule, error)
	// GetCircuitBreakerRulesForCache get increment circuitbreaker rules
	GetCircuitBreakerRulesForCache(mtime time.Time, firstUpdate bool) ([]*rules.CircuitBreakerRule, error)
	// EnableCircuitBreakerRule enable specific circuitbreaker rule
	EnableCircuitBreakerRule(cbRule *rules.CircuitBreakerRule) error
	// LockCircuitBreakerRule 锁住一个熔断规则
	LockCircuitBreakerRule(tx Tx, name string) (*rules.CircuitBreakerRule, error)

	// 关于规则发布
	// GetActiveCircuitBreakerRule 获取处于使用状态的熔断规则
	GetActiveCircuitBreakerRule(tx Tx, release *rules.CircuitBreakerRelease) (*rules.CircuitBreakerRelease, error)
	// ActiveCircuitBreakerRule 设置某个熔断规则发布为使用状态
	ActiveCircuitBreakerRule(tx Tx, release *rules.CircuitBreakerRelease) error
	// InactiveCircuitBreakerRule 设置某个熔断规则的发布为不使用状态
	InactiveCircuitBreakerRule(tx Tx, release *rules.CircuitBreakerRelease) error
	// PublishCircuitBreakerRule 发布熔断规则规则
	PublishCircuitBreakerRule(tx Tx, rule *rules.CircuitBreakerRelease) error
}

// RouterRuleConfigStore 路由配置表的存储接口
type RouterRuleConfigStore interface {
	// EnableRouting 设置路由规则是否启用
	EnableRouting(conf *rules.RouterConfig) error
	// CreateRoutingConfig 新增一个路由配置
	CreateRoutingConfig(conf *rules.RouterConfig) error
	// CreateRoutingConfigTx 新增一个路由配置
	CreateRoutingConfigTx(tx Tx, conf *rules.RouterConfig) error
	// UpdateRoutingConfig 更新一个路由配置
	UpdateRoutingConfig(conf *rules.RouterConfig) error
	// UpdateRoutingConfigTx 更新一个路由配置
	UpdateRoutingConfigTx(tx Tx, conf *rules.RouterConfig) error
	// DeleteRoutingConfig 删除一个路由配置
	DeleteRoutingConfig(serviceID string) error
	// GetRoutingConfigsForCache 通过mtime拉取增量的路由配置信息
	// 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
	GetRoutingConfigsForCache(mtime time.Time, firstUpdate bool) ([]*rules.RouterConfig, error)
	// GetRoutingConfigWithID 根据服务ID拉取路由配置
	GetRoutingConfigWithID(id string) (*rules.RouterConfig, error)
	// GetRoutingConfigWithIDTx 根据服务ID拉取路由配置
	GetRoutingConfigWithIDTx(tx Tx, id string) (*rules.RouterConfig, error)
	// LockRouterRule 锁住一个路由规则
	LockRouterRule(tx Tx, name string) (*rules.RouterConfig, error)

	// 关于规则发布
	// GetActiveRouterRule 获取处于使用状态的路由规则
	GetActiveRouterRule(tx Tx, release *rules.CustomRouteRelease) (*rules.CustomRouteRelease, error)
	// ActiveRouterRule 设置某个路由规则发布为使用状态
	ActiveRouterRule(tx Tx, release *rules.CustomRouteRelease) error
	// InactiveRouterRule 设置某个路由规则的发布为不使用状态
	InactiveRouterRule(tx Tx, release *rules.CustomRouteRelease) error
	// PublishRouterRule 发布路由规则规则
	PublishRouterRule(tx Tx, rule *rules.CustomRouteRelease) error
}

// FaultDetectRuleStore store api for the fault detector config
type FaultDetectRuleStore interface {
	// CreateFaultDetectRule create fault detect rule
	CreateFaultDetectRule(conf *rules.FaultDetectRule) error
	// UpdateFaultDetectRule update fault detect rule
	UpdateFaultDetectRule(conf *rules.FaultDetectRule) error
	// DeleteFaultDetectRule delete fault detect rule
	DeleteFaultDetectRule(id string) error
	// HasFaultDetectRule check fault detect rule exists
	HasFaultDetectRule(id string) (bool, error)
	// HasFaultDetectRuleByName check fault detect rule exists by name
	HasFaultDetectRuleByName(name string, namespace string) (bool, error)
	// HasFaultDetectRuleByNameExcludeId check fault detect rule exists by name not this id
	HasFaultDetectRuleByNameExcludeId(name string, namespace string, id string) (bool, error)
	// GetFaultDetectRules get all fault detect rules by query and limit
	GetFaultDetectRules(filter map[string]string, offset uint32, limit uint32) (uint32, []*rules.FaultDetectRule, error)
	// GetFaultDetectRulesForCache get increment fault detect rules
	GetFaultDetectRulesForCache(mtime time.Time, firstUpdate bool) ([]*rules.FaultDetectRule, error)
	// LockFaultDetectRule 锁住一个探测规则
	LockFaultDetectRule(tx Tx, name string) (*rules.FaultDetectRule, error)

	// 关于规则发布
	// GetActiveFaultDetectRule 获取处于使用状态的探测规则
	GetActiveFaultDetectRule(tx Tx, release *rules.FaultDetectRelease) (*rules.FaultDetectRelease, error)
	// ActiveFaultDetectRule 设置某个探测规则发布为使用状态
	ActiveFaultDetectRule(tx Tx, release *rules.FaultDetectRelease) error
	// InactiveFaultDetectRule 设置某个探测规则的发布为不使用状态
	InactiveFaultDetectRule(tx Tx, release *rules.FaultDetectRelease) error
	// PublishFaultDetectRule 发布探测规则规则
	PublishFaultDetectRule(tx Tx, rule *rules.FaultDetectRelease) error
}

// LaneStore 泳道资源存储操作
type LaneStore interface {
	// AddLaneGroup 添加泳道组
	AddLaneGroup(tx Tx, item *rules.LaneGroup) error
	// UpdateLaneGroup 更新泳道组
	UpdateLaneGroup(tx Tx, item *rules.LaneGroup) error
	// GetLaneGroup 按照名称查询泳道组
	GetLaneGroup(name string) (*rules.LaneGroup, error)
	// GetLaneGroupByID 按照名称查询泳道组
	GetLaneGroupByID(id string) (*rules.LaneGroup, error)
	// GetLaneGroups 查询泳道组
	GetLaneGroups(filter map[string]string, offset, limit uint32) (uint32, []*rules.LaneGroup, error)
	// GetMoreLaneGroups 获取泳道规则列表到缓存层
	GetMoreLaneGroups(mtime time.Time, firstUpdate bool) (map[string]*rules.LaneGroup, error)
	// DeleteLaneGroup 删除泳道组
	DeleteLaneGroup(id string) error
	// GetLaneRuleMaxPriority 获取泳道规则中当前最大的泳道规则优先级信息
	GetLaneRuleMaxPriority() (int32, error)
	// LockLaneGroup 锁住一个泳道分组
	LockLaneGroup(tx Tx, name string) (*rules.LaneGroup, error)
	// GetLaneRule 查询泳道规则
	GetLaneRule(id string) (*rules.LaneRule, error)
	// AddLaneRules 添加泳道规则
	AddLaneRules(tx Tx, rules []*rules.LaneRule) error
	// UpdateLaneRules 更新泳道规则
	UpdateLaneRules(tx Tx, rules []*rules.LaneRule) error
	// DeleteLaneRules 删除泳道规则
	DeleteLaneRules(tx Tx, group string, ids []string) error

	// 关于规则发布
	// GetActiveLaneGroup 获取处于使用状态的泳道组
	GetActiveLaneGroup(tx Tx, release *rules.LaneGroupRelease) (*rules.LaneGroupRelease, error)
	// ActiveLaneGroup 设置某个泳道组发布为使用状态
	ActiveLaneGroup(tx Tx, release *rules.LaneGroupRelease) error
	// InactiveLaneGroup 设置某个泳道组的发布为不使用状态
	InactiveLaneGroup(tx Tx, release *rules.LaneGroupRelease) error
	// PublishLaneGroup 发布泳道组规则
	PublishLaneGroup(tx Tx, rule *rules.LaneGroupRelease) error
}
