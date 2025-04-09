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

package namespace

import (
	"golang.org/x/sync/singleflight"

	"github.com/pole-io/pole-server/apis/observability/history"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/cache"
)

const (
	// SystemNamespace polaris system namespace
	SystemNamespace = "pole-system"
	// DefaultNamespace default namespace
	DefaultNamespace = "default"
	// ProductionNamespace default namespace
	ProductionNamespace = "Production"
	// DefaultTLL default ttl
	DefaultTLL = 5
)

var _ NamespaceOperateServer = (*Server)(nil)

type Server struct {
	storage               store.Store
	caches                *cache.CacheManager
	createNamespaceSingle *singleflight.Group
	cfg                   Config
}

// RecordHistory server对外提供history插件的简单封装
func (s *Server) RecordHistory(entry *types.RecordEntry) {
	// 如果插件没有初始化，那么不记录history
	if history.GetHistory() == nil {
		return
	}
	// 如果数据为空，则不需要打印了
	if entry == nil {
		return
	}

	// 调用插件记录history
	history.GetHistory().Record(entry)
}
