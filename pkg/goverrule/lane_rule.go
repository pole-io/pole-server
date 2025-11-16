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

package goverrule

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
)

// CreateLaneGroups 批量创建泳道组
func (s *Server) CreateLaneGroups(ctx context.Context, req []*apitraffic.LaneGroup) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range req {
		resp := s.CreateLaneGroup(ctx, req[i])
		api.Collect(responses, resp)
	}
	return api.FormatBatchWriteResponse(responses)
}

// CreateLaneGroup 创建泳道组
func (s *Server) CreateLaneGroup(ctx context.Context, req *apitraffic.LaneGroup) *apimodel.Response {
	tx, err := s.storage.StartTx()
	if err != nil {
		log.Error("[Service][Lane] open store transaction fail", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	defer func() {
		_ = tx.Rollback()
	}()

	saveVal, err := s.storage.LockLaneGroup(tx, req.GetName())
	if err != nil {
		log.Error("[Service][Lane] lock one lane_group", utils.RequestID(ctx),
			zap.String("name", req.GetName()), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	if saveVal != nil {
		return api.NewResponse(apimodel.Code_ExistedResource)
	}
	saveData := &rules.LaneGroup{}
	if err := saveData.FromSpec(req); err != nil {
		log.Error("[Service][Lane] create lane_group transfer spec to model", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(apimodel.Code_ExecuteException)
	}
	saveData.ID = utils.DefaultString(req.GetId(), utils.NewUUID())
	saveData.Revision = utils.DefaultString(req.GetRevision(), utils.NewUUID())

	// 由于这里是新建，所以需要手动再把两个 flag 字段设置为 true 状态
	for i := range saveData.LaneRules {
		saveData.LaneRules[i].SetAddFlag(true)
		saveData.LaneRules[i].SetChangeEnable(true)
	}

	if err := s.storage.AddLaneGroup(tx, saveData); err != nil {
		log.Error("[Service][Lane] save lane_group", utils.RequestID(ctx), zap.String("name", saveData.Name), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	req.Id = saveData.ID

	if err := tx.Commit(); err != nil {
		log.Error("[Service][Lane] commit store transaction fail", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}

	s.RecordHistory(ctx, laneGroupRecordEntry(ctx, req, saveData, types.OCreate))
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req)
}

// UpdateLaneGroups 批量更新泳道组
func (s *Server) UpdateLaneGroups(ctx context.Context, req []*apitraffic.LaneGroup) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range req {
		resp := s.UpdateLaneGroup(ctx, req[i])
		api.Collect(responses, resp)
	}
	return api.FormatBatchWriteResponse(responses)
}

// UpdateLaneGroup 更新泳道组
func (s *Server) UpdateLaneGroup(ctx context.Context, req *apitraffic.LaneGroup) *apimodel.Response {
	tx, err := s.storage.StartTx()
	if err != nil {
		log.Error("[Service][Lane] open store transaction fail", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	defer func() {
		_ = tx.Rollback()
	}()

	saveData, err := s.storage.LockLaneGroup(tx, req.GetName())
	if err != nil {
		log.Error("[Service][Lane] lock one lane_group", utils.RequestID(ctx),
			zap.String("name", req.GetName()), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData == nil {
		log.Error("[Service][Lane] lock one lane_group not found", utils.RequestID(ctx),
			zap.String("name", req.GetName()))
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}

	needUpdate, err := updateLaneGroupAttribute(req, saveData)
	if err != nil {
		log.Error("[Service][Lane] update lane_group transfer spec to model", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(apimodel.Code_ExecuteException)
	}
	if !needUpdate {
		return api.NewResponse(apimodel.Code_NoNeedUpdate)
	}

	saveData.Revision = utils.DefaultString(req.GetRevision(), utils.NewUUID())
	if err := s.storage.UpdateLaneGroup(tx, saveData); err != nil {
		log.Error("[Service][Lane] update lane_group", utils.RequestID(ctx), zap.String("name", saveData.Name), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	req.Id = saveData.ID

	if err := tx.Commit(); err != nil {
		log.Error("[Service][Lane] commit store transaction fail", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}

	s.RecordHistory(ctx, laneGroupRecordEntry(ctx, req, saveData, types.OUpdate))
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req)
}

// DeleteLaneGroups 批量删除泳道组
func (s *Server) DeleteLaneGroups(ctx context.Context, req []*apitraffic.LaneGroup) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range req {
		resp := s.DeleteLaneGroup(ctx, req[i])
		api.Collect(responses, resp)
	}
	return api.FormatBatchWriteResponse(responses)
}

// DeleteLaneGroup 删除泳道组
func (s *Server) DeleteLaneGroup(ctx context.Context, req *apitraffic.LaneGroup) *apimodel.Response {
	var saveData *rules.LaneGroup
	var err error
	if req.GetId() != "" {
		saveData, err = s.storage.GetLaneGroupByID(req.GetId())
	} else {
		saveData, err = s.storage.GetLaneGroup(req.GetName())
	}
	if err != nil {
		log.Error("[Server][LaneGroup] get target lane_group when delete", zap.String("id", req.GetId()),
			zap.String("name", req.GetName()), utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData == nil {
		log.Info("[Server][LaneGroup] delete target lane_group but not found", zap.String("id", req.GetId()),
			zap.String("name", req.GetName()), utils.RequestID(ctx))
		return api.NewResponse(apimodel.Code_ExecuteSuccess)
	}

	saveData.Revision = utils.DefaultString(req.GetRevision(), utils.NewUUID())
	if err := s.storage.DeleteLaneGroup(saveData.ID); err != nil {
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	req.Id = saveData.ID
	s.RecordHistory(ctx, laneGroupRecordEntry(ctx, req, saveData, types.ODelete))
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req)
}

// GetLaneGroups 查询泳道组列表
func (s *Server) GetLaneGroups(ctx context.Context, filter map[string]string) *apimodel.BatchQueryResponse {
	offset, limit, _ := valid.ParseOffsetAndLimit(filter)
	total, ret, err := s.caches.LaneRule().Query(ctx, &cacheapi.LaneGroupArgs{
		Filter: filter,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Error("[Server][LaneGroup][Query] get lane_groups from store", utils.RequestID(ctx), zap.Error(err))
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}

	rsp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	rsp.Amount = total
	rsp.Size = uint32(len(ret))
	rsp.Data = make([]*anypb.Any, 0, len(ret))

	for i := range ret {
		data, err := ret[i].ToProto()
		if err != nil {
			log.Error("[Server][LaneGroup][Query] lane_group convert to proto", utils.RequestID(ctx), zap.Error(err))
			return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
		}
		anyData, err := anypb.New(proto.MessageV2(data.Proto))
		if err != nil {
			log.Error("[Server][LaneGroup][Query] lane_group convert to anypb", utils.RequestID(ctx), zap.Error(err))
			return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
		}
		rsp.Data = append(rsp.Data, anyData)
	}
	return rsp
}

// GetOneLaneGroup 查询单个泳道组
func (s *Server) GetOneLaneGroup(ctx context.Context, req *apitraffic.LaneGroup) *apimodel.Response {
	saveData, err := s.storage.GetLaneGroupByID(req.GetId())
	if err != nil {
		log.Error("[goverrule][lanegroup] get one lane_group from store", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData == nil {
		log.Info("[goverrule][lanegroup] lane_group not found", utils.RequestID(ctx))
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}

	view, err := saveData.ToSpec()
	if err != nil {
		log.Error("[goverrule][lanegroup] lane_group convert to spec", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(apimodel.Code_ExecuteException)
	}
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, view)
}

// CreateLaneRules 批量创建泳道规则
func (s *Server) CreateLaneRules(ctx context.Context, req []*apitraffic.LaneRule) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range req {
		resp := s.CreateLaneRule(ctx, req[i])
		api.Collect(responses, resp)
	}
	return api.FormatBatchWriteResponse(responses)
}

// CreateLaneRule 创建泳道规则
func (s *Server) CreateLaneRule(ctx context.Context, req *apitraffic.LaneRule) *apimodel.Response {
	tx, err := s.storage.StartTx()
	if err != nil {
		log.Error("[Service][Lane] open store transaction fail", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	defer func() {
		_ = tx.Rollback()
	}()

	group, err := s.storage.LockLaneGroup(tx, req.GetGroupName())
	if err != nil {
		log.Error("[Service][Lane] lock one lane_group", utils.RequestID(ctx),
			zap.String("name", req.GetGroupName()), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	if group == nil {
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}
	if len(group.LaneRules) > 10 {
		// 泳道组规则数量超过限制
		log.Error("[Service][Lane] create lane_group over limit", utils.RequestID(ctx), zap.String("name", req.GetGroupName()))
		return api.NewResponse(apimodel.Code_BatchSizeOverLimit)
	}

	saveData := &rules.LaneRule{}
	if err := saveData.FromSpec(req); err != nil {
		log.Error("[Service][Lane] create lane_group transfer spec to model", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(apimodel.Code_ExecuteException)
	}
	saveData.ID = utils.DefaultString(req.GetId(), utils.NewUUID())
	saveData.Revision = utils.DefaultString(req.GetRevision(), utils.NewUUID())

	if err := s.storage.AddLaneRules(tx, []*rules.LaneRule{saveData}); err != nil {
		log.Error("[Service][Lane] save lane_group", utils.RequestID(ctx), zap.String("name", saveData.Name), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	req.Id = saveData.ID

	if err := tx.Commit(); err != nil {
		log.Error("[Service][Lane] commit store transaction fail", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}

	s.RecordHistory(ctx, laneRuleRecordEntry(ctx, req, saveData, types.OCreate))
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req)
}

// UpdateLaneRules 批量更新泳道组
func (s *Server) UpdateLaneRules(ctx context.Context, req []*apitraffic.LaneRule) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range req {
		resp := s.UpdateLaneRule(ctx, req[i])
		api.Collect(responses, resp)
	}
	return api.FormatBatchWriteResponse(responses)
}

// UpdateLaneRule 更新泳道组
func (s *Server) UpdateLaneRule(ctx context.Context, req *apitraffic.LaneRule) *apimodel.Response {
	tx, err := s.storage.StartTx()
	if err != nil {
		log.Error("[Service][Lane] open store transaction fail", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	defer func() {
		_ = tx.Rollback()
	}()

	group, err := s.storage.LockLaneGroup(tx, req.GetGroupName())
	if err != nil {
		log.Error("[Service][Lane] lock one lane_group", utils.RequestID(ctx),
			zap.String("name", req.GetGroupName()), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	if group == nil {
		log.Error("[Service][Lane] lock one lane_group not found", utils.RequestID(ctx),
			zap.String("name", req.GetGroupName()))
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}
	saveData := &rules.LaneRule{}
	if err := saveData.FromSpec(req); err != nil {
		log.Error("[Service][Lane] create lane_group transfer spec to model", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(apimodel.Code_ExecuteException)
	}

	saveData.Revision = utils.DefaultString(req.GetRevision(), utils.NewUUID())
	if err := s.storage.UpdateLaneRules(tx, []*rules.LaneRule{saveData}); err != nil {
		log.Error("[Service][Lane] update lane_group", utils.RequestID(ctx), zap.String("name", saveData.Name), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	req.Id = saveData.ID

	if err := tx.Commit(); err != nil {
		log.Error("[Service][Lane] commit store transaction fail", utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}

	s.RecordHistory(ctx, laneRuleRecordEntry(ctx, req, saveData, types.OUpdate))
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req)
}

// DeleteLaneRules 批量删除泳道规则
func (s *Server) DeleteLaneRules(ctx context.Context, req []*apitraffic.LaneRule) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range req {
		resp := s.DeleteLaneRule(ctx, req[i])
		api.Collect(responses, resp)
	}
	return api.FormatBatchWriteResponse(responses)
}

// DeleteLaneRule 删除泳道规则
func (s *Server) DeleteLaneRule(ctx context.Context, req *apitraffic.LaneRule) *apimodel.Response {
	saveData, err := s.storage.GetLaneRule(req.GetId())
	if err != nil {
		log.Error("[Server][LaneGroup] get target lane_group when delete", zap.String("id", req.GetId()),
			zap.String("name", req.GetName()), utils.RequestID(ctx), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData == nil {
		log.Info("[Server][LaneGroup] delete target lane_group but not found", zap.String("id", req.GetId()),
			zap.String("name", req.GetName()), utils.RequestID(ctx))
		return api.NewResponse(apimodel.Code_ExecuteSuccess)
	}

	saveData.Revision = utils.DefaultString(req.GetRevision(), utils.NewUUID())
	if err := s.storage.DeleteLaneGroup(saveData.ID); err != nil {
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	req.Id = saveData.ID
	s.RecordHistory(ctx, laneRuleRecordEntry(ctx, req, saveData, types.ODelete))
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req)
}
func updateLaneGroupAttribute(req *apitraffic.LaneGroup, saveData *rules.LaneGroup) (bool, error) {
	updateData := &rules.LaneGroup{}
	if err := updateData.FromSpec(req); err != nil {
		return false, err
	}

	saveData.Description = updateData.Description
	return true, nil
}

// laneGroupRecordEntry 转换为鉴权策略的记录结构体
func laneGroupRecordEntry(ctx context.Context, req *apitraffic.LaneGroup, md *rules.LaneGroup,
	operationType types.OperationType) *types.RecordEntry {

	marshaler := jsonpb.Marshaler{}
	detail, _ := marshaler.MarshalToString(req)

	entry := &types.RecordEntry{
		ResourceType:  types.RLaneGroup,
		ResourceName:  fmt.Sprintf("%s(%s)", md.Name, md.ID),
		OperationType: operationType,
		Operator:      utils.ParseOperator(ctx),
		Detail:        detail,
		HappenTime:    time.Now(),
	}
	return entry
}

// laneRuleRecordEntry 转换为鉴权策略的记录结构体
func laneRuleRecordEntry(ctx context.Context, req *apitraffic.LaneRule, md *rules.LaneRule,
	operationType types.OperationType) *types.RecordEntry {

	marshaler := jsonpb.Marshaler{}
	detail, _ := marshaler.MarshalToString(req)

	entry := &types.RecordEntry{
		ResourceType:  types.RLaneRule,
		ResourceName:  fmt.Sprintf("%s(%s)", md.Name, md.ID),
		OperationType: operationType,
		Operator:      utils.ParseOperator(ctx),
		Detail:        detail,
		HappenTime:    time.Now(),
	}
	return entry
}
