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

package goverrule_auth

import (
	"context"

	"go.uber.org/zap"

	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	"github.com/pole-io/pole-server/apis/access_control/auth"
	cacheapi "github.com/pole-io/pole-server/apis/cache"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/pole-server/pkg/goverrule"
)

// Server 带有鉴权能力的 discoverServer
//
//	该层会对请求参数做一些调整，根据具体的请求发起人，设置为数据对应的 owner，不可为为别人进行创建资源
type Server struct {
	nextSvr   goverrule.GoverRuleServer
	userSvr   auth.UserServer
	policySvr auth.StrategyServer
}

func NewServer(nextSvr goverrule.GoverRuleServer,
	userSvr auth.UserServer, policySvr auth.StrategyServer) goverrule.GoverRuleServer {
	proxy := &Server{
		nextSvr:   nextSvr,
		userSvr:   userSvr,
		policySvr: policySvr,
	}
	return proxy
}

// Cache Get cache management
func (svr *Server) Cache() cacheapi.CacheManager {
	return svr.nextSvr.Cache()
}

// collectServiceAuthContext 对于服务的处理，收集所有的与鉴权的相关信息
//
//	@receiver svr Server
//	@param ctx 请求上下文 ctx
//	@param req 实际请求对象
//	@param resourceOp 该接口的数据操作类型
//	@return *authtypes.AcquireContext 返回鉴权上下文
func (svr *Server) collectServiceAuthContext(ctx context.Context, req []*apiservice.Service,
	resourceOp authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {
	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(resourceOp),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(svr.queryServiceResource(req)),
	)
}

// collectServiceAliasAuthContext 对于服务别名的处理，收集所有的与鉴权的相关信息
//
//	@receiver svr Server
//	@param ctx 请求上下文 ctx
//	@param req 实际请求对象
//	@param resourceOp 该接口的数据操作类型
//	@return *authtypes.AcquireContext 返回鉴权上下文
func (svr *Server) collectServiceAliasAuthContext(ctx context.Context, req []*apiservice.ServiceAlias,
	resourceOp authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {
	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(resourceOp),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(svr.queryServiceAliasResource(req)),
	)
}

// collectInstanceAuthContext 对于服务实例的处理，收集所有的与鉴权的相关信息
//
//	@receiver svr Server
//	@param ctx 请求上下文 ctx
//	@param req 实际请求对象
//	@param resourceOp 该接口的数据操作类型
//	@return *authtypes.AcquireContext 返回鉴权上下文
func (svr *Server) collectInstanceAuthContext(ctx context.Context, req []*apiservice.Instance,
	resourceOp authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {
	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(resourceOp),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(svr.queryInstanceResource(req)),
	)
}

// collectClientInstanceAuthContext 对于服务实例的处理，收集所有的与鉴权的相关信息
func (svr *Server) collectClientInstanceAuthContext(ctx context.Context, req []*apiservice.Instance,
	resourceOp authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {
	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(resourceOp),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(methodName),
		authtypes.WithFromClient(),
		authtypes.WithAccessResources(svr.queryInstanceResource(req)),
	)
}

// collectRouteRuleAuthContext 对于服务路由规则的处理，收集所有的与鉴权的相关信息
//
//	@receiver svr Server
//	@param ctx 请求上下文 ctx
//	@param req 实际请求对象
//	@param resourceOp 该接口的数据操作类型
//	@return *authtypes.AcquireContext 返回鉴权上下文
func (svr *Server) collectRouteRuleAuthContext(ctx context.Context, req []*apitraffic.RouteRule,
	resourceOp authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {
	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(resourceOp),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(svr.queryRouteRuleResource(req)),
	)
}

// collectRateLimitAuthContext 对于服务限流规则的处理，收集所有的与鉴权的相关信息
//
//	@receiver svr Server
//	@param ctx 请求上下文 ctx
//	@param req 实际请求对象
//	@param resourceOp 该接口的数据操作类型
//	@return *authtypes.AcquireContext 返回鉴权上下文
func (svr *Server) collectRateLimitAuthContext(ctx context.Context, req []*apitraffic.RateLimit,
	resourceOp authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {

	resources := make([]authtypes.ResourceEntry, 0, len(req))
	for i := range req {
		saveRule := svr.Cache().RateLimit().GetRule(req[i].GetId())
		if saveRule != nil {
			resources = append(resources, authtypes.ResourceEntry{
				Type:     apisecurity.ResourceType_RateLimitRules,
				ID:       saveRule.ID,
				Metadata: saveRule.Metadata,
			})
		}
	}

	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(resourceOp),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_RateLimitRules: resources,
		}),
	)
}

// collectRateLimitAuthContext 对于服务限流规则的处理，收集所有的与鉴权的相关信息
//
//	@receiver svr Server
//	@param ctx 请求上下文 ctx
//	@param req 实际请求对象
//	@param resourceOp 该接口的数据操作类型
//	@return *authtypes.AcquireContext 返回鉴权上下文
func (svr *Server) collectLosslessAuthContext(ctx context.Context, req []*apitraffic.LosslessRule,
	resourceOp authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {

	resources := make([]authtypes.ResourceEntry, 0, len(req))
	for i := range req {
		saveRule := svr.Cache().Lossless().GetRule(req[i].GetId())
		if saveRule != nil {
			resources = append(resources, authtypes.ResourceEntry{
				Type:     apisecurity.ResourceType_LosslessRules,
				ID:       saveRule.ID,
				Metadata: saveRule.Metadata,
			})
		}
	}

	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(resourceOp),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_LosslessRules: resources,
		}),
	)
}

// collectRouteRuleV2AuthContext 收集路由v2规则
func (svr *Server) collectRouteRuleV2AuthContext(ctx context.Context, req []*apitraffic.RouteRule,
	resourceOp authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {

	resources := make([]authtypes.ResourceEntry, 0, len(req))
	for i := range req {
		saveRule := svr.Cache().RoutingConfig().GetRule(req[i].GetId())
		if saveRule != nil {
			resources = append(resources, authtypes.ResourceEntry{
				Type:     apisecurity.ResourceType_RouteRules,
				ID:       saveRule.ID,
				Metadata: saveRule.Metadata,
			})
		}
	}

	accessResources := map[apisecurity.ResourceType][]authtypes.ResourceEntry{}
	if len(resources) != 0 {
		accessResources[apisecurity.ResourceType_RouteRules] = resources
	}

	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(resourceOp),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(accessResources),
	)
}

// collectCircuitBreakerRuleV2 收集熔断v2规则
func (svr *Server) collectCircuitBreakerRuleV2(ctx context.Context, req []*apifault.CircuitBreakerRule,
	op authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {

	resources := make([]authtypes.ResourceEntry, 0, len(req))
	for i := range req {
		saveRule := svr.Cache().CircuitBreaker().GetRule(req[i].GetId())
		if saveRule != nil {
			resources = append(resources, authtypes.ResourceEntry{
				Type:     apisecurity.ResourceType_CircuitBreakerRules,
				ID:       saveRule.ID,
				Metadata: saveRule.Proto.GetMetadata(),
			})
		}
	}

	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(op),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_CircuitBreakerRules: resources,
		}),
	)
}

// collectFaultDetectAuthContext 收集主动探测规则
func (svr *Server) collectFaultDetectAuthContext(ctx context.Context, req []*apifault.FaultDetectRule,
	op authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {

	resources := make([]authtypes.ResourceEntry, 0, len(req))
	for i := range req {
		saveRule := svr.Cache().FaultDetector().GetRule(req[i].GetId())
		if saveRule != nil {
			resources = append(resources, authtypes.ResourceEntry{
				Type:     apisecurity.ResourceType_FaultDetectRules,
				ID:       saveRule.ID,
				Metadata: saveRule.Metadata,
			})
		}
	}

	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(op),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_FaultDetectRules: resources,
		}),
	)
}

// collectRuleReleases 收集规则发布的资源
func (svr *Server) collectRuleReleases(ctx context.Context, req []*apimodel.RuleRelease,
	op authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {

	resources := map[apisecurity.ResourceType][]authtypes.ResourceEntry{
		apisecurity.ResourceType_CircuitBreakerRules: {},
		apisecurity.ResourceType_LaneRules:           {},
		apisecurity.ResourceType_FaultDetectRules:    {},
		apisecurity.ResourceType_RouteRules:          {},
		apisecurity.ResourceType_RateLimitRules:      {},
		apisecurity.ResourceType_LosslessRules:       {},
		apisecurity.ResourceType_SecurityRules:       {},
		apisecurity.ResourceType_MirrorRules:         {},
		apisecurity.ResourceType_MockRules:           {},
	}

	for i := range req {
		switch req[i].GetResource() {
		case apimodel.RuleRelease_LaneRules:
			saveRule := svr.Cache().LaneRule().GetRule(req[i].GetId())
			if saveRule != nil {
				resources[apisecurity.ResourceType_LaneRules] = append(resources[apisecurity.ResourceType_LaneRules], authtypes.ResourceEntry{
					Type:     apisecurity.ResourceType_LaneRules,
					ID:       saveRule.ID,
					Metadata: saveRule.Labels,
				})
			}
		case apimodel.RuleRelease_CircuitBreakerRules:
			saveRule := svr.Cache().CircuitBreaker().GetRule(req[i].GetId())
			if saveRule != nil {
				resources[apisecurity.ResourceType_CircuitBreakerRules] = append(resources[apisecurity.ResourceType_CircuitBreakerRules], authtypes.ResourceEntry{
					Type:     apisecurity.ResourceType_CircuitBreakerRules,
					ID:       saveRule.ID,
					Metadata: circuitBreakerRuleMetadata(saveRule),
				})
			}
		case apimodel.RuleRelease_FaultDetectRules:
			saveRule := svr.Cache().FaultDetector().GetRule(req[i].GetId())
			if saveRule != nil {
				resources[apisecurity.ResourceType_FaultDetectRules] = append(resources[apisecurity.ResourceType_FaultDetectRules], authtypes.ResourceEntry{
					Type:     apisecurity.ResourceType_FaultDetectRules,
					ID:       saveRule.ID,
					Metadata: saveRule.Metadata,
				})
			}
		case apimodel.RuleRelease_RouteRules:
			saveRule := svr.Cache().RoutingConfig().GetRule(req[i].GetId())
			if saveRule != nil {
				resources[apisecurity.ResourceType_RouteRules] = append(resources[apisecurity.ResourceType_RouteRules], authtypes.ResourceEntry{
					Type:     apisecurity.ResourceType_RouteRules,
					ID:       saveRule.ID,
					Metadata: saveRule.Metadata,
				})
			}
		case apimodel.RuleRelease_RateLimitRules:
			saveRule := svr.Cache().RateLimit().GetRule(req[i].GetId())
			if saveRule != nil {
				resources[apisecurity.ResourceType_RateLimitRules] = append(resources[apisecurity.ResourceType_RateLimitRules], authtypes.ResourceEntry{
					Type:     apisecurity.ResourceType_RateLimitRules,
					ID:       saveRule.ID,
					Metadata: saveRule.Metadata,
				})
			}
		case apimodel.RuleRelease_LosslessRules:
			saveRule := svr.Cache().Lossless().GetRule(req[i].GetId())
			if saveRule != nil {
				resources[apisecurity.ResourceType_LosslessRules] = append(resources[apisecurity.ResourceType_LosslessRules], authtypes.ResourceEntry{
					Type:     apisecurity.ResourceType_LosslessRules,
					ID:       saveRule.ID,
					Metadata: saveRule.Metadata,
				})
			}
		case apimodel.RuleRelease_TrafficSecurityRules:
			saveRule := svr.Cache().TrafficSecurity().GetRule(req[i].GetId())
			if saveRule != nil {
				resources[apisecurity.ResourceType_SecurityRules] = append(resources[apisecurity.ResourceType_SecurityRules], authtypes.ResourceEntry{
					Type:     apisecurity.ResourceType_SecurityRules,
					ID:       saveRule.ID,
					Metadata: saveRule.Metadata,
				})
			}
		case apimodel.RuleRelease_TrafficMirrorRules:
			saveRule := svr.Cache().TrafficMirror().GetRule(req[i].GetId())
			if saveRule != nil {
				resources[apisecurity.ResourceType_MirrorRules] = append(resources[apisecurity.ResourceType_MirrorRules], authtypes.ResourceEntry{
					Type:     apisecurity.ResourceType_MirrorRules,
					ID:       saveRule.ID,
					Metadata: saveRule.Metadata,
				})
			}
		case apimodel.RuleRelease_TrafficMockRules:
			saveRule := svr.Cache().TrafficMock().GetRule(req[i].GetId())
			if saveRule != nil {
				resources[apisecurity.ResourceType_MockRules] = append(resources[apisecurity.ResourceType_MockRules], authtypes.ResourceEntry{
					Type:     apisecurity.ResourceType_MockRules,
					ID:       saveRule.ID,
					Metadata: saveRule.Metadata,
				})
			}
		}
	}

	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(op),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(resources),
	)
}

// queryServiceResource  根据所给的 service 信息，收集对应的 ResourceEntry 列表
func (svr *Server) queryServiceResource(
	req []*apiservice.Service) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	if len(req) == 0 {
		return make(map[apisecurity.ResourceType][]authtypes.ResourceEntry)
	}

	names := container.NewSet[string]()
	svcSet := map[string]*svctypes.Service{}

	for index := range req {
		svcName := req[index].GetName()
		svcNamespace := req[index].GetNamespace()
		names.Add(svcNamespace)
		svc := svr.Cache().Service().GetServiceByName(svcName, svcNamespace)
		if svc != nil {
			svcSet[svc.ID] = svc
		}
	}

	ret := svr.convertToDiscoverResourceEntryMaps(names, svcSet)
	if authLog.DebugEnabled() {
		authLog.Debug("[Auth][Server] collect service access res", zap.Any("res", ret))
	}
	return ret
}

// queryServiceAliasResource  根据所给的 servicealias 信息，收集对应的 ResourceEntry 列表
func (svr *Server) queryServiceAliasResource(
	req []*apiservice.ServiceAlias) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	if len(req) == 0 {
		return make(map[apisecurity.ResourceType][]authtypes.ResourceEntry)
	}

	names := container.NewSet[string]()
	svcSet := map[string]*svctypes.Service{}

	for index := range req {
		refSvcName := req[index].GetService()
		refSvcNamespace := req[index].GetNamespace()
		svcNamespace := req[index].GetNamespace()
		names.Add(svcNamespace)
		refSvc := svr.Cache().Service().GetServiceByName(refSvcName, refSvcNamespace)
		if refSvc != nil {
			svcSet[refSvc.ID] = refSvc
		}
	}

	ret := svr.convertToDiscoverResourceEntryMaps(names, svcSet)
	if authLog.DebugEnabled() {
		authLog.Debug("[Auth][Server] collect service alias access res", zap.Any("res", ret))
	}
	return ret
}

// queryInstanceResource 根据所给的 instances 信息，收集对应的 ResourceEntry 列表
// 由于实例是注册到服务下的，因此只需要判断，是否有对应服务的权限即可
func (svr *Server) queryInstanceResource(
	req []*apiservice.Instance) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	if len(req) == 0 {
		return make(map[apisecurity.ResourceType][]authtypes.ResourceEntry)
	}

	names := container.NewSet[string]()
	svcSet := map[string]*svctypes.Service{}

	for index := range req {
		svcName := req[index].GetService()
		svcNamespace := req[index].GetNamespace()
		item := req[index]
		if svcNamespace != "" && svcName != "" {
			svc := svr.Cache().Service().GetServiceByName(svcName, svcNamespace)
			if svc != nil {
				svcSet[svc.ID] = svc
			} else {
				names.Add(svcNamespace)
			}
		} else {
			ins := svr.Cache().Instance().GetInstance(item.GetId())
			if ins != nil {
				svc := svr.Cache().Service().GetServiceByID(ins.ServiceID)
				if svc != nil {
					svcSet[svc.ID] = svc
				} else {
					names.Add(svcNamespace)
				}
			}
		}
	}

	ret := svr.convertToDiscoverResourceEntryMaps(names, svcSet)
	if authLog.DebugEnabled() {
		authLog.Debug("[Auth][Server] collect instance access res", zap.Any("res", ret))
	}
	return ret
}

// queryRouteRuleResource 根据所给的 RouteRule 信息，收集对应的 ResourceEntry 列表
func (svr *Server) queryRouteRuleResource(
	req []*apitraffic.RouteRule) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	if len(req) == 0 {
		return make(map[apisecurity.ResourceType][]authtypes.ResourceEntry)
	}

	names := container.NewSet[string]()
	svcSet := map[string]*svctypes.Service{}

	for index := range req {
		// RouteRule 中的服务信息在 routing_config 中，根据不同类型解析
		// 这里先使用 namespace 来收集相关资源
		svcNamespace := req[index].GetNamespace()
		if svcNamespace != "" {
			names.Add(svcNamespace)
		}
	}

	ret := svr.convertToDiscoverResourceEntryMaps(names, svcSet)
	if authLog.DebugEnabled() {
		authLog.Debug("[Auth][Server] collect route-rule access res", zap.Any("res", ret))
	}
	return ret
}

// queryRateLimitConfigResource 根据所给的 RateLimit 信息，收集对应的 ResourceEntry 列表
func (svr *Server) queryRateLimitConfigResource(
	req []*apitraffic.RateLimit) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	if len(req) == 0 {
		return make(map[apisecurity.ResourceType][]authtypes.ResourceEntry)
	}

	names := container.NewSet[string]()
	svcSet := map[string]*svctypes.Service{}

	for index := range req {
		svcName := req[index].GetService()
		svcNamespace := req[index].GetNamespace()
		svc := svr.Cache().Service().GetServiceByName(svcName, svcNamespace)
		if svc != nil {
			svcSet[svc.ID] = svc
		}
	}

	ret := svr.convertToDiscoverResourceEntryMaps(names, svcSet)
	if authLog.DebugEnabled() {
		authLog.Debug("[Auth][Server] collect rate-limit access res", zap.Any("res", ret))
	}
	return ret
}

// convertToDiscoverResourceEntryMaps 通用方法，进行转换为期望的、服务相关的 ResourceEntry
func (svr *Server) convertToDiscoverResourceEntryMaps(nsSet *container.Set[string],
	svcSet map[string]*svctypes.Service) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	var (
		param = nsSet.ToSlice()
		nsArr = svr.Cache().Namespace().GetNamespacesByName(param)
		nsRet = make([]authtypes.ResourceEntry, 0, len(nsArr))
	)
	for index := range nsArr {
		ns := nsArr[index]
		nsRet = append(nsRet, authtypes.ResourceEntry{
			Type:     apisecurity.ResourceType_Namespaces,
			ID:       ns.Name,
			Owner:    ns.Owner,
			Metadata: ns.Metadata,
		})
	}

	svcRet := make([]authtypes.ResourceEntry, 0, len(svcSet))
	for _, svc := range svcSet {
		svcRet = append(svcRet, authtypes.ResourceEntry{
			Type:     apisecurity.ResourceType_Services,
			ID:       svc.ID,
			Owner:    svc.Owner,
			Metadata: svc.Meta,
		})
	}

	return map[apisecurity.ResourceType][]authtypes.ResourceEntry{
		apisecurity.ResourceType_Namespaces: nsRet,
		apisecurity.ResourceType_Services:   svcRet,
	}
}
