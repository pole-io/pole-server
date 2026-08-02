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

package store

import (
	"errors"
	"fmt"
	"sync"

	"github.com/pole-io/pole-server/pluginapi"
)

// Config Store的通用配置
type Config struct {
	Name   string
	Option map[string]interface{}
}

var (
	// StoreSlots 已弃用，仅保留旧注册入口的实例投影。
	StoreSlots = make(map[string]Store)
	slotsMu    sync.Mutex

	initMu              sync.Mutex
	config              = &Config{}
	initializedRegistry *pluginapi.Registry
	initializedName     string
	initErr             error
)

// RegisterStore 注册一个新的Store
func RegisterStore(s Store) error {
	if s == nil {
		return errors.New("store is nil")
	}
	name := s.Name()
	if err := RegisterStoreFactory(pluginapi.DefaultRegistry(), pluginapi.Descriptor{
		Kind:   pluginapi.KindStore,
		Name:   name,
		Origin: pluginapi.OriginLegacy,
	}, func() (Store, error) {
		return s, nil
	}); err != nil {
		return err
	}
	slotsMu.Lock()
	StoreSlots[name] = s
	slotsMu.Unlock()
	return nil
}

type Factory func() (Store, error)

func RegisterStoreFactory(registry *pluginapi.Registry, descriptor pluginapi.Descriptor,
	factory Factory) error {
	if factory == nil {
		return fmt.Errorf("store factory is nil: name=%s", descriptor.Name)
	}
	descriptor.Kind = pluginapi.KindStore
	return registry.Register(descriptor, func() (any, error) {
		store, err := factory()
		if err != nil {
			return nil, err
		}
		if store == nil {
			return nil, fmt.Errorf("store factory returned nil: name=%s", descriptor.Name)
		}
		if descriptor.Name != store.Name() {
			return nil, fmt.Errorf("store name mismatch: registered=%s actual=%s",
				descriptor.Name, store.Name())
		}
		return store, nil
	})
}

func ResolveStore(name string) (Store, error) {
	return pluginapi.ResolveAs[Store](pluginapi.ActiveRegistry(), pluginapi.KindStore, name)
}

// GetStore 获取Store
func GetStore() (Store, error) {
	name := config.Name
	if name == "" {
		return nil, errors.New("store name is empty")
	}

	store, err := ResolveStore(name)
	if err != nil {
		return nil, fmt.Errorf("resolve store %q: %w", name, err)
	}

	if err := initialize(name, store); err != nil {
		return nil, err
	}
	return store, nil
}

// SetStoreConfig 设置store的conf
func SetStoreConfig(conf *Config) {
	config = conf
}

func GetStoreConfig() *Config {
	return config
}

// initialize  包裹了初始化函数，在GetStore的时候会在自动调用，全局初始化一次
func initialize(name string, store Store) error {
	initMu.Lock()
	defer initMu.Unlock()
	activeRegistry := pluginapi.ActiveRegistry()
	if initializedRegistry == activeRegistry && initializedName == name {
		return initErr
	}
	initializedRegistry = activeRegistry
	initializedName = name
	initErr = nil
	fmt.Printf("[Store][Info] current use store plugin : %s\n", store.Name())
	if err := store.Initialize(config); err != nil {
		initErr = fmt.Errorf("initialize store %q: %w", store.Name(), err)
	}
	return initErr
}
