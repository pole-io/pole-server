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

package sqldb

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pole-io/specification/source/go/api/v1/ai"
)

func TestNewMCPID_MatchesSchemaLength(t *testing.T) {
	id := newMCPID()

	assert.Len(t, id, 32)
	assert.False(t, strings.Contains(id, "-"))
}

func TestNormalizeMCPServerBackend_Service(t *testing.T) {
	server := &ai.MCPServer{
		BackendServiceNamespace: "default",
		BackendServiceName:      "order-service",
		BackendAddress:          "http://127.0.0.1:8080/mcp",
	}

	normalizeMCPServerBackend(server)

	assert.Equal(t, "service", server.BackendType)
	assert.Equal(t, "default", server.Namespace)
	assert.Equal(t, "order-service", server.Name)
	assert.Equal(t, "order-service", server.Reference)
	assert.Empty(t, server.BackendAddress)
	assert.NoError(t, validateMCPServerBackend(server))
}

func TestNormalizeMCPServerBackend_Address(t *testing.T) {
	server := &ai.MCPServer{
		Name:                    "external-mcp",
		Namespace:               "default",
		BackendType:             "address",
		BackendServiceNamespace: "default",
		BackendServiceName:      "order-service",
		BackendAddress:          "http://127.0.0.1:8080/mcp",
	}

	normalizeMCPServerBackend(server)

	assert.Equal(t, "address", server.BackendType)
	assert.Empty(t, server.BackendServiceNamespace)
	assert.Empty(t, server.BackendServiceName)
	assert.Equal(t, "http://127.0.0.1:8080/mcp", server.BackendAddress)
	assert.NoError(t, validateMCPServerBackend(server))
}

func TestValidateMCPServerBackend_MissingRequiredFields(t *testing.T) {
	assert.Error(t, validateMCPServerBackend(&ai.MCPServer{BackendType: "service", BackendServiceNamespace: "default"}))
	assert.Error(t, validateMCPServerBackend(&ai.MCPServer{BackendType: "address"}))
	assert.NoError(t, validateMCPServerBackend(&ai.MCPServer{}))
}
