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

package ai

import (
	"time"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"golang.org/x/sync/singleflight"
)

type skillSubscriptionCache struct {
	*cachebase.BaseCache
	storage store.Store

	// id -> *ai.SkillSubscription
	ids *container.SyncMap[string, *ai.SkillSubscription]

	// clientID -> []*ai.SkillSubscription
	clientSubscriptions *container.SyncMap[string, []*ai.SkillSubscription]

	// namespace/skillName -> []*ai.SkillSubscription
	skillSubscriptions *container.SyncMap[string, []*ai.SkillSubscription]

	singleFlight *singleflight.Group
}

func NewSkillSubscriptionCache(storage store.Store, cacheMgr cacheapi.CacheManager) cacheapi.Cache {
	return &skillSubscriptionCache{
		BaseCache:           cachebase.NewBaseCache(storage, cacheMgr),
		storage:             storage,
		ids:                container.NewSyncMap[string, *ai.SkillSubscription](),
		clientSubscriptions: container.NewSyncMap[string, []*ai.SkillSubscription](),
		skillSubscriptions:  container.NewSyncMap[string, []*ai.SkillSubscription](),
		singleFlight:       &singleflight.Group{},
	}
}

func (ssc *skillSubscriptionCache) Name() string {
	return cacheapi.SkillSubscriptionName
}

func (ssc *skillSubscriptionCache) Initialize(_ map[string]any) error {
	return nil
}

func (ssc *skillSubscriptionCache) Update() error {
	_, err, _ := ssc.singleFlight.Do(ssc.Name(), func() (any, error) {
		return nil, ssc.DoCacheUpdate(ssc.Name(), ssc.realUpdate)
	})
	return err
}

func (ssc *skillSubscriptionCache) realUpdate() (map[string]time.Time, int64, error) {
	results := make(map[string]time.Time)

	// SkillSubscriptionStore 目前没有增量更新接口，使用全量获取所有订阅
	// TODO: 未来存储层增加增量更新接口后可改为增量更新
	count, subscriptions, err := ssc.storage.QuerySkillSubscriptions(nil, 0, 0xFFFFFFFF)
	if err != nil {
		return nil, 0, err
	}

	upsert := 0
	del := 0

	for _, subscription := range subscriptions {
		if subscription.MTime.After(ssc.LastFetchTime()) {
			ssc.LastFetchTime().Add(time.Nanosecond)
		}

		key := skillSubscriptionKey(subscription)
		results[key] = subscription.MTime

		if !subscription.Active { // Active=false 表示已删除
			ssc.removeSubscription(subscription)
			del++
		} else {
			ssc.storeSubscription(subscription)
			upsert++
		}
	}

	log.Infof("[Cache][SkillSubscription] update subscriptions, upsert: %d, delete: %d, total: %d",
		upsert, del, len(subscriptions))
	return results, int64(count), nil
}

func (ssc *skillSubscriptionCache) Clear() error {
	ssc.BaseCache.Clear()
	// 重新初始化 SyncMap
	ssc.ids = container.NewSyncMap[string, *ai.SkillSubscription]()
	ssc.clientSubscriptions = container.NewSyncMap[string, []*ai.SkillSubscription]()
	ssc.skillSubscriptions = container.NewSyncMap[string, []*ai.SkillSubscription]()
	return nil
}

func (ssc *skillSubscriptionCache) Close() error {
	return nil
}

// GetSubscriptionByID 实现 SkillSubscriptionCache 接口
func (ssc *skillSubscriptionCache) GetSubscriptionByID(id string) *ai.SkillSubscription {
	if id == "" {
		return nil
	}
	val, ok := ssc.ids.Load(id)
	if !ok {
		return nil
	}
	return val
}

// GetSubscriptionsByClient 实现 SkillSubscriptionCache 接口
func (ssc *skillSubscriptionCache) GetSubscriptionsByClient(clientID string) []*ai.SkillSubscription {
	if clientID == "" {
		return nil
	}
	val, ok := ssc.clientSubscriptions.Load(clientID)
	if !ok {
		return nil
	}
	return val
}

// GetSubscriptionsBySkill 实现 SkillSubscriptionCache 接口
func (ssc *skillSubscriptionCache) GetSubscriptionsBySkill(skillName, namespace string) []*ai.SkillSubscription {
	if skillName == "" || namespace == "" {
		return nil
	}
	key := skillSubscriptionKeyBySkill(namespace, skillName)
	val, ok := ssc.skillSubscriptions.Load(key)
	if !ok {
		return nil
	}
	return val
}

// 辅助方法
func skillSubscriptionKey(subscription *ai.SkillSubscription) string {
	return subscription.ID
}

func skillSubscriptionKeyBySkill(namespace, skillName string) string {
	return namespace + "/" + skillName
}

func (ssc *skillSubscriptionCache) storeSubscription(subscription *ai.SkillSubscription) {
	// 存储 ID 索引
	ssc.ids.Store(subscription.ID, subscription)

	// 更新客户端订阅索引
	ssc.updateClientSubscriptions(subscription)

	// 更新技能订阅索引
	ssc.updateSkillSubscriptions(subscription)
}

func (ssc *skillSubscriptionCache) removeSubscription(subscription *ai.SkillSubscription) {
	// 删除 ID 索引
	ssc.ids.Delete(subscription.ID)

	// 从客户端订阅索引中移除
	ssc.removeFromClientSubscriptions(subscription)

	// 从技能订阅索引中移除
	ssc.removeFromSkillSubscriptions(subscription)
}

func (ssc *skillSubscriptionCache) updateClientSubscriptions(subscription *ai.SkillSubscription) {
	clientID := subscription.ClientID
	subscriptions, _ := ssc.clientSubscriptions.ComputeIfAbsent(clientID, func(k string) []*ai.SkillSubscription {
		return make([]*ai.SkillSubscription, 0)
	})
	// 先移除已存在的
	for i, s := range subscriptions {
		if s.ID == subscription.ID {
			subscriptions = append(subscriptions[:i], subscriptions[i+1:]...)
			break
		}
	}
	subscriptions = append(subscriptions, subscription)
	ssc.clientSubscriptions.Store(clientID, subscriptions)
}

func (ssc *skillSubscriptionCache) removeFromClientSubscriptions(subscription *ai.SkillSubscription) {
	clientID := subscription.ClientID
	subscriptions, ok := ssc.clientSubscriptions.Load(clientID)
	if !ok {
		return
	}
	// 移除
	for i, s := range subscriptions {
		if s.ID == subscription.ID {
			subscriptions = append(subscriptions[:i], subscriptions[i+1:]...)
			break
		}
	}
	if len(subscriptions) == 0 {
		ssc.clientSubscriptions.Store(clientID, []*ai.SkillSubscription{})
	} else {
		ssc.clientSubscriptions.Store(clientID, subscriptions)
	}
}

func (ssc *skillSubscriptionCache) updateSkillSubscriptions(subscription *ai.SkillSubscription) {
	key := skillSubscriptionKeyBySkill(subscription.Namespace, subscription.SkillName)
	subscriptions, _ := ssc.skillSubscriptions.ComputeIfAbsent(key, func(k string) []*ai.SkillSubscription {
		return make([]*ai.SkillSubscription, 0)
	})
	// 先移除已存在的
	for i, s := range subscriptions {
		if s.ID == subscription.ID {
			subscriptions = append(subscriptions[:i], subscriptions[i+1:]...)
			break
		}
	}
	subscriptions = append(subscriptions, subscription)
	ssc.skillSubscriptions.Store(key, subscriptions)
}

func (ssc *skillSubscriptionCache) removeFromSkillSubscriptions(subscription *ai.SkillSubscription) {
	key := skillSubscriptionKeyBySkill(subscription.Namespace, subscription.SkillName)
	subscriptions, ok := ssc.skillSubscriptions.Load(key)
	if !ok {
		return
	}
	// 移除
	for i, s := range subscriptions {
		if s.ID == subscription.ID {
			subscriptions = append(subscriptions[:i], subscriptions[i+1:]...)
			break
		}
	}
	if len(subscriptions) == 0 {
		ssc.skillSubscriptions.Store(key, []*ai.SkillSubscription{})
	} else {
		ssc.skillSubscriptions.Store(key, subscriptions)
	}
}
