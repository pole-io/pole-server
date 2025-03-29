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
	"github.com/polarismesh/specification/source/go/api/v1/security"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	cachetypes "github.com/pole-io/pole-server/cache/api"
	api "github.com/pole-io/pole-server/common/api/v1"
	"github.com/pole-io/pole-server/common/model"
	authcommon "github.com/pole-io/pole-server/common/model/auth"
	"github.com/pole-io/pole-server/common/utils"
)

func (svr *Server) CreateFaultDetectRules(
	ctx context.Context, request []*apifault.FaultDetectRule) *apiservice.BatchWriteResponse {

	authCtx := svr.collectFaultDetectAuthContext(ctx, request, authcommon.Create, authcommon.CreateFaultDetectRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authcommon.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)
	return svr.nextSvr.CreateFaultDetectRules(ctx, request)
}

func (svr *Server) DeleteFaultDetectRules(
	ctx context.Context, request []*apifault.FaultDetectRule) *apiservice.BatchWriteResponse {

	authCtx := svr.collectFaultDetectAuthContext(ctx, request, authcommon.Delete, authcommon.DeleteFaultDetectRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authcommon.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)
	return svr.nextSvr.DeleteFaultDetectRules(ctx, request)
}

func (svr *Server) UpdateFaultDetectRules(
	ctx context.Context, request []*apifault.FaultDetectRule) *apiservice.BatchWriteResponse {

	authCtx := svr.collectFaultDetectAuthContext(ctx, request, authcommon.Modify, authcommon.UpdateFaultDetectRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authcommon.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)
	return svr.nextSvr.UpdateFaultDetectRules(ctx, request)
}

func (svr *Server) GetFaultDetectRules(
	ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse {
	authCtx := svr.collectFaultDetectAuthContext(ctx, nil, authcommon.Read, authcommon.DescribeFaultDetectRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authcommon.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	ctx = cachetypes.AppendFaultDetectRulePredicate(ctx, func(ctx context.Context, cbr *model.FaultDetectRule) bool {
		return svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authcommon.ResourceEntry{
			Type:     security.ResourceType_FaultDetectRules,
			ID:       cbr.ID,
			Metadata: cbr.Proto.GetMetadata(),
		})
	})
	authCtx.SetRequestContext(ctx)

	resp := svr.nextSvr.GetFaultDetectRules(ctx, query)

	for index := range resp.Data {
		item := &apifault.FaultDetectRule{}
		_ = anypb.UnmarshalTo(resp.Data[index], item, proto.UnmarshalOptions{})
		authCtx.SetAccessResources(map[security.ResourceType][]authcommon.ResourceEntry{
			security.ResourceType_FaultDetectRules: {
				{
					Type:     apisecurity.ResourceType_FaultDetectRules,
					ID:       item.GetId(),
					Metadata: item.Metadata,
				},
			},
		})

		// 检查 write 操作权限
		authCtx.SetMethod([]authcommon.ServerFunctionName{authcommon.UpdateFaultDetectRules, authcommon.EnableFaultDetectRules})
		// 如果检查不通过，设置 editable 为 false
		if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Editable = false
		}

		// 检查 delete 操作权限
		authCtx.SetMethod([]authcommon.ServerFunctionName{authcommon.DeleteFaultDetectRules})
		// 如果检查不通过，设置 editable 为 false
		if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Deleteable = false
		}
		_ = anypb.MarshalFrom(resp.Data[index], item, proto.MarshalOptions{})
	}
	return resp
}
