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

package cache

import (
	"github.com/pole-io/pole-server/apis/pkg/types/ai"
)

// SkillCache Skill 缓存接口
type SkillCache interface {
	Cache

	// GetSkillByID 根据 ID 获取 Skill
	GetSkillByID(id string) *ai.Skill

	// GetSkillByName 根据名称和命名空间获取 Skill
	GetSkillByName(name, namespace string) *ai.Skill

	// GetSkillsByNamespace 获取命名空间下的所有 Skill
	GetSkillsByNamespace(namespace string) []*ai.Skill

	// GetSkillsByCategory 获取指定类型的 Skill (使用 SkillType)
	GetSkillsByCategory(category string) []*ai.Skill
}

// MCPServerCache MCP Server 缓存接口
type MCPServerCache interface {
	Cache

	// GetMCPServerByID 根据 ID 获取 MCP Server
	GetMCPServerByID(id string) *ai.MCPServer

	// GetMCPServerByName 根据名称和命名空间获取 MCP Server
	GetMCPServerByName(name, namespace string) *ai.MCPServer

	// GetMCPServersByNamespace 获取命名空间下的所有 MCP Server
	GetMCPServersByNamespace(namespace string) []*ai.MCPServer

	// GetMCPServerTools 获取 MCP Server 的所有工具
	GetMCPServerTools(serverID string) []*ai.MCPServerTool
}

// SkillVersionCache Skill 版本缓存接口
type SkillVersionCache interface {
	Cache

	// GetSkillVersionByID 根据 ID 获取 SkillVersion
	GetSkillVersionByID(id string) *ai.SkillVersion

	// GetSkillVersionByVersion 根据技能名、命名空间和版本号获取
	GetSkillVersionByVersion(skillName, namespace string, version uint64) *ai.SkillVersion

	// GetActiveSkillVersion 获取活跃版本
	GetActiveSkillVersion(skillName, namespace string) *ai.SkillVersion

	// GetSkillVersions 获取技能的所有版本
	GetSkillVersions(skillName, namespace string) []*ai.SkillVersion
}

// SkillSubscriptionCache Skill 订阅缓存接口
type SkillSubscriptionCache interface {
	Cache

	// GetSubscriptionByID 根据 ID 获取订阅
	GetSubscriptionByID(id string) *ai.SkillSubscription

	// GetSubscriptionsByClient 获取客户端的所有订阅
	GetSubscriptionsByClient(clientID string) []*ai.SkillSubscription

	// GetSubscriptionsBySkill 获取技能的所有订阅
	GetSubscriptionsBySkill(skillName, namespace string) []*ai.SkillSubscription
}
