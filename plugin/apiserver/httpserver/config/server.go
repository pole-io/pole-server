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
	"github.com/emicklei/go-restful/v3"

	"github.com/pole-io/pole-server/apis/apiserver"
	"github.com/pole-io/pole-server/pkg/admin"
	"github.com/pole-io/pole-server/pkg/config"
	"github.com/pole-io/pole-server/pkg/namespace"
	"github.com/pole-io/pole-server/plugin/apiserver/httpserver/docs"
)

// HTTPServer
type HTTPServer struct {
	maintainServer  admin.AdminOperateServer
	namespaceServer namespace.NamespaceOperateServer
	configServer    config.ConfigCenterServer
}

// NewServer 创建配置中心的 HttpServer
func NewServer(
	maintainServer admin.AdminOperateServer,
	namespaceServer namespace.NamespaceOperateServer) (*HTTPServer, error) {
	// 初始化配置中心模块
	configServer, err := config.GetServer()
	if err != nil {
		log.Errorf("set config server to http server error. %v", err)
		return nil, err
	}
	return &HTTPServer{
		maintainServer:  maintainServer,
		namespaceServer: namespaceServer,
		configServer:    configServer,
	}, nil
}

const (
	defaultReadAccess   string = "default-read"
	defaultAccess       string = "default"
	configConsoleAccess string = "config"
)

// GetConfigAccessServer 获取配置中心接口
func (h *HTTPServer) GetConsoleAccessServer(include []string) *restful.WebService {
	consoleAccess := []string{defaultAccess}

	ws := new(restful.WebService)
	ws.Path("/config/v1").Consumes(restful.MIME_JSON, "multipart/form-data").Produces(restful.MIME_JSON, "application/zip")

	if len(include) == 0 {
		include = consoleAccess
	}

	for _, item := range include {
		switch item {
		case defaultReadAccess:
			h.addDefaultReadAccess(ws)
		case configConsoleAccess, defaultAccess:
			h.addDefaultAccess(ws)
		}
	}

	return ws
}

func (h *HTTPServer) addDefaultReadAccess(ws *restful.WebService) {
	ws.Route(docs.EnrichQueryConfigFileGroupsApiDocs(ws.GET("/groups").To(h.QueryConfigFileGroups)))
	ws.Route(docs.EnrichGetConfigFileApiDocs(ws.GET("/files").To(h.GetConfigFile)))
	ws.Route(docs.EnrichQueryConfigFilesByGroupApiDocs(ws.GET("/files/by-group").To(h.SearchConfigFiles)))
	ws.Route(docs.EnrichSearchConfigFileApiDocs(ws.GET("/files/search").To(h.SearchConfigFiles)))
	ws.Route(docs.EnrichGetAllConfigEncryptAlgorithms(ws.GET("/files/encryptalgorithm").
		To(h.GetAllConfigEncryptAlgorithms)))
	ws.Route(docs.EnrichGetConfigFileReleaseApiDocs(ws.GET("/files/release").To(h.GetConfigFileRelease)))
	ws.Route(docs.EnrichGetConfigFileReleaseHistoryApiDocs(ws.GET("/files/releasehistory").
		To(h.GetConfigFileReleaseHistory)))
	ws.Route(docs.EnrichGetAllConfigFileTemplatesApiDocs(ws.GET("/templates").To(h.GetAllConfigFileTemplates)))
}

func (h *HTTPServer) addDefaultAccess(ws *restful.WebService) {
	// 配置文件组
	ws.Route(docs.EnrichCreateConfigFileGroupApiDocs(ws.POST("/groups").To(h.CreateConfigFileGroup)))
	ws.Route(docs.EnrichUpdateConfigFileGroupApiDocs(ws.PUT("/groups").To(h.UpdateConfigFileGroup)))
	ws.Route(docs.EnrichDeleteConfigFileGroupApiDocs(ws.DELETE("/groups").To(h.DeleteConfigFileGroup)))
	ws.Route(docs.EnrichQueryConfigFileGroupsApiDocs(ws.GET("/groups").To(h.QueryConfigFileGroups)))

	// 配置文件
	ws.Route(docs.EnrichCreateConfigFileApiDocs(ws.POST("/files").To(h.CreateConfigFile)))
	ws.Route(docs.EnrichGetConfigFileApiDocs(ws.GET("/files").To(h.GetConfigFile)))
	ws.Route(docs.EnrichQueryConfigFilesByGroupApiDocs(ws.GET("/files/by-group").To(h.SearchConfigFiles)))
	ws.Route(docs.EnrichSearchConfigFileApiDocs(ws.GET("/files/search").To(h.SearchConfigFiles)))
	ws.Route(docs.EnrichUpdateConfigFileApiDocs(ws.PUT("/files").To(h.UpdateConfigFile)))
	ws.Route(docs.EnrichDeleteConfigFileApiDocs(ws.DELETE("/files").To(h.DeleteConfigFiles)))
	ws.Route(docs.EnrichExportConfigFileApiDocs(ws.POST("/files/export").To(h.ExportConfigFile)))
	ws.Route(docs.EnrichImportConfigFileApiDocs(ws.POST("/files/import").To(h.ImportConfigFile)))
	ws.Route(docs.EnrichGetAllConfigEncryptAlgorithms(ws.GET("/files/encryptalgorithm").
		To(h.GetAllConfigEncryptAlgorithms)))

	// 配置文件发布
	ws.Route(docs.EnrichPublishConfigFileApiDocs(ws.POST("/files/release").To(h.PublishConfigFile)))
	ws.Route(docs.EnrichGetConfigFileReleaseApiDocs(ws.PUT("/files/releases/rollback").To(h.RollbackConfigFileReleases)))
	ws.Route(docs.EnrichGetConfigFileReleaseApiDocs(ws.GET("/files/release").To(h.GetConfigFileRelease)))
	ws.Route(docs.EnrichGetConfigFileReleaseApiDocs(ws.GET("/files/releases").To(h.GetConfigFileReleases)))
	ws.Route(docs.EnrichGetConfigFileReleaseApiDocs(ws.POST("/files/releases/delete").To(h.DeleteConfigFileReleases)))
	ws.Route(docs.EnrichGetConfigFileReleaseApiDocs(ws.GET("/files/release/versions").To(h.GetConfigFileReleaseVersions)))
	ws.Route(docs.EnrichUpsertAndReleaseConfigFileApiDocs(ws.POST("/files/createandpub").To(h.UpsertAndReleaseConfigFile)))
	ws.Route(docs.EnrichStopBetaReleaseConfigFileApiDocs(ws.POST("/files/releases/stopbeta").To(h.StopGrayConfigFileReleases)))

	// 配置文件发布历史
	ws.Route(docs.EnrichGetConfigFileReleaseHistoryApiDocs(ws.GET("/files/releases/history").
		To(h.GetConfigFileReleaseHistory)))

	// config file template
	ws.Route(docs.EnrichGetAllConfigFileTemplatesApiDocs(ws.GET("/templates").To(h.GetAllConfigFileTemplates)))
	ws.Route(docs.EnrichCreateConfigFileTemplateApiDocs(ws.POST("/templates").To(h.CreateConfigFileTemplates)))
	ws.Route(docs.EnrichUpdateConfigFileTemplateApiDocs(ws.PUT("/templates").To(h.UpdateConfigFileTemplate)))
}

// GetClientAccessServer 获取配置中心接口
func (h *HTTPServer) GetClientAccessServer(ws *restful.WebService, include []string) {
	clientAccess := []string{apiserver.DiscoverAccess, apiserver.CreateFileAccess}

	if len(include) == 0 {
		include = clientAccess
	}

	for _, item := range include {
		switch item {
		case apiserver.DiscoverAccess:
			h.addDiscover(ws)
		}
	}
}

func (h *HTTPServer) addDiscover(ws *restful.WebService) {
	ws.Route(docs.EnrichConfigDiscoverApiDocs(ws.POST("/ConfigDiscover").To(h.Discover)))
	ws.Route(docs.EnrichGetConfigFileForClientApiDocs(ws.GET("/GetConfigFile").To(h.ClientGetConfigFile)))
	ws.Route(docs.EnrichWatchConfigFileForClientApiDocs(ws.POST("/WatchConfigFile").To(h.ClientWatchConfigFile)))
	ws.Route(docs.EnrichGetConfigFileMetadataList(ws.POST("/GetConfigFileMetadataList").To(h.GetConfigFileMetadataList)))
}
