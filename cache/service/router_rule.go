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
	"fmt"
	"time"

	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	types "github.com/pole-io/pole-server/cache/api"
	"github.com/pole-io/pole-server/common/model"
	"github.com/pole-io/pole-server/common/utils"
	"github.com/pole-io/pole-server/store"
)

type (
	// RouteRuleCache Routing rules cache
	RouteRuleCache struct {
		*types.BaseCache

		serviceCache types.ServiceCache
		storage      store.Store

		container *RouteRuleContainer

		lastMtime time.Time

		singleFlight singleflight.Group
	}
)

// NewRouteRuleCache Return a object of operating RouteRuleCache
func NewRouteRuleCache(s store.Store, cacheMgr types.CacheManager) types.RoutingConfigCache {
	return &RouteRuleCache{
		BaseCache: types.NewBaseCache(s, cacheMgr),
		storage:   s,
	}
}

// initialize The function of implementing the cache interface
func (rc *RouteRuleCache) Initialize(_ map[string]interface{}) error {
	rc.lastMtime = time.Unix(0, 0)
	rc.container = newRouteRuleContainer()
	rc.serviceCache = rc.BaseCache.CacheMgr.GetCacher(types.CacheService).(*serviceCache)
	return nil
}

// Update The function of implementing the cache interface
func (rc *RouteRuleCache) Update() error {
	// Multiple thread competition, only one thread is updated
	_, err, _ := rc.singleFlight.Do(rc.Name(), func() (interface{}, error) {
		return nil, rc.DoCacheUpdate(rc.Name(), rc.realUpdate)
	})
	return err
}

// update The function of implementing the cache interface
func (rc *RouteRuleCache) realUpdate() (map[string]time.Time, int64, error) {
	outV2, err := rc.storage.GetRoutingConfigsV2ForCache(rc.LastFetchTime(), rc.IsFirstUpdate())
	if err != nil {
		log.Errorf("[Cache] routing config v2 cache get from store err: %s", err.Error())
		return nil, -1, err
	}

	lastMtimes := map[string]time.Time{}
	rc.setRouterRules(lastMtimes, outV2)
	rc.container.reload()
	return lastMtimes, int64(len(outV2)), err
}

// Clear The function of implementing the cache interface
func (rc *RouteRuleCache) Clear() error {
	rc.BaseCache.Clear()
	rc.container = newRouteRuleContainer()
	rc.lastMtime = time.Unix(0, 0)
	return nil
}

// Name The function of implementing the cache interface
func (rc *RouteRuleCache) Name() string {
	return types.RoutingConfigName
}

func (rc *RouteRuleCache) ListRouterRule(service, namespace string) []*model.ExtendRouterConfig {
	routerRules := rc.container.SearchCustomRules(service, namespace)
	ret := make([]*model.ExtendRouterConfig, 0, len(routerRules))
	ret = append(ret, routerRules...)
	return ret
}

// GetRouterConfigV2 Obtain routing configuration based on serviceid
func (rc *RouteRuleCache) GetRouterConfigV2(id, service, namespace string) (*apitraffic.Routing, error) {
	if id == "" && service == "" && namespace == "" {
		return nil, nil
	}

	routerRules := rc.container.SearchCustomRules(service, namespace)
	revisions := make([]string, 0, len(routerRules))
	rulesV2 := make([]*apitraffic.RouteRule, 0, len(routerRules))
	for i := range routerRules {
		item := routerRules[i]
		entry, err := item.ToApi()
		if err != nil {
			return nil, err
		}
		rulesV2 = append(rulesV2, entry)
		revisions = append(revisions, entry.GetRevision())
	}
	revision, err := types.CompositeComputeRevision(revisions)
	if err != nil {
		log.Warn("[Cache][Routing] v2=>v1 compute revisions fail, use fake revision", zap.Error(err))
		revision = utils.NewV2Revision()
	}

	resp := &apitraffic.Routing{
		Namespace: utils.NewStringValue(namespace),
		Service:   utils.NewStringValue(service),
		Rules:     rulesV2,
		Revision:  utils.NewStringValue(revision),
	}
	return resp, nil
}

// GetRouterConfig Obtain routing configuration based on serviceid
func (rc *RouteRuleCache) GetRouterConfig(id, svcName, namespace string) (*apitraffic.Routing, error) {
	if id == "" && svcName == "" && namespace == "" {
		return nil, nil
	}

	key := model.ServiceKey{Namespace: namespace, Name: svcName}

	revisions := []string{}
	inRule, inRevision := rc.container.customContainers[model.TrafficDirection_INBOUND].SearchCustomRuleV1(key)
	revisions = append(revisions, inRevision...)
	outRule, outRevision := rc.container.customContainers[model.TrafficDirection_OUTBOUND].SearchCustomRuleV1(key)
	revisions = append(revisions, outRevision...)

	revision, err := types.CompositeComputeRevision(revisions)
	if err != nil {
		log.Warn("[Cache][Routing] v2=>v1 compute revisions fail, use fake revision", zap.Error(err))
		revision = utils.NewV2Revision()
	}

	return &apitraffic.Routing{
		Namespace: utils.NewStringValue(namespace),
		Service:   utils.NewStringValue(svcName),
		Inbounds:  inRule.Inbounds,
		Outbounds: outRule.Outbounds,
		Revision:  utils.NewStringValue(revision),
	}, nil
}

// GetNearbyRouteRule 根据服务名查询就近路由数据
func (rc *RouteRuleCache) GetNearbyRouteRule(service, namespace string) ([]*apitraffic.RouteRule, string, error) {
	if service == "" && namespace == "" {
		return nil, "", nil
	}

	svcKey := model.ServiceKey{
		Namespace: namespace,
		Name:      service,
	}

	routerRules := rc.container.nearbyContainers.SearchRouteRuleV2(svcKey)
	revisions := make([]string, 0, len(routerRules))
	ret := make([]*apitraffic.RouteRule, 0, len(routerRules))
	for i := range routerRules {
		item := routerRules[i]
		entry, err := item.ToApi()
		if err != nil {
			return nil, "", err
		}
		ret = append(ret, entry)
		revisions = append(revisions, entry.GetRevision())
	}
	revision, err := types.CompositeComputeRevision(revisions)
	if err != nil {
		log.Warn("[Cache][Routing] v2=>v1 compute revisions fail, use fake revision", zap.Error(err))
		revision = utils.NewV2Revision()
	}

	return ret, revision, nil
}

// IteratorRouterRule
func (rc *RouteRuleCache) IteratorRouterRule(iterProc types.RouterRuleIterProc) {
	// need to traverse the Routing cache bucket of V2 here
	rc.container.foreach(iterProc)
}

// GetRoutingConfigCount Get the total number of routing configuration cache
func (rc *RouteRuleCache) GetRoutingConfigCount() int {
	return rc.container.size()
}

// GetRule implements api.RoutingConfigCache.
func (rc *RouteRuleCache) GetRule(id string) *model.ExtendRouterConfig {
	rule, _ := rc.container.rules.Load(id)
	return rule
}

// setRouterRules Store V2 Router Caches
func (rc *RouteRuleCache) setRouterRules(lastMtimes map[string]time.Time, cs []*model.RouterConfig) {
	if len(cs) == 0 {
		return
	}

	lastMtimeV2 := rc.LastMtime(rc.Name() + "v2").Unix()
	for _, entry := range cs {
		if entry.ID == "" {
			continue
		}
		if entry.ModifyTime.Unix() > lastMtimeV2 {
			lastMtimeV2 = entry.ModifyTime.Unix()
		}
		if !entry.Valid {
			rc.container.deleteV2(entry.ID)
			continue
		}
		extendEntry, err := entry.ToExpendRoutingConfig()
		if err != nil {
			log.Error("[Cache] routing config v2 convert to expend", zap.Error(err))
			continue
		}
		rc.container.saveV2(extendEntry)
	}
	lastMtimes[rc.Name()+"v2"] = time.Unix(lastMtimeV2, 0)
}

func (rc *RouteRuleCache) IsConvertFromV1(id string) (string, bool) {
	val, ok := rc.container.v1rulesToOld[id]
	return val, ok
}

func (rc *RouteRuleCache) convertV1toV2(rule *model.RoutingConfig) (bool, []*model.ExtendRouterConfig, error) {
	svc := rc.serviceCache.GetServiceByID(rule.ID)
	if svc == nil {
		s, err := rc.storage.GetServiceByID(rule.ID)
		if err != nil {
			return false, nil, err
		}
		if s == nil {
			return true, nil, nil
		}
		svc = s
	}
	if svc.IsAlias() {
		return false, nil, fmt.Errorf("svc: %+v is alias", svc)
	}

	in, out, err := model.ConvertRoutingV1ToExtendV2(svc.Name, svc.Namespace, rule)
	if err != nil {
		return false, nil, err
	}

	ret := make([]*model.ExtendRouterConfig, 0, len(in)+len(out))
	ret = append(ret, in...)
	ret = append(ret, out...)

	return true, ret, nil
}
