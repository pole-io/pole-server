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
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	configtemplate "github.com/pole-io/pole-server/pkg/config/template"
)

// PreviewConfigTemplate performs a deterministic reference render. The result is
// intended for preview, validation and cross-language hash comparison; it is not
// the runtime configuration payload.
func (s *Server) PreviewConfigTemplate(
	ctx context.Context, req *apiconfig.RenderPreviewRequest) *apiconfig.RenderPreview {
	out := &apiconfig.RenderPreview{
		TemplateReleaseId: req.GetTemplateReleaseId(),
		ValueReleaseId:    req.GetValueReleaseId(),
		Code:              uint32(apimodel.Code_ExecuteSuccess),
		Info:              apimodel.Code_ExecuteSuccess.String(),
		Engine: &apiconfig.ConfigTemplateEngine{
			Name: configtemplate.EnginePoleMustache, Version: configtemplate.EngineVersionV1,
		},
	}
	previewReq, err := configtemplate.RenderRequestFromSpec(req.GetInput())
	if err != nil {
		out.Diagnostics = []*apiconfig.RenderDiagnostic{{
			Severity: apiconfig.RenderDiagnostic_DIAGNOSTIC_ERROR,
			Code:     string(configtemplate.DiagnosticInvalidValue),
			Message:  err.Error(),
		}}
		return out
	}
	result, err := configtemplate.Preview(ctx, previewReq)
	if err != nil {
		out.Diagnostics = []*apiconfig.RenderDiagnostic{{
			Severity: apiconfig.RenderDiagnostic_DIAGNOSTIC_ERROR,
			Code:     string(configtemplate.DiagnosticInvalidValue),
			Message:  err.Error(),
		}}
		return out
	}
	out.Valid = result.Valid
	out.Format = result.Format
	out.RenderedContent = result.RenderedContent
	out.RenderedSha256 = result.RenderedSHA256
	out.Diagnostics = make([]*apiconfig.RenderDiagnostic, 0, len(result.Diagnostics))
	for _, diagnostic := range result.Diagnostics {
		out.Diagnostics = append(out.Diagnostics, &apiconfig.RenderDiagnostic{
			Severity:  apiconfig.RenderDiagnostic_DIAGNOSTIC_ERROR,
			Code:      string(diagnostic.Code),
			Message:   diagnostic.Message,
			Parameter: diagnostic.Parameter,
		})
	}
	return out
}

// CreateConfigFileTemplates create config file template
func (s *Server) CreateConfigFileTemplates(
	ctx context.Context, reqs []*apiconfig.ConfigFileTemplate) *apimodel.Response {
	for _, req := range reqs {
		rsp := s.CreateConfigFileTemplate(ctx, req)
		if !api.IsSuccess(rsp) {
			return rsp
		}
	}
	return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

// CreateConfigFileTemplate create config file template
func (s *Server) CreateConfigFileTemplate(
	ctx context.Context, req *apiconfig.ConfigFileTemplate) *apimodel.Response {
	name := req.GetName()

	saveData, err := s.storage.GetConfigFileTemplate(name)
	if err != nil {
		log.Error("[Config][Service] get config file template error.",
			utils.RequestID(ctx), zap.String("name", name), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData != nil {
		return api.NewConfigResponse(apimodel.Code_ExistedResource)
	}

	saveData = conftypes.ToConfigFileTemplateStore(req)
	if _, err := s.storage.SaveConfigFileTemplate(saveData); err != nil {
		log.Error("[Config][Service] create config file template error.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

func (s *Server) GetConfigTemplateLabels(_ context.Context, templateID uint64) *apimodel.Response {
	template, err := s.findConfigTemplateByID(templateID)
	if err != nil {
		return api.NewConfigResponseWithInfo(apimodel.Code_NotFoundResource, err.Error())
	}
	labels := make(map[string]any, len(template.Labels))
	for key, value := range template.Labels {
		labels[key] = value
	}
	data, err := structpb.NewStruct(map[string]any{"template_id": float64(templateID), "labels": labels})
	if err != nil {
		return api.NewConfigResponse(apimodel.Code_ExecuteException)
	}
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, data)
}

func (s *Server) SaveConfigTemplateLabels(_ context.Context, templateID uint64,
	labels map[string]string) *apimodel.Response {
	template, err := s.findConfigTemplateByID(templateID)
	if err != nil {
		return api.NewConfigResponseWithInfo(apimodel.Code_NotFoundResource, err.Error())
	}
	template.Labels = labels
	if _, err := s.storage.SaveConfigFileTemplate(template); err != nil {
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

func (s *Server) findConfigTemplateByID(templateID uint64) (*conftypes.ConfigFileTemplate, error) {
	template, err := s.storage.GetConfigFileTemplateByID(templateID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, fmt.Errorf("config template %d not found", templateID)
	}
	return template, nil
}

// UpdateConfigFileTemplates create config file template
func (s *Server) UpdateConfigFileTemplates(
	ctx context.Context, reqs []*apiconfig.ConfigFileTemplate) *apimodel.Response {
	for _, req := range reqs {
		rsp := s.UpdateConfigFileTemplate(ctx, req)
		if !api.IsSuccess(rsp) {
			return rsp
		}
	}
	return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

// UpdateConfigFileTemplate create config file template
func (s *Server) UpdateConfigFileTemplate(
	ctx context.Context, req *apiconfig.ConfigFileTemplate) *apimodel.Response {
	name := req.GetName()

	saveData, err := s.storage.GetConfigFileTemplate(name)
	if err != nil {
		log.Error("[Config][Service] get config file template error.",
			utils.RequestID(ctx), zap.String("name", name), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData == nil {
		return api.NewConfigResponse(apimodel.Code_NotFoundResource)
	}

	labels := saveData.Labels
	saveData = conftypes.ToConfigFileTemplateStore(req)
	saveData.Labels = labels
	if _, err := s.storage.SaveConfigFileTemplate(saveData); err != nil {
		log.Error("[Config][Service] update config file template error.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

// GetConfigFileTemplate get config file template by name
func (s *Server) GetConfigFileTemplate(ctx context.Context, name string) *apimodel.Response {
	if len(name) == 0 {
		return api.NewConfigResponse(apimodel.Code_BadRequest)
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
	template := conftypes.ToConfigFileTemplateAPI(saveData)
	return api.NewConfigFileTemplateResponse(apimodel.Code_ExecuteSuccess, template)
}

// GetAllConfigFileTemplates get all config file templates
func (s *Server) GetAllConfigFileTemplates(ctx context.Context) *apimodel.BatchQueryResponse {
	templates, err := s.storage.QueryAllConfigFileTemplates()
	if err != nil {
		log.Error("[Config][Service]query all config file templates error.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}

	out := api.NewConfigBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	out.Amount = uint32(len(templates))
	out.Size = uint32(len(templates))
	for _, template := range templates {
		item := conftypes.ToConfigFileTemplateAPI(template)
		if err := api.AddAnyDataIntoBatchQuery(out, item); err != nil {
			log.Error("[Config][Service] add config file template to response data", utils.RequestID(ctx), zap.Error(err))
			return api.NewConfigBatchQueryResponse(apimodel.Code_ExecuteException)
		}
	}
	return out
}
