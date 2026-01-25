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

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
	"github.com/pole-io/pole-server/pkg/config"
)

// UpsertAndReleaseConfigFileFromClient 创建/更新配置文件并发布
func (s *Server) UpsertAndReleaseConfigFileFromClient(ctx context.Context,
	req *apiconfig.ConfigFilePublishInfo) *apimodel.Response {
	if err := valid.CheckResourceName(req.GetNamespace()); err != nil {
		return api.NewConfigResponseWithInfo(apimodel.Code_BadRequest, "invalid config namespace")
	}
	if err := valid.CheckResourceName(req.GetGroup()); err != nil {
		return api.NewConfigResponseWithInfo(apimodel.Code_BadRequest, "invalid config group")
	}
	if err := CheckFileName(req.GetFileName()); err != nil {
		return api.NewConfigResponseWithInfo(apimodel.Code_BadRequest, "invalid config file_name")
	}
	return s.nextServer.UpsertAndReleaseConfigFileFromClient(ctx, req)
}

// CreateConfigFileFromClient 调用config_file的方法创建配置文件
func (s *Server) CreateConfigFileFromClient(ctx context.Context,
	req *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse {
	if checkRsp := s.checkConfigFileParams(req); checkRsp != nil {
		return &apiconfig.ConfigDiscoverResponse{
			Code: checkRsp.Code,
			Info: checkRsp.Info,
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		}
	}
	return s.nextServer.CreateConfigFileFromClient(ctx, req)
}

// UpdateConfigFileFromClient 调用config_file的方法更新配置文件
func (s *Server) UpdateConfigFileFromClient(ctx context.Context,
	req *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse {
	if checkRsp := s.checkConfigFileParams(req); checkRsp != nil {
		return &apiconfig.ConfigDiscoverResponse{
			Code: checkRsp.Code,
			Info: checkRsp.Info,
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		}
	}
	return s.nextServer.UpdateConfigFileFromClient(ctx, req)
}

// DeleteConfigFileFromClient 删除配置文件，删除配置文件同时会通知客户端 Not_Found
func (s *Server) DeleteConfigFileFromClient(ctx context.Context,
	req *apiconfig.ConfigFile) *apimodel.Response {

	if req.GetNamespace() == "" {
		return api.NewConfigResponseWithInfo(
			apimodel.Code_BadRequest, "namespace is empty")
	}

	if req.GetGroup() == "" {
		return api.NewConfigResponseWithInfo(
			apimodel.Code_BadRequest, "file group is empty")
	}

	if req.GetName() == "" {
		return api.NewConfigResponseWithInfo(
			apimodel.Code_BadRequest, "filename is empty")
	}

	return s.nextServer.DeleteConfigFileFromClient(ctx, req)
}

// PublishConfigFileFromClient 调用config_file_release的方法发布配置文件
func (s *Server) PublishConfigFileFromClient(ctx context.Context,
	req *apiconfig.ConfigFileRelease) *apiconfig.ConfigDiscoverResponse {

	if err := CheckFileName(req.GetFileName()); err != nil {
		ret := api.NewConfigResponse(apimodel.Code_InvalidParameter)
		return &apiconfig.ConfigDiscoverResponse{
			Code: ret.Code,
			Info: ret.Info,
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		}
	}
	if err := valid.CheckResourceName(req.GetNamespace()); err != nil {
		ret := api.NewConfigResponse(apimodel.Code_InvalidParameter)
		return &apiconfig.ConfigDiscoverResponse{
			Code: ret.Code,
			Info: ret.Info,
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		}
	}
	if err := valid.CheckResourceName(req.GetGroup()); err != nil {
		ret := api.NewConfigResponse(apimodel.Code_InvalidParameter)
		return &apiconfig.ConfigDiscoverResponse{
			Code: ret.Code,
			Info: ret.Info,
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		}
	}
	if !s.checkNamespaceExisted(req.GetNamespace()) {
		ret := api.NewConfigResponse(apimodel.Code_NotFoundResource)
		return &apiconfig.ConfigDiscoverResponse{
			Code: ret.Code,
			Info: ret.Info,
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		}
	}
	if req.GetReleaseType() == conftypes.ReleaseTypeGray && len(req.GetBetaLabels()) == 0 {
		ret := api.NewConfigResponse(apimodel.Code_InvalidMatchRule)
		return &apiconfig.ConfigDiscoverResponse{
			Code: ret.Code,
			Info: ret.Info,
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		}
	}

	return s.nextServer.PublishConfigFileFromClient(ctx, req)
}

// GetConfigFileWithCache 从缓存中获取配置文件，如果客户端的版本号大于服务端，则服务端重新加载缓存
func (s *Server) GetConfigFileWithCache(ctx context.Context,
	req *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse {

	if req.GetNamespace() == "" {
		return &apiconfig.ConfigDiscoverResponse{
			Code: uint32(apimodel.Code_BadRequest),
			Info: "namespace is empty",
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		}
	}

	if req.GetGroup() == "" {
		return &apiconfig.ConfigDiscoverResponse{
			Code: uint32(apimodel.Code_BadRequest),
			Info: "file group is empty",
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		}
	}

	if req.GetName() == "" {
		return &apiconfig.ConfigDiscoverResponse{
			Code: uint32(apimodel.Code_BadRequest),
			Info: "filename is empty",
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		}
	}
	return s.nextServer.GetConfigFileWithCache(ctx, req)
}

// WatchConfigFiles 监听配置文件变化 (gRPC 版本) - 但目前简化跳过实现
// func (s *Server) WatchConfigFiles(ctx context.Context,
//	request *apiconfig.ClientWatchConfigFileRequest) (*apiconfig.ConfigClientResponse, error) {
//	// 暂时跳过此方法的实现，因为接口不匹配
//	return s.nextServer.WatchConfigFiles(ctx, request)
//}

// LongPullWatchFile 监听配置文件变化 (HTTP 版本)
func (s *Server) LongPullWatchFile(ctx context.Context,
	request *apiconfig.WatchConfigFileRequest) (config.WatchCallback, error) {

	if request.GetConfigFileGroup().GetNamespace() == "" {
		return func() *apiconfig.ConfigDiscoverResponse {
			return &apiconfig.ConfigDiscoverResponse{
				Code: uint32(apimodel.Code_BadRequest),
				Info: "namespace is empty",
				Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
			}
		}, nil
	}

	if request.GetConfigFileGroup().GetName() == "" {
		return func() *apiconfig.ConfigDiscoverResponse {
			return &apiconfig.ConfigDiscoverResponse{
				Code: uint32(apimodel.Code_BadRequest),
				Info: "file group is empty",
				Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
			}
		}, nil
	}

	return s.nextServer.LongPullWatchFile(ctx, request)
}

// GetConfigFileNamesWithCache 获取某个配置分组下的配置文件
func (s *Server) GetConfigFileNamesWithCache(ctx context.Context,
	req *apiconfig.ConfigFileGroupRequest) *apiconfig.ConfigDiscoverResponse {

	if req.GetConfigFileGroup().GetNamespace() == "" {
		return &apiconfig.ConfigDiscoverResponse{
			Code: uint32(apimodel.Code_BadRequest),
			Info: "namespace is empty",
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE_NAMES,
		}
	}

	if req.GetConfigFileGroup().GetName() == "" {
		return &apiconfig.ConfigDiscoverResponse{
			Code: uint32(apimodel.Code_BadRequest),
			Info: "file group is empty",
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE_NAMES,
		}
	}

	return s.nextServer.GetConfigFileNamesWithCache(ctx, req)
}

func (s *Server) GetConfigGroupsWithCache(ctx context.Context,
	req *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse {

	namespace := req.GetNamespace()
	out := api.NewConfigDiscoverResponse(apimodel.Code_ExecuteSuccess)
	if namespace == "" {
		out.Code = uint32(apimodel.Code_BadRequest)
		out.Info = "invalid namespace"
		return out
	}

	return s.nextServer.GetConfigGroupsWithCache(ctx, req)
}
