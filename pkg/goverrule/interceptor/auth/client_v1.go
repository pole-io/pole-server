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

	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// GetRoutingConfigWithCache is the interface for getting routing config with cache
func (svr *Server) GetRoutingConfigWithCache(
	ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {

	authCtx := svr.collectServiceAuthContext(
		ctx, []*apiservice.Service{req}, authtypes.Read, authtypes.DiscoverRouterRule)
	if _, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx); err != nil {
		return api.NewDiscoverResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetRoutingConfigWithCache(ctx, req)
}

// GetRateLimitWithCache is the interface for getting rate limit with cache
func (svr *Server) GetRateLimitWithCache(
	ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {

	authCtx := svr.collectServiceAuthContext(
		ctx, []*apiservice.Service{req}, authtypes.Read, authtypes.DiscoverRateLimitRule)
	if _, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx); err != nil {
		return api.NewDiscoverResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetRateLimitWithCache(ctx, req)
}

// GetCircuitBreakerWithCache is the interface for getting a circuit breaker with cache
func (svr *Server) GetCircuitBreakerWithCache(
	ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {

	authCtx := svr.collectServiceAuthContext(
		ctx, []*apiservice.Service{req}, authtypes.Read, authtypes.DiscoverCircuitBreakerRule)
	if _, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx); err != nil {
		return api.NewDiscoverResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetCircuitBreakerWithCache(ctx, req)
}

// GetFaultDetectWithCache 获取主动探测规则列表
func (svr *Server) GetFaultDetectWithCache(
	ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {

	authCtx := svr.collectServiceAuthContext(
		ctx, []*apiservice.Service{req}, authtypes.Read, authtypes.DiscoverFaultDetectRule)
	if _, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx); err != nil {
		return api.NewDiscoverResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetFaultDetectWithCache(ctx, req)
}

// GetLaneRuleWithCache fetch lane rules by client
func (svr *Server) GetLaneRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	authCtx := svr.collectServiceAuthContext(
		ctx, []*apiservice.Service{req}, authtypes.Read, authtypes.DiscoverLaneRule)
	if _, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx); err != nil {
		return api.NewDiscoverResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetLaneRuleWithCache(ctx, req)
}

// GetRouterRuleWithCache .
func (svr *Server) GetRouterRuleWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	authCtx := svr.collectServiceAuthContext(
		ctx, []*apiservice.Service{req}, authtypes.Read, authtypes.DiscoverRouterRule)
	if _, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx); err != nil {
		return api.NewDiscoverResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetRouterRuleWithCache(ctx, req)
}
