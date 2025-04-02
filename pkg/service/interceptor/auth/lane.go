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

	"github.com/polarismesh/specification/source/go/api/v1/security"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// CreateLaneGroups 批量创建泳道组
func (svr *Server) CreateLaneGroups(ctx context.Context, reqs []*apitraffic.LaneGroup) *apiservice.BatchWriteResponse {

	authCtx := svr.collectLaneRuleAuthContext(ctx, reqs, authtypes.Create, authtypes.CreateLaneGroups)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	rsp := svr.nextSvr.CreateLaneGroups(ctx, reqs)
	for index := range rsp.Responses {
		item := rsp.GetResponses()[index].GetData()
		rule := &apitraffic.LaneGroup{}
		_ = anypb.UnmarshalTo(item, rule, proto.UnmarshalOptions{})
		_ = svr.afterRuleResource(ctx, types.RLaneGroup, authtypes.ResourceEntry{
			ID:   rule.Id,
			Type: security.ResourceType_LaneRules,
		}, false)
	}
	return rsp
}

// UpdateLaneGroups 批量更新泳道组
func (svr *Server) UpdateLaneGroups(ctx context.Context, reqs []*apitraffic.LaneGroup) *apiservice.BatchWriteResponse {
	authCtx := svr.collectLaneRuleAuthContext(ctx, reqs, authtypes.Modify, authtypes.UpdateLaneGroups)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.UpdateLaneGroups(ctx, reqs)
}

// DeleteLaneGroups 批量删除泳道组
func (svr *Server) DeleteLaneGroups(ctx context.Context, reqs []*apitraffic.LaneGroup) *apiservice.BatchWriteResponse {
	authCtx := svr.collectLaneRuleAuthContext(ctx, reqs, authtypes.Delete, authtypes.DeleteLaneGroups)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	rsp := svr.nextSvr.DeleteLaneGroups(ctx, reqs)
	for index := range rsp.Responses {
		item := rsp.GetResponses()[index].GetData()
		rule := &apitraffic.LaneGroup{}
		_ = anypb.UnmarshalTo(item, rule, proto.UnmarshalOptions{})
		_ = svr.afterRuleResource(ctx, types.RLaneGroup, authtypes.ResourceEntry{
			ID:   rule.Id,
			Type: security.ResourceType_LaneRules,
		}, true)
	}
	return rsp
}

// GetLaneGroups 查询泳道组列表
func (svr *Server) GetLaneGroups(ctx context.Context, filter map[string]string) *apiservice.BatchQueryResponse {
	authCtx := svr.collectFaultDetectAuthContext(ctx, nil, authtypes.Read, authtypes.DescribeFaultDetectRules)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendLaneRulePredicate(ctx, func(ctx context.Context, cbr *rules.LaneGroupProto) bool {
		return svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_LaneRules,
			ID:       cbr.ID,
			Metadata: cbr.Proto.Metadata,
		})
	})
	authCtx.SetRequestContext(ctx)

	resp := svr.nextSvr.GetLaneGroups(ctx, filter)

	for index := range resp.Data {
		item := &apitraffic.LaneGroup{}
		_ = anypb.UnmarshalTo(resp.Data[index], item, proto.UnmarshalOptions{})
		authCtx.SetAccessResources(map[security.ResourceType][]authtypes.ResourceEntry{
			security.ResourceType_LaneRules: {
				{
					Type:     apisecurity.ResourceType_LaneRules,
					ID:       item.GetId(),
					Metadata: item.Metadata,
				},
			},
		})

		// 检查 write 操作权限
		authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.UpdateLaneGroups, authtypes.EnableLaneGroups})
		// 如果检查不通过，设置 editable 为 false
		if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Editable = false
		}

		// 检查 delete 操作权限
		authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.DeleteLaneGroups})
		// 如果检查不通过，设置 editable 为 false
		if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Deleteable = false
		}
		_ = anypb.MarshalFrom(resp.Data[index], item, proto.MarshalOptions{})
	}
	return resp
}

// collectLaneRuleAuthContext 收集全链路灰度规则
func (svr *Server) collectLaneRuleAuthContext(ctx context.Context, req []*apitraffic.LaneGroup,
	op authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {

	resources := make([]authtypes.ResourceEntry, 0, len(req))
	for i := range req {
		saveRule := svr.Cache().LaneRule().GetRule(req[i].GetId())
		if saveRule != nil {
			resources = append(resources, authtypes.ResourceEntry{
				Type:     apisecurity.ResourceType_LaneRules,
				ID:       saveRule.ID,
				Metadata: saveRule.Labels,
			})
		}
	}

	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(op),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_LaneRules: resources,
		}),
	)
}
