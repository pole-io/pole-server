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
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// CreateServices 批量创建服务
func (svr *Server) CreateServices(
	ctx context.Context, reqs []*apiservice.Service) *apimodel.BatchWriteResponse {
	authCtx := svr.collectServiceAuthContext(ctx, reqs, authtypes.Create, authtypes.CreateServices)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	// 填充 ownerID 信息数据
	ownerID := utils.ParseOwnerID(ctx)
	if len(ownerID) > 0 {
		for index := range reqs {
			req := reqs[index]
			req.Owners = ownerID
		}
	}

	resp := svr.nextSvr.CreateServices(ctx, reqs)

	nRsp := api.NewBatchWriteResponse(apimodel.Code(resp.Code))
	for index := range resp.Responses {
		r := resp.Responses[index]
		var svc *apiservice.Service
		if r != nil && r.GetData() != nil {
			svc = &apiservice.Service{}
			if err := anypb.UnmarshalTo(r.GetData(), svc, proto.UnmarshalOptions{}); err != nil {
				svc = nil
			}
		}
		if err := svr.afterServiceResource(ctx, svc, false); err != nil {
			api.Collect(nRsp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(nRsp, r)
		}
	}
	return nRsp
}

// DeleteServices 批量删除服务
func (svr *Server) DeleteServices(
	ctx context.Context, reqs []*apiservice.Service) *apimodel.BatchWriteResponse {
	authCtx := svr.collectServiceAuthContext(ctx, reqs, authtypes.Delete, authtypes.DeleteServices)

	accessRes := authCtx.GetAccessResources()
	delete(accessRes, apisecurity.ResourceType_Namespaces)
	authCtx.SetAccessResources(accessRes)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	resp := svr.nextSvr.DeleteServices(ctx, reqs)

	nRsp := api.NewBatchWriteResponse(apimodel.Code(resp.Code))
	for index := range resp.Responses {
		r := resp.Responses[index]
		var svc *apiservice.Service
		if r != nil && r.GetData() != nil {
			svc = &apiservice.Service{}
			if err := anypb.UnmarshalTo(r.GetData(), svc, proto.UnmarshalOptions{}); err != nil {
				svc = nil
			}
		}
		if err := svr.afterServiceResource(ctx, svc, true); err != nil {
			api.Collect(nRsp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(nRsp, r)
		}
	}
	return nRsp
}

// UpdateServices 对于服务修改来说，只针对服务本身，而不需要检查命名空间
func (svr *Server) UpdateServices(
	ctx context.Context, reqs []*apiservice.Service) *apimodel.BatchWriteResponse {
	authCtx := svr.collectServiceAuthContext(ctx, reqs, authtypes.Modify, authtypes.UpdateServices)

	accessRes := authCtx.GetAccessResources()
	delete(accessRes, apisecurity.ResourceType_Namespaces)
	authCtx.SetAccessResources(accessRes)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	resp := svr.nextSvr.UpdateServices(ctx, reqs)

	nRsp := api.NewBatchWriteResponse(apimodel.Code(resp.Code))
	for index := range resp.Responses {
		r := resp.Responses[index]
		var svc *apiservice.Service
		if r != nil && r.GetData() != nil {
			svc = &apiservice.Service{}
			if err := anypb.UnmarshalTo(r.GetData(), svc, proto.UnmarshalOptions{}); err != nil {
				svc = nil
			}
		}
		if err := svr.afterServiceResource(ctx, svc, false); err != nil {
			api.Collect(nRsp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(nRsp, r)
		}
	}

	return nRsp
}

func (svr *Server) GetAllServices(ctx context.Context,
	query map[string]string) *apimodel.BatchQueryResponse {
	authCtx := svr.collectServiceAuthContext(ctx, nil, authtypes.Read, authtypes.DescribeAllServices)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendServicePredicate(ctx, func(ctx context.Context, cbr *svctypes.Service) bool {
		ok := svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_Services,
			ID:       cbr.ID,
			Metadata: cbr.Meta,
		})
		if ok {
			return true
		}
		saveNs := svr.Cache().Namespace().GetNamespace(cbr.Namespace)
		if saveNs == nil {
			return false
		}
		// 检查下是否可以访问对应的 namespace
		return svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_Namespaces,
			ID:       saveNs.Name,
			Metadata: saveNs.Metadata,
		})
	})
	authCtx.SetRequestContext(ctx)

	return svr.nextSvr.GetAllServices(ctx, query)
}

// GetServices 批量获取服务
func (svr *Server) GetServices(
	ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	authCtx := svr.collectServiceAuthContext(ctx, nil, authtypes.Read, authtypes.DescribeServices)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	// 注入查询条件拦截器
	ctx = cacheapi.AppendServicePredicate(ctx, func(ctx context.Context, cbr *svctypes.Service) bool {
		ok := svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_Services,
			ID:       cbr.ID,
			Metadata: cbr.Meta,
		})
		if ok {
			return true
		}
		saveNs := svr.Cache().Namespace().GetNamespace(cbr.Namespace)
		if saveNs == nil {
			return false
		}
		// 检查下是否可以访问对应的 namespace
		return svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_Namespaces,
			ID:       saveNs.Name,
			Metadata: saveNs.Metadata,
		})
	})
	authCtx.SetRequestContext(ctx)

	resp := svr.nextSvr.GetServices(ctx, query)
	return svr.filterServicesByPermission(ctx, resp, authCtx)
}

// filterServicesByPermission 基于 BatchQueryResponse.Data 做服务级权限过滤，仅保留有权限的服务
func (svr *Server) filterServicesByPermission(
	ctx context.Context, resp *apimodel.BatchQueryResponse, authCtx *authtypes.AcquireContext) *apimodel.BatchQueryResponse {
	if resp == nil || resp.Code != uint32(apimodel.Code_ExecuteSuccess) || len(resp.Data) == 0 {
		return resp
	}
	filtered := make([]*anypb.Any, 0, len(resp.Data))
	for _, anyData := range resp.Data {
		if anyData == nil {
			continue
		}
		svc := &apiservice.Service{}
		if err := anypb.UnmarshalTo(anyData, svc, proto.UnmarshalOptions{}); err != nil {
			continue
		}
		ok := svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_Services,
			ID:       svc.GetId(),
			Metadata: svc.GetMetadata(),
		})
		if !ok {
			saveNs := svr.Cache().Namespace().GetNamespace(svc.GetNamespace())
			if saveNs == nil {
				continue
			}
			ok = svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
				Type:     security.ResourceType_Namespaces,
				ID:       saveNs.Name,
				Metadata: saveNs.Metadata,
			})
		}
		if ok {
			filtered = append(filtered, anyData)
		}
	}
	resp.Data = filtered
	resp.Size = uint32(len(filtered))
	return resp
}

// GetServicesCount 批量获取服务数量
func (svr *Server) GetServicesCount(ctx context.Context) *apimodel.BatchQueryResponse {
	authCtx := svr.collectServiceAuthContext(ctx, nil, authtypes.Read, authtypes.DescribeServicesCount)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.GetServicesCount(ctx)
}

// GetServiceOwner 获取服务的 owner
func (svr *Server) GetServiceSubscribers(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	authCtx := svr.collectServiceAuthContext(ctx, []*apiservice.Service{}, authtypes.Read, authtypes.DescribeServiceOwner)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetServiceSubscribers(ctx, query)
}
