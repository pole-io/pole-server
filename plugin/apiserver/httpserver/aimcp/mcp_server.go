package aimcp

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	"github.com/pole-io/pole-server/apis/pkg/types/ai"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

// addToolQueryMCPServers 添加查询 MCP Servers 的 tool
func (h *HTTPServer) addToolQueryMCPServers(mcpSvr *server.MCPServer) {
	mcpSvr.AddTool(
		mcp.NewTool("list_mcp_servers",
			mcp.WithDescription("此工具用于查询 MCP Server 列表，支持按名称、命名空间、业务线、部门、协议等条件过滤"),
			mcp.WithString("name",
				mcp.Description("MCP Server 名称过滤"),
			),
			mcp.WithString("namespace",
				mcp.Description("MCP Server 命名空间过滤"),
			),
			mcp.WithString("business",
				mcp.Description("业务线过滤"),
			),
			mcp.WithString("department",
				mcp.Description("部门过滤"),
			),
			mcp.WithString("protocol",
				mcp.Description("协议类型过滤"),
			),
			mcp.WithNumber("offset",
				mcp.Description("查询的偏移量"),
				mcp.DefaultNumber(0),
			),
			mcp.WithNumber("limit",
				mcp.Description("限制返回的 MCP Server 数量"),
				mcp.DefaultNumber(100),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] handleQueryMCPServers", zap.Any("params", req.Params))
			args := req.Params.Arguments
			if args == nil {
				return mcp.NewToolResultError("invalid: args is empty"), nil
			}

			filter := make(map[string]string)

			if name, ok := args["name"].(string); ok && name != "" {
				filter["name"] = name
			}
			if namespace, ok := args["namespace"].(string); ok && namespace != "" {
				filter["namespace"] = namespace
			}
			if business, ok := args["business"].(string); ok && business != "" {
				filter["business"] = business
			}
			if department, ok := args["department"].(string); ok && department != "" {
				filter["department"] = department
			}
			if protocol, ok := args["protocol"].(string); ok && protocol != "" {
				filter["protocol"] = protocol
			}

			searchOffset, _ := args["offset"].(float64)
			searchLimit, _ := args["limit"].(float64)

			rsp := h.mcpServerQuery(ctx, filter, uint32(searchOffset), uint32(searchLimit))
			if !api.IsSuccess(rsp) {
				return mcp.NewToolResultError(rsp.GetInfo()), nil
			}

			ret, err := httpcommon.MarshalPBJson(rsp)
			return mcp.NewToolResultText(ret), err
		})
}

// addToolCreateMCPServers 添加创建 MCP Servers 的 tool
func (h *HTTPServer) addToolCreateMCPServers(mcpSvr *server.MCPServer) {
	mcpSvr.AddTool(
		mcp.NewTool("create_mcp_servers",
			mcp.WithDescription("此工具用于创建多个 MCP Server"),
			mcp.WithArray("servers",
				mcp.Description("MCP Server 数组"),
				mcp.Items(httpcommon.MarshalPBJsonToMap(&ai.MCPServer{})),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] handleCreateMCPServers", zap.Any("params", req.Params))
			args := req.Params.Arguments
			if args == nil {
				return mcp.NewToolResultError("invalid: args is empty"), nil
			}

			servers, ok := args["servers"].([]interface{})
			if !ok || len(servers) == 0 {
				return mcp.NewToolResultError("invalid: servers is empty"), nil
			}

			reqs, err := httpcommon.UnmarshalArray(json.NewDecoder(bytes.NewBuffer(jsonMarshal(servers))),
				func() *ai.MCPServer { return &ai.MCPServer{} })
			if err != nil {
				log.Error("[apiserver][ai-mcp] handleCreateMCPServers", zap.Error(err))
				return mcp.NewToolResultError("invalid: servers parse fail: " + err.Error()), nil
			}

			rsp := h.mcpServerCreate(ctx, reqs)
			if !api.IsSuccess(rsp) {
				return mcp.NewToolResultError(rsp.GetInfo()), nil
			}

			ret, err := httpcommon.MarshalPBJson(rsp)
			return mcp.NewToolResultText(ret), err
		})
}

// addToolUpdateMCPServers 添加更新 MCP Servers 的 tool
func (h *HTTPServer) addToolUpdateMCPServers(mcpSvr *server.MCPServer) {
	mcpSvr.AddTool(
		mcp.NewTool("update_mcp_servers",
			mcp.WithDescription("此工具用于更新多个 MCP Server"),
			mcp.WithArray("servers",
				mcp.Description("MCP Server 数组"),
				mcp.Items(httpcommon.MarshalPBJsonToMap(&ai.MCPServer{})),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] handleUpdateMCPServers", zap.Any("params", req.Params))
			args := req.Params.Arguments
			if args == nil {
				return mcp.NewToolResultError("invalid: args is empty"), nil
			}

			servers, ok := args["servers"].([]*ai.MCPServer)
			if !ok || len(servers) == 0 {
				return mcp.NewToolResultError("invalid: servers is empty or invalid"), nil
			}

			rsp := h.mcpServerUpdate(ctx, servers)
			if !api.IsSuccess(rsp) {
				return mcp.NewToolResultError(rsp.GetInfo()), nil
			}

			ret, err := httpcommon.MarshalPBJson(rsp)
			return mcp.NewToolResultText(ret), err
		})
}

// addToolDeleteMCPServers 添加删除 MCP Servers 的 tool
func (h *HTTPServer) addToolDeleteMCPServers(mcpSvr *server.MCPServer) {
	mcpSvr.AddTool(
		mcp.NewTool("delete_mcp_servers",
			mcp.WithDescription("此工具用于删除多个 MCP Server"),
			mcp.WithArray("server_ids",
				mcp.Description("MCP Server ID 数组"),
				mcp.Items(map[string]interface{}{
					"id": "",
				}),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] handleDeleteMCPServers", zap.Any("params", req.Params))
			args := req.Params.Arguments
			if args == nil {
				return mcp.NewToolResultError("invalid: args is empty"), nil
			}

			serverIDs, ok := args["server_ids"].([]interface{})
			if !ok || len(serverIDs) == 0 {
				return mcp.NewToolResultError("invalid: server_ids is empty"), nil
			}

			var ids []string
			for _, v := range serverIDs {
				if id, ok := v.(string); ok {
					ids = append(ids, id)
				}
			}

			rsp := h.mcpServerDelete(ctx, ids)
			if !api.IsSuccess(rsp) {
				return mcp.NewToolResultError(rsp.GetInfo()), nil
			}

			ret, err := httpcommon.MarshalPBJson(rsp)
			return mcp.NewToolResultText(ret), err
		})
}

// addToolQueryMCPServerTools 添加查询 MCP Server Tools 的 tool
func (h *HTTPServer) addToolQueryMCPServerTools(mcpSvr *server.MCPServer) {
	mcpSvr.AddTool(
		mcp.NewTool("list_mcp_server_tools",
			mcp.WithDescription("此工具用于查询 MCP Server 的 Tools 列表"),
			mcp.WithString("server_id",
				mcp.Description("MCP Server ID"),
			),
			mcp.WithString("server_name",
				mcp.Description("MCP Server 名称"),
			),
			mcp.WithString("server_namespace",
				mcp.Description("MCP Server 命名空间"),
			),
			mcp.WithNumber("offset",
				mcp.Description("查询的偏移量"),
				mcp.DefaultNumber(0),
			),
			mcp.WithNumber("limit",
				mcp.Description("限制返回的 Tools 数量"),
				mcp.DefaultNumber(100),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] handleQueryMCPServerTools", zap.Any("params", req.Params))
			args := req.Params.Arguments
			if args == nil {
				return mcp.NewToolResultError("invalid: args is empty"), nil
			}

			filter := make(map[string]string)

			if serverID, ok := args["server_id"].(string); ok && serverID != "" {
				filter["server_id"] = serverID
			}
			if serverName, ok := args["server_name"].(string); ok && serverName != "" {
				filter["server_name"] = serverName
			}
			if serverNamespace, ok := args["server_namespace"].(string); ok && serverNamespace != "" {
				filter["server_namespace"] = serverNamespace
			}

			searchOffset, _ := args["offset"].(float64)
			searchLimit, _ := args["limit"].(float64)

			rsp := h.mcpServerToolQuery(ctx, filter, uint32(searchOffset), uint32(searchLimit))
			if !api.IsSuccess(rsp) {
				return mcp.NewToolResultError(rsp.GetInfo()), nil
			}

			ret, err := httpcommon.MarshalPBJson(rsp)
			return mcp.NewToolResultText(ret), err
		})
}

// addToolsMCPServer 添加 MCP Server 相关的 tools
func (h *HTTPServer) addToolsMCPServer(mcpSvr *server.MCPServer) {
	h.addToolQueryMCPServers(mcpSvr)
	h.addToolCreateMCPServers(mcpSvr)
	h.addToolUpdateMCPServers(mcpSvr)
	h.addToolDeleteMCPServers(mcpSvr)
	h.addToolQueryMCPServerTools(mcpSvr)
}

// mcpServerQuery 查询 MCP Servers
func (h *HTTPServer) mcpServerQuery(ctx context.Context, filter map[string]string, offset, limit uint32) *api.BatchQueryResponse {
	// TODO: 实现 MCP Server 查询逻辑
	return api.NewBatchQueryResponse(api.ExecuteSuccess)
}

// mcpServerCreate 创建 MCP Servers
func (h *HTTPServer) mcpServerCreate(ctx context.Context, servers []*ai.MCPServer) api.ResponseMessage {
	// TODO: 实现 MCP Server 创建逻辑
	return api.NewResponse(api.ExecuteSuccess)
}

// mcpServerUpdate 更新 MCP Servers
func (h *HTTPServer) mcpServerUpdate(ctx context.Context, servers []*ai.MCPServer) api.ResponseMessage {
	// TODO: 实现 MCP Server 更新逻辑
	return api.NewResponse(api.ExecuteSuccess)
}

// mcpServerDelete 删除 MCP Servers
func (h *HTTPServer) mcpServerDelete(ctx context.Context, ids []string) api.ResponseMessage {
	// TODO: 实现 MCP Server 删除逻辑
	return api.NewResponse(api.ExecuteSuccess)
}

// mcpServerToolQuery 查询 MCP Server Tools
func (h *HTTPServer) mcpServerToolQuery(ctx context.Context, filter map[string]string, offset, limit uint32) *api.BatchQueryResponse {
	// TODO: 实现 MCP Server Tool 查询逻辑
	return api.NewBatchQueryResponse(api.ExecuteSuccess)
}

// jsonMarshal 序列化 interface{} 为 JSON
func jsonMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

// bytes.NewReader 包装
func bytesNewBuffer(data []byte) *bytes.Buffer {
	return bytes.NewBuffer(data)
}
