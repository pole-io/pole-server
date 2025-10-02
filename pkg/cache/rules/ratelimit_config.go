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
	"sync"
	"time"

	cachetypes "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
)

// rateLimitCache的实现
type rateLimitCache struct {
	*cachebase.BaseCache

	lock         sync.RWMutex
	waitFixRules map[string]struct{}
	svcCache     cachetypes.ServiceCache
	storage      store.Store

	// 用于控制台查询列表
	ids *container.SyncMap[string, *rules.RateLimit]

	// 用于客户端查询规则缓存
	// increment cache
	rules *container.SyncMap[string, *rules.RateLimitRelease]
	// fetched service cache
	// key1: namespace, key2: service
	svcSpecificRules *container.SyncMap[string, *container.SyncMap[string, *rules.ServiceWithRateLimits]]
	// key1: namespace
	nsWildcardRules *container.SyncMap[string, *rules.ServiceWithRateLimits]
	// all rules are wildcard specific
	allWildcardRules *rules.ServiceWithRateLimits
}

// NewRateLimitCache 返回一个操作RateLimitCache的对象
func NewRateLimitCache(s store.Store, cacheMgr cachetypes.CacheManager) cachetypes.RateLimitCache {
	return &rateLimitCache{
		BaseCache:    cachebase.NewBaseCache(s, cacheMgr),
		storage:      s,
		waitFixRules: map[string]struct{}{},
	}
}

// Initialize 实现Cache接口的initialize函数
func (rlc *rateLimitCache) Initialize(_ map[string]interface{}) error {
	rlc.ids = container.NewSyncMap[string, *rules.RateLimit]()
	rlc.rules = container.NewSyncMap[string, *rules.RateLimitRelease]()
	rlc.svcSpecificRules = container.NewSyncMap[string, *container.SyncMap[string, *rules.ServiceWithRateLimits]]()
	rlc.nsWildcardRules = container.NewSyncMap[string, *rules.ServiceWithRateLimits]()
	rlc.allWildcardRules = rules.NewServiceWithRateLimits(svctypes.ServiceKey{
		Namespace: cachetypes.AllMatched,
		Name:      cachetypes.AllMatched,
	})
	rlc.svcCache = rlc.CacheMgr.GetCacher(cachetypes.CacheService).(cachetypes.ServiceCache)
	return nil
}

// Update 实现Cache接口的update函数
func (rlc *rateLimitCache) Update() error {
	// 多个线程竞争，只有一个线程进行更新
	_, err, _ := rlc.GetSingle().Do(rlc.Name(), func() (interface{}, error) {
		return nil, rlc.DoCacheUpdate(rlc.Name(), rlc.realUpdate)
	})
	return err
}

func (rlc *rateLimitCache) realUpdate() (map[string]time.Time, int64, error) {
	rateLimits, err := rlc.storage.GetMoreRateLimits(rlc.LastFetchTime(), rlc.IsFirstUpdate())
	if err != nil {
		log.Errorf("[cache][rate_limit] console cache update err: %s", err.Error())
		return nil, -1, err
	}
	clastMtimes := rlc.setRateLimitConsole(rateLimits)

	releases, err := rlc.storage.GetMoreRateLimitReleases(rlc.LastFetchTime(), rlc.IsFirstUpdate())
	if err != nil {
		log.Errorf("[cache][rate_limit] client cache update err: %s", err.Error())
		return nil, -1, err
	}
	plastMtimes := rlc.setRateLimitClient(releases)

	lastMtimes := map[string]time.Time{}
	if !clastMtimes.IsZero() {
		lastMtimes[rlc.Name()+"_console"] = clastMtimes
	}
	if !plastMtimes.IsZero() {
		lastMtimes[rlc.Name()] = plastMtimes
	}

	return lastMtimes, int64(len(rateLimits) + len(releases)), nil
}

// Name 获取资源名称
func (rlc *rateLimitCache) Name() string {
	return cachetypes.RateLimitConfigName
}

// Clear 实现Cache接口的clear函数
func (rlc *rateLimitCache) Clear() error {
	rlc.BaseCache.Clear()
	rlc.ids = container.NewSyncMap[string, *rules.RateLimit]()
	rlc.rules = container.NewSyncMap[string, *rules.RateLimitRelease]()
	rlc.svcSpecificRules = container.NewSyncMap[string, *container.SyncMap[string, *rules.ServiceWithRateLimits]]()
	rlc.nsWildcardRules = container.NewSyncMap[string, *rules.ServiceWithRateLimits]()
	rlc.allWildcardRules = rules.NewServiceWithRateLimits(svctypes.ServiceKey{
		Namespace: cachetypes.AllMatched,
		Name:      cachetypes.AllMatched,
	})
	return nil
}

func (rlc *rateLimitCache) toProto(item *rules.RateLimit) {
	item.ToSpec()
	namespace := item.Proto.GetNamespace().GetValue()
	name := item.Proto.GetService().GetValue()
	if namespace == "" || name == "" {
		rlc.fixRuleServiceInfo(item)
	}
}

// setRateLimitConsole 更新限流规则到缓存中
func (rlc *rateLimitCache) setRateLimitConsole(rateLimits []*rules.RateLimit) time.Time {
	if len(rateLimits) == 0 {
		return time.Time{}
	}
	lastMtime := rlc.LastMtime(rlc.Name() + "_console").Unix()
	for _, item := range rateLimits {
		if item.ModifyTime.Unix() > lastMtime {
			lastMtime = item.ModifyTime.Unix()
		}

		// 待删除的rateLimit
		if !item.Valid {
			rlc.ids.Delete(item.ID)
			continue
		}
		rlc.ids.Store(item.ID, item)
	}

	return time.Unix(lastMtime, 0)
}

// setRateLimitClient 更新限流规则到缓存中
func (rlc *rateLimitCache) setRateLimitClient(rateLimits []*rules.RateLimitRelease) time.Time {
	if len(rateLimits) == 0 {
		return time.Time{}
	}
	rlc.fixRulesServiceInfo()
	lastMtime := rlc.LastMtime(rlc.Name()).Unix()
	for _, item := range rateLimits {
		rlc.toProto(item.Rule)
		if item.Mtime.Unix() > lastMtime {
			lastMtime = item.Mtime.Unix()
		}

		// 待删除的rateLimit
		if !item.Valid {
			rlc.rules.Delete(item.Key())
			rlc.deleteWaitFixRule(item)
			continue
		}
		rlc.rules.Store(item.Key(), item)
	}

	return time.Unix(lastMtime, 0)
}

// IteratorRateLimit 根据serviceID进行迭代回调
func (rlc *rateLimitCache) IteratorRateLimit(proc cachetypes.RateLimitIterProc) {
	rlc.ids.Range(func(key string, val *rules.RateLimit) {
		proc(val)
	})
}

// GetRateLimitByServiceID 根据serviceID获取限流数据
func (rlc *rateLimitCache) GetRateLimitRules(serviceKey svctypes.ServiceKey) ([]*rules.RateLimit, string) {
	// 获取对应服务的限流规则配置
	svcRules := rlc.GetRateLimitConfig(serviceKey.Name, serviceKey.Namespace)
	if svcRules == nil {
		return nil, ""
	}

	// 收集所有规则
	var rateLimitRules []*rules.RateLimit
	svcRules.Rules.Range(func(key string, release *rules.RateLimitRelease) {
		if release != nil && release.Rule != nil {
			rateLimitRules = append(rateLimitRules, release.Rule)
		}
	})

	return rateLimitRules, svcRules.Revision
}

// GetRateLimitsCount 获取限流规则总数
func (rlc *rateLimitCache) GetRateLimitsCount() int {
	return rlc.ids.Len()
}

func (rlc *rateLimitCache) deleteWaitFixRule(rule *rules.RateLimitRelease) {
	rlc.lock.Lock()
	defer rlc.lock.Unlock()
	delete(rlc.waitFixRules, rule.Key())
}

func (rlc *rateLimitCache) fixRulesServiceInfo() {
	rlc.lock.Lock()
	defer rlc.lock.Unlock()
	for id := range rlc.waitFixRules {
		rule, ok := rlc.rules.Load(id)
		if !ok {
			delete(rlc.waitFixRules, id)
			continue
		}
		svcId := rule.Rule.ServiceID
		svc := rlc.svcCache.GetServiceByID(svcId)
		if svc == nil {
			svc2, err := rlc.storage.GetServiceByID(svcId)
			if err != nil {
				continue
			}
			svc = svc2
		}
		if svc != nil {
			rule.Rule.Proto.Namespace = protobuf.NewStringValue(svc.Namespace)
			rule.Rule.Proto.Name = protobuf.NewStringValue(svc.Name)
			delete(rlc.waitFixRules, rule.Key())
		}
	}
}

func (rlc *rateLimitCache) fixRuleServiceInfo(rateLimit *rules.RateLimit) {
	rlc.lock.Lock()
	defer rlc.lock.Unlock()
	svcId := rateLimit.ServiceID
	svc := rlc.svcCache.GetServiceByID(svcId)
	if svc == nil {
		svc2, err := rlc.storage.GetServiceByID(svcId)
		if err != nil {
			rlc.waitFixRules[rateLimit.ID] = struct{}{}
			return
		}
		if svc2 == nil {
			// 存储层确实不存在，直接跳过
			delete(rlc.waitFixRules, rateLimit.ID)
			return
		}
		svc = svc2
	}

	if svc != nil {
		rateLimit.Proto.Namespace = protobuf.NewStringValue(svc.Namespace)
		rateLimit.Proto.Name = protobuf.NewStringValue(svc.Name)
	}
	delete(rlc.waitFixRules, rateLimit.ID)
}

// GetRule implements api.RateLimitCache.
func (rlc *rateLimitCache) GetRule(id string) *rules.RateLimit {
	rule, _ := rlc.ids.Load(id)
	return rule
}

// GetRateLimitConfig 根据 service name 和 namespace 获取限流规则
func (rlc *rateLimitCache) GetRateLimitConfig(name string, namespace string) *rules.ServiceWithRateLimits {
	log.Infof("[cache][rate_limit] GetRateLimitConfig: name %s, namespace %s", name, namespace)
	// check service specific
	rules := rlc.checkServiceSpecificCache(name, namespace)
	if nil != rules {
		return rules
	}
	if rules, ok := rlc.nsWildcardRules.Load(namespace); ok {
		return rules
	}
	return rlc.allWildcardRules
}

func (rlc *rateLimitCache) checkServiceSpecificCache(
	name string, namespace string) *rules.ServiceWithRateLimits {
	rlc.lock.RLock()
	defer rlc.lock.RUnlock()
	log.Infof(
		"[cache][rate_limit] checkServiceSpecificCache name %s, namespace %s, values %v", name, namespace, rlc.svcSpecificRules)
	svcRules, ok := rlc.svcSpecificRules.Load(namespace)
	if ok {
		if rule, exists := svcRules.Load(name); exists {
			return rule
		}
	}
	return nil
}
