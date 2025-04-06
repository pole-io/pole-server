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

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

var (
	_ cacheapi.NamespaceCache = (*namespaceCache)(nil)
)

type namespaceCache struct {
	*cachebase.BaseCache
	storage store.Store
	ids     *container.SyncMap[string, *types.Namespace]
	updater *singleflight.Group
	// exportNamespace 某个命名空间下的所有服务的可见性
	exportNamespace *container.SyncMap[string, *container.SyncSet[string]]
}

func NewNamespaceCache(storage store.Store, cacheMgr cacheapi.CacheManager) cacheapi.NamespaceCache {
	return &namespaceCache{
		BaseCache: cachebase.NewBaseCache(storage, cacheMgr),
		storage:   storage,
	}
}

// Initialize
func (nsCache *namespaceCache) Initialize(c map[string]interface{}) error {
	nsCache.ids = container.NewSyncMap[string, *types.Namespace]()
	nsCache.updater = new(singleflight.Group)
	nsCache.exportNamespace = container.NewSyncMap[string, *container.SyncSet[string]]()
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

func (nsCache *namespaceCache) setNamespaces(nsSlice []*types.Namespace) map[string]time.Time {
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

func (nsCache *namespaceCache) handleNamespaceChange(et eventhub.EventType, oldItem, item *types.Namespace) {
	switch et {
	case eventhub.EventUpdated, eventhub.EventCreated:
		exportTo := item.ServiceExportTo
		viewer := container.NewSyncSet[string]()
		for i := range exportTo {
			viewer.Add(i)
		}
		nsCache.exportNamespace.Store(item.Name, viewer)
	case eventhub.EventDeleted:
		nsCache.exportNamespace.Delete(item.Name)
	}
}

func (nsCache *namespaceCache) GetVisibleNamespaces(namespace string) []*types.Namespace {
	ret := make(map[string]*types.Namespace, 8)

	// 根据命名空间级别的可见性进行查询
	// 先看精确的
	nsCache.exportNamespace.Range(func(exportNs string, viewerNs *container.SyncSet[string]) {
		exactMatch := viewerNs.Contains(namespace)
		allMatch := viewerNs.Contains(cacheapi.AllMatched)
		if !exactMatch && !allMatch {
			return
		}
		val := nsCache.GetNamespace(exportNs)
		if val != nil {
			ret[val.Name] = val
		}
	})

	values := make([]*types.Namespace, 0, len(ret))
	for _, item := range ret {
		values = append(values, item)
	}
	return values
}

// Clear .
func (nsCache *namespaceCache) Clear() error {
	nsCache.BaseCache.Clear()
	nsCache.ids = container.NewSyncMap[string, *types.Namespace]()
	nsCache.exportNamespace = container.NewSyncMap[string, *container.SyncSet[string]]()
	return nil
}

// Name .
func (nsCache *namespaceCache) Name() string {
	return cacheapi.NamespaceName
}

// GetNamespace get namespace by id
func (nsCache *namespaceCache) GetNamespace(id string) *types.Namespace {
	val, ok := nsCache.ids.Load(id)
	if !ok {
		return nil
	}
	return val
}

// GetNamespacesByName batch get namespace by name
func (nsCache *namespaceCache) GetNamespacesByName(names []string) []*types.Namespace {
	nsArr := make([]*types.Namespace, 0, len(names))
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
//	@return []*types.Namespace
func (nsCache *namespaceCache) GetNamespaceList() []*types.Namespace {
	nsArr := make([]*types.Namespace, 0, 8)

	nsCache.ids.Range(func(key string, ns *types.Namespace) {
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

func (nsCache *namespaceCache) Query(ctx context.Context, args *cacheapi.NamespaceArgs) (uint32, []*types.Namespace, error) {
	if err := nsCache.forceQueryUpdate(); err != nil {
		return 0, nil, err
	}

	ret := make([]*types.Namespace, 0, 32)

	predicates := cacheapi.LoadNamespacePredicates(ctx)

	searchName, hasName := args.Filter["name"]
	searchOwner, hasOwner := args.Filter["owner"]

	nsCache.ids.ReadRange(func(key string, val *types.Namespace) {
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

func (c *namespaceCache) toPage(total int, items []*types.Namespace,
	args *cacheapi.NamespaceArgs) (int, []*types.Namespace) {
	if len(items) == 0 {
		return 0, []*types.Namespace{}
	}
	if args.Limit == 0 {
		return total, items
	}
	if args.Offset > total {
		return total, []*types.Namespace{}
	}
	endIdx := args.Offset + args.Limit
	if endIdx > total {
		endIdx = total
	}
	return total, items[args.Offset:endIdx]
}
