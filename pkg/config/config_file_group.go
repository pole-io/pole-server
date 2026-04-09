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

package config

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
)

// CreateConfigFileGroups 批量创建配置文件组
func (s *Server) CreateConfigFileGroups(ctx context.Context, reqs []*apiconfig.ConfigFileGroup) *apimodel.BatchWriteResponse {
	bRsp := api.NewConfigBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, req := range reqs {
		rsp := s.CreateConfigFileGroup(ctx, req)
		api.ConfigCollect(bRsp, rsp)
	}
	return bRsp
}

// CreateConfigFileGroup 创建配置文件组
func (s *Server) CreateConfigFileGroup(ctx context.Context, req *apiconfig.ConfigFileGroup) *apimodel.Response {
	namespace := req.Namespace
	groupName := req.Name

	// 如果 namespace 不存在则自动创建
	if _, rsp := s.namespaceOperator.CreateNamespaceIfAbsent(ctx, &apimodel.Namespace{
		Name: req.GetNamespace(),
	}); !api.IsSuccess(rsp) {
		log.Error("[Config][Group] create namespace failed.", utils.RequestID(ctx),
			utils.ZapNamespace(namespace), utils.ZapGroup(groupName), zap.String("err", rsp.String()))
		return api.NewConfigResponse(apimodel.Code(rsp.Code))
	}

	fileGroup, err := s.storage.GetConfigFileGroup(namespace, groupName)
	if err != nil {
		log.Error("[Config][Group] get config file group error.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if fileGroup != nil {
		return api.NewConfigResponse(apimodel.Code_ExistedResource)
	}

	saveData := conftypes.ToConfigGroupStore(req)
	saveData.CreateBy = utils.ParseUserName(ctx)
	saveData.ModifyBy = utils.ParseUserName(ctx)

	ret, err := s.storage.CreateConfigFileGroup(saveData)
	if err != nil {
		log.Error("[Config][Group] create config file group error.", utils.RequestID(ctx),
			utils.ZapNamespace(namespace), utils.ZapGroup(groupName), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	log.Info("[Config][Group] create config file group successful.", utils.RequestID(ctx),
		utils.ZapNamespace(namespace), utils.ZapGroup(groupName))

	// 这里设置在 config-group 的 id 信息
	req.Id = ret.Id
	s.RecordHistory(ctx, configGroupRecordEntry(ctx, req, saveData, types.OCreate))
	return api.NewConfigGroupResponse(apimodel.Code_ExecuteSuccess, &apiconfig.ConfigFileGroup{
		Id:        saveData.Id,
		Namespace: saveData.Namespace,
		Name:      saveData.Name,
	})
}

// UpdateConfigFileGroups 批量更新配置文件组
func (s *Server) UpdateConfigFileGroups(ctx context.Context, reqs []*apiconfig.ConfigFileGroup) *apimodel.BatchWriteResponse {
	bRsp := api.NewConfigBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, req := range reqs {
		rsp := s.UpdateConfigFileGroup(ctx, req)
		api.ConfigCollect(bRsp, rsp)
	}
	return bRsp
}

// UpdateConfigFileGroup 更新配置文件组
func (s *Server) UpdateConfigFileGroup(ctx context.Context, req *apiconfig.ConfigFileGroup) *apimodel.Response {
	namespace := req.Namespace
	groupName := req.Name

	saveData, err := s.storage.GetConfigFileGroup(namespace, groupName)
	if err != nil {
		log.Error("[Config][Group] get config file group failed. ", utils.RequestID(ctx),
			utils.ZapNamespace(namespace), utils.ZapGroup(groupName), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData == nil {
		return api.NewConfigResponse(apimodel.Code_NotFoundResource)
	}

	updateData := conftypes.ToConfigGroupStore(req)
	updateData.ModifyBy = utils.ParseOperator(ctx)
	updateData, needUpdate := s.UpdateGroupAttribute(saveData, updateData)
	if !needUpdate {
		return api.NewConfigResponse(apimodel.Code_NoNeedUpdate)
	}

	if err := s.storage.UpdateConfigFileGroup(updateData); err != nil {
		log.Error("[Config][Group] update config file group failed. ", utils.RequestID(ctx),
			utils.ZapNamespace(namespace), utils.ZapGroup(groupName), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	req.Id = saveData.Id
	s.RecordHistory(ctx, configGroupRecordEntry(ctx, req, updateData, types.OUpdate))
	return api.NewConfigGroupResponse(apimodel.Code_ExecuteSuccess, &apiconfig.ConfigFileGroup{
		Id:        updateData.Id,
		Namespace: updateData.Namespace,
		Name:      updateData.Name,
	})
}

func (s *Server) UpdateGroupAttribute(saveData, updateData *conftypes.ConfigFileGroup) (*conftypes.ConfigFileGroup, bool) {
	needUpdate := false
	if saveData.Comment != updateData.Comment {
		needUpdate = true
		saveData.Comment = updateData.Comment
	}
	if saveData.Business != updateData.Business {
		needUpdate = true
		saveData.Business = updateData.Business
	}
	if saveData.Department != updateData.Department {
		needUpdate = true
		saveData.Department = updateData.Department
	}
	if utils.IsNotEqualMap(updateData.Metadata, saveData.Metadata) {
		needUpdate = true
		saveData.Metadata = updateData.Metadata
	}
	return saveData, needUpdate
}

// createConfigFileGroupIfAbsent 如果不存在配置文件组，则自动创建
func (s *Server) createConfigFileGroupIfAbsent(ctx context.Context,
	configFileGroup *apiconfig.ConfigFileGroup) *apimodel.Response {
	var (
		namespace = configFileGroup.Namespace
		name      = configFileGroup.Name
	)

	group, err := s.storage.GetConfigFileGroup(namespace, name)
	if err != nil {
		log.Error("[Config][Group] query config file group error.", utils.RequestID(ctx),
			utils.ZapNamespace(namespace), utils.ZapGroup(name), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if group != nil {
		return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
	}
	return s.CreateConfigFileGroup(ctx, configFileGroup)
}

// DeleteConfigFileGroups 批量删除配置文件组
func (s *Server) DeleteConfigFileGroups(ctx context.Context, reqs []*apiconfig.ConfigFileGroup) *apimodel.BatchWriteResponse {
	bRsp := api.NewConfigBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, req := range reqs {
		rsp := s.DeleteConfigFileGroup(ctx, req.Namespace, req.Name)
		api.ConfigCollect(bRsp, rsp)
	}
	return bRsp
}

// DeleteConfigFileGroup 删除配置文件组
func (s *Server) DeleteConfigFileGroup(ctx context.Context, namespace, name string) *apimodel.Response {
	log.Info("[Config][Group] delete config file group. ", utils.RequestID(ctx),
		utils.ZapNamespace(namespace), utils.ZapGroup(name))

	configGroup, err := s.storage.GetConfigFileGroup(namespace, name)
	if err != nil {
		log.Error("[Config][Group] get config file group failed. ", utils.RequestID(ctx),
			utils.ZapNamespace(namespace), utils.ZapGroup(name), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if configGroup == nil {
		return api.NewConfigResponse(apimodel.Code_NotFoundResource)
	}
	if errResp := s.hasResourceInConfigGroup(ctx, namespace, name); errResp != nil {
		return errResp
	}

	if err := s.storage.DeleteConfigFileGroup(namespace, name); err != nil {
		log.Error("[Config][Group] delete config file group failed. ", utils.RequestID(ctx),
			utils.ZapNamespace(namespace), utils.ZapGroup(name), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	s.RecordHistory(ctx, configGroupRecordEntry(ctx, &apiconfig.ConfigFileGroup{
		Id:        configGroup.Id,
		Namespace: configGroup.Namespace,
		Name:      configGroup.Name,
	}, configGroup, types.ODelete))
	return api.NewConfigGroupResponse(apimodel.Code_ExecuteSuccess, &apiconfig.ConfigFileGroup{
		Id:        configGroup.Id,
		Namespace: configGroup.Namespace,
		Name:      configGroup.Name,
	})
}

func (s *Server) hasResourceInConfigGroup(ctx context.Context, namespace, name string) *apimodel.Response {
	total, err := s.storage.CountConfigFiles(namespace, name)
	if err != nil {
		log.Error("[Config][Group] get config file group failed. ", utils.RequestID(ctx),
			utils.ZapNamespace(namespace), utils.ZapGroup(name), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if total != 0 {
		return api.NewConfigResponse(apimodel.Code_ExistedResource)
	}
	total, err = s.storage.CountConfigReleases(namespace, name, true)
	if err != nil {
		log.Error("[Config][Group] get config file group failed. ", utils.RequestID(ctx),
			utils.ZapNamespace(namespace), utils.ZapGroup(name), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if total != 0 {
		return api.NewConfigResponse(apimodel.Code_ExistedResource)
	}
	return nil
}

// QueryConfigFileGroups 查询配置文件组
func (s *Server) QueryConfigFileGroups(ctx context.Context,
	searchFilters map[string]string) *apimodel.BatchQueryResponse {

	offset, limit, _ := valid.ParseOffsetAndLimit(searchFilters)

	args := &cacheapi.ConfigGroupArgs{
		Namespace:  searchFilters["namespace"],
		Name:       searchFilters["name"],
		Business:   searchFilters["business"],
		Department: searchFilters["department"],
		Offset:     offset,
		Limit:      limit,
		OrderField: searchFilters["order_field"],
		OrderType:  searchFilters["order_type"],
	}

	total, ret, err := s.groupCache.Query(args)
	if err != nil {
		resp := api.NewConfigBatchQueryResponse(storeapi.StoreCode2APICode(err))
		resp.Info = err.Error()
		return resp
	}
	values := make([]*apiconfig.ConfigFileGroup, 0, len(ret))
	for i := range ret {
		item := conftypes.ToConfigGroupAPI(ret[i])
		fileCount, err := s.storage.CountConfigFiles(ret[i].Namespace, ret[i].Name)
		if err != nil {
			log.Error("[Config][Service] get config file count for group error.", utils.RequestID(ctx),
				utils.ZapNamespace(ret[i].Namespace), utils.ZapGroup(ret[i].Name), zap.Error(err))
		}

		// 如果包含特殊标签，也不允许修改
		if _, ok := item.GetMetadata()[types.MetaKey3RdPlatform]; ok {
			item.Editable = false
		}

		item.FileCount = fileCount
		values = append(values, item)
	}

	resp := api.NewConfigBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = total
	resp.Size = uint32(len(values))
	for i := range values {
		if err := api.AddAnyDataIntoBatchQuery(resp, values[i]); err != nil {
			log.Error("[Config][Service] add config file group to response data", utils.RequestID(ctx), zap.Error(err))
			return api.NewConfigBatchQueryResponse(apimodel.Code_ExecuteException)
		}
	}
	return resp
}

// configGroupRecordEntry 生成服务的记录entry
func configGroupRecordEntry(ctx context.Context, req *apiconfig.ConfigFileGroup, md *conftypes.ConfigFileGroup,
	operationType types.OperationType) *types.RecordEntry {

	detail, _ := protojson.Marshal(req)
	detailStr := string(detail)

	entry := &types.RecordEntry{
		ResourceType:  types.RConfigGroup,
		ResourceName:  req.GetName(),
		Namespace:     req.GetNamespace(),
		OperationType: operationType,
		Operator:      utils.ParseOperator(ctx),
		Detail:        detailStr,
		HappenTime:    time.Now(),
	}

	return entry
}
