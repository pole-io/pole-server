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

package service

import (
	"context"

	"golang.org/x/sync/singleflight"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/access_control/auth"
	cacheapi "github.com/pole-io/pole-server/apis/cache"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/cache"
	"github.com/pole-io/pole-server/pkg/namespace"
	"github.com/pole-io/pole-server/pkg/service/batch"
	"github.com/pole-io/pole-server/pkg/service/healthcheck"
)

// GetBatchController .
func (s *Server) GetBatchController() *batch.Controller {
	return s.bc
}

// MockBatchController .
func (s *Server) MockBatchController(bc *batch.Controller) {
	s.bc = bc
}

func TestNewServer(mockStore store.Store, nsSvr namespace.NamespaceOperateServer,
	cacheMgr *cache.CacheManager) *Server {
	return &Server{
		storage:             mockStore,
		namespaceSvr:        nsSvr,
		caches:              cacheMgr,
		createServiceSingle: &singleflight.Group{},
	}
}

// TestInitialize 初始化
func TestInitialize(ctx context.Context, namingOpt *Config, cacheOpt *cache.Config,
	cacheEntries []cacheapi.ConfigEntry, bc *batch.Controller, cacheMgr *cache.CacheManager,
	storage store.Store, namespaceSvr namespace.NamespaceOperateServer,
	healthSvr *healthcheck.Server,
	userMgn auth.UserServer, strategyMgn auth.StrategyServer) (DiscoverServer, DiscoverServer, error) {
	entrites := []cacheapi.ConfigEntry{}
	if len(cacheEntries) != 0 {
		entrites = cacheEntries
	} else {
		entrites = GetRegisterCaches()
	}

	actualSvr, proxySvr, err := InitServer(ctx, namingOpt,
		WithBatchController(bc),
		WithCacheManager(cacheOpt, cacheMgr, entrites...),
		WithHealthCheckSvr(healthSvr),
		WithNamespaceSvr(namespaceSvr),
		WithStorage(storage),
	)
	namingServer = actualSvr
	return proxySvr, namingServer, err
}

// TestSerialCreateInstance .
func (s *Server) TestSerialCreateInstance(
	ctx context.Context, svcId string, req *apiservice.Instance, ins *apiservice.Instance) (
	*svctypes.Instance, *apiservice.Response) {
	return s.serialCreateInstance(ctx, svcId, req, ins)
}

// TestSetStore .
func (s *Server) TestSetStore(storage store.Store) {
	s.storage = storage
}

// TestIsEmptyLocation .
func TestIsEmptyLocation(loc *apimodel.Location) bool {
	return isEmptyLocation(loc)
}
