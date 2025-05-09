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
	"crypto/sha1"
	"fmt"
	"sort"
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
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

type faultDetectCache struct {
	*cachebase.BaseCache

	storage store.Store
	// rules record id -> *rules.FaultDetectRule
	rules *container.SyncMap[string, *rules.FaultDetectRule]
	// increment cache
	// fetched service cache
	// key1: namespace, key2: service
	svcSpecificRules map[string]map[string]*rules.ServiceWithFaultDetectRules
	// key1: namespace
	nsWildcardRules map[string]*rules.ServiceWithFaultDetectRules
	// all rules are wildcard specific
	allWildcardRules *rules.ServiceWithFaultDetectRules
	lock             sync.RWMutex
	singleFlight     singleflight.Group
}

// NewFaultDetectCache faultDetectCache constructor
func NewFaultDetectCache(s store.Store, cacheMgr types.CacheManager) types.FaultDetectCache {
	return &faultDetectCache{
		BaseCache:        cachebase.NewBaseCache(s, cacheMgr),
		storage:          s,
		rules:            container.NewSyncMap[string, *rules.FaultDetectRule](),
		svcSpecificRules: make(map[string]map[string]*rules.ServiceWithFaultDetectRules),
		nsWildcardRules:  make(map[string]*rules.ServiceWithFaultDetectRules),
		allWildcardRules: rules.NewServiceWithFaultDetectRules(svctypes.ServiceKey{
			Namespace: types.AllMatched,
			Name:      types.AllMatched,
		}),
	}
}

// Initialize 实现Cache接口的函数
func (f *faultDetectCache) Initialize(_ map[string]interface{}) error {
	return nil
}

func (f *faultDetectCache) Update() error {
	_, err, _ := f.singleFlight.Do(f.Name(), func() (interface{}, error) {
		return nil, f.DoCacheUpdate(f.Name(), f.realUpdate)
	})
	return err
}

// update 实现Cache接口的函数
func (f *faultDetectCache) realUpdate() (map[string]time.Time, int64, error) {
	fdRules, err := f.storage.GetFaultDetectRulesForCache(f.LastFetchTime(), f.IsFirstUpdate())
	if err != nil {
		log.Errorf("[Cache] fault detect config cache update err:%s", err.Error())
		return nil, -1, err
	}
	lastMtimes := f.setFaultDetectRules(fdRules)

	return lastMtimes, int64(len(fdRules)), nil
}

// clear 实现Cache接口的函数
func (f *faultDetectCache) Clear() error {
	f.BaseCache.Clear()
	f.lock.Lock()
	f.allWildcardRules.Clear()
	f.rules = container.NewSyncMap[string, *rules.FaultDetectRule]()
	f.nsWildcardRules = make(map[string]*rules.ServiceWithFaultDetectRules)
	f.svcSpecificRules = make(map[string]map[string]*rules.ServiceWithFaultDetectRules)
	f.lock.Unlock()
	return nil
}

// Name 实现资源名称
func (f *faultDetectCache) Name() string {
	return types.FaultDetectRuleName
}

// GetFaultDetectConfig 根据serviceID获取探测规则
func (f *faultDetectCache) GetFaultDetectConfig(name string, namespace string) *rules.ServiceWithFaultDetectRules {
	log.Infof("GetFaultDetectConfig: name %s, namespace %s", name, namespace)
	// check service specific
	rules := f.checkServiceSpecificCache(name, namespace)
	if nil != rules {
		return rules
	}
	rules = f.checkNamespaceSpecificCache(namespace)
	if nil != rules {
		return rules
	}
	return f.allWildcardRules
}

func (f *faultDetectCache) checkServiceSpecificCache(
	name string, namespace string) *rules.ServiceWithFaultDetectRules {
	f.lock.RLock()
	defer f.lock.RUnlock()
	log.Infof(
		"checkServiceSpecificCache name %s, namespace %s, values %v", name, namespace, f.svcSpecificRules)
	svcRules, ok := f.svcSpecificRules[namespace]
	if ok {
		return svcRules[name]
	}
	return nil
}

func (f *faultDetectCache) checkNamespaceSpecificCache(namespace string) *rules.ServiceWithFaultDetectRules {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.nsWildcardRules[namespace]
}

func (f *faultDetectCache) reloadRevision(svcRules *rules.ServiceWithFaultDetectRules) {
	rulesCount := svcRules.CountFaultDetectRules()
	if rulesCount == 0 {
		svcRules.Revision = ""
		return
	}
	revisions := make([]string, 0, rulesCount)
	svcRules.IterateFaultDetectRules(func(rule *rules.FaultDetectRule) {
		revisions = append(revisions, rule.Revision)
	})
	sort.Strings(revisions)
	h := sha1.New()
	revision, err := types.ComputeRevisionBySlice(h, revisions)
	if err != nil {
		log.Errorf("[Server][Service][FaultDetector] compute revision service(%s) err: %s",
			svcRules.Service, err.Error())
		return
	}
	svcRules.Revision = revision
}

func (f *faultDetectCache) deleteAndReloadFaultDetectRules(svcRules *rules.ServiceWithFaultDetectRules, id string) {
	svcRules.DelFaultDetectRule(id)
	f.reloadRevision(svcRules)
}

func (f *faultDetectCache) deleteFaultDetectRuleFromServiceCache(id string, svcKeys map[svctypes.ServiceKey]bool) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if len(svcKeys) == 0 {
		// all wildcard
		f.deleteAndReloadFaultDetectRules(f.allWildcardRules, id)
		for _, rules := range f.nsWildcardRules {
			f.deleteAndReloadFaultDetectRules(rules, id)
		}
		for _, svcRules := range f.svcSpecificRules {
			for _, rules := range svcRules {
				f.deleteAndReloadFaultDetectRules(rules, id)
			}
		}
		return
	}
	svcToReloads := make(map[svctypes.ServiceKey]bool)
	for svcKey := range svcKeys {
		if svcKey.Name == types.AllMatched {
			rules, ok := f.nsWildcardRules[svcKey.Namespace]
			if ok {
				f.deleteAndReloadFaultDetectRules(rules, id)
			}
			svcRules, ok := f.svcSpecificRules[svcKey.Namespace]
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
			svcRules, ok := f.svcSpecificRules[svcToReload.Namespace]
			if ok {
				rules, ok := svcRules[svcToReload.Name]
				if ok {
					f.deleteAndReloadFaultDetectRules(rules, id)
				}
			}
		}
	}
}

func (f *faultDetectCache) storeAndReloadFaultDetectRules(
	svcRules *rules.ServiceWithFaultDetectRules, cbRule *rules.FaultDetectRule) {
	svcRules.AddFaultDetectRule(cbRule)
	f.reloadRevision(svcRules)
}

func createAndStoreServiceWithFaultDetectRules(svcKey svctypes.ServiceKey, key string,
	values map[string]*rules.ServiceWithFaultDetectRules) *rules.ServiceWithFaultDetectRules {
	rules := rules.NewServiceWithFaultDetectRules(svcKey)
	values[key] = rules
	return rules
}

func (f *faultDetectCache) storeFaultDetectRuleToServiceCache(
	entry *rules.FaultDetectRule, svcKeys map[svctypes.ServiceKey]bool) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if len(svcKeys) == 0 {
		// all wildcard
		f.storeAndReloadFaultDetectRules(f.allWildcardRules, entry)
		for _, rules := range f.nsWildcardRules {
			f.storeAndReloadFaultDetectRules(rules, entry)
		}
		for _, svcRules := range f.svcSpecificRules {
			for _, rules := range svcRules {
				f.storeAndReloadFaultDetectRules(rules, entry)
			}
		}
		return
	}
	svcToReloads := make(map[svctypes.ServiceKey]bool)
	for svcKey := range svcKeys {
		if svcKey.Name == types.AllMatched {
			var wildcardRules *rules.ServiceWithFaultDetectRules
			var ok bool
			wildcardRules, ok = f.nsWildcardRules[svcKey.Namespace]
			if !ok {
				wildcardRules = createAndStoreServiceWithFaultDetectRules(svcKey, svcKey.Namespace, f.nsWildcardRules)
			}
			f.storeAndReloadFaultDetectRules(wildcardRules, entry)
			svcRules, ok := f.svcSpecificRules[svcKey.Namespace]
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
			var detectrules *rules.ServiceWithFaultDetectRules
			var svcRules map[string]*rules.ServiceWithFaultDetectRules
			var ok bool
			svcRules, ok = f.svcSpecificRules[svcToReload.Namespace]
			if !ok {
				svcRules = make(map[string]*rules.ServiceWithFaultDetectRules)
				f.svcSpecificRules[svcToReload.Namespace] = svcRules
			}
			detectrules, ok = svcRules[svcToReload.Name]
			if !ok {
				detectrules = createAndStoreServiceWithFaultDetectRules(svcToReload, svcToReload.Name, svcRules)
			}
			f.storeAndReloadFaultDetectRules(detectrules, entry)
		}
	}
}

func getServicesInvolveByFaultDetectRule(fdRule *rules.FaultDetectRule) map[svctypes.ServiceKey]bool {
	svcKeys := make(map[svctypes.ServiceKey]bool)
	addService := func(name string, namespace string) {
		if name == types.AllMatched && namespace == types.AllMatched {
			return
		}
		svcKeys[svctypes.ServiceKey{
			Namespace: namespace,
			Name:      name,
		}] = true
	}
	addService(fdRule.DstService, fdRule.DstNamespace)
	return svcKeys
}

// setCircuitBreaker 更新store的数据到cache中
func (f *faultDetectCache) setFaultDetectRules(fdRules []*rules.FaultDetectRule) map[string]time.Time {
	if len(fdRules) == 0 {
		return nil
	}

	lastMtime := f.LastMtime(f.Name()).Unix()

	for _, fdRule := range fdRules {
		oldRule, ok := f.rules.Load(fdRule.ID)
		if ok {
			// 对比规则前后绑定的服务是否出现了变化，清理掉之前所绑定的信息数据
			if oldRule.IsServiceChange(fdRule) {
				// 从老的规则中获取所有的 svcKeys 信息列表
				svcKeys := getServicesInvolveByFaultDetectRule(oldRule)
				log.Info("[Cache][FaultDetect] clean rule bind old service info",
					zap.String("svc-keys", fmt.Sprintf("%#v", svcKeys)), zap.String("rule-id", fdRule.ID))
				// 挨个清空
				f.deleteFaultDetectRuleFromServiceCache(fdRule.ID, svcKeys)
			}
		}

		if fdRule.ModifyTime.Unix() > lastMtime {
			lastMtime = fdRule.ModifyTime.Unix()
		}
		svcKeys := getServicesInvolveByFaultDetectRule(fdRule)
		if !fdRule.Valid {
			f.rules.Delete(fdRule.ID)
			f.deleteFaultDetectRuleFromServiceCache(fdRule.ID, svcKeys)
			continue
		}
		f.rules.Store(fdRule.ID, fdRule)
		f.storeFaultDetectRuleToServiceCache(fdRule, svcKeys)
	}

	return map[string]time.Time{
		f.Name(): time.Unix(lastMtime, 0),
	}
}

// GetFaultDetectRuleCount 获取探测规则总数
func (f *faultDetectCache) GetFaultDetectRuleCount(fun func(k, v interface{}) bool) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	for k, v := range f.svcSpecificRules {
		if !fun(k, v) {
			break
		}
	}
}

var (
	ignoreFaultDetectRuleFilter = map[string]struct{}{
		"brief":            {},
		"service":          {},
		"serviceNamespace": {},
		"exactName":        {},
		"excludeId":        {},
	}

	fdBlurSearchFields = map[string]func(*rules.FaultDetectRule) string{
		"name": func(cbr *rules.FaultDetectRule) string {
			return cbr.Name
		},
		"description": func(cbr *rules.FaultDetectRule) string {
			return cbr.Description
		},
		"dstservice": func(cbr *rules.FaultDetectRule) string {
			return cbr.DstService
		},
		"dstmethod": func(cbr *rules.FaultDetectRule) string {
			return cbr.DstMethod
		},
	}

	faultDetectSort = map[string]func(asc bool, a, b *rules.FaultDetectRule) bool{
		"mtime": func(asc bool, a, b *rules.FaultDetectRule) bool {
			ret := a.ModifyTime.Before(b.ModifyTime)
			return ret && asc
		},
		"id": func(asc bool, a, b *rules.FaultDetectRule) bool {
			ret := a.ID < b.ID
			return ret && asc
		},
		"name": func(asc bool, a, b *rules.FaultDetectRule) bool {
			ret := a.Name < b.Name
			return ret && asc
		},
	}
)

// Query implements api.FaultDetectCache.
func (f *faultDetectCache) Query(ctx context.Context, args *types.FaultDetectArgs) (uint32, []*rules.FaultDetectRule, error) {
	if err := f.Update(); err != nil {
		return 0, nil, err
	}

	results := make([]*rules.FaultDetectRule, 0, 32)

	predicates := types.LoadFaultDetectRulePredicates(ctx)

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

	f.rules.ReadRange(func(key string, val *rules.FaultDetectRule) {
		if hasSvc && hasSvcNs {
			dstServiceValue := val.DstService
			dstNamespaceValue := val.DstNamespace
			if !(dstServiceValue == searchSvc && dstNamespaceValue == searchNs) {
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
			getter, isBlur := fdBlurSearchFields[fieldKey]
			if isBlur {
				if utils.IsWildMatch(getter(val), filterValue) {
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

	sortFunc, ok := faultDetectSort[args.Filter["order_field"]]
	if !ok {
		sortFunc = faultDetectSort["mtime"]
	}
	asc := "asc" == strings.ToLower(args.Filter["order_type"])
	sort.Slice(results, func(i, j int) bool {
		return sortFunc(asc, results[i], results[j])
	})

	total, ret := f.toPage(uint32(len(results)), results, args)
	return total, ret, nil
}

func (f *faultDetectCache) toPage(total uint32, items []*rules.FaultDetectRule,
	args *types.FaultDetectArgs) (uint32, []*rules.FaultDetectRule) {
	if args.Limit == 0 {
		return total, items
	}
	endIdx := args.Offset + args.Limit
	if endIdx > total {
		endIdx = total
	}
	return total, items[args.Offset:endIdx]
}

// GetRule implements api.FaultDetectCache.
func (f *faultDetectCache) GetRule(id string) *rules.FaultDetectRule {
	rule, _ := f.rules.Load(id)
	return rule
}
