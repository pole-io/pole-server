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

package config_auth

import (
	"context"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/auth"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

func (s *Server) PreviewConfigTemplate(
	ctx context.Context, req *apiconfig.RenderPreviewRequest) *apiconfig.RenderPreview {
	authCtx := s.collectConfigFileTemplateAuthContext(
		ctx, nil, auth.Read, auth.DescribeConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		code := auth.ConvertToErrCode(err)
		return &apiconfig.RenderPreview{
			Code: uint32(code),
			Info: err.Error(),
		}
	}
	ctx = context.WithValue(
		authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.PreviewConfigTemplate(ctx, req)
}

func (s *Server) PublishConfigTemplateRelease(
	ctx context.Context, req *apiconfig.ConfigTemplateRelease) *apimodel.Response {
	authCtx := s.collectConfigFileTemplateAuthContext(ctx, nil, auth.Create, auth.CreateConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.PublishConfigTemplateRelease(ctx, req)
}

func (s *Server) GetConfigTemplateLabels(ctx context.Context, templateID uint64) *apimodel.Response {
	authCtx := s.collectConfigFileTemplateAuthContext(ctx, nil, auth.Read, auth.DescribeConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.GetConfigTemplateLabels(ctx, templateID)
}

func (s *Server) SaveConfigTemplateLabels(ctx context.Context, templateID uint64,
	labels map[string]string) *apimodel.Response {
	authCtx := s.collectConfigFileTemplateAuthContext(ctx, nil, auth.Modify, auth.CreateConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.SaveConfigTemplateLabels(ctx, templateID, labels)
}

func (s *Server) SaveNamespaceTemplateValues(
	ctx context.Context, req *apiconfig.NamespaceTemplateValues) *apimodel.Response {
	authCtx := s.collectConfigFileTemplateAuthContext(ctx, nil, auth.Modify, auth.CreateConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.SaveNamespaceTemplateValues(ctx, req)
}

func (s *Server) GetNamespaceConfigTemplateDraft(
	ctx context.Context, namespace string, templateID uint64) *apimodel.Response {
	authCtx := s.collectConfigFileTemplateAuthContext(ctx, nil, auth.Read, auth.DescribeConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.GetNamespaceConfigTemplateDraft(ctx, namespace, templateID)
}

func (s *Server) SaveNamespaceConfigTemplateDraft(ctx context.Context,
	draft *conftypes.NamespaceConfigTemplateDraft) *apimodel.Response {
	authCtx := s.collectConfigFileTemplateAuthContext(ctx, nil, auth.Modify, auth.CreateConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.SaveNamespaceConfigTemplateDraft(ctx, draft)
}

func (s *Server) PublishNamespaceTemplateValueRelease(
	ctx context.Context, req *apiconfig.NamespaceTemplateValueRelease) *apimodel.Response {
	authCtx := s.collectConfigFileTemplateAuthContext(ctx, nil, auth.Create, auth.CreateConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.PublishNamespaceTemplateValueRelease(ctx, req)
}

func (s *Server) BindConfigFileTemplate(
	ctx context.Context, file *apiconfig.ConfigFile) *apimodel.Response {
	authCtx := s.collectConfigFileAuthContext(
		ctx, []*apiconfig.ConfigFile{file}, auth.Modify, auth.UpdateConfigFile)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.BindConfigFileTemplate(ctx, file)
}

func (s *Server) ListConfigTemplateReleases(
	ctx context.Context, templateID uint64) *apimodel.BatchQueryResponse {
	authCtx := s.collectConfigFileTemplateAuthContext(
		ctx, nil, auth.Read, auth.DescribeConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchQueryResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.ListConfigTemplateReleases(ctx, templateID)
}

func (s *Server) GetNamespaceTemplateValues(
	ctx context.Context, namespace string, templateID uint64) *apimodel.Response {
	authCtx := s.collectConfigFileTemplateAuthContext(
		ctx, nil, auth.Read, auth.DescribeConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.GetNamespaceTemplateValues(ctx, namespace, templateID)
}

func (s *Server) ListNamespaceTemplateValueReleases(
	ctx context.Context, namespace string, templateID uint64) *apimodel.BatchQueryResponse {
	authCtx := s.collectConfigFileTemplateAuthContext(
		ctx, nil, auth.Read, auth.DescribeConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchQueryResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.ListNamespaceTemplateValueReleases(ctx, namespace, templateID)
}

func (s *Server) ListConfigTemplateBindings(
	ctx context.Context, namespace, group, fileName string) *apimodel.BatchQueryResponse {
	authCtx := s.collectConfigFileAuthContext(ctx, []*apiconfig.ConfigFile{{
		Namespace: namespace,
		Group:     group,
		Name:      fileName,
	}}, auth.Read, auth.DescribeConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchQueryResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}
	ctx = context.WithValue(authCtx.GetRequestContext(), types.ContextAuthContextKey, authCtx)
	return s.nextServer.ListConfigTemplateBindings(ctx, namespace, group, fileName)
}

// GetAllConfigFileTemplates get all config file templates
func (s *Server) GetAllConfigFileTemplates(ctx context.Context) *apimodel.BatchQueryResponse {
	authCtx := s.collectConfigFileTemplateAuthContext(ctx,
		[]*apiconfig.ConfigFileTemplate{}, auth.Read, auth.DescribeAllConfigFileTemplates)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigFileBatchQueryResponseWithMessage(auth.ConvertToErrCode(err), err.Error())
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return s.nextServer.GetAllConfigFileTemplates(ctx)
}

// GetConfigFileTemplate get config file template
func (s *Server) GetConfigFileTemplate(ctx context.Context, name string) *apimodel.Response {
	authCtx := s.collectConfigFileTemplateAuthContext(ctx,
		[]*apiconfig.ConfigFileTemplate{}, auth.Read, auth.DescribeConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return s.nextServer.GetConfigFileTemplate(ctx, name)
}

// CreateConfigFileTemplates create config file template
func (s *Server) CreateConfigFileTemplates(ctx context.Context,
	reqs []*apiconfig.ConfigFileTemplate) *apimodel.Response {

	authCtx := s.collectConfigFileTemplateAuthContext(ctx, reqs, auth.Create, auth.CreateConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return s.nextServer.CreateConfigFileTemplates(ctx, reqs)
}

// UpdateConfigFileTemplates create config file template
func (s *Server) UpdateConfigFileTemplates(ctx context.Context,
	reqs []*apiconfig.ConfigFileTemplate) *apimodel.Response {

	authCtx := s.collectConfigFileTemplateAuthContext(ctx, reqs, auth.Create, auth.CreateConfigFileTemplate)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponseWithInfo(auth.ConvertToErrCode(err), err.Error())
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return s.nextServer.UpdateConfigFileTemplates(ctx, reqs)
}
