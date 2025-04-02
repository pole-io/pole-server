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

package auth

import (
	"context"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	"go.uber.org/zap"

	"github.com/pole-io/pole-server/apis/access_control/auth"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	cacheapi "github.com/pole-io/pole-server/apis/cache"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/namespace"
)

var _ namespace.NamespaceOperateServer = (*Server)(nil)

// Server 带有鉴权能力的 NamespaceOperateServer
// 该层会对请求参数做一些调整，根据具体的请求发起人，设置为数据对应的 owner，不可为为别人进行创建资源
type Server struct {
	nextSvr   namespace.NamespaceOperateServer
	userSvr   auth.UserServer
	policySvr auth.StrategyServer
	cacheSvr  cacheapi.CacheManager
}

func NewServer(nextSvr namespace.NamespaceOperateServer, userSvr auth.UserServer,
	policySvr auth.StrategyServer, cacheSvr cacheapi.CacheManager) namespace.NamespaceOperateServer {
	proxy := &Server{
		nextSvr:   nextSvr,
		userSvr:   userSvr,
		policySvr: policySvr,
		cacheSvr:  cacheSvr,
	}
	return proxy
}

// CreateNamespaceIfAbsent Create a single name space
func (svr *Server) CreateNamespaceIfAbsent(ctx context.Context,
	req *apimodel.Namespace) (string, *apiservice.Response) {
	n, rsp := svr.nextSvr.CreateNamespaceIfAbsent(ctx, req)
	if api.IsSuccess(rsp) {
		_ = svr.afterNamespaceResource(ctx, rsp.Namespace, false)
	}
	return n, rsp
}

// CreateNamespace 创建命名空间，只需要要后置鉴权，将数据添加到资源策略中
func (svr *Server) CreateNamespace(ctx context.Context, req *apimodel.Namespace) *apiservice.Response {
	authCtx := svr.collectNamespaceAuthContext(
		ctx, []*apimodel.Namespace{req}, authtypes.Create, authtypes.CreateNamespace)
	// 验证 token 信息
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	// 填充 ownerId 信息数据
	if ownerId := utils.ParseOwnerID(ctx); len(ownerId) > 0 {
		req.Owners = utils.NewStringValue(ownerId)
	}

	resp := svr.nextSvr.CreateNamespace(ctx, req)
	_ = svr.afterNamespaceResource(ctx, resp.Namespace, false)
	return resp
}

// CreateNamespaces 创建命名空间，只需要要后置鉴权，将数据添加到资源策略中
func (svr *Server) CreateNamespaces(
	ctx context.Context, reqs []*apimodel.Namespace) *apiservice.BatchWriteResponse {
	authCtx := svr.collectNamespaceAuthContext(ctx, reqs, authtypes.Create, authtypes.CreateNamespaces)

	// 验证 token 信息
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	// 填充 ownerId 信息数据
	ownerId := utils.ParseOwnerID(ctx)
	if len(ownerId) > 0 {
		for index := range reqs {
			req := reqs[index]
			req.Owners = utils.NewStringValue(ownerId)
		}
	}
	resp := svr.nextSvr.CreateNamespaces(ctx, reqs)

	for i := range resp.Responses {
		item := resp.Responses[i].Namespace
		_ = svr.afterNamespaceResource(ctx, item, false)
	}
	return resp
}

// DeleteNamespaces 删除命名空间，需要先走权限检查
func (svr *Server) DeleteNamespaces(
	ctx context.Context, reqs []*apimodel.Namespace) *apiservice.BatchWriteResponse {
	authCtx := svr.collectNamespaceAuthContext(ctx, reqs, authtypes.Delete, authtypes.DeleteNamespaces)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	resp := svr.nextSvr.DeleteNamespaces(ctx, reqs)

	for i := range resp.Responses {
		item := resp.Responses[i].Namespace
		_ = svr.afterNamespaceResource(ctx, item, true)
	}
	return resp
}

// UpdateNamespaces 更新命名空间，需要先走权限检查
func (svr *Server) UpdateNamespaces(
	ctx context.Context, req []*apimodel.Namespace) *apiservice.BatchWriteResponse {
	authCtx := svr.collectNamespaceAuthContext(ctx, req, authtypes.Modify, authtypes.UpdateNamespaces)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	resp := svr.nextSvr.UpdateNamespaces(ctx, req)
	for i := range resp.Responses {
		item := resp.Responses[i].Namespace
		_ = svr.afterNamespaceResource(ctx, item, false)
	}
	return resp
}

// UpdateNamespaceToken 更新命名空间的token信息，需要先走权限检查
func (svr *Server) UpdateNamespaceToken(ctx context.Context, req *apimodel.Namespace) *apiservice.Response {
	authCtx := svr.collectNamespaceAuthContext(
		ctx, []*apimodel.Namespace{req}, authtypes.Modify, authtypes.UpdateNamespaceToken)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	return svr.nextSvr.UpdateNamespaceToken(ctx, req)
}

// GetNamespaces 获取命名空间列表信息，暂时不走权限检查
func (svr *Server) GetNamespaces(
	ctx context.Context, query map[string][]string) *apiservice.BatchQueryResponse {
	authCtx := svr.collectNamespaceAuthContext(ctx, nil, authtypes.Read, authtypes.DescribeNamespaces)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendNamespacePredicate(ctx, func(ctx context.Context, n *types.Namespace) bool {
		return svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     apisecurity.ResourceType_Namespaces,
			ID:       n.Name,
			Metadata: n.Metadata,
		})
	})

	authCtx.SetRequestContext(ctx)
	resp := svr.nextSvr.GetNamespaces(ctx, query)
	for i := range resp.Namespaces {
		item := resp.Namespaces[i]
		authCtx.SetAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_Namespaces: {
				{
					Type:     apisecurity.ResourceType_Namespaces,
					ID:       item.GetId().GetValue(),
					Metadata: item.GetMetadata(),
				},
			},
		})

		// 检查 write 操作权限
		authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.UpdateNamespaces})
		// 如果检查不通过，设置 editable 为 false
		if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Editable = utils.NewBoolValue(false)
		}

		// 检查 delete 操作权限
		authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.DeleteNamespaces})
		// 如果检查不通过，设置 editable 为 false
		if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
			item.Deleteable = utils.NewBoolValue(false)
		}
	}
	return resp
}

// GetNamespaceToken 获取命名空间的token信息，暂时不走权限检查
func (svr *Server) GetNamespaceToken(ctx context.Context, req *apimodel.Namespace) *apiservice.Response {
	authCtx := svr.collectNamespaceAuthContext(
		ctx, []*apimodel.Namespace{req}, authtypes.Read, authtypes.DescribeNamespaceToken)
	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, utils.ContextAuthContextKey, authCtx)
	return svr.nextSvr.GetNamespaceToken(ctx, req)
}

// collectNamespaceAuthContext 对于命名空间的处理，收集所有的与鉴权的相关信息
func (svr *Server) collectNamespaceAuthContext(ctx context.Context, req []*apimodel.Namespace,
	resourceOp authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {
	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(resourceOp),
		authtypes.WithModule(authtypes.CoreModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(svr.queryNamespaceResource(req)),
	)
}

// queryNamespaceResource 根据所给的 namespace 信息，收集对应的 ResourceEntry 列表
func (svr *Server) queryNamespaceResource(
	req []*apimodel.Namespace) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	if len(req) == 0 {
		return map[apisecurity.ResourceType][]authtypes.ResourceEntry{}
	}

	names := utils.NewSet[string]()
	for index := range req {
		names.Add(req[index].Name.GetValue())
	}
	param := names.ToSlice()
	nsArr := svr.cacheSvr.Namespace().GetNamespacesByName(param)

	temp := make([]authtypes.ResourceEntry, 0, len(nsArr))

	for index := range nsArr {
		ns := nsArr[index]
		temp = append(temp, authtypes.ResourceEntry{
			Type:  apisecurity.ResourceType_Namespaces,
			ID:    ns.Name,
			Owner: ns.Owner,
		})
	}

	ret := map[apisecurity.ResourceType][]authtypes.ResourceEntry{
		apisecurity.ResourceType_Namespaces: temp,
	}
	authLog.Debug("[Auth][Server] collect namespace access res", zap.Any("res", ret))
	return ret
}
