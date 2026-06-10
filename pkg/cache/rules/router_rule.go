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

package rules

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	cachetypes "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	revisionapi "github.com/pole-io/pole-server/apis/pkg/utils/revision"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/pole-server/pkg/common/utils"
	matchs "github.com/pole-io/pole-server/pkg/common/utils/match"
)

type (
	// RouteRuleCache Routing rules cache
	RouteRuleCache struct {
		*cachebase.BaseCache
		// serviceCache 服务缓存
		serviceCache cachetypes.ServiceCache
		// ids 路由规则保存
		ids *container.SyncMap[string, *rules.RouterConfig]
		// container 这里保存的是已经发布的规则
		container *RouteRuleContainer
		// lastMtime 最新的规则更新时间
		lastMtime time.Time
	}
)

// NewRouteRuleCache Return a object of operating RouteRuleCache
func NewRouteRuleCache(s store.Store, cacheMgr cachetypes.CacheManager) cachetypes.RouterRuleCache {
	return &RouteRuleCache{
		BaseCache: cachebase.NewBaseCache(s, cacheMgr),
	}
}

// initialize The function of implementing the cache interface
func (rc *RouteRuleCache) Initialize(_ map[string]any) error {
	rc.lastMtime = time.Unix(0, 0)
	rc.ids = container.NewSyncMap[string, *rules.RouterConfig]()
	rc.container = newRouteRuleContainer()
	rc.serviceCache = rc.BaseCache.CacheMgr.GetCacher(cachetypes.CacheService).(cachetypes.ServiceCache)
	registerGovernanceRuleWatcher(rc.Store(), rc.CacheMgr, rc)
	return nil
}

// Update The function of implementing the cache interface
func (rc *RouteRuleCache) Update() error {
	if ok, err := updateGovernanceRuleCache(rc.Store()); ok {
		return err
	}
	// Multiple thread competition, only one thread is updated
	_, err, _ := rc.GetSingle().Do(rc.Name(), func() (any, error) {
		return nil, rc.DoCacheUpdate(rc.Name(), rc.realUpdate)
	})
	return err
}

// update The function of implementing the cache interface
func (rc *RouteRuleCache) realUpdate() (map[string]time.Time, int64, error) {
	allLastMtimes := map[string]time.Time{}

	rules, err := rc.Store().GetMoreRouterRule(rc.LastFetchTime(), rc.IsFirstUpdate())
	if err != nil {
		log.Errorf("[cache][router] console cache get from store err: %s", err.Error())
		return nil, -1, err
	}
	lastMtime, upsert, del := rc.setRouterRuleConsole(rules)
	log.Info("[cache][router] console cache update",
		zap.Int("pull-from-store", len(rules)), zap.Int("upsert", upsert), zap.Int("delete", del),
		zap.Time("last", lastMtime))
	allLastMtimes[rc.Name()+"_console"] = lastMtime

	releases, err := rc.Store().GetMoreRouterRuleReleases(rc.IsFirstUpdate(), rc.LastFetchTime())
	if err != nil {
		log.Errorf("[cache][router] release cache get from store err: %s", err.Error())
		return nil, -1, err
	}

	lastMtime, upsert, del = rc.setRouterRuleClient(releases)
	log.Info("[cache][router] release cache update",
		zap.Int("pull-from-store", len(releases)), zap.Int("upsert", upsert), zap.Int("delete", del),
		zap.Time("last", lastMtime))
	allLastMtimes[rc.Name()+"_release"] = lastMtime

	rc.container.reload()
	return allLastMtimes, int64(len(releases)), err
}

// Clear The function of implementing the cache interface
func (rc *RouteRuleCache) Clear() error {
	resetGovernanceRuleUpdateCache(rc.Store())
	rc.BaseCache.Clear()
	rc.container = newRouteRuleContainer()
	rc.lastMtime = time.Unix(0, 0)
	return nil
}

// Name The function of implementing the cache interface
func (rc *RouteRuleCache) Name() string {
	return cachetypes.RoutingConfigName
}

func (rc *RouteRuleCache) ListRouterRule(service, namespace string) []*rules.ExtendRouterConfig {
	routerRules := rc.container.SearchCustomRules(service, namespace)
	ret := make([]*rules.ExtendRouterConfig, 0, len(routerRules))
	ret = append(ret, routerRules...)
	return ret
}

// GetRouterRule Obtain routing configuration based on serviceid
func (rc *RouteRuleCache) GetRouterRule(id, service, namespace string) ([]*apitraffic.RouteRule, string, error) {
	if id == "" && service == "" && namespace == "" {
		return nil, "", nil
	}

	routerRules := rc.container.SearchCustomRules(service, namespace)
	revisions := make([]string, 0, len(routerRules))
	rules := make([]*apitraffic.RouteRule, 0, len(routerRules))
	for i := range routerRules {
		item := routerRules[i]
		entry, err := item.ToApi()
		if err != nil {
			return nil, "", err
		}
		rules = append(rules, entry)
		revisions = append(revisions, entry.GetRevision())
	}
	revision, err := revisionapi.CompositeComputeRevision(revisions)
	if err != nil {
		log.Warn("[Cache][Routing] v2=>v1 compute revisions fail, use fake revision", zap.Error(err))
		revision = utils.NewRevision()
	}

	return rules, revision, nil
}

// GetNearbyRouteRule 根据服务名查询就近路由数据
func (rc *RouteRuleCache) GetNearbyRouteRule(service, namespace string) ([]*apitraffic.RouteRule, string, error) {
	if service == "" && namespace == "" {
		return nil, "", nil
	}

	svcKey := svctypes.ServiceKey{
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
	revision, err := revisionapi.CompositeComputeRevision(revisions)
	if err != nil {
		log.Warn("[Cache][Routing] v2=>v1 compute revisions fail, use fake revision", zap.Error(err))
		revision = utils.NewRevision()
	}

	return ret, revision, nil
}

// IteratorRouterRule
func (rc *RouteRuleCache) IteratorRouterRule(iterProc cachetypes.RouterRuleIterProc) {
	// need to traverse the Routing cache bucket of V2 here
	rc.ids.Range(func(key string, val *rules.RouterConfig) {
		iterProc(key, val)
	})
}

// GetRouterCount Get the total number of routing configuration cache
func (rc *RouteRuleCache) GetRouterCount() int {
	return rc.container.size()
}

// GetRule implements api.RouterRuleCache.
func (rc *RouteRuleCache) GetRule(id string) *rules.RouterConfig {
	rule, _ := rc.ids.Load(id)
	return rule
}

// setRouterRuleConsole 用于控制台列表查询缓存
func (rc *RouteRuleCache) setRouterRuleConsole(cs []*rules.RouterConfig) (time.Time, int, int) {
	if len(cs) == 0 {
		return time.Time{}, 0, 0
	}

	upsert := 0
	del := 0

	lastMtime := rc.LastMtime(rc.Name() + "_client").Unix()
	for _, entry := range cs {
		if entry.ID == "" {
			continue
		}
		if entry.ModifyTime.Unix() > lastMtime {
			lastMtime = entry.ModifyTime.Unix()
		}
		if !entry.Valid {
			del++
			rc.container.deleteV2(entry.ID)
			continue
		}
		upsert++
		rc.ids.Store(entry.ID, entry)
	}
	return time.Unix(lastMtime, 0), upsert, del
}

// setRouterRuleClient 用于客户端规则查询缓存
func (rc *RouteRuleCache) setRouterRuleClient(cs []*rules.RouterRuleRelease) (time.Time, int, int) {
	if len(cs) == 0 {
		return time.Time{}, 0, 0
	}

	upsert := 0
	del := 0

	lastMtime := rc.LastMtime(rc.Name() + "_client").Unix()
	for _, entry := range cs {
		if entry.Key() == "" {
			continue
		}
		if entry.Mtime.Unix() > lastMtime {
			lastMtime = entry.Mtime.Unix()
		}
		if !entry.Valid {
			del++
			rc.container.deleteV2(entry.Key())
			continue
		}
		upsert++
		rc.container.saveV2(entry)
	}

	return time.Unix(lastMtime, 0), upsert, del
}

// ServiceWithRouterRules 与服务绑定的路由规则数据
type ServiceWithRouterRules struct {
	direction rules.TrafficDirection
	mutex     sync.RWMutex
	Service   svctypes.ServiceKey
	// sortKeys: 针对 customv2Rules 做了排序
	sortKeys []string
	rules    map[string]*rules.RouterRuleRelease
	revision string
}

func NewServiceWithRouterRules(svcKey svctypes.ServiceKey, direction rules.TrafficDirection) *ServiceWithRouterRules {
	return &ServiceWithRouterRules{
		direction: direction,
		Service:   svcKey,
		rules:     make(map[string]*rules.RouterRuleRelease),
	}
}

// AddRouterRule 添加路由规则，注意，这里只会保留处于 Enable 状态的路由规则
func (s *ServiceWithRouterRules) AddRouterRule(rule *rules.RouterRuleRelease) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !rule.Active {
		delete(s.rules, rule.Key())
	} else {
		s.rules[rule.Key()] = rule
	}
}

func (s *ServiceWithRouterRules) DelRouterRule(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.rules, id)
}

// IterateRouterRules 这里是可以保证按照路由规则优先顺序进行遍历
func (s *ServiceWithRouterRules) IterateRouterRules(callback func(*rules.RouterRuleRelease)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, key := range s.sortKeys {
		val, ok := s.rules[key]
		if ok {
			callback(val)
		}
	}
}

func (s *ServiceWithRouterRules) CountRouterRules() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.rules)
}

func (s *ServiceWithRouterRules) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.rules = make(map[string]*rules.RouterRuleRelease)
	s.revision = ""
}

func (s *ServiceWithRouterRules) reload() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.reloadRuleOrder()
	s.reloadRevision()
}

func (s *ServiceWithRouterRules) reloadRuleOrder() {
	curRules := make([]*rules.RouterRuleRelease, 0, len(s.rules))
	for i := range s.rules {
		curRules = append(curRules, s.rules[i])
	}

	sort.Slice(curRules, func(i, j int) bool {
		return rules.CompareRoutingV2(curRules[i].Rule, curRules[j].Rule)
	})

	curKeys := make([]string, 0, len(curRules))
	for i := range curRules {
		curKeys = append(curKeys, curRules[i].Key())
	}

	s.sortKeys = curKeys
}

func (s *ServiceWithRouterRules) reloadRevision() {
	revisioins := make([]string, 0, len(s.rules))
	for i := range s.sortKeys {
		revisioins = append(revisioins, strconv.Itoa(int(s.rules[s.sortKeys[i]].Version)))
	}
	s.revision, _ = revisionapi.CompositeComputeRevision(revisioins)
}

func newClientRouteRuleContainer(direction rules.TrafficDirection) *ClientRouteRuleContainer {
	return &ClientRouteRuleContainer{
		direction:        direction,
		exactRules:       container.NewSyncMap[string, *ServiceWithRouterRules](),
		nsWildcardRules:  container.NewSyncMap[string, *ServiceWithRouterRules](),
		allWildcardRules: NewServiceWithRouterRules(svctypes.ServiceKey{Namespace: cachetypes.AllMatched, Name: cachetypes.AllMatched}, direction),
	}
}

type ClientRouteRuleContainer struct {
	// lock .
	lock sync.RWMutex

	direction rules.TrafficDirection
	// key1: namespace, key2: service
	exactRules *container.SyncMap[string, *ServiceWithRouterRules]
	// key1: namespace is exact, service is full match
	nsWildcardRules *container.SyncMap[string, *ServiceWithRouterRules]
	// all rules are wildcard specific
	allWildcardRules *ServiceWithRouterRules
}

func (c *ClientRouteRuleContainer) SearchRouteRuleV2(svc svctypes.ServiceKey) []*rules.ExtendRouterConfig {
	ret := make([]*rules.ExtendRouterConfig, 0, 32)

	c.lock.RLock()
	defer c.lock.RUnlock()

	exactRule, existExactRule := c.exactRules.Load(svc.Domain())
	if existExactRule {
		exactRule.IterateRouterRules(func(erc *rules.RouterRuleRelease) {
			ret = append(ret, erc.Rule)
		})
	}

	nsWildcardRule, existNsWildcardRule := c.nsWildcardRules.Load(svc.Namespace)
	if existNsWildcardRule {
		nsWildcardRule.IterateRouterRules(func(erc *rules.RouterRuleRelease) {
			ret = append(ret, erc.Rule)
		})
	}

	c.allWildcardRules.IterateRouterRules(func(erc *rules.RouterRuleRelease) {
		ret = append(ret, erc.Rule)
	})

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Priority < ret[j].Priority
	})
	return ret
}

func (c *ClientRouteRuleContainer) SaveRule(svcKey svctypes.ServiceKey, item *rules.RouterRuleRelease) {
	// level1 级别 cache 处理
	if svcKey.Name != rules.MatchAll && svcKey.Namespace != rules.MatchAll {
		c.exactRules.ComputeIfAbsent(svcKey.Domain(), func(k string) *ServiceWithRouterRules {
			return NewServiceWithRouterRules(svcKey, c.direction)
		})
		svcContainer, _ := c.exactRules.Load(svcKey.Domain())
		svcContainer.AddRouterRule(item)
	}
	// level2 级别 cache 处理
	if svcKey.Name == rules.MatchAll && svcKey.Namespace != rules.MatchAll {
		c.nsWildcardRules.ComputeIfAbsent(svcKey.Namespace, func(k string) *ServiceWithRouterRules {
			return NewServiceWithRouterRules(svcKey, c.direction)
		})

		nsRules, _ := c.nsWildcardRules.Load(svcKey.Namespace)
		nsRules.AddRouterRule(item)
	}
	// level3 级别 cache 处理
	if svcKey.Name == rules.MatchAll && svcKey.Namespace == rules.MatchAll {
		c.allWildcardRules.AddRouterRule(item)
	}
}

func (c *ClientRouteRuleContainer) RemoveRule(svcKey svctypes.ServiceKey, ruleId string) {
	// level1 级别 cache 处理
	if svcKey.Name != rules.MatchAll && svcKey.Namespace != rules.MatchAll {
		svcContainer, ok := c.exactRules.Load(svcKey.Domain())
		if !ok {
			return
		}
		svcContainer.DelRouterRule(ruleId)
	}
	// level2 级别 cache 处理
	if svcKey.Name == rules.MatchAll && svcKey.Namespace != rules.MatchAll {
		nsRules, ok := c.nsWildcardRules.Load(svcKey.Namespace)
		if !ok {
			return
		}
		nsRules.DelRouterRule(ruleId)
	}
	// level3 级别 cache 处理
	if svcKey.Name == rules.MatchAll && svcKey.Namespace == rules.MatchAll {
		c.allWildcardRules.DelRouterRule(ruleId)
	}
}

func (c *ClientRouteRuleContainer) CleanAllRule(ruleId string) {
	// level1 级别 cache 处理
	c.exactRules.Range(func(key string, svcContainer *ServiceWithRouterRules) {
		svcContainer.DelRouterRule(ruleId)
	})
	// level2 级别 cache 处理
	c.nsWildcardRules.Range(func(key string, svcContainer *ServiceWithRouterRules) {
		svcContainer.DelRouterRule(ruleId)
	})
	// level3 级别 cache 处理
	c.allWildcardRules.DelRouterRule(ruleId)
}

func newRouteRuleContainer() *RouteRuleContainer {
	return &RouteRuleContainer{
		rules:            container.NewSyncMap[string, *rules.RouterRuleRelease](),
		nearbyContainers: newClientRouteRuleContainer(rules.TrafficDirection_INBOUND),
		customContainers: map[rules.TrafficDirection]*ClientRouteRuleContainer{
			rules.TrafficDirection_INBOUND:  newClientRouteRuleContainer(rules.TrafficDirection_INBOUND),
			rules.TrafficDirection_OUTBOUND: newClientRouteRuleContainer(rules.TrafficDirection_OUTBOUND),
		},
		effect: container.NewSyncSet[svctypes.ServiceKey](),
	}
}

// RouteRuleContainer v2 路由规则缓存 bucket
type RouteRuleContainer struct {
	// rules id => routing rule
	rules *container.SyncMap[string, *rules.RouterRuleRelease]

	// 就近路由规则缓存
	nearbyContainers *ClientRouteRuleContainer
	// 自定义路由规则缓存
	customContainers map[rules.TrafficDirection]*ClientRouteRuleContainer

	// effect 记录一次缓存更新中，那些服务的路由出现了更新
	effect *container.SyncSet[svctypes.ServiceKey]
}

func (b *RouteRuleContainer) saveV2(conf *rules.RouterRuleRelease) {
	b.rules.Store(conf.Key(), conf)
	handler := func(container *ClientRouteRuleContainer, svcKey svctypes.ServiceKey) {
		// 避免读取到中间状态数据
		container.lock.Lock()
		defer container.lock.Unlock()

		b.effect.Add(svcKey)
		// 先删除，再保存
		container.CleanAllRule(conf.Key())
		container.SaveRule(svcKey, conf)
	}

	switch conf.Rule.GetRoutePolicy() {
	case apitraffic.RoutePolicy_RulePolicy:
		handler(b.customContainers[rules.TrafficDirection_OUTBOUND], conf.Rule.RuleRouting.Caller)
		handler(b.customContainers[rules.TrafficDirection_INBOUND], conf.Rule.RuleRouting.Callee)
	case apitraffic.RoutePolicy_NearbyPolicy:
		handler(b.nearbyContainers, svctypes.ServiceKey{
			Namespace: conf.Rule.NearbyRouting.Namespace,
			Name:      conf.Rule.NearbyRouting.Service,
		})
	}

}

func (b *RouteRuleContainer) deleteV2(id string) {
	rule, exist := b.rules.Load(id)
	b.rules.Delete(id)
	if !exist {
		return
	}

	handler := func(container *ClientRouteRuleContainer, svcKey svctypes.ServiceKey) {
		b.effect.Add(svcKey)
		container.RemoveRule(svcKey, id)
	}

	switch rule.Rule.GetRoutePolicy() {
	case apitraffic.RoutePolicy_RulePolicy:
		handler(b.customContainers[rules.TrafficDirection_OUTBOUND], rule.Rule.RuleRouting.Caller)
		handler(b.customContainers[rules.TrafficDirection_INBOUND], rule.Rule.RuleRouting.Callee)
	case apitraffic.RoutePolicy_NearbyPolicy:
		handler(b.nearbyContainers, svctypes.ServiceKey{
			Namespace: rule.Rule.NearbyRouting.Namespace,
			Name:      rule.Rule.NearbyRouting.Service,
		})
	}
}

// size Number of routing-v2 cache rules
func (b *RouteRuleContainer) size() int {
	return b.rules.Len()
}

func (b *RouteRuleContainer) SearchCustomRules(svcName, namespace string) []*rules.ExtendRouterConfig {
	ruleIds := map[string]struct{}{}

	svcKey := svctypes.ServiceKey{Namespace: namespace, Name: svcName}

	ret := make([]*rules.ExtendRouterConfig, 0, 32)

	routerrules := b.customContainers[rules.TrafficDirection_INBOUND].SearchRouteRuleV2(svcKey)
	ret = append(ret, routerrules...)
	for i := range routerrules {
		ruleIds[routerrules[i].ID] = struct{}{}
	}

	routerrules = b.customContainers[rules.TrafficDirection_OUTBOUND].SearchRouteRuleV2(svcKey)
	for i := range routerrules {
		if _, ok := ruleIds[routerrules[i].ID]; !ok {
			ret = append(ret, routerrules[i])
		}
	}

	return ret
}

func (b *RouteRuleContainer) reload() {
	b.effect.Range(func(val svctypes.ServiceKey) {
		b.reloadCustom(val)
		b.reloadNearby(val)
	})
}

func (b *RouteRuleContainer) reloadCustom(val svctypes.ServiceKey) {
	// 处理自定义路由
	// 处理 exact
	rrules, ok := b.customContainers[rules.TrafficDirection_INBOUND].exactRules.Load(val.Domain())
	if ok {
		rrules.reload()
	}
	rrules, ok = b.customContainers[rules.TrafficDirection_OUTBOUND].exactRules.Load(val.Domain())
	if ok {
		rrules.reload()
	}

	// 处理 ns wildcard
	rrules, ok = b.customContainers[rules.TrafficDirection_INBOUND].nsWildcardRules.Load(val.Namespace)
	if ok {
		rrules.reload()
	}
	rrules, ok = b.customContainers[rules.TrafficDirection_OUTBOUND].nsWildcardRules.Load(val.Namespace)
	if ok {
		rrules.reload()
	}

	// 处理 all wildcard
	b.customContainers[rules.TrafficDirection_INBOUND].allWildcardRules.reload()
	b.customContainers[rules.TrafficDirection_OUTBOUND].allWildcardRules.reload()
}

func (b *RouteRuleContainer) reloadNearby(val svctypes.ServiceKey) {
	// 处理 exact
	rules, ok := b.nearbyContainers.exactRules.Load(val.Domain())
	if ok {
		rules.reload()
	}
	// 处理 ns wildcard
	rules, ok = b.nearbyContainers.nsWildcardRules.Load(val.Namespace)
	if ok {
		rules.reload()
	}
	// 处理 all wildcard
	b.nearbyContainers.allWildcardRules.reload()
}

func queryRoutingRuleV2ByService(rule *rules.ExtendRouterConfig, sourceNamespace, sourceService,
	destNamespace, destService string, both bool) bool {
	var (
		sourceFind bool
		destFind   bool
	)

	hasSourceSvc := len(sourceService) != 0
	hasSourceNamespace := len(sourceNamespace) != 0
	hasDestSvc := len(destService) != 0
	hasDestNamespace := len(destNamespace) != 0

	sourceService, isWildSourceSvc := matchs.ParseWildName(sourceService)
	sourceNamespace, isWildSourceNamespace := matchs.ParseWildName(sourceNamespace)
	destService, isWildDestSvc := matchs.ParseWildName(destService)
	destNamespace, isWildDestNamespace := matchs.ParseWildName(destNamespace)

	customRule := rule.RuleRouting.RuleRouting
	if hasSourceNamespace || hasSourceSvc {
		if hasSourceSvc {
			if isWildSourceSvc {
				if !strings.Contains(customRule.GetCaller().Service, sourceService) {
					return false
				}
			} else if customRule.GetCaller().Service != sourceService {
				return false
			}
		}
		if hasSourceNamespace {
			if isWildSourceNamespace {
				if !strings.Contains(customRule.GetCaller().Namespace, sourceNamespace) {
					return false
				}
			} else if customRule.GetCaller().Namespace != sourceNamespace {
				return false
			}
		}
	}

	if hasDestNamespace || hasDestSvc {
		if hasDestSvc {
			if isWildDestSvc && !strings.Contains(customRule.GetCallee().Service, destService) {
				return false
			}
			if customRule.GetCallee().Service != destService {
				return false
			}
		}
		if hasDestNamespace {
			if isWildDestNamespace && !strings.Contains(customRule.GetCallee().Namespace, destNamespace) {
				return false
			}
			if customRule.GetCallee().Namespace != destNamespace {
				return false
			}
		}
	}

	if both {
		if sourceFind && destFind {
			return true
		}
	} else if sourceFind || destFind {
		return true
	}
	return false
}

// QueryRouterRules Query Route Configuration List
func (rc *RouteRuleCache) QueryRouterRules(ctx context.Context, args *cacheapi.RoutingArgs) (uint32, []*rules.ExtendRouterConfig, error) {
	if err := rc.Update(); err != nil {
		return 0, nil, err
	}
	hasSvcQuery := len(args.Service) != 0 || len(args.Namespace) != 0
	hasSourceQuery := len(args.SourceService) != 0 || len(args.SourceNamespace) != 0
	hasDestQuery := len(args.DestinationService) != 0 || len(args.DestinationNamespace) != 0
	needBoth := hasSourceQuery && hasDestQuery

	res := make([]*rules.ExtendRouterConfig, 0, 8)

	var process = func(_ string, routeRule *rules.ExtendRouterConfig) {
		if args.ID != "" && args.ID != routeRule.ID {
			return
		}

		if routeRule.GetRoutePolicy() == apitraffic.RoutePolicy_NearbyPolicy {
			if args.Namespace != "" {
				if args.SourceNamespace == "" {
					args.SourceNamespace = args.Namespace
				}
				if args.DestinationNamespace == "" {
					args.DestinationNamespace = args.Namespace
				}
			}
			if args.Service != "" {
				if args.SourceService == "" {
					args.SourceService = args.Service
				}
				if args.DestinationService == "" {
					args.DestinationService = args.Service
				}
			}
			if hasSvcQuery || hasSourceQuery || hasDestQuery {
				if !queryRoutingRuleV2ByService(routeRule,
					args.SourceNamespace, args.SourceService,
					args.DestinationNamespace, args.DestinationService,
					needBoth) {
					return
				}
			}
		}

		if args.Name != "" {
			name, isWild := matchs.ParseWildName(args.Name)
			if isWild {
				if !strings.Contains(routeRule.Name, name) {
					return
				}
			} else if args.Name != routeRule.Name {
				return
			}
		}

		if args.Enable != nil && *args.Enable != routeRule.Enable {
			return
		}

		res = append(res, routeRule)
	}

	predicates := cacheapi.LoadRouterRulePredicates(ctx)

	rc.IteratorRouterRule(func(key string, value *rules.RouterConfig) {
		for i := range predicates {
			if !predicates[i](ctx, value) {
				return
			}
		}
		pdata, _ := value.ToExpendRoutingConfig()
		process(key, pdata)
	})

	amount, routings := rc.sortBeforeTrim(res, args)
	return amount, routings, nil
}

func (rc *RouteRuleCache) sortBeforeTrim(routings []*rules.ExtendRouterConfig,
	args *cacheapi.RoutingArgs) (uint32, []*rules.ExtendRouterConfig) {
	amount := uint32(len(routings))
	if args.Offset >= amount || args.Limit == 0 {
		return amount, nil
	}
	sort.Slice(routings, func(i, j int) bool {
		asc := strings.ToLower(args.OrderType) == "asc" || args.OrderType == ""
		if strings.ToLower(args.OrderField) == "priority" {
			return orderByRoutingPriority(routings[i], routings[j], asc)
		}
		return orderByRoutingModifyTime(routings[i], routings[j], asc)
	})
	endIdx := args.Offset + args.Limit
	if endIdx > amount {
		endIdx = amount
	}
	return amount, routings[args.Offset:endIdx]
}

func orderByRoutingPriority(a, b *rules.ExtendRouterConfig, asc bool) bool {
	if a.Priority < b.Priority {
		return asc
	}
	if a.Priority > b.Priority {
		// false && asc always false
		return false
	}
	return strings.Compare(a.ID, b.ID) < 0 && asc
}

func orderByRoutingModifyTime(a, b *rules.ExtendRouterConfig, asc bool) bool {
	if a.ModifyTime.After(b.ModifyTime) {
		return asc
	}
	if a.ModifyTime.Before(b.ModifyTime) {
		// false && asc always false
		return false
	}
	return strings.Compare(a.ID, b.ID) < 0 && asc
}
