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

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
	configtemplate "github.com/pole-io/pole-server/pkg/config/template"
)

func (s *Server) PreviewConfigTemplate(
	ctx context.Context, req *apiconfig.RenderPreviewRequest) *apiconfig.RenderPreview {
	if req == nil || req.GetInput() == nil ||
		CheckContentLength(req.GetInput().GetContent(), int(s.cfg.ContentMaxLength)) != nil {
		return &apiconfig.RenderPreview{
			Code: uint32(apimodel.Code_InvalidParameter),
			Info: "template render input is missing or exceeds the configured content limit",
		}
	}
	engine := req.GetInput().GetEngine()
	if engine == nil || engine.GetName() != configtemplate.EnginePoleMustache ||
		engine.GetVersion() != configtemplate.EngineVersionV1 {
		return &apiconfig.RenderPreview{
			Code: uint32(apimodel.Code_InvalidParameter),
			Info: "unsupported template engine",
		}
	}
	return s.nextServer.PreviewConfigTemplate(ctx, req)
}

func (s *Server) PublishConfigTemplateRelease(
	ctx context.Context, req *apiconfig.ConfigTemplateRelease) *apimodel.Response {
	if req == nil || CheckContentLength(req.GetContent(), int(s.cfg.ContentMaxLength)) != nil {
		return api.NewConfigResponse(apimodel.Code_InvalidParameter)
	}
	return s.nextServer.PublishConfigTemplateRelease(ctx, req)
}

func (s *Server) GetConfigTemplateLabels(ctx context.Context, templateID uint64) *apimodel.Response {
	if templateID == 0 {
		return api.NewConfigResponse(apimodel.Code_InvalidParameter)
	}
	return s.nextServer.GetConfigTemplateLabels(ctx, templateID)
}

func (s *Server) SaveConfigTemplateLabels(ctx context.Context, templateID uint64,
	labels map[string]string) *apimodel.Response {
	if templateID == 0 || len(labels) > 64 {
		return api.NewConfigResponse(apimodel.Code_InvalidParameter)
	}
	for key, value := range labels {
		if key == "" || len(key) > 128 || len(value) > 256 {
			return api.NewConfigResponse(apimodel.Code_InvalidParameter)
		}
	}
	return s.nextServer.SaveConfigTemplateLabels(ctx, templateID, labels)
}

func (s *Server) SaveNamespaceTemplateValues(
	ctx context.Context, req *apiconfig.NamespaceTemplateValues) *apimodel.Response {
	if req == nil || req.GetNamespace() == "" || len(req.GetValues()) > 1024 {
		return api.NewConfigResponse(apimodel.Code_InvalidParameter)
	}
	return s.nextServer.SaveNamespaceTemplateValues(ctx, req)
}

func (s *Server) PublishNamespaceTemplateValueRelease(
	ctx context.Context, req *apiconfig.NamespaceTemplateValueRelease) *apimodel.Response {
	if req == nil || req.GetNamespace() == "" || len(req.GetValues()) > 1024 {
		return api.NewConfigResponse(apimodel.Code_InvalidParameter)
	}
	return s.nextServer.PublishNamespaceTemplateValueRelease(ctx, req)
}

func (s *Server) BindConfigFileTemplate(
	ctx context.Context, file *apiconfig.ConfigFile) *apimodel.Response {
	if file == nil || CheckContentLength(file.GetContent(), int(s.cfg.ContentMaxLength)) != nil {
		return api.NewConfigResponse(apimodel.Code_InvalidParameter)
	}
	return s.nextServer.BindConfigFileTemplate(ctx, file)
}

func (s *Server) ListConfigTemplateReleases(
	ctx context.Context, templateID uint64) *apimodel.BatchQueryResponse {
	if templateID == 0 {
		return api.NewConfigBatchQueryResponseWithInfo(apimodel.Code_BadRequest, "invalid template_id")
	}
	return s.nextServer.ListConfigTemplateReleases(ctx, templateID)
}

func (s *Server) GetNamespaceTemplateValues(
	ctx context.Context, namespace string, templateID uint64) *apimodel.Response {
	if templateID == 0 || valid.CheckResourceName(namespace) != nil {
		return api.NewConfigResponseWithInfo(apimodel.Code_BadRequest, "invalid namespace or template_id")
	}
	return s.nextServer.GetNamespaceTemplateValues(ctx, namespace, templateID)
}

func (s *Server) ListNamespaceTemplateValueReleases(
	ctx context.Context, namespace string, templateID uint64) *apimodel.BatchQueryResponse {
	if templateID == 0 || valid.CheckResourceName(namespace) != nil {
		return api.NewConfigBatchQueryResponseWithInfo(
			apimodel.Code_BadRequest, "invalid namespace or template_id")
	}
	return s.nextServer.ListNamespaceTemplateValueReleases(ctx, namespace, templateID)
}

func (s *Server) ListConfigTemplateBindings(
	ctx context.Context, namespace, group, fileName string) *apimodel.BatchQueryResponse {
	if valid.CheckResourceName(namespace) != nil || valid.CheckResourceName(group) != nil ||
		CheckFileName(fileName) != nil {
		return api.NewConfigBatchQueryResponseWithInfo(
			apimodel.Code_BadRequest, "invalid namespace, group, or file_name")
	}
	return s.nextServer.ListConfigTemplateBindings(ctx, namespace, group, fileName)
}

// GetAllConfigFileTemplates get all config file templates
func (s *Server) GetAllConfigFileTemplates(ctx context.Context) *apimodel.BatchQueryResponse {

	return s.nextServer.GetAllConfigFileTemplates(ctx)
}

// GetConfigFileTemplate get config file template
func (s *Server) GetConfigFileTemplate(ctx context.Context, name string) *apimodel.Response {

	return s.nextServer.GetConfigFileTemplate(ctx, name)
}

// CreateConfigFileTemplate create config file template
func (s *Server) CreateConfigFileTemplates(ctx context.Context,
	reqs []*apiconfig.ConfigFileTemplate) *apimodel.Response {
	if len(reqs) == 0 {
		return api.NewConfigResponse(apimodel.Code_BadRequest)
	}
	for _, t := range reqs {
		if checkRsp := s.checkConfigFileTemplateParam(t); checkRsp != nil {
			return checkRsp
		}
	}
	return s.nextServer.CreateConfigFileTemplates(ctx, reqs)
}

// UpdateConfigFileTemplates create config file template
func (s *Server) UpdateConfigFileTemplates(ctx context.Context,
	reqs []*apiconfig.ConfigFileTemplate) *apimodel.Response {
	if len(reqs) == 0 {
		return api.NewConfigResponse(apimodel.Code_BadRequest)
	}
	for _, t := range reqs {
		if checkRsp := s.checkConfigFileTemplateParam(t); checkRsp != nil {
			return checkRsp
		}
	}
	return s.nextServer.UpdateConfigFileTemplates(ctx, reqs)
}

func (s *Server) checkConfigFileTemplateParam(template *apiconfig.ConfigFileTemplate) *apimodel.Response {
	if err := CheckFileName(template.GetName()); err != nil {
		return api.NewConfigResponse(apimodel.Code_InvalidParameter)
	}
	if err := CheckContentLength(template.Content, int(s.cfg.ContentMaxLength)); err != nil {
		return api.NewConfigResponse(apimodel.Code_InvalidParameter)
	}
	if len(template.Content) == 0 {
		return api.NewConfigFileTemplateResponseWithMessage(apimodel.Code_BadRequest, "content can not be blank.")
	}
	return nil
}
