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

type skillCache struct {
	*cachebase.BaseCache
	storage store.Store

	// 数据存储
	// id -> *ai.Skill
	ids *container.SyncMap[string, *ai.Skill]

	// namespace/name -> *ai.Skill
	names *container.SyncMap[string, *ai.Skill]

	// namespace -> []*ai.Skill
	namespaceIndex *container.SyncMap[string, []*ai.Skill]

	// skillType -> []*ai.Skill (作为 category 索引)
	categoryIndex *container.SyncMap[string, []*ai.Skill]

	singleFlight *singleflight.Group
}

func NewSkillCache(storage store.Store, cacheMgr cacheapi.CacheManager) cacheapi.Cache {
	return &skillCache{
		BaseCache:       cachebase.NewBaseCache(storage, cacheMgr),
		storage:         storage,
		ids:             container.NewSyncMap[string, *ai.Skill](),
		names:           container.NewSyncMap[string, *ai.Skill](),
		namespaceIndex:  container.NewSyncMap[string, []*ai.Skill](),
		categoryIndex:   container.NewSyncMap[string, []*ai.Skill](),
		singleFlight:    &singleflight.Group{},
	}
}

func (sc *skillCache) Name() string {
	return cacheapi.SkillName
}

func (sc *skillCache) Initialize(_ map[string]any) error {
	return nil
}

func (sc *skillCache) Update() error {
	_, err, _ := sc.singleFlight.Do(sc.Name(), func() (any, error) {
		return nil, sc.DoCacheUpdate(sc.Name(), sc.realUpdate)
	})
	return err
}

func (sc *skillCache) realUpdate() (map[string]time.Time, int64, error) {
	results := make(map[string]time.Time)

	skills, err := sc.storage.GetMoreSkills(sc.LastFetchTime(), sc.IsFirstUpdate())
	if err != nil {
		return nil, 0, err
	}

	upsert := 0
	del := 0

	for _, skill := range skills {
		if skill.MTime.After(sc.LastFetchTime()) {
			sc.LastFetchTime().Add(time.Nanosecond)
		}

		key := skillKey(skill)
		results[key] = skill.MTime

		if skill.Flag == 1 { // Flag=1 表示已删除
			sc.removeSkill(skill)
			del++
		} else {
			sc.storeSkill(skill)
			upsert++
		}
	}

	log.Infof("[Cache][Skill] update skills, upsert: %d, delete: %d, total: %d",
		upsert, del, len(skills))
	return results, int64(len(skills)), nil
}

func (sc *skillCache) Clear() error {
	sc.BaseCache.Clear()
	// 重新初始化 SyncMap
	sc.ids = container.NewSyncMap[string, *ai.Skill]()
	sc.names = container.NewSyncMap[string, *ai.Skill]()
	sc.namespaceIndex = container.NewSyncMap[string, []*ai.Skill]()
	sc.categoryIndex = container.NewSyncMap[string, []*ai.Skill]()
	return nil
}

func (sc *skillCache) Close() error {
	return nil
}

// GetSkillByID 实现 SkillCache 接口
func (sc *skillCache) GetSkillByID(id string) *ai.Skill {
	if id == "" {
		return nil
	}
	val, ok := sc.ids.Load(id)
	if !ok {
		return nil
	}
	return val
}

// GetSkillByName 实现 SkillCache 接口
func (sc *skillCache) GetSkillByName(name, namespace string) *ai.Skill {
	if name == "" || namespace == "" {
		return nil
	}
	key := skillNameKey(namespace, name)
	val, ok := sc.names.Load(key)
	if !ok {
		return nil
	}
	return val
}

// GetSkillsByNamespace 实现 SkillCache 接口
func (sc *skillCache) GetSkillsByNamespace(namespace string) []*ai.Skill {
	if namespace == "" {
		return nil
	}
	val, ok := sc.namespaceIndex.Load(namespace)
	if !ok {
		return nil
	}
	return val
}

// GetSkillsByCategory 实现 SkillCache 接口
func (sc *skillCache) GetSkillsByCategory(category string) []*ai.Skill {
	if category == "" {
		return nil
	}
	val, ok := sc.categoryIndex.Load(category)
	if !ok {
		return nil
	}
	return val
}

// 辅助方法
func skillKey(skill *ai.Skill) string {
	return skill.ID
}

func skillNameKey(namespace, name string) string {
	return namespace + "/" + name
}

func (sc *skillCache) storeSkill(skill *ai.Skill) {
	// 存储 ID 索引
	sc.ids.Store(skill.ID, skill)

	// 存储名称索引
	key := skillNameKey(skill.Namespace, skill.Name)
	sc.names.Store(key, skill)

	// 更新命名空间索引
	sc.updateNamespaceIndex(skill)

	// 更新分类索引
	sc.updateCategoryIndex(skill)
}

func (sc *skillCache) removeSkill(skill *ai.Skill) {
	// 删除 ID 索引
	sc.ids.Delete(skill.ID)

	// 删除名称索引
	key := skillNameKey(skill.Namespace, skill.Name)
	sc.names.Delete(key)

	// 从命名空间索引中移除
	sc.removeFromNamespaceIndex(skill)

	// 从分类索引中移除
	sc.removeFromCategoryIndex(skill)
}

func (sc *skillCache) updateNamespaceIndex(skill *ai.Skill) {
	namespace := skill.Namespace
	skills, _ := sc.namespaceIndex.ComputeIfAbsent(namespace, func(k string) []*ai.Skill {
		return make([]*ai.Skill, 0)
	})
	// 先移除已存在的
	for i, s := range skills {
		if s.ID == skill.ID {
			skills = append(skills[:i], skills[i+1:]...)
			break
		}
	}
	skills = append(skills, skill)
	sc.namespaceIndex.Store(namespace, skills)
}

func (sc *skillCache) updateCategoryIndex(skill *ai.Skill) {
	category := skill.SkillType
	skills, _ := sc.categoryIndex.ComputeIfAbsent(category, func(k string) []*ai.Skill {
		return make([]*ai.Skill, 0)
	})
	// 先移除已存在的
	for i, s := range skills {
		if s.ID == skill.ID {
			skills = append(skills[:i], skills[i+1:]...)
			break
		}
	}
	skills = append(skills, skill)
	sc.categoryIndex.Store(category, skills)
}

func (sc *skillCache) removeFromNamespaceIndex(skill *ai.Skill) {
	namespace := skill.Namespace
	skills, ok := sc.namespaceIndex.Load(namespace)
	if !ok {
		return
	}
	// 移除
	for i, s := range skills {
		if s.ID == skill.ID {
			skills = append(skills[:i], skills[i+1:]...)
			break
		}
	}
	if len(skills) == 0 {
		sc.namespaceIndex.Store(namespace, []*ai.Skill{})
	} else {
		sc.namespaceIndex.Store(namespace, skills)
	}
}

func (sc *skillCache) removeFromCategoryIndex(skill *ai.Skill) {
	category := skill.SkillType
	skills, ok := sc.categoryIndex.Load(category)
	if !ok {
		return
	}
	// 移除
	for i, s := range skills {
		if s.ID == skill.ID {
			skills = append(skills[:i], skills[i+1:]...)
			break
		}
	}
	if len(skills) == 0 {
		sc.categoryIndex.Store(category, []*ai.Skill{})
	} else {
		sc.categoryIndex.Store(category, skills)
	}
}
