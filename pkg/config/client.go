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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strconv"
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
	configtemplate "github.com/pole-io/pole-server/pkg/config/template"
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
	// 将客户端标签转换为灰度匹配需要的标签格式
	clientLabels := make(map[string]string)
	if req.Labels != nil {
		for k, v := range req.Labels {
			clientLabels[k] = v
		}
	}
	release = selectMatchedGrayRelease(s.fileCache.GetActiveGrayReleases(namespace, group, fileName),
		func(release *conftypes.SimpleConfigFileRelease) bool {
			return s.grayCache.HitGrayRule(GetGrayConfigReaseKey(release), clientLabels)
		})
	match = release != nil
	if !match {
		if release = s.fileCache.GetActiveRelease(namespace, group, fileName); release == nil {
			response := api.NewConfigDiscoverResponse(apimodel.Code_NotFoundResource)
			response.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE
			return response
		}
	}
	snapshotRevision := configReleaseSnapshotRevision(release)
	var renderSnapshot *apiconfig.RenderSnapshot
	releaseView := conftypes.ToConfiogFileReleaseApi(release)
	if releaseView.GetConfigType() == apiconfig.ConfigFileRelease_CONFIG_TEMPLATE {
		if !clientSupportsTemplateEngine(ctx, configtemplate.EnginePoleMustache, configtemplate.EngineVersionV1) {
			response := api.NewConfigDiscoverResponse(apimodel.Code_BadRequest)
			response.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE
			response.Info = "client does not declare support for pole-mustache v1"
			return response
		}
		var resolveErr error
		renderSnapshot, resolveErr = s.resolveTemplateSnapshot(
			ctx, release.ToFileKey(), releaseView.GetTemplateBinding(), clientLabels)
		if resolveErr != nil {
			log.Error("[Config][Service] resolve template snapshot", utils.RequestID(ctx), zap.Error(resolveErr))
			response := api.NewConfigDiscoverResponse(apimodel.Code_ExecuteException)
			response.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE
			response.Info = resolveErr.Error()
			return response
		}
		snapshotRevision = renderSnapshot.GetRevision()
	}
	if req.Id != "" && req.Id == snapshotRevision {
		log.Debug("[Config][Service] get config file to client", utils.RequestID(ctx),
			zap.String("client-version", req.Id), zap.Uint64("server-version", release.Version))
		response := api.NewConfigDiscoverResponse(apimodel.Code_DataNoChange)
		response.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE
		response.Revision = snapshotRevision
		return response
	}
	configFile, err := toClientInfo(req, release)
	if err != nil {
		log.Error("[Config][Service] get config file to client", utils.RequestID(ctx), zap.Error(err))
		response := api.NewConfigDiscoverResponse(apimodel.Code_ExecuteException)
		response.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE
		response.Info = err.Error()
		return response
	}
	// 将 configFile 设置到响应中
	response := api.NewConfigDiscoverResponse(apimodel.Code_ExecuteSuccess)
	response.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE
	response.File = configFile
	response.RenderSnapshot = renderSnapshot
	if renderSnapshot != nil {
		// Template source and Values are transported atomically in RenderSnapshot.
		// Leaving ordinary content populated would allow an old SDK to apply an
		// unrendered template as if it were a plain config file.
		response.File.Content = ""
		response.File.TemplateBinding = renderSnapshot.GetTemplateBinding()
	}
	response.Revision = snapshotRevision
	return response
}

func clientSupportsTemplateEngine(ctx context.Context, name, version string) bool {
	filter, _ := ctx.Value(types.ContextDiscoverFilter).(*apiconfig.ConfigDiscoverFilter)
	for _, engine := range filter.GetSupportedTemplateEngines() {
		if engine.GetName() == name && engine.GetVersion() == version {
			return true
		}
	}
	return false
}

func configReleaseSnapshotRevision(release *conftypes.ConfigFileRelease) string {
	if release == nil || release.SimpleConfigFileRelease == nil {
		return ""
	}
	hash := sha256.New()
	for _, part := range []string{
		release.Id,
		release.Name,
		string(release.ReleaseType),
		strconv.FormatUint(release.Version, 10),
		release.Md5,
	} {
		_, _ = hash.Write([]byte(part))
		_, _ = hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// formatClientRequest 自动填充客户端的相关标签数据
func formatClientRequest(ctx context.Context, client *apiconfig.ConfigFile) *apiconfig.ConfigFile {
	// 添加客户端IP标签
	clientIP := utils.ParseClientIP(ctx)
	if client.Labels == nil {
		client.Labels = make(map[string]string)
	}
	client.Labels[types.ClientLabel_IP] = clientIP
	return client
}

// LongPullWatchFile .
func (s *Server) LongPullWatchFile(ctx context.Context,
	req *apiconfig.WatchConfigFileRequest) (WatchCallback, error) {
	// 构建 watchFiles，使用 ConfigFile 类型
	watchFiles := req.Files

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

func BuildTimeoutWatchCtx(ctx context.Context, req *apiconfig.WatchConfigFileRequest,
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
			watchConfigFiles: map[string]*apiconfig.ConfigFileRelease{},
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
		response := api.NewConfigDiscoverResponse(apimodel.Code_ExecuteSuccess)
		response.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE_NAMES
		return response
	}
	if revision == req.GetRevision() {
		response := api.NewConfigDiscoverResponse(apimodel.Code_DataNoChange)
		response.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE_NAMES
		return response
	}
	ret := make([]*apiconfig.ConfigFileRelease, 0, len(releases))
	for i := range releases {
		ret = append(ret, conftypes.ToConfiogFileReleaseApi(releases[i]))
	}

	response := api.NewConfigDiscoverResponse(apimodel.Code_ExecuteSuccess)
	response.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE_NAMES
	response.Revision = revision
	response.FileNames = ret
	return response
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
	out.Revision = revision
	out.FileGroups = ret
	return out
}

func toClientInfo(client *apiconfig.ConfigFile,
	release *conftypes.ConfigFileRelease) (*apiconfig.ConfigFileRelease, error) {

	namespace := client.Namespace
	group := client.Group
	fileName := client.Name
	publicKey := ""

	configFile := conftypes.ToConfiogFileReleaseApi(release)
	configFile.Namespace = namespace
	configFile.Group = group
	configFile.FileName = fileName
	configFile.Encrypted = release.IsEncrypted()
	configFile.EncryptAlgo = release.GetEncryptAlgo()

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
		if configFile.Labels == nil {
			configFile.Labels = make(map[string]string)
		}
		configFile.Labels[types.MetaKeyConfigFileDataKey] = dataKey
		configFile.Labels[types.MetaKeyConfigFileEncryptAlgo] = encryptAlgo
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
		Code: configResponse.GetCode(),
		Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		Info: configResponse.GetInfo(),
	}
}

// UpdateConfigFileFromClient 调用config_file接口更新配置文件
func (s *Server) UpdateConfigFileFromClient(ctx context.Context,
	client *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse {
	configResponse := s.UpdateConfigFile(ctx, client)
	return &apiconfig.ConfigDiscoverResponse{
		Code: configResponse.GetCode(),
		Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		Info: configResponse.GetInfo(),
	}
}

// PublishConfigFileFromClient 调用config_file_release接口发布配置文件
func (s *Server) PublishConfigFileFromClient(ctx context.Context,
	client *apiconfig.ConfigFileRelease) *apiconfig.ConfigDiscoverResponse {
	configResponse := s.PublishConfigFile(ctx, client)
	return &apiconfig.ConfigDiscoverResponse{
		Code: configResponse.GetCode(),
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
				for _, gray := range s.fileCache.GetActiveGrayReleases(ns, group, filename) {
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
