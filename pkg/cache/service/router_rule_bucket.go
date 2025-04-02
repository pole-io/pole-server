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
	"sort"
	"sync"

	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// ServiceWithRouterRules 与服务绑定的路由规则数据
type ServiceWithRouterRules struct {
	direction rules.TrafficDirection
	mutex     sync.RWMutex
	Service   svctypes.ServiceKey
	// sortKeys: 针对 customv2Rules 做了排序
	sortKeys []string
	rules    map[string]*rules.ExtendRouterConfig
	revision string

	customv1RuleRef *utils.AtomicValue[*apitraffic.Routing]
}

func NewServiceWithRouterRules(svcKey svctypes.ServiceKey, direction rules.TrafficDirection) *ServiceWithRouterRules {
	return &ServiceWithRouterRules{
		direction: direction,
		Service:   svcKey,
		rules:     make(map[string]*rules.ExtendRouterConfig),
	}
}

// AddRouterRule 添加路由规则，注意，这里只会保留处于 Enable 状态的路由规则
func (s *ServiceWithRouterRules) AddRouterRule(rule *rules.ExtendRouterConfig) {
	if rule.GetRoutingPolicy() == apitraffic.RoutingPolicy_RulePolicy {
		s.customv1RuleRef = utils.NewAtomicValue[*apitraffic.Routing](&apitraffic.Routing{
			Inbounds:  []*apitraffic.Route{},
			Outbounds: []*apitraffic.Route{},
		})
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !rule.Enable {
		delete(s.rules, rule.ID)
	} else {
		s.rules[rule.ID] = rule
	}
}

func (s *ServiceWithRouterRules) DelRouterRule(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.rules, id)
}

// IterateRouterRules 这里是可以保证按照路由规则优先顺序进行遍历
func (s *ServiceWithRouterRules) IterateRouterRules(callback func(*rules.ExtendRouterConfig)) {
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

func (s *ServiceWithRouterRules) GetRouteRuleV1() *apitraffic.Routing {
	if !s.customv1RuleRef.HasValue() {
		return nil
	}
	return s.customv1RuleRef.Load()
}

func (s *ServiceWithRouterRules) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.rules = make(map[string]*rules.ExtendRouterConfig)
	s.revision = ""
}

func (s *ServiceWithRouterRules) reload() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.reloadRuleOrder()
	s.reloadRevision()
	s.reloadV1Rules()
}

func (s *ServiceWithRouterRules) reloadRuleOrder() {
	curRules := make([]*rules.ExtendRouterConfig, 0, len(s.rules))
	for i := range s.rules {
		curRules = append(curRules, s.rules[i])
	}

	sort.Slice(curRules, func(i, j int) bool {
		return rules.CompareRoutingV2(curRules[i], curRules[j])
	})

	curKeys := make([]string, 0, len(curRules))
	for i := range curRules {
		curKeys = append(curKeys, curRules[i].ID)
	}

	s.sortKeys = curKeys
}

func (s *ServiceWithRouterRules) reloadRevision() {
	revisioins := make([]string, 0, len(s.rules))
	for i := range s.sortKeys {
		revisioins = append(revisioins, s.rules[s.sortKeys[i]].Revision)
	}
	s.revision, _ = cacheapi.CompositeComputeRevision(revisioins)
}

func (s *ServiceWithRouterRules) reloadV1Rules() {
	if !s.customv1RuleRef.HasValue() {
		return
	}

	routerrules := make([]*rules.ExtendRouterConfig, 0, 32)
	for i := range s.sortKeys {
		rule, ok := s.rules[s.sortKeys[i]]
		if !ok {
			continue
		}
		routerrules = append(routerrules, rule)
	}

	routes := make([]*apitraffic.Route, 0, 32)

	for i := range routerrules {
		if routerrules[i].Policy != apitraffic.RoutingPolicy_RulePolicy.String() {
			continue
		}
		routes = append(routes, rules.BuildRoutes(routerrules[i], s.direction)...)
	}

	customv1Rules := &apitraffic.Routing{}
	switch s.direction {
	case rules.TrafficDirection_INBOUND:
		customv1Rules.Inbounds = routes
	case rules.TrafficDirection_OUTBOUND:
		customv1Rules.Outbounds = routes
	}

	s.customv1RuleRef.Store(customv1Rules)
}

func newClientRouteRuleContainer(direction rules.TrafficDirection) *ClientRouteRuleContainer {
	return &ClientRouteRuleContainer{
		direction:        direction,
		exactRules:       utils.NewSyncMap[string, *ServiceWithRouterRules](),
		nsWildcardRules:  utils.NewSyncMap[string, *ServiceWithRouterRules](),
		allWildcardRules: NewServiceWithRouterRules(svctypes.ServiceKey{Namespace: cacheapi.AllMatched, Name: cacheapi.AllMatched}, direction),
	}
}

type ClientRouteRuleContainer struct {
	// lock .
	lock sync.RWMutex

	direction rules.TrafficDirection
	// key1: namespace, key2: service
	exactRules *utils.SyncMap[string, *ServiceWithRouterRules]
	// key1: namespace is exact, service is full match
	nsWildcardRules *utils.SyncMap[string, *ServiceWithRouterRules]
	// all rules are wildcard specific
	allWildcardRules *ServiceWithRouterRules
}

func (c *ClientRouteRuleContainer) SearchRouteRuleV2(svc svctypes.ServiceKey) []*rules.ExtendRouterConfig {
	ret := make([]*rules.ExtendRouterConfig, 0, 32)

	c.lock.RLock()
	defer c.lock.RUnlock()

	exactRule, existExactRule := c.exactRules.Load(svc.Domain())
	if existExactRule {
		exactRule.IterateRouterRules(func(erc *rules.ExtendRouterConfig) {
			ret = append(ret, erc)
		})
	}

	nsWildcardRule, existNsWildcardRule := c.nsWildcardRules.Load(svc.Namespace)
	if existNsWildcardRule {
		nsWildcardRule.IterateRouterRules(func(erc *rules.ExtendRouterConfig) {
			ret = append(ret, erc)
		})
	}

	c.allWildcardRules.IterateRouterRules(func(erc *rules.ExtendRouterConfig) {
		ret = append(ret, erc)
	})

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Priority < ret[j].Priority
	})
	return ret
}

// SearchCustomRuleV1 针对 v1 客户端拉取路由规则
func (c *ClientRouteRuleContainer) SearchCustomRuleV1(svc svctypes.ServiceKey) (*apitraffic.Routing, []string) {
	ret := &apitraffic.Routing{
		Inbounds:  make([]*apitraffic.Route, 0, 8),
		Outbounds: make([]*apitraffic.Route, 0, 8),
	}
	exactRule, existExactRule := c.exactRules.Load(svc.Domain())
	nsWildcardRule, existNsWildcardRule := c.nsWildcardRules.Load(svc.Namespace)

	revisions := make([]string, 0, 2)

	switch c.direction {
	case rules.TrafficDirection_INBOUND:
		if existExactRule {
			ret.Inbounds = append(ret.Inbounds, exactRule.GetRouteRuleV1().GetInbounds()...)
		}
		if existNsWildcardRule {
			ret.Inbounds = append(ret.Inbounds, nsWildcardRule.GetRouteRuleV1().GetInbounds()...)
		}
	default:
		if existExactRule {
			ret.Outbounds = append(ret.Outbounds, exactRule.GetRouteRuleV1().GetOutbounds()...)
			revisions = append(revisions, exactRule.revision)
		}
		if existNsWildcardRule {
			ret.Outbounds = append(ret.Outbounds, nsWildcardRule.GetRouteRuleV1().GetOutbounds()...)
		}
	}
	if existExactRule {
		revisions = append(revisions, exactRule.revision)
	}
	if existNsWildcardRule {
		revisions = append(revisions, nsWildcardRule.revision)
	}

	// 最终在做一次排序
	sort.Slice(ret.Inbounds, func(i, j int) bool {
		return rules.CompareRoutingV1(ret.Inbounds[i], ret.Inbounds[j])
	})
	sort.Slice(ret.Outbounds, func(i, j int) bool {
		return rules.CompareRoutingV1(ret.Outbounds[i], ret.Outbounds[j])
	})

	return ret, revisions
}

func (c *ClientRouteRuleContainer) SaveRule(svcKey svctypes.ServiceKey, item *rules.ExtendRouterConfig) {
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
		rules:            utils.NewSyncMap[string, *rules.ExtendRouterConfig](),
		v1rules:          map[string][]*rules.ExtendRouterConfig{},
		v1rulesToOld:     map[string]string{},
		nearbyContainers: newClientRouteRuleContainer(rules.TrafficDirection_INBOUND),
		customContainers: map[rules.TrafficDirection]*ClientRouteRuleContainer{
			rules.TrafficDirection_INBOUND:  newClientRouteRuleContainer(rules.TrafficDirection_INBOUND),
			rules.TrafficDirection_OUTBOUND: newClientRouteRuleContainer(rules.TrafficDirection_OUTBOUND),
		},
		effect: utils.NewSyncSet[svctypes.ServiceKey](),
	}
}

// RouteRuleContainer v2 路由规则缓存 bucket
type RouteRuleContainer struct {
	// rules id => routing rule
	rules *utils.SyncMap[string, *rules.ExtendRouterConfig]

	// 就近路由规则缓存
	nearbyContainers *ClientRouteRuleContainer
	// 自定义路由规则缓存
	customContainers map[rules.TrafficDirection]*ClientRouteRuleContainer

	// effect 记录一次缓存更新中，那些服务的路由出现了更新
	effect *utils.SyncSet[svctypes.ServiceKey]

	// ------- 这里的逻辑都是为了兼容老的数据规则，这个将在1.18.2代码中移除，通过升级工具一次性处理 ------
	lock sync.RWMutex
	// v1rules service-id => []*rules.ExtendRouterConfig v1 版本的规则自动转为 v2 版本的规则，用于 v2 接口的数据查看
	v1rules map[string][]*rules.ExtendRouterConfig
	// v1rulesToOld 转为 v2 规则id 对应的原本的 v1 规则id 信息
	v1rulesToOld map[string]string
}

func (b *RouteRuleContainer) saveV2(conf *rules.ExtendRouterConfig) {
	b.rules.Store(conf.ID, conf)
	handler := func(container *ClientRouteRuleContainer, svcKey svctypes.ServiceKey) {
		// 避免读取到中间状态数据
		container.lock.Lock()
		defer container.lock.Unlock()

		b.effect.Add(svcKey)
		// 先删除，再保存
		container.CleanAllRule(conf.ID)
		container.SaveRule(svcKey, conf)
	}

	switch conf.GetRoutingPolicy() {
	case apitraffic.RoutingPolicy_RulePolicy:
		handler(b.customContainers[rules.TrafficDirection_OUTBOUND], conf.RuleRouting.Caller)
		handler(b.customContainers[rules.TrafficDirection_INBOUND], conf.RuleRouting.Callee)
	case apitraffic.RoutingPolicy_NearbyPolicy:
		handler(b.nearbyContainers, svctypes.ServiceKey{
			Namespace: conf.NearbyRouting.Namespace,
			Name:      conf.NearbyRouting.Service,
		})
	}

}

// saveV1 保存 v1 级别的路由规则
func (b *RouteRuleContainer) saveV1(v1rule *rules.RoutingConfig, v2rules []*rules.ExtendRouterConfig) {
	for i := range v2rules {
		b.saveV2(v2rules[i])
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	b.v1rules[v1rule.ID] = v2rules

	for i := range v2rules {
		item := v2rules[i]
		b.v1rulesToOld[item.ID] = v1rule.ID
	}
}

func (b *RouteRuleContainer) convertV2Size() uint32 {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return uint32(len(b.v1rulesToOld))
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

	switch rule.GetRoutingPolicy() {
	case apitraffic.RoutingPolicy_RulePolicy:
		handler(b.customContainers[rules.TrafficDirection_OUTBOUND], rule.RuleRouting.Caller)
		handler(b.customContainers[rules.TrafficDirection_INBOUND], rule.RuleRouting.Callee)
	case apitraffic.RoutingPolicy_NearbyPolicy:
		handler(b.nearbyContainers, svctypes.ServiceKey{
			Namespace: rule.NearbyRouting.Namespace,
			Name:      rule.NearbyRouting.Service,
		})
	}
}

// deleteV1 删除 v1 的路由规则
func (b *RouteRuleContainer) deleteV1(serviceId string) {
	b.lock.Lock()
	defer b.lock.Unlock()

	items, ok := b.v1rules[serviceId]
	if !ok {
		delete(b.v1rules, serviceId)
		return
	}

	for i := range items {
		delete(b.v1rulesToOld, items[i].ID)
		b.deleteV2(items[i].ID)
	}
	delete(b.v1rules, serviceId)
}

// size Number of routing-v2 cache rules
func (b *RouteRuleContainer) size() int {
	b.lock.RLock()
	defer b.lock.RUnlock()

	cnt := b.rules.Len()
	for k := range b.v1rules {
		cnt += len(b.v1rules[k])
	}

	return cnt
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

// foreach Traversing all routing rules
func (b *RouteRuleContainer) foreach(proc cacheapi.RouterRuleIterProc) {
	b.rules.Range(func(key string, val *rules.ExtendRouterConfig) {
		proc(key, val)
	})

	for _, rules := range b.v1rules {
		for i := range rules {
			proc(rules[i].ID, rules[i])
		}
	}
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
