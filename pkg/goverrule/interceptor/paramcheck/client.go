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

package paramcheck

import (
	"context"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/service"
)

var (
	clientFilterAttributes = map[string]struct{}{
		"type":    {},
		"host":    {},
		"limit":   {},
		"offset":  {},
		"version": {},
	}
)

// GetRoutingConfigWithCache User Client Get Service Routing Configuration Information
func (s *Server) GetRoutingConfigWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := service.CreateCommonDiscoverResponse(req, apiservice.DiscoverResponse_ROUTING)
	if !s.commonCheckDiscoverRequest(req, resp) {
		return resp
	}
	return s.nextSvr.GetRoutingConfigWithCache(ctx, req)
}

// GetRateLimitWithCache User Client Get Service Limit Configuration Information
func (s *Server) GetRateLimitWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := service.CreateCommonDiscoverResponse(req, apiservice.DiscoverResponse_RATE_LIMIT)
	if !s.commonCheckDiscoverRequest(req, resp) {
		return resp
	}
	return s.nextSvr.GetRateLimitWithCache(ctx, req)
}

// GetCircuitBreakerWithCache Fuse configuration information for obtaining services for clients
func (s *Server) GetCircuitBreakerWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := service.CreateCommonDiscoverResponse(req, apiservice.DiscoverResponse_CIRCUIT_BREAKER)
	if !s.commonCheckDiscoverRequest(req, resp) {
		return resp
	}
	return s.nextSvr.GetCircuitBreakerWithCache(ctx, req)
}

// GetFaultDetectWithCache User Client Get FaultDetect Rule Information
func (s *Server) GetFaultDetectWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := service.CreateCommonDiscoverResponse(req, apiservice.DiscoverResponse_FAULT_DETECTOR)
	if !s.commonCheckDiscoverRequest(req, resp) {
		return resp
	}
	return s.nextSvr.GetFaultDetectWithCache(ctx, req)
}

// GetLaneRuleWithCache fetch lane rule by client
func (s *Server) GetLaneRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := service.CreateCommonDiscoverResponse(req, apiservice.DiscoverResponse_LANE)
	if !s.commonCheckDiscoverRequest(req, resp) {
		return resp
	}
	return s.nextSvr.GetLaneRuleWithCache(ctx, req)
}

// GetRouterRuleWithCache .
func (s *Server) GetRouterRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := service.CreateCommonDiscoverResponse(req, apiservice.DiscoverResponse_CUSTOM_ROUTE_RULE)
	if !s.commonCheckDiscoverRequest(req, resp) {
		return resp
	}
	return s.nextSvr.GetRouterRuleWithCache(ctx, req)
}

func (s *Server) commonCheckDiscoverRequest(req *apiservice.Service, resp *apiservice.DiscoverResponse) bool {
	if s.nextSvr.Cache() == nil {
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
