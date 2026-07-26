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

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pole-io/pole-server/apis/pkg/types"
	authcommon "github.com/pole-io/pole-server/apis/pkg/types/auth"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// CreateServiceContracts .
func (svr *Server) CreateServiceContracts(ctx context.Context,
	req []*apiservice.ServiceContract) *apimodel.BatchWriteResponse {
	services := make([]*apiservice.Service, 0, len(req))
	for i := range req {
		services = append(services, &apiservice.Service{
			Namespace: req[i].Namespace,
			Name:      req[i].Service,
		})
	}

	authCtx := svr.collectServiceAuthContext(ctx, services, authcommon.Create, authcommon.CreateServiceContracts)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authcommon.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.CreateServiceContracts(ctx, req)
}

// GetServiceContracts .
func (svr *Server) GetServiceContracts(ctx context.Context,
	query map[string]string) *apimodel.BatchQueryResponse {
	services := contractQueryServices(query)
	authCtx := svr.collectServiceAuthContext(ctx, services, authcommon.Read, authcommon.DescribeServiceContracts)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authcommon.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.filterServiceContractsByPermission(svr.nextSvr.GetServiceContracts(ctx, query), authCtx)
}

// GetServiceContractVersions .
func (svr *Server) GetServiceContractVersions(ctx context.Context,
	filter map[string]string) *apimodel.BatchQueryResponse {

	authCtx := svr.collectServiceAuthContext(ctx, contractQueryServices(filter), authcommon.Read,
		authcommon.DescribeServiceContractVersions)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authcommon.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.filterServiceContractsByPermission(svr.nextSvr.GetServiceContractVersions(ctx, filter), authCtx)
}

func (svr *Server) filterServiceContractsByPermission(
	resp *apimodel.BatchQueryResponse, authCtx *authcommon.AcquireContext,
) *apimodel.BatchQueryResponse {
	if resp == nil || resp.Code != uint32(apimodel.Code_ExecuteSuccess) || len(resp.Data) == 0 {
		return resp
	}
	filtered := make([]*anypb.Any, 0, len(resp.Data))
	for _, data := range resp.Data {
		contract := &apiservice.ServiceContract{}
		if data == nil || anypb.UnmarshalTo(data, contract, proto.UnmarshalOptions{}) != nil {
			continue
		}

		allowed := false
		if svc := svr.Cache().Service().GetServiceByName(contract.GetService(), contract.GetNamespace()); svc != nil {
			allowed = svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authcommon.ResourceEntry{
				Type:     apisecurity.ResourceType_Services,
				ID:       svc.ID,
				Metadata: svc.Meta,
			})
		}
		if !allowed {
			if ns := svr.Cache().Namespace().GetNamespace(contract.GetNamespace()); ns != nil {
				allowed = svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authcommon.ResourceEntry{
					Type:     apisecurity.ResourceType_Namespaces,
					ID:       ns.Name,
					Metadata: ns.Metadata,
				})
			}
		}
		if allowed {
			filtered = append(filtered, data)
		}
	}
	resp.Data = filtered
	resp.Size = uint32(len(filtered))
	resp.Amount = uint32(len(filtered))
	return resp
}

func contractQueryServices(query map[string]string) []*apiservice.Service {
	if query["namespace"] == "" || query["service"] == "" {
		return nil
	}
	return []*apiservice.Service{{
		Namespace: query["namespace"],
		Name:      query["service"],
	}}
}

// DeleteServiceContracts .
func (svr *Server) DeleteServiceContracts(ctx context.Context,
	req []*apiservice.ServiceContract) *apimodel.BatchWriteResponse {
	services := make([]*apiservice.Service, 0, len(req))
	for i := range req {
		services = append(services, &apiservice.Service{
			Namespace: req[i].Namespace,
			Name:      req[i].Service,
		})
	}

	authCtx := svr.collectServiceAuthContext(ctx, services, authcommon.Delete, authcommon.DeleteServiceContracts)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authcommon.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.DeleteServiceContracts(ctx, req)
}

// CreateServiceContractInterfaces .
func (svr *Server) CreateServiceContractInterfaces(ctx context.Context, contract *apiservice.ServiceContract,
	source apiservice.InterfaceDescriptor_Source) *apimodel.Response {
	authCtx := svr.collectServiceAuthContext(ctx, []*apiservice.Service{
		{
			Namespace: contract.Namespace,
			Name:      contract.Service,
		},
	}, authcommon.Modify, authcommon.CreateServiceContractInterfaces)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authcommon.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.CreateServiceContractInterfaces(ctx, contract, source)
}

// AppendServiceContractInterfaces .
func (svr *Server) AppendServiceContractInterfaces(ctx context.Context,
	contract *apiservice.ServiceContract, source apiservice.InterfaceDescriptor_Source) *apimodel.Response {
	authCtx := svr.collectServiceAuthContext(ctx, []*apiservice.Service{
		{
			Namespace: contract.Namespace,
			Name:      contract.Service,
		},
	}, authcommon.Modify, authcommon.AppendServiceContractInterfaces)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authcommon.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.AppendServiceContractInterfaces(ctx, contract, source)
}

// DeleteServiceContractInterfaces .
func (svr *Server) DeleteServiceContractInterfaces(ctx context.Context,
	contract *apiservice.ServiceContract) *apimodel.Response {
	authCtx := svr.collectServiceAuthContext(ctx, []*apiservice.Service{
		{
			Namespace: contract.Namespace,
			Name:      contract.Service,
		},
	}, authcommon.Modify, authcommon.DeleteServiceContractInterfaces)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authcommon.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.DeleteServiceContractInterfaces(ctx, contract)
}
