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

package auth

import (
	"context"
	"sort"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// NewRoleCache
func NewRoleCache(storage store.Store, cacheMgr cacheapi.CacheManager) cacheapi.RoleCache {
	return &roleCache{
		BaseCache:    cachebase.NewBaseCache(storage, cacheMgr),
		singleFlight: new(singleflight.Group),
	}
}

type roleCache struct {
	*cachebase.BaseCache
	// roles
	roles *container.SyncMap[string, *authtypes.Role]
	// principalRoles
	principalRoles map[authtypes.PrincipalType]*container.SyncMap[string, *container.SyncSet[string]]
	singleFlight   *singleflight.Group
}

// Initialize implements api.RoleCache.
func (r *roleCache) Initialize(c map[string]interface{}) error {
	r.roles = container.NewSyncMap[string, *authtypes.Role]()
	r.principalRoles = map[authtypes.PrincipalType]*container.SyncMap[string, *container.SyncSet[string]]{
		authtypes.PrincipalUser:  container.NewSyncMap[string, *container.SyncSet[string]](),
		authtypes.PrincipalGroup: container.NewSyncMap[string, *container.SyncSet[string]](),
	}
	return nil
}

// Name implements api.RoleCache.
func (r *roleCache) Name() string {
	return cacheapi.RolesName
}

// Clear implements api.RoleCache.
// Subtle: this method shadows the method (*BaseCache).Clear of roleCache.BaseCache.
func (r *roleCache) Clear() error {
	r.roles = container.NewSyncMap[string, *authtypes.Role]()
	r.principalRoles = map[authtypes.PrincipalType]*container.SyncMap[string, *container.SyncSet[string]]{
		authtypes.PrincipalUser:  container.NewSyncMap[string, *container.SyncSet[string]](),
		authtypes.PrincipalGroup: container.NewSyncMap[string, *container.SyncSet[string]](),
	}
	return nil
}

// Update implements api.RoleCache.
func (r *roleCache) Update() error {
	// 多个线程竞争，只有一个线程进行更新
	_, err, _ := r.singleFlight.Do(r.Name(), func() (interface{}, error) {
		return nil, r.DoCacheUpdate(r.Name(), r.realUpdate)
	})
	return err
}

func (r *roleCache) realUpdate() (map[string]time.Time, int64, error) {
	// 获取几秒前的全部数据
	var (
		start      = time.Now()
		lastTime   = r.LastFetchTime()
		roles, err = r.BaseCache.Store().GetMoreRoles(r.IsFirstUpdate(), lastTime)
	)
	if err != nil {
		log.Errorf("[Cache][Roles] refresh auth roles cache err: %s", err.Error())
		return nil, -1, err
	}

	lastMtime, add, update, del := r.setRoles(roles)
	log.Info("[Cache][Roles] get more auth role",
		zap.Int("add", add), zap.Int("update", update), zap.Int("delete", del),
		zap.Time("last", lastTime), zap.Duration("used", time.Since(start)))
	return map[string]time.Time{
		r.Name(): lastMtime,
	}, int64(len(roles)), nil
}

func (r *roleCache) setRoles(roles []*authtypes.Role) (time.Time, int, int, int) {
	var add, remove, update int
	lastMtime := r.LastMtime(r.Name()).Unix()

	for i := range roles {
		item := roles[i]
		oldVal, exist := r.roles.Load(item.ID)
		r.dealPrincipalRoles(oldVal, true)
		if !item.Valid {
			remove++
			r.roles.Delete(item.ID)
		} else {
			if exist {
				update++
			} else {
				add++
			}
			r.dealPrincipalRoles(item, false)
			r.roles.Store(item.ID, item)
		}
	}
	r.cleanEmptyPrincipalRoles()
	return time.Unix(lastMtime, 0), add, update, remove
}

func (r *roleCache) cleanEmptyPrincipalRoles() {
	// 清理掉 principal 没有关联任何 role 的容器
	for pt := range r.principalRoles {
		r.principalRoles[pt].Range(func(key string, val *container.SyncSet[string]) {
			if val.Len() == 0 {
				r.principalRoles[pt].Delete(key)
			}
		})
	}
}

// dealPrincipalRoles 处理 principal 和 role 的关联关系
func (r *roleCache) dealPrincipalRoles(role *authtypes.Role, isDel bool) {
	if role == nil {
		return
	}
	if isDel {
		users := role.Users
		for i := range users {
			container, _ := r.principalRoles[authtypes.PrincipalUser].ComputeIfAbsent(users[i].PrincipalID,
				func(k string) *container.SyncSet[string] {
					return container.NewSyncSet[string]()
				})
			container.Remove(role.ID)
		}
		groups := role.UserGroups
		for i := range groups {
			container, _ := r.principalRoles[authtypes.PrincipalGroup].ComputeIfAbsent(groups[i].PrincipalID,
				func(k string) *container.SyncSet[string] {
					return container.NewSyncSet[string]()
				})
			container.Remove(role.ID)
		}
		return
	}
	users := role.Users
	for i := range users {
		container, _ := r.principalRoles[authtypes.PrincipalUser].ComputeIfAbsent(users[i].PrincipalID,
			func(k string) *container.SyncSet[string] {
				return container.NewSyncSet[string]()
			})
		container.Add(role.ID)
	}
	groups := role.UserGroups
	for i := range groups {
		container, _ := r.principalRoles[authtypes.PrincipalGroup].ComputeIfAbsent(groups[i].PrincipalID,
			func(k string) *container.SyncSet[string] {
				return container.NewSyncSet[string]()
			})
		container.Add(role.ID)
	}
}

// Query implements api.RoleCache.
func (r *roleCache) Query(ctx context.Context, args cacheapi.RoleSearchArgs) (uint32, []*authtypes.Role, error) {
	if err := r.Update(); err != nil {
		return 0, nil, err
	}
	var (
		total uint32
		roles []*authtypes.Role
	)

	searchId, hasId := args.Filters["id"]
	searchName, hasName := args.Filters["name"]
	searchSource, hasSource := args.Filters["source"]

	predicates := cacheapi.LoadAuthRolePredicates(ctx)

	r.roles.Range(func(key string, val *authtypes.Role) {
		if hasId && searchId != "" && key != searchId {
			return
		}
		if hasName && searchName != "" {
			if !utils.IsWildMatch(val.Name, searchName) {
				return
			}
		}
		if hasSource && searchSource != "" {
			if !utils.IsWildMatch(val.Source, searchSource) {
				return
			}
		}
		for i := range predicates {
			if !predicates[i](ctx, val) {
				return
			}
		}
		total++
		roles = append(roles, val)
	})

	sort.Slice(roles, func(i int, j int) bool {
		return roles[i].ModifyTime.After(roles[j].ModifyTime)
	})

	total, roles = r.toPage(total, roles, args)
	return total, roles, nil
}

func (r *roleCache) toPage(total uint32, roles []*authtypes.Role, args cacheapi.RoleSearchArgs) (uint32, []*authtypes.Role) {
	if args.Limit == 0 {
		return total, roles
	}
	if args.Offset >= total || args.Limit == 0 {
		return total, nil
	}
	endIdx := args.Offset + args.Limit
	if endIdx > total {
		endIdx = total
	}
	return total, roles[args.Offset:endIdx]
}

// GetPrincipalRoles implements api.RoleCache.
func (r *roleCache) GetPrincipalRoles(p authtypes.Principal) []*authtypes.Role {
	roleContainers, ok := r.principalRoles[p.PrincipalType]
	if !ok {
		return nil
	}
	containers, ok := roleContainers.Load(p.PrincipalID)
	if !ok {
		return nil
	}

	result := make([]*authtypes.Role, 0, containers.Len())
	containers.Range(func(val string) {
		role, ok := r.roles.Load(val)
		if !ok {
			return
		}
		result = append(result, role)
	})
	return result
}

// GetRole implements api.RoleCache.
func (r *roleCache) GetRole(id string) *authtypes.Role {
	ret, _ := r.roles.Load(id)
	return ret
}
