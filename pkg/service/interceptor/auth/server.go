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

package service_auth

import (
	"context"

	apifault "github.com/polarismesh/specification/source/go/api/v1/fault_tolerance"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"
	"go.uber.org/zap"

	"github.com/pole-io/pole-server/apis/access_control/auth"
	cacheapi "github.com/pole-io/pole-server/apis/cache"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/service"
)

// Server 带有鉴权能力的 discoverServer
//
//	该层会对请求参数做一些调整，根据具体的请求发起人，设置为数据对应的 owner，不可为为别人进行创建资源
type Server struct {
	nextSvr   service.DiscoverServer
	userSvr   auth.UserServer
	policySvr auth.StrategyServer
}

func NewServer(nextSvr service.DiscoverServer,
	userSvr auth.UserServer, policySvr auth.StrategyServer) service.DiscoverServer {
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

// GetServiceInstanceRevision 获取服务实例的版本号
func (svr *Server) GetServiceInstanceRevision(serviceID string,
	instances []*svctypes.Instance) (string, error) {
	return svr.nextSvr.GetServiceInstanceRevision(serviceID, instances)
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
func (svr *Server) collectRouteRuleAuthContext(ctx context.Context, req []*apitraffic.Routing,
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
func (svr *Server) collectRateLimitAuthContext(ctx context.Context, req []*apitraffic.Rule,
	resourceOp authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {

	resources := make([]authtypes.ResourceEntry, 0, len(req))
	for i := range req {
		saveRule := svr.Cache().RateLimit().GetRule(req[i].GetId().GetValue())
		if saveRule != nil {
			resources = append(resources, authtypes.ResourceEntry{
				Type:     apisecurity.ResourceType_RouteRules,
				ID:       saveRule.ID,
				Metadata: saveRule.Proto.Metadata,
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
			apisecurity.ResourceType_FaultDetectRules: resources,
		}),
	)
}

// queryServiceResource  根据所给的 service 信息，收集对应的 ResourceEntry 列表
func (svr *Server) queryServiceResource(
	req []*apiservice.Service) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	if len(req) == 0 {
		return make(map[apisecurity.ResourceType][]authtypes.ResourceEntry)
	}

	names := utils.NewSet[string]()
	svcSet := utils.NewMap[string, *svctypes.Service]()

	for index := range req {
		svcName := req[index].GetName().GetValue()
		svcNamespace := req[index].GetNamespace().GetValue()
		names.Add(svcNamespace)
		svc := svr.Cache().Service().GetServiceByName(svcName, svcNamespace)
		if svc != nil {
			svcSet.Store(svc.ID, svc)
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

	names := utils.NewSet[string]()
	svcSet := utils.NewMap[string, *svctypes.Service]()

	for index := range req {
		refSvcName := req[index].GetService().GetValue()
		refSvcNamespace := req[index].GetNamespace().GetValue()
		svcNamespace := req[index].GetNamespace().GetValue()
		names.Add(svcNamespace)
		refSvc := svr.Cache().Service().GetServiceByName(refSvcName, refSvcNamespace)
		if refSvc != nil {
			svcSet.Store(refSvc.ID, refSvc)
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

	names := utils.NewSet[string]()
	svcSet := utils.NewMap[string, *svctypes.Service]()

	for index := range req {
		svcName := req[index].GetService().GetValue()
		svcNamespace := req[index].GetNamespace().GetValue()
		item := req[index]
		if svcNamespace != "" && svcName != "" {
			svc := svr.Cache().Service().GetServiceByName(svcName, svcNamespace)
			if svc != nil {
				svcSet.Store(svc.ID, svc)
			} else {
				names.Add(svcNamespace)
			}
		} else {
			ins := svr.Cache().Instance().GetInstance(item.GetId().GetValue())
			if ins != nil {
				svc := svr.Cache().Service().GetServiceByID(ins.ServiceID)
				if svc != nil {
					svcSet.Store(svc.ID, svc)
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
	req []*apitraffic.Routing) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	if len(req) == 0 {
		return make(map[apisecurity.ResourceType][]authtypes.ResourceEntry)
	}

	names := utils.NewSet[string]()
	svcSet := utils.NewMap[string, *svctypes.Service]()

	for index := range req {
		svcName := req[index].GetService().GetValue()
		svcNamespace := req[index].GetNamespace().GetValue()
		svc := svr.Cache().Service().GetServiceByName(svcName, svcNamespace)
		if svc != nil {
			svcSet.Store(svc.ID, svc)
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
	req []*apitraffic.Rule) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	if len(req) == 0 {
		return make(map[apisecurity.ResourceType][]authtypes.ResourceEntry)
	}

	names := utils.NewSet[string]()
	svcSet := utils.NewMap[string, *svctypes.Service]()

	for index := range req {
		svcName := req[index].GetService().GetValue()
		svcNamespace := req[index].GetNamespace().GetValue()
		svc := svr.Cache().Service().GetServiceByName(svcName, svcNamespace)
		if svc != nil {
			svcSet.Store(svc.ID, svc)
		}
	}

	ret := svr.convertToDiscoverResourceEntryMaps(names, svcSet)
	if authLog.DebugEnabled() {
		authLog.Debug("[Auth][Server] collect rate-limit access res", zap.Any("res", ret))
	}
	return ret
}

// convertToDiscoverResourceEntryMaps 通用方法，进行转换为期望的、服务相关的 ResourceEntry
func (svr *Server) convertToDiscoverResourceEntryMaps(nsSet *utils.Set[string],
	svcSet *utils.Map[string, *svctypes.Service]) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
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

	svcRet := make([]authtypes.ResourceEntry, 0, svcSet.Len())
	svcSet.Range(func(key string, svc *svctypes.Service) {
		svcRet = append(svcRet, authtypes.ResourceEntry{
			Type:     apisecurity.ResourceType_Services,
			ID:       svc.ID,
			Owner:    svc.Owner,
			Metadata: svc.Meta,
		})
	})

	return map[apisecurity.ResourceType][]authtypes.ResourceEntry{
		apisecurity.ResourceType_Namespaces: nsRet,
		apisecurity.ResourceType_Services:   svcRet,
	}
}
