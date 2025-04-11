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
	"google.golang.org/protobuf/types/known/wrapperspb"

	apiconfig "github.com/polarismesh/specification/source/go/api/v1/config_manage"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/cmdb"
	"github.com/pole-io/pole-server/apis/crypto"
	"github.com/pole-io/pole-server/apis/pkg/types"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	commontime "github.com/pole-io/pole-server/pkg/common/time"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

type (
	CompareFunction func(clientInfo *apiconfig.ClientConfigFileInfo, file *conftypes.ConfigFileRelease) bool
)

// GetConfigFileWithCache 从缓存中获取配置文件，如果客户端的版本号大于服务端，则服务端重新加载缓存
func (s *Server) GetConfigFileWithCache(ctx context.Context,
	req *apiconfig.ClientConfigFileInfo) *apiconfig.ConfigClientResponse {
	namespace := req.GetNamespace().GetValue()
	group := req.GetGroup().GetValue()
	fileName := req.GetFileName().GetValue()

	req = formatClientRequest(ctx, req)
	// 从缓存中获取灰度文件
	var release *conftypes.ConfigFileRelease
	var match = false
	if len(req.GetTags()) > 0 {
		if release = s.fileCache.GetActiveGrayRelease(namespace, group, fileName); release != nil {
			key := utils.GetGrayConfigRealseKey(release.SimpleConfigFileRelease)
			match = s.grayCache.HitGrayRule(key, conftypes.ToTagMap(req.GetTags()))
		}
	}
	if !match {
		if release = s.fileCache.GetActiveRelease(namespace, group, fileName); release == nil {
			return api.NewConfigClientResponse(apimodel.Code_NotFoundResource, req)
		}
	}
	// 客户端版本号大于服务端版本号，服务端不返回变更
	if req.GetVersion().GetValue() > release.Version {
		log.Debug("[Config][Service] get config file to client", utils.RequestID(ctx),
			zap.Uint64("client-version", req.GetVersion().GetValue()), zap.Uint64("server-version", release.Version))
		return api.NewConfigClientResponse(apimodel.Code_DataNoChange, req)
	}
	configFile, err := toClientInfo(req, release)
	if err != nil {
		log.Error("[Config][Service] get config file to client", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigClientResponseWithInfo(apimodel.Code_ExecuteException, err.Error())
	}
	return api.NewConfigClientResponse(apimodel.Code_ExecuteSuccess, configFile)
}

// formatClientRequest 自动填充客户端的相关标签数据
func formatClientRequest(ctx context.Context, client *apiconfig.ClientConfigFileInfo) *apiconfig.ClientConfigFileInfo {
	clientIP := utils.ParseClientIP(ctx)
	newTags := []*apiconfig.ConfigFileTag{
		{
			Key:   wrapperspb.String(types.ClientLabel_IP),
			Value: wrapperspb.String(clientIP),
		},
	}
	loc, err := cmdb.GetCMDB().GetLocation(clientIP)
	if err == nil && loc != nil {
		newTags = append(newTags, &apiconfig.ConfigFileTag{
			Key:   wrapperspb.String(types.ClientLabel_Region),
			Value: wrapperspb.String(loc.Proto.GetRegion().GetValue()),
		})
		newTags = append(newTags, &apiconfig.ConfigFileTag{
			Key:   wrapperspb.String(types.ClientLabel_Zone),
			Value: wrapperspb.String(loc.Proto.GetZone().GetValue()),
		})
		newTags = append(newTags, &apiconfig.ConfigFileTag{
			Key:   wrapperspb.String(types.ClientLabel_Campus),
			Value: wrapperspb.String(loc.Proto.GetCampus().GetValue()),
		})
		for k, v := range loc.Labels {
			newTags = append(newTags, &apiconfig.ConfigFileTag{
				Key:   wrapperspb.String(k),
				Value: wrapperspb.String(v),
			})
		}
	}

	if len(client.Tags) == 0 {
		client.Tags = []*apiconfig.ConfigFileTag{}
	}
	client.Tags = append(newTags, client.Tags...)
	return client
}

// LongPullWatchFile .
func (s *Server) LongPullWatchFile(ctx context.Context,
	req *apiconfig.ClientWatchConfigFileRequest) (WatchCallback, error) {
	watchFiles := req.GetWatchFiles()

	tmpWatchCtx := BuildTimeoutWatchCtx(ctx, req, 0)("", s.watchCenter.MatchBetaReleaseFile)
	for _, file := range watchFiles {
		tmpWatchCtx.AppendInterest(file)
	}
	if quickResp := s.watchCenter.CheckQuickResponseClient(tmpWatchCtx); quickResp != nil {
		_ = tmpWatchCtx.Close()
		return func() *apiconfig.ConfigClientResponse {
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
	return func() *apiconfig.ConfigClientResponse {
		return (watchCtx.(*LongPollWatchContext)).GetNotifieResult()
	}, nil
}

func BuildTimeoutWatchCtx(ctx context.Context, req *apiconfig.ClientWatchConfigFileRequest,
	watchTimeOut time.Duration) WatchContextFactory {
	labels := map[string]string{
		types.ClientLabel_IP: utils.ParseClientIP(ctx),
	}
	if len(req.GetClientIp().GetValue()) != 0 {
		labels[types.ClientLabel_IP] = req.GetClientIp().GetValue()
	}
	return func(clientId string, matcher BetaReleaseMatcher) WatchContext {
		watchCtx := &LongPollWatchContext{
			clientId:         clientId,
			labels:           labels,
			finishTime:       time.Now().Add(watchTimeOut),
			finishChan:       make(chan *apiconfig.ConfigClientResponse, 1),
			watchConfigFiles: map[string]*apiconfig.ClientConfigFileInfo{},
			betaMatcher:      matcher,
		}
		return watchCtx
	}
}

// GetConfigFileNamesWithCache
func (s *Server) GetConfigFileNamesWithCache(ctx context.Context,
	req *apiconfig.ConfigFileGroupRequest) *apiconfig.ConfigClientListResponse {

	namespace := req.GetConfigFileGroup().GetNamespace().GetValue()
	group := req.GetConfigFileGroup().GetName().GetValue()

	releases, revision := s.fileCache.GetGroupActiveReleases(namespace, group)
	if revision == "" {
		return api.NewConfigClientListResponse(apimodel.Code_ExecuteSuccess)
	}
	if revision == req.GetRevision().GetValue() {
		return api.NewConfigClientListResponse(apimodel.Code_DataNoChange)
	}
	ret := make([]*apiconfig.ClientConfigFileInfo, 0, len(releases))
	for i := range releases {
		ret = append(ret, &apiconfig.ClientConfigFileInfo{
			Namespace:   protobuf.NewStringValue(releases[i].Namespace),
			Group:       protobuf.NewStringValue(releases[i].Group),
			FileName:    protobuf.NewStringValue(releases[i].FileName),
			Name:        protobuf.NewStringValue(releases[i].Name),
			Version:     protobuf.NewUInt64Value(releases[i].Version),
			ReleaseTime: protobuf.NewStringValue(commontime.Time2String(releases[i].ModifyTime)),
			Tags:        conftypes.FromTagMap(releases[i].Metadata),
		})
	}

	return &apiconfig.ConfigClientListResponse{
		Code:            protobuf.NewUInt32Value(uint32(apimodel.Code_ExecuteSuccess)),
		Info:            protobuf.NewStringValue(api.Code2Info(uint32(apimodel.Code_ExecuteSuccess))),
		Revision:        protobuf.NewStringValue(revision),
		Namespace:       namespace,
		Group:           group,
		ConfigFileInfos: ret,
	}
}

func (s *Server) GetConfigGroupsWithCache(ctx context.Context, req *apiconfig.ClientConfigFileInfo) *apiconfig.ConfigDiscoverResponse {
	namespace := req.GetNamespace().GetValue()
	out := api.NewConfigDiscoverResponse(apimodel.Code_ExecuteSuccess)

	groups, revision := s.groupCache.ListGroups(namespace)
	if revision == "" {
		out = api.NewConfigDiscoverResponse(apimodel.Code_ExecuteSuccess)
		out.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE_GROUPS
		return out
	}
	if revision == req.GetMd5().GetValue() {
		out = api.NewConfigDiscoverResponse(apimodel.Code_DataNoChange)
		out.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE_GROUPS
		return out
	}

	ret := make([]*apiconfig.ConfigFileGroup, 0, len(groups))
	for i := range groups {
		item := groups[i]
		ret = append(ret, &apiconfig.ConfigFileGroup{
			Namespace: wrapperspb.String(item.Namespace),
			Name:      wrapperspb.String(item.Name),
		})
	}

	out.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE_GROUPS
	out.ConfigFile = &apiconfig.ClientConfigFileInfo{Namespace: wrapperspb.String(namespace)}
	out.Revision = revision
	out.ConfigFileGroups = ret
	return out
}

func toClientInfo(client *apiconfig.ClientConfigFileInfo,
	release *conftypes.ConfigFileRelease) (*apiconfig.ClientConfigFileInfo, error) {

	namespace := client.GetNamespace().GetValue()
	group := client.GetGroup().GetValue()
	fileName := client.GetFileName().GetValue()
	publicKey := client.GetPublicKey().GetValue()

	copyMetadata := func() map[string]string {
		ret := map[string]string{}
		for k, v := range release.Metadata {
			ret[k] = v
		}
		delete(ret, types.MetaKeyConfigFileDataKey)
		return ret
	}()

	configFile := &apiconfig.ClientConfigFileInfo{
		Namespace: protobuf.NewStringValue(namespace),
		Group:     protobuf.NewStringValue(group),
		FileName:  protobuf.NewStringValue(fileName),
		Content:   protobuf.NewStringValue(release.Content),
		Version:   protobuf.NewUInt64Value(release.Version),
		Md5:       protobuf.NewStringValue(release.Md5),
		Encrypted: protobuf.NewBoolValue(release.IsEncrypted()),
		Tags:      conftypes.FromTagMap(copyMetadata),
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
		configFile.Tags = append(configFile.Tags,
			&apiconfig.ConfigFileTag{
				Key:   protobuf.NewStringValue(types.MetaKeyConfigFileDataKey),
				Value: protobuf.NewStringValue(dataKey),
			},
		)
	}
	return configFile, nil
}

// UpsertAndReleaseConfigFile 创建/更新配置文件并发布
func (s *Server) UpsertAndReleaseConfigFileFromClient(ctx context.Context,
	req *apiconfig.ConfigFilePublishInfo) *apiconfig.ConfigResponse {
	return s.UpsertAndReleaseConfigFile(ctx, req)
}

// DeleteConfigFileFromClient 调用config_file的方法更新配置文件
func (s *Server) DeleteConfigFileFromClient(ctx context.Context, req *apiconfig.ConfigFile) *apiconfig.ConfigResponse {
	return s.DeleteConfigFile(ctx, req)
}

// CreateConfigFileFromClient 调用config_file接口获取配置文件
func (s *Server) CreateConfigFileFromClient(ctx context.Context,
	client *apiconfig.ConfigFile) *apiconfig.ConfigClientResponse {
	configResponse := s.CreateConfigFile(ctx, client)
	return api.NewConfigClientResponseFromConfigResponse(configResponse)
}

// UpdateConfigFileFromClient 调用config_file接口更新配置文件
func (s *Server) UpdateConfigFileFromClient(ctx context.Context,
	client *apiconfig.ConfigFile) *apiconfig.ConfigClientResponse {
	configResponse := s.UpdateConfigFile(ctx, client)
	return api.NewConfigClientResponseFromConfigResponse(configResponse)
}

// PublishConfigFileFromClient 调用config_file_release接口发布配置文件
func (s *Server) PublishConfigFileFromClient(ctx context.Context,
	client *apiconfig.ConfigFileRelease) *apiconfig.ConfigClientResponse {
	configResponse := s.PublishConfigFile(ctx, client)
	return api.NewConfigClientResponseFromConfigResponse(configResponse)
}

// GetConfigSubscribers 根据配置视角获取订阅者列表
func (s *Server) GetConfigSubscribers(ctx context.Context, filter map[string]string) *types.CommonResponse {
	namespace := filter["namespace"]
	group := filter["group"]
	fileName := filter["file_name"]

	key := utils.GenFileId(namespace, group, fileName)
	clientIds, _ := s.watchCenter.watchers.Load(key)
	if clientIds == nil {
		return types.NewCommonResponse(uint32(apimodel.Code_NotFoundResource))
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

		watchCtx.ClientLabels()

		versionClients[curVer] = append(versionClients[curVer], &conftypes.Subscriber{
			ID:         watchCtx.ClientID(),
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
			Name:      fileName,
		},
		VersionClients: func() []*conftypes.VersionClient {
			ret := make([]*conftypes.VersionClient, 0, len(versionClients))
			for ver, clients := range versionClients {
				ret = append(ret, &conftypes.VersionClient{
					Versoin:     ver,
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
		key := conftypes.BuildKeyForClientConfigFileInfo(file)
		curVer := watchCtx.CurWatchVersion(key)

		ns := file.GetNamespace().GetValue()
		group := file.GetGroup().GetValue()
		filename := file.GetFileName().GetValue()

		data.Files = append(data.Files, conftypes.FileReleaseSubscribeInfo{
			Name:      file.GetName().GetValue(),
			Namespace: ns,
			Group:     group,
			FileName:  filename,
			ReleaseType: func() conftypes.ReleaseType {
				if gray := s.fileCache.GetActiveGrayRelease(ns, group, filename); gray != nil {
					if gray.Version == curVer {
						return conftypes.ReleaseTypeGray
					}
				}
				return conftypes.ReleaseTypeNormal
			}(),
			Version: curVer,
		})
	}

	rsp := types.NewCommonResponse(uint32(apimodel.Code_ExecuteSuccess))
	rsp.Data = data
	return rsp
}
