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
	"time"

	"golang.org/x/sync/singleflight"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/observability/history"
	"github.com/pole-io/pole-server/apis/pkg/types"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/store"
	cacheservice "github.com/pole-io/pole-server/pkg/cache/service"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/namespace"
	"github.com/pole-io/pole-server/pkg/service/batch"
	"github.com/pole-io/pole-server/pkg/service/healthcheck"
)

// Server 对接API层的server层，用以处理业务逻辑
type Server struct {
	// config 配置
	config Config
	// store 数据存储
	storage store.Store
	// namespaceSvr 命名空间相关的操作
	namespaceSvr namespace.NamespaceOperateServer
	// caches 缓存
	caches cacheapi.CacheManager
	// bc 批量写执行器
	bc *batch.Controller
	// healthCheckServer 健康检查服务
	healthServer *healthcheck.Server
	// createServiceSingle 确保服务创建的并发控制
	createServiceSingle *singleflight.Group
	// subCtxs eventhub 的 subscriber 的控制
	subCtxs []*eventhub.SubscribtionContext
	// instanceChains 实例信息变化回调
	instanceChains []InstanceChain
	// emptyPushProtectSvs 开启了推空保护的服务数据
	emptyPushProtectSvs *container.SyncMap[string, time.Time]
}

func (s *Server) allowAutoCreate() bool {
	if s.config.AutoCreate == nil {
		return true
	}
	return *s.config.AutoCreate
}

func (s *Server) Store() store.Store {
	return s.storage
}

// HealthServer 健康检查Server
func (s *Server) HealthServer() *healthcheck.Server {
	return s.healthServer
}

// Cache 返回Cache
func (s *Server) Cache() cacheapi.CacheManager {
	return s.caches
}

// Namespace 返回NamespaceOperateServer
func (s *Server) Namespace() namespace.NamespaceOperateServer {
	return s.namespaceSvr
}

// RecordHistory server对外提供history插件的简单封装
func (s *Server) RecordHistory(ctx context.Context, entry *types.RecordEntry) {
	// 如果数据为空，则不需要打印了
	if entry == nil {
		return
	}

	fromClient, _ := ctx.Value(types.ContextIsFromClient).(bool)
	if fromClient {
		return
	}
	// 调用插件记录history
	history.GetHistory().Record(entry)
}

// AddInstanceChain not thread safe
func (s *Server) AddInstanceChain(chain ...InstanceChain) {
	s.instanceChains = append(s.instanceChains, chain...)
}

// GetServiceInstanceRevision 获取服务实例的revision
func (s *Server) GetServiceInstanceRevision(serviceID string, instances []*svctypes.Instance) (string, error) {
	if revision := s.caches.Service().GetRevisionWorker().GetServiceInstanceRevision(serviceID); revision != "" {
		return revision, nil
	}

	svc := s.Cache().Service().GetServiceByID(serviceID)
	if svc == nil {
		return "", types.ErrorNoService
	}

	data, err := cacheservice.ComputeRevision(svc.Revision, instances)
	if err != nil {
		return "", err
	}

	return data, nil
}

func AllowAutoCreate(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, utils.ContextKeyAutoCreateService{}, true)
	return ctx
}
