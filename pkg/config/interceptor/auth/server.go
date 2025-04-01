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

	apiconfig "github.com/polarismesh/specification/source/go/api/v1/config_manage"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
	"go.uber.org/zap"

	"github.com/pole-io/pole-server/apis/access_control/auth"
	authcommon "github.com/pole-io/pole-server/apis/pkg/types/auth"
	cachetypes "github.com/pole-io/pole-server/pkg/cache/api"
	"github.com/pole-io/pole-server/pkg/common/model"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/config"
)

var _ config.ConfigCenterServer = (*Server)(nil)

// Server 配置中心核心服务
type Server struct {
	cacheMgr   cachetypes.CacheManager
	nextServer config.ConfigCenterServer
	userSvr    auth.UserServer
	policySvr  auth.StrategyServer
}

func New(nextServer config.ConfigCenterServer, cacheMgr cachetypes.CacheManager,
	userSvr auth.UserServer, policySvr auth.StrategyServer) config.ConfigCenterServer {
	proxy := &Server{
		nextServer: nextServer,
		cacheMgr:   cacheMgr,
		userSvr:    userSvr,
		policySvr:  policySvr,
	}
	return proxy
}

func (s *Server) collectConfigFileAuthContext(ctx context.Context, req []*apiconfig.ConfigFile,
	op authcommon.ResourceOperation, methodName authcommon.ServerFunctionName) *authcommon.AcquireContext {
	return authcommon.NewAcquireContext(
		authcommon.WithRequestContext(ctx),
		authcommon.WithModule(authcommon.ConfigModule),
		authcommon.WithOperation(op),
		authcommon.WithMethod(methodName),
		authcommon.WithAccessResources(s.queryConfigFileResource(ctx, req)),
	)
}

func (s *Server) collectClientConfigFileAuthContext(ctx context.Context, req []*apiconfig.ConfigFile,
	op authcommon.ResourceOperation, methodName authcommon.ServerFunctionName) *authcommon.AcquireContext {
	return authcommon.NewAcquireContext(
		authcommon.WithRequestContext(ctx),
		authcommon.WithModule(authcommon.ConfigModule),
		authcommon.WithOperation(op),
		authcommon.WithMethod(methodName),
		authcommon.WithFromClient(),
		authcommon.WithAccessResources(s.queryConfigFileResource(ctx, req)),
	)
}

func (s *Server) collectClientWatchConfigFiles(ctx context.Context, req *apiconfig.ClientWatchConfigFileRequest,
	op authcommon.ResourceOperation, methodName authcommon.ServerFunctionName) *authcommon.AcquireContext {
	return authcommon.NewAcquireContext(
		authcommon.WithRequestContext(ctx),
		authcommon.WithModule(authcommon.ConfigModule),
		authcommon.WithOperation(op),
		authcommon.WithMethod(methodName),
		authcommon.WithFromClient(),
		authcommon.WithAccessResources(s.queryWatchConfigFilesResource(ctx, req)),
	)
}

func (s *Server) collectConfigFileReleaseAuthContext(ctx context.Context, req []*apiconfig.ConfigFileRelease,
	op authcommon.ResourceOperation, methodName authcommon.ServerFunctionName) *authcommon.AcquireContext {
	return authcommon.NewAcquireContext(
		authcommon.WithRequestContext(ctx),
		authcommon.WithModule(authcommon.ConfigModule),
		authcommon.WithOperation(op),
		authcommon.WithMethod(methodName),
		authcommon.WithAccessResources(s.queryConfigFileReleaseResource(ctx, req)),
	)
}

func (s *Server) collectConfigFilePublishAuthContext(ctx context.Context, req []*apiconfig.ConfigFilePublishInfo,
	op authcommon.ResourceOperation, methodName authcommon.ServerFunctionName) *authcommon.AcquireContext {
	return authcommon.NewAcquireContext(
		authcommon.WithRequestContext(ctx),
		authcommon.WithModule(authcommon.ConfigModule),
		authcommon.WithOperation(op),
		authcommon.WithMethod(methodName),
		authcommon.WithAccessResources(s.queryConfigFilePublishResource(ctx, req)),
	)
}

func (s *Server) collectClientConfigFileRelease(ctx context.Context, req []*apiconfig.ConfigFileRelease,
	op authcommon.ResourceOperation, methodName authcommon.ServerFunctionName) *authcommon.AcquireContext {
	return authcommon.NewAcquireContext(
		authcommon.WithRequestContext(ctx),
		authcommon.WithModule(authcommon.ConfigModule),
		authcommon.WithOperation(op),
		authcommon.WithMethod(methodName),
		authcommon.WithFromClient(),
		authcommon.WithAccessResources(s.queryConfigFileReleaseResource(ctx, req)),
	)
}

func (s *Server) collectConfigFileReleaseHistoryAuthContext(
	ctx context.Context,
	req []*apiconfig.ConfigFileReleaseHistory,
	op authcommon.ResourceOperation, methodName authcommon.ServerFunctionName) *authcommon.AcquireContext {
	return authcommon.NewAcquireContext(
		authcommon.WithRequestContext(ctx),
		authcommon.WithModule(authcommon.ConfigModule),
		authcommon.WithOperation(op),
		authcommon.WithMethod(methodName),
		authcommon.WithAccessResources(s.queryConfigFileReleaseHistoryResource(ctx, req)),
	)
}

func (s *Server) collectConfigGroupAuthContext(ctx context.Context, req []*apiconfig.ConfigFileGroup,
	op authcommon.ResourceOperation, methodName authcommon.ServerFunctionName) *authcommon.AcquireContext {
	return authcommon.NewAcquireContext(
		authcommon.WithRequestContext(ctx),
		authcommon.WithModule(authcommon.ConfigModule),
		authcommon.WithOperation(op),
		authcommon.WithMethod(methodName),
		authcommon.WithAccessResources(s.queryConfigGroupResource(ctx, req)),
	)
}

func (s *Server) collectConfigFileTemplateAuthContext(ctx context.Context, req []*apiconfig.ConfigFileTemplate,
	op authcommon.ResourceOperation, methodName authcommon.ServerFunctionName) *authcommon.AcquireContext {
	return authcommon.NewAcquireContext(
		authcommon.WithRequestContext(ctx),
		authcommon.WithModule(authcommon.ConfigModule),
	)
}

func (s *Server) queryConfigGroupResource(ctx context.Context,
	req []*apiconfig.ConfigFileGroup) map[apisecurity.ResourceType][]authcommon.ResourceEntry {

	if len(req) == 0 {
		return nil
	}

	names := utils.NewSet[string]()
	namespace := req[0].GetNamespace().GetValue()
	for index := range req {
		if req[index] == nil {
			continue
		}
		names.Add(req[index].GetName().GetValue())
	}
	entries, err := s.queryConfigGroupRsEntryByNames(ctx, namespace, names.ToSlice())
	if err != nil {
		authLog.Error("[Config][Server] collect config_file_group res",
			utils.RequestID(ctx), zap.Error(err))
		return nil
	}
	ret := map[apisecurity.ResourceType][]authcommon.ResourceEntry{
		apisecurity.ResourceType_ConfigGroups: entries,
	}
	authLog.Debug("[Config][Server] collect config_file_group access res",
		utils.RequestID(ctx), zap.Any("res", ret))
	return ret
}

// queryConfigFileResource config file资源的鉴权转换为config group的鉴权
func (s *Server) queryConfigFileResource(ctx context.Context,
	req []*apiconfig.ConfigFile) map[apisecurity.ResourceType][]authcommon.ResourceEntry {

	if len(req) == 0 {
		return nil
	}
	namespace := req[0].Namespace.GetValue()
	groupNames := utils.NewSet[string]()

	for _, apiConfigFile := range req {
		groupNames.Add(apiConfigFile.Group.GetValue())
	}
	entries, err := s.queryConfigGroupRsEntryByNames(ctx, namespace, groupNames.ToSlice())
	if err != nil {
		authLog.Error("[Config][Server] collect config_file res",
			utils.RequestID(ctx), zap.Error(err))
		return nil
	}
	ret := map[apisecurity.ResourceType][]authcommon.ResourceEntry{
		apisecurity.ResourceType_ConfigGroups: entries,
	}
	authLog.Debug("[Config][Server] collect config_file access res",
		utils.RequestID(ctx), zap.Any("res", ret))
	return ret
}

func (s *Server) queryConfigFileReleaseResource(ctx context.Context,
	req []*apiconfig.ConfigFileRelease) map[apisecurity.ResourceType][]authcommon.ResourceEntry {

	if len(req) == 0 {
		return nil
	}
	namespace := req[0].Namespace.GetValue()
	groupNames := utils.NewSet[string]()

	for _, apiConfigFile := range req {
		groupNames.Add(apiConfigFile.Group.GetValue())
	}
	entries, err := s.queryConfigGroupRsEntryByNames(ctx, namespace, groupNames.ToSlice())
	if err != nil {
		authLog.Debug("[Config][Server] collect config_file res",
			utils.RequestID(ctx), zap.Error(err))
		return nil
	}
	ret := map[apisecurity.ResourceType][]authcommon.ResourceEntry{
		apisecurity.ResourceType_ConfigGroups: entries,
	}
	authLog.Debug("[Config][Server] collect config_file access res",
		utils.RequestID(ctx), zap.Any("res", ret))
	return ret
}

func (s *Server) queryConfigFilePublishResource(ctx context.Context,
	req []*apiconfig.ConfigFilePublishInfo) map[apisecurity.ResourceType][]authcommon.ResourceEntry {

	if len(req) == 0 {
		return nil
	}
	namespace := req[0].GetNamespace().GetValue()
	groupNames := utils.NewSet[string]()

	for _, apiConfigFile := range req {
		groupNames.Add(apiConfigFile.GetGroup().GetValue())
	}
	entries, err := s.queryConfigGroupRsEntryByNames(ctx, namespace, groupNames.ToSlice())
	if err != nil {
		authLog.Debug("[Config][Server] collect config_file res", utils.RequestID(ctx), zap.Error(err))
		return nil
	}
	ret := map[apisecurity.ResourceType][]authcommon.ResourceEntry{
		apisecurity.ResourceType_ConfigGroups: entries,
	}
	authLog.Debug("[Config][Server] collect config_file access res", utils.RequestID(ctx), zap.Any("res", ret))
	return ret
}

func (s *Server) queryConfigFileReleaseHistoryResource(ctx context.Context,
	req []*apiconfig.ConfigFileReleaseHistory) map[apisecurity.ResourceType][]authcommon.ResourceEntry {

	if len(req) == 0 {
		return nil
	}
	namespace := req[0].Namespace.GetValue()
	groupNames := utils.NewSet[string]()

	for _, apiConfigFile := range req {
		groupNames.Add(apiConfigFile.Group.GetValue())
	}
	entries, err := s.queryConfigGroupRsEntryByNames(ctx, namespace, groupNames.ToSlice())
	if err != nil {
		authLog.Debug("[Config][Server] collect config_file res",
			utils.RequestID(ctx), zap.Error(err))
		return nil
	}
	ret := map[apisecurity.ResourceType][]authcommon.ResourceEntry{
		apisecurity.ResourceType_ConfigGroups: entries,
	}
	authLog.Debug("[Config][Server] collect config_file access res",
		utils.RequestID(ctx), zap.Any("res", ret))
	return ret
}

func (s *Server) queryConfigGroupRsEntryByNames(ctx context.Context, namespace string,
	names []string) ([]authcommon.ResourceEntry, error) {

	configFileGroups := make([]*model.ConfigFileGroup, 0, len(names))
	for i := range names {
		data := s.cacheMgr.ConfigGroup().GetGroupByName(namespace, names[i])
		if data == nil {
			continue
		}

		configFileGroups = append(configFileGroups, data)
	}

	entries := make([]authcommon.ResourceEntry, 0, len(configFileGroups))

	for index := range configFileGroups {
		group := configFileGroups[index]
		entries = append(entries, authcommon.ResourceEntry{
			ID:    strconv.FormatUint(group.Id, 10),
			Owner: group.Owner,
		})
	}
	return entries, nil
}

func (s *Server) queryWatchConfigFilesResource(ctx context.Context,
	req *apiconfig.ClientWatchConfigFileRequest) map[apisecurity.ResourceType][]authcommon.ResourceEntry {
	files := req.GetWatchFiles()
	if len(files) == 0 {
		return nil
	}
	temp := map[string]struct{}{}
	entries := make([]authcommon.ResourceEntry, 0, len(files))
	for _, apiConfigFile := range files {
		namespace := apiConfigFile.GetNamespace().GetValue()
		groupName := apiConfigFile.GetGroup().GetValue()
		key := namespace + "@@" + groupName
		if _, ok := temp[key]; ok {
			continue
		}
		temp[key] = struct{}{}
		data := s.cacheMgr.ConfigGroup().GetGroupByName(namespace, groupName)
		if data == nil {
			continue
		}
		entries = append(entries, authcommon.ResourceEntry{
			ID:    strconv.FormatUint(data.Id, 10),
			Owner: data.Owner,
		})
	}

	ret := map[apisecurity.ResourceType][]authcommon.ResourceEntry{
		apisecurity.ResourceType_ConfigGroups: entries,
	}
	authLog.Debug("[Config][Server] collect config_file watch access res",
		utils.RequestID(ctx), zap.Any("res", ret))
	return ret
}

// ResourceEvent 资源事件
type ResourceEvent struct {
	ConfigGroup   *apiconfig.ConfigFileGroup
	AddPrincipals []authcommon.Principal
	DelPrincipals []authcommon.Principal
	IsRemove      bool
}

func (s *Server) afterConfigGroupResource(ctx context.Context, req *apiconfig.ConfigFileGroup, isRemove bool) error {
	event := &ResourceEvent{
		ConfigGroup: req,
		IsRemove:    isRemove,
	}

	return s.After(ctx, model.RConfigGroup, event)
}
