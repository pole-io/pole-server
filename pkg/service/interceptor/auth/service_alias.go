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

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	"github.com/pole-io/specification/source/go/api/v1/security"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// CreateServiceAlias creates a service alias
func (svr *Server) CreateServiceAlias(
	ctx context.Context, req *apiservice.ServiceAlias) *apimodel.Response {
	authCtx := svr.collectServiceAliasAuthContext(
		ctx, []*apiservice.ServiceAlias{req}, authtypes.Create, authtypes.CreateServiceAlias)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewServiceAliasResponse(authtypes.ConvertToErrCode(err), req)
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	// 填充 ownerId 信息数据
	ownerId := utils.ParseOwnerID(ctx)
	if len(ownerId) > 0 {
		req.Owners = string(ownerId)
	}

	return svr.nextSvr.CreateServiceAlias(ctx, req)
}

// DeleteServiceAliases deletes service aliases
func (svr *Server) DeleteServiceAliases(ctx context.Context,
	reqs []*apiservice.ServiceAlias) *apimodel.BatchWriteResponse {
	authCtx := svr.collectServiceAliasAuthContext(ctx, reqs, authtypes.Delete, authtypes.DeleteServiceAliases)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.DeleteServiceAliases(ctx, reqs)
}

// UpdateServiceAlias updates service alias
func (svr *Server) UpdateServiceAlias(
	ctx context.Context, req *apiservice.ServiceAlias) *apimodel.Response {
	authCtx := svr.collectServiceAliasAuthContext(
		ctx, []*apiservice.ServiceAlias{req}, authtypes.Modify, authtypes.UpdateServiceAlias)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewServiceAliasResponse(authtypes.ConvertToErrCode(err), req)
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.UpdateServiceAlias(ctx, req)
}

// GetServiceAliases gets service aliases
func (svr *Server) GetServiceAliases(ctx context.Context,
	query map[string]string) *apimodel.BatchQueryResponse {
	authCtx := svr.collectServiceAliasAuthContext(ctx, nil, authtypes.Read, authtypes.DescribeServiceAliases)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendServicePredicate(ctx, func(ctx context.Context, cbr *svctypes.Service) bool {
		sourceSvc := svr.Cache().Service().GetServiceByID(cbr.Reference)
		if sourceSvc == nil {
			return false
		}
		return svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_Services,
			ID:       sourceSvc.ID,
			Metadata: sourceSvc.Meta,
		})
	})

	authCtx.SetRequestContext(ctx)

	resp := svr.nextSvr.GetServiceAliases(ctx, query)
	return svr.filterServiceAliasesByPermission(ctx, resp, authCtx)
}

// filterServiceAliasesByPermission 基于 BatchQueryResponse.Data 做别名级权限过滤，仅保留对源服务有读权限的别名
func (svr *Server) filterServiceAliasesByPermission(
	ctx context.Context, resp *apimodel.BatchQueryResponse, authCtx *authtypes.AcquireContext) *apimodel.BatchQueryResponse {
	if resp == nil || resp.Code != uint32(apimodel.Code_ExecuteSuccess) || len(resp.Data) == 0 {
		return resp
	}
	filtered := make([]*anypb.Any, 0, len(resp.Data))
	for _, anyData := range resp.Data {
		if anyData == nil {
			continue
		}
		alias := &apiservice.ServiceAlias{}
		if err := anypb.UnmarshalTo(anyData, alias, proto.UnmarshalOptions{}); err != nil {
			continue
		}
		// 通过源服务名与命名空间解析源服务（API 层 ServiceAlias 无 Reference 字段时）
		sourceSvc := svr.Cache().Service().GetServiceByName(alias.GetService(), alias.GetNamespace())
		if sourceSvc == nil {
			continue
		}
		ok := svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_Services,
			ID:       sourceSvc.ID,
			Metadata: sourceSvc.Meta,
		})
		if ok {
			filtered = append(filtered, anyData)
		}
	}
	resp.Data = filtered
	resp.Size = uint32(len(filtered))
	return resp
}
