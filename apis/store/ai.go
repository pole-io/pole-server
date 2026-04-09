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

	"github.com/pole-io/pole-server/apis/pkg/types/ai"
)

// DiscoverStore Service discovery storage interface
type AIStore interface {
	MCPServerStore
	SkillStore
	SkillGroupStore
	SkillVersionStore
	SkillSubscriptionStore
}

// ===== MCP Server Store =====

type MCPServerStore interface {
	// CreateMCPServer 创建 MCP Server
	CreateMCPServer(server *ai.MCPServer) error
	// UpdateMCPServer 更新 MCP Server
	UpdateMCPServer(server *ai.MCPServer) error
	// DeleteMCPServer 删除 MCP Server
	DeleteMCPServer(id string) error
	// GetMCPServer 获取 MCP Server
	GetMCPServer(id string) (*ai.MCPServer, error)
	// GetMCPServerByName 获取 MCP Server by name and namespace
	GetMCPServerByName(name, namespace string) (*ai.MCPServer, error)
	// GetMoreMCPServers 增量获取 MCP Servers
	GetMoreMCPServers(mtime time.Time, firstUpdate bool) ([]*ai.MCPServer, error)
	// HasMCPServer 检查 MCP Server 是否存在
	HasMCPServer(id string) (bool, error)
	// HasMCPServerByName 检查 MCP Server 是否存在 by name
	HasMCPServerByName(name, namespace string) (bool, error)
	// HasMCPServerByNameExcludeId 检查 MCP Server 是否存在 by name exclude id
	HasMCPServerByNameExcludeId(name, namespace, id string) (bool, error)

	// MCP Server Tool 相关方法
	// CreateMCPServerTool 创建 MCP Server Tool
	CreateMCPServerTool(tool *ai.MCPServerTool) error
	// UpdateMCPServerTool 更新 MCP Server Tool
	UpdateMCPServerTool(tool *ai.MCPServerTool) error
	// DeleteMCPServerTool 删除 MCP Server Tool
	DeleteMCPServerTool(id string) error
	// GetMCPServerTool 获取 MCP Server Tool
	GetMCPServerTool(id string) (*ai.MCPServerTool, error)
	// GetMCPServerToolsByServerID 获取 MCP Server 的所有 Tools
	GetMCPServerToolsByServerID(serverID string) ([]*ai.MCPServerTool, error)
	// GetMCPServerTools 增量获取 MCP Server Tools
	GetMCPServerTools(mtime time.Time, firstUpdate bool) ([]*ai.MCPServerTool, error)

	// QueryMCPServers 查询 MCP Servers（支持过滤和分页）
	QueryMCPServers(filter map[string]string, offset, limit uint32) (uint32, []*ai.MCPServer, error)
}

// ===== Skill Store =====

// SkillStore 技能存储接口
type SkillStore interface {
	// CreateSkill 创建技能
	CreateSkill(skill *ai.Skill) error
	// UpdateSkill 更新技能
	UpdateSkill(skill *ai.Skill) error
	// DeleteSkill 删除技能 (逻辑删除)
	DeleteSkill(id string) error
	// GetSkill 获取技能 by ID
	GetSkill(id string) (*ai.Skill, error)
	// GetSkillByName 获取技能 by name and namespace
	GetSkillByName(name, namespace string) (*ai.Skill, error)
	// GetMoreSkills 增量获取技能
	GetMoreSkills(mtime time.Time, firstUpdate bool) ([]*ai.Skill, error)
	// HasSkill 检查技能是否存在
	HasSkill(id string) (bool, error)
	// HasSkillByName 检查技能是否存在 by name
	HasSkillByName(name, namespace string) (bool, error)
	// HasSkillByNameExcludeId 检查技能是否存在 by name exclude id
	HasSkillByNameExcludeId(name, namespace, id string) (bool, error)
	// QuerySkills 分页查询技能（支持过滤）
	QuerySkills(filter map[string]string, offset, limit uint32) (uint32, []*ai.Skill, error)
}

// ===== Skill Group Store =====

// SkillGroupStore 技能分组存储接口
type SkillGroupStore interface {
	// CreateSkillGroup 创建技能分组
	CreateSkillGroup(group *ai.SkillGroup) (*ai.SkillGroup, error)
	// UpdateSkillGroup 更新技能分组
	UpdateSkillGroup(group *ai.SkillGroup) error
	// GetSkillGroup 获取技能分组
	GetSkillGroup(namespace, name string) (*ai.SkillGroup, error)
	// DeleteSkillGroup 删除技能分组
	DeleteSkillGroup(namespace, name string) error
	// GetMoreSkillGroups 增量获取技能分组
	GetMoreSkillGroups(firstUpdate bool, mtime time.Time) ([]*ai.SkillGroup, error)
	// CountSkillGroups 获取一个命名空间下的技能分组数量
	CountSkillGroups(namespace string) (uint64, error)
	// QuerySkillGroups 分页查询技能分组（支持过滤）
	QuerySkillGroups(filter map[string]string, offset, limit uint32) (uint32, []*ai.SkillGroup, error)
}

// ===== Skill Version Store =====

// SkillVersionStore 技能版本存储接口
type SkillVersionStore interface {
	// CreateSkillVersion 创建技能版本
	CreateSkillVersion(version *ai.SkillVersion) error
	// UpdateSkillVersion 更新技能版本
	UpdateSkillVersion(version *ai.SkillVersion) error
	// DeleteSkillVersion 删除技能版本
	DeleteSkillVersion(id string) error
	// GetSkillVersion 获取技能版本
	GetSkillVersion(id string) (*ai.SkillVersion, error)
	// GetSkillVersionByVersion 获取技能版本 by version
	GetSkillVersionByVersion(skillName, namespace string, version uint64) (*ai.SkillVersion, error)
	// GetSkillVersionsBySkillID 获取技能的所有版本
	GetSkillVersionsBySkillID(skillID string) ([]*ai.SkillVersion, error)
	// QuerySkillVersions 分页查询技能版本
	QuerySkillVersions(filter map[string]string, offset, limit uint32) (uint32, []*ai.SkillVersion, error)
	// GetActiveSkillVersion 获取技能的活跃版本
	GetActiveSkillVersion(skillName, namespace string) (*ai.SkillVersion, error)
	// ActiveSkillVersion 激活技能版本
	ActiveSkillVersion(version *ai.SkillVersion) error
	// InactiveSkillVersion 取消激活技能版本
	InactiveSkillVersion(version *ai.SkillVersion) error
}

// ===== Skill Subscription Store =====

// SkillSubscriptionStore 技能订阅存储接口
type SkillSubscriptionStore interface {
	// CreateSkillSubscription 创建技能订阅
	CreateSkillSubscription(sub *ai.SkillSubscription) error
	// UpdateSkillSubscription 更新技能订阅
	UpdateSkillSubscription(sub *ai.SkillSubscription) error
	// DeleteSkillSubscription 删除技能订阅
	DeleteSkillSubscription(id string) error
	// GetSkillSubscription 获取技能订阅
	GetSkillSubscription(id string) (*ai.SkillSubscription, error)
	// GetSkillSubscriptionByClient 获取客户端的技能订阅
	GetSkillSubscriptionByClient(clientID string) ([]*ai.SkillSubscription, error)
	// GetSkillSubscriptionsBySkill 获取技能的所有订阅
	GetSkillSubscriptionsBySkill(skillName, namespace string) ([]*ai.SkillSubscription, error)
	// QuerySkillSubscriptions 分页查询技能订阅
	QuerySkillSubscriptions(filter map[string]string, offset, limit uint32) (uint32, []*ai.SkillSubscription, error)
	// UpdateSubscriptionVersion 更新订阅的版本
	UpdateSubscriptionVersion(clientID, skillName, namespace string, version uint64) error
	// DeactiveSubscription 取消订阅
	DeactiveSubscription(clientID, skillName, namespace string) error
}

