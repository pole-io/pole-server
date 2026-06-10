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
	"time"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	protoV2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	types "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	revisionapi "github.com/pole-io/pole-server/apis/pkg/utils/revision"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	matchs "github.com/pole-io/pole-server/pkg/common/utils/match"
)

func NewLaneCache(storage store.Store, cacheMgr types.CacheManager) types.LaneCache {
	return &LaneCache{
		BaseCache: cachebase.NewBaseCache(storage, cacheMgr),
	}
}

type LaneCache struct {
	*cachebase.BaseCache
	// 这里的缓存用于列表页面查询
	// groups id -> *rules.LaneGroupProto
	rules *container.SyncMap[string, *rules.LaneGroup]
	// 这里的缓存用于客户端规则发布查询, 因为每个规则只会有一个发布状态
	// {rule-name + release_type} => *rules.LaneGroupRelease
	ids *container.SyncMap[string, *rules.LaneGroupRelease]
	// serviceRules namespace -> service -> []*rules.LaneRuleProto
	serviceRules *container.SyncMap[string, *container.SyncMap[string, *container.SyncMap[string, *rules.LaneGroupRelease]]]
	// revisions namespace -> service -> revision
	revisions *container.SyncMap[string, *container.SyncMap[string, string]]
}

// Initialize .
func (lc *LaneCache) Initialize(c map[string]interface{}) error {
	lc.rules = container.NewSyncMap[string, *rules.LaneGroup]()
	lc.ids = container.NewSyncMap[string, *rules.LaneGroupRelease]()
	lc.serviceRules = container.NewSyncMap[string, *container.SyncMap[string, *container.SyncMap[string, *rules.LaneGroupRelease]]]()
	lc.revisions = container.NewSyncMap[string, *container.SyncMap[string, string]]()
	registerGovernanceRuleWatcher(lc.Store(), lc.CacheMgr, lc)
	return nil
}

// Update .
func (lc *LaneCache) Update() error {
	if ok, err := updateGovernanceRuleCache(lc.Store()); ok {
		return err
	}
	// 多个线程竞争，只有一个线程进行更新
	err, _ := lc.singleUpdate()
	return err
}

func (lc *LaneCache) singleUpdate() (error, bool) {
	// 多个线程竞争，只有一个线程进行更新
	_, err, shared := lc.GetSingle().Do(lc.Name(), func() (interface{}, error) {
		return nil, lc.DoCacheUpdate(lc.Name(), lc.realUpdate)
	})
	return err, shared
}

// Update .
func (lc *LaneCache) realUpdate() (map[string]time.Time, int64, error) {
	start := time.Now()

	// 获取泳道规则信息
	rules, err := lc.Store().GetMoreLaneGroups(lc.LastFetchTime(), lc.IsFirstUpdate())
	if err != nil {
		log.Errorf("[cache][lane] query cache update err: %s", err.Error())
		return nil, -1, err
	}

	// 获取泳道发布信息
	// 这里的泳道发布信息是客户端发布的泳道规则
	// 需要注意的是，泳道规则的发布信息可能会有多个版本
	releases, err := lc.Store().GetMoreLaneGroupReleases(lc.IsFirstUpdate(), lc.LastFetchTime())
	if err != nil {
		log.Errorf("[cache][lane] publish rule cache update err: %s", err.Error())
		return nil, -1, err
	}

	cmtime, addCnt, updateCnt, delCnt := lc.setLaneRulesConsole(rules)
	log.Info("[cache][lane] console cache update",
		zap.Int("pull-from-store", len(rules)), zap.Int("add", addCnt), zap.Int("update", updateCnt),
		zap.Int("delete", delCnt), zap.Time("last", lc.LastMtime()), zap.Duration("used", time.Since(start)))

	pmtime, addCnt, updateCnt, delCnt := lc.setLaneRulesClient(releases)
	log.Info("[cache][lane] client cache update",
		zap.Int("pull-from-store", len(releases)), zap.Int("add", addCnt), zap.Int("update", updateCnt),
		zap.Int("delete", delCnt), zap.Time("last", lc.LastMtime()), zap.Duration("used", time.Since(start)))

	return map[string]time.Time{
		lc.Name() + "_console": cmtime,
		lc.Name() + "_client":  pmtime,
	}, int64(len(rules)), err
}

func (lc *LaneCache) setLaneRulesConsole(items map[string]*rules.LaneGroup) (time.Time, int, int, int) {
	lastMtime := lc.LastMtime().Unix()
	add := 0
	update := 0
	del := 0

	for i := range items {
		item := items[i]
		if item.ModifyTime.Unix() > lastMtime {
			lastMtime = item.ModifyTime.Unix()
		}
		if !item.Valid {
			del++
			_, _ = lc.rules.Delete(item.ID)
			continue
		}
		if _, exist := lc.rules.Load(item.ID); exist {
			update++
		} else {
			add++
		}
		lc.rules.Store(item.ID, item)
	}
	return time.Unix(lastMtime, 0), add, update, del
}

func (lc *LaneCache) setLaneRulesClient(items []*rules.LaneGroupRelease) (time.Time, int, int, int) {
	lastMtime := lc.LastMtime().Unix()
	add := 0
	update := 0
	del := 0
	affectSvcs := map[string]map[string]struct{}{}

	for i := range items {
		item := items[i]
		if item.Mtime.Unix() > lastMtime {
			lastMtime = item.Mtime.Unix()
		}

		oldVal, exist := lc.ids.Load(item.Key())
		if !item.Valid {
			if exist && oldVal.Version > item.Version {
				continue // 这里的版本号比当前的版本号大，说明是旧的发布信息
			}
			del++
			_, _ = lc.ids.Delete(item.Key())
			if exist {
				lc.processLaneRuleDelete(oldVal, affectSvcs)
			}
			continue
		}
		if exist {
			update++
		} else {
			add++
		}

		lc.ids.Store(item.Key(), item)
		lc.processLaneRuleUpsert(oldVal, item, affectSvcs)
	}
	lc.postUpdateRevisions(affectSvcs)
	return time.Unix(lastMtime, 0), add, update, del
}

func (lc *LaneCache) processLaneRuleUpsert(old, item *rules.LaneGroupRelease, affectSvcs map[string]map[string]struct{}) {
	waitDelServices := map[string]map[string]struct{}{}
	addService := func(ns, svc string) {
		if _, ok := waitDelServices[ns]; !ok {
			waitDelServices[ns] = map[string]struct{}{}
		}
		waitDelServices[ns][svc] = struct{}{}
	}
	removeServiceIfExist := func(ns, svc string) {
		waitDelServices[ns] = map[string]struct{}{}
		delete(waitDelServices[ns], svc)
	}

	handle := func(rule *rules.LaneGroupRelease, serviceOp func(ns, svc string), ruleOp func(string, string, *rules.LaneGroupRelease)) {
		if rule == nil {
			return
		}

		for i := range rule.Rule.Proto.Destinations {
			dest := rule.Rule.Proto.Destinations[i]
			serviceOp(dest.Namespace, dest.Service)
			ruleOp(dest.Namespace, dest.Service, rule)
		}

		for i := range rule.Rule.Proto.Entries {
			entry := rule.Rule.Proto.Entries[i]
			switch rules.TrafficEntryType(entry.Type) {
			case rules.TrafficEntry_MicroService:
				selector := &apitraffic.ServiceSelector{}
				if err := anyToSelector(entry.Selector, selector); err != nil {
					continue
				}
				serviceOp(selector.Namespace, selector.Service)
				ruleOp(selector.Namespace, selector.Service, rule)
			case rules.TrafficEntry_SpringCloudGateway:
				selector := &apitraffic.ServiceGatewaySelector{}
				if err := anyToSelector(entry.Selector, selector); err != nil {
					continue
				}
				serviceOp(selector.Namespace, selector.Service)
				ruleOp(selector.Namespace, selector.Service, rule)
			default:
				// do nothing
			}
		}
	}

	handle(item, addService, func(ns, svc string, group *rules.LaneGroupRelease) {
		if _, ok := affectSvcs[ns]; !ok {
			affectSvcs[ns] = map[string]struct{}{}
		}
		affectSvcs[ns][svc] = struct{}{}
		lc.upsertServiceRule(ns, svc, group)
	})
	handle(old, removeServiceIfExist, func(ns, svc string, group *rules.LaneGroupRelease) {
		if _, ok := affectSvcs[ns]; !ok {
			affectSvcs[ns] = map[string]struct{}{}
		}
		affectSvcs[ns][svc] = struct{}{}
	})

	for ns := range waitDelServices {
		for svc := range waitDelServices[ns] {
			lc.cleanServiceRule(ns, svc, old)
		}
	}
}

func (lc *LaneCache) processLaneRuleDelete(item *rules.LaneGroupRelease, affectSvcs map[string]map[string]struct{}) {
	message := item.Rule.Proto
	// 先清理 destinations
	for i := range message.Destinations {
		dest := message.Destinations[i]
		if _, ok := affectSvcs[dest.Namespace]; !ok {
			affectSvcs[dest.Namespace] = map[string]struct{}{}
		}
		affectSvcs[dest.Namespace][dest.Service] = struct{}{}
		lc.cleanServiceRule(dest.Namespace, dest.Service, item)
	}

	for i := range message.Entries {
		entry := message.Entries[i]
		var ns string
		var svc string
		switch rules.TrafficEntryType(entry.Type) {
		case rules.TrafficEntry_MicroService:
			selector := &apitraffic.ServiceSelector{}
			if err := anyToSelector(entry.Selector, selector); err != nil {
				continue
			}
			ns = selector.Namespace
			svc = selector.Service
			lc.cleanServiceRule(selector.Namespace, selector.Service, item)
		case rules.TrafficEntry_SpringCloudGateway:
			selector := &apitraffic.ServiceGatewaySelector{}
			if err := anyToSelector(entry.Selector, selector); err != nil {
				continue
			}
			ns = selector.Namespace
			svc = selector.Service
			lc.cleanServiceRule(selector.Namespace, selector.Service, item)
		}
		if _, ok := affectSvcs[ns]; !ok {
			affectSvcs[ns] = map[string]struct{}{}
		}
		affectSvcs[ns][svc] = struct{}{}
	}
}

func (lc *LaneCache) upsertServiceRule(namespace, service string, item *rules.LaneGroupRelease) {
	namespaceContainer, _ := lc.serviceRules.ComputeIfAbsent(namespace,
		func(k string) *container.SyncMap[string, *container.SyncMap[string, *rules.LaneGroupRelease]] {
			return container.NewSyncMap[string, *container.SyncMap[string, *rules.LaneGroupRelease]]()
		})
	serviceContainer, _ := namespaceContainer.ComputeIfAbsent(service,
		func(k string) *container.SyncMap[string, *rules.LaneGroupRelease] {
			return container.NewSyncMap[string, *rules.LaneGroupRelease]()
		})
	serviceContainer.Store(item.Key(), item)
}

func (lc *LaneCache) cleanServiceRule(namespace, service string, item *rules.LaneGroupRelease) {
	namespaceContainer, ok := lc.serviceRules.Load(namespace)
	if !ok {
		return
	}
	serviceContainer, ok := namespaceContainer.Load(service)
	if !ok {
		return
	}
	if item == nil {
		return
	}

	serviceContainer.Delete(item.Key())

	if serviceContainer.Len() == 0 {
		namespaceContainer.Delete(service)
	}
}

func (lc *LaneCache) postUpdateRevisions(affectSvcs map[string]map[string]struct{}) {
	for ns, svsList := range affectSvcs {
		nsContainer, ok := lc.serviceRules.Load(ns)
		if !ok {
			continue
		}
		lc.revisions.ComputeIfAbsent(ns, func(k string) *container.SyncMap[string, string] {
			return container.NewSyncMap[string, string]()
		})
		nsRevisions, _ := lc.revisions.Load(ns)
		for svc := range svsList {
			revisions := make([]string, 0, 32)
			svcContainer, ok := nsContainer.Load(svc)
			if !ok {
				continue
			}
			svcContainer.Range(func(key string, val *rules.LaneGroupRelease) {
				revisions = append(revisions, strconv.Itoa(int(val.Version)))
			})
			revision, err := revisionapi.CompositeComputeRevision(revisions)
			if err != nil {
				continue
			}
			nsRevisions.Store(svc, revision)
		}
	}
}

func (lc *LaneCache) GetLaneRules(serviceKey *svctypes.Service) ([]*rules.LaneGroupProto, string) {
	namespaceContainer, ok := lc.serviceRules.Load(serviceKey.Namespace)
	if !ok {
		return []*rules.LaneGroupProto{}, ""
	}
	serviceContainer, ok := namespaceContainer.Load(serviceKey.Name)
	if !ok {
		return []*rules.LaneGroupProto{}, ""
	}
	ret := make([]*rules.LaneGroupProto, 0, 32)
	serviceContainer.Range(func(ruleId string, val *rules.LaneGroupRelease) {
		ret = append(ret, val.Rule)
	})

	nsRevision, ok := lc.revisions.Load(serviceKey.Namespace)
	if !ok {
		return ret, ""
	}
	revision, _ := nsRevision.Load(serviceKey.Name)
	return ret, revision
}

func (lc *LaneCache) LastMtime() time.Time {
	return lc.BaseCache.LastMtime(lc.Name())
}

// Clear .
func (lc *LaneCache) Clear() error {
	resetGovernanceRuleUpdateCache(lc.Store())
	lc.rules = container.NewSyncMap[string, *rules.LaneGroup]()

	lc.ids = container.NewSyncMap[string, *rules.LaneGroupRelease]()
	lc.serviceRules = container.NewSyncMap[string, *container.SyncMap[string, *container.SyncMap[string, *rules.LaneGroupRelease]]]()
	lc.revisions = container.NewSyncMap[string, *container.SyncMap[string, string]]()
	return nil
}

// Name .
func (lc *LaneCache) Name() string {
	return types.LaneRuleName
}

func anyToSelector(data *anypb.Any, msg proto.Message) error {
	if err := anypb.UnmarshalTo(data, proto.MessageV2(msg),
		protoV2.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}); err != nil {
		return err
	}
	return nil
}

var (
	laneGroupSort = map[string]func(asc bool, a, b *rules.LaneGroupProto) bool{
		"mtime": func(asc bool, a, b *rules.LaneGroupProto) bool {
			ret := a.ModifyTime.Before(b.ModifyTime)
			return ret && asc
		},
		"id": func(asc bool, a, b *rules.LaneGroupProto) bool {
			ret := a.ID < b.ID
			return ret && asc
		},
		"name": func(asc bool, a, b *rules.LaneGroupProto) bool {
			ret := a.Name < b.Name
			return ret && asc
		},
	}
)

// Query implements api.LaneCache.
func (lc *LaneCache) Query(ctx context.Context, args *types.LaneGroupArgs) (uint32, []*rules.LaneGroupProto, error) {
	if err := lc.Update(); err != nil {
		return 0, nil, err
	}

	predicates := types.LoadLaneRulePredicates(ctx)

	searchName, hasName := args.Filter["name"]
	searchId, hasId := args.Filter["id"]

	results := make([]*rules.LaneGroupProto, 0, 32)

	lc.rules.ReadRange(func(key string, val *rules.LaneGroup) {
		if hasName && !matchs.IsWildMatch(val.Name, searchName) {
			return
		}
		if hasId && val.ID != searchId {
			return
		}

		for i := range predicates {
			if !predicates[i](ctx, val) {
				return
			}
		}

		protoVal, _ := val.ToProto()
		results = append(results, protoVal)
	})

	sortFunc, ok := laneGroupSort[args.Filter["order_field"]]
	if !ok {
		sortFunc = laneGroupSort["mtime"]
	}
	asc := strings.ToLower(args.Filter["order_type"]) == "asc"
	sort.Slice(results, func(i, j int) bool {
		return sortFunc(asc, results[i], results[j])
	})

	total, ret := lc.toPage(uint32(len(results)), results, args)
	return total, ret, nil
}

func (lc *LaneCache) toPage(total uint32, items []*rules.LaneGroupProto,
	args *types.LaneGroupArgs) (uint32, []*rules.LaneGroupProto) {
	if args.Limit == 0 {
		return total, items
	}
	endIdx := args.Offset + args.Limit
	if endIdx > total {
		endIdx = total
	}
	return total, items[args.Offset:endIdx]
}

// GetRule implements api.LaneCache.
func (lc *LaneCache) GetRule(id string) *rules.LaneGroup {
	rule, _ := lc.rules.Load(id)
	if rule == nil {
		return nil
	}
	return rule
}
