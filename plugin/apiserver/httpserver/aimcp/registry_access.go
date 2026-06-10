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
	"encoding/json"
	"strconv"

	restful "github.com/emicklei/go-restful/v3"

	"github.com/pole-io/specification/source/go/api/v1/ai"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// ListMCPServers 查询 MCP Server registry。
func (h *HTTPServer) ListMCPServers(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	query := &ai.MCPServerQuery{
		Name:                    req.QueryParameter("name"),
		Namespace:               req.QueryParameter("namespace"),
		Business:                req.QueryParameter("business"),
		Department:              req.QueryParameter("department"),
		Protocol:                req.QueryParameter("protocol"),
		BackendType:             req.QueryParameter("backend_type"),
		BackendServiceNamespace: req.QueryParameter("backend_service_namespace"),
		BackendServiceName:      req.QueryParameter("backend_service_name"),
		Offset:                  uint32Query(req, "offset", 0),
		Limit:                   uint32Query(req, "limit", 100),
	}
	handler.WriteHeaderAndProto(h.mcpServerQuery(handler.ParseHeaderContext(), query))
}

// CreateMCPServers 创建 MCP Server registry 记录。
func (h *HTTPServer) CreateMCPServers(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	servers, err := readMCPServers(req)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.mcpServerCreate(handler.ParseHeaderContext(), servers))
}

// UpdateMCPServers 更新 MCP Server registry 记录。
func (h *HTTPServer) UpdateMCPServers(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	servers, err := readMCPServers(req)
	if err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.mcpServerUpdate(handler.ParseHeaderContext(), servers))
}

// DeleteMCPServers 删除 MCP Server registry 记录。
func (h *HTTPServer) DeleteMCPServers(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	var deleteReq ai.MCPServerDeleteRequest
	if err := req.ReadEntity(&deleteReq); err != nil {
		handler.WriteHeaderAndProto(api.NewResponseWithMsg(apimodel.Code_ParseException, err.Error()))
		return
	}
	handler.WriteHeaderAndProto(h.mcpServerDelete(handler.ParseHeaderContext(), deleteReq.GetServerIds()))
}

// ListMCPServerTools 查询 MCP Server tools。
func (h *HTTPServer) ListMCPServerTools(req *restful.Request, rsp *restful.Response) {
	handler := &httpcommon.Handler{Request: req, Response: rsp}
	query := &ai.MCPServerToolQuery{
		ServerId:        req.QueryParameter("server_id"),
		ServerName:      req.QueryParameter("server_name"),
		ServerNamespace: req.QueryParameter("server_namespace"),
		Offset:          uint32Query(req, "offset", 0),
		Limit:           uint32Query(req, "limit", 100),
	}
	handler.WriteHeaderAndProto(h.mcpServerToolQuery(handler.ParseHeaderContext(), query))
}

func readMCPServers(req *restful.Request) ([]*ai.MCPServer, error) {
	var servers []*ai.MCPServer
	if err := json.NewDecoder(req.Request.Body).Decode(&servers); err != nil {
		return nil, err
	}
	return servers, nil
}

func uint32Query(req *restful.Request, key string, defaultValue uint32) uint32 {
	raw := req.QueryParameter(key)
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return defaultValue
	}
	return uint32(value)
}
