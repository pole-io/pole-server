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

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	api "github.com/pole-io/pole-server/common/api/v1"
	authcommon "github.com/pole-io/pole-server/common/model/auth"
	"github.com/pole-io/pole-server/common/utils"
	"github.com/pole-io/pole-server/service"
)

// CreateInstances create instances
func (svr *Server) CreateInstances(ctx context.Context,
	reqs []*apiservice.Instance) *apiservice.BatchWriteResponse {
	authCtx := svr.collectInstanceAuthContext(ctx, reqs, authcommon.Create, authcommon.CreateInstances)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		resp := api.NewResponseWithMsg(authcommon.ConvertToErrCode(err), err.Error())
		batchResp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
		api.Collect(batchResp, resp)
		return batchResp
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	return svr.nextSvr.CreateInstances(ctx, reqs)
}

// DeleteInstances delete instances
func (svr *Server) DeleteInstances(ctx context.Context,
	reqs []*apiservice.Instance) *apiservice.BatchWriteResponse {
	authCtx := svr.collectInstanceAuthContext(ctx, reqs, authcommon.Delete, authcommon.DeleteInstances)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		resp := api.NewResponseWithMsg(authcommon.ConvertToErrCode(err), err.Error())
		batchResp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
		api.Collect(batchResp, resp)
		return batchResp
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	return svr.nextSvr.DeleteInstances(ctx, reqs)
}

// DeleteInstancesByHost 根据 host 信息进行数据删除
func (svr *Server) DeleteInstancesByHost(ctx context.Context,
	reqs []*apiservice.Instance) *apiservice.BatchWriteResponse {
	authCtx := svr.collectInstanceAuthContext(ctx, reqs, authcommon.Delete, authcommon.DeleteInstancesByHost)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authcommon.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)
	if authcommon.ParseUserRole(ctx) == authcommon.SubAccountUserRole {
		ret := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
		api.Collect(ret, api.NewResponse(apimodel.Code_NotAllowedAccess))
		return ret
	}

	return svr.nextSvr.DeleteInstancesByHost(ctx, reqs)
}

// UpdateInstances update instances
func (svr *Server) UpdateInstances(ctx context.Context,
	reqs []*apiservice.Instance) *apiservice.BatchWriteResponse {
	authCtx := svr.collectInstanceAuthContext(ctx, reqs, authcommon.Modify, authcommon.UpdateInstances)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return api.NewBatchWriteResponseWithMsg(authcommon.ConvertToErrCode(err), err.Error())
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	return svr.nextSvr.UpdateInstances(ctx, reqs)
}

// UpdateInstancesIsolate update instances
func (svr *Server) UpdateInstancesIsolate(ctx context.Context,
	reqs []*apiservice.Instance) *apiservice.BatchWriteResponse {
	authCtx := svr.collectInstanceAuthContext(ctx, reqs, authcommon.Modify, authcommon.UpdateInstancesIsolate)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return api.NewBatchWriteResponseWithMsg(authcommon.ConvertToErrCode(err), err.Error())
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	return svr.nextSvr.UpdateInstancesIsolate(ctx, reqs)
}

// GetInstances get instances
func (svr *Server) GetInstances(ctx context.Context,
	query map[string]string) *apiservice.BatchQueryResponse {
	authCtx := svr.collectInstanceAuthContext(ctx, nil, authcommon.Read, authcommon.DescribeInstances)
	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return api.NewBatchQueryResponseWithMsg(authcommon.ConvertToErrCode(err), err.Error())
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetInstances(ctx, query)
}

// GetInstancesCount get instances to count
func (svr *Server) GetInstancesCount(ctx context.Context) *apiservice.BatchQueryResponse {
	authCtx := svr.collectInstanceAuthContext(ctx, nil, authcommon.Read, authcommon.DescribeInstancesCount)
	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return api.NewBatchQueryResponseWithMsg(authcommon.ConvertToErrCode(err), err.Error())
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetInstancesCount(ctx)
}

// GetInstanceLabels 获取某个服务下的实例标签集合
func (svr *Server) GetInstanceLabels(ctx context.Context,
	query map[string]string) *apiservice.Response {

	var (
		serviceId string
		namespace = service.DefaultNamespace
	)

	if val, ok := query["namespace"]; ok {
		namespace = val
	}

	if svcName, ok := query["service"]; ok {
		if svc := svr.Cache().Service().GetServiceByName(svcName, namespace); svc != nil {
			serviceId = svc.ID
		}
	}

	if id, ok := query["service_id"]; ok {
		serviceId = id
	}

	// TODO 如果在鉴权的时候发现资源不存在，怎么处理？
	svc := svr.Cache().Service().GetServiceByID(serviceId)
	if svc == nil {
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}

	authCtx := svr.collectServiceAuthContext(ctx, []*apiservice.Service{
		svc.ToSpec(),
	}, authcommon.Read, authcommon.DescribeInstanceLabels)
	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return api.NewResponseWithMsg(authcommon.ConvertToErrCode(err), err.Error())
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)
	return svr.nextSvr.GetInstanceLabels(ctx, query)
}
