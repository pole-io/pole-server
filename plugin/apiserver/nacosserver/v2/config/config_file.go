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
	"time"

	"go.uber.org/zap"

	"github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/observability/statis"
	"github.com/pole-io/pole-server/apis/pkg/types"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/types/metrics"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/pole-server/pkg/common/utils"
	commontime "github.com/pole-io/pole-server/pkg/common/utils/time"
	"github.com/pole-io/pole-server/pkg/config"
	nacosmodel "github.com/pole-io/pole-server/plugin/apiserver/nacosserver/model"
	nacospb "github.com/pole-io/pole-server/plugin/apiserver/nacosserver/v2/pb"
	"github.com/pole-io/pole-server/plugin/apiserver/nacosserver/v2/remote"
)

const (
	ErrorConfigNotFound      = 300
	ErrorConfigQueryConflict = 400
)

func (h *ConfigServer) handlePublishConfigRequest(ctx context.Context, req nacospb.BaseRequest,
	meta nacospb.RequestMeta) (nacospb.BaseResponse, error) {
	configReq, ok := req.(*nacospb.ConfigPublishRequest)
	if !ok {
		return nil, remote.ErrorInvalidRequestBodyType
	}

	var resp *apimodel.Response
	startTime := commontime.CurrentMillisecond()
	defer func() {
		statis.GetStatis().ReportDiscoverCall(metrics.ClientDiscoverMetric{
			Action:    nacosmodel.ActionGrpcPublishConfigFile,
			ClientIP:  meta.ConnectionID,
			Namespace: configReq.Tenant,
			Resource:  metrics.ResourceOfConfigFile(configReq.Group, configReq.DataId),
			Timestamp: startTime,
			CostTime:  commontime.CurrentMillisecond() - startTime,
			Success:   resp.GetCode() == uint32(apimodel.Code_ExecuteSuccess),
		})
	}()

	resp = h.configSvr.UpsertAndReleaseConfigFileFromClient(ctx, configReq.ToSpec())
	if resp.GetCode() != uint32(apimodel.Code_ExecuteSuccess) {
		nacoslog.Error("[NACOS-V2][Config] publish config file fail", zap.String("tenant", configReq.Tenant),
			utils.ZapGroup(configReq.Group), utils.ZapFileName(configReq.DataId),
			zap.Uint32("code", resp.GetCode()), zap.String("msg", resp.GetInfo()))
		return &nacospb.ConfigPublishResponse{
			Response: &nacospb.Response{
				Success:    false,
				ResultCode: int(nacosmodel.Response_Fail.Code),
				ErrorCode:  int(resp.GetCode()),
				Message:    resp.GetInfo(),
			},
		}, nil
	}

	return &nacospb.ConfigPublishResponse{
		Response: &nacospb.Response{
			Success:    true,
			ResultCode: int(nacosmodel.Response_Success.Code),
			Message:    nacosmodel.Response_Success.Desc,
		},
	}, nil
}

func (h *ConfigServer) handleGetConfigRequest(ctx context.Context, req nacospb.BaseRequest,
	meta nacospb.RequestMeta) (nacospb.BaseResponse, error) {
	configReq, ok := req.(*nacospb.ConfigQueryRequest)
	if !ok {
		return nil, remote.ErrorInvalidRequestBodyType
	}
	var rsp *nacospb.ConfigQueryResponse

	startTime := commontime.CurrentMillisecond()
	defer func() {
		statis.GetStatis().ReportDiscoverCall(metrics.ClientDiscoverMetric{
			Action:    nacosmodel.ActionGrpcGetConfigFile,
			ClientIP:  meta.ConnectionID,
			Namespace: configReq.Tenant,
			Resource:  metrics.ResourceOfConfigFile(configReq.Group, configReq.DataId),
			Timestamp: startTime,
			CostTime:  commontime.CurrentMillisecond() - startTime,
			Revision:  rsp.Md5,
			Success:   rsp.Success,
		})
	}()

	queryReq := configReq.ToQuerySpec()
	queryResp := h.configSvr.GetConfigFileWithCache(ctx, queryReq)
	if queryResp.GetCode() != uint32(apimodel.Code_ExecuteSuccess) {
		nacoslog.Error("[NACOS-V2][Config] query config file fail", zap.String("tenant", configReq.Tenant),
			utils.ZapNamespace(queryReq.GetNamespace()), utils.ZapGroup(configReq.Group),
			utils.ZapFileName(configReq.DataId), zap.Uint32("code", queryResp.GetCode()),
			zap.String("msg", queryResp.GetInfo()))
		switch queryResp.GetCode() {
		case uint32(apimodel.Code_NotFoundResource):
			rsp = &nacospb.ConfigQueryResponse{
				Response: &nacospb.Response{
					ResultCode: int(nacosmodel.Response_Fail.Code),
					ErrorCode:  ErrorConfigNotFound,
					Message:    "config data not exist",
				},
			}
			return rsp, nil
		default:
			rsp = &nacospb.ConfigQueryResponse{
				Response: &nacospb.Response{
					ResultCode: int(nacosmodel.Response_Fail.Code),
					ErrorCode:  int(queryResp.GetCode()),
					Message:    queryResp.GetInfo(),
				},
			}
			return rsp, nil
		}
	}

	viewRelease := queryResp.GetFile()

	rsp = &nacospb.ConfigQueryResponse{
		Response: &nacospb.Response{
			ResultCode: int(nacosmodel.Response_Success.Code),
			Success:    true,
		},
		Content:      viewRelease.GetContent(),
		Md5:          viewRelease.GetMd5(),
		LastModified: stringToTimestamp(viewRelease.GetMtime()),
	}
	return rsp, nil
}

func (h *ConfigServer) handleDeleteConfigRequest(ctx context.Context, req nacospb.BaseRequest,
	meta nacospb.RequestMeta) (nacospb.BaseResponse, error) {
	configReq, ok := req.(*nacospb.ConfigRemoveRequest)
	if !ok {
		return nil, remote.ErrorInvalidRequestBodyType
	}
	delResp := h.configSvr.DeleteConfigFileFromClient(ctx, configReq.ToSpec())
	if delResp.GetCode() != uint32(apimodel.Code_ExecuteSuccess) {
		nacoslog.Error("[NACOS-V2][Config] delete config file fail", zap.String("tenant", configReq.Tenant),
			utils.ZapGroup(configReq.Group), utils.ZapFileName(configReq.DataId),
			zap.Uint32("code", delResp.GetCode()), zap.String("msg", delResp.GetInfo()))
		return &nacospb.ConfigRemoveResponse{
			Response: &nacospb.Response{
				Success:    false,
				ResultCode: int(nacosmodel.Response_Fail.Code),
				ErrorCode:  int(delResp.GetCode()),
				Message:    delResp.GetInfo(),
			},
		}, nil
	}

	return &nacospb.ConfigRemoveResponse{
		Response: &nacospb.Response{
			Success:    true,
			ResultCode: int(nacosmodel.Response_Success.Code),
			Message:    nacosmodel.Response_Success.Desc,
		},
	}, nil
}

func (h *ConfigServer) handleWatchConfigRequest(ctx context.Context, req nacospb.BaseRequest,
	meta nacospb.RequestMeta) (nacospb.BaseResponse, error) {
	watchReq, ok := req.(*nacospb.ConfigBatchListenRequest)
	if !ok {
		return nil, remote.ErrorInvalidRequestBodyType
	}
	configSvr := h.originConfigSvr.(*config.Server)

	listenResp := nacospb.NewConfigChangeBatchListenResponse()
	clientId := meta.ConnectionID
	specReq := watchReq.ToSpec()
	if watchReq.Listen {
		// 转换ConfigFileRelease到ConfigFile
		configFiles := make([]*config_manage.ConfigFile, len(specReq.GetFiles()))
		for i, fileRelease := range specReq.GetFiles() {
			// 将MD5信息保存在Tags中以便后续比较
			tags := make(map[string]string)
			tags["md5"] = fileRelease.GetMd5()

			configFiles[i] = &config_manage.ConfigFile{
				Id:        fileRelease.GetConfigFileId(),
				Name:      fileRelease.GetFileName(),
				Namespace: fileRelease.GetNamespace(),
				Group:     fileRelease.GetGroup(),
				Content:   fileRelease.GetContent(),
				Format:    fileRelease.GetFormat(),
				Comment:   fileRelease.GetComment(),
				Tags:      tags,
			}
		}
		watchCtx := configSvr.WatchCenter().AddWatcher(clientId, configFiles, h.BuildGrpcWatchCtx(ctx))
		for i := range specReq.GetFiles() {
			item := specReq.GetFiles()[i]
			namespace := item.GetNamespace()
			group := item.GetGroup()
			dataId := item.GetFileName()
			mdval := item.GetMd5()

			var active *conftypes.ConfigFileRelease
			var match bool
			if betaActive := h.cacheSvr.ConfigFile().GetActiveGrayRelease(namespace, group, dataId); betaActive != nil {
				match = h.cacheSvr.Gray().HitGrayRule(config.GetGrayConfigReaseKey(betaActive.SimpleConfigFileRelease), watchCtx.ClientLabels())
				active = betaActive
			}
			if !match {
				active = h.cacheSvr.ConfigFile().GetActiveRelease(namespace, group, dataId)
			}

			// 如果 client 过来的 MD5 是一个空字符串
			if (active == nil && mdval != "") || (active != nil && active.Md5 != mdval) {
				listenResp.ChangedConfigs = append(listenResp.ChangedConfigs, nacospb.ConfigContext{
					Tenant: nacosmodel.ToNacosConfigNamespace(namespace),
					Group:  group,
					DataId: dataId,
				})
			}
		}
	} else {
		watchCtx, ok := configSvr.WatchCenter().GetWatchContext(clientId)
		if ok {
			for i := range specReq.GetFiles() {
				fileRelease := specReq.GetFiles()[i]
				// 转换ConfigFileRelease到ConfigFile
				tags := make(map[string]string)
				tags["md5"] = fileRelease.GetMd5()

				configFile := &config_manage.ConfigFile{
					Id:        fileRelease.GetConfigFileId(),
					Name:      fileRelease.GetFileName(),
					Namespace: fileRelease.GetNamespace(),
					Group:     fileRelease.GetGroup(),
					Content:   fileRelease.GetContent(),
					Format:    fileRelease.GetFormat(),
					Comment:   fileRelease.GetComment(),
					Tags:      tags,
				}
				watchCtx.RemoveInterest(configFile)
			}
		}
	}
	return listenResp, nil
}

// BuildGrpcWatchCtx .
func (h *ConfigServer) BuildGrpcWatchCtx(ctx context.Context) config.WatchContextFactory {
	labels := map[string]string{}
	labels[types.ClientLabel_IP] = utils.ParseClientIP(ctx)

	return func(clientId string, matcher config.BetaReleaseMatcher) config.WatchContext {
		watchCtx := &StreamWatchContext{
			clientId:         clientId,
			connMgr:          h.connMgr,
			labels:           labels,
			watchConfigFiles: container.NewSyncMap[string, *config_manage.ConfigFile](),
			betaMatcher: func(clientLabels map[string]string, event *conftypes.SimpleConfigFileRelease) bool {
				return h.cacheSvr.Gray().HitGrayRule(config.GetGrayConfigReaseKey(event), clientLabels)
			},
		}
		return watchCtx
	}
}

// stringToTimestamp Convert string to timestamp
func stringToTimestamp(val string) int64 {
	lastModified, _ := time.Parse("2006-01-02 15:04:05", val)
	return lastModified.UnixMilli()
}
