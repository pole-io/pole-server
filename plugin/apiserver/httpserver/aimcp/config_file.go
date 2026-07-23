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
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	"go.uber.org/zap"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) addToolsConfigFile(mcpSvr *server.MCPServer) {
	h.addToolGetConfigFile(mcpSvr)
	h.addToolSearchConfigFiles(mcpSvr)
}

func (h *HTTPServer) addToolGetConfigFile(mcpSvr *server.MCPServer) {
	mcpSvr.AddTool(
		mcp.NewTool("get_config_file",
			mcp.WithDescription("查询指定命名空间、配置分组和文件名的配置文件草稿及发布状态，只读，不会修改或发布配置"),
			mcp.WithString("namespace",
				mcp.Description("配置文件所属命名空间"),
				mcp.Required(),
			),
			mcp.WithString("group",
				mcp.Description("配置分组名称"),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("配置文件名称"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] get config file", zap.Any("params", req.Params))

			namespace, group, name, errResult := exactConfigFileArguments(req.Params.Arguments)
			if errResult != nil {
				return errResult, nil
			}
			rsp := h.configServer.GetConfigFileRichInfo(ctx, &apiconfig.ConfigFile{
				Namespace: namespace,
				Group:     group,
				Name:      name,
			})
			if !api.IsSuccess(rsp) {
				return mcp.NewToolResultError(rsp.GetInfo()), nil
			}
			ret, err := httpcommon.MarshalPBJson(rsp)
			return mcp.NewToolResultText(ret), err
		},
	)
}

func (h *HTTPServer) addToolSearchConfigFiles(mcpSvr *server.MCPServer) {
	mcpSvr.AddTool(
		mcp.NewTool("search_config_files",
			mcp.WithDescription("按命名空间、配置分组或文件名查询配置文件列表，只读，不会修改或发布配置"),
			mcp.WithString("namespace",
				mcp.Description("配置文件所属命名空间"),
				mcp.Required(),
			),
			mcp.WithString("group",
				mcp.Description("配置分组名称，支持服务端已有的模糊查询语法"),
			),
			mcp.WithString("name",
				mcp.Description("配置文件名称，支持服务端已有的模糊查询语法"),
			),
			mcp.WithNumber("offset",
				mcp.Description("查询偏移量"),
				mcp.DefaultNumber(0),
			),
			mcp.WithNumber("limit",
				mcp.Description("单次返回数量，默认 20，最大 100"),
				mcp.DefaultNumber(20),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] search config files", zap.Any("params", req.Params))

			filters, errResult := configFileSearchFilters(req.Params.Arguments)
			if errResult != nil {
				return errResult, nil
			}
			rsp := h.configServer.SearchConfigFiles(ctx, filters)
			if !api.IsSuccess(rsp) {
				return mcp.NewToolResultError(rsp.GetInfo()), nil
			}
			ret, err := httpcommon.MarshalPBJson(rsp)
			return mcp.NewToolResultText(ret), err
		},
	)
}

func exactConfigFileArguments(args map[string]interface{}) (string, string, string, *mcp.CallToolResult) {
	if args == nil {
		return "", "", "", mcp.NewToolResultError("invalid: args is empty")
	}
	namespace, _ := args["namespace"].(string)
	group, _ := args["group"].(string)
	name, _ := args["name"].(string)
	namespace = strings.TrimSpace(namespace)
	group = strings.TrimSpace(group)
	name = strings.TrimSpace(name)
	if namespace == "" || group == "" || name == "" {
		return "", "", "", mcp.NewToolResultError("invalid: namespace, group and name are required")
	}
	return namespace, group, name, nil
}

func configFileSearchFilters(args map[string]interface{}) (map[string]string, *mcp.CallToolResult) {
	if args == nil {
		return nil, mcp.NewToolResultError("invalid: args is empty")
	}
	namespace, _ := args["namespace"].(string)
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return nil, mcp.NewToolResultError("invalid: namespace is required")
	}
	filters := map[string]string{
		"namespace": namespace,
		"offset":    strconv.FormatUint(uint64(mcpUint32Argument(args["offset"], 0, 0)), 10),
		"limit":     strconv.FormatUint(uint64(mcpUint32Argument(args["limit"], 20, 100)), 10),
	}
	if group, _ := args["group"].(string); strings.TrimSpace(group) != "" {
		filters["group"] = strings.TrimSpace(group)
	}
	if name, _ := args["name"].(string); strings.TrimSpace(name) != "" {
		filters["name"] = strings.TrimSpace(name)
	}
	return filters, nil
}

func mcpUint32Argument(value interface{}, defaultValue, maxValue uint32) uint32 {
	var parsed uint64
	switch typed := value.(type) {
	case float64:
		if typed < 0 {
			return defaultValue
		}
		parsed = uint64(typed)
	case float32:
		if typed < 0 {
			return defaultValue
		}
		parsed = uint64(typed)
	case int:
		if typed < 0 {
			return defaultValue
		}
		parsed = uint64(typed)
	case uint32:
		parsed = uint64(typed)
	case uint64:
		parsed = typed
	case nil:
		return defaultValue
	default:
		return defaultValue
	}
	if parsed > uint64(^uint32(0)) {
		return defaultValue
	}
	result := uint32(parsed)
	if maxValue > 0 && result > maxValue {
		return maxValue
	}
	return result
}
