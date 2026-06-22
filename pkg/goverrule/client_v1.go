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

package goverrule

import (
	"context"

	// 注释：移除golang/protobuf/ptypes/wrappers导入 - 不再使用wrapper类型
	"go.uber.org/zap"

	// 注释：新增apifault导入 - 用于熔断器规则类型
	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	// 注释：移除protobuf工具包导入 - 改用基础类型
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// GetOldRouterRuleWithCache 获取缓存中的路由配置信息
func (s *Server) GetOldRouterRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_SERVICE_CONTRACTS)
	aliasFor := s.findServiceAlias(req)

	// 注释：缓存方法调用改动 - 从GetOldRouterRule改为GetRouterRule，现在返回三个值
	out, revision, err := s.caches.RoutingConfig().GetRouterRule(aliasFor.ID, aliasFor.Name, aliasFor.Namespace)
	if err != nil {
		log.Error("[Server][Service][Routing] discover routing", utils.RequestID(ctx), zap.Error(err))
		return api.NewDiscoverRoutingResponse(apimodel.Code_ExecuteException, req)
	}
	if out == nil {
		return resp
	}

	// 获取路由数据，并对比revision
	if revision == req.GetRevision() {
		return api.NewDiscoverRoutingResponse(apimodel.Code_DataNoChange, req)
	}

	// 数据不一致，发生了改变
	// 数据格式转换，service只需要返回二元组与routing的revision
	resp.Service.Revision = revision
	resp.CustomRouteRules = out
	resp.AliasFor = &apiservice.Service{
		Name:      aliasFor.Name,
		Namespace: aliasFor.Namespace,
	}
	return resp
}

// GetRateLimitWithCache 获取缓存中的限流规则信息
// GetRateLimitWithCache 获取限流规则
func (s *Server) GetRateLimitWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_RATE_LIMIT)
	aliasFor := s.findServiceAlias(req)

	rules, revision := s.caches.RateLimit().GetRateLimitRules(svctypes.ServiceKey{
		Namespace: aliasFor.Namespace,
		Name:      aliasFor.Name,
	})
	if len(rules) == 0 || revision == "" {
		return resp
	}
	// 注释：版本比较改动 - req.GetRevision()从*wrapperspb.StringValue改为string
	if req.GetRevision() == revision {
		return api.NewDiscoverRateLimitResponse(apimodel.Code_DataNoChange, req)
	}
	resp.RateLimit = make([]*apitraffic.RateLimit, 0, len(rules))
	for i := range rules {
		rateLimit, err := rateLimit2Client(req.GetName(), req.GetNamespace(), rules[i])
		if rateLimit == nil || err != nil {
			continue
		}
		rateLimit.Revision = revision
		resp.RateLimit = append(resp.RateLimit, rateLimit)
	}

	// 塞入源服务信息数据
	resp.AliasFor = &apiservice.Service{
		// 注释：服务字段类型改动 - 从*wrapperspb.StringValue改为string
		Namespace: aliasFor.Namespace,
		Name:      aliasFor.Name,
	}
	// 服务名和request保持一致
	resp.Service = &apiservice.Service{
		Name:      req.GetName(),
		Namespace: req.GetNamespace(),
		// 注释：Revision字段类型改动 - 从*wrapperspb.StringValue改为string
		Revision: revision,
	}
	return resp
}

func (s *Server) GetFaultDetectWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_FAULT_DETECTOR)
	aliasFor := s.findServiceAlias(req)

	out := s.caches.FaultDetector().GetFaultDetectConfig(aliasFor.Name, aliasFor.Namespace)
	if out == nil || out.Revision == "" {
		return resp
	}

	if req.GetRevision() == out.Revision {
		return api.NewDiscoverFaultDetectorResponse(apimodel.Code_DataNoChange, req)
	}

	// 数据不一致，发生了改变
	var err error
	resp.AliasFor = &apiservice.Service{
		Name:      aliasFor.Name,
		Namespace: aliasFor.Namespace,
	}
	resp.Service.Revision = out.Revision
	resp.FaultDetectRules, err = faultDetectRule2ClientAPI(out)
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewDiscoverFaultDetectorResponse(apimodel.Code_ExecuteException, req)
	}
	return resp
}

// GetCircuitBreakerWithCache 获取缓存中的熔断规则信息
func (s *Server) GetCircuitBreakerWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_CIRCUIT_BREAKER)
	// 获取源服务
	aliasFor := s.findServiceAlias(req)
	out := s.caches.CircuitBreaker().GetCircuitBreakerConfig(aliasFor.Name, aliasFor.Namespace)
	if out == nil || out.Revision == "" {
		return resp
	}

	// 获取熔断规则数据，并对比revision
	// 注释：版本比较改动 - req.GetRevision()从*wrapperspb.StringValue改为string
	if len(req.GetRevision()) > 0 && req.GetRevision() == out.Revision {
		return api.NewDiscoverCircuitBreakerResponse(apimodel.Code_DataNoChange, req)
	}

	// 数据不一致，发生了改变
	var err error
	resp.AliasFor = &apiservice.Service{
		// 注释：服务字段类型改动 - Name和Namespace从*wrapperspb.StringValue改为string
		Name:      aliasFor.Name,
		Namespace: aliasFor.Namespace,
	}
	// 注释：Revision字段类型改动 - 从*wrapperspb.StringValue改为string
	resp.Service.Revision = out.Revision
	// 注释：重大API改动 - circuitBreaker2ClientAPI现在返回单个CircuitBreakerRule而非CircuitBreaker
	circuitBreakerRule, err := circuitBreaker2ClientAPI(out, req.GetName(), req.GetNamespace())
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewDiscoverCircuitBreakerResponse(apimodel.Code_ExecuteException, req)
	}
	// 注释：响应结构重大改动 - CircuitBreaker字段现在是[]*apifault.CircuitBreakerRule数组
	if circuitBreakerRule != nil {
		resp.CircuitBreaker = []*apifault.CircuitBreakerRule{circuitBreakerRule}
	}
	return resp
}

// GetLaneRuleWithCache fetch lane rule by client
func (s *Server) GetLaneRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_LANE)
	// 获取源服务
	aliasFor := s.findServiceAlias(req)
	out, revision := s.caches.LaneRule().GetLaneRules(aliasFor)
	if out == nil || revision == "" {
		return resp
	}

	// 获取泳道规则数据，并对比revision
	if len(req.GetRevision()) > 0 && req.GetRevision() == revision {
		return api.NewDiscoverLaneResponse(apimodel.Code_DataNoChange, req)
	}

	resp.AliasFor = &apiservice.Service{
		Name:      aliasFor.Name,
		Namespace: aliasFor.Namespace,
	}
	resp.Service.Revision = revision
	resp.Lanes = make([]*apitraffic.LaneGroup, 0, len(out))
	for i := range out {
		resp.Lanes = append(resp.Lanes, out[i].Proto)
	}
	return resp
}

// GetRouterRuleWithCache fetch lane rules by client
func (s *Server) GetRouterRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_CUSTOM_ROUTE_RULE)
	aliasFor := s.findServiceAlias(req)

	out, revision, err := s.caches.RoutingConfig().GetRouterRule(aliasFor.ID, aliasFor.Name, aliasFor.Namespace)
	if err != nil {
		log.Error("[Server][Service][Routing] discover routing", utils.RequestID(ctx), zap.Error(err))
		return api.NewDiscoverRoutingResponse(apimodel.Code_ExecuteException, req)
	}
	if out == nil {
		return resp
	}

	// 获取路由数据，并对比revision
	if revision == req.GetRevision() {
		return api.NewDiscoverRoutingResponse(apimodel.Code_DataNoChange, req)
	}

	// 数据不一致，发生了改变
	// 数据格式转换，service只需要返回二元组与routing的revision
	resp.Service.Revision = revision
	resp.CustomRouteRules = out
	resp.AliasFor = &apiservice.Service{
		Name:      aliasFor.Name,
		Namespace: aliasFor.Namespace,
	}
	return resp
}

// GetLosslessRuleWithCache fetch service list by client
func (s *Server) GetLosslessRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_LOSSLESS)
	aliasFor := s.findServiceAlias(req)

	out := s.caches.Lossless().GetLosslessConfig(aliasFor.Namespace, aliasFor.Name)
	if out == nil || out.Revision == "" {
		return resp
	}

	// 获取无损规则数据，并对比revision
	if len(req.GetRevision()) > 0 && req.GetRevision() == out.Revision {
		return api.NewDiscoverLosslessResponse(apimodel.Code_DataNoChange, req)
	}

	resp.AliasFor = &apiservice.Service{
		Name:      aliasFor.Name,
		Namespace: aliasFor.Namespace,
	}
	resp.Service.Revision = out.Revision
	resp.LosslessRules = []*apitraffic.LosslessRule{out.ToSpec()}
	return resp
}

func (s *Server) GetTrafficSecurityRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_TRAFFIC_SECURITY_RULE)
	aliasFor := s.findServiceAlias(req)
	out, revision := s.caches.TrafficSecurity().GetRulesForService(aliasFor.Namespace, aliasFor.Name)
	if revision == "" {
		return resp
	}
	if len(req.GetRevision()) > 0 && req.GetRevision() == revision {
		return api.NewDiscoverTrafficSecurityResponse(apimodel.Code_DataNoChange, req)
	}
	resp.AliasFor = &apiservice.Service{Name: aliasFor.Name, Namespace: aliasFor.Namespace}
	resp.Service.Revision = revision
	resp.TrafficSecurityRules = make([]*apisecurity.TrafficSecurityRule, 0, len(out))
	for i := range out {
		resp.TrafficSecurityRules = append(resp.TrafficSecurityRules, out[i].ToTrafficSecuritySpec())
	}
	return resp
}

func (s *Server) GetTrafficMirrorRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_TRAFFIC_MIRROR_RULE)
	aliasFor := s.findServiceAlias(req)
	out, revision := s.caches.TrafficMirror().GetRulesForService(aliasFor.Namespace, aliasFor.Name)
	if revision == "" {
		return resp
	}
	if len(req.GetRevision()) > 0 && req.GetRevision() == revision {
		return api.NewDiscoverTrafficMirrorResponse(apimodel.Code_DataNoChange, req)
	}
	resp.AliasFor = &apiservice.Service{Name: aliasFor.Name, Namespace: aliasFor.Namespace}
	resp.Service.Revision = revision
	resp.TrafficMirrorRules = make([]*apitraffic.TrafficMirror, 0, len(out))
	for i := range out {
		resp.TrafficMirrorRules = append(resp.TrafficMirrorRules, out[i].ToTrafficMirrorSpec())
	}
	return resp
}

func (s *Server) GetTrafficMockRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_TRAFFIC_MOCK_RULE)
	aliasFor := s.findServiceAlias(req)
	out, revision := s.caches.TrafficMock().GetRulesForService(aliasFor.Namespace, aliasFor.Name)
	if revision == "" {
		return resp
	}
	if len(req.GetRevision()) > 0 && req.GetRevision() == revision {
		return api.NewDiscoverTrafficMockResponse(apimodel.Code_DataNoChange, req)
	}
	resp.AliasFor = &apiservice.Service{Name: aliasFor.Name, Namespace: aliasFor.Namespace}
	resp.Service.Revision = revision
	resp.TrafficMockRules = make([]*apitraffic.TrafficMock, 0, len(out))
	for i := range out {
		resp.TrafficMockRules = append(resp.TrafficMockRules, out[i].ToTrafficMockSpec())
	}
	return resp
}

func (s *Server) findServiceAlias(req *apiservice.Service) *svctypes.Service {
	// 获取源服务
	aliasFor := s.getServiceCache(req.GetName(), req.GetNamespace())
	if aliasFor == nil {
		aliasFor = &svctypes.Service{
			Namespace: req.GetNamespace(),
			Name:      req.GetName(),
		}
	}
	return aliasFor
}

func CreateCommonDiscoverResponse(req *apiservice.Service,
	dT apiservice.DiscoverResponse_DiscoverResponseType) *apiservice.DiscoverResponse {
	return createCommonDiscoverResponse(req, dT)
}

func createCommonDiscoverResponse(req *apiservice.Service,
	dT apiservice.DiscoverResponse_DiscoverResponseType) *apiservice.DiscoverResponse {
	return &apiservice.DiscoverResponse{
		Code: uint32(apimodel.Code_ExecuteSuccess),
		Info: api.Code2Info(uint32(apimodel.Code_ExecuteSuccess)),
		Type: dT,
		Service: &apiservice.Service{
			Name:      req.GetName(),
			Namespace: req.GetNamespace(),
		},
	}
}

// 根据服务名获取服务缓存数据
// 注意，如果是服务别名查询，这里会返回别名的源服务，不会返回别名
func (s *Server) getServiceCache(name string, namespace string) *svctypes.Service {
	sc := s.caches.Service()
	service := sc.GetServiceByName(name, namespace)
	if service == nil {
		return nil
	}
	// 如果是服务别名，继续查找一下
	if service.IsAlias() {
		service = sc.GetServiceByID(service.Reference)
		if service == nil {
			return nil
		}
	}

	if service.Meta == nil {
		service.Meta = make(map[string]string)
	}
	return service
}
