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

package paramcheck

import (
	"github.com/pole-io/pole-server/apis/access_control/ratelimit"
	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/goverrule"
)

// Server 带有鉴权能力的 discoverServer
//
//	该层会对请求参数做一些调整，根据具体的请求发起人，设置为数据对应的 owner，不可为为别人进行创建资源
type Server struct {
	storage   store.Store
	nextSvr   goverrule.GoverRuleServer
	ratelimit ratelimit.Ratelimit
}

func NewServer(nextSvr goverrule.GoverRuleServer, s store.Store) goverrule.GoverRuleServer {
	proxy := &Server{
		nextSvr: nextSvr,
		storage: s,
	}
	// 获取限流插件
	proxy.ratelimit = ratelimit.GetRatelimit()
	if proxy.ratelimit == nil {
		log.Warnf("Not found Ratelimit Plugin")
	}
	return proxy
}

// Cache Get cache management
func (svr *Server) Cache() cacheapi.CacheManager {
	return svr.nextSvr.Cache()
}
