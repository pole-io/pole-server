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
)

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
	if err := CheckContentLength(template.GetContent(), int(s.cfg.ContentMaxLength)); err != nil {
		return api.NewConfigResponse(apimodel.Code_InvalidParameter)
	}
	if len(template.GetContent()) == 0 {
		return api.NewConfigFileTemplateResponseWithMessage(apimodel.Code_BadRequest, "content can not be blank.")
	}
	return nil
}
