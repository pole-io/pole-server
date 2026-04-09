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

type skillVersionCache struct {
	*cachebase.BaseCache
	storage store.Store

	// id -> *ai.SkillVersion
	ids *container.SyncMap[string, *ai.SkillVersion]

	// namespace/skillName/version -> *ai.SkillVersion
	versionKeys *container.SyncMap[string, *ai.SkillVersion]

	// namespace/skillName -> *ai.SkillVersion (active version)
	activeVersions *container.SyncMap[string, *ai.SkillVersion]

	// namespace/skillName -> []*ai.SkillVersion
	skillVersions *container.SyncMap[string, []*ai.SkillVersion]

	singleFlight *singleflight.Group
}

func NewSkillVersionCache(storage store.Store, cacheMgr cacheapi.CacheManager) cacheapi.Cache {
	return &skillVersionCache{
		BaseCache:       cachebase.NewBaseCache(storage, cacheMgr),
		storage:         storage,
		ids:             container.NewSyncMap[string, *ai.SkillVersion](),
		versionKeys:     container.NewSyncMap[string, *ai.SkillVersion](),
		activeVersions:  container.NewSyncMap[string, *ai.SkillVersion](),
		skillVersions:   container.NewSyncMap[string, []*ai.SkillVersion](),
		singleFlight:    &singleflight.Group{},
	}
}

func (svc *skillVersionCache) Name() string {
	return cacheapi.SkillVersionName
}

func (svc *skillVersionCache) Initialize(_ map[string]any) error {
	return nil
}

func (svc *skillVersionCache) Update() error {
	_, err, _ := svc.singleFlight.Do(svc.Name(), func() (any, error) {
		return nil, svc.DoCacheUpdate(svc.Name(), svc.realUpdate)
	})
	return err
}

func (svc *skillVersionCache) realUpdate() (map[string]time.Time, int64, error) {
	results := make(map[string]time.Time)

	// SkillVersionStore 目前没有增量更新接口，使用全量获取所有版本
	// TODO: 未来存储层增加 GetMoreSkillVersions 方法后可改为增量更新
	skillVersionMap := make(map[string]*ai.SkillVersion)

	// 获取所有技能的所有版本
	skills, err := svc.storage.GetMoreSkills(svc.LastFetchTime(), svc.IsFirstUpdate())
	if err != nil {
		return nil, 0, err
	}

	upsert := 0
	del := 0

	for _, skill := range skills {
		if skill.Flag == 1 {
			continue
		}

		versions, err := svc.storage.GetSkillVersionsBySkillID(skill.ID)
		if err != nil {
			log.Warnf("[Cache][SkillVersion] get versions for skill %s err: %s", skill.ID, err.Error())
			continue
		}

		for _, version := range versions {
			skillVersionMap[version.ID] = version

			if version.MTime.After(svc.LastFetchTime()) {
				svc.LastFetchTime().Add(time.Nanosecond)
			}

			key := skillVersionKey(version)
			results[key] = version.MTime

			if version.Flag == 1 { // Flag=1 表示已删除
				svc.removeSkillVersion(version)
				del++
			} else {
				svc.storeSkillVersion(version)
				upsert++
			}
		}
	}

	log.Infof("[Cache][SkillVersion] update versions, upsert: %d, delete: %d, total: %d",
		upsert, del, len(skillVersionMap))
	return results, int64(len(skillVersionMap)), nil
}

func (svc *skillVersionCache) Clear() error {
	svc.BaseCache.Clear()
	// 重新初始化 SyncMap
	svc.ids = container.NewSyncMap[string, *ai.SkillVersion]()
	svc.versionKeys = container.NewSyncMap[string, *ai.SkillVersion]()
	svc.activeVersions = container.NewSyncMap[string, *ai.SkillVersion]()
	svc.skillVersions = container.NewSyncMap[string, []*ai.SkillVersion]()
	return nil
}

func (svc *skillVersionCache) Close() error {
	return nil
}

// GetSkillVersionByID 实现 SkillVersionCache 接口
func (svc *skillVersionCache) GetSkillVersionByID(id string) *ai.SkillVersion {
	if id == "" {
		return nil
	}
	val, ok := svc.ids.Load(id)
	if !ok {
		return nil
	}
	return val
}

// GetSkillVersionByVersion 实现 SkillVersionCache 接口
func (svc *skillVersionCache) GetSkillVersionByVersion(skillName, namespace string, version uint64) *ai.SkillVersion {
	if skillName == "" || namespace == "" {
		return nil
	}
	key := skillVersionKeyBySkill(namespace, skillName, version)
	val, ok := svc.versionKeys.Load(key)
	if !ok {
		return nil
	}
	return val
}

// GetActiveSkillVersion 实现 SkillVersionCache 接口
func (svc *skillVersionCache) GetActiveSkillVersion(skillName, namespace string) *ai.SkillVersion {
	if skillName == "" || namespace == "" {
		return nil
	}
	key := skillVersionKeyBySkill(namespace, skillName, 0)
	val, ok := svc.activeVersions.Load(key)
	if !ok {
		return nil
	}
	return val
}

// GetSkillVersions 实现 SkillVersionCache 接口
func (svc *skillVersionCache) GetSkillVersions(skillName, namespace string) []*ai.SkillVersion {
	if skillName == "" || namespace == "" {
		return nil
	}
	key := skillVersionKeyBySkill(namespace, skillName, 0)
	val, ok := svc.skillVersions.Load(key)
	if !ok {
		return nil
	}
	return val
}

// 辅助方法
func skillVersionKey(version *ai.SkillVersion) string {
	return version.ID
}

func skillVersionKeyBySkill(namespace, skillName string, version uint64) string {
	return namespace + "/" + skillName
}

func (svc *skillVersionCache) storeSkillVersion(version *ai.SkillVersion) {
	// 存储 ID 索引
	svc.ids.Store(version.ID, version)

	// 存储版本键索引
	key := skillVersionKey(version)
	svc.versionKeys.Store(key, version)

	// 更新技能版本列表
	svc.updateSkillVersions(version)

	// 更新活跃版本
	if version.Active {
		svc.updateActiveVersion(version)
	}
}

func (svc *skillVersionCache) removeSkillVersion(version *ai.SkillVersion) {
	// 删除 ID 索引
	svc.ids.Delete(version.ID)

	// 删除版本键索引
	key := skillVersionKey(version)
	svc.versionKeys.Delete(key)

	// 从技能版本列表中移除
	svc.removeFromSkillVersions(version)

	// 从活跃版本中移除
	if version.Active {
		svc.removeActiveVersion(version)
	}
}

func (svc *skillVersionCache) updateSkillVersions(version *ai.SkillVersion) {
	key := skillVersionKeyBySkill(version.Namespace, version.SkillName, 0)
	versions, _ := svc.skillVersions.ComputeIfAbsent(key, func(k string) []*ai.SkillVersion {
		return make([]*ai.SkillVersion, 0)
	})
	// 先移除已存在的
	for i, v := range versions {
		if v.ID == version.ID {
			versions = append(versions[:i], versions[i+1:]...)
			break
		}
	}
	versions = append(versions, version)
	svc.skillVersions.Store(key, versions)
}

func (svc *skillVersionCache) removeFromSkillVersions(version *ai.SkillVersion) {
	key := skillVersionKeyBySkill(version.Namespace, version.SkillName, 0)
	versions, ok := svc.skillVersions.Load(key)
	if !ok {
		return
	}
	// 移除
	for i, v := range versions {
		if v.ID == version.ID {
			versions = append(versions[:i], versions[i+1:]...)
			break
		}
	}
	if len(versions) == 0 {
		svc.skillVersions.Store(key, []*ai.SkillVersion{})
	} else {
		svc.skillVersions.Store(key, versions)
	}
}

func (svc *skillVersionCache) updateActiveVersion(version *ai.SkillVersion) {
	key := skillVersionKeyBySkill(version.Namespace, version.SkillName, 0)
	if version.Active {
		svc.activeVersions.Store(key, version)
	}
}

func (svc *skillVersionCache) removeActiveVersion(version *ai.SkillVersion) {
	key := skillVersionKeyBySkill(version.Namespace, version.SkillName, 0)
	svc.activeVersions.Delete(key)
}
