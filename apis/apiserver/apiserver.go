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

package apiserver

import (
	"context"
	"fmt"
	"sync"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/pluginapi"
)

const (
	DiscoverAccess    string = "discover"
	RegisterAccess    string = "register"
	HealthcheckAccess string = "healthcheck"
	ConfigAccess      string = "config"
	CreateFileAccess  string = "createfile"
)

// Config API服务器配置
type Config struct {
	Name   string
	Option map[string]interface{}
	API    map[string]APIConfig
}

// APIConfig API配置
type APIConfig struct {
	Enable  bool
	Include []string
}

// Apiserver API服务器接口
type Apiserver interface {
	// GetProtocol API协议名
	GetProtocol() string
	// GetPort API的监听端口
	GetPort() uint32
	// Initialize API初始化逻辑
	Initialize(ctx context.Context, option map[string]interface{}, api map[string]APIConfig) error
	// Run API服务的主逻辑循环
	Run(errCh chan error)
	// Stop 停止API端口监听
	Stop()
	// Restart 重启API
	Restart(option map[string]interface{}, api map[string]APIConfig, errCh chan error) error
}

type EnrichApiserver interface {
	Apiserver
	DebugHandlers() []types.DebugHandler
}

var (
	slotsMu sync.RWMutex

	// Slots 已弃用，仅保留源代码兼容。运行时代码应使用 RuntimeSlots。
	Slots = make(map[string]Apiserver)
)

// Register 注册API服务器
func Register(name string, server Apiserver) error {
	if server == nil {
		return fmt.Errorf("apiserver is nil: name=%s", name)
	}
	slotsMu.Lock()
	defer slotsMu.Unlock()
	if err := RegisterFactory(pluginapi.DefaultRegistry(), pluginapi.Descriptor{
		Kind:   pluginapi.KindAPIServer,
		Name:   name,
		Origin: pluginapi.OriginLegacy,
	}, func() (Apiserver, error) {
		return server, nil
	}); err != nil {
		return err
	}
	Slots[name] = server

	return nil
}

func ReplaceRuntimeSlots(slots map[string]Apiserver) {
	next := cloneSlots(slots)
	slotsMu.Lock()
	Slots = next
	slotsMu.Unlock()
}

func RuntimeSlots() map[string]Apiserver {
	slotsMu.RLock()
	snapshot := cloneSlots(Slots)
	slotsMu.RUnlock()
	return snapshot
}

func cloneSlots(slots map[string]Apiserver) map[string]Apiserver {
	snapshot := make(map[string]Apiserver, len(slots))
	for name, server := range slots {
		snapshot[name] = server
	}
	return snapshot
}

type Factory func() (Apiserver, error)

func RegisterFactory(registry *pluginapi.Registry, descriptor pluginapi.Descriptor,
	factory Factory) error {
	if factory == nil {
		return fmt.Errorf("apiserver factory is nil: name=%s", descriptor.Name)
	}
	descriptor.Kind = pluginapi.KindAPIServer
	return registry.Register(descriptor, func() (any, error) {
		server, err := factory()
		if err != nil {
			return nil, err
		}
		if server == nil {
			return nil, fmt.Errorf("apiserver factory returned nil: name=%s", descriptor.Name)
		}
		return server, nil
	})
}

func Resolve(name string) (Apiserver, error) {
	return pluginapi.ResolveAs[Apiserver](pluginapi.ActiveRegistry(), pluginapi.KindAPIServer, name)
}

func Registered(name string) bool {
	return pluginapi.ActiveRegistry().Contains(pluginapi.KindAPIServer, name)
}
