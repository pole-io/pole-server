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

package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/wrappers"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"
	"go.uber.org/zap"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	apiv1 "github.com/pole-io/pole-server/pkg/common/api/v1"
	commonstore "github.com/pole-io/pole-server/pkg/common/store"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// CreateRoutingConfigs Create a routing configuration
func (s *Server) CreateRoutingConfigs(
	ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse {
	resp := apiv1.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, entry := range req {
		apiv1.Collect(resp, s.createRoutingConfig(ctx, entry))
	}

	return apiv1.FormatBatchWriteResponse(resp)
}

// createRoutingConfig Create a routing configuration
func (s *Server) createRoutingConfig(ctx context.Context, req *apitraffic.RouteRule) *apiservice.Response {
	conf, err := Api2RoutingConfig(req)
	if err != nil {
		log.Error("[Routing][] parse routing config  from request for create",
			utils.RequestID(ctx), zap.Error(err))
		return apiv1.NewResponse(apimodel.Code_ExecuteException)
	}

	if err := s.storage.CreateRoutingConfig(conf); err != nil {
		log.Error("[Routing][] create routing config  store layer",
			utils.RequestID(ctx), zap.Error(err))
		return apiv1.NewResponse(commonstore.StoreCode2APICode(err))
	}

	s.RecordHistory(ctx, routeRuleRecordEntry(ctx, req, conf, types.OCreate))
	req.Id = conf.ID
	return apiv1.NewRouterResponse(apimodel.Code_ExecuteSuccess, req)
}

// DeleteRoutingConfigs Batch delete routing configuration
func (s *Server) DeleteRoutingConfigs(
	ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse {
	out := apiv1.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, entry := range req {
		resp := s.deleteRoutingConfig(ctx, entry)
		apiv1.Collect(out, resp)
	}

	return apiv1.FormatBatchWriteResponse(out)
}

// DeleteRoutingConfig Delete a routing configuration
func (s *Server) deleteRoutingConfig(ctx context.Context, req *apitraffic.RouteRule) *apiservice.Response {
	if err := s.storage.DeleteRoutingConfig(req.Id); err != nil {
		log.Error("[Routing][] delete routing config  store layer",
			utils.RequestID(ctx), zap.Error(err))
		return apiv1.NewResponse(commonstore.StoreCode2APICode(err))
	}

	s.RecordHistory(ctx, routeRuleRecordEntry(ctx, req, &rules.RouterConfig{
		ID:   req.GetId(),
		Name: req.GetName(),
	}, types.ODelete))
	return apiv1.NewRouterResponse(apimodel.Code_ExecuteSuccess, req)
}

// UpdateRoutingConfigs Batch update routing configuration
func (s *Server) UpdateRoutingConfigs(
	ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse {
	out := apiv1.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, entry := range req {
		resp := s.updateRoutingConfig(ctx, entry)
		apiv1.Collect(out, resp)
	}

	return apiv1.FormatBatchWriteResponse(out)
}

// updateRoutingConfig Update a single routing configuration
func (s *Server) updateRoutingConfig(ctx context.Context, req *apitraffic.RouteRule) *apiservice.Response {
	// Check whether the routing configuration exists
	conf, err := s.storage.GetRoutingConfigWithID(req.Id)
	if err != nil {
		log.Error("[Routing][] get routing config  store layer",
			utils.RequestID(ctx), zap.Error(err))
		return apiv1.NewResponse(commonstore.StoreCode2APICode(err))
	}
	if conf == nil {
		return apiv1.NewResponse(apimodel.Code_NotFoundRouting)
	}

	reqModel, err := Api2RoutingConfig(req)
	reqModel.Revision = utils.NewRevision()
	if err != nil {
		log.Error("[Routing][] parse routing config  from request for update",
			utils.RequestID(ctx), zap.Error(err))
		return apiv1.NewResponse(apimodel.Code_ExecuteException)
	}

	if err := s.storage.UpdateRoutingConfig(reqModel); err != nil {
		log.Error("[Routing][] update routing config  store layer",
			utils.RequestID(ctx), zap.Error(err))
		return apiv1.NewResponse(commonstore.StoreCode2APICode(err))
	}

	s.RecordHistory(ctx, routeRuleRecordEntry(ctx, req, reqModel, types.OUpdate))
	return apiv1.NewResponse(apimodel.Code_ExecuteSuccess)
}

// QueryRoutingConfigs The interface of the query configuration to the OSS
func (s *Server) QueryRoutingConfigs(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse {
	args, presp := parseRoutingArgs(query, ctx)
	if presp != nil {
		return apiv1.NewBatchQueryResponse(apimodel.Code(presp.GetCode().GetValue()))
	}

	total, ret, err := s.Cache().RoutingConfig().QueryRoutingConfigs(ctx, args)
	if err != nil {
		log.Error("[Routing][] query routing list from cache", utils.RequestID(ctx), zap.Error(err))
		return apiv1.NewBatchQueryResponse(apimodel.Code_ExecuteException)
	}

	routers, err := marshalRoutingtoAnySlice(ret)
	if err != nil {
		log.Error("[Routing][] marshal routing list to anypb.Any list",
			utils.RequestID(ctx), zap.Error(err))
		return apiv1.NewBatchQueryResponse(apimodel.Code_ExecuteException)
	}

	resp := apiv1.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = &wrappers.UInt32Value{Value: total}
	resp.Size = &wrappers.UInt32Value{Value: uint32(len(ret))}
	resp.Data = routers
	return resp
}

// GetAllRouterRules Query all router_rule rules
func (s *Server) GetAllRouterRules(ctx context.Context) *apiservice.BatchQueryResponse {
	return nil
}

// EnableRoutings batch enable routing rules
func (s *Server) EnableRoutings(ctx context.Context, req []*apitraffic.RouteRule) *apiservice.BatchWriteResponse {
	out := apiv1.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, entry := range req {
		resp := s.enableRoutings(ctx, entry)
		apiv1.Collect(out, resp)
	}

	return apiv1.FormatBatchWriteResponse(out)
}

func (s *Server) enableRoutings(ctx context.Context, req *apitraffic.RouteRule) *apiservice.Response {
	conf, err := s.storage.GetRoutingConfigWithID(req.Id)
	if err != nil {
		log.Error("[Routing][] get routing config  store layer",
			utils.RequestID(ctx), zap.Error(err))
		return apiv1.NewResponse(commonstore.StoreCode2APICode(err))
	}
	if conf == nil {
		return apiv1.NewResponse(apimodel.Code_NotFoundRouting)
	}

	conf.Enable = req.GetEnable()
	conf.Revision = utils.NewRevision()

	if err := s.storage.EnableRouting(conf); err != nil {
		log.Error("[Routing][] enable routing config  store layer",
			utils.RequestID(ctx), zap.Error(err))
		return apiv1.NewResponse(commonstore.StoreCode2APICode(err))
	}

	s.RecordHistory(ctx, routeRuleRecordEntry(ctx, req, conf, types.OUpdate))
	return apiv1.NewResponse(apimodel.Code_ExecuteSuccess)
}

// parseServiceArgs The query conditions of the analysis service
func parseRoutingArgs(filter map[string]string, ctx context.Context) (*cacheapi.RoutingArgs, *apiservice.Response) {
	offset, limit, _ := utils.ParseOffsetAndLimit(filter)
	res := &cacheapi.RoutingArgs{
		Filter:     filter,
		Name:       filter["name"],
		ID:         filter["id"],
		OrderField: filter["order_field"],
		OrderType:  filter["order_type"],
		Offset:     offset,
		Limit:      limit,
	}

	if _, ok := filter["service"]; ok {
		res.Namespace = filter["namespace"]
		res.Service = filter["service"]
	} else {
		res.SourceService = filter["source_service"]
		res.SourceNamespace = filter["source_namespace"]

		res.DestinationService = filter["destination_service"]
		res.DestinationNamespace = filter["destination_namespace"]
	}

	if enableStr, ok := filter["enable"]; ok {
		enable, err := strconv.ParseBool(enableStr)
		if err == nil {
			res.Enable = &enable
		} else {
			log.Error("[Service][Routing][Query] search with routing enable", zap.Error(err))
		}
	}
	log.Infof("[Service][Routing][Query] routing query args: %+v", res)
	return res, nil
}

// Api2RoutingConfig Convert the API parameter to internal data structure
func Api2RoutingConfig(req *apitraffic.RouteRule) (*rules.RouterConfig, error) {
	out := &rules.RouterConfig{
		Valid: true,
	}
	if req.Id == "" {
		req.Id = utils.NewUUID()
	}
	if req.Revision == "" {
		req.Revision = utils.NewUUID()
	}

	if err := out.ParseRouteRuleFromAPI(req); err != nil {
		return nil, err
	}
	return out, nil
}

// marshalRoutingtoAnySlice Converted to []*anypb.Any array
func marshalRoutingtoAnySlice(routings []*rules.ExtendRouterConfig) ([]*any.Any, error) {
	ret := make([]*any.Any, 0, len(routings))

	for i := range routings {
		entry, err := routings[i].ToApi()
		if err != nil {
			return nil, err
		}
		item, err := ptypes.MarshalAny(entry)
		if err != nil {
			return nil, err
		}

		ret = append(ret, item)
	}

	return ret, nil
}

// routeRuleRecordEntry Construction of RoutingConfig's record Entry
func routeRuleRecordEntry(ctx context.Context, req *apitraffic.RouteRule, md *rules.RouterConfig,
	opt types.OperationType) *types.RecordEntry {

	marshaler := jsonpb.Marshaler{}
	detail, _ := marshaler.MarshalToString(req)

	entry := &types.RecordEntry{
		ResourceType:  types.RRouting,
		ResourceName:  fmt.Sprintf("%s(%s)", md.Name, md.ID),
		Namespace:     req.GetNamespace(),
		OperationType: opt,
		Operator:      utils.ParseOperator(ctx),
		Detail:        detail,
		HappenTime:    time.Now(),
	}
	return entry
}
