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

// PublishConfigFile 发布配置文件
func (s *Server) PublishConfigFile(ctx context.Context,
	configFileRelease *apiconfig.ConfigFileRelease) *apiconfig.ConfigResponse {

	authCtx := s.collectConfigFileReleaseAuthContext(ctx,
		[]*apiconfig.ConfigFileRelease{configFileRelease}, auth.Modify, "PublishConfigFile")

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponse(auth.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return s.nextServer.PublishConfigFile(ctx, configFileRelease)
}

// GetConfigFileRelease 获取配置文件发布内容
func (s *Server) GetConfigFileRelease(ctx context.Context,
	req *apiconfig.ConfigFileRelease) *apiconfig.ConfigResponse {

	authCtx := s.collectConfigFileReleaseAuthContext(ctx,
		[]*apiconfig.ConfigFileRelease{req}, auth.Read, auth.DescribeConfigFileRelease)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponse(auth.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return s.nextServer.GetConfigFileRelease(ctx, req)
}

// DeleteConfigFileReleases implements ConfigCenterServer.
func (s *Server) DeleteConfigFileReleases(ctx context.Context,
	reqs []*apiconfig.ConfigFileRelease) *apiconfig.ConfigBatchWriteResponse {

	authCtx := s.collectConfigFileReleaseAuthContext(ctx, reqs, auth.Delete, auth.DeleteConfigFileReleases)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchWriteResponse(auth.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return s.nextServer.DeleteConfigFileReleases(ctx, reqs)
}

// GetConfigFileReleaseVersions implements ConfigCenterServer.
func (s *Server) GetConfigFileReleaseVersions(ctx context.Context,
	filters map[string]string) *apiconfig.ConfigBatchQueryResponse {
	authCtx := s.collectConfigFileReleaseAuthContext(ctx, nil, auth.Read, auth.DescribeConfigFileReleaseVersions)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchQueryResponse(auth.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return s.nextServer.GetConfigFileReleaseVersions(ctx, filters)
}

// GetConfigFileReleases implements ConfigCenterServer.
func (s *Server) GetConfigFileReleases(ctx context.Context,
	filters map[string]string) *apiconfig.ConfigBatchQueryResponse {

	authCtx := s.collectConfigFileReleaseAuthContext(ctx, nil, auth.Read, auth.DescribeConfigFileReleases)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchQueryResponse(auth.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return s.nextServer.GetConfigFileReleases(ctx, filters)
}

// RollbackConfigFileReleases implements ConfigCenterServer.
func (s *Server) RollbackConfigFileReleases(ctx context.Context,
	reqs []*apiconfig.ConfigFileRelease) *apiconfig.ConfigBatchWriteResponse {

	authCtx := s.collectConfigFileReleaseAuthContext(ctx, reqs, auth.Modify, auth.RollbackConfigFileReleases)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchWriteResponse(auth.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	return s.nextServer.RollbackConfigFileReleases(ctx, reqs)
}

// UpsertAndReleaseConfigFile .
func (s *Server) UpsertAndReleaseConfigFile(ctx context.Context,
	req *apiconfig.ConfigFilePublishInfo) *apiconfig.ConfigResponse {
	authCtx := s.collectConfigFilePublishAuthContext(ctx, []*apiconfig.ConfigFilePublishInfo{req},
		auth.Modify, auth.UpsertAndReleaseConfigFile)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponse(auth.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return s.nextServer.UpsertAndReleaseConfigFile(ctx, req)
}

func (s *Server) StopGrayConfigFileReleases(ctx context.Context,
	reqs []*apiconfig.ConfigFileRelease) *apiconfig.ConfigBatchWriteResponse {

	authCtx := s.collectConfigFileReleaseAuthContext(ctx, reqs,
		auth.Modify, auth.StopGrayConfigFileReleases)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchWriteResponse(auth.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return s.nextServer.StopGrayConfigFileReleases(ctx, reqs)
}
