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
	"net/http"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	apiconfig "github.com/polarismesh/specification/source/go/api/v1/config_manage"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) addFilesAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichCreateConfigFileApiDocs(ws.POST("/files").To(h.CreateConfigFiles)))
	ws.Route(docs.EnrichUpdateConfigFileApiDocs(ws.PUT("/files").To(h.UpdateConfigFiles)))
	ws.Route(docs.EnrichDeleteConfigFileApiDocs(ws.POST("/files/delete").To(h.DeleteConfigFiles)))
	ws.Route(docs.EnrichGetConfigFileApiDocs(ws.GET("/files/detail").To(h.GetConfigFile)))
	ws.Route(docs.EnrichSearchConfigFileApiDocs(ws.GET("/files/search").To(h.SearchConfigFiles)))
	ws.Route(docs.EnrichExportConfigFileApiDocs(ws.POST("/files/export").To(h.ExportConfigFile)))
	ws.Route(docs.EnrichImportConfigFileApiDocs(ws.POST("/files/import").To(h.ImportConfigFile)))
	ws.Route(docs.EnrichGetAllConfigEncryptAlgorithms(ws.GET("/files/encrypt/algorithms").
		To(h.GetAllConfigEncryptAlgorithms)))
	ws.Route(docs.EnrichGetConfigFileReleaseHistoryApiDocs(ws.GET("/files/op/history").
		To(h.GetConfigFileReleaseHistory)))
}

// CreateConfigFile 创建配置文件
func (h *HTTPServer) CreateConfigFiles(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var files []*apiconfig.ConfigFile
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiconfig.ConfigFile{}
		files = append(files, msg)
		return msg
	})
	if err != nil {
		configLog.Error("[Config][HttpServer] parse config file from request error.",
			utils.RequestID(ctx), zap.Error(err))
		handler.WriteHeaderAndProto(api.NewConfigResponseWithInfo(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.CreateConfigFiles(ctx, files))
}

// GetConfigFile 获取单个配置文件
func (h *HTTPServer) GetConfigFile(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	namespace := handler.Request.QueryParameter("namespace")
	group := handler.Request.QueryParameter("group")
	name := handler.Request.QueryParameter("name")

	fileReq := &apiconfig.ConfigFile{
		Namespace: protobuf.NewStringValue(namespace),
		Group:     protobuf.NewStringValue(group),
		Name:      protobuf.NewStringValue(name),
	}

	response := h.configServer.GetConfigFileRichInfo(handler.ParseHeaderContext(), fileReq)
	handler.WriteHeaderAndProto(response)
}

// SearchConfigFiles 按照 group 和 name 模糊搜索配置文件，按照 tag 搜索，多个tag之间或的关系
func (h *HTTPServer) SearchConfigFiles(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	filters := httpcommon.ParseQueryParams(req)
	response := h.configServer.SearchConfigFiles(handler.ParseHeaderContext(), filters)

	handler.WriteHeaderAndProto(response)
}

// UpdateConfigFile 更新配置文件
func (h *HTTPServer) UpdateConfigFiles(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var files []*apiconfig.ConfigFile
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiconfig.ConfigFile{}
		files = append(files, msg)
		return msg
	})
	if err != nil {
		configLog.Error("[Config][HttpServer] parse config file from request error.",
			utils.RequestID(ctx), zap.Error(err))
		handler.WriteHeaderAndProto(api.NewConfigResponseWithInfo(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.UpdateConfigFiles(ctx, files))
}

// DeleteConfigFiles 批量删除配置文件
func (h *HTTPServer) DeleteConfigFiles(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	var files []*apiconfig.ConfigFile
	ctx, err := handler.ParseArray(func() proto.Message {
		msg := &apiconfig.ConfigFile{}
		files = append(files, msg)
		return msg
	})
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}

	handler.WriteHeaderAndProto(h.configServer.DeleteConfigFiles(ctx, files))
}

// ExportConfigFile 导出配置文件
func (h *HTTPServer) ExportConfigFile(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	configFileExport := &apiconfig.ConfigFileExportRequest{}
	ctx, err := handler.Parse(configFileExport)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	response := h.configServer.ExportConfigFile(ctx, configFileExport)
	if response.Code.Value != api.ExecuteSuccess {
		handler.WriteHeaderAndProto(response)
	} else {
		handler.WriteHeader(api.ExecuteSuccess, http.StatusOK)
		handler.Response.AddHeader("Content-Type", "application/zip")
		handler.Response.AddHeader("Content-Disposition", "attachment; filename=config.zip")
		if _, err := handler.Response.ResponseWriter.Write(response.Data.Value); err != nil {
			configLog.Error("[Config][HttpServer] response write error.",
				utils.RequestID(ctx),
				zap.String("error", err.Error()))
		}
	}
}

// ImportConfigFile 导入配置文件
func (h *HTTPServer) ImportConfigFile(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{
		Request:  req,
		Response: rsp,
	}

	ctx := handler.ParseHeaderContext()
	configFiles, err := handler.ParseFile()
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	namespace := handler.Request.QueryParameter("namespace")
	group := handler.Request.QueryParameter("group")
	conflictHandling := handler.Request.QueryParameter("conflict_handling")

	for _, file := range configFiles {
		file.Namespace = protobuf.NewStringValue(namespace)
		if group != "" {
			file.Group = protobuf.NewStringValue(group)
		}
	}

	var filenames []string
	for _, file := range configFiles {
		filenames = append(filenames, file.String())
	}
	configLog.Info("[Config][HttpServer]import config file",
		zap.String("namespace", namespace),
		zap.String("group", group),
		zap.String("conflict_handling", conflictHandling),
		zap.String("files", strings.Join(filenames, ",")),
	)

	response := h.configServer.ImportConfigFile(ctx, configFiles, conflictHandling)
	handler.WriteHeaderAndProto(response)
}
