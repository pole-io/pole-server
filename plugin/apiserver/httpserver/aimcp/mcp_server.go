package aimcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
	"github.com/pole-io/specification/source/go/api/v1/ai"
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
			mcp.WithString("backend_type",
				mcp.Description("Backend 类型过滤，支持 service 或 address"),
			),
			mcp.WithString("backend_service_namespace",
				mcp.Description("关联 Pole 注册服务的命名空间过滤"),
			),
			mcp.WithString("backend_service_name",
				mcp.Description("关联 Pole 注册服务的服务名过滤"),
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

			query := parseMCPServerQuery(args)
			rsp := h.mcpServerQuery(ctx, query)
			if rsp.GetCode() != uint32(apimodel.Code_ExecuteSuccess) {
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
				mcp.Description("MCP Server 数组。关联 Pole 注册服务时传 backend_type=service、backend_service_namespace、backend_service_name；自定义地址时传 backend_type=address、backend_address"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] handleCreateMCPServers", zap.Any("params", req.Params))
			args := req.Params.Arguments
			if args == nil {
				return mcp.NewToolResultError("invalid: args is empty"), nil
			}

			serversRaw, ok := args["servers"]
			if !ok {
				return mcp.NewToolResultError("invalid: servers is empty"), nil
			}

			servers, err := parseMCPServers(serversRaw)
			if err != nil || len(servers.GetServers()) == 0 {
				return mcp.NewToolResultError(fmt.Sprintf("invalid: parse servers fail: %v", err)), nil
			}

			rsp := h.mcpServerCreate(ctx, servers.GetServers())
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
				mcp.Description("MCP Server 数组。关联 Pole 注册服务时传 backend_type=service、backend_service_namespace、backend_service_name；自定义地址时传 backend_type=address、backend_address"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] handleUpdateMCPServers", zap.Any("params", req.Params))
			args := req.Params.Arguments
			if args == nil {
				return mcp.NewToolResultError("invalid: args is empty"), nil
			}

			serversRaw, ok := args["servers"]
			if !ok {
				return mcp.NewToolResultError("invalid: servers is empty"), nil
			}

			servers, err := parseMCPServers(serversRaw)
			if err != nil || len(servers.GetServers()) == 0 {
				return mcp.NewToolResultError(fmt.Sprintf("invalid: parse servers fail: %v", err)), nil
			}

			rsp := h.mcpServerUpdate(ctx, servers.GetServers())
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
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] handleDeleteMCPServers", zap.Any("params", req.Params))
			args := req.Params.Arguments
			if args == nil {
				return mcp.NewToolResultError("invalid: args is empty"), nil
			}

			deleteReq, err := parseMCPServerDeleteRequest(args)
			if err != nil || len(deleteReq.GetServerIds()) == 0 {
				return mcp.NewToolResultError("invalid: server_ids is empty"), nil
			}

			rsp := h.mcpServerDelete(ctx, deleteReq.GetServerIds())
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

			query := parseMCPServerToolQuery(args)
			rsp := h.mcpServerToolQuery(ctx, query)
			if rsp.GetCode() != uint32(apimodel.Code_ExecuteSuccess) {
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
func (h *HTTPServer) mcpServerQuery(ctx context.Context, query *ai.MCPServerQuery) *apimodel.BatchQueryResponse {
	if h.cacheMgr == nil {
		return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
	}

	count, servers := h.cacheMgr.MCPServer().Query(query)

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = count
	resp.Size = uint32(len(servers))
	for _, s := range servers {
		if err := appendMCPServerToResp(resp, s); err != nil {
			log.Warnf("[apiserver][ai-mcp] marshal mcp server to resp data err: %s", err.Error())
			continue
		}
	}
	return resp
}

// mcpServerCreate 创建 MCP Servers
func (h *HTTPServer) mcpServerCreate(ctx context.Context, servers []*ai.MCPServer) *apimodel.Response {
	if h.storage == nil {
		return api.NewResponse(apimodel.Code_StoreLayerException)
	}

	for _, s := range servers {
		if err := h.storage.CreateMCPServer(s); err != nil {
			log.Errorf("[apiserver][ai-mcp] create mcp server err: %s", err.Error())
			return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
		}
	}
	return api.NewResponse(apimodel.Code_ExecuteSuccess)
}

// mcpServerUpdate 更新 MCP Servers
func (h *HTTPServer) mcpServerUpdate(ctx context.Context, servers []*ai.MCPServer) *apimodel.Response {
	if h.storage == nil {
		return api.NewResponse(apimodel.Code_StoreLayerException)
	}

	for _, s := range servers {
		if err := h.storage.UpdateMCPServer(s); err != nil {
			log.Errorf("[apiserver][ai-mcp] update mcp server err: %s", err.Error())
			return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
		}
	}
	return api.NewResponse(apimodel.Code_ExecuteSuccess)
}

// mcpServerDelete 删除 MCP Servers
func (h *HTTPServer) mcpServerDelete(ctx context.Context, ids []string) *apimodel.Response {
	if h.storage == nil {
		return api.NewResponse(apimodel.Code_StoreLayerException)
	}

	for _, id := range ids {
		if err := h.storage.DeleteMCPServer(id); err != nil {
			log.Errorf("[apiserver][ai-mcp] delete mcp server err: %s", err.Error())
			return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
		}
	}
	return api.NewResponse(apimodel.Code_ExecuteSuccess)
}

// mcpServerToolQuery 查询 MCP Server Tools
func (h *HTTPServer) mcpServerToolQuery(ctx context.Context, query *ai.MCPServerToolQuery) *apimodel.BatchQueryResponse {
	if h.cacheMgr == nil {
		return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
	}
	if query == nil {
		query = &ai.MCPServerToolQuery{}
	}

	// Resolve server_id from server_name + server_namespace if needed
	serverID := query.ServerId
	if serverID == "" {
		serverName := query.ServerName
		serverNamespace := query.ServerNamespace
		if serverName != "" && serverNamespace != "" {
			svr := h.cacheMgr.MCPServer().GetMCPServerByName(serverName, serverNamespace)
			if svr == nil {
				return api.NewBatchQueryResponseWithMsg(apimodel.Code_NotFoundResource, "mcp server not found")
			}
			serverID = svr.Id
		}
	}

	if serverID == "" {
		return api.NewBatchQueryResponseWithMsg(apimodel.Code_BadRequest, "server_id or server_name+server_namespace is required")
	}

	tools := h.cacheMgr.MCPServer().GetMCPServerTools(serverID)

	// Apply offset/limit
	total := uint32(len(tools))
	if query.Offset > total {
		query.Offset = total
	}
	end := query.Offset + query.Limit
	if end > total {
		end = total
	}
	result := tools[query.Offset:end]

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = total
	resp.Size = uint32(len(result))
	for _, t := range result {
		if err := appendMCPServerToolToResp(resp, t); err != nil {
			log.Warnf("[apiserver][ai-mcp] marshal mcp server tool to resp data err: %s", err.Error())
			continue
		}
	}
	return resp
}

// appendMCPServerToResp 将 specification MCPServer 直接包装为 Any，保留 proto 类型信息。
func appendMCPServerToResp(resp *apimodel.BatchQueryResponse, s *ai.MCPServer) error {
	val, err := anypb.New(s)
	if err != nil {
		return err
	}
	resp.Data = append(resp.Data, val)
	return nil
}

// appendMCPServerToolToResp 同 appendMCPServerToResp，处理 MCPServerTool。
func appendMCPServerToolToResp(resp *apimodel.BatchQueryResponse, t *ai.MCPServerTool) error {
	val, err := anypb.New(t)
	if err != nil {
		return err
	}
	resp.Data = append(resp.Data, val)
	return nil
}

// parseMCPServers parses MCP Server objects from MCP tool arguments
func parseMCPServers(raw interface{}) (*ai.MCPServers, error) {
	data, err := json.Marshal(map[string]interface{}{"servers": raw})
	if err != nil {
		return nil, fmt.Errorf("marshal servers: %w", err)
	}

	var servers ai.MCPServers
	if err := json.Unmarshal(data, &servers); err != nil {
		return nil, fmt.Errorf("unmarshal servers: %w", err)
	}
	return &servers, nil
}

func parseMCPServerDeleteRequest(args map[string]interface{}) (*ai.MCPServerDeleteRequest, error) {
	data, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal delete request: %w", err)
	}

	var req ai.MCPServerDeleteRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("unmarshal delete request: %w", err)
	}
	return &req, nil
}

func parseMCPServerQuery(args map[string]interface{}) *ai.MCPServerQuery {
	return &ai.MCPServerQuery{
		Name:                    mcpStringArg(args, "name"),
		Namespace:               mcpStringArg(args, "namespace"),
		Business:                mcpStringArg(args, "business"),
		Department:              mcpStringArg(args, "department"),
		Protocol:                mcpStringArg(args, "protocol"),
		BackendType:             mcpStringArg(args, "backend_type"),
		BackendServiceNamespace: mcpStringArg(args, "backend_service_namespace"),
		BackendServiceName:      mcpStringArg(args, "backend_service_name"),
		Offset:                  mcpUint32Arg(args, "offset", 0),
		Limit:                   mcpUint32Arg(args, "limit", 100),
	}
}

func parseMCPServerToolQuery(args map[string]interface{}) *ai.MCPServerToolQuery {
	return &ai.MCPServerToolQuery{
		ServerId:        mcpStringArg(args, "server_id"),
		ServerName:      mcpStringArg(args, "server_name"),
		ServerNamespace: mcpStringArg(args, "server_namespace"),
		Offset:          mcpUint32Arg(args, "offset", 0),
		Limit:           mcpUint32Arg(args, "limit", 100),
	}
}

func mcpStringArg(args map[string]interface{}, key string) string {
	value, _ := args[key].(string)
	return value
}

func mcpUint32Arg(args map[string]interface{}, key string, defaultValue uint32) uint32 {
	value, ok := args[key]
	if !ok {
		return defaultValue
	}
	switch typed := value.(type) {
	case float64:
		if typed < 0 {
			return defaultValue
		}
		return uint32(typed)
	case int:
		if typed < 0 {
			return defaultValue
		}
		return uint32(typed)
	case uint32:
		return typed
	default:
		return defaultValue
	}
}
