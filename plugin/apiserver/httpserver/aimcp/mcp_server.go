package aimcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

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
				mcp.Description("MCP Server 数组，每个元素包含 name, namespace, ports, business, department, description, protocol, export_to 等字段"),
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
			if err != nil || len(servers) == 0 {
				return mcp.NewToolResultError(fmt.Sprintf("invalid: parse servers fail: %v", err)), nil
			}

			rsp := h.mcpServerCreate(ctx, servers)
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
				mcp.Description("MCP Server 数组，每个元素包含 id, name, namespace, ports, business, department, description, protocol, export_to 等字段"),
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
			if err != nil || len(servers) == 0 {
				return mcp.NewToolResultError(fmt.Sprintf("invalid: parse servers fail: %v", err)), nil
			}

			rsp := h.mcpServerUpdate(ctx, servers)
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
func (h *HTTPServer) mcpServerQuery(ctx context.Context, filter map[string]string, offset, limit uint32) *apimodel.BatchQueryResponse {
	if h.cacheMgr == nil {
		return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
	}

	count, servers := h.cacheMgr.MCPServer().Query(filter, offset, limit)

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
func (h *HTTPServer) mcpServerToolQuery(ctx context.Context, filter map[string]string, offset, limit uint32) *apimodel.BatchQueryResponse {
	if h.cacheMgr == nil {
		return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
	}

	// Resolve server_id from server_name + server_namespace if needed
	serverID := filter["server_id"]
	if serverID == "" {
		serverName := filter["server_name"]
		serverNamespace := filter["server_namespace"]
		if serverName != "" && serverNamespace != "" {
			svr := h.cacheMgr.MCPServer().GetMCPServerByName(serverName, serverNamespace)
			if svr == nil {
				return api.NewBatchQueryResponseWithMsg(apimodel.Code_NotFoundResource, "mcp server not found")
			}
			serverID = svr.ID
		}
	}

	if serverID == "" {
		return api.NewBatchQueryResponseWithMsg(apimodel.Code_BadRequest, "server_id or server_name+server_namespace is required")
	}

	tools := h.cacheMgr.MCPServer().GetMCPServerTools(serverID)

	// Apply offset/limit
	total := uint32(len(tools))
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	result := tools[offset:end]

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

// appendMCPServerToResp 将 ai.MCPServer 通过 JSON → structpb.Struct → anypb.Any 路径塞进 BatchQueryResponse.Data
func appendMCPServerToResp(resp *apimodel.BatchQueryResponse, s *ai.MCPServer) error {
	val, err := mcpObjectToAny(s)
	if err != nil {
		return err
	}
	resp.Data = append(resp.Data, val)
	return nil
}

// appendMCPServerToolToResp 同 appendMCPServerToResp，处理 MCPServerTool
func appendMCPServerToolToResp(resp *apimodel.BatchQueryResponse, t *ai.MCPServerTool) error {
	val, err := mcpObjectToAny(t)
	if err != nil {
		return err
	}
	resp.Data = append(resp.Data, val)
	return nil
}

// mcpObjectToAny 把任意 Go 对象走 JSON 字段映射后包装成 anypb.Any（载荷为 google.protobuf.Struct）
func mcpObjectToAny(obj interface{}) (*anypb.Any, error) {
	raw, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var asMap map[string]interface{}
	if err := json.Unmarshal(raw, &asMap); err != nil {
		return nil, err
	}
	st, err := structpb.NewStruct(asMap)
	if err != nil {
		return nil, err
	}
	return anypb.New(st)
}

// parseMCPServers parses MCP Server objects from MCP tool arguments
func parseMCPServers(raw interface{}) ([]*ai.MCPServer, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshal servers: %w", err)
	}

	var servers []*ai.MCPServer
	if err := json.Unmarshal(data, &servers); err != nil {
		return nil, fmt.Errorf("unmarshal servers: %w", err)
	}
	return servers, nil
}
