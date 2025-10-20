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
	"strconv"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
)

var (
	_allowLaneGroupFilters = map[string]struct{}{
		"id":          {},
		"name":        {},
		"offset":      {},
		"brief":       {},
		"limit":       {},
		"order_type":  {},
		"order_field": {},
	}
)

// CreateLaneGroups 批量创建泳道组
func (svr *Server) CreateLaneGroups(ctx context.Context, reqs []*apitraffic.LaneGroup) *apimodel.BatchWriteResponse {
	if err := checkBatchLaneGroupRules(reqs); err != nil {
		return err
	}
	batchRsp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range reqs {
		rsp := checkLaneGroupParam(reqs[i], false)
		api.Collect(batchRsp, rsp)
	}

	if !api.IsSuccess(batchRsp) {
		return batchRsp
	}
	return svr.nextSvr.CreateLaneGroups(ctx, reqs)
}

// UpdateLaneGroups 批量更新泳道组
func (svr *Server) UpdateLaneGroups(ctx context.Context, reqs []*apitraffic.LaneGroup) *apimodel.BatchWriteResponse {
	if err := checkBatchLaneGroupRules(reqs); err != nil {
		return err
	}
	batchRsp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range reqs {
		rsp := checkLaneGroupParam(reqs[i], true)
		api.Collect(batchRsp, rsp)
	}

	if !api.IsSuccess(batchRsp) {
		return batchRsp
	}
	return svr.nextSvr.UpdateLaneGroups(ctx, reqs)
}

// DeleteLaneGroups 批量删除泳道组
func (svr *Server) DeleteLaneGroups(ctx context.Context, reqs []*apitraffic.LaneGroup) *apimodel.BatchWriteResponse {
	if err := checkBatchLaneGroupRules(reqs); err != nil {
		return err
	}
	return svr.nextSvr.DeleteLaneGroups(ctx, reqs)
}

func (svr *Server) GetOneLaneGroup(ctx context.Context, req *apitraffic.LaneGroup) *apimodel.Response {
	return svr.nextSvr.GetOneLaneGroup(ctx, req)
}

// GetLaneGroups 查询泳道组列表
func (svr *Server) GetLaneGroups(ctx context.Context, filter map[string]string) *apimodel.BatchQueryResponse {
	offset, limit, err := valid.ParseOffsetAndLimit(filter)
	if err != nil {
		return api.NewBatchQueryResponseWithMsg(apimodel.Code_BadRequest, err.Error())
	}

	for k := range filter {
		if _, ok := _allowLaneGroupFilters[k]; !ok {
			log.Error("[Server][LaneGroup][Query] not allowed", zap.String("attribute", k), utils.RequestID(ctx))
			return api.NewBatchQueryResponseWithMsg(apimodel.Code_InvalidParameter, k+" is not allowed")
		}
		if filter[k] == "" {
			delete(filter, k)
		}
	}

	if _, ok := filter["order_field"]; !ok {
		filter["order_field"] = "mtime"
	}
	if _, ok := filter["order_type"]; !ok {
		filter["order_type"] = "desc"
	}

	filter["offset"] = strconv.FormatUint(uint64(offset), 10)
	filter["limit"] = strconv.FormatUint(uint64(limit), 10)

	return svr.nextSvr.GetLaneGroups(ctx, filter)
}

// CreateLaneRules 批量创建泳道规则
func (svr *Server) CreateLaneRules(ctx context.Context, req []*apitraffic.LaneRule) *apimodel.BatchWriteResponse {
	if len(req) == 0 {
		return api.NewBatchWriteResponse(apimodel.Code_EmptyRequest)
	}

	if len(req) > valid.MaxBatchSize {
		return api.NewBatchWriteResponse(apimodel.Code_BatchSizeOverLimit)
	}

	batchRsp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range req {
		if len(req[i].GetName()) >= valid.MaxRuleName {
			api.Collect(batchRsp, api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "lane_rule name size must be <= 64"))
			continue
		}
		if err := valid.CheckResourceName(wrapperspb.String(req[i].GetName())); err != nil {
			api.Collect(batchRsp, api.NewResponseWithMsg(apimodel.Code_InvalidParameter, err.Error()))
			continue
		}
		if len(req[i].GetGroupName()) == 0 {
			api.Collect(batchRsp, api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "lane_rule group name is required"))
			continue
		}
	}

	if !api.IsSuccess(batchRsp) {
		return batchRsp
	}
	return svr.nextSvr.CreateLaneRules(ctx, req)
}

// UpdateLaneRules 批量更新泳道规则
func (svr *Server) UpdateLaneRules(ctx context.Context, req []*apitraffic.LaneRule) *apimodel.BatchWriteResponse {
	if len(req) == 0 {
		return api.NewBatchWriteResponse(apimodel.Code_EmptyRequest)
	}

	if len(req) > valid.MaxBatchSize {
		return api.NewBatchWriteResponse(apimodel.Code_BatchSizeOverLimit)
	}
	batchRsp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range req {
		if len(req[i].GetGroupName()) == 0 {
			api.Collect(batchRsp, api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "lane_rule group name is required"))
			continue
		}
	}

	if !api.IsSuccess(batchRsp) {
		return batchRsp
	}

	return svr.nextSvr.UpdateLaneRules(ctx, req)
}

// DeleteLaneRules 批量删除泳道规则
func (svr *Server) DeleteLaneRules(ctx context.Context, req []*apitraffic.LaneRule) *apimodel.BatchWriteResponse {
	if len(req) == 0 {
		return api.NewBatchWriteResponse(apimodel.Code_EmptyRequest)
	}
	if len(req) > valid.MaxBatchSize {
		return api.NewBatchWriteResponse(apimodel.Code_BatchSizeOverLimit)
	}
	batchRsp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range req {
		if len(req[i].GetGroupName()) == 0 {
			api.Collect(batchRsp, api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "lane_rule group name is required"))
			continue
		}
	}

	if !api.IsSuccess(batchRsp) {
		return batchRsp
	}
	return svr.nextSvr.DeleteLaneRules(ctx, req)
}

func checkBatchLaneGroupRules(req []*apitraffic.LaneGroup) *apimodel.BatchWriteResponse {
	if len(req) == 0 {
		return api.NewBatchWriteResponse(apimodel.Code_EmptyRequest)
	}

	if len(req) > valid.MaxBatchSize {
		return api.NewBatchWriteResponse(apimodel.Code_BatchSizeOverLimit)
	}
	return nil
}

func checkLaneGroupParam(req *apitraffic.LaneGroup, update bool) *apimodel.Response {
	if len(req.GetName()) >= valid.MaxRuleName {
		return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "lane_group name size must be <= 64")
	}
	if err := valid.CheckResourceName(wrapperspb.String(req.GetName())); err != nil {
		return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, err.Error())
	}
	if len(req.Rules) > valid.MaxBatchSize {
		return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "lane_rule size must be <= 100")
	}
	for i := range req.Rules {
		rule := req.Rules[i]
		if err := valid.CheckResourceName(wrapperspb.String(rule.GetName())); err != nil {
			return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, err.Error())
		}
		if len(rule.GetName()) >= valid.MaxRuleName {
			return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "lane_rule name size must be <= 64")
		}
	}

	if update {
		if req.GetId() == "" {
			return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "lane_group id is empty")
		}
	}
	return nil
}
