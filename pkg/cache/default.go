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

package cache

import (
	"context"
	"errors"
	"sync"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/store"
	cacheauth "github.com/pole-io/pole-server/pkg/cache/auth"
	cacheclient "github.com/pole-io/pole-server/pkg/cache/client"
	cacheconfig "github.com/pole-io/pole-server/pkg/cache/config"
	cachegray "github.com/pole-io/pole-server/pkg/cache/gray"
	cachens "github.com/pole-io/pole-server/pkg/cache/namespace"
	cacherules "github.com/pole-io/pole-server/pkg/cache/rules"
	cachesvc "github.com/pole-io/pole-server/pkg/cache/service"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
)

func init() {
	RegisterCache(cacheapi.NamespaceName, cacheapi.CacheNamespace)
	RegisterCache(cacheapi.ServiceName, cacheapi.CacheService)
	RegisterCache(cacheapi.InstanceName, cacheapi.CacheInstance)
	RegisterCache(cacheapi.RoutingConfigName, cacheapi.CacheRoutingConfig)
	RegisterCache(cacheapi.RateLimitConfigName, cacheapi.CacheRateLimit)
	RegisterCache(cacheapi.FaultDetectRuleName, cacheapi.CacheFaultDetector)
	RegisterCache(cacheapi.CircuitBreakerName, cacheapi.CacheCircuitBreaker)
	RegisterCache(cacheapi.ConfigFileCacheName, cacheapi.CacheConfigFile)
	RegisterCache(cacheapi.ConfigGroupCacheName, cacheapi.CacheConfigGroup)
	RegisterCache(cacheapi.UsersName, cacheapi.CacheUser)
	RegisterCache(cacheapi.StrategyRuleName, cacheapi.CacheAuthStrategy)
	RegisterCache(cacheapi.ClientName, cacheapi.CacheClient)
	RegisterCache(cacheapi.ServiceContractName, cacheapi.CacheServiceContract)
	RegisterCache(cacheapi.GrayName, cacheapi.CacheGray)
	RegisterCache(cacheapi.LaneRuleName, cacheapi.CacheLaneRule)
	RegisterCache(cacheapi.RolesName, cacheapi.CacheRole)
}

var (
	cacheMgn   *CacheManager
	once       sync.Once
	finishInit bool
)

// Initialize 初始化
func Initialize(ctx context.Context, cacheOpt *Config, storage store.Store) error {
	var err error
	once.Do(func() {
		err = initialize(ctx, cacheOpt, storage)
	})

	if err != nil {
		return err
	}

	finishInit = true
	return nil
}

// initialize cache 初始化
func initialize(ctx context.Context, cacheOpt *Config, storage store.Store) error {
	var err error
	cacheMgn, err = newCacheManager(ctx, cacheOpt, storage)
	return err
}

func newCacheManager(ctx context.Context, cacheOpt *Config, storage store.Store) (*CacheManager, error) {
	SetCacheConfig(cacheOpt)
	mgr := &CacheManager{
		storage:  storage,
		caches:   make([]cacheapi.Cache, cacheapi.CacheLast),
		needLoad: container.NewSyncSet[string](),
	}

	// 命名空间缓存
	mgr.RegisterCacher(cacheapi.CacheNamespace, cachens.NewNamespaceCache(storage, mgr))
	// 注册发现缓存
	mgr.RegisterCacher(cacheapi.CacheService, cachesvc.NewServiceCache(storage, mgr))
	mgr.RegisterCacher(cacheapi.CacheInstance, cachesvc.NewInstanceCache(storage, mgr))
	// 治理规则缓存
	mgr.RegisterCacher(cacheapi.CacheServiceContract, cachesvc.NewServiceContractCache(storage, mgr))
	mgr.RegisterCacher(cacheapi.CacheRoutingConfig, cacherules.NewRouteRuleCache(storage, mgr))
	mgr.RegisterCacher(cacheapi.CacheRateLimit, cacherules.NewRateLimitCache(storage, mgr))
	mgr.RegisterCacher(cacheapi.CacheCircuitBreaker, cacherules.NewCircuitBreakerCache(storage, mgr))
	mgr.RegisterCacher(cacheapi.CacheFaultDetector, cacherules.NewFaultDetectCache(storage, mgr))
	mgr.RegisterCacher(cacheapi.CacheLaneRule, cacherules.NewLaneCache(storage, mgr))
	// 配置分组 & 配置发布缓存
	mgr.RegisterCacher(cacheapi.CacheConfigFile, cacheconfig.NewConfigFileCache(storage, mgr))
	mgr.RegisterCacher(cacheapi.CacheConfigGroup, cacheconfig.NewConfigGroupCache(storage, mgr))
	// 用户/用户组 & 鉴权规则缓存
	mgr.RegisterCacher(cacheapi.CacheUser, cacheauth.NewUserCache(storage, mgr))
	mgr.RegisterCacher(cacheapi.CacheAuthStrategy, cacheauth.NewStrategyCache(storage, mgr))
	mgr.RegisterCacher(cacheapi.CacheRole, cacheauth.NewRoleCache(storage, mgr))
	// pole-serverSDK Client
	mgr.RegisterCacher(cacheapi.CacheClient, cacheclient.NewClientCache(storage, mgr))
	// 灰度规则
	mgr.RegisterCacher(cacheapi.CacheGray, cachegray.NewGrayCache(storage, mgr))

	if len(mgr.caches) != int(cacheapi.CacheLast) {
		return nil, errors.New("some Cache implement not loaded into CacheManager")
	}

	if err := mgr.Initialize(); err != nil {
		return nil, err
	}
	return mgr, nil
}

func Run(cacheMgr *CacheManager, ctx context.Context) error {
	if startErr := cacheMgr.Start(ctx); startErr != nil {
		log.Errorf("[Cache][Server] start cache err: %s", startErr.Error())
		return startErr
	}

	return nil
}

// GetCacheManager
func GetCacheManager() (*CacheManager, error) {
	if !finishInit {
		return nil, errors.New("cache has not done Initialize")
	}
	return cacheMgn, nil
}
