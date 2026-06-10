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
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/specification/source/go/api/v1/ai"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

func TestParseMCPServerQuery_UsesSpecMessage(t *testing.T) {
	query := parseMCPServerQuery(map[string]interface{}{
		"name":                      "order",
		"namespace":                 "default",
		"business":                  "payment",
		"department":                "infra",
		"protocol":                  "http",
		"backend_type":              "service",
		"backend_service_namespace": "default",
		"backend_service_name":      "order-service",
		"offset":                    float64(2),
		"limit":                     float64(20),
	})

	assert.Equal(t, &ai.MCPServerQuery{
		Name:                    "order",
		Namespace:               "default",
		Business:                "payment",
		Department:              "infra",
		Protocol:                "http",
		BackendType:             "service",
		BackendServiceNamespace: "default",
		BackendServiceName:      "order-service",
		Offset:                  2,
		Limit:                   20,
	}, query)
}

func TestParseMCPServers_UsesSpecContainer(t *testing.T) {
	servers, err := parseMCPServers([]interface{}{
		map[string]interface{}{
			"id":                        "server-1",
			"backend_type":              "service",
			"backend_service_namespace": "default",
			"backend_service_name":      "order",
			"protocol":                  "http",
		},
	})

	require.NoError(t, err)
	require.Len(t, servers.GetServers(), 1)
	assert.Equal(t, "server-1", servers.GetServers()[0].GetId())
	assert.Equal(t, "service", servers.GetServers()[0].GetBackendType())
	assert.Equal(t, "default", servers.GetServers()[0].GetBackendServiceNamespace())
	assert.Equal(t, "order", servers.GetServers()[0].GetBackendServiceName())
}

func TestAppendMCPServerToResp_UsesSpecAny(t *testing.T) {
	resp := &apimodel.BatchQueryResponse{}
	err := appendMCPServerToResp(resp, &ai.MCPServer{
		Id:        "server-1",
		Name:      "order",
		Namespace: "default",
	})

	require.NoError(t, err)
	require.Len(t, resp.GetData(), 1)
	assert.NotContains(t, resp.GetData()[0].GetTypeUrl(), "google.protobuf.Struct")

	var got ai.MCPServer
	require.NoError(t, resp.GetData()[0].UnmarshalTo(&got))
	assert.Equal(t, "server-1", got.GetId())
	assert.Equal(t, "order", got.GetName())
}

func TestAppendMCPServerToolToResp_UsesSpecAny(t *testing.T) {
	resp := &apimodel.BatchQueryResponse{}
	err := appendMCPServerToolToResp(resp, &ai.MCPServerTool{
		Id:          "tool-1",
		McpServerId: "server-1",
		Name:        "query_order",
	})

	require.NoError(t, err)
	require.Len(t, resp.GetData(), 1)
	assert.NotContains(t, resp.GetData()[0].GetTypeUrl(), "google.protobuf.Struct")

	var got ai.MCPServerTool
	require.NoError(t, resp.GetData()[0].UnmarshalTo(&got))
	assert.Equal(t, "tool-1", got.GetId())
	assert.Equal(t, "server-1", got.GetMcpServerId())
}

func TestStartMCPServerCache_UpdatesImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache := &testMCPServerCache{}
	require.NoError(t, startMCPServerCache(ctx, cache, time.Hour))
	assert.Equal(t, int32(1), cache.updateCount.Load())
}

func TestStartMCPServerCache_ReturnsUpdateError(t *testing.T) {
	expected := errors.New("update failed")
	cache := &testMCPServerCache{updateErr: expected}

	err := startMCPServerCache(context.Background(), cache, time.Hour)
	require.ErrorIs(t, err, expected)
	assert.Equal(t, int32(1), cache.updateCount.Load())
}

type testMCPServerCache struct {
	updateCount atomic.Int32
	updateErr   error
}

func (t *testMCPServerCache) Initialize(map[string]interface{}) error { return nil }
func (t *testMCPServerCache) Update() error {
	t.updateCount.Add(1)
	return t.updateErr
}
func (t *testMCPServerCache) Clear() error { return nil }
func (t *testMCPServerCache) Name() string { return "mcpServer" }
func (t *testMCPServerCache) Close() error { return nil }

func (t *testMCPServerCache) GetMCPServerByID(string) *ai.MCPServer { return nil }
func (t *testMCPServerCache) GetMCPServerByName(string, string) *ai.MCPServer {
	return nil
}
func (t *testMCPServerCache) GetMCPServersByNamespace(string) []*ai.MCPServer { return nil }
func (t *testMCPServerCache) GetMCPServerTools(string) []*ai.MCPServerTool    { return nil }
func (t *testMCPServerCache) Query(*ai.MCPServerQuery) (uint32, []*ai.MCPServer) {
	return 0, nil
}
