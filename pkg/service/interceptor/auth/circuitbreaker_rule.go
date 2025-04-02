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

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	apifault "github.com/polarismesh/specification/source/go/api/v1/fault_tolerance"
	"github.com/polarismesh/specification/source/go/api/v1/security"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

func (svr *Server) CreateCircuitBreakerRules(
	ctx context.Context, request []*apifault.CircuitBreakerRule) *apiservice.BatchWriteResponse {
	authCtx := svr.collectCircuitBreakerRuleV2(ctx, request, authtypes.Create,
		authtypes.CreateCircuitBreakerRules)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	rsp := svr.nextSvr.CreateCircuitBreakerRules(ctx, request)
	for index := range rsp.Responses {
		item := rsp.GetResponses()[index].GetData()
		rule := &apifault.CircuitBreakerRule{}
		_ = anypb.UnmarshalTo(item, rule, proto.UnmarshalOptions{})
		_ = svr.afterRuleResource(ctx, types.RRouting, authtypes.ResourceEntry{
			ID:   rule.Id,
			Type: security.ResourceType_CircuitBreakerRules,
		}, false)
	}
	return rsp
}

func (svr *Server) DeleteCircuitBreakerRules(
	ctx context.Context, request []*apifault.CircuitBreakerRule) *apiservice.BatchWriteResponse {
	authCtx := svr.collectCircuitBreakerRuleV2(ctx, request, authtypes.Delete,
		authtypes.DeleteCircuitBreakerRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	rsp := svr.nextSvr.DeleteCircuitBreakerRules(ctx, request)
	for index := range rsp.Responses {
		item := rsp.GetResponses()[index].GetData()
		rule := &apifault.CircuitBreakerRule{}
		_ = anypb.UnmarshalTo(item, rule, proto.UnmarshalOptions{})
		_ = svr.afterRuleResource(ctx, types.RRouting, authtypes.ResourceEntry{
			ID:   rule.Id,
			Type: security.ResourceType_CircuitBreakerRules,
		}, true)
	}
	return rsp
}

func (svr *Server) EnableCircuitBreakerRules(
	ctx context.Context, request []*apifault.CircuitBreakerRule) *apiservice.BatchWriteResponse {
	authCtx := svr.collectCircuitBreakerRuleV2(ctx, request, authtypes.Modify,
		authtypes.EnableCircuitBreakerRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.EnableCircuitBreakerRules(ctx, request)
}

func (svr *Server) UpdateCircuitBreakerRules(
	ctx context.Context, request []*apifault.CircuitBreakerRule) *apiservice.BatchWriteResponse {
	authCtx := svr.collectCircuitBreakerRuleV2(ctx, request, authtypes.Modify,
		authtypes.UpdateCircuitBreakerRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.UpdateCircuitBreakerRules(ctx, request)
}

func (svr *Server) GetCircuitBreakerRules(
	ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse {
	authCtx := svr.collectCircuitBreakerRuleV2(ctx, nil, authtypes.Read,
		authtypes.DescribeCircuitBreakerRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendCircuitBreakerRulePredicate(ctx,
		func(ctx context.Context, cbr *rules.CircuitBreakerRule) bool {
			return svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
				Type:     security.ResourceType_CircuitBreakerRules,
				ID:       cbr.ID,
				Metadata: cbr.Proto.Metadata,
			})
		})
	authCtx.SetRequestContext(ctx)

	resp := svr.nextSvr.GetCircuitBreakerRules(ctx, query)

	for index := range resp.Data {
		item := &apifault.CircuitBreakerRule{}
		_ = anypb.UnmarshalTo(resp.Data[index], item, proto.UnmarshalOptions{})
		authCtx.SetAccessResources(map[security.ResourceType][]authtypes.ResourceEntry{
			security.ResourceType_CircuitBreakerRules: {
				{
					Type:     apisecurity.ResourceType_CircuitBreakerRules,
					ID:       item.GetId(),
					Metadata: item.Metadata,
				},
			},
		})

		// 检查 write 操作权限
		authCtx.SetMethod([]authtypes.ServerFunctionName{
			authtypes.UpdateCircuitBreakerRules,
			authtypes.EnableCircuitBreakerRules,
		})
		// 如果检查不通过，设置 editable 为 false
		if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Editable = false
		}

		// 检查 delete 操作权限
		authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.DeleteCircuitBreakerRules})
		// 如果检查不通过，设置 editable 为 false
		if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Deleteable = false
		}
		_ = anypb.MarshalFrom(resp.Data[index], item, proto.MarshalOptions{})
	}
	return resp
}
