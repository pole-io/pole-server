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

	"github.com/polarismesh/specification/source/go/api/v1/security"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	cacheapi "github.com/pole-io/pole-server/apis/cache"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// CreateRateLimits creates rate limits for a namespace.
func (svr *Server) CreateRateLimits(
	ctx context.Context, reqs []*apitraffic.Rule) *apiservice.BatchWriteResponse {
	authCtx := svr.collectRateLimitAuthContext(ctx, reqs, authtypes.Create, authtypes.CreateRateLimitRules)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	rsp := svr.nextSvr.CreateRateLimits(ctx, reqs)
	for index := range rsp.Responses {
		_ = svr.afterRuleResource(ctx, types.RRateLimit, authtypes.ResourceEntry{
			ID:   rsp.Responses[index].GetRateLimit().GetId().GetValue(),
			Type: security.ResourceType_RateLimitRules,
		}, false)
	}
	return rsp
}

// DeleteRateLimits deletes rate limits for a namespace.
func (svr *Server) DeleteRateLimits(
	ctx context.Context, reqs []*apitraffic.Rule) *apiservice.BatchWriteResponse {
	authCtx := svr.collectRateLimitAuthContext(ctx, reqs, authtypes.Delete, authtypes.DeleteRateLimitRules)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	rsp := svr.nextSvr.DeleteRateLimits(ctx, reqs)
	for index := range rsp.Responses {
		_ = svr.afterRuleResource(ctx, types.RRateLimit, authtypes.ResourceEntry{
			ID:   rsp.Responses[index].GetRateLimit().GetId().GetValue(),
			Type: security.ResourceType_RateLimitRules,
		}, true)
	}
	return rsp
}

// UpdateRateLimits updates rate limits for a namespace.
func (svr *Server) UpdateRateLimits(
	ctx context.Context, reqs []*apitraffic.Rule) *apiservice.BatchWriteResponse {
	authCtx := svr.collectRateLimitAuthContext(ctx, reqs, authtypes.Modify, authtypes.UpdateRateLimitRules)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	return svr.nextSvr.UpdateRateLimits(ctx, reqs)
}

// EnableRateLimits 启用限流规则
func (svr *Server) EnableRateLimits(
	ctx context.Context, reqs []*apitraffic.Rule) *apiservice.BatchWriteResponse {
	authCtx := svr.collectRateLimitAuthContext(ctx, reqs, authtypes.Read, authtypes.EnableRateLimitRules)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	return svr.nextSvr.EnableRateLimits(ctx, reqs)
}

// GetRateLimits gets rate limits for a namespace.
func (svr *Server) GetRateLimits(
	ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse {
	authCtx := svr.collectRateLimitAuthContext(ctx, nil, authtypes.Read, authtypes.DescribeRateLimitRules)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewAuthBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendRatelimitRulePredicate(ctx, func(ctx context.Context, cbr *rules.RateLimit) bool {
		return svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_RateLimitRules,
			ID:       cbr.ID,
			Metadata: cbr.Proto.Metadata,
		})
	})
	authCtx.SetRequestContext(ctx)

	resp := svr.nextSvr.GetRateLimits(ctx, query)

	for index := range resp.RateLimits {
		item := resp.RateLimits[index]
		authCtx.SetAccessResources(map[security.ResourceType][]authtypes.ResourceEntry{
			security.ResourceType_RateLimitRules: {
				{
					Type:     apisecurity.ResourceType_RateLimitRules,
					ID:       item.GetId().GetValue(),
					Metadata: item.Metadata,
				},
			},
		})

		// 检查 write 操作权限
		authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.UpdateRateLimitRules, authtypes.EnableRateLimitRules})
		// 如果检查不通过，设置 editable 为 false
		if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Editable = false
		}

		// 检查 delete 操作权限
		authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.DeleteRateLimitRules})
		// 如果检查不通过，设置 editable 为 false
		if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Deleteable = false
		}
	}

	return resp
}
