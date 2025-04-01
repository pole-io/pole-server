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

	"github.com/pole-io/pole-server/apis/cmdb"
	"github.com/pole-io/pole-server/apis/observability/history"
	"github.com/pole-io/pole-server/apis/pkg/types"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/store"
	cachetypes "github.com/pole-io/pole-server/pkg/cache/api"
	cacheservice "github.com/pole-io/pole-server/pkg/cache/service"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	"github.com/pole-io/pole-server/pkg/common/model"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/namespace"
	"github.com/pole-io/pole-server/pkg/service/batch"
	"github.com/pole-io/pole-server/pkg/service/healthcheck"
)

// Server 对接API层的server层，用以处理业务逻辑
type Server struct {
	config Config

	storage store.Store

	namespaceSvr namespace.NamespaceOperateServer

	caches cachetypes.CacheManager
	bc     *batch.Controller

	healthServer *healthcheck.Server

	cmdb    cmdb.CMDB
	history history.History

	createServiceSingle *singleflight.Group
	subCtxs             []*eventhub.SubscribtionContext

	// instanceChains 实例信息变化回调
	instanceChains []InstanceChain
}

func (s *Server) isSupportL5() bool {
	if s.config.L5Open != nil {
		return *s.config.L5Open
	}
	return true
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
func (s *Server) Cache() cachetypes.CacheManager {
	return s.caches
}

// Namespace 返回NamespaceOperateServer
func (s *Server) Namespace() namespace.NamespaceOperateServer {
	return s.namespaceSvr
}

// RecordHistory server对外提供history插件的简单封装
func (s *Server) RecordHistory(ctx context.Context, entry *types.RecordEntry) {
	// 如果插件没有初始化，那么不记录history
	if s.history == nil {
		return
	}
	// 如果数据为空，则不需要打印了
	if entry == nil {
		return
	}

	fromClient, _ := ctx.Value(utils.ContextIsFromClient).(bool)
	if fromClient {
		return
	}
	// 调用插件记录history
	s.history.Record(entry)
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

// 封装一下cmdb的GetLocation
func (s *Server) getLocation(host string) *svctypes.Location {
	if s.cmdb == nil {
		return nil
	}

	location, err := s.cmdb.GetLocation(host)
	if err != nil {
		log.Errorf("[Server] get location(%s) err: %s", host, err.Error())
		return nil
	}
	return location
}

func AllowAutoCreate(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, model.ContextKeyAutoCreateService{}, true)
	return ctx
}
