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

package paramcheck

import (
	"context"

	"google.golang.org/protobuf/types/known/wrapperspb"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/service"
)

var (
	clientFilterAttributes = map[string]struct{}{
		"type":    {},
		"host":    {},
		"limit":   {},
		"offset":  {},
		"version": {},
	}
)

// GetPrometheusTargets implements service.DiscoverServer.
func (svr *Server) GetPrometheusTargets(ctx context.Context,
	query map[string]string) *types.PrometheusDiscoveryResponse {
	return svr.nextSvr.GetPrometheusTargets(ctx, query)
}

// RegisterInstance create one instance by client
func (s *Server) RegisterInstance(ctx context.Context, req *apiservice.Instance) *apiservice.Response {
	// 参数检查
	if err := checkMetadata(req.GetMetadata()); err != nil {
		return api.NewInstanceResponse(apimodel.Code_InvalidMetadata, req)
	}
	instanceID, rsp := checkCreateInstance(req)
	if rsp != nil {
		return rsp
	}
	req.Id = protobuf.NewStringValue(instanceID)
	return s.nextSvr.RegisterInstance(ctx, req)
}

// DeregisterInstance delete onr instance by client
func (s *Server) DeregisterInstance(ctx context.Context, req *apiservice.Instance) *apiservice.Response {
	instanceID, resp := checkReviseInstance(req)
	if resp != nil {
		return resp
	}
	req.Id = wrapperspb.String(instanceID)
	return s.nextSvr.DeregisterInstance(ctx, req)
}

// ReportClient Client gets geographic location information
func (s *Server) ReportClient(ctx context.Context, req *apiservice.Client) *apiservice.Response {
	if s.nextSvr.Cache() == nil {
		return api.NewResponse(apimodel.Code_ClientAPINotOpen)
	}
	return s.nextSvr.ReportClient(ctx, req)
}

// GetServiceWithCache Used for client acquisition service information
func (s *Server) GetServiceWithCache(ctx context.Context, req *apiservice.Service) *apiservice.DiscoverResponse {
	if s.nextSvr.Cache() == nil {
		return api.NewDiscoverServiceResponse(apimodel.Code_ClientAPINotOpen, req)
	}
	if req == nil {
		return api.NewDiscoverServiceResponse(apimodel.Code_EmptyRequest, req)
	}
	return s.nextSvr.GetServiceWithCache(ctx, req)
}

// ServiceInstancesCache Used for client acquisition service instance information
func (s *Server) ServiceInstancesCache(ctx context.Context, filter *apiservice.DiscoverFilter,
	req *apiservice.Service) *apiservice.DiscoverResponse {
	resp := service.CreateCommonDiscoverResponse(req, apiservice.DiscoverResponse_INSTANCE)

	namespaceName := req.GetNamespace().GetValue()

	// 消费服务为了兼容，可以不带namespace，server端使用默认的namespace
	if namespaceName == "" {
		namespaceName = service.DefaultNamespace
		req.Namespace = protobuf.NewStringValue(namespaceName)
	}
	if !s.commonCheckDiscoverRequest(req, resp) {
		return resp
	}
	return s.nextSvr.ServiceInstancesCache(ctx, filter, req)
}

// GetServiceContractWithCache User Client Get ServiceContract Rule Information
func (s *Server) GetServiceContractWithCache(ctx context.Context, req *apiservice.ServiceContract) *apiservice.Response {
	resp := api.NewResponse(apimodel.Code_ExecuteSuccess)
	if !s.serviceContractCheckDiscoverRequest(req, resp) {
		return resp
	}

	return s.nextSvr.GetServiceContractWithCache(ctx, req)
}

// UpdateInstance update one instance by client
func (s *Server) UpdateInstance(ctx context.Context, req *apiservice.Instance) *apiservice.Response {
	// 参数检查
	if err := checkMetadata(req.GetMetadata()); err != nil {
		return api.NewInstanceResponse(apimodel.Code_InvalidMetadata, req)
	}
	instanceID, rsp := checkReviseInstance(req)
	if rsp != nil {
		return rsp
	}
	req.Id = protobuf.NewStringValue(instanceID)
	return s.nextSvr.UpdateInstance(ctx, req)
}

// ReportServiceContract client report service_contract
func (s *Server) ReportServiceContract(ctx context.Context, req *apiservice.ServiceContract) *apiservice.Response {
	return s.nextSvr.ReportServiceContract(ctx, req)
}

func (s *Server) commonCheckDiscoverRequest(req *apiservice.Service, resp *apiservice.DiscoverResponse) bool {
	if s.nextSvr.Cache() == nil {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_ClientAPINotOpen))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = req
		return false
	}
	if req == nil {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_EmptyRequest))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = req
		return false
	}

	if req.GetName().GetValue() == "" {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_InvalidServiceName))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = req
		return false
	}
	if req.GetNamespace().GetValue() == "" {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_InvalidNamespaceName))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = req
		return false
	}

	return true
}

func (s *Server) serviceContractCheckDiscoverRequest(req *apiservice.ServiceContract, resp *apiservice.Response) bool {
	svc := &apiservice.Service{
		Name:      wrapperspb.String(req.GetService()),
		Namespace: wrapperspb.String(req.GetNamespace()),
	}

	if s.nextSvr.Cache() == nil {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_ClientAPINotOpen))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = svc
		resp.ServiceContract = req
		return false
	}
	if req == nil {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_EmptyRequest))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = svc
		return false
	}

	if req.GetName() == "" {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_InvalidParameter))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = svc
		resp.ServiceContract = req
		return false
	}
	if req.GetNamespace() == "" {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_InvalidNamespaceName))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = svc
		resp.ServiceContract = req
		return false
	}
	if req.GetProtocol() == "" {
		resp.Code = protobuf.NewUInt32Value(uint32(apimodel.Code_InvalidParameter))
		resp.Info = protobuf.NewStringValue(api.Code2Info(resp.GetCode().GetValue()))
		resp.Service = svc
		resp.ServiceContract = req
		return false
	}
	return true
}
