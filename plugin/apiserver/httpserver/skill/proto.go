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

package skill

import (
	"time"

	aiTypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
)

// Skill API request/response types

// Skill request for creating a skill
type Skill struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Namespace     string            `json:"namespace"`
	Description  string            `json:"description"`
	SkillType     string            `json:"skill_type"`
	InputSchema   string            `json:"input_schema"`
	OutputSchema  string            `json:"output_schema"`
	Author        string            `json:"author"`
	Business      string            `json:"business"`
	Department    string            `json:"department"`
	Metadata      map[string]string `json:"metadata"`
	Tags          []string          `json:"tags"`
	Revision     string            `json:"revision"`
	ExportTo      string            `json:"export_to"`
}

// SkillArr is an array of Skill
type SkillArr []*Skill

// ToAIType converts SkillArr to []*aiTypes.Skill
func (arr SkillArr) ToAIType() []*aiTypes.Skill {
	if arr == nil {
		return nil
	}
	result := make([]*aiTypes.Skill, 0, len(arr))
	for _, s := range arr {
		result = append(result, s.ToAIType())
	}
	return result
}

// SkillReference is a reference to a skill in a group
type SkillReference struct {
	SkillID   string `json:"skill_id"`
	SkillName string `json:"skill_name"`
	Version   uint64 `json:"version"` // 0 means use active version
}

// SkillGroup API request/response types

// SkillGroup request for creating a group
type SkillGroup struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Description string            `json:"description"`
	Skills      []*SkillReference `json:"skills"`
	Metadata    map[string]string `json:"metadata"`
	Owner       string            `json:"owner"`
	Business    string            `json:"business"`
	Department  string            `json:"department"`
}

// SkillGroupArr is an array of SkillGroup
type SkillGroupArr []*SkillGroup

// ToAIType converts SkillGroupArr to []*aiTypes.SkillGroup
func (arr SkillGroupArr) ToAIType() []*aiTypes.SkillGroup {
	if arr == nil {
		return nil
	}
	result := make([]*aiTypes.SkillGroup, 0, len(arr))
	for _, g := range arr {
		result = append(result, g.ToAIType())
	}
	return result
}

// Convert Skill to ai.Skill
func (s *Skill) ToAIType() *aiTypes.Skill {
	if s == nil {
		return nil
	}
	return &aiTypes.Skill{
		ID:           s.ID,
		Name:         s.Name,
		Namespace:    s.Namespace,
		Description:  s.Description,
		SkillType:    s.SkillType,
		InputSchema:  s.InputSchema,
		OutputSchema: s.OutputSchema,
		Author:       s.Author,
		Business:     s.Business,
		Department:   s.Department,
		Metadata:     s.Metadata,
		Flag:        0,
		Revision:    s.Revision,
		ExportTo:     s.ExportTo,
	}
}

// Convert SkillGroup to ai.SkillGroup
func (g *SkillGroup) ToAIType() *aiTypes.SkillGroup {
	if g == nil {
		return nil
	}
	return &aiTypes.SkillGroup{
		ID:          g.ID,
		Name:        g.Name,
		Namespace:   g.Namespace,
		Comment:     g.Description,
		Metadata:   g.Metadata,
		Owner:      g.Owner,
		Business:   g.Business,
		Department: g.Department,
		Flag:      0,
	}
}

// SkillVersion represents a version of a skill
type SkillVersion struct {
	ID           string            `json:"id"`
	SkillID      string            `json:"skill_id"`
	SkillName    string            `json:"skill_name"`
	Namespace    string            `json:"namespace"`
	Version      uint64            `json:"version"`
	Comment      string            `json:"comment"`
	InputSchema  string            `json:"input_schema"`
	OutputSchema string            `json:"output_schema"`
	SkillType    string            `json:"skill_type"`
	Metadata     map[string]string `json:"metadata"`
	Active       bool              `json:"active"`
	CreateTime   time.Time         `json:"create_time"`
	ModifyTime   time.Time         `json:"modify_time"`
}

// SkillVersionArr is an array of SkillVersion
type SkillVersionArr []*SkillVersion

// ToAIType converts SkillVersionArr to []*aiTypes.SkillVersion
func (arr SkillVersionArr) ToAIType() []*aiTypes.SkillVersion {
	if arr == nil {
		return nil
	}
	result := make([]*aiTypes.SkillVersion, 0, len(arr))
	for _, v := range arr {
		result = append(result, v.ToAIType())
	}
	return result
}

// Convert SkillVersion to ai.SkillVersion
func (v *SkillVersion) ToAIType() *aiTypes.SkillVersion {
	if v == nil {
		return nil
	}
	return &aiTypes.SkillVersion{
		ID:           v.ID,
		SkillID:      v.SkillID,
		SkillName:    v.SkillName,
		Namespace:    v.Namespace,
		Version:      v.Version,
		Comment:      v.Comment,
		InputSchema:  v.InputSchema,
		OutputSchema: v.OutputSchema,
		SkillType:    v.SkillType,
		Metadata:     v.Metadata,
		Active:       v.Active,
		CTime:        v.CreateTime,
		MTime:        v.ModifyTime,
	}
}

// SkillSubscription represents a subscription to a skill
type SkillSubscription struct {
	ID         string    `json:"id"`
	SkillID    string    `json:"skill_id"`
	SkillName  string    `json:"skill_name"`
	Namespace  string    `json:"namespace"`
	ClientID   string    `json:"client_id"`
	ClientHost string    `json:"client_host"`
	ClientType string    `json:"client_type"`
	Version    uint64    `json:"version"`
	Active     bool      `json:"active"`
	CreateTime time.Time `json:"create_time"`
	ModifyTime time.Time `json:"modify_time"`
}

// SkillSubscriptionArr is an array of SkillSubscription
type SkillSubscriptionArr []*SkillSubscription

// ToAIType converts SkillSubscriptionArr to []*aiTypes.SkillSubscription
func (arr SkillSubscriptionArr) ToAIType() []*aiTypes.SkillSubscription {
	if arr == nil {
		return nil
	}
	result := make([]*aiTypes.SkillSubscription, 0, len(arr))
	for _, s := range arr {
		result = append(result, s.ToAIType())
	}
	return result
}

// Convert SkillSubscription to ai.SkillSubscription
func (s *SkillSubscription) ToAIType() *aiTypes.SkillSubscription {
	if s == nil {
		return nil
	}
	return &aiTypes.SkillSubscription{
		ID:         s.ID,
		SkillID:    s.SkillID,
		SkillName:  s.SkillName,
		Namespace:  s.Namespace,
		ClientID:   s.ClientID,
		ClientHost: s.ClientHost,
		ClientType: s.ClientType,
		Version:    s.Version,
		Active:     s.Active,
		CTime:     s.CreateTime,
		MTime:     s.ModifyTime,
	}
}
