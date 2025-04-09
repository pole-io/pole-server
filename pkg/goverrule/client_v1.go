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

	"github.com/golang/protobuf/ptypes/wrappers"
	"go.uber.org/zap"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// GetRoutingConfigWithCache 获取缓存中的路由配置信息
func (s *Server) GetRoutingConfigWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := createCommonDiscoverResponse(req, apiservice.DiscoverResponse_ROUTING)
	aliasFor := s.findServiceAlias(req)

	out, err := s.caches.RoutingConfig().GetRouterConfig(aliasFor.ID, aliasFor.Name, aliasFor.Namespace)
	if err != nil {
		log.Error("[Server][Service][Routing] discover routing", utils.RequestID(ctx), zap.Error(err))
		return api.NewDiscoverRoutingResponse(apimodel.Code_ExecuteException, req)
	}
	if out == nil {
		return resp
	}

	// 获取路由数据，并对比revision
	if out.GetRevision().GetValue() == req.GetRevision().GetValue() {
		return api.NewDiscoverRoutingResponse(apimodel.Code_DataNoChange, req)
	}

	// 数据不一致，发生了改变
	// 数据格式转换，service只需要返回二元组与routing的revision
	resp.Service.Revision = out.GetRevision()
	resp.Routing = out
	resp.AliasFor = &apiservice.Service{
		Name:      protobuf.NewStringValue(aliasFor.Name),
		Namespace: protobuf.NewStringValue(aliasFor.Namespace),
	}
	return resp
}

// GetRateLimitWithCache 获取缓存中的限流规则信息
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
	if req.GetRevision().GetValue() == revision {
		return api.NewDiscoverRateLimitResponse(apimodel.Code_DataNoChange, req)
	}
	resp.RateLimit = &apitraffic.RateLimit{
		Revision: protobuf.NewStringValue(revision),
		Rules:    []*apitraffic.Rule{},
	}
	for i := range rules {
		rateLimit, err := rateLimit2Client(req.GetName().GetValue(), req.GetNamespace().GetValue(), rules[i])
		if rateLimit == nil || err != nil {
			continue
		}
		resp.RateLimit.Rules = append(resp.RateLimit.Rules, rateLimit)
	}

	// 塞入源服务信息数据
	resp.AliasFor = &apiservice.Service{
		Namespace: protobuf.NewStringValue(aliasFor.Namespace),
		Name:      protobuf.NewStringValue(aliasFor.Name),
	}
	// 服务名和request保持一致
	resp.Service = &apiservice.Service{
		Name:      req.GetName(),
		Namespace: req.GetNamespace(),
		Revision:  protobuf.NewStringValue(revision),
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

	if req.GetRevision().GetValue() == out.Revision {
		return api.NewDiscoverFaultDetectorResponse(apimodel.Code_DataNoChange, req)
	}

	// 数据不一致，发生了改变
	var err error
	resp.AliasFor = &apiservice.Service{
		Name:      protobuf.NewStringValue(aliasFor.Name),
		Namespace: protobuf.NewStringValue(aliasFor.Namespace),
	}
	resp.Service.Revision = protobuf.NewStringValue(out.Revision)
	resp.FaultDetector, err = faultDetectRule2ClientAPI(out)
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
	if len(req.GetRevision().GetValue()) > 0 && req.GetRevision().GetValue() == out.Revision {
		return api.NewDiscoverCircuitBreakerResponse(apimodel.Code_DataNoChange, req)
	}

	// 数据不一致，发生了改变
	var err error
	resp.AliasFor = &apiservice.Service{
		Name:      protobuf.NewStringValue(aliasFor.Name),
		Namespace: protobuf.NewStringValue(aliasFor.Namespace),
	}
	resp.Service.Revision = protobuf.NewStringValue(out.Revision)
	resp.CircuitBreaker, err = circuitBreaker2ClientAPI(out, req.GetName().GetValue(), req.GetNamespace().GetValue())
	if err != nil {
		log.Error(err.Error(), utils.RequestID(ctx))
		return api.NewDiscoverCircuitBreakerResponse(apimodel.Code_ExecuteException, req)
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
	if len(req.GetRevision().GetValue()) > 0 && req.GetRevision().GetValue() == revision {
		return api.NewDiscoverLaneResponse(apimodel.Code_DataNoChange, req)
	}

	resp.AliasFor = &apiservice.Service{
		Name:      protobuf.NewStringValue(aliasFor.Name),
		Namespace: protobuf.NewStringValue(aliasFor.Namespace),
	}
	resp.Service.Revision = protobuf.NewStringValue(revision)
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

	out, err := s.caches.RoutingConfig().GetRouterConfigV2(aliasFor.ID, aliasFor.Name, aliasFor.Namespace)
	if err != nil {
		log.Error("[Server][Service][Routing] discover routing", utils.RequestID(ctx), zap.Error(err))
		return api.NewDiscoverRoutingResponse(apimodel.Code_ExecuteException, req)
	}
	if out == nil {
		return resp
	}

	// 获取路由数据，并对比revision
	if out.GetRevision().GetValue() == req.GetRevision().GetValue() {
		return api.NewDiscoverRoutingResponse(apimodel.Code_DataNoChange, req)
	}

	// 数据不一致，发生了改变
	// 数据格式转换，service只需要返回二元组与routing的revision
	resp.Service.Revision = out.GetRevision()
	resp.CustomRouteRules = out.Rules
	resp.AliasFor = &apiservice.Service{
		Name:      protobuf.NewStringValue(aliasFor.Name),
		Namespace: protobuf.NewStringValue(aliasFor.Namespace),
	}
	return resp
}

func (s *Server) findServiceAlias(req *apiservice.Service) *svctypes.Service {
	// 获取源服务
	aliasFor := s.getServiceCache(req.GetName().GetValue(), req.GetNamespace().GetValue())
	if aliasFor == nil {
		aliasFor = &svctypes.Service{
			Namespace: req.GetNamespace().GetValue(),
			Name:      req.GetName().GetValue(),
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
		Code: &wrappers.UInt32Value{Value: uint32(apimodel.Code_ExecuteSuccess)},
		Info: &wrappers.StringValue{Value: api.Code2Info(uint32(apimodel.Code_ExecuteSuccess))},
		Type: dT,
		Service: &apiservice.Service{
			Name:      req.GetName(),
			Namespace: req.GetNamespace(),
		},
	}
}

// 获取顶级服务ID
// 没有顶级ID，则返回自身
func (s *Server) getSourceServiceID(service *svctypes.Service) string {
	if service == nil || service.ID == "" {
		return ""
	}
	// 找到parent服务，最多两级，因此不用递归查找
	if service.IsAlias() {
		return service.Reference
	}

	return service.ID
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

func (s *Server) commonCheckDiscoverRequest(req *apiservice.Service, resp *apiservice.DiscoverResponse) bool {
	if s.caches == nil {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_ClientAPINotOpen))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = req
		return false
	}
	if req == nil {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_EmptyRequest))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = req
		return false
	}

	if req.GetName().GetValue() == "" {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_InvalidServiceName))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = req
		return false
	}
	if req.GetNamespace().GetValue() == "" {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_InvalidNamespaceName))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = req
		return false
	}

	return true
}
