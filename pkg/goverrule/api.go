package goverrule

import (
	"context"

	apifault "github.com/polarismesh/specification/source/go/api/v1/fault_tolerance"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

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
	// ClientServer Client operation interface definition
	ClientServer
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
	// PublishCircuitBreakerRules Publish the CircuitBreaker rule 发布多个熔断规则
	PublishCircuitBreakerRules(ctx context.Context, request []*apifault.CircuitBreakerRule) *apiservice.BatchWriteResponse
	// RollbackCircuitBreakerRules Rollback the CircuitBreaker rule
	RollbackCircuitBreakerRules(ctx context.Context, request []*apifault.CircuitBreakerRule) *apiservice.BatchWriteResponse
	// StopbetaCircuitBreakerRules Rollback the CircuitBreaker rule
	StopbetaCircuitBreakerRules(ctx context.Context, request []*apifault.CircuitBreakerRule) *apiservice.BatchWriteResponse
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
	// PublishRateLimits 发布多个限流规则
	PublishRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse
	// RollbackRateLimits Rollback the RateLimit rule
	RollbackRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse
	// StopbetaRateLimits Rollback the RateLimit rule
	StopbetaRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse
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
	// PublishRouterRules 发布多个路由规则
	PublishRouterRules(ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse
	// RollbackRouterRules Rollback the routing rule
	RollbackRouterRules(ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse
	// StopbetaRouterRules Rollback the routing rule
	StopbetaRouterRules(ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse
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
	// PublishFaultDetectRules 发布多个故障检测规则
	PublishFaultDetectRules(ctx context.Context, request []*apifault.FaultDetectRule) *apiservice.BatchWriteResponse
	// RollbackFaultDetectRules Rollback the fault detect rule
	RollbackFaultDetectRules(ctx context.Context, request []*apifault.FaultDetectRule) *apiservice.BatchWriteResponse
	// StopbetaFaultDetectRules Rollback the fault detect rule
	StopbetaFaultDetectRules(ctx context.Context, request []*apifault.FaultDetectRule) *apiservice.BatchWriteResponse
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
	// PublishLaneGroups 发布多个泳道组
	PublishLaneGroups(ctx context.Context, req []*apitraffic.LaneGroup) *apiservice.BatchWriteResponse
	// RollbackLaneGroups 回滚泳道组
	RollbackLaneGroups(ctx context.Context, req []*apitraffic.LaneGroup) *apiservice.BatchWriteResponse
	// StopbetaLaneGroups 回滚泳道组
	StopbetaLaneGroups(ctx context.Context, req []*apitraffic.LaneGroup) *apiservice.BatchWriteResponse
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
}
