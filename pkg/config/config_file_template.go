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

	"go.uber.org/zap"

	apiconfig "github.com/polarismesh/specification/source/go/api/v1/config_manage"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// CreateConfigFileTemplates create config file template
func (s *Server) CreateConfigFileTemplates(
	ctx context.Context, reqs []*apiconfig.ConfigFileTemplate) *apiconfig.ConfigResponse {
	for _, req := range reqs {
		rsp := s.CreateConfigFileTemplate(ctx, req)
		if api.IsSuccess(rsp) {
			return rsp
		}
	}
	return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

// CreateConfigFileTemplate create config file template
func (s *Server) CreateConfigFileTemplate(
	ctx context.Context, req *apiconfig.ConfigFileTemplate) *apiconfig.ConfigResponse {
	name := req.GetName().GetValue()

	saveData, err := s.storage.GetConfigFileTemplate(name)
	if err != nil {
		log.Error("[Config][Service] get config file template error.",
			utils.RequestID(ctx), zap.String("name", name), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData != nil {
		return api.NewConfigResponse(apimodel.Code_ExistedResource)
	}

	userName := utils.ParseUserName(ctx)
	req.CreateBy = protobuf.NewStringValue(userName)
	req.ModifyBy = protobuf.NewStringValue(userName)
	saveData = conftypes.ToConfigFileTemplateStore(req)
	if _, err := s.storage.SaveConfigFileTemplate(saveData); err != nil {
		log.Error("[Config][Service] create config file template error.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

// UpdateConfigFileTemplates create config file template
func (s *Server) UpdateConfigFileTemplates(
	ctx context.Context, reqs []*apiconfig.ConfigFileTemplate) *apiconfig.ConfigResponse {
	for _, req := range reqs {
		rsp := s.UpdateConfigFileTemplate(ctx, req)
		if api.IsSuccess(rsp) {
			return rsp
		}
	}
	return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

// UpdateConfigFileTemplate create config file template
func (s *Server) UpdateConfigFileTemplate(
	ctx context.Context, req *apiconfig.ConfigFileTemplate) *apiconfig.ConfigResponse {
	name := req.GetName().GetValue()

	saveData, err := s.storage.GetConfigFileTemplate(name)
	if err != nil {
		log.Error("[Config][Service] get config file template error.",
			utils.RequestID(ctx), zap.String("name", name), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData == nil {
		return api.NewConfigResponse(apimodel.Code_NotFoundResource)
	}

	userName := utils.ParseUserName(ctx)
	req.CreateBy = protobuf.NewStringValue(saveData.CreateBy)
	req.ModifyBy = protobuf.NewStringValue(userName)
	saveData = conftypes.ToConfigFileTemplateStore(req)
	if _, err := s.storage.SaveConfigFileTemplate(saveData); err != nil {
		log.Error("[Config][Service] update config file template error.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

// GetConfigFileTemplate get config file template by name
func (s *Server) GetConfigFileTemplate(ctx context.Context, name string) *apiconfig.ConfigResponse {
	if len(name) == 0 {
		return api.NewConfigResponse(apimodel.Code_InvalidConfigFileTemplateName)
	}

	saveData, err := s.storage.GetConfigFileTemplate(name)
	if err != nil {
		log.Error("[Config][Service] get config file template error.",
			utils.RequestID(ctx), zap.String("name", name), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData == nil {
		return api.NewConfigResponse(apimodel.Code_NotFoundResource)
	}
	out := api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
	out.ConfigFileTemplate = conftypes.ToConfigFileTemplateAPI(saveData)
	return out
}

// GetAllConfigFileTemplates get all config file templates
func (s *Server) GetAllConfigFileTemplates(ctx context.Context) *apiconfig.ConfigBatchQueryResponse {
	templates, err := s.storage.QueryAllConfigFileTemplates()
	if err != nil {
		log.Error("[Config][Service]query all config file templates error.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}

	var apiTemplates []*apiconfig.ConfigFileTemplate
	for _, template := range templates {
		apiTemplates = append(apiTemplates, conftypes.ToConfigFileTemplateAPI(template))
	}
	return api.NewConfigFileTemplateBatchQueryResponse(apimodel.Code_ExecuteSuccess,
		uint32(len(templates)), apiTemplates)
}
