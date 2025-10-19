package goverrule

import (
	"context"

	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
)

// GoverRuleServer Server discovered by the service
type GoverRuleServer interface {
	// CircuitBreakerOperateServer Fuse rule operation interface definition
	CircuitBreakerOperateServer
	// RateLimitOperateServer Lamflow rule operation interface definition
	RateLimitOperateServer
	// RouterRuleOperateServer Routing rules operation interface definition
	RouterRuleOperateServer
	// FaultDetectRuleOperateServer fault detect rules operation interface definition
	FaultDetectRuleOperateServer
	// LaneOperateServer lane rule operation interface definition
	LaneOperateServer
	// LossLessOperateServer lossless rule operation interface definition
	LossLessOperateServer
	// ClientServer Client operation interface definition
	ClientServer
	// GovernanceRuleReleaseServer Governance rule operation interface definition
	RuleReleaseServer
	// Cache Get cache management
	Cache() cacheapi.CacheManager
}

// CircuitBreakerOperateServer Melting rule related treatment
type CircuitBreakerOperateServer interface {
	// CreateCircuitBreakerRules Create a CircuitBreaker rule
	CreateCircuitBreakerRules(ctx context.Context, request []*apifault.CircuitBreakerRule) *apiservice.BatchWriteResponse
	// DeleteCircuitBreakerRules Delete current CircuitBreaker rules
	DeleteCircuitBreakerRules(ctx context.Context, request []*apifault.CircuitBreakerRule) *apiservice.BatchWriteResponse
	// UpdateCircuitBreakerRules Modify the CircuitBreaker rule
	UpdateCircuitBreakerRules(ctx context.Context, request []*apifault.CircuitBreakerRule) *apiservice.BatchWriteResponse
	// GetCircuitBreakerRules Query CircuitBreaker rules
	GetCircuitBreakerRules(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse
	// GetOneCircuitBreakerRule Query a single CircuitBreaker rule
	GetOneCircuitBreakerRule(ctx context.Context, req *apifault.CircuitBreakerRule) *apiservice.Response
}

// RateLimitOperateServer Lamflow rule related operation
type RateLimitOperateServer interface {
	// CreateRateLimits Create a RateLimit rule
	CreateRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse
	// DeleteRateLimits Delete current RateLimit rules
	DeleteRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse
	// UpdateRateLimits Modify the RateLimit rule
	UpdateRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse
	// GetRateLimits Query RateLimit rules
	GetRateLimits(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse
	// GetOneRateLimitRule Query a single RateLimit rule
	GetOneRateLimitRule(ctx context.Context, req *apitraffic.Rule) *apiservice.Response
}

// RouterRuleOperateServer Routing rules related operations
type RouterRuleOperateServer interface {
	// CreateRouterRules Batch creation routing configuration
	CreateRouterRules(ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse
	// DeleteRouterRules Batch delete routing configuration
	DeleteRouterRules(ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse
	// UpdateRouterRules Batch update routing configuration
	UpdateRouterRules(ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse
	// QueryRouterRules Inquiry route configuration to OSS
	QueryRouterRules(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse
	// GetOneRouterRule Query a single routing rule
	GetOneRouterRule(ctx context.Context, req *apitraffic.RouteRule) *apiservice.Response
}

// FaultDetectRuleOperateServer Fault detect rules related operations
type FaultDetectRuleOperateServer interface {
	// CreateFaultDetectRules create the fault detect rule by request
	CreateFaultDetectRules(ctx context.Context, request []*apifault.FaultDetectRule) *apiservice.BatchWriteResponse
	// DeleteFaultDetectRules delete the fault detect rule by request
	DeleteFaultDetectRules(ctx context.Context, request []*apifault.FaultDetectRule) *apiservice.BatchWriteResponse
	// UpdateFaultDetectRules update the fault detect rule by request
	UpdateFaultDetectRules(ctx context.Context, request []*apifault.FaultDetectRule) *apiservice.BatchWriteResponse
	// GetFaultDetectRules get the fault detect rule by request
	GetFaultDetectRules(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse
	// GetOneFaultDetectRule get one fault detect rule by request
	GetOneFaultDetectRule(ctx context.Context, req *apifault.FaultDetectRule) *apiservice.Response
}

// LaneOperateServer lane operations
type LaneOperateServer interface {
	// CreateLaneGroups 批量创建泳道组
	CreateLaneGroups(ctx context.Context, req []*apitraffic.LaneGroup) *apiservice.BatchWriteResponse
	// UpdateLaneGroups 批量更新泳道组
	UpdateLaneGroups(ctx context.Context, req []*apitraffic.LaneGroup) *apiservice.BatchWriteResponse
	// DeleteLaneGroups 批量删除泳道组
	DeleteLaneGroups(ctx context.Context, req []*apitraffic.LaneGroup) *apiservice.BatchWriteResponse
	// GetLaneGroups 查询泳道组列表
	GetLaneGroups(ctx context.Context, filter map[string]string) *apiservice.BatchQueryResponse
	// CreateLaneRules 批量创建泳道规则
	CreateLaneRules(ctx context.Context, req []*apitraffic.LaneRule) *apiservice.BatchWriteResponse
	// UpdateLaneRules 批量更新泳道规则
	UpdateLaneRules(ctx context.Context, req []*apitraffic.LaneRule) *apiservice.BatchWriteResponse
	// DeleteLaneRules 批量删除泳道规则
	DeleteLaneRules(ctx context.Context, req []*apitraffic.LaneRule) *apiservice.BatchWriteResponse
	// GetOneLaneGroup 查询单个泳道组
	GetOneLaneGroup(ctx context.Context, req *apitraffic.LaneGroup) *apiservice.Response
}

// LossLessOperateServer lane operations
type LossLessOperateServer interface {
	// CreateLossLessRules 批量创建无损规则
	CreateLossLessRules(ctx context.Context, req []*apitraffic.LosslessRule) *apiservice.BatchWriteResponse
	// UpdateLossLessRules 批量更新无损规则
	UpdateLossLessRules(ctx context.Context, req []*apitraffic.LosslessRule) *apiservice.BatchWriteResponse
	// DeleteLossLessRules 批量删除无损规则
	DeleteLossLessRules(ctx context.Context, req []*apitraffic.LosslessRule) *apiservice.BatchWriteResponse
	// GetLossLessRules 查询无损规则列表
	GetLossLessRules(ctx context.Context, filter map[string]string) *apiservice.BatchQueryResponse
	// GetOneLossLessRule 查询单个无损规则
	GetOneLossLessRule(ctx context.Context, req *apitraffic.LosslessRule) *apiservice.Response
}

// ClientServer Client related operation  Client operation interface definition
type ClientServer interface {
	// GetOldRouterRuleWithCache User Client Get Service Routing Configuration Information
	GetOldRouterRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse
	// GetRateLimitWithCache User Client Get Service Limit Configuration Information
	GetRateLimitWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse
	// GetCircuitBreakerWithCache Fuse configuration information for obtaining services for clients
	GetCircuitBreakerWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse
	// GetFaultDetectWithCache User Client Get FaultDetect Rule Information
	GetFaultDetectWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse
	// GetLaneRuleWithCache fetch lane rules by client
	GetLaneRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse
	// GetRouterRuleWithCache fetch lane rules by client
	GetRouterRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse
	// GetServiceWithCache fetch service list by client
	GetLosslessRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse
}

type RuleReleaseServer interface {
	// GetRuleReleases 获取已发布的规则
	GetRuleReleases(ctx context.Context, filter map[string]string) *apiservice.BatchQueryResponse
	// PublishLaneGroups 发布多个治理规则
	PublishGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse
	// DeleteLaneGroups 删除多个治理规则已发布版本
	DeleteGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse
	// RollbackLaneGroups 回滚多个治理规则到目标版本
	RollbackGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse
	// StopbetaLaneGroups 停止多个治理规则灰度发布版本
	StopbetaGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse
}
