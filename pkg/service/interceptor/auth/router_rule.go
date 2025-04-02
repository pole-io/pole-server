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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// CreateRoutingConfigsV2 批量创建路由配置
func (svr *Server) CreateRoutingConfigs(ctx context.Context,
	req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse {

	// TODO not support RouteRuleV2 resource auth, so we set op is read
	authCtx := svr.collectRouteRuleV2AuthContext(ctx, req, authtypes.Create, authtypes.CreateRouteRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	resp := svr.nextSvr.CreateRoutingConfigs(ctx, req)

	for index := range resp.Responses {
		item := resp.GetResponses()[index].GetData()
		rule := &apitraffic.RouteRule{}
		_ = anypb.UnmarshalTo(item, rule, proto.UnmarshalOptions{})
		_ = svr.afterRuleResource(ctx, types.RRouting, authtypes.ResourceEntry{
			ID:   rule.Id,
			Type: security.ResourceType_RouteRules,
		}, false)
	}
	return resp
}

// DeleteRoutingConfigs 批量删除路由配置
func (svr *Server) DeleteRoutingConfigs(ctx context.Context,
	req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse {

	authCtx := svr.collectRouteRuleV2AuthContext(ctx, req, authtypes.Delete, authtypes.DeleteRouteRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	resp := svr.nextSvr.DeleteRoutingConfigs(ctx, req)

	for index := range resp.Responses {
		item := resp.GetResponses()[index].GetData()
		rule := &apitraffic.RouteRule{}
		_ = anypb.UnmarshalTo(item, rule, proto.UnmarshalOptions{})
		_ = svr.afterRuleResource(ctx, types.RRouting, authtypes.ResourceEntry{
			ID:   rule.Id,
			Type: security.ResourceType_RouteRules,
		}, true)
	}
	return resp
}

// UpdateRoutingConfigsV2 批量更新路由配置
func (svr *Server) UpdateRoutingConfigs(ctx context.Context,
	req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse {

	authCtx := svr.collectRouteRuleV2AuthContext(ctx, req, authtypes.Modify, authtypes.UpdateRouteRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.UpdateRoutingConfigs(ctx, req)
}

// EnableRoutings batch enable routing rules
func (svr *Server) EnableRoutings(ctx context.Context,
	req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse {

	authCtx := svr.collectRouteRuleV2AuthContext(ctx, req, authtypes.Modify, authtypes.EnableRouteRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.EnableRoutings(ctx, req)
}

// QueryRoutingConfigsV2 提供给OSS的查询路由配置的接口
func (svr *Server) QueryRoutingConfigs(ctx context.Context,
	query map[string]string) *apiservice.BatchQueryResponse {
	authCtx := svr.collectRouteRuleV2AuthContext(ctx, nil, authtypes.Read, authtypes.DescribeRouteRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendRouterRulePredicate(ctx, func(ctx context.Context, cbr *rules.ExtendRouterConfig) bool {
		return svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_RouteRules,
			ID:       cbr.ID,
			Metadata: cbr.Metadata,
		})
	})
	authCtx.SetRequestContext(ctx)

	resp := svr.nextSvr.QueryRoutingConfigs(ctx, query)
	for index := range resp.Data {
		item := &apitraffic.RouteRule{}
		_ = anypb.UnmarshalTo(resp.Data[index], item, proto.UnmarshalOptions{})
		authCtx.SetAccessResources(map[security.ResourceType][]authtypes.ResourceEntry{
			security.ResourceType_RouteRules: {
				{
					Type:     apisecurity.ResourceType_RouteRules,
					ID:       item.GetId(),
					Metadata: item.Metadata,
				},
			},
		})

		// 检查 write 操作权限
		authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.UpdateRouteRules, authtypes.EnableRouteRules})
		// 如果检查不通过，设置 editable 为 false
		if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Editable = false
		}

		// 检查 delete 操作权限
		authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.DeleteRouteRules})
		// 如果检查不通过，设置 editable 为 false
		if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Deleteable = false
		}
		_ = anypb.MarshalFrom(resp.Data[index], item, proto.MarshalOptions{})
	}
	return resp
}
