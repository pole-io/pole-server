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

package paramcheck

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	aiTypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
	"github.com/pole-io/pole-server/pkg/skill"
)

// Server 参数校验拦截器（最外层）
type Server struct {
	nextSvr skill.SkillServer
	logger  *log.Scope
}

// NewServer 创建带参数校验的 SkillServer
func NewServer(nextSvr skill.SkillServer) skill.SkillServer {
	return &Server{
		nextSvr: nextSvr,
		logger:  log.RegisterScope("skill_paramcheck_interceptor", "Skill paramcheck interceptor logging", 1),
	}
}

// CreateSkill 创建 Skill（带参数校验）
func (svr *Server) CreateSkill(ctx context.Context, s *aiTypes.Skill) error {
	if err := svr.checkCreateSkill(s); err != nil {
		return err
	}
	return svr.nextSvr.CreateSkill(ctx, s)
}

// checkCreateSkill 校验创建 Skill 的参数
func (svr *Server) checkCreateSkill(s *aiTypes.Skill) error {
	if s == nil {
		return errors.New("skill cannot be nil")
	}

	if s.Name == "" {
		return errors.New("skill name cannot be empty")
	}

	if s.Namespace == "" {
		return errors.New("skill namespace cannot be empty")
	}

	if s.SkillType == "" {
		return errors.New("skill type cannot be empty")
	}

	// 校验名称格式
	if err := valid.CheckResourceName(s.Name); err != nil {
		return fmt.Errorf("invalid skill name: %w", err)
	}

	return nil
}

// UpdateSkill 更新 Skill（带参数校验）
func (svr *Server) UpdateSkill(ctx context.Context, s *aiTypes.Skill) error {
	if err := svr.checkUpdateSkill(s); err != nil {
		return err
	}
	return svr.nextSvr.UpdateSkill(ctx, s)
}

// checkUpdateSkill 校验更新 Skill 的参数
func (svr *Server) checkUpdateSkill(s *aiTypes.Skill) error {
	if s == nil {
		return errors.New("skill cannot be nil")
	}

	if s.ID == "" {
		return errors.New("skill id cannot be empty")
	}

	// 校验名称格式
	if s.Name != "" {
		if err := valid.CheckResourceName(s.Name); err != nil {
			return fmt.Errorf("invalid skill name: %w", err)
		}
	}

	return nil
}

// DeleteSkill 删除 Skill（带参数校验）
func (svr *Server) DeleteSkill(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("skill id cannot be empty")
	}
	return svr.nextSvr.DeleteSkill(ctx, id)
}

// GetSkill 获取 Skill
func (svr *Server) GetSkill(ctx context.Context, id string) (*aiTypes.Skill, error) {
	if id == "" {
		return nil, errors.New("skill id cannot be empty")
	}
	return svr.nextSvr.GetSkill(ctx, id)
}

// GetSkillByName 根据名称获取 Skill
func (svr *Server) GetSkillByName(ctx context.Context, name, namespace string) (*aiTypes.Skill, error) {
	if name == "" {
		return nil, errors.New("skill name cannot be empty")
	}
	if namespace == "" {
		return nil, errors.New("skill namespace cannot be empty")
	}
	return svr.nextSvr.GetSkillByName(ctx, name, namespace)
}

// CreateSkillGroup 创建 SkillGroup（带参数校验）
func (svr *Server) CreateSkillGroup(ctx context.Context, group *aiTypes.SkillGroup) error {
	if group == nil {
		return errors.New("skill group cannot be nil")
	}
	if group.Name == "" {
		return errors.New("skill group name cannot be empty")
	}
	if group.Namespace == "" {
		return errors.New("skill group namespace cannot be empty")
	}
	return svr.nextSvr.CreateSkillGroup(ctx, group)
}

// UpdateSkillGroup 更新 SkillGroup
func (svr *Server) UpdateSkillGroup(ctx context.Context, group *aiTypes.SkillGroup) error {
	if group == nil {
		return errors.New("skill group cannot be nil")
	}
	return svr.nextSvr.UpdateSkillGroup(ctx, group)
}

// DeleteSkillGroup 删除 SkillGroup
func (svr *Server) DeleteSkillGroup(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		return errors.New("namespace cannot be empty")
	}
	if name == "" {
		return errors.New("name cannot be empty")
	}
	return svr.nextSvr.DeleteSkillGroup(ctx, namespace, name)
}

// GetSkillGroup 获取 SkillGroup
func (svr *Server) GetSkillGroup(ctx context.Context, namespace, name string) (*aiTypes.SkillGroup, error) {
	return svr.nextSvr.GetSkillGroup(ctx, namespace, name)
}

// CreateSkillVersion 创建 SkillVersion
func (svr *Server) CreateSkillVersion(ctx context.Context, version *aiTypes.SkillVersion) error {
	if version == nil {
		return errors.New("skill version cannot be nil")
	}
	return svr.nextSvr.CreateSkillVersion(ctx, version)
}

// ActivateSkillVersion 激活 SkillVersion
func (svr *Server) ActivateSkillVersion(ctx context.Context, versionID string) error {
	if versionID == "" {
		return errors.New("version id cannot be empty")
	}
	return svr.nextSvr.ActivateSkillVersion(ctx, versionID)
}

// CreateSkillSubscription 创建 SkillSubscription
func (svr *Server) CreateSkillSubscription(ctx context.Context, sub *aiTypes.SkillSubscription) error {
	if sub == nil {
		return errors.New("skill subscription cannot be nil")
	}
	return svr.nextSvr.CreateSkillSubscription(ctx, sub)
}

// DeleteSkillSubscription 删除 SkillSubscription
func (svr *Server) DeleteSkillSubscription(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("subscription id cannot be empty")
	}
	return svr.nextSvr.DeleteSkillSubscription(ctx, id)
}

// GetSkillSubscriptionsBySkill 获取 Skill 订阅列表
func (svr *Server) GetSkillSubscriptionsBySkill(ctx context.Context, skillName, namespace string) ([]*aiTypes.SkillSubscription, error) {
	return svr.nextSvr.GetSkillSubscriptionsBySkill(ctx, skillName, namespace)
}

// GetSkillSubscriptionsByClient 获取客户端订阅列表
func (svr *Server) GetSkillSubscriptionsByClient(ctx context.Context, clientID string) ([]*aiTypes.SkillSubscription, error) {
	return svr.nextSvr.GetSkillSubscriptionsByClient(ctx, clientID)
}

// ===== Batch Operations (delegate to next server) =====

// CreateSkills creates multiple skills (batch)
func (svr *Server) CreateSkills(ctx context.Context, skills []*aiTypes.Skill) *apimodel.BatchWriteResponse {
	return svr.nextSvr.CreateSkills(ctx, skills)
}

// UpdateSkills updates multiple skills (batch)
func (svr *Server) UpdateSkills(ctx context.Context, skills []*aiTypes.Skill) *apimodel.BatchWriteResponse {
	return svr.nextSvr.UpdateSkills(ctx, skills)
}

// DeleteSkills deletes multiple skills (batch)
func (svr *Server) DeleteSkills(ctx context.Context, skills []*aiTypes.Skill) *apimodel.BatchWriteResponse {
	return svr.nextSvr.DeleteSkills(ctx, skills)
}

// GetSkills queries skills with filters
func (svr *Server) GetSkills(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return svr.nextSvr.GetSkills(ctx, query)
}

// GetAllSkills gets all skills
func (svr *Server) GetAllSkills(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return svr.nextSvr.GetAllSkills(ctx, query)
}

// GetSkillsCount gets the total count of skills
func (svr *Server) GetSkillsCount(ctx context.Context) *apimodel.BatchQueryResponse {
	return svr.nextSvr.GetSkillsCount(ctx)
}

// CreateSkillGroups creates multiple skill groups (batch)
func (svr *Server) CreateSkillGroups(ctx context.Context, groups []*aiTypes.SkillGroup) *apimodel.BatchWriteResponse {
	return svr.nextSvr.CreateSkillGroups(ctx, groups)
}

// UpdateSkillGroups updates multiple skill groups (batch)
func (svr *Server) UpdateSkillGroups(ctx context.Context, groups []*aiTypes.SkillGroup) *apimodel.BatchWriteResponse {
	return svr.nextSvr.UpdateSkillGroups(ctx, groups)
}

// DeleteSkillGroups deletes multiple skill groups (batch)
func (svr *Server) DeleteSkillGroups(ctx context.Context, groups []*aiTypes.SkillGroup) *apimodel.BatchWriteResponse {
	return svr.nextSvr.DeleteSkillGroups(ctx, groups)
}

// GetSkillGroups queries skill groups with filters
func (svr *Server) GetSkillGroups(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return svr.nextSvr.GetSkillGroups(ctx, query)
}
