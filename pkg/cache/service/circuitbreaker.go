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
	"crypto/sha1"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	types "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// circuitBreaker的实现
type circuitBreakerCache struct {
	*cachebase.BaseCache

	storage store.Store
	// rules record id -> *rules.CircuitBreakerRule
	rules *utils.SyncMap[string, *rules.CircuitBreakerRule]
	// increment cache
	// fetched service cache
	// key1: namespace, key2: service
	circuitBreakers map[string]map[string]*rules.ServiceWithCircuitBreakerRules
	// key1: namespace
	nsWildcardRules map[string]*rules.ServiceWithCircuitBreakerRules
	// all rules are wildcard specific
	allWildcardRules *rules.ServiceWithCircuitBreakerRules
	lock             sync.RWMutex

	singleFlight singleflight.Group
}

// NewCircuitBreakerCache 返回一个操作CircuitBreakerCache的对象
func NewCircuitBreakerCache(s store.Store, cacheMgr types.CacheManager) types.CircuitBreakerCache {
	return &circuitBreakerCache{
		BaseCache:       cachebase.NewBaseCache(s, cacheMgr),
		storage:         s,
		rules:           utils.NewSyncMap[string, *rules.CircuitBreakerRule](),
		circuitBreakers: make(map[string]map[string]*rules.ServiceWithCircuitBreakerRules),
		nsWildcardRules: make(map[string]*rules.ServiceWithCircuitBreakerRules),
		allWildcardRules: rules.NewServiceWithCircuitBreakerRules(svctypes.ServiceKey{
			Namespace: types.AllMatched,
			Name:      types.AllMatched,
		}),
	}
}

// Initialize 实现Cache接口的函数
func (c *circuitBreakerCache) Initialize(_ map[string]interface{}) error {
	return nil
}

// Update 实现Cache接口的函数
func (c *circuitBreakerCache) Update() error {
	// 多个线程竞争，只有一个线程进行更新
	_, err, _ := c.singleFlight.Do(c.Name(), func() (interface{}, error) {
		return nil, c.DoCacheUpdate(c.Name(), c.realUpdate)
	})
	return err
}

func (c *circuitBreakerCache) realUpdate() (map[string]time.Time, int64, error) {
	start := time.Now()
	cbRules, err := c.storage.GetCircuitBreakerRulesForCache(c.LastFetchTime(), c.IsFirstUpdate())
	if err != nil {
		log.Errorf("[Cache][CircuitBreaker] cache update err:%s", err.Error())
		return nil, -1, err
	}
	lastMtimes, upsert, del := c.setCircuitBreaker(cbRules)
	log.Info("[Cache][CircuitBreaker] get more rules",
		zap.Int("pull-from-store", len(cbRules)), zap.Int("upsert", upsert), zap.Int("delete", del),
		zap.Time("last", c.LastMtime(c.Name())), zap.Duration("used", time.Since(start)))
	return lastMtimes, int64(len(cbRules)), nil
}

// clear 实现Cache接口的函数
func (c *circuitBreakerCache) Clear() error {
	c.BaseCache.Clear()
	c.lock.Lock()
	c.allWildcardRules.Clear()
	c.rules = utils.NewSyncMap[string, *rules.CircuitBreakerRule]()
	c.nsWildcardRules = make(map[string]*rules.ServiceWithCircuitBreakerRules)
	c.circuitBreakers = make(map[string]map[string]*rules.ServiceWithCircuitBreakerRules)
	c.lock.Unlock()
	return nil
}

// name 实现资源名称
func (c *circuitBreakerCache) Name() string {
	return types.CircuitBreakerName
}

// GetCircuitBreakerConfig 根据serviceID获取熔断规则
func (c *circuitBreakerCache) GetCircuitBreakerConfig(
	name string, namespace string) *rules.ServiceWithCircuitBreakerRules {
	// check service specific
	rules := c.checkServiceSpecificCache(name, namespace)
	if nil != rules {
		return rules
	}
	rules = c.checkNamespaceSpecificCache(namespace)
	if nil != rules {
		return rules
	}
	return c.allWildcardRules
}

func (c *circuitBreakerCache) checkServiceSpecificCache(
	name string, namespace string) *rules.ServiceWithCircuitBreakerRules {
	c.lock.RLock()
	defer c.lock.RUnlock()
	svcRules, ok := c.circuitBreakers[namespace]
	if ok {
		return svcRules[name]
	}
	return nil
}

func (c *circuitBreakerCache) checkNamespaceSpecificCache(namespace string) *rules.ServiceWithCircuitBreakerRules {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.nsWildcardRules[namespace]
}

func (c *circuitBreakerCache) reloadRevision(svcRules *rules.ServiceWithCircuitBreakerRules) {
	rulesCount := svcRules.CountCircuitBreakerRules()
	if rulesCount == 0 {
		svcRules.Revision = ""
		return
	}
	revisions := make([]string, 0, rulesCount)
	svcRules.IterateCircuitBreakerRules(func(rule *rules.CircuitBreakerRule) {
		revisions = append(revisions, rule.Revision)
	})
	sort.Strings(revisions)
	h := sha1.New()
	revision, err := types.ComputeRevisionBySlice(h, revisions)
	if err != nil {
		log.Errorf("[Server][Service][CircuitBreaker] compute revision service(%s) err: %s",
			svcRules.Service, err.Error())
		return
	}
	svcRules.Revision = revision
}

func (c *circuitBreakerCache) deleteAndReloadCircuitBreakerRules(
	svcRules *rules.ServiceWithCircuitBreakerRules, id string) {
	svcRules.DelCircuitBreakerRule(id)
	c.reloadRevision(svcRules)
}

func (c *circuitBreakerCache) deleteCircuitBreakerFromServiceCache(id string, svcKeys map[svctypes.ServiceKey]bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if len(svcKeys) == 0 {
		// all wildcard
		c.deleteAndReloadCircuitBreakerRules(c.allWildcardRules, id)
		for _, rules := range c.nsWildcardRules {
			c.deleteAndReloadCircuitBreakerRules(rules, id)
		}
		for _, svcRules := range c.circuitBreakers {
			for _, rules := range svcRules {
				c.deleteAndReloadCircuitBreakerRules(rules, id)
			}
		}
		return
	}
	svcToReloads := make(map[svctypes.ServiceKey]bool)
	for svcKey := range svcKeys {
		if svcKey.Name == types.AllMatched {
			rules, ok := c.nsWildcardRules[svcKey.Namespace]
			if ok {
				c.deleteAndReloadCircuitBreakerRules(rules, id)
			}
			svcRules, ok := c.circuitBreakers[svcKey.Namespace]
			if ok {
				for svc := range svcRules {
					svcToReloads[svctypes.ServiceKey{Namespace: svcKey.Namespace, Name: svc}] = true
				}
			}
		} else {
			svcToReloads[svcKey] = true
		}
	}
	if len(svcToReloads) > 0 {
		for svcToReload := range svcToReloads {
			svcRules, ok := c.circuitBreakers[svcToReload.Namespace]
			if ok {
				rules, ok := svcRules[svcToReload.Name]
				if ok {
					c.deleteAndReloadCircuitBreakerRules(rules, id)
				}
			}
		}
	}
}

func (c *circuitBreakerCache) storeAndReloadCircuitBreakerRules(
	svcRules *rules.ServiceWithCircuitBreakerRules, cbRule *rules.CircuitBreakerRule) {
	svcRules.AddCircuitBreakerRule(cbRule)
	c.reloadRevision(svcRules)
}

func createAndStoreServiceWithCircuitBreakerRules(svcKey svctypes.ServiceKey, key string,
	values map[string]*rules.ServiceWithCircuitBreakerRules) *rules.ServiceWithCircuitBreakerRules {
	rules := rules.NewServiceWithCircuitBreakerRules(svcKey)
	values[key] = rules
	return rules
}

func (c *circuitBreakerCache) storeCircuitBreakerToServiceCache(
	entry *rules.CircuitBreakerRule, svcKeys map[svctypes.ServiceKey]bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(svcKeys) == 0 {
		// all wildcard
		c.storeAndReloadCircuitBreakerRules(c.allWildcardRules, entry)
		for _, rules := range c.nsWildcardRules {
			c.storeAndReloadCircuitBreakerRules(rules, entry)
		}
		for _, svcRules := range c.circuitBreakers {
			for _, rules := range svcRules {
				c.storeAndReloadCircuitBreakerRules(rules, entry)
			}
		}
		return
	}
	svcToReloads := make(map[svctypes.ServiceKey]bool)
	for svcKey := range svcKeys {
		if svcKey.Name == types.AllMatched {
			var wildcardRules *rules.ServiceWithCircuitBreakerRules
			var ok bool
			wildcardRules, ok = c.nsWildcardRules[svcKey.Namespace]
			if !ok {
				wildcardRules = createAndStoreServiceWithCircuitBreakerRules(svcKey, svcKey.Namespace, c.nsWildcardRules)
				// add all exists wildcard rules
				c.allWildcardRules.IterateCircuitBreakerRules(func(rule *rules.CircuitBreakerRule) {
					wildcardRules.AddCircuitBreakerRule(rule)
				})
			}
			c.storeAndReloadCircuitBreakerRules(wildcardRules, entry)
			svcRules, ok := c.circuitBreakers[svcKey.Namespace]
			if ok {
				for svc := range svcRules {
					svcToReloads[svctypes.ServiceKey{Namespace: svcKey.Namespace, Name: svc}] = true
				}
			}
		} else {
			svcToReloads[svcKey] = true
		}
	}
	if len(svcToReloads) > 0 {
		for svcToReload := range svcToReloads {
			var breakerrules *rules.ServiceWithCircuitBreakerRules
			var svcRules map[string]*rules.ServiceWithCircuitBreakerRules
			var ok bool
			svcRules, ok = c.circuitBreakers[svcToReload.Namespace]
			if !ok {
				svcRules = make(map[string]*rules.ServiceWithCircuitBreakerRules)
				c.circuitBreakers[svcToReload.Namespace] = svcRules
			}
			breakerrules, ok = svcRules[svcToReload.Name]
			if !ok {
				breakerrules = createAndStoreServiceWithCircuitBreakerRules(svcToReload, svcToReload.Name, svcRules)
				// add all exists wildcard rules
				c.allWildcardRules.IterateCircuitBreakerRules(func(rule *rules.CircuitBreakerRule) {
					breakerrules.AddCircuitBreakerRule(rule)
				})
				// add all namespace wildcard rules
				nsRules, ok := c.nsWildcardRules[svcToReload.Namespace]
				if ok {
					nsRules.IterateCircuitBreakerRules(func(rule *rules.CircuitBreakerRule) {
						breakerrules.AddCircuitBreakerRule(rule)
					})
				}
			}
			c.storeAndReloadCircuitBreakerRules(breakerrules, entry)
		}
	}
}

func getServicesInvolveByCircuitBreakerRule(cbRule *rules.CircuitBreakerRule) map[svctypes.ServiceKey]bool {
	svcKeys := make(map[svctypes.ServiceKey]bool)
	addService := func(name string, namespace string) {
		if len(name) == 0 && len(namespace) == 0 {
			return
		}
		if name == types.AllMatched && namespace == types.AllMatched {
			return
		}
		svcKeys[svctypes.ServiceKey{
			Namespace: namespace,
			Name:      name,
		}] = true
	}
	addService(cbRule.DstService, cbRule.DstNamespace)
	return svcKeys
}

// setCircuitBreaker 更新store的数据到cache中
func (c *circuitBreakerCache) setCircuitBreaker(
	cbRules []*rules.CircuitBreakerRule) (map[string]time.Time, int, int) {

	if len(cbRules) == 0 {
		return nil, 0, 0
	}

	var upsert, del int

	lastMtime := c.LastMtime(c.Name()).Unix()

	for _, cbRule := range cbRules {
		if cbRule.ModifyTime.Unix() > lastMtime {
			lastMtime = cbRule.ModifyTime.Unix()
		}

		oldRule, ok := c.rules.Load(cbRule.ID)
		if ok {
			// 对比规则前后绑定的服务是否出现了变化，清理掉之前所绑定的信息数据
			if oldRule.IsServiceChange(cbRule) {
				// 从老的规则中获取所有的 svcKeys 信息列表
				svcKeys := getServicesInvolveByCircuitBreakerRule(oldRule)
				log.Info("[Cache][CircuitBreaker] clean rule bind old service info",
					zap.String("svc-keys", fmt.Sprintf("%#v", svcKeys)), zap.String("rule-id", cbRule.ID))
				// 挨个清空
				c.deleteCircuitBreakerFromServiceCache(cbRule.ID, svcKeys)
			}
		}
		svcKeys := getServicesInvolveByCircuitBreakerRule(cbRule)
		if !cbRule.Valid {
			del++
			c.rules.Delete(cbRule.ID)
			c.deleteCircuitBreakerFromServiceCache(cbRule.ID, svcKeys)
			continue
		}
		upsert++
		c.rules.Store(cbRule.ID, cbRule)
		c.storeCircuitBreakerToServiceCache(cbRule, svcKeys)
	}

	return map[string]time.Time{
		c.Name(): time.Unix(lastMtime, 0),
	}, upsert, del
}

// GetCircuitBreakerCount 获取熔断规则总数
func (c *circuitBreakerCache) GetCircuitBreakerCount() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	names := make(map[string]bool)
	c.allWildcardRules.IterateCircuitBreakerRules(func(rule *rules.CircuitBreakerRule) {
		names[rule.Name] = true
	})
	for _, breakerrules := range c.nsWildcardRules {
		breakerrules.IterateCircuitBreakerRules(func(rule *rules.CircuitBreakerRule) {
			names[rule.Name] = true
		})
	}
	for _, values := range c.circuitBreakers {
		for _, breakerrules := range values {
			breakerrules.IterateCircuitBreakerRules(func(rule *rules.CircuitBreakerRule) {
				names[rule.Name] = true
			})
		}
	}
	return len(names)
}

var (
	ignoreCircuitBreakerRuleFilter = map[string]struct{}{
		"brief":            {},
		"service":          {},
		"serviceNamespace": {},
		"exactName":        {},
		"excludeId":        {},
	}

	cbBlurSearchFields = map[string]func(*rules.CircuitBreakerRule) string{
		"name": func(cbr *rules.CircuitBreakerRule) string {
			return cbr.Name
		},
		"description": func(cbr *rules.CircuitBreakerRule) string {
			return cbr.Description
		},
		"srcservice": func(cbr *rules.CircuitBreakerRule) string {
			return cbr.SrcService
		},
		"dstservice": func(cbr *rules.CircuitBreakerRule) string {
			return cbr.DstService
		},
		"dstmethod": func(cbr *rules.CircuitBreakerRule) string {
			return cbr.DstMethod
		},
	}

	circuitBreakerSort = map[string]func(asc bool, a, b *rules.CircuitBreakerRule) bool{
		"mtime": func(asc bool, a, b *rules.CircuitBreakerRule) bool {
			ret := a.ModifyTime.Before(b.ModifyTime)
			return ret && asc
		},
		"id": func(asc bool, a, b *rules.CircuitBreakerRule) bool {
			ret := a.ID < b.ID
			return ret && asc
		},
		"name": func(asc bool, a, b *rules.CircuitBreakerRule) bool {
			ret := a.Name < b.Name
			return ret && asc
		},
	}
)

// Query implements api.CircuitBreakerCache.
func (c *circuitBreakerCache) Query(ctx context.Context, args *types.CircuitBreakerRuleArgs) (uint32, []*rules.CircuitBreakerRule, error) {
	if err := c.Update(); err != nil {
		return 0, nil, err
	}

	predicates := types.LoadCircuitBreakerRulePredicates(ctx)

	searchSvc, hasSvc := args.Filter["service"]
	searchNs, hasSvcNs := args.Filter["serviceNamespace"]
	exactNameValue, hasExactName := args.Filter["exactName"]
	excludeIdValue, hasExcludeId := args.Filter["excludeId"]

	lowerFilter := make(map[string]string, len(args.Filter))
	for k, v := range args.Filter {
		if _, ok := ignoreCircuitBreakerRuleFilter[k]; ok {
			continue
		}
		lowerFilter[strings.ToLower(k)] = v
	}

	results := make([]*rules.CircuitBreakerRule, 0, 32)
	c.rules.ReadRange(func(key string, val *rules.CircuitBreakerRule) {
		if hasSvcNs {
			srcNsValue := val.SrcNamespace
			dstNsValue := val.DstNamespace
			if !((srcNsValue == "*" || srcNsValue == searchNs) || (dstNsValue == "*" || dstNsValue == searchNs)) {
				return
			}
		}
		if hasSvc {
			srcSvcValue := val.SrcService
			dstSvcValue := val.DstService
			if !((srcSvcValue == searchSvc || srcSvcValue == "*") || (dstSvcValue == searchSvc || dstSvcValue == "*")) {
				return
			}
		}
		if hasExactName && exactNameValue != val.Name {
			return
		}
		if hasExcludeId && excludeIdValue != val.ID {
			return
		}
		for fieldKey, filterValue := range lowerFilter {
			getter, isBlur := cbBlurSearchFields[fieldKey]
			if isBlur {
				if utils.IsWildMatch(getter(val), filterValue) {
					return
				}
			} else if fieldKey == "enable" {
				if filterValue != strconv.FormatBool(val.Enable) {
					return
				}
			} else if fieldKey == "level" {
				levels := strings.Split(filterValue, ",")
				var inLevel = false
				for _, level := range levels {
					levelInt, _ := strconv.Atoi(level)
					if int64(levelInt) == int64(val.Level) {
						inLevel = true
						break
					}
				}
				if !inLevel {
					return
				}
			} else {
				// FIXME 暂时不知道还有什么字段查询需要适配，等待自测验证
			}
		}
		for i := range predicates {
			if !predicates[i](ctx, val) {
				return
			}
		}

		results = append(results, val)
	})

	sortFunc, ok := circuitBreakerSort[args.Filter["order_field"]]
	if !ok {
		sortFunc = circuitBreakerSort["mtime"]
	}
	asc := "asc" == strings.ToLower(args.Filter["order_type"])
	sort.Slice(results, func(i, j int) bool {
		return sortFunc(asc, results[i], results[j])
	})

	total, ret := c.toPage(uint32(len(results)), results, args)
	return total, ret, nil
}

func (c *circuitBreakerCache) toPage(total uint32, items []*rules.CircuitBreakerRule,
	args *types.CircuitBreakerRuleArgs) (uint32, []*rules.CircuitBreakerRule) {
	if args.Limit == 0 {
		return total, items
	}
	endIdx := args.Offset + args.Limit
	if endIdx > total {
		endIdx = total
	}
	return total, items[args.Offset:endIdx]
}

// GetRule implements api.CircuitBreakerCache.
func (c *circuitBreakerCache) GetRule(id string) *rules.CircuitBreakerRule {
	rule, _ := c.rules.Load(id)
	return rule
}
