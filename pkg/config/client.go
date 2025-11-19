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
	"encoding/base64"
	"time"

	"go.uber.org/zap"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/crypto"
	"github.com/pole-io/pole-server/apis/pkg/types"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

type (
	CompareFunction func(clientInfo *apiconfig.ConfigFile, file *conftypes.ConfigFileRelease) bool
)

// GetConfigFileWithCache 从缓存中获取配置文件，如果客户端的版本号大于服务端，则服务端重新加载缓存
func (s *Server) GetConfigFileWithCache(ctx context.Context, req *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse {
	namespace := req.Namespace
	group := req.Group
	fileName := req.Name

	req = formatClientRequest(ctx, req)
	// 从缓存中获取灰度文件
	var release *conftypes.ConfigFileRelease
	var match = false
	if release = s.fileCache.GetActiveGrayRelease(namespace, group, fileName); release != nil {
		key := GetGrayConfigReaseKey(release.SimpleConfigFileRelease)
		// 将客户端标签转换为灰度匹配需要的标签格式
		clientLabels := make(map[string]string)
		if req.Tags != nil {
			for k, v := range req.Tags {
				clientLabels[k] = v
			}
		}
		match = s.grayCache.HitGrayRule(key, clientLabels)
	}
	if !match {
		if release = s.fileCache.GetActiveRelease(namespace, group, fileName); release == nil {
			return &apiconfig.ConfigDiscoverResponse{
				Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
				Info: "NotFoundResource",
			}
		}
	}
	// 客户端版本号大于服务端版本号，服务端不返回变更
	if req.Id > 0 && release.Version > 0 {
		log.Debug("[Config][Service] get config file to client", utils.RequestID(ctx),
				zap.Uint64("client-version", req.Id), zap.Uint64("server-version", release.Version))
		return &apiconfig.ConfigDiscoverResponse{
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
			Info: "DataNoChange",
		}
	}
	configFile, err := toClientInfo(req, release)
	if err != nil {
		log.Error("[Config][Service] get config file to client", utils.RequestID(ctx), zap.Error(err))
		return &apiconfig.ConfigDiscoverResponse{
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
			Info: err.Error(),
		}
	}
	// 将 configFile 设置到响应中
	response := &apiconfig.ConfigDiscoverResponse{
		Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		Info: "ExecuteSuccess",
	}
	_ = configFile // 使用 configFile 避免 unused 错误
	return response
}

// formatClientRequest 自动填充客户端的相关标签数据
func formatClientRequest(ctx context.Context, client *apiconfig.ConfigFile) *apiconfig.ConfigFile {
	// 添加客户端IP标签
	clientIP := utils.ParseClientIP(ctx)
	if client.Tags == nil {
		client.Tags = make(map[string]string)
	}
	client.Tags[types.ClientLabel_IP] = clientIP
	return client
}

// LongPullWatchFile .
func (s *Server) LongPullWatchFile(ctx context.Context,
	req *apiconfig.ConfigFileGroupRequest) (WatchCallback, error) {
	// 构建 watchFiles，使用 ConfigFile 类型
	watchFiles := []*apiconfig.ConfigFile{
		{
			Namespace: req.GetConfigFileGroup().Namespace,
			Group:     req.GetConfigFileGroup().Name,
			Name:      req.GetConfigFileGroup().Name, // 使用组名作为文件名
		},
	}

	tmpWatchCtx := BuildTimeoutWatchCtx(ctx, req, 0)("", s.watchCenter.MatchBetaReleaseFile)
	for _, file := range watchFiles {
		tmpWatchCtx.AppendInterest(file)
	}
	if quickResp := s.watchCenter.CheckQuickResponseClient(tmpWatchCtx); quickResp != nil {
		_ = tmpWatchCtx.Close()
		return func() *apiconfig.ConfigDiscoverResponse {
			return quickResp
		}, nil
	}

	watchTimeOut := defaultLongPollingTimeout
	if timeoutVal, ok := ctx.Value(utils.WatchTimeoutCtx{}).(time.Duration); ok {
		watchTimeOut = timeoutVal
	}

	// 3. 监听配置变更，hold 请求 30s，30s 内如果有配置发布，则响应请求
	clientId := utils.ParseClientAddress(ctx) + "@" + utils.NewUUID()[0:8]
	watchCtx := s.WatchCenter().AddWatcher(clientId, watchFiles, BuildTimeoutWatchCtx(ctx, req, watchTimeOut))
	return func() *apiconfig.ConfigDiscoverResponse {
		return (watchCtx.(*LongPollWatchContext)).GetNotifieResult()
	}, nil
}

func BuildTimeoutWatchCtx(ctx context.Context, req *apiconfig.ConfigFileGroupRequest,
	watchTimeOut time.Duration) WatchContextFactory {
	labels := map[string]string{
		types.ClientLabel_IP: utils.ParseClientIP(ctx),
	}
	return func(clientId string, matcher BetaReleaseMatcher) WatchContext {
		watchCtx := &LongPollWatchContext{
			clientId:         clientId,
			labels:           labels,
			finishTime:       time.Now().Add(watchTimeOut),
			finishChan:       make(chan *apiconfig.ConfigDiscoverResponse, 1),
			watchConfigFiles: map[string]*apiconfig.ConfigFile{},
			betaMatcher:      matcher,
		}
		return watchCtx
	}
}

// GetConfigFileNamesWithCache
func (s *Server) GetConfigFileNamesWithCache(ctx context.Context,
	req *apiconfig.ConfigFileGroupRequest) *apiconfig.ConfigDiscoverResponse {

	namespace := req.GetConfigFileGroup().GetNamespace()
	group := req.GetConfigFileGroup().GetName()

	releases, revision := s.fileCache.GetGroupActiveReleases(namespace, group)
	if revision == "" {
		return &apiconfig.ConfigDiscoverResponse{
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE_Names,
			Info: "ExecuteSuccess",
		}
	}
	if revision == req.GetRevision() {
		return &apiconfig.ConfigDiscoverResponse{
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE_Names,
			Info: "DataNoChange",
		}
	}
	ret := make([]*apiconfig.ConfigFile, 0, len(releases))
	for i := range releases {
		ret = append(ret, &apiconfig.ConfigFile{
			Namespace:   releases[i].Namespace,
			Group:       releases[i].Group,
			Name:        releases[i].Name,
			Id:     	 releases[i].Version,
		})
	}

	return &apiconfig.ConfigDiscoverResponse{
		Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE_Names,
		Info: "ExecuteSuccess",
	}
}

func (s *Server) GetConfigGroupsWithCache(ctx context.Context, req *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse {
	namespace := req.Namespace
	out := api.NewConfigDiscoverResponse(apimodel.Code_ExecuteSuccess)

	groups, revision := s.groupCache.ListGroups(namespace)
	if revision == "" {
		out = api.NewConfigDiscoverResponse(apimodel.Code_ExecuteSuccess)
		out.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE_GROUPS
		return out
	}
	if revision == req.Mtime {
		out = api.NewConfigDiscoverResponse(apimodel.Code_DataNoChange)
		out.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE_GROUPS
		return out
	}

	ret := make([]*apiconfig.ConfigFileGroup, 0, len(groups))
	for i := range groups {
		item := groups[i]
		ret = append(ret, &apiconfig.ConfigFileGroup{
			Namespace: item.Namespace,
			Name:      item.Name,
		})
	}

	out.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE_GROUPS
	return out
}

func toClientInfo(client *apiconfig.ConfigFile,
	release *conftypes.ConfigFileRelease) (*apiconfig.ConfigFile, error) {

	namespace := client.Namespace
	group := client.Group
	fileName := client.Name
	publicKey := ""

	configFile := &apiconfig.ConfigFile{
		Namespace: namespace,
		Group:     group,
		Name:      fileName,
		Content:   release.Content,
		Id:        release.Version,
		Encrypted: release.IsEncrypted(),
	}

	dataKey := release.GetEncryptDataKey()
	encryptAlgo := release.GetEncryptAlgo()
	if dataKey != "" && encryptAlgo != "" {
		dataKeyBytes, err := base64.StdEncoding.DecodeString(dataKey)
		if err != nil {
			log.Error("[config][client] decode data key error.", zap.String("dataKey", dataKey), zap.Error(err))
			return nil, err
		}
		if publicKey != "" {
			rsacrypto, err := crypto.GetCryptoManager().GetCrypto("rsa")
			if err != nil {
				log.Error("[config][client] get rsa crypto fail", zap.Error(err))
				return nil, err
			}
			cipherDataKey, err := rsacrypto.Encrypt(string(dataKeyBytes), []byte(publicKey))
			if err != nil {
				log.Error("[config][client] rsa encrypt data key error.",
					zap.String("dataKey", dataKey), zap.Error(err))
			} else {
				dataKey = cipherDataKey
			}
		}
		// 设置加密相关的标签
		if configFile.Tags == nil {
			configFile.Tags = make(map[string]string)
		}
		configFile.Tags[types.MetaKeyConfigFileDataKey] = dataKey
		configFile.Tags[types.MetaKeyConfigFileEncryptAlgo] = encryptAlgo
	}
	return configFile, nil
}

// UpsertAndReleaseConfigFile 创建/更新配置文件并发布
func (s *Server) UpsertAndReleaseConfigFileFromClient(ctx context.Context,
	req *apiconfig.ConfigFilePublishInfo) *apimodel.Response {
	return s.UpsertAndReleaseConfigFile(ctx, req)
}

// DeleteConfigFileFromClient 调用config_file的方法更新配置文件
func (s *Server) DeleteConfigFileFromClient(ctx context.Context, req *apiconfig.ConfigFile) *apimodel.Response {
	return s.DeleteConfigFile(ctx, req)
}

// CreateConfigFileFromClient 调用config_file接口获取配置文件
func (s *Server) CreateConfigFileFromClient(ctx context.Context,
	client *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse {
	configResponse := s.CreateConfigFile(ctx, client)
	return &apiconfig.ConfigDiscoverResponse{
		Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		Info: configResponse.GetInfo(),
	}
}

// UpdateConfigFileFromClient 调用config_file接口更新配置文件
func (s *Server) UpdateConfigFileFromClient(ctx context.Context,
	client *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse {
	configResponse := s.UpdateConfigFile(ctx, client)
	return &apiconfig.ConfigDiscoverResponse{
		Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		Info: configResponse.GetInfo(),
	}
}

// PublishConfigFileFromClient 调用config_file_release接口发布配置文件
func (s *Server) PublishConfigFileFromClient(ctx context.Context,
	client *apiconfig.ConfigFileRelease) *apiconfig.ConfigDiscoverResponse {
	configResponse := s.PublishConfigFile(ctx, client)
	return &apiconfig.ConfigDiscoverResponse{
		Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		Info: configResponse.GetInfo(),
	}
}

// GetConfigSubscribers 根据配置视角获取订阅者列表
func (s *Server) GetConfigSubscribers(ctx context.Context, filter map[string]string) *types.CommonResponse {
	namespace := filter["namespace"]
	group := filter["group"]
	filename := filter["file_name"]

	key := GenFileId(namespace, group, filename)
	clientIds, _ := s.watchCenter.watchers.Load(key)
	if clientIds == nil {
		rsp := types.NewCommonResponse(uint32(apimodel.Code_ExecuteSuccess))
		rsp.Data = &conftypes.ConfigSubscribers{}
		return rsp
	}

	versionClients := map[uint64][]*conftypes.Subscriber{}
	clientIds.Range(func(val string) {
		watchCtx, ok := s.watchCenter.clients.Load(val)
		if !ok {
			return
		}
		curVer := watchCtx.CurWatchVersion(key)
		if _, ok := versionClients[curVer]; !ok {
			versionClients[curVer] = []*conftypes.Subscriber{}
		}
		versionClients[curVer] = append(versionClients[curVer], &conftypes.Subscriber{
			ID: watchCtx.ClientID(),
			ReleaseName: s.fileCache.GetRelease(conftypes.ConfigFileReleaseKey{
				Namespace: namespace,
				Group:     group,
				Name:      filename,
				Version:   uint64(curVer),
			}).Name,
			Host:       watchCtx.ClientLabels()[types.ClientLabel_Host],
			Version:    watchCtx.ClientLabels()[types.ClientLabel_Version],
			ClientType: watchCtx.ClientLabels()[types.ClientLabel_Language],
		})
	})

	rsp := types.NewCommonResponse(uint32(apimodel.Code_ExecuteSuccess))
	rsp.Data = &conftypes.ConfigSubscribers{
		Key: conftypes.ConfigFileKey{
			Namespace: namespace,
			Group:     group,
			Name:      filename,
		},
		VersionClients: func() []*conftypes.VersionClient {
			ret := make([]*conftypes.VersionClient, 0, len(versionClients))
			for ver, clients := range versionClients {
				ret = append(ret, &conftypes.VersionClient{
					Version:     ver,
					Subscribers: clients,
				})
			}
			return ret
		}(),
	}
	return rsp
}

// GetClientSubscribers 根据客户端视角获取订阅的配置文件列表
func (s *Server) GetClientSubscribers(ctx context.Context, filter map[string]string) *types.CommonResponse {
	clientId := filter["client_id"]
	watchCtx, ok := s.watchCenter.clients.Load(clientId)
	if !ok {
		return types.NewCommonResponse(uint32(apimodel.Code_NotFoundResource))
	}

	watchFiles := watchCtx.ListWatchFiles()
	data := &conftypes.ClientSubscriber{
		Subscriber: conftypes.Subscriber{
			ID:         watchCtx.ClientID(),
			Host:       watchCtx.ClientLabels()[types.ClientLabel_Host],
			Version:    watchCtx.ClientLabels()[types.ClientLabel_Version],
			ClientType: watchCtx.ClientLabels()[types.ClientLabel_Language],
		},
		Files: []conftypes.FileReleaseSubscribeInfo{},
	}

	for _, file := range watchFiles {
		key := GenFileId(file.GetNamespace(), file.GetGroup(), file.GetName())
		curVer := watchCtx.CurWatchVersion(key)

		ns := file.GetNamespace()
		group := file.GetGroup()
		filename := file.GetName()

		data.Files = append(data.Files, conftypes.FileReleaseSubscribeInfo{
			Name:      file.GetName(),
			Namespace: ns,
			Group:     group,
			FileName:  filename,
			ReleaseType: func() rules.ReleaseType {
				if gray := s.fileCache.GetActiveGrayRelease(ns, group, filename); gray != nil {
					if gray.Version == curVer {
						return conftypes.ReleaseTypeGray
					}
				}
				return conftypes.ReleaseTypeNormal
			}(),
			Version: curVer,
			ReleaseName: s.fileCache.GetRelease(conftypes.ConfigFileReleaseKey{
				Namespace: ns,
				Group:     group,
				Name:      filename,
				Version:   uint64(curVer),
			}).Name,
		})
	}

	rsp := types.NewCommonResponse(uint32(apimodel.Code_ExecuteSuccess))
	rsp.Data = data
	return rsp
}
