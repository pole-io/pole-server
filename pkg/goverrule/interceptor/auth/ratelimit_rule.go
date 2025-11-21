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

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	"github.com/pole-io/specification/source/go/api/v1/security"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// CreateRateLimits 为命名空间创建限流规则
func (svr *Server) CreateRateLimits(
	ctx context.Context, reqs []*apitraffic.RateLimit) *apimodel.BatchWriteResponse {
	authCtx := svr.collectRateLimitAuthContext(ctx, reqs, authtypes.Create, authtypes.CreateRateLimitRules)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	rsp := svr.nextSvr.CreateRateLimits(ctx, reqs)
	for range rsp.Responses {
		// 跳过限流规则的资源处理，因为 Response 不包含限流数据
		_ = svr.afterRuleResource(ctx, types.RRateLimit, authtypes.ResourceEntry{
			ID:   "",
			Type: security.ResourceType_RateLimitRules,
		}, false)
	}
	return rsp
}

// DeleteRateLimits 为命名空间删除限流规则
func (svr *Server) DeleteRateLimits(
	ctx context.Context, reqs []*apitraffic.RateLimit) *apimodel.BatchWriteResponse {
	authCtx := svr.collectRateLimitAuthContext(ctx, reqs, authtypes.Delete, authtypes.DeleteRateLimitRules)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	rsp := svr.nextSvr.DeleteRateLimits(ctx, reqs)
	for range rsp.Responses {
		// 跳过限流规则的资源处理，因为 Response 不包含限流数据
		_ = svr.afterRuleResource(ctx, types.RRateLimit, authtypes.ResourceEntry{
			ID:   "",
			Type: security.ResourceType_RateLimitRules,
		}, true)
	}
	return rsp
}

// UpdateRateLimits 为命名空间更新限流规则
func (svr *Server) UpdateRateLimits(
	ctx context.Context, reqs []*apitraffic.RateLimit) *apimodel.BatchWriteResponse {
	authCtx := svr.collectRateLimitAuthContext(ctx, reqs, authtypes.Modify, authtypes.UpdateRateLimitRules)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.UpdateRateLimits(ctx, reqs)
}

func (svr *Server) GetOneRateLimitRule(ctx context.Context, req *apitraffic.RateLimit) *apimodel.Response {
	authCtx := svr.collectRateLimitAuthContext(ctx, []*apitraffic.RateLimit{req}, authtypes.Read, authtypes.DescribeRateLimitRules)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	resp := svr.nextSvr.GetOneRateLimitRule(ctx, req)
	rule := &apitraffic.RateLimit{}
	_ = anypb.UnmarshalTo(resp.Data, rule, proto.UnmarshalOptions{})
	rule.Editable = true
	rule.Deleteable = true

	authCtx.SetAccessResources(map[security.ResourceType][]authtypes.ResourceEntry{
		security.ResourceType_RateLimitRules: {
			{
				Type:     apisecurity.ResourceType_RateLimitRules,
				ID:       rule.GetId(),
				Metadata: rule.Metadata,
			},
		},
	})

	// 检查 write 操作权限
	authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.UpdateRateLimitRules, authtypes.PublishRouteRules})
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		rule.Editable = false
	}

	// 检查 delete 操作权限
	authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.DeleteRouteRules})
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		rule.Deleteable = false
	}
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, rule)
}

// GetRateLimits 为命名空间获取限流规则
func (svr *Server) GetRateLimits(
	ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	authCtx := svr.collectRateLimitAuthContext(ctx, nil, authtypes.Read, authtypes.DescribeRateLimitRules)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewAuthBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendRatelimitRulePredicate(ctx, func(ctx context.Context, cbr *rules.RateLimit) bool {
		return svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_RateLimitRules,
			ID:       cbr.ID,
			Metadata: cbr.Metadata,
		})
	})
	authCtx.SetRequestContext(ctx)

	resp := svr.nextSvr.GetRateLimits(ctx, query)

	// 从响应中处理限流数据
	if resp.Data != nil {
		for _, anyData := range resp.Data {
			item := &apitraffic.RateLimit{}
			if err := anypb.UnmarshalTo(anyData, item, proto.UnmarshalOptions{}); err != nil {
				continue
			}

			item.Editable = true
			item.Deleteable = true
			authCtx.SetAccessResources(map[security.ResourceType][]authtypes.ResourceEntry{
				security.ResourceType_RateLimitRules: {
					{
						Type:     apisecurity.ResourceType_RateLimitRules,
						ID:       item.GetId(),
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

			// 将数据更新回响应
			if updatedData, err := anypb.New(item); err == nil {
				anyData.Value = updatedData.Value
				anyData.TypeUrl = updatedData.TypeUrl
			}
		}
	}

	return resp
}
