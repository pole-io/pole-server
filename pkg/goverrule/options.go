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

package goverrule

import (
	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/cache"
	"github.com/pole-io/pole-server/pkg/namespace"
)

/*
   - name: users # Load user and user group data
   - name: strategyRule # Loading the rules of appraisal
   - name: namespace # Load the naming space data
   - name: client # Load Client-SDK instance data
*/

func GetCaches() []cacheapi.ConfigEntry {
	ret := []cacheapi.ConfigEntry{}
	ret = append(ret, governanceCacheEntries...)
	return ret
}

var (
	governanceCacheEntries = []cacheapi.ConfigEntry{
		{
			Name: cacheapi.RoutingConfigName,
		},
		{
			Name: cacheapi.RateLimitConfigName,
		},
		{
			Name: cacheapi.CircuitBreakerName,
		},
		{
			Name: cacheapi.FaultDetectRuleName,
		},
		{
			Name: cacheapi.LaneRuleName,
		},
	}
)

type InitOption func(s *Server)

func WithNamespaceSvr(svr namespace.NamespaceOperateServer) InitOption {
	return func(s *Server) {
		s.namespaceSvr = svr
	}
}

func WithStorage(storage store.Store) InitOption {
	return func(s *Server) {
		s.storage = storage
	}
}

func WithCacheManager(cacheOpt *cache.Config, c cacheapi.CacheManager, entries ...cacheapi.ConfigEntry) InitOption {
	return func(s *Server) {
		log.Infof("[GoverRule][Server] cache is open, can access the client api function")

		openentries := []cacheapi.ConfigEntry{}
		if len(entries) != 0 {
			openentries = append(openentries, entries...)
		} else {
			openentries = append(openentries, governanceCacheEntries...)
		}

		for i := range openentries {
			if _, ok := s.config.Caches[openentries[i].Name]; !ok {
				continue
			}
			openentries[i].Option = s.config.Caches[openentries[i].Name]
		}

		if err := c.OpenResourceCache(openentries...); err != nil {
			log.Errorf("[GoverRule][Server] open cache error: %v", err)
			return
		}
		s.caches = c
	}
}
