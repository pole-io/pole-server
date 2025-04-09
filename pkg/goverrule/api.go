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
	// EnableCircuitBreakerRules Enable the CircuitBreaker rule
	EnableCircuitBreakerRules(ctx context.Context, request []*apifault.CircuitBreakerRule) *apiservice.BatchWriteResponse
	// UpdateCircuitBreakerRules Modify the CircuitBreaker rule
	UpdateCircuitBreakerRules(ctx context.Context, request []*apifault.CircuitBreakerRule) *apiservice.BatchWriteResponse
	// GetCircuitBreakerRules Query CircuitBreaker rules
	GetCircuitBreakerRules(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse
}

// RateLimitOperateServer Lamflow rule related operation
type RateLimitOperateServer interface {
	// CreateRateLimits Create a RateLimit rule
	CreateRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse
	// DeleteRateLimits Delete current RateLimit rules
	DeleteRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse
	// EnableRateLimits Enable the RateLimit rule
	EnableRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse
	// UpdateRateLimits Modify the RateLimit rule
	UpdateRateLimits(ctx context.Context, request []*apitraffic.Rule) *apiservice.BatchWriteResponse
	// GetRateLimits Query RateLimit rules
	GetRateLimits(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse
}

// RouterRuleOperateServer Routing rules related operations
type RouterRuleOperateServer interface {
	// CreateRoutingConfigs Batch creation routing configuration
	CreateRoutingConfigs(ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse
	// DeleteRoutingConfigs Batch delete routing configuration
	DeleteRoutingConfigs(ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse
	// UpdateRoutingConfigs Batch update routing configuration
	UpdateRoutingConfigs(ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse
	// QueryRoutingConfigs Inquiry route configuration to OSS
	QueryRoutingConfigs(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse
	// EnableRoutings batch enable routing rules
	EnableRoutings(ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse
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
}

// ClientServer Client related operation  Client operation interface definition
type ClientServer interface {
	// GetRoutingConfigWithCache User Client Get Service Routing Configuration Information
	GetRoutingConfigWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse
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
