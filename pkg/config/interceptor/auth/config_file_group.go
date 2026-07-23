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
	"github.com/pole-io/specification/source/go/api/v1/security"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// CreateConfigFileGroups 创建配置文件组
func (s *Server) CreateConfigFileGroups(ctx context.Context,
	reqs []*apiconfig.ConfigFileGroup) *apimodel.BatchWriteResponse {
	authCtx := s.collectConfigGroupAuthContext(ctx, reqs, authtypes.Create, authtypes.CreateConfigFileGroup)

	// 验证 token 信息
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	resp := s.nextServer.CreateConfigFileGroups(ctx, reqs)

	return resp
}

// UpdateConfigFileGroups 更新配置文件组
func (s *Server) UpdateConfigFileGroups(ctx context.Context, reqs []*apiconfig.ConfigFileGroup) *apimodel.BatchWriteResponse {
	authCtx := s.collectConfigGroupAuthContext(ctx, reqs, authtypes.Modify, authtypes.UpdateConfigFileGroup)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	resp := s.nextServer.UpdateConfigFileGroups(ctx, reqs)

	return resp
}

// DeleteConfigFileGroups 删除配置文件组
func (s *Server) DeleteConfigFileGroups(ctx context.Context, reqs []*apiconfig.ConfigFileGroup) *apimodel.BatchWriteResponse {
	authCtx := s.collectConfigGroupAuthContext(ctx, reqs, authtypes.Delete, authtypes.DeleteConfigFileGroup)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	resp := s.nextServer.DeleteConfigFileGroups(ctx, reqs)
	return resp
}

// QueryConfigFileGroups 查询配置文件组
func (s *Server) QueryConfigFileGroups(ctx context.Context,
	filter map[string]string) *apimodel.BatchQueryResponse {
	authCtx := s.collectConfigGroupAuthContext(ctx, nil, authtypes.Read, authtypes.DescribeConfigFileGroups)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendConfigGroupPredicate(ctx, func(ctx context.Context, cfg *conftypes.ConfigFileGroup) bool {
		ok := s.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     apisecurity.ResourceType_ConfigGroups,
			ID:       cfg.Id,
			Metadata: cfg.Metadata,
		})
		if ok {
			return true
		}
		saveNs := s.cacheMgr.Namespace().GetNamespace(cfg.Namespace)
		if saveNs == nil {
			return false
		}
		// 检查下是否可以访问对应的 namespace
		return s.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     security.ResourceType_Namespaces,
			ID:       saveNs.Name,
			Metadata: saveNs.Metadata,
		})
	})
	authCtx.SetRequestContext(ctx)

	resp := s.nextServer.QueryConfigFileGroups(ctx, filter)
	return s.filterConfigGroupQuery(ctx, authCtx, resp)
}
