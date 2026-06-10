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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_LoadsEmbeddedConsoleConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "pole-server.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`
bootstrap:
  mode: all
  console:
    webServer:
      listenIP: 127.0.0.1
      listenPort: 8080
      webPath: console/web/dist/
    poleServer:
      address: 127.0.0.1:8090
`), 0o644))

	cfg, err := Load(cfgPath)

	require.NoError(t, err)
	assert.Equal(t, StartModeAll, cfg.Bootstrap.Mode)
	assert.Equal(t, "127.0.0.1", cfg.Bootstrap.Console.WebServer.ListenIP)
	assert.Equal(t, 8080, cfg.Bootstrap.Console.WebServer.ListenPort)
	assert.Equal(t, "console/web/dist/", cfg.Bootstrap.Console.WebServer.WebPath)
	assert.Equal(t, "127.0.0.1:8090", cfg.Bootstrap.Console.PoleServer.Address)
}

func TestLoad_LoadsDeployDefaultAllConfig(t *testing.T) {
	cfgPath := filepath.Join("..", "..", "deploy", "conf", "pole-server.yaml")

	cfg, err := Load(cfgPath)

	require.NoError(t, err)
	assert.Equal(t, StartModeAll, cfg.Bootstrap.Mode)
	assert.Equal(t, 8080, cfg.Bootstrap.Console.WebServer.ListenPort)
	require.NotNil(t, cfg.Naming.Batch)
	require.NotNil(t, cfg.Naming.Batch["register"])
	require.NotNil(t, cfg.Naming.HealthChecks.Batch)
	require.NotEmpty(t, cfg.Naming.HealthChecks.Checkers)
}

func TestLoadAPIEntries_LoadsDeployAIMCPAccess(t *testing.T) {
	cfgPath := filepath.Join("..", "..", "deploy", "conf", "pole-apiserver.yaml")

	entries, err := LoadAPIEntries(cfgPath)

	require.NoError(t, err)
	for _, entry := range entries {
		if entry.Name != "api-http" {
			continue
		}
		aimcpConfig, ok := entry.API["aimcp"]
		require.True(t, ok)
		assert.True(t, aimcpConfig.Enable)
		assert.Contains(t, aimcpConfig.Include, "default")
		return
	}
	t.Fatal("api-http entry not found")
}
