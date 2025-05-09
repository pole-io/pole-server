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

	apiconfig "github.com/polarismesh/specification/source/go/api/v1/config_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/auth"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// CreateConfigFile 创建配置文件
func (s *Server) CreateConfigFiles(ctx context.Context, reqs []*apiconfig.ConfigFile) *apiconfig.ConfigBatchWriteResponse {
	authCtx := s.collectConfigFileAuthContext(ctx, reqs, auth.Create, auth.CreateConfigFile)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchWriteResponse(auth.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return s.nextServer.CreateConfigFiles(ctx, reqs)
}

// UpdateConfigFile 更新配置文件
func (s *Server) UpdateConfigFiles(ctx context.Context, reqs []*apiconfig.ConfigFile) *apiconfig.ConfigBatchWriteResponse {
	authCtx := s.collectConfigFileAuthContext(ctx, reqs, auth.Modify, auth.UpdateConfigFile)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchWriteResponse(auth.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return s.nextServer.UpdateConfigFiles(ctx, reqs)
}

// DeleteConfigFile 删除配置文件，删除配置文件同时会通知客户端 Not_Found
func (s *Server) DeleteConfigFiles(ctx context.Context, reqs []*apiconfig.ConfigFile) *apiconfig.ConfigBatchWriteResponse {

	authCtx := s.collectConfigFileAuthContext(ctx, reqs, auth.Delete, auth.DeleteConfigFile)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchWriteResponse(auth.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return s.nextServer.DeleteConfigFiles(ctx, reqs)
}

// GetConfigFileRichInfo 获取单个配置文件基础信息，包含发布状态等信息
func (s *Server) GetConfigFileRichInfo(ctx context.Context,
	req *apiconfig.ConfigFile) *apiconfig.ConfigResponse {

	authCtx := s.collectConfigFileAuthContext(
		ctx, []*apiconfig.ConfigFile{req}, auth.Read, auth.DescribeConfigFileRichInfo)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponse(auth.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return s.nextServer.GetConfigFileRichInfo(ctx, req)
}

// SearchConfigFiles 查询配置文件
func (s *Server) SearchConfigFiles(ctx context.Context,
	filter map[string]string) *apiconfig.ConfigBatchQueryResponse {

	authCtx := s.collectConfigFileAuthContext(ctx, nil, auth.Read, auth.DescribeConfigFiles)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchQueryResponse(auth.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return s.nextServer.SearchConfigFiles(ctx, filter)
}

func (s *Server) ExportConfigFile(ctx context.Context,
	configFileExport *apiconfig.ConfigFileExportRequest) *apiconfig.ConfigExportResponse {
	var configFiles []*apiconfig.ConfigFile
	for _, group := range configFileExport.Groups {
		configFile := &apiconfig.ConfigFile{
			Namespace: configFileExport.Namespace,
			Group:     group,
		}
		configFiles = append(configFiles, configFile)
	}
	authCtx := s.collectConfigFileAuthContext(ctx, configFiles, auth.Read, auth.ExportConfigFiles)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigFileExportResponse(auth.ConvertToErrCode(err), nil)
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return s.nextServer.ExportConfigFile(ctx, configFileExport)
}

func (s *Server) ImportConfigFile(ctx context.Context,
	configFiles []*apiconfig.ConfigFile, conflictHandling string) *apiconfig.ConfigImportResponse {
	authCtx := s.collectConfigFileAuthContext(ctx, configFiles, auth.Create, auth.ImportConfigFiles)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewSimpleConfigFileImportResponse(auth.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return s.nextServer.ImportConfigFile(ctx, configFiles, conflictHandling)
}

func (s *Server) GetAllConfigEncryptAlgorithms(
	ctx context.Context) *apiconfig.ConfigEncryptAlgorithmResponse {
	return s.nextServer.GetAllConfigEncryptAlgorithms(ctx)
}

// GetClientSubscribers 获取客户端订阅者
func (s *Server) GetClientSubscribers(ctx context.Context, filter map[string]string) *types.CommonResponse {
	return s.nextServer.GetClientSubscribers(ctx, filter)
}

// GetConfigSubscribers 获取配置订阅者
func (s *Server) GetConfigSubscribers(ctx context.Context, filter map[string]string) *types.CommonResponse {
	return s.nextServer.GetConfigSubscribers(ctx, filter)
}
