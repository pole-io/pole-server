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

package ratelimitv2

import (
	"context"
	"sync"

	"github.com/pole-io/pole-server/pkg/limiter/internal/config"
	"github.com/pole-io/pole-server/pkg/limiter/internal/statistics"
)

// Server v2版本的主server逻辑
type Server struct {
	counterMng         *CounterManagerV2
	clientMng          *ClientManager
	cfg                config.Config
	statics            statistics.Statis
	leaseMutex         sync.Mutex
	quotaLeases        map[string]*quotaLease
	leaseByIdempotency map[string]string
	reservedByCounter  map[uint32]uint64
}

// CounterMng 获取计数器管理类
func (s *Server) CounterMng() *CounterManagerV2 {
	return s.counterMng
}

// CleanupClient 清理客户端
func (s *Server) CleanupClient(client Client, streamCtxId string) {
	if s.clientMng.DelClient(client, streamCtxId) {
		client.Cleanup()
	}
}

// NewServer 构造独立的限流核心实例。
func NewServer(ctx context.Context, cfg *config.Config, statics statistics.Statis) (*Server, error) {
	newConfig, err := config.ParseConfig(cfg)
	if err != nil {
		return nil, err
	}
	server := &Server{
		cfg:                *newConfig,
		statics:            statics,
		quotaLeases:        make(map[string]*quotaLease),
		leaseByIdempotency: make(map[string]string),
		reservedByCounter:  make(map[uint32]uint64),
	}
	server.counterMng = NewCounterManagerV2(
		newConfig.MaxCounter, newConfig.PurgeCounterInterval, statics)
	server.clientMng = NewClientManager(newConfig.MaxClient, statics)
	server.counterMng.Start(ctx)
	server.startLeaseCleanup(ctx)
	return server, nil
}
