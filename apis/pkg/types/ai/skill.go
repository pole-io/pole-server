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
	"fmt"
	"time"
)

// Skill 表示一个可复用的技能单元
type Skill struct {
	ID          string    `json:"id"`            // Skill ID
	Name        string    `json:"name"`          // Skill 名称，在 namespace 唯一
	Namespace   string    `json:"namespace"`     // 所属命名空间
	Description string    `json:"description"`   // 描述信息
	InputSchema string    `json:"input_schema"`  // 输入 JSON Schema
	OutputSchema string   `json:"output_schema"` // 输出 JSON Schema
	// Skill 类型: "function", "tool", "agent" 等
	SkillType string            `json:"skill_type"`
	// 作者信息
	Author      string            `json:"author"`
	Business    string            `json:"business"`      // 业务线
	Department  string            `json:"department"`    // 部门
	Metadata    map[string]string `json:"metadata"`      // 扩展元数据
	Flag        int8              `json:"flag"`          // 逻辑删除标志, 0 可见, 1 已删除
	Reference   string            `json:"reference"`     // 引用的资源名称
	Protocol    string            `json:"protocol"`      // 协议类型
	Revision    string            `json:"revision"`      // 版本号
	ExportTo    string            `json:"export_to"`     // 导出到的命名空间
	CTime       time.Time         `json:"ctime"`         // 创建时间
	MTime       time.Time         `json:"mtime"`         // 修改时间
}

// Key 获取 Skill 的唯一键
func (s *Skill) Key() string {
	return s.Namespace + "/" + s.Name
}

// KeyWithID 获取包含 ID 的唯一键
func (s *Skill) KeyWithID() string {
	return s.ID
}

// TableName 设置数据库表名
func (Skill) TableName() string {
	return "skill"
}

// SkillGroup 表示 Skill 的分组
type SkillGroup struct {
	ID          string            `json:"id"`            // SkillGroup ID
	Name        string            `json:"name"`          // SkillGroup 名称，在 namespace 唯一
	Namespace   string            `json:"namespace"`     // 所属命名空间
	Comment     string            `json:"comment"`       // 备注说明
	Metadata    map[string]string `json:"metadata"`      // 扩展元数据
	Owner       string            `json:"owner"`         // 所有者
	Business    string            `json:"business"`      // 业务线
	Department  string            `json:"department"`    // 部门
	Flag        int8              `json:"flag"`          // 逻辑删除标志, 0 可见, 1 已删除
	CTime       time.Time         `json:"ctime"`         // 创建时间
	MTime       time.Time         `json:"mtime"`         // 修改时间
}

// Key 获取 SkillGroup 的唯一键
func (s *SkillGroup) Key() string {
	return s.Namespace + "/" + s.Name
}

// TableName 设置数据库表名
func (SkillGroup) TableName() string {
	return "skill_group"
}

// SkillVersion 表示 Skill 的版本信息
type SkillVersion struct {
	ID            string    `json:"id"`            // Version ID
	SkillID       string    `json:"skill_id"`      // 所属 Skill ID
	SkillName     string    `json:"skill_name"`    // Skill 名称
	Namespace     string    `json:"namespace"`     // 命名空间
	Version       uint64    `json:"version"`       // 版本号
	Comment       string    `json:"comment"`       // 版本备注
	InputSchema   string    `json:"input_schema"`  // 输入 JSON Schema
	OutputSchema  string    `json:"output_schema"` // 输出 JSON Schema
	SkillType     string    `json:"skill_type"`    // 技能类型
	Metadata      map[string]string `json:"metadata"`  // 元数据
	Active        bool      `json:"active"`        // 是否为当前版本
	Flag          int8      `json:"flag"`          // 逻辑删除标志
	CTime         time.Time `json:"ctime"`         // 创建时间
	MTime         time.Time `json:"mtime"`         // 修改时间
}

// Key 获取 SkillVersion 的唯一键
func (s *SkillVersion) Key() string {
	return s.Namespace + "/" + s.SkillName + "/" + fmt.Sprintf("%d", s.Version)
}

// TableName 设置数据库表名
func (SkillVersion) TableName() string {
	return "skill_version"
}

// SkillSubscription 表示客户端对 Skill 的订阅关系
type SkillSubscription struct {
	ID           string    `json:"id"`
	SkillID      string    `json:"skill_id"`
	SkillName    string    `json:"skill_name"`
	Namespace    string    `json:"namespace"`
	ClientID     string    `json:"client_id"`
	ClientHost   string    `json:"client_host"`
	ClientType   string    `json:"client_type"`   // 客户端类型: "go", "java", "python" 等
	Version      uint64    `json:"version"`       // 订阅的版本
	Active       bool      `json:"active"`        // 是否活跃订阅
	CTime        time.Time `json:"ctime"`         // 创建时间
	MTime        time.Time `json:"mtime"`         // 修改时间
}

// Key 获取订阅关系的唯一键
func (s *SkillSubscription) Key() string {
	return s.Namespace + "/" + s.SkillName + "/" + s.ClientID
}

// TableName 设置数据库表名
func (SkillSubscription) TableName() string {
	return "skill_subscription"
}
