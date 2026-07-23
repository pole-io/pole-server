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
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	configapi "github.com/pole-io/pole-server/pkg/config"
)

func TestConfigFileArguments(t *testing.T) {
	namespace, group, name, result := exactConfigFileArguments(map[string]interface{}{
		"namespace": " default ",
		"group":     " application ",
		"name":      " app.yaml ",
	})
	require.Nil(t, result)
	assert.Equal(t, "default", namespace)
	assert.Equal(t, "application", group)
	assert.Equal(t, "app.yaml", name)

	_, _, _, result = exactConfigFileArguments(map[string]interface{}{"namespace": "default"})
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestConfigFileSearchFilters_HandlesJSONNumbersAndLimitsPageSize(t *testing.T) {
	filters, result := configFileSearchFilters(map[string]interface{}{
		"namespace": " default ",
		"group":     " app* ",
		"offset":    float64(10),
		"limit":     float64(500),
	})
	require.Nil(t, result)
	assert.Equal(t, map[string]string{
		"namespace": "default",
		"group":     "app*",
		"offset":    "10",
		"limit":     "100",
	}, filters)
}

func TestConfigFileTools_RegisterAndGetConfigFile(t *testing.T) {
	mcpServer := server.NewMCPServer("pole-test", "1.0.0")
	h := &HTTPServer{configServer: &testConfigCenterServer{}}
	h.addToolsConfigFile(mcpServer)

	testServer := server.NewTestServer(mcpServer)
	defer testServer.Close()

	mcpClient, err := client.NewSSEMCPClient(testServer.URL + "/sse")
	require.NoError(t, err)
	defer mcpClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, mcpClient.Start(ctx))

	initialize := mcp.InitializeRequest{}
	initialize.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initialize.Params.ClientInfo = mcp.Implementation{Name: "pole-test-client", Version: "1.0.0"}
	_, err = mcpClient.Initialize(ctx, initialize)
	require.NoError(t, err)

	tools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	require.NoError(t, err)
	require.Len(t, tools.Tools, 2)
	assert.Equal(t, "get_config_file", tools.Tools[0].Name)
	assert.Equal(t, []string{"namespace", "group", "name"}, tools.Tools[0].InputSchema.Required)
	assert.Equal(t, "search_config_files", tools.Tools[1].Name)

	call := mcp.CallToolRequest{}
	call.Params.Name = "get_config_file"
	call.Params.Arguments = map[string]interface{}{
		"namespace": "default",
		"group":     "application",
		"name":      "app.yaml",
	}
	result, err := mcpClient.CallTool(ctx, call)
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, text.Text, "server:")
	assert.Contains(t, text.Text, "app.yaml")
}

type testConfigCenterServer struct {
	configapi.ConfigCenterServer
}

func (t *testConfigCenterServer) GetConfigFileRichInfo(
	_ context.Context,
	req *apiconfig.ConfigFile,
) *apimodel.Response {
	return api.NewConfigFileResponse(apimodel.Code_ExecuteSuccess, &apiconfig.ConfigFile{
		Namespace: req.GetNamespace(),
		Group:     req.GetGroup(),
		Name:      req.GetName(),
		Content:   "server:\n  port: 8080\n",
	})
}
