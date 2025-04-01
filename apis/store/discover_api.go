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
	"context"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
)

// NamingModuleStore Service discovery, governance center module storage interface
type NamingModuleStore interface {
	// ServiceStore 服务接口
	ServiceStore
	// InstanceStore 实例接口
	InstanceStore
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

// ServiceStore 服务存储接口
type ServiceStore interface {
	// AddService 保存一个服务
	AddService(service *svctypes.Service) error
	// DeleteService 删除服务
	DeleteService(id, serviceName, namespaceName string) error
	// DeleteServiceAlias 删除服务别名
	DeleteServiceAlias(name string, namespace string) error
	// UpdateServiceAlias 修改服务别名
	UpdateServiceAlias(alias *svctypes.Service, needUpdateOwner bool) error
	// UpdateService 更新服务
	UpdateService(service *svctypes.Service, needUpdateOwner bool) error
	// UpdateServiceToken 更新服务token
	UpdateServiceToken(serviceID string, token string, revision string) error
	// GetSourceServiceToken 获取源服务的token信息
	GetSourceServiceToken(name string, namespace string) (*svctypes.Service, error)
	// GetService 根据服务名和命名空间获取服务的详情
	GetService(name string, namespace string) (*svctypes.Service, error)
	// GetServiceByID 根据服务ID查询服务详情
	GetServiceByID(id string) (*svctypes.Service, error)
	// GetServices 根据相关条件查询对应服务及数目
	GetServices(serviceFilters, serviceMetas map[string]string, instanceFilters *InstanceArgs, offset, limit uint32) (
		uint32, []*svctypes.Service, error)
	// GetServicesCount 获取所有服务总数
	GetServicesCount() (uint32, error)
	// GetMoreServices 获取增量services
	// 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
	GetMoreServices(mtime time.Time, firstUpdate, disableBusiness, needMeta bool) (map[string]*svctypes.Service, error)
	// GetServiceAliases 获取服务别名列表
	GetServiceAliases(filter map[string]string, offset uint32, limit uint32) (uint32, []*svctypes.ServiceAlias, error)
	// GetSystemServices 获取系统服务
	GetSystemServices() ([]*svctypes.Service, error)
	// GetServicesBatch 批量获取服务id、负责人等信息
	GetServicesBatch(services []*svctypes.Service) ([]*svctypes.Service, error)
}

// InstanceStore 实例存储接口
type InstanceStore interface {
	// AddInstance 增加一个实例
	AddInstance(instance *svctypes.Instance) error
	// BatchAddInstances 增加多个实例
	BatchAddInstances(instances []*svctypes.Instance) error
	// UpdateInstance 更新实例
	UpdateInstance(instance *svctypes.Instance) error
	// DeleteInstance 删除一个实例，实际是把valid置为false
	DeleteInstance(instanceID string) error
	// BatchDeleteInstances 批量删除实例，flag=1
	BatchDeleteInstances(ids []interface{}) error
	// CleanInstance 清空一个实例，真正删除
	CleanInstance(instanceID string) error
	// BatchGetInstanceIsolate 检查ID是否存在，并且返回存在的ID，以及ID的隔离状态
	BatchGetInstanceIsolate(ids map[string]bool) (map[string]bool, error)
	// GetInstancesBrief 获取实例关联的token
	GetInstancesBrief(ids map[string]bool) (map[string]*svctypes.Instance, error)
	// GetInstance 查询一个实例的详情，只返回有效的数据
	GetInstance(instanceID string) (*svctypes.Instance, error)
	// GetInstancesCount 获取有效的实例总数
	GetInstancesCount() (uint32, error)
	// GetInstancesCountTx 获取有效的实例总数
	GetInstancesCountTx(tx Tx) (uint32, error)
	// GetInstancesMainByService 根据服务和Host获取实例（不包括metadata）
	GetInstancesMainByService(serviceID, host string) ([]*svctypes.Instance, error)
	// GetExpandInstances 根据过滤条件查看实例详情及对应数目
	GetExpandInstances(
		filter, metaFilter map[string]string, offset uint32, limit uint32) (uint32, []*svctypes.Instance, error)
	// GetMoreInstances 根据mtime获取增量instances，返回所有store的变更信息
	// 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
	GetMoreInstances(tx Tx, mtime time.Time, firstUpdate, needMeta bool, serviceID []string) (map[string]*svctypes.Instance, error)
	// SetInstanceHealthStatus 设置实例的健康状态
	SetInstanceHealthStatus(instanceID string, flag int, revision string) error
	// BatchSetInstanceHealthStatus 批量设置实例的健康状态
	BatchSetInstanceHealthStatus(ids []interface{}, healthy int, revision string) error
	// BatchSetInstanceIsolate 批量修改实例的隔离状态
	BatchSetInstanceIsolate(ids []interface{}, isolate int, revision string) error
	// AppendInstanceMetadata 追加实例 metadata
	BatchAppendInstanceMetadata(requests []*InstanceMetadataRequest) error
	// RemoveInstanceMetadata 删除实例指定的 metadata
	BatchRemoveInstanceMetadata(requests []*InstanceMetadataRequest) error
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
	// GetExtendRateLimits 根据过滤条件拉取限流规则
	GetExtendRateLimits(query map[string]string, offset uint32, limit uint32) (uint32, []*rules.ExtendRateLimit, error)
	// GetRateLimitWithID 根据限流ID拉取限流规则
	GetRateLimitWithID(id string) (*rules.RateLimit, error)
	// GetRateLimitsForCache 根据修改时间拉取增量限流规则及最新版本号
	// 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
	GetRateLimitsForCache(mtime time.Time, firstUpdate bool) ([]*rules.RateLimit, error)
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
}

// ClientStore store interface for client info
type ClientStore interface {
	// BatchAddClients insert the client info
	BatchAddClients(clients []*types.Client) error
	// BatchDeleteClients delete the client info
	BatchDeleteClients(ids []string) error
	// GetMoreClients 根据mtime获取增量clients，返回所有store的变更信息
	// 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
	GetMoreClients(mtime time.Time, firstUpdate bool) (map[string]*types.Client, error)
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
}

type ServiceContractStore interface {
	// CreateServiceContract 创建服务契约
	CreateServiceContract(contract *svctypes.ServiceContract) error
	// UpdateServiceContract 更新服务契约
	UpdateServiceContract(contract *svctypes.ServiceContract) error
	// DeleteServiceContract 删除服务契约
	DeleteServiceContract(contract *svctypes.ServiceContract) error
	// GetServiceContract 查询服务契约数据
	GetServiceContract(id string) (data *svctypes.EnrichServiceContract, err error)
	// GetServiceContracts 查询服务契约公共属性列表
	GetServiceContracts(ctx context.Context, filter map[string]string, offset, limit uint32) (uint32, []*svctypes.EnrichServiceContract, error)
	// AddServiceContractInterfaces 创建服务契约API接口
	AddServiceContractInterfaces(contract *svctypes.EnrichServiceContract) error
	// AppendServiceContractInterfaces 追加服务契约API接口
	AppendServiceContractInterfaces(contract *svctypes.EnrichServiceContract) error
	// DeleteServiceContractInterfaces 批量删除服务契约API接口
	DeleteServiceContractInterfaces(contract *svctypes.EnrichServiceContract) error
	// GetInterfaceDescriptors 查询服务接口列表
	GetInterfaceDescriptors(ctx context.Context, filter map[string]string, offset, limit uint32) (uint32, []*svctypes.InterfaceDescriptor, error)
	// ListVersions .
	ListVersions(ctx context.Context, service, namespace string) ([]*svctypes.ServiceContract, error)
	// GetMoreServiceContracts 查询服务契约公共属性列表
	GetMoreServiceContracts(firstUpdate bool, mtime time.Time) ([]*svctypes.EnrichServiceContract, error)
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
	// LockLaneGroup 锁住一个泳道分组
	LockLaneGroup(tx Tx, name string) (*rules.LaneGroup, error)
	// GetMoreLaneGroups 获取泳道规则列表到缓存层
	GetMoreLaneGroups(mtime time.Time, firstUpdate bool) (map[string]*rules.LaneGroup, error)
	// DeleteLaneGroup 删除泳道组
	DeleteLaneGroup(id string) error
	// GetLaneRuleMaxPriority 获取泳道规则中当前最大的泳道规则优先级信息
	GetLaneRuleMaxPriority() (int32, error)
}
