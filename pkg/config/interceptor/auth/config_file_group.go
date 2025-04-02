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
	"strconv"

	"go.uber.org/zap"

	apiconfig "github.com/polarismesh/specification/source/go/api/v1/config_manage"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	"github.com/polarismesh/specification/source/go/api/v1/security"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// CreateConfigFileGroup 创建配置文件组
func (s *Server) CreateConfigFileGroup(ctx context.Context,
	configFileGroup *apiconfig.ConfigFileGroup) *apiconfig.ConfigResponse {
	authCtx := s.collectConfigGroupAuthContext(ctx, []*apiconfig.ConfigFileGroup{configFileGroup},
		authtypes.Create, authtypes.CreateConfigFileGroup)

	// 验证 token 信息
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	resp := s.nextServer.CreateConfigFileGroup(ctx, configFileGroup)
	if err := s.afterConfigGroupResource(ctx, resp.GetConfigFileGroup(), false); err != nil {
		log.Error("[Config][Group] create config_file_group after resource",
			utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(apimodel.Code_ExecuteException)
	}
	return resp
}

// UpdateConfigFileGroup 更新配置文件组
func (s *Server) UpdateConfigFileGroup(ctx context.Context,
	configFileGroup *apiconfig.ConfigFileGroup) *apiconfig.ConfigResponse {
	authCtx := s.collectConfigGroupAuthContext(ctx, []*apiconfig.ConfigFileGroup{configFileGroup},
		authtypes.Modify, authtypes.UpdateConfigFileGroup)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)
	resp := s.nextServer.UpdateConfigFileGroup(ctx, configFileGroup)
	if err := s.afterConfigGroupResource(ctx, resp.GetConfigFileGroup(), false); err != nil {
		log.Error("[Config][Group] update config_file_group after resource",
			utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(apimodel.Code_ExecuteException)
	}
	return resp
}

// DeleteConfigFileGroup 删除配置文件组
func (s *Server) DeleteConfigFileGroup(
	ctx context.Context, namespace, name string) *apiconfig.ConfigResponse {
	authCtx := s.collectConfigGroupAuthContext(ctx, []*apiconfig.ConfigFileGroup{{Name: utils.NewStringValue(name),
		Namespace: utils.NewStringValue(namespace)}}, authtypes.Delete, authtypes.DeleteConfigFileGroup)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	resp := s.nextServer.DeleteConfigFileGroup(ctx, namespace, name)
	if err := s.afterConfigGroupResource(ctx, resp.GetConfigFileGroup(), true); err != nil {
		log.Error("[Config][Group] delete config_file_group after resource",
			utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(apimodel.Code_ExecuteException)
	}
	return resp
}

// QueryConfigFileGroups 查询配置文件组
func (s *Server) QueryConfigFileGroups(ctx context.Context,
	filter map[string]string) *apiconfig.ConfigBatchQueryResponse {
	authCtx := s.collectConfigGroupAuthContext(ctx, nil, authtypes.Read, authtypes.DescribeConfigFileGroups)

	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewConfigBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	ctx = cacheapi.AppendConfigGroupPredicate(ctx, func(ctx context.Context, cfg *conftypes.ConfigFileGroup) bool {
		ok := s.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     apisecurity.ResourceType_ConfigGroups,
			ID:       strconv.FormatUint(cfg.Id, 10),
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
	if len(resp.ConfigFileGroups) != 0 {
		for index := range resp.ConfigFileGroups {
			item := resp.ConfigFileGroups[index]
			authCtx.SetAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
				apisecurity.ResourceType_ConfigGroups: {
					{
						Type:     apisecurity.ResourceType_ConfigGroups,
						ID:       strconv.FormatUint(item.GetId().GetValue(), 10),
						Metadata: item.Metadata,
					},
				},
			})

			// 检查 write 操作权限
			authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.UpdateConfigFileGroup})
			// 如果检查不通过，设置 editable 为 false
			if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
				item.Editable = utils.NewBoolValue(false)
			}

			// 检查 delete 操作权限
			authCtx.SetMethod([]authtypes.ServerFunctionName{authtypes.DeleteConfigFileGroup})
			// 如果检查不通过，设置 editable 为 false
			if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
				item.Deleteable = utils.NewBoolValue(false)
			}
		}
	}
	return resp
}
