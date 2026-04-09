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

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	aiTypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/access_control/auth"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/skill"
)

// Server Skill 认证拦截器（装饰器模式）
type Server struct {
	nextSvr   skill.SkillServer
	userSvr   auth.UserServer
	policySvr auth.StrategyServer
	cacheMgr  cacheapi.CacheManager
	logger    *log.Scope
}

// NewServer 创建带认证的 SkillServer
func NewServer(nextSvr skill.SkillServer,
	userSvr auth.UserServer,
	policySvr auth.StrategyServer,
	cacheMgr cacheapi.CacheManager) skill.SkillServer {
	return &Server{
		nextSvr:   nextSvr,
		userSvr:   userSvr,
		policySvr: policySvr,
		cacheMgr:  cacheMgr,
		logger:    log.RegisterScope("skill_auth_interceptor", "Skill auth interceptor logging", 1),
	}
}

// Cache 获取缓存管理器
func (svr *Server) Cache() cacheapi.CacheManager {
	return svr.cacheMgr
}

// collectSkillAuthContext 收集 Skill 认证上下文
func (svr *Server) collectSkillAuthContext(ctx context.Context,
	s *aiTypes.Skill, resourceOp authtypes.ResourceOperation,
	methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {

	svr.logger.Debugf("[Auth][Skill] collect auth context for skill: %s, operation: %d, method: %s",
		s.Name, resourceOp, methodName)

	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(resourceOp),
		authtypes.WithModule(authtypes.CoreModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(svr.querySkillResource(s)),
	)

	svr.logger.Debug("[Auth][Skill] auth context collected for skill")
	return authCtx
}

// querySkillResource 查询 Skill 资源信息
func (svr *Server) querySkillResource(s *aiTypes.Skill) map[apisecurity.ResourceType][]authtypes.ResourceEntry {
	if s == nil {
		return make(map[apisecurity.ResourceType][]authtypes.ResourceEntry)
	}

	nsRet := make([]authtypes.ResourceEntry, 0)
	if s.Namespace != "" && svr.cacheMgr != nil {
		ns := svr.cacheMgr.Namespace().GetNamespace(s.Namespace)
		if ns != nil {
			nsRet = append(nsRet, authtypes.ResourceEntry{
				Type:     apisecurity.ResourceType_Namespaces,
				ID:       ns.Name,
				Owner:    ns.Owner,
				Metadata: ns.Metadata,
			})
		}
	}

	return map[apisecurity.ResourceType][]authtypes.ResourceEntry{
		apisecurity.ResourceType_Namespaces: nsRet,
	}
}

// setOwnerFromContext 从上下文获取 owner 并设置到 skill
func (svr *Server) setOwnerFromContext(ctx context.Context, s *aiTypes.Skill) {
	ownerID := utils.ParseOwnerID(ctx)
	if len(ownerID) > 0 && s.Metadata != nil {
		svr.logger.Infof("[Auth][Skill] set owner for skill %s from context: %s", s.Name, ownerID)
		s.Metadata["owner"] = ownerID
	}
}

// CreateSkill 创建 Skill（带认证）
func (svr *Server) CreateSkill(ctx context.Context, s *aiTypes.Skill) error {
	authCtx := svr.collectSkillAuthContext(ctx, s, authtypes.Create, "CreateSkill")

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	// 填充 owner 信息
	svr.setOwnerFromContext(ctx, s)

	return svr.nextSvr.CreateSkill(ctx, s)
}

// UpdateSkill 更新 Skill（带认证）
func (svr *Server) UpdateSkill(ctx context.Context, s *aiTypes.Skill) error {
	authCtx := svr.collectSkillAuthContext(ctx, s, authtypes.Modify, "UpdateSkill")

	// 移除命名空间资源检查，只检查 Skill 本身
	accessRes := authCtx.GetAccessResources()
	delete(accessRes, apisecurity.ResourceType_Namespaces)
	authCtx.SetAccessResources(accessRes)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.UpdateSkill(ctx, s)
}

// DeleteSkill 删除 Skill（带认证）
func (svr *Server) DeleteSkill(ctx context.Context, id string) error {
	// 先获取 Skill 信息用于权限检查
	s, err := svr.nextSvr.GetSkill(ctx, id)
	if err != nil {
		return err
	}

	authCtx := svr.collectSkillAuthContext(ctx, s, authtypes.Delete, "DeleteSkill")

	// 移除命名空间资源检查
	accessRes := authCtx.GetAccessResources()
	delete(accessRes, apisecurity.ResourceType_Namespaces)
	authCtx.SetAccessResources(accessRes)

	_, err = svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.DeleteSkill(ctx, id)
}

// GetSkill 获取 Skill（带认证）
func (svr *Server) GetSkill(ctx context.Context, id string) (*aiTypes.Skill, error) {
	authCtx := svr.collectSkillAuthContext(ctx, nil, authtypes.Read, "GetSkill")

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return nil, err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetSkill(ctx, id)
}

// GetSkillByName 根据名称获取 Skill（带认证）
func (svr *Server) GetSkillByName(ctx context.Context, name, namespace string) (*aiTypes.Skill, error) {
	authCtx := svr.collectSkillAuthContext(ctx, &aiTypes.Skill{
		Name:      name,
		Namespace: namespace,
	}, authtypes.Read, "GetSkillByName")

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return nil, err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetSkillByName(ctx, name, namespace)
}

// CreateSkillGroup 创建 SkillGroup（带认证）
func (svr *Server) CreateSkillGroup(ctx context.Context, group *aiTypes.SkillGroup) error {
	authCtx := svr.collectSkillGroupAuthContext(ctx, group, authtypes.Create, "CreateSkillGroup")

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.CreateSkillGroup(ctx, group)
}

// UpdateSkillGroup 更新 SkillGroup（带认证）
func (svr *Server) UpdateSkillGroup(ctx context.Context, group *aiTypes.SkillGroup) error {
	authCtx := svr.collectSkillGroupAuthContext(ctx, group, authtypes.Modify, "UpdateSkillGroup")

	accessRes := authCtx.GetAccessResources()
	delete(accessRes, apisecurity.ResourceType_Namespaces)
	authCtx.SetAccessResources(accessRes)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.UpdateSkillGroup(ctx, group)
}

// DeleteSkillGroup 删除 SkillGroup（带认证）
func (svr *Server) DeleteSkillGroup(ctx context.Context, namespace, name string) error {
	group, err := svr.nextSvr.GetSkillGroup(ctx, namespace, name)
	if err != nil {
		return err
	}

	authCtx := svr.collectSkillGroupAuthContext(ctx, group, authtypes.Delete, "DeleteSkillGroup")

	accessRes := authCtx.GetAccessResources()
	delete(accessRes, apisecurity.ResourceType_Namespaces)
	authCtx.SetAccessResources(accessRes)

	_, err = svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.DeleteSkillGroup(ctx, namespace, name)
}

// GetSkillGroup 获取 SkillGroup（带认证）
func (svr *Server) GetSkillGroup(ctx context.Context, namespace, name string) (*aiTypes.SkillGroup, error) {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Read),
		authtypes.WithModule(authtypes.CoreModule),
		authtypes.WithMethod("GetSkillGroup"),
	)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return nil, err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetSkillGroup(ctx, namespace, name)
}

// collectSkillGroupAuthContext 收集 SkillGroup 认证上下文
func (svr *Server) collectSkillGroupAuthContext(ctx context.Context, group *aiTypes.SkillGroup,
	resourceOp authtypes.ResourceOperation, methodName authtypes.ServerFunctionName) *authtypes.AcquireContext {

	svr.logger.Debugf("[Auth][SkillGroup] collect auth context for group: %s, operation: %d, method: %s",
		group.Name, resourceOp, methodName)

	nsRet := make([]authtypes.ResourceEntry, 0)
	if group.Namespace != "" && svr.cacheMgr != nil {
		ns := svr.cacheMgr.Namespace().GetNamespace(group.Namespace)
		if ns != nil {
			nsRet = append(nsRet, authtypes.ResourceEntry{
				Type:     apisecurity.ResourceType_Namespaces,
				ID:       ns.Name,
				Owner:    ns.Owner,
				Metadata: ns.Metadata,
			})
		}
	}

	return authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(resourceOp),
		authtypes.WithModule(authtypes.CoreModule),
		authtypes.WithMethod(methodName),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_Namespaces: nsRet,
		}),
	)
}

// CreateSkillVersion 创建 SkillVersion（带认证）
func (svr *Server) CreateSkillVersion(ctx context.Context, version *aiTypes.SkillVersion) error {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Create),
		authtypes.WithModule(authtypes.CoreModule),
		authtypes.WithMethod("CreateSkillVersion"),
	)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.CreateSkillVersion(ctx, version)
}

// ActivateSkillVersion 激活 SkillVersion（带认证）
func (svr *Server) ActivateSkillVersion(ctx context.Context, versionID string) error {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Modify),
		authtypes.WithModule(authtypes.CoreModule),
		authtypes.WithMethod("ActivateSkillVersion"),
	)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.ActivateSkillVersion(ctx, versionID)
}

// CreateSkillSubscription 创建 SkillSubscription（带认证）
func (svr *Server) CreateSkillSubscription(ctx context.Context, sub *aiTypes.SkillSubscription) error {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Create),
		authtypes.WithModule(authtypes.CoreModule),
		authtypes.WithMethod("CreateSkillSubscription"),
	)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.CreateSkillSubscription(ctx, sub)
}

// DeleteSkillSubscription 删除 SkillSubscription（带认证）
func (svr *Server) DeleteSkillSubscription(ctx context.Context, id string) error {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Delete),
		authtypes.WithModule(authtypes.CoreModule),
		authtypes.WithMethod("DeleteSkillSubscription"),
	)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.DeleteSkillSubscription(ctx, id)
}

// GetSkillSubscriptionsBySkill 获取 Skill 订阅列表（带认证）
func (svr *Server) GetSkillSubscriptionsBySkill(ctx context.Context, skillName, namespace string) ([]*aiTypes.SkillSubscription, error) {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Read),
		authtypes.WithModule(authtypes.CoreModule),
		authtypes.WithMethod("GetSkillSubscriptionsBySkill"),
	)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return nil, err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

	return svr.nextSvr.GetSkillSubscriptionsBySkill(ctx, skillName, namespace)
}

// GetSkillSubscriptionsByClient 获取客户端订阅列表（带认证）
func (svr *Server) GetSkillSubscriptionsByClient(ctx context.Context, clientID string) ([]*aiTypes.SkillSubscription, error) {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Read),
		authtypes.WithModule(authtypes.CoreModule),
		authtypes.WithMethod("GetSkillSubscriptionsByClient"),
	)

	_, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx)
	if err != nil {
		return nil, err
	}

	ctx = authCtx.GetRequestContext()
	ctx = context.WithValue(ctx, types.ContextAuthContextKey, authCtx)

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
