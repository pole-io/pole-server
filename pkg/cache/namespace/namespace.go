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

package namespace

import (
	"context"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"github.com/pole-io/pole-server/apis/store"
	types "github.com/pole-io/pole-server/pkg/cache/api"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/model"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

var (
	_ types.NamespaceCache = (*namespaceCache)(nil)
)

type namespaceCache struct {
	*types.BaseCache
	storage store.Store
	ids     *utils.SyncMap[string, *model.Namespace]
	updater *singleflight.Group
	// exportNamespace 某个命名空间下的所有服务的可见性
	exportNamespace *utils.SyncMap[string, *utils.SyncSet[string]]
}

func NewNamespaceCache(storage store.Store, cacheMgr types.CacheManager) types.NamespaceCache {
	return &namespaceCache{
		BaseCache: types.NewBaseCache(storage, cacheMgr),
		storage:   storage,
	}
}

// Initialize
func (nsCache *namespaceCache) Initialize(c map[string]interface{}) error {
	nsCache.ids = utils.NewSyncMap[string, *model.Namespace]()
	nsCache.updater = new(singleflight.Group)
	nsCache.exportNamespace = utils.NewSyncMap[string, *utils.SyncSet[string]]()
	return nil
}

// Update
func (nsCache *namespaceCache) Update() error {
	// 多个线程竞争，只有一个线程进行更新
	err, _ := nsCache.singleUpdate()
	return err
}

func (nsCache *namespaceCache) realUpdate() (map[string]time.Time, int64, error) {
	var (
		lastTime = nsCache.LastFetchTime()
		ret, err = nsCache.storage.GetMoreNamespaces(lastTime)
	)
	if err != nil {
		log.Error("[Cache][Namespace] get storage more", zap.Error(err))
		return nil, -1, err
	}
	lastMtimes := nsCache.setNamespaces(ret)
	return lastMtimes, int64(len(ret)), nil
}

func (nsCache *namespaceCache) setNamespaces(nsSlice []*model.Namespace) map[string]time.Time {
	lastMtime := nsCache.LastMtime(nsCache.Name()).Unix()

	for index := range nsSlice {
		ns := nsSlice[index]
		oldNs, hasOldVal := nsCache.ids.Load(ns.Name)
		eventType := eventhub.EventCreated
		if !ns.Valid {
			eventType = eventhub.EventDeleted
			nsCache.ids.Delete(ns.Name)
		} else {
			if !hasOldVal {
				eventType = eventhub.EventCreated
			} else {
				eventType = eventhub.EventUpdated
			}
			nsCache.ids.Store(ns.Name, ns)
		}
		nsCache.handleNamespaceChange(eventType, oldNs, ns)
		_ = eventhub.Publish(eventhub.CacheNamespaceEventTopic, &eventhub.CacheNamespaceEvent{
			OldItem:   oldNs,
			Item:      ns,
			EventType: eventType,
		})
		lastMtime = int64(math.Max(float64(lastMtime), float64(ns.ModifyTime.Unix())))
	}

	return map[string]time.Time{
		nsCache.Name(): time.Unix(lastMtime, 0),
	}
}

func (nsCache *namespaceCache) handleNamespaceChange(et eventhub.EventType, oldItem, item *model.Namespace) {
	switch et {
	case eventhub.EventUpdated, eventhub.EventCreated:
		exportTo := item.ServiceExportTo
		viewer := utils.NewSyncSet[string]()
		for i := range exportTo {
			viewer.Add(i)
		}
		nsCache.exportNamespace.Store(item.Name, viewer)
	case eventhub.EventDeleted:
		nsCache.exportNamespace.Delete(item.Name)
	}
}

func (nsCache *namespaceCache) GetVisibleNamespaces(namespace string) []*model.Namespace {
	ret := make(map[string]*model.Namespace, 8)

	// 根据命名空间级别的可见性进行查询
	// 先看精确的
	nsCache.exportNamespace.Range(func(exportNs string, viewerNs *utils.SyncSet[string]) {
		exactMatch := viewerNs.Contains(namespace)
		allMatch := viewerNs.Contains(types.AllMatched)
		if !exactMatch && !allMatch {
			return
		}
		val := nsCache.GetNamespace(exportNs)
		if val != nil {
			ret[val.Name] = val
		}
	})

	values := make([]*model.Namespace, 0, len(ret))
	for _, item := range ret {
		values = append(values, item)
	}
	return values
}

// Clear .
func (nsCache *namespaceCache) Clear() error {
	nsCache.BaseCache.Clear()
	nsCache.ids = utils.NewSyncMap[string, *model.Namespace]()
	nsCache.exportNamespace = utils.NewSyncMap[string, *utils.SyncSet[string]]()
	return nil
}

// Name .
func (nsCache *namespaceCache) Name() string {
	return types.NamespaceName
}

// GetNamespace get namespace by id
func (nsCache *namespaceCache) GetNamespace(id string) *model.Namespace {
	val, ok := nsCache.ids.Load(id)
	if !ok {
		return nil
	}
	return val
}

// GetNamespacesByName batch get namespace by name
func (nsCache *namespaceCache) GetNamespacesByName(names []string) []*model.Namespace {
	nsArr := make([]*model.Namespace, 0, len(names))
	for _, name := range names {
		if ns := nsCache.GetNamespace(name); ns != nil {
			nsArr = append(nsArr, ns)
		}
	}

	return nsArr
}

// GetNamespaceList
//
//	@receiver nsCache
//	@return []*model.Namespace
func (nsCache *namespaceCache) GetNamespaceList() []*model.Namespace {
	nsArr := make([]*model.Namespace, 0, 8)

	nsCache.ids.Range(func(key string, ns *model.Namespace) {
		nsArr = append(nsArr, ns)
	})

	return nsArr
}

// forceQueryUpdate 为了确保读取的数据是最新的，这里需要做一个强制 update 的动作进行数据读取处理
func (nsCache *namespaceCache) forceQueryUpdate() error {
	err, shared := nsCache.singleUpdate()
	// shared == true，表示当前已经有正在 update 执行的任务，这个任务不一定能够读取到最新的数据
	// 为了避免读取到脏数据，在发起一次 singleUpdate
	if shared {
		log.Debug("[Cache][Namespace] force query update from store")
		err, _ = nsCache.singleUpdate()
	}
	return err
}

func (nsCache *namespaceCache) singleUpdate() (error, bool) {
	// 多个线程竞争，只有一个线程进行更新
	_, err, shared := nsCache.updater.Do(nsCache.Name(), func() (interface{}, error) {
		return nil, nsCache.DoCacheUpdate(nsCache.Name(), nsCache.realUpdate)
	})
	return err, shared
}

func (nsCache *namespaceCache) Query(ctx context.Context, args *types.NamespaceArgs) (uint32, []*model.Namespace, error) {
	if err := nsCache.forceQueryUpdate(); err != nil {
		return 0, nil, err
	}

	ret := make([]*model.Namespace, 0, 32)

	predicates := types.LoadNamespacePredicates(ctx)

	searchName, hasName := args.Filter["name"]
	searchOwner, hasOwner := args.Filter["owner"]

	nsCache.ids.ReadRange(func(key string, val *model.Namespace) {
		for i := range predicates {
			if !predicates[i](ctx, val) {
				return
			}
		}

		if hasName {
			matchOne := false
			for i := range searchName {
				if utils.IsWildMatch(val.Name, searchName[i]) {
					matchOne = true
					break
				}
			}
			// 如果没有匹配到，直接返回
			if !matchOne {
				return
			}
		}

		if hasOwner {
			matchOne := false
			for i := range searchOwner {
				if utils.IsWildMatch(val.Owner, searchOwner[i]) {
					matchOne = true
					break
				}
			}
			// 如果没有匹配到，直接返回
			if !matchOne {
				return
			}
		}

		ret = append(ret, val)
	})

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].ModifyTime.After(ret[j].ModifyTime)
	})

	total, ret := nsCache.toPage(len(ret), ret, args)
	return uint32(total), ret, nil
}

func (c *namespaceCache) toPage(total int, items []*model.Namespace,
	args *types.NamespaceArgs) (int, []*model.Namespace) {
	if len(items) == 0 {
		return 0, []*model.Namespace{}
	}
	if args.Limit == 0 {
		return total, items
	}
	endIdx := args.Offset + args.Limit
	if endIdx > total {
		endIdx = total
	}
	return total, items[args.Offset:endIdx]
}
