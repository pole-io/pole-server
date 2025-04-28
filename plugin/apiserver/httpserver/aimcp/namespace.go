package aimcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	"github.com/polarismesh/specification/source/go/api/v1/service_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

func (h *HTTPServer) addToolsNamespace(mcpSvr *server.MCPServer) {
	h.addToolQueryNamespaces(mcpSvr)
	h.addToolCreateNamespaces(mcpSvr)
	h.addToolUpdateNamespaces(mcpSvr)
	h.addToolDeleteNamespaces(mcpSvr)
}

// 命名空间相关的 tools
func (h *HTTPServer) addToolQueryNamespaces(mcpSvr *server.MCPServer) {
	mcpSvr.AddTool(
		mcp.NewTool("list_namespaces",
			mcp.WithDescription("此工具用于查询服务治理中心下的命名空间列表，可以查询所有命名空间，也可以根据名称进行模糊查询或者进行分页查询"),
			mcp.WithString("name",
				mcp.Description("根据名称查询过滤，支持模糊查询，后缀查询：*test、前缀查询：test*、全模糊查询：*test*"),
			),
			mcp.WithNumber("offset",
				mcp.Description("查询的偏移量，默认值为0"),
				mcp.DefaultNumber(0),
			),
			mcp.WithNumber("limit",
				mcp.Description("限制返回的命名空间数量"),
				mcp.DefaultNumber(100),
			),
			mcp.WithBoolean("all",
				mcp.Description("是否查询所有命名空间"),
				mcp.DefaultBool(false),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] handleQueryNamespaces", zap.Any("params", req.Params))
			args := req.Params.Arguments
			if args == nil {
				return mcp.NewToolResultError("invalid: args is empty"), nil
			}

			var rsp *service_manage.BatchQueryResponse

			searchName, _ := args["name"].(string)
			searchAll, _ := args["all"].(bool)

			if searchAll {
				offset := uint32(0)
				limit := uint32(100)
				rsp = api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
				for {
					filter := map[string][]string{
						"offset": {strconv.Itoa(int(offset))},
						"limit":  {strconv.Itoa(int(limit))},
					}
					if searchName != "" {
						filter["name"] = []string{searchName}
					}
					subRsp := h.namespaceServer.GetNamespaces(ctx, filter)
					if !api.IsSuccess(subRsp) {
						return mcp.NewToolResultError(subRsp.GetInfo().GetValue()), nil
					}
					if len(subRsp.GetNamespaces()) == 0 {
						break
					}
					rsp.Namespaces = append(rsp.Namespaces, subRsp.GetNamespaces()...)
					offset += limit
				}
			} else {
				searchOffset, _ := args["offset"].(uint32)
				searchLimit, _ := args["limit"].(uint32)
				filters := make(map[string][]string)

				if searchName != "" {
					filters["name"] = []string{searchName}
				}
				if searchOffset != 0 {
					filters["offset"] = []string{strconv.Itoa(int(searchOffset))}
				}
				if searchLimit != 0 {
					filters["limit"] = []string{strconv.Itoa(int(searchLimit))}
				}
				rsp = h.namespaceServer.GetNamespaces(ctx, filters)
			}

			if !api.IsSuccess(rsp) {
				return mcp.NewToolResultError(rsp.GetInfo().GetValue()), nil
			}
			ret, err := httpcommon.MarshalPBJson(rsp)
			return mcp.NewToolResultText(ret), err
		})
}

func (h *HTTPServer) addToolCreateNamespaces(mcpSvr *server.MCPServer) {
	mcpSvr.AddTool(
		mcp.NewTool("create_namespaces",
			mcp.WithDescription("此工具用于在服务治理中心下的创建多个命名空间"),
			mcp.WithArray("namespaces",
				mcp.Description("命名空间数组"),
				mcp.Items(httpcommon.MarshalPBJsonToMap(&apimodel.Namespace{})),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] handleCreateNamespaces", zap.Any("params", req.Params))
			args := req.Params.Arguments
			if args == nil {
				return mcp.NewToolResultError("invalid: args is empty"), nil
			}

			namespaces, ok := args["namespaces"].([]interface{})
			if !ok || len(namespaces) == 0 {
				return mcp.NewToolResultError("invalid: namespaces is empty"), nil
			}
			d, _ := json.Marshal(namespaces)
			reqs, err := httpcommon.UnmarshalArray(json.NewDecoder(bytes.NewReader(d)),
				func() *apimodel.Namespace { return &apimodel.Namespace{} })
			if err != nil {
				log.Error("[apiserver][ai-mcp] handleCreateNamespaces", zap.Error(err))
				return mcp.NewToolResultError("invalid: namespaces parse fail: " + err.Error()), nil
			}
			rsp := h.namespaceServer.CreateNamespaces(ctx, reqs)
			if !api.IsSuccess(rsp) {
				return mcp.NewToolResultError(rsp.GetInfo().GetValue()), nil
			}

			ret, err := httpcommon.MarshalPBJson(rsp)
			return mcp.NewToolResultText(ret), err
		})
}

func (h *HTTPServer) addToolDeleteNamespaces(mcpSvr *server.MCPServer) {
	mcpSvr.AddTool(
		mcp.NewTool("delete_namespaces",
			mcp.WithDescription("此工具用于在服务治理中心下的删除多个命名空间"),
			mcp.WithArray("namespaces",
				mcp.Description("命名空间数组"),
				mcp.Items(map[string]interface{}{
					"name": "",
				}),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] handleDeleteNamespaces", zap.Any("params", req.Params))
			args := req.Params.Arguments
			if args == nil {
				return mcp.NewToolResultError("invalid: args is empty"), nil
			}

			namespaces, ok := args["namespaces"].([]*apimodel.Namespace)
			if !ok || len(namespaces) == 0 {
				return mcp.NewToolResultError("invalid: namespaces is empty or invalid"), nil
			}

			rsp := h.namespaceServer.DeleteNamespaces(ctx, namespaces)
			if !api.IsSuccess(rsp) {
				return mcp.NewToolResultError(rsp.GetInfo().GetValue()), nil
			}

			ret, err := httpcommon.MarshalPBJson(rsp)
			return mcp.NewToolResultText(ret), err
		})
}

func (h *HTTPServer) addToolUpdateNamespaces(mcpSvr *server.MCPServer) {
	mcpSvr.AddTool(
		mcp.NewTool("update_namespaces",
			mcp.WithDescription("此工具用于在服务治理中心下的更新多个命名空间"),
			mcp.WithArray("namespaces",
				mcp.Description("命名空间数组"),
				mcp.Items(httpcommon.MarshalPBJsonToMap(&apimodel.Namespace{})),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Info("[apiserver][ai-mcp] handleUpdatwNamespaces", zap.Any("params", req.Params))
			args := req.Params.Arguments
			if args == nil {
				return mcp.NewToolResultError("invalid: args is empty"), nil
			}

			namespaces, ok := args["namespaces"].([]*apimodel.Namespace)
			if !ok || len(namespaces) == 0 {
				return mcp.NewToolResultError("invalid: namespaces is empty or invalid"), nil
			}

			rsp := h.namespaceServer.DeleteNamespaces(ctx, namespaces)
			if !api.IsSuccess(rsp) {
				return mcp.NewToolResultError(rsp.GetInfo().GetValue()), nil
			}

			ret, err := httpcommon.MarshalPBJson(rsp)
			return mcp.NewToolResultText(ret), err
		})
}
