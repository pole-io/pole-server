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
	"strings"

	"github.com/emicklei/go-restful/v3"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/observability/statis"
	"github.com/pole-io/pole-server/apis/pkg/types/metrics"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	commontime "github.com/pole-io/pole-server/pkg/common/utils/time"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) addDiscover(ws *restful.WebService) {
	ws.Route(docs.EnrichConfigDiscoverApiDocs(ws.POST("/ConfigDiscover").To(h.Discover)))
	ws.Route(docs.EnrichGetConfigFileForClientApiDocs(ws.GET("/GetConfigFile").To(h.ClientGetConfigFile)))
	ws.Route(docs.EnrichWatchConfigFileForClientApiDocs(ws.POST("/WatchConfigFile").To(h.ClientWatchConfigFile)))
	ws.Route(docs.EnrichGetConfigFileMetadataList(ws.POST("/GetConfigFileMetadataList").To(h.GetConfigFileMetadataList)))
}

func (h *HTTPServer) ClientGetConfigFile(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	configFile := &apiconfig.ConfigFile{
		Namespace: handler.Request.QueryParameter("namespace"),
		Group:     handler.Request.QueryParameter("group"),
		Name:      handler.Request.QueryParameter("fileName"),
		Labels: func() map[string]string {
			tags := handler.Request.QueryParameters("tags")
			ret := make(map[string]string, len(tags))
			for i := range tags {
				kv := strings.Split(tags[i], "=")
				if len(kv) == 2 {
					ret[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
				}
			}
			return ret
		}(),
	}

	ctx := handler.ParseHeaderContext()
	startTime := commontime.CurrentMillisecond()
	var ret *apiconfig.ConfigDiscoverResponse
	defer func() {
		statis.GetStatis().ReportDiscoverCall(metrics.ClientDiscoverMetric{
			Action:    metrics.ActionGetConfigFile,
			ClientIP:  utils.ParseClientAddress(ctx),
			Namespace: configFile.GetNamespace(),
			Resource:  metrics.ResourceOfConfigFile(configFile.GetGroup(), configFile.GetName()),
			Timestamp: startTime,
			CostTime:  commontime.CurrentMillisecond() - startTime,
			Revision:  ret.GetRevision(),
			Success:   ret.GetCode() > uint32(apimodel.Code_DataNoChange),
		})
	}()

	ret = h.configServer.GetConfigFileWithCache(ctx, configFile)
	handler.WriteHeaderAndProto(ret)
}

func (h *HTTPServer) ClientWatchConfigFile(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	// 1. 解析出客户端监听的配置文件列表
	watchConfigFileRequest := &apiconfig.WatchConfigFileRequest{}
	if _, err := handler.Parse(watchConfigFileRequest); err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	// 阻塞等待响应
	// 将 ClientWatchConfigFileRequest 转换为 ConfigFileGroupRequest
	// 这里需要根据实际业务逻辑进行适当的转换
	groupRequest := &apiconfig.ConfigFileGroupRequest{
		// 根据实际需要填充字段
	}
	callback, err := h.configServer.LongPullWatchFile(handler.ParseHeaderContext(), groupRequest)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(callback())
}

// GetConfigFileMetadataList 统一发现接口
func (h *HTTPServer) GetConfigFileMetadataList(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	in := &apiconfig.ConfigFileGroupRequest{}
	ctx, err := handler.Parse(in)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	var out *apiconfig.ConfigDiscoverResponse
	startTime := commontime.CurrentMillisecond()
	defer func() {
		statis.GetStatis().ReportDiscoverCall(metrics.ClientDiscoverMetric{
			Action:    metrics.ActionListConfigFiles,
			ClientIP:  utils.ParseClientAddress(ctx),
			Namespace: in.GetConfigFileGroup().GetNamespace(),
			Resource:  metrics.ResourceOfConfigFileList(in.GetConfigFileGroup().GetName()),
			Timestamp: startTime,
			CostTime:  commontime.CurrentMillisecond() - startTime,
			Revision:  out.GetRevision(),
			Success:   out.GetCode() > uint32(apimodel.Code_DataNoChange),
		})
	}()

	out = h.configServer.GetConfigFileNamesWithCache(ctx, in)
	handler.WriteHeaderAndProto(out)
}

// Discover 统一发现接口
func (h *HTTPServer) Discover(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	in := &apiconfig.ConfigDiscoverRequest{}
	ctx, err := handler.Parse(in)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	var out *apiconfig.ConfigDiscoverResponse
	var action string
	startTime := commontime.CurrentMillisecond()
	defer func() {
		statis.GetStatis().ReportDiscoverCall(metrics.ClientDiscoverMetric{
			Action:    action,
			ClientIP:  utils.ParseClientAddress(ctx),
			Namespace: in.GetFile().GetNamespace(),
			Resource:  metrics.ResourceOfConfigFile(in.GetFile().GetGroup(), in.GetFile().GetName()),
			Timestamp: startTime,
			CostTime:  commontime.CurrentMillisecond() - startTime,
			Revision:  out.GetRevision(),
			Success:   out.GetCode() > uint32(apimodel.Code_DataNoChange),
		})
	}()

	switch in.Type {
	case apiconfig.ConfigDiscoverRequest_CONFIG_FILE:
		action = metrics.ActionGetConfigFile
		ret := h.configServer.GetConfigFileWithCache(ctx, &apiconfig.ConfigFile{
			Namespace: in.GetFile().GetNamespace(),
			Group:     in.GetFile().GetGroup(),
			Name:      in.GetFile().GetName(),
		})
		out = api.NewConfigDiscoverResponse(apimodel.Code(ret.GetCode()))
		out.File = ret.GetFile()
		out.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE
		out.Revision = ret.GetRevision()
	case apiconfig.ConfigDiscoverRequest_CONFIG_FILE_NAMES:
		action = metrics.ActionListConfigFiles
		ret := h.configServer.GetConfigFileNamesWithCache(ctx, &apiconfig.ConfigFileGroupRequest{
			Revision: in.GetRevision(),
			ConfigFileGroup: &apiconfig.ConfigFileGroup{
				Namespace: in.GetFile().GetNamespace(),
				Name:      in.GetFile().GetGroup(),
			},
		})
		out = api.NewConfigDiscoverResponse(apimodel.Code(ret.GetCode()))
		out.FileNames = ret.GetFileNames()
		out.Type = apiconfig.ConfigDiscoverResponse_CONFIG_FILE_NAMES
		out.Revision = ret.GetRevision()
	case apiconfig.ConfigDiscoverRequest_CONFIG_FILE_GROUPS:
		action = metrics.ActionListConfigGroups
		req := in.GetFile()
		out = h.configServer.GetConfigGroupsWithCache(ctx, req)
	default:
		out = api.NewConfigDiscoverResponse(apimodel.Code_InvalidDiscoverResource)
	}

	handler.WriteHeaderAndProtoV2(out)
}
