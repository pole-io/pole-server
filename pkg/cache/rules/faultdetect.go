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
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	types "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	revisionapi "github.com/pole-io/pole-server/apis/pkg/utils/revision"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	matchs "github.com/pole-io/pole-server/pkg/common/utils/match"
)

type faultDetectCache struct {
	*cachebase.BaseCache

	storage store.Store
	// rules 用于 console 查询
	rules *container.SyncMap[string, *rules.FaultDetectRule]
	// --------- 以下缓存均用于客户端数据查询 --------- //
	// increment cache
	ids *container.SyncMap[string, *rules.FaultDetectRelease]
	// fetched service cache
	// key1: namespace, key2: service
	svcSpecificRules *container.SyncMap[string, *container.SyncMap[string, *rules.ServiceWithFaultDetectRules]]
	// key1: namespace
	nsWildcardRules *container.SyncMap[string, *rules.ServiceWithFaultDetectRules]
	// all rules are wildcard specific
	allWildcardRules *rules.ServiceWithFaultDetectRules
	lock             sync.RWMutex
}

// NewFaultDetectCache faultDetectCache constructor
func NewFaultDetectCache(s store.Store, cacheMgr types.CacheManager) types.FaultDetectCache {
	return &faultDetectCache{
		BaseCache: cachebase.NewBaseCache(s, cacheMgr),
		storage:   s,
	}
}

// Initialize 实现Cache接口的函数
func (f *faultDetectCache) Initialize(_ map[string]interface{}) error {
	f.rules = container.NewSyncMap[string, *rules.FaultDetectRule]()
	f.ids = container.NewSyncMap[string, *rules.FaultDetectRelease]()
	f.svcSpecificRules = container.NewSyncMap[string, *container.SyncMap[string, *rules.ServiceWithFaultDetectRules]]()
	f.nsWildcardRules = container.NewSyncMap[string, *rules.ServiceWithFaultDetectRules]()
	f.allWildcardRules = rules.NewServiceWithFaultDetectRules(svctypes.ServiceKey{
		Namespace: types.AllMatched,
		Name:      types.AllMatched,
	})
	return nil
}

func (f *faultDetectCache) Update() error {
	_, err, _ := f.GetSingle().Do(f.Name(), func() (interface{}, error) {
		return nil, f.DoCacheUpdate(f.Name(), f.realUpdate)
	})
	return err
}

// update 实现Cache接口的函数
func (f *faultDetectCache) realUpdate() (map[string]time.Time, int64, error) {
	allLastTimes := map[string]time.Time{}

	fdRules, err := f.storage.GetMoreFaultDetects(f.LastFetchTime(), f.IsFirstUpdate())
	if err != nil {
		log.Errorf("[cache][fault_detect] console cache update err:%s", err.Error())
		return nil, -1, err
	}

	lastMtime, upsert, del := f.setFaultDetectConsole(fdRules)
	log.Info("[cache][fault_detect] console cache update",
		zap.Int("pull-from-store", len(fdRules)), zap.Int("upsert", upsert), zap.Int("delete", del),
		zap.Time("last", lastMtime))
	allLastTimes[f.Name()+"_console"] = lastMtime

	releases, err := f.storage.GetMoreFaultDetectReleases(f.LastFetchTime(), f.IsFirstUpdate())
	if err != nil {
		log.Errorf("[cache][fault_detect] release cache update err:%s", err.Error())
		return nil, -1, err
	}

	lastMtime, upsert, del = f.setFaultDetectClient(releases)
	log.Info("[cache][fault_detect] client cache update",
		zap.Int("pull-from-store", len(fdRules)), zap.Int("upsert", upsert), zap.Int("delete", del),
		zap.Time("last", lastMtime))
	allLastTimes[f.Name()+"_client"] = lastMtime

	return allLastTimes, int64(len(fdRules) + len(releases)), nil
}

// clear 实现Cache接口的函数
func (f *faultDetectCache) Clear() error {
	f.BaseCache.Clear()
	f.lock.Lock()
	f.rules = container.NewSyncMap[string, *rules.FaultDetectRule]()
	f.ids = container.NewSyncMap[string, *rules.FaultDetectRelease]()
	f.svcSpecificRules = container.NewSyncMap[string, *container.SyncMap[string, *rules.ServiceWithFaultDetectRules]]()
	f.nsWildcardRules = container.NewSyncMap[string, *rules.ServiceWithFaultDetectRules]()
	f.allWildcardRules.Clear()
	f.lock.Unlock()
	return nil
}

// Name 实现资源名称
func (f *faultDetectCache) Name() string {
	return types.FaultDetectRuleName
}

// GetFaultDetectConfig 根据serviceID获取探测规则
func (f *faultDetectCache) GetFaultDetectConfig(name string, namespace string) *rules.ServiceWithFaultDetectRules {
	log.Infof("[cache][fault_detect] GetFaultDetectConfig: name %s, namespace %s", name, namespace)
	// check service specific
	rules := f.checkServiceSpecificCache(name, namespace)
	if nil != rules {
		return rules
	}
	if rules, ok := f.nsWildcardRules.Load(namespace); ok {
		return rules
	}
	return f.allWildcardRules
}

func (f *faultDetectCache) checkServiceSpecificCache(
	name string, namespace string) *rules.ServiceWithFaultDetectRules {
	f.lock.RLock()
	defer f.lock.RUnlock()
	log.Infof(
		"[cache][fault_detect] checkServiceSpecificCache name %s, namespace %s, values %v", name, namespace, f.svcSpecificRules)
	svcRules, ok := f.svcSpecificRules.Load(namespace)
	if ok {
		return svcRules.MustLoad(name)
	}
	return nil
}

func (f *faultDetectCache) reloadRevision(svcRules *rules.ServiceWithFaultDetectRules) {
	rulesCount := svcRules.CountFaultDetectRules()
	if rulesCount == 0 {
		svcRules.Revision = ""
		return
	}
	revisions := make([]string, 0, rulesCount)
	svcRules.IterateFaultDetectRules(func(rule *rules.FaultDetectRelease) {
		revisions = append(revisions, strconv.Itoa(int(rule.Version)))
	})
	sort.Strings(revisions)
	h := sha1.New()
	revision, err := revisionapi.ComputeRevisionBySlice(h, revisions)
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
		// 遍历所有的 svc specific rules
		// 这里的 svc specific rules 是指 namespace 下的所有 service 的规则
		f.nsWildcardRules.Range(func(key string, rules *rules.ServiceWithFaultDetectRules) {
			f.deleteAndReloadFaultDetectRules(rules, id)
		})
		f.svcSpecificRules.Range(func(key string, svcRules *container.SyncMap[string, *rules.ServiceWithFaultDetectRules]) {
			svcRules.Range(func(svcName string, rules *rules.ServiceWithFaultDetectRules) {
				f.deleteAndReloadFaultDetectRules(rules, id)
			})
		})
		return
	}
	svcToReloads := make(map[svctypes.ServiceKey]bool)
	for svcKey := range svcKeys {
		if svcKey.Name == types.AllMatched {
			nsRules, ok := f.nsWildcardRules.Load(svcKey.Namespace)
			if ok {
				f.deleteAndReloadFaultDetectRules(nsRules, id)
			}
			if svcRules, ok := f.svcSpecificRules.Load(svcKey.Namespace); ok {
				svcRules.Range(func(svcName string, _ *rules.ServiceWithFaultDetectRules) {
					svcToReloads[svctypes.ServiceKey{Namespace: svcKey.Namespace, Name: svcName}] = true
				})
			}
		} else {
			svcToReloads[svcKey] = true
		}
	}
	if len(svcToReloads) > 0 {
		for svcToReload := range svcToReloads {
			if svcRules, ok := f.svcSpecificRules.Load(svcToReload.Namespace); ok {
				if rules, ok := svcRules.Load(svcToReload.Name); ok {
					f.deleteAndReloadFaultDetectRules(rules, id)
				}
			}
		}
	}
}

func (f *faultDetectCache) storeAndReloadFaultDetectRules(
	svcRules *rules.ServiceWithFaultDetectRules, cbRule *rules.FaultDetectRelease) {
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
	entry *rules.FaultDetectRelease, svcKeys map[svctypes.ServiceKey]bool) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if len(svcKeys) == 0 {
		// all wildcard
		f.storeAndReloadFaultDetectRules(f.allWildcardRules, entry)
		f.nsWildcardRules.Range(func(key string, rules *rules.ServiceWithFaultDetectRules) {
			f.storeAndReloadFaultDetectRules(rules, entry)
		})
		f.svcSpecificRules.Range(func(key string, svcRules *container.SyncMap[string, *rules.ServiceWithFaultDetectRules]) {
			svcRules.Range(func(svcName string, rules *rules.ServiceWithFaultDetectRules) {
				f.storeAndReloadFaultDetectRules(rules, entry)
			})
		})
		return
	}
	svcToReloads := make(map[svctypes.ServiceKey]bool)
	for svcKey := range svcKeys {
		if svcKey.Name == types.AllMatched {
			wildcardRules, _ := f.nsWildcardRules.ComputeIfAbsent(svcKey.Namespace, func(k string) *rules.ServiceWithFaultDetectRules {
				return rules.NewServiceWithFaultDetectRules(svcKey)
			})
			f.storeAndReloadFaultDetectRules(wildcardRules, entry)
			if svcRules, ok := f.svcSpecificRules.Load(svcKey.Namespace); ok {
				svcRules.Range(func(key string, val *rules.ServiceWithFaultDetectRules) {
					svcToReloads[svctypes.ServiceKey{Namespace: svcKey.Namespace, Name: key}] = true
				})
			}
		} else {
			svcToReloads[svcKey] = true
		}
	}
	if len(svcToReloads) > 0 {
		for svcToReload := range svcToReloads {
			svcRules, _ := f.svcSpecificRules.ComputeIfAbsent(svcToReload.Namespace, func(k string) *container.SyncMap[string, *rules.ServiceWithFaultDetectRules] {
				return container.NewSyncMap[string, *rules.ServiceWithFaultDetectRules]()
			})
			detectrules, _ := svcRules.ComputeIfAbsent(svcToReload.Name, func(k string) *rules.ServiceWithFaultDetectRules {
				return rules.NewServiceWithFaultDetectRules(svcToReload)
			})
			f.storeAndReloadFaultDetectRules(detectrules, entry)
		}
	}
}

func getServicesInvolveByFaultDetectRule(fdRule *rules.FaultDetectRelease) map[svctypes.ServiceKey]bool {
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
	addService(fdRule.Rule.DstService, fdRule.Rule.DstNamespace)
	return svcKeys
}

// setFaultDetectConsole 更新store的数据到cache中
func (f *faultDetectCache) setFaultDetectConsole(fdRules []*rules.FaultDetectRule) (time.Time, int, int) {
	if len(fdRules) == 0 {
		return time.Time{}, 0, 0
	}

	upsert := 0
	del := 0
	lastMtime := f.LastMtime(f.Name()).Unix()

	for _, fdRule := range fdRules {
		if fdRule.ModifyTime.Unix() > lastMtime {
			lastMtime = fdRule.ModifyTime.Unix()
		}
		if !fdRule.Valid {
			del++
			f.rules.Delete(fdRule.ID)
			continue
		}
		f.rules.Store(fdRule.ID, fdRule)
		upsert++
	}
	return time.Unix(lastMtime, 0), upsert, del
}

// setFaultDetectClient 更新store的数据到cache中
func (f *faultDetectCache) setFaultDetectClient(fdRules []*rules.FaultDetectRelease) (time.Time, int, int) {
	if len(fdRules) == 0 {
		return time.Time{}, 0, 0
	}

	upsert := 0
	del := 0
	lastMtime := f.LastMtime(f.Name()).Unix()

	for _, fdRule := range fdRules {
		oldRule, ok := f.ids.Load(fdRule.Key())
		if ok {
			// 对比规则前后绑定的服务是否出现了变化，清理掉之前所绑定的信息数据
			if oldRule.Rule.IsServiceChange(fdRule.Rule) {
				// 从老的规则中获取所有的 svcKeys 信息列表
				svcKeys := getServicesInvolveByFaultDetectRule(oldRule)
				log.Info("[cache][fault_detect] clean rule bind old service info",
					zap.String("svc-keys", fmt.Sprintf("%#v", svcKeys)), zap.String("rule-id", fdRule.Key()))
				// 挨个清空
				f.deleteFaultDetectRuleFromServiceCache(fdRule.Key(), svcKeys)
			}
		}

		if fdRule.Mtime.Unix() > lastMtime {
			lastMtime = fdRule.Mtime.Unix()
		}
		svcKeys := getServicesInvolveByFaultDetectRule(fdRule)
		if !fdRule.Valid {
			del++
			f.ids.Delete(fdRule.Key())
			f.deleteFaultDetectRuleFromServiceCache(fdRule.Key(), svcKeys)
			continue
		}
		f.ids.Store(fdRule.Key(), fdRule)
		f.storeFaultDetectRuleToServiceCache(fdRule, svcKeys)
		upsert++
	}

	return time.Unix(lastMtime, 0), upsert, del
}

var (
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
				if matchs.IsWildMatch(getter(val), filterValue) {
					return
				}
			} else if fieldKey == "namespace" {
				// 精确匹配命名空间
				if filterValue != val.Namespace {
					return
				}
			} else if fieldKey == "dstnamespace" {
				// 精确匹配目标命名空间
				if filterValue != val.DstNamespace {
					return
				}
			} else if fieldKey == "valid" {
				// 精确匹配有效性状态
				validStr := strconv.FormatBool(val.Valid)
				if filterValue != validStr {
					return
				}
			} else if fieldKey == "revision" {
				// 精确匹配版本号
				if filterValue != val.Revision {
					return
				}
			}
			// 其他未知字段忽略，保持向后兼容
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
