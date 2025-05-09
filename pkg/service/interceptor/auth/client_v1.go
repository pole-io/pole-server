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

	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/polarismesh/specification/source/go/api/v1/security"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// RegisterInstance create one instance
func (svr *Server) RegisterInstance(ctx context.Context, req *apiservice.Instance) *apiservice.Response {
	authCtx := svr.collectClientInstanceAuthContext(
		ctx, []*apiservice.Instance{req}, authtypes.Create, authtypes.RegisterInstance)

	_, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx)
	if err != nil {
		resp := api.NewResponseWithMsg(authtypes.ConvertToErrCode(err), err.Error())
		return resp
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.RegisterInstance(ctx, req)
}

// DeregisterInstance delete onr instance
func (svr *Server) DeregisterInstance(ctx context.Context, req *apiservice.Instance) *apiservice.Response {
	authCtx := svr.collectClientInstanceAuthContext(
		ctx, []*apiservice.Instance{req}, authtypes.Create, authtypes.DeregisterInstance)

	if _, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.DeregisterInstance(ctx, req)
}

// ReportClient is the interface for reporting client authability
func (svr *Server) ReportClient(ctx context.Context, req *apiservice.Client) *apiservice.Response {
	return svr.nextSvr.ReportClient(ctx, req)
}

// ReportServiceContract .
func (svr *Server) ReportServiceContract(ctx context.Context, req *apiservice.ServiceContract) *apiservice.Response {
	authCtx := svr.collectServiceAuthContext(
		ctx, []*apiservice.Service{{
			Name:      wrapperspb.String(req.GetService()),
			Namespace: wrapperspb.String(req.GetNamespace()),
		}}, authtypes.Create, authtypes.ReportServiceContract)

	if _, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return svr.nextSvr.ReportServiceContract(ctx, req)
}

// GetPrometheusTargets Used for client acquisition service information
func (svr *Server) GetPrometheusTargets(ctx context.Context,
	query map[string]string) *types.PrometheusDiscoveryResponse {

	return svr.nextSvr.GetPrometheusTargets(ctx, query)
}

// GetServiceWithCache is the interface for getting service with cache
func (svr *Server) GetServiceWithCache(
	ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {

	authCtx := svr.collectServiceAuthContext(
		ctx, []*apiservice.Service{req}, authtypes.Read, authtypes.DiscoverServices)
	if _, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx); err != nil {
		return api.NewDiscoverResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendServicePredicate(ctx, func(ctx context.Context, cbr *svctypes.Service) bool {
		return svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_Services,
			ID:       cbr.ID,
			Metadata: cbr.Meta,
		})
	})
	authCtx.SetRequestContext(ctx)

	return svr.nextSvr.GetServiceWithCache(ctx, req)
}

// ServiceInstancesCache is the interface for getting service instances cache
func (svr *Server) ServiceInstancesCache(
	ctx context.Context, filter *apiservice.DiscoverFilter, req *apiservice.Service) *apiservice.DiscoverResponse {

	authCtx := svr.collectServiceAuthContext(
		ctx, []*apiservice.Service{req}, authtypes.Read, authtypes.DiscoverInstances)
	if _, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx); err != nil {
		return api.NewDiscoverResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.ServiceInstancesCache(ctx, filter, req)
}

// UpdateInstance update single instance
func (svr *Server) UpdateInstance(ctx context.Context, req *apiservice.Instance) *apiservice.Response {
	authCtx := svr.collectClientInstanceAuthContext(
		ctx, []*apiservice.Instance{req}, authtypes.Modify, authtypes.UpdateInstance)

	if _, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.UpdateInstance(ctx, req)
}

// GetServiceContractWithCache User Client Get ServiceContract Rule Information
func (svr *Server) GetServiceContractWithCache(ctx context.Context,
	req *apiservice.ServiceContract) *apiservice.Response {
	authCtx := svr.collectServiceAuthContext(ctx, []*apiservice.Service{{
		Namespace: wrapperspb.String(req.Namespace),
		Name:      wrapperspb.String(req.Service),
	}}, authtypes.Read, authtypes.DiscoverServiceContract)

	if _, err := svr.policySvr.GetAuthChecker().CheckClientPermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetServiceContractWithCache(ctx, req)
}
