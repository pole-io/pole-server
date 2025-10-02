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

package aimcp

import (
	"context"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/pkg/admin"
	commonlog "github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/version"
	"github.com/pole-io/pole-server/pkg/config"
	"github.com/pole-io/pole-server/pkg/namespace"
	"github.com/pole-io/pole-server/pkg/service"
)

const (
	defaultReadAccess string = "default-read"
	defaultAccess     string = "default"
	aimcpAccess       string = "aimcp"

	basePath string = "/ai/mcp/v1"

	sseEp string = "/sse"
	msgEp string = "/message"
)

// HTTPServer
type HTTPServer struct {
	maintainServer  admin.AdminOperateServer
	namespaceServer namespace.NamespaceOperateServer
	configServer    config.ConfigCenterServer
	discoverySvr    service.DiscoverServer
	mcpSvr          *server.MCPServer
	sseSvr          *server.SSEServer
}

// NewServer 创建配置中心的 HttpServer
func NewServer(
	maintainServer admin.AdminOperateServer,
	namespaceServer namespace.NamespaceOperateServer) (*HTTPServer, error) {
	// 初始化配置中心模块
	configServer, err := config.GetServer()
	if err != nil {
		commonlog.Errorf("set config server to http server error. %v", err)
		return nil, err
	}
	// 初始化服务发现模块
	discoverySvr, err := service.GetServer()
	if err != nil {
		commonlog.Errorf("set discovery server to http server error. %v", err)
		return nil, err
	}

	mcpSvr := server.NewMCPServer("pole.io", version.Get(),
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)
	sseSvr := server.NewSSEServer(mcpSvr,
		server.WithBasePath(basePath),
		server.WithSSEContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			newheaders := make(http.Header)
			for k, v := range r.Header {
				if len(v) == 0 {
					continue
				}
				if k == types.HeaderAuthorizationKey {
					v[0] = strings.TrimPrefix(v[0], "Bearer ")
				}
				newheaders[k] = v
			}
			if log.DebugEnabled() {
				log.Debug("[apiserver][ai-mcp] sse rewrite header", zap.Any("origin", r.Header), zap.Any("new", newheaders))
			}
			return types.AppendRequestHeader(ctx, newheaders)
		}),
		server.WithSSEEndpoint(sseEp),
		server.WithMessageEndpoint(msgEp),
	)

	return &HTTPServer{
		maintainServer:  maintainServer,
		namespaceServer: namespaceServer,
		configServer:    configServer,
		discoverySvr:    discoverySvr,
		mcpSvr:          mcpSvr,
		sseSvr:          sseSvr,
	}, nil
}

// GetConfigAccessServer 获取配置中心接口
func (h *HTTPServer) GetMCPAccessServer(include []string) *restful.WebService {
	commonlog.Info("enable ai-mcp access server")

	h.addMcpTools()
	consoleAccess := []string{defaultAccess}

	ws := new(restful.WebService)
	ws.Path(basePath).Consumes(restful.MIME_JSON, "multipart/form-data", "text/event-stream").Produces(restful.MIME_JSON, "text/event-stream", "application/zip")

	if len(include) == 0 {
		include = consoleAccess
	}

	for _, item := range include {
		switch item {
		case aimcpAccess, defaultAccess:
			h.addDefaultAccess(ws)
		}
	}

	return ws
}

func (h *HTTPServer) addMcpTools() {
	h.addToolsNamespace(h.mcpSvr)
}

func (h *HTTPServer) addDefaultAccess(ws *restful.WebService) {
	// MCP sse handler
	ws.Route(ws.GET(sseEp).To(func(req *restful.Request, rsp *restful.Response) {
		h.sseSvr.ServeHTTP(rsp, req.Request)
	}))
	ws.Route(ws.POST(sseEp).To(func(req *restful.Request, rsp *restful.Response) {
		h.sseSvr.ServeHTTP(rsp, req.Request)
	}))

	// MCP message handler
	ws.Route(ws.GET(msgEp).To(func(req *restful.Request, rsp *restful.Response) {
		h.sseSvr.ServeHTTP(rsp, req.Request)
	}))
	ws.Route(ws.POST(msgEp).To(func(req *restful.Request, rsp *restful.Response) {
		h.sseSvr.ServeHTTP(rsp, req.Request)
	}))
}
