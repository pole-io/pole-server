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
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/observability/statis"
	"github.com/pole-io/pole-server/apis/pkg/types"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/types/metrics"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	commontime "github.com/pole-io/pole-server/pkg/common/utils/time"
	"github.com/pole-io/pole-server/pkg/config"
	"github.com/pole-io/pole-server/plugin/apiserver/nacosserver/model"
	nacoshttp "github.com/pole-io/pole-server/plugin/apiserver/nacosserver/v1/http"
)

func (n *ConfigServer) handlePublishConfig(ctx context.Context, req *model.ConfigFile) (bool, error) {
	resp := n.configSvr.UpsertAndReleaseConfigFileFromClient(ctx, req.ToSpecConfigFile())
	if resp.GetCode() == uint32(apimodel.Code_ExecuteSuccess) {
		return true, nil
	}
	nacoslog.Error("[NACOS-V1][Config] publish config file fail",
		zap.Uint32("code", resp.GetCode()), zap.String("msg", resp.GetInfo()))
	return false, &model.NacosError{
		ErrCode: int32(model.ExceptionCode_ServerError),
		ErrMsg:  resp.GetInfo(),
	}
}

func (n *ConfigServer) handleDeleteConfig(ctx context.Context, req *model.ConfigFile) (bool, error) {
	resp := n.configSvr.DeleteConfigFileFromClient(ctx, req.ToDeleteSpec())
	if resp.GetCode() == uint32(apimodel.Code_ExecuteSuccess) {
		return true, nil
	}
	nacoslog.Error("[NACOS-V1][Config] delete config file fail",
		zap.Uint32("code", resp.GetCode()), zap.String("msg", resp.GetInfo()))
	return false, &model.NacosError{
		ErrCode: int32(model.ExceptionCode_ServerError),
		ErrMsg:  resp.GetInfo(),
	}
}

func (n *ConfigServer) handleGetConfig(ctx context.Context, req *model.ConfigFile, rsp *restful.Response) (string, error) {
	var queryResp *config_manage.ConfigDiscoverResponse
	startTime := commontime.CurrentMillisecond()
	defer func() {
		statis.GetStatis().ReportDiscoverCall(metrics.ClientDiscoverMetric{
			Action:    model.ActionGetConfigFile,
			ClientIP:  utils.ParseClientAddress(ctx),
			Namespace: req.Namespace,
			Resource:  metrics.ResourceOfConfigFile(req.Group, req.DataId),
			Timestamp: startTime,
			CostTime:  commontime.CurrentMillisecond() - startTime,
			Revision:  queryResp.GetFile().GetMd5(),
			Success:   queryResp.GetCode() > uint32(apimodel.Code_DataNoChange),
		})
	}()

	queryResp = n.configSvr.GetConfigFileWithCache(ctx, req.ToQuerySpec())
	if queryResp.GetCode() != uint32(apimodel.Code_ExecuteSuccess) {
		nacoslog.Error("[NACOS-V1][Config] query config file fail",
			zap.Uint32("code", queryResp.GetCode()), zap.String("msg", queryResp.GetInfo()))
		switch queryResp.GetCode() {
		case uint32(apimodel.Code_NotFoundResource):
			return "", &model.NacosError{
				ErrCode: int32(http.StatusNotFound),
				ErrMsg:  "config data not exist",
			}
		default:
			return "", &model.NacosError{
				ErrCode: int32(model.ExceptionCode_ServerError),
				ErrMsg:  queryResp.GetInfo(),
			}
		}
	}

	viewRelease := queryResp.GetFile()
	disableCache(rsp)
	rsp.AddHeader(model.HeaderLastModified, viewRelease.GetMtime())
	rsp.AddHeader(model.HeaderContentMD5, viewRelease.GetMd5())

	return viewRelease.GetContent(), nil
}

func (n *ConfigServer) handleWatch(ctx context.Context, listenCtx *model.ConfigWatchContext,
	rsp *restful.Response) {

	specWatchReq := listenCtx.ToSpecWatch()
	timeout, ok := listenCtx.IsSupportLongPolling()
	if !ok {
		changeKeys := n.diffChangeFiles(ctx, specWatchReq)
		oldResult := md5OldResult(changeKeys)
		newResult := md5ResultString(changeKeys)

		rsp.WriteHeader(http.StatusOK)
		disableCache(rsp)
		rsp.AddHeader(model.HeaderProbeModifyResponse, oldResult)
		rsp.AddHeader(model.HeaderProbeModifyResponseNew, newResult)
		listenCtx.Request.SetAttribute(model.HeaderContent, newResult)
		return
	}
	if changeKeys := n.diffChangeFiles(ctx, specWatchReq); len(changeKeys) > 0 {
		newResult := md5ResultString(changeKeys)
		nacoslog.Info("[NACOS-V1][Config] client quick compare file result.", zap.String("result", newResult))
		rsp.WriteHeader(http.StatusOK)
		disableCache(rsp)
		_, _ = rsp.Write([]byte(newResult))
		return
	}
	if listenCtx.IsNoHangUp() {
		// 该场景只会在 nacos-client 第一次发起订阅任务的时候
		// /com/alibaba/nacos/nacos-client/1.4.6/nacos-client-1.4.6-sources.jar!/com/alibaba/nacos/client/config/impl/ClientWorker.java:368
		nacoslog.Info("[NACOS-V1][Config] client set listen no hangup, quick return")
		return
	}
	clientId := utils.ParseClientAddress(ctx) + "@" + utils.NewUUID()[0:8]
	configSvr := n.originConfigSvr.(*config.Server)

	// 将 ConfigFileRelease 转换为 ConfigFile
	configFiles := make([]*config_manage.ConfigFile, 0, len(specWatchReq.GetFiles()))
	for _, release := range specWatchReq.GetFiles() {
		configFiles = append(configFiles, &config_manage.ConfigFile{
			Namespace: release.GetNamespace(),
			Group:     release.GetGroup(),
			Name:      release.GetName(),
		})
	}

	watchCtx := configSvr.WatchCenter().AddWatcher(clientId, configFiles,
		n.BuildTimeoutWatchCtx(ctx, timeout))
	nacoslog.Info("[NACOS-V1][Config] client start waitting server send notify message")
	notifyRet := (watchCtx.(*LongPollWatchContext)).GetNotifieResult()
	notifyCode := notifyRet.GetCode()
	if notifyCode != uint32(apimodel.Code_ExecuteSuccess) && notifyCode != uint32(apimodel.Code_DataNoChange) {
		nacoslog.Error("[NACOS-V1][Config] notify client config change",
			zap.String("remote", listenCtx.Request.Request.RemoteAddr), zap.Uint32("code", notifyCode),
			zap.String("msg", notifyRet.GetInfo()))
		rsp.WriteHeader(api.CalcCode(notifyRet))
		return
	}

	var changeKeys []*model.ConfigListenItem
	if notifyCode == uint32(apimodel.Code_DataNoChange) {
		// 按照 Nacos 原本的设计，只有 WatchClient 超时后才会再次全部 diff 比较
		changeKeys = n.diffChangeFiles(ctx, specWatchReq)
	} else {
		// 如果收到一个事件变化，就立即通知这个文件的变化信息
		changeKeys = []*model.ConfigListenItem{
			{
				Tenant: notifyRet.GetFile().GetNamespace(),
				Group:  notifyRet.GetFile().GetGroup(),
				DataId: notifyRet.GetFile().GetFileName(),
			},
		}
	}
	if len(changeKeys) == 0 {
		nacoslog.Debug("[NACOS-V1][Config] client receive empty watch result.", zap.Any("ret", notifyRet))
		rsp.WriteHeader(http.StatusOK)
		return
	}
	newResult := md5ResultString(changeKeys)
	nacoslog.Info("[NACOS-V1][Config] client receive watch result.", zap.String("result", newResult))
	rsp.WriteHeader(http.StatusOK)
	disableCache(rsp)
	_, _ = rsp.Write([]byte(newResult))
}

func (n *ConfigServer) diffChangeFiles(ctx context.Context,
	listenCtx *config_manage.ClientWatchConfigFileRequest) []*model.ConfigListenItem {
	clientLabels := map[string]string{
		types.ClientLabel_IP: utils.ParseClientIP(ctx),
	}
	changeKeys := make([]*model.ConfigListenItem, 0, 4)
	// quick get file and compare
	for _, item := range listenCtx.Files {
		namespace := item.GetNamespace()
		group := item.GetGroup()
		dataId := item.GetFileName()
		mdval := item.GetMd5()
		if beta := n.cacheSvr.ConfigFile().GetActiveGrayRelease(namespace, group, dataId); beta != nil {
			if n.cacheSvr.Gray().HitGrayRule(beta.FileKey(), clientLabels) {
				changeKeys = append(changeKeys, &model.ConfigListenItem{
					Tenant: model.ToNacosConfigNamespace(beta.Namespace),
					Group:  beta.Group,
					DataId: dataId,
				})
				continue
			}
		}

		active := n.cacheSvr.ConfigFile().GetActiveRelease(namespace, group, dataId)
		if (active == nil && mdval != "") || (active != nil && active.Md5 != mdval) {
			changeKeys = append(changeKeys, &model.ConfigListenItem{
				Tenant: model.ToNacosConfigNamespace(namespace),
				Group:  group,
				DataId: dataId,
			})
		}
	}
	return changeKeys
}

func (n *ConfigServer) BuildTimeoutWatchCtx(ctx context.Context, watchTimeOut time.Duration) config.WatchContextFactory {
	labels := map[string]string{}
	labels[types.ClientLabel_IP] = utils.ParseClientIP(ctx)

	return func(clientId string, matcher config.BetaReleaseMatcher) config.WatchContext {
		watchCtx := &LongPollWatchContext{
			clientId:         clientId,
			labels:           labels,
			finishTime:       time.Now().Add(watchTimeOut),
			finishChan:       make(chan *config_manage.ConfigDiscoverResponse),
			watchConfigFiles: map[string]*config_manage.ConfigFile{},
			betaMatcher: func(clientLabels map[string]string, event *conftypes.SimpleConfigFileRelease) bool {
				return n.cacheSvr.Gray().HitGrayRule(config.GetGrayConfigReaseKey(event), clientLabels)
			},
		}
		return watchCtx
	}
}

const (
	ConfigExportMetadata   = ".meta.yml"
	ConfigExpotrMetadataV2 = ".metadata.yml"
)

func (n *ConfigServer) handleConfigImport(ctx context.Context, policy string, result *UnZipResult, rsp *restful.Response) {
	var files []*model.ConfigFile
	var err error

	if result.Meta != nil && result.Meta.Name == ConfigExpotrMetadataV2 {
		files, err = n.parseImportV2(result)
	}

	if err != nil {
		nacoshttp.WrirteNacosResponse(map[string]interface{}{
			"errMsg": "解析数据失败",
			"code":   100004,
		}, rsp)
		return
	}

	if len(files) == 0 {
		nacoshttp.WrirteNacosResponse(map[string]interface{}{
			"errMsg": "导入的文件数据为空",
			"code":   100005,
		}, rsp)
		return
	}

	for _, file := range files {
		// 看下之前有没有存在已发布的配置文件
		existVal := n.cacheSvr.ConfigFile().GetActiveRelease(file.Namespace, file.Group, file.DataId)
		switch policy {
		case "ABORT":
			if existVal != nil {
				return
			}
		case "SKIP":
			if existVal != nil {
				continue
			}
			fallthrough
		default:
			if _, err := n.handlePublishConfig(ctx, file); err != nil {
				nacoshttp.WrirteNacosErrorResponse(err, rsp)
				return
			}
		}
	}
	nacoshttp.WrirteNacosResponse(map[string]interface{}{
		"errMsg": "导入成功",
		"code":   200,
	}, rsp)
}

func (n *ConfigServer) parseImportV1(result *UnZipResult) ([]*model.ConfigFile, error) {
	metaData := map[string]string{}
	if result.Meta != nil {
		metaDataStr := strings.ReplaceAll(string(result.Meta.Data), "[\r\n]+", "|")
		metaDataArr := strings.Split(metaDataStr, "\\|")
		for _, meta := range metaDataArr {
			metaArr := strings.Split(meta, "=")
			if len(metaArr) == 2 {
				metaData[metaArr[0]] = metaArr[1]
			}
		}
	}

	files := make([]*model.ConfigFile, 0, len(result.Items))
	for _, item := range result.Items {
		name := item.Name
		groupDataId := strings.Split(name, "/")
		if len(groupDataId) != 2 {
			return nil, nil
		}
		group := groupDataId[0]
		dataId := groupDataId[1]
		if strings.Contains(dataId, ":") {
			dataId = dataId[0:strings.LastIndex(dataId, ".")] + "~" + dataId[strings.LastIndex(dataId, ".")+1:]
		}
		metaDataId := group + "." + dataId + ".app"
		appName := ""
		if val, ok := metaData[metaDataId]; ok {
			appName = val
		}

		files = append(files, &model.ConfigFile{
			ConfigFileBase: model.ConfigFileBase{
				Namespace: result.Namespace,
				Group:     group,
				DataId:    dataId,
			},
			Content: string(item.Data),
			AppName: appName,
		})
	}
	return files, nil
}

func (n *ConfigServer) parseImportV2(result *UnZipResult) ([]*model.ConfigFile, error) {
	meta := &ConfigMetadata{}
	if err := yaml.Unmarshal(result.Meta.Data, meta); err != nil {
		return nil, err
	}
	if err := meta.Valid(); err != nil {
		return nil, err
	}

	metaKeys := meta.CalcKeys()
	for _, item := range result.Items {
		name := item.Name
		groupDataId := strings.Split(name, "/")
		if len(groupDataId) != 2 {
			return nil, nil
		}
		group := groupDataId[0]
		if _, ok := metaKeys[group]; !ok {
			return nil, nil
		}
		dataId := groupDataId[1]
		if _, ok := metaKeys[group][dataId]; !ok {
			return nil, nil
		}

		metaKeys[group][dataId] = string(item.Data)
	}

	files := make([]*model.ConfigFile, 0, len(meta.Metadata))
	for _, item := range meta.Metadata {
		if _, ok := metaKeys[item.Group]; !ok {
			return nil, nil
		}
		if _, ok := metaKeys[item.Group][item.DataId]; !ok {
			return nil, nil
		}

		files = append(files, &model.ConfigFile{
			ConfigFileBase: model.ConfigFileBase{
				Namespace: result.Namespace,
				Group:     item.Group,
				DataId:    item.DataId,
			},
			Content:     metaKeys[item.Group][item.DataId],
			Type:        item.Type,
			Description: item.Desc,
			AppName:     item.AppName,
		})
	}
	return files, nil
}

type ConfigMetadata struct {
	Metadata []ConfigExportItem `yaml:"metadata"`
}

func (c *ConfigMetadata) CalcKeys() map[string]map[string]string {
	return nil
}

// Valid .
func (c *ConfigMetadata) Valid() error {
	return &model.NacosError{
		ErrCode: 100002,
		ErrMsg:  "导入的元数据非法",
	}
}

type ConfigExportItem struct {
	Group   string `yaml:"group"`
	DataId  string `yaml:"dataId"`
	Desc    string `yaml:"desc"`
	Type    string `yaml:"type"`
	AppName string `yaml:"appName"`
}

type UnZipResult struct {
	Namespace string
	Meta      *ZipItem
	Items     []*ZipItem
}

type ZipItem struct {
	Name string
	Data []byte
}

func md5OldResult(items []*model.ConfigListenItem) string {
	sb := strings.Builder{}
	for i := range items {
		item := items[i]
		sb.WriteString(item.DataId)
		sb.WriteString(":")
		sb.WriteString(item.Group)
		sb.WriteString(";")
	}
	return sb.String()
}

func md5ResultString(items []*model.ConfigListenItem) string {
	if len(items) == 0 {
		return url.QueryEscape("")
	}
	sb := strings.Builder{}
	for i := range items {
		item := items[i]
		sb.WriteString(item.DataId)
		sb.WriteRune(model.WordSeparatorRune)
		sb.WriteString(item.Group)
		tenant := model.ToNacosConfigNamespace(item.Tenant)
		if len(tenant) != 0 {
			sb.WriteRune(model.WordSeparatorRune)
			sb.WriteString(model.ToNacosConfigNamespace(tenant))
		}
		sb.WriteRune(model.LineSeparatorRune)
	}
	return url.QueryEscape(sb.String())
}

func disableCache(rsp *restful.Response) {
	rsp.AddHeader(model.HeaderExpires, "0")
	rsp.AddHeader(model.HeaderPragma, "no-cache")
	rsp.AddHeader(model.HeaderCacheControl, "no-cache,no-store")
}
