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

package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	cachetypes "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pluginapi"
)

const (
	ContextKeyUserSvr   = "userSvr"
	ContextKeyPolicySvr = "policySvr"
)

const (
	// DefaultUserMgnPluginName default user server name
	DefaultUserMgnPluginName = "defaultUser"
	// DefaultPolicyPluginName default strategy server name
	DefaultPolicyPluginName = "defaultStrategy"
)

// Config 鉴权能力的相关配置参数
type Config struct {
	// Name 原AuthServer名称，已废弃
	Name string
	// Option 原AuthServer的option，已废弃
	// Deprecated
	Option map[string]interface{}
	// User UserOperator的相关配置
	User *UserConfig `yaml:"user"`
	// Strategy StrategyOperator的相关配置
	Strategy *StrategyConfig `yaml:"strategy"`
	// Interceptors .
	Interceptors []string `yaml:"-"`
}

func (c *Config) SetDefault() {
	if c.User == nil {
		c.User = &UserConfig{
			Name:   DefaultUserMgnPluginName,
			Option: map[string]interface{}{},
		}
	}
	if c.Strategy == nil {
		c.Strategy = &StrategyConfig{
			Name:   DefaultPolicyPluginName,
			Option: map[string]interface{}{},
		}
	}
}

// UserConfig UserOperator的相关配置
type UserConfig struct {
	// Name UserOperator的名称
	Name string `yaml:"name"`
	// Option UserOperator的option
	Option map[string]interface{} `yaml:"option"`
}

// StrategyConfig StrategyOperator的相关配置
type StrategyConfig struct {
	// Name StrategyOperator的名称
	Name string `yaml:"name"`
	// Option StrategyOperator的option
	Option map[string]interface{} `yaml:"option"`
}

var (
	initMu       sync.RWMutex
	initBuildMu  sync.Mutex
	initRegistry *pluginapi.Registry
	userMgr      UserServer
	policyMgr    StrategyServer
	finishInit   bool
)

func InjectUserMgr(svr UserServer) {
	initMu.Lock()
	defer initMu.Unlock()
	userMgr = svr
}

func InjectPolicyMgr(svr StrategyServer) {
	initMu.Lock()
	defer initMu.Unlock()
	policyMgr = svr
}

// RegisterUserServer 注册一个新的 UserServer
func RegisterUserServer(s UserServer) error {
	if s == nil {
		return errors.New("UserServer is nil")
	}
	name := s.Name()
	return RegisterUserServerFactory(pluginapi.DefaultRegistry(), pluginapi.Descriptor{
		Kind:   pluginapi.KindAuthUser,
		Name:   name,
		Origin: pluginapi.OriginLegacy,
	}, func() (UserServer, error) {
		return s, nil
	})
}

type UserServerFactory func() (UserServer, error)

func RegisterUserServerFactory(registry *pluginapi.Registry, descriptor pluginapi.Descriptor,
	factory UserServerFactory) error {
	if factory == nil {
		return fmt.Errorf("UserServer factory is nil: name=%s", descriptor.Name)
	}
	descriptor.Kind = pluginapi.KindAuthUser
	return registry.Register(descriptor, func() (any, error) {
		server, err := factory()
		if err != nil {
			return nil, err
		}
		if server == nil {
			return nil, fmt.Errorf("UserServer factory returned nil: name=%s", descriptor.Name)
		}
		if descriptor.Name != server.Name() {
			return nil, fmt.Errorf("UserServer name mismatch: registered=%s actual=%s",
				descriptor.Name, server.Name())
		}
		return server, nil
	})
}

func ResolveUserServer(name string) (UserServer, error) {
	return pluginapi.ResolveAs[UserServer](pluginapi.ActiveRegistry(), pluginapi.KindAuthUser, name)
}

// GetUserServer 获取一个 UserServer
func GetUserServer() (UserServer, error) {
	initMu.RLock()
	defer initMu.RUnlock()
	if !finishInit {
		return nil, errors.New("UserServer has not done Initialize")
	}
	return userMgr, nil
}

// GetUserServerContext 获取一个 UserServer
func GetUserServerContext(ctx context.Context) (UserServer, error) {
	userSvr, err := GetUserServer()
	if err != nil {
		return nil, err
	}
	userSvrVal := ctx.Value(ContextKeyUserSvr)
	if userSvrVal != nil {
		userSvr = userSvrVal.(UserServer)
	}
	return userSvr, nil
}

// RegisterStrategyServer 注册一个新的 StrategyServer
func RegisterStrategyServer(s StrategyServer) error {
	if s == nil {
		return errors.New("StrategyServer is nil")
	}
	name := s.Name()
	return RegisterStrategyServerFactory(pluginapi.DefaultRegistry(), pluginapi.Descriptor{
		Kind:   pluginapi.KindAuthStrategy,
		Name:   name,
		Origin: pluginapi.OriginLegacy,
	}, func() (StrategyServer, error) {
		return s, nil
	})
}

type StrategyServerFactory func() (StrategyServer, error)

func RegisterStrategyServerFactory(registry *pluginapi.Registry, descriptor pluginapi.Descriptor,
	factory StrategyServerFactory) error {
	if factory == nil {
		return fmt.Errorf("StrategyServer factory is nil: name=%s", descriptor.Name)
	}
	descriptor.Kind = pluginapi.KindAuthStrategy
	return registry.Register(descriptor, func() (any, error) {
		server, err := factory()
		if err != nil {
			return nil, err
		}
		if server == nil {
			return nil, fmt.Errorf("StrategyServer factory returned nil: name=%s", descriptor.Name)
		}
		if descriptor.Name != server.Name() {
			return nil, fmt.Errorf("StrategyServer name mismatch: registered=%s actual=%s",
				descriptor.Name, server.Name())
		}
		return server, nil
	})
}

func ResolveStrategyServer(name string) (StrategyServer, error) {
	return pluginapi.ResolveAs[StrategyServer](pluginapi.ActiveRegistry(), pluginapi.KindAuthStrategy, name)
}

// GetStrategyServer 获取一个 StrategyServer
func GetStrategyServer() (StrategyServer, error) {
	initMu.RLock()
	defer initMu.RUnlock()
	if !finishInit {
		return nil, errors.New("StrategyServer has not done Initialize")
	}
	return policyMgr, nil
}

// GetStrategyServerContext 获取一个 UserServer
func GetStrategyServerContext(ctx context.Context) (StrategyServer, error) {
	policySvr, err := GetStrategyServer()
	if err != nil {
		return nil, err
	}
	policySvrVal := ctx.Value(ContextKeyPolicySvr)
	if policySvrVal != nil {
		policySvr = policySvrVal.(StrategyServer)
	}
	return policySvr, nil
}

// Initialize 初始化
func Initialize(ctx context.Context, authOpt *Config, storage store.Store, cacheMgr cachetypes.CacheManager) error {
	initBuildMu.Lock()
	defer initBuildMu.Unlock()

	activeRegistry := pluginapi.ActiveRegistry()
	initMu.RLock()
	if finishInit && initRegistry == activeRegistry {
		initMu.RUnlock()
		return nil
	}
	initMu.RUnlock()

	initMu.Lock()
	finishInit = false
	initMu.Unlock()

	nextUserManager, nextPolicyManager, err := BuildAuthComponent(ctx, authOpt, storage, cacheMgr)
	if err != nil {
		return err
	}

	initMu.Lock()
	userMgr = nextUserManager
	policyMgr = nextPolicyManager
	initRegistry = activeRegistry
	finishInit = true
	initMu.Unlock()
	return nil
}

// BuildAuthComponent 包裹了初始化函数，在 Initialize 的时候会在自动调用，全局初始化一次
func BuildAuthComponent(_ context.Context, authOpt *Config, storage store.Store,
	cacheMgr cachetypes.CacheManager) (UserServer, StrategyServer, error) {
	authOpt.SetDefault()

	userMgrName := authOpt.User.Name
	if userMgrName == "" {
		return nil, nil, errors.New("UserServer Name is empty")
	}
	policyMgrName := authOpt.Strategy.Name
	if policyMgrName == "" {
		return nil, nil, errors.New("StrategyServer Name is empty")
	}

	userMgr, err := ResolveUserServer(userMgrName)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve UserServer plugin %q: %w", userMgrName, err)
	}
	policyMgr, err := ResolveStrategyServer(policyMgrName)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve StrategyServer plugin %q: %w", policyMgrName, err)
	}

	if err := userMgr.Initialize(authOpt, storage, policyMgr, cacheMgr); err != nil {
		log.Printf("UserServer do initialize err: %s", err.Error())
		return nil, nil, err
	}
	if err := policyMgr.Initialize(authOpt, storage, cacheMgr, userMgr); err != nil {
		log.Printf("StrategyServer do initialize err: %s", err.Error())
		return nil, nil, err
	}
	return userMgr, policyMgr, nil
}
