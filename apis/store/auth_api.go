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

package store

import (
	"time"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
)

type AuthStore interface {
	// UserStore 用户接口
	UserStore
	// GroupStore 用户组接口
	GroupStore
	// StrategyStore 鉴权策略接口
	StrategyStore
	// RoleStore 角色接口
	RoleStore
}

// UserStore User-related operation interface
type UserStore interface {
	// GetMainUser Get the main account
	GetMainUser() (*authtypes.User, error)
	// AddUser Create a user
	AddUser(tx Tx, user *authtypes.User) error
	// UpdateUser Update user
	UpdateUser(user *authtypes.User) error
	// DeleteUser delete users
	DeleteUser(tx Tx, user *authtypes.User) error
	// GetSubCount Number of getting a child account
	GetSubCount(user *authtypes.User) (uint32, error)
	// GetUser Obtain user
	GetUser(id string) (*authtypes.User, error)
	// GetUserByName Get a unique user according to Name + Owner
	GetUserByName(name, ownerId string) (*authtypes.User, error)
	// GetUserByIDS Get users according to USER IDS batch
	GetUserByIds(ids []string) ([]*authtypes.User, error)
	// GetMoreUsers Used to refresh user cache
	// 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
	GetMoreUsers(mtime time.Time, firstUpdate bool) ([]*authtypes.User, error)
}

// GroupStore User group storage operation interface
type GroupStore interface {
	// AddGroup Add a user group
	AddGroup(tx Tx, group *authtypes.UserGroupDetail) error
	// UpdateGroup Update user group
	UpdateGroup(group *authtypes.ModifyUserGroup) error
	// DeleteGroup Delete user group
	DeleteGroup(tx Tx, group *authtypes.UserGroupDetail) error
	// GetGroup Get user group details
	GetGroup(id string) (*authtypes.UserGroupDetail, error)
	// GetGroupByName Get user groups according to Name and Owner
	GetGroupByName(name, owner string) (*authtypes.UserGroup, error)
	// GetMoreGroups Refresh of getting user groups for cache
	// 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
	GetMoreGroups(mtime time.Time, firstUpdate bool) ([]*authtypes.UserGroupDetail, error)
}

// StrategyStore Authentication policy related storage operation interface
type StrategyStore interface {
	// AddStrategy Create authentication strategy
	AddStrategy(tx Tx, strategy *authtypes.StrategyDetail) error
	// UpdateStrategy Update authentication strategy
	UpdateStrategy(strategy *authtypes.ModifyStrategyDetail) error
	// DeleteStrategy Delete authentication strategy
	DeleteStrategy(id string) error
	// CleanPrincipalPolicies Clean all the policies associated with the principal
	CleanPrincipalPolicies(tx Tx, p authtypes.Principal) error
	// LooseAddStrategyResources Song requires the resources of the authentication strategy,
	//   allowing the issue of ignoring the primary key conflict
	LooseAddStrategyResources(resources []authtypes.StrategyResource) error
	// RemoveStrategyResources Clean all the strategies associated with corresponding resources
	RemoveStrategyResources(resources []authtypes.StrategyResource) error
	// GetStrategyResources Gets a Principal's corresponding resource ID data information
	GetStrategyResources(principalId string, principalRole authtypes.PrincipalType) ([]authtypes.StrategyResource, error)
	// GetDefaultStrategyDetailByPrincipal Get a default policy for a Principal
	GetDefaultStrategyDetailByPrincipal(principalId string,
		principalType authtypes.PrincipalType) (*authtypes.StrategyDetail, error)
	// GetStrategyDetail Get strategy details
	GetStrategyDetail(id string) (*authtypes.StrategyDetail, error)
	// GetMoreStrategies Used to refresh policy cache
	// 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
	GetMoreStrategies(mtime time.Time, firstUpdate bool) ([]*authtypes.StrategyDetail, error)
}

// RoleStore Role related storage operation interface
type RoleStore interface {
	// GetRole
	GetRole(id string) (*authtypes.Role, error)
	// AddRole Add a role
	AddRole(role *authtypes.Role) error
	// UpdateRole Update a role
	UpdateRole(role *authtypes.Role) error
	// DeleteRole Delete a role
	DeleteRole(tx Tx, role *authtypes.Role) error
	// CleanPrincipalRoles Clean all the roles associated with the principal
	CleanPrincipalRoles(tx Tx, p *authtypes.Principal) error
	// GetRole get more role for cache update
	GetMoreRoles(firstUpdate bool, modifyTime time.Time) ([]*authtypes.Role, error)
}
