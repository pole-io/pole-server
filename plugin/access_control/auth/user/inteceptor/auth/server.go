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

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	cachetypes "github.com/pole-io/pole-server/apis/cache"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

func NewServer(nextSvr authapi.UserServer) authapi.UserServer {
	return &Server{
		nextSvr: nextSvr,
	}
}

type Server struct {
	nextSvr   authapi.UserServer
	policySvr authapi.StrategyServer
}

// Initialize 初始化
func (svr *Server) Initialize(authOpt *authapi.Config, storage store.Store, policySvr authapi.StrategyServer,
	cacheMgr cachetypes.CacheManager) error {
	svr.policySvr = policySvr
	return svr.nextSvr.Initialize(authOpt, storage, policySvr, cacheMgr)
}

// Name 用户数据管理server名称
func (svr *Server) Name() string {
	return svr.nextSvr.Name()
}

// Login 登录动作
func (svr *Server) Login(req *apisecurity.LoginRequest) *apiservice.Response {
	return svr.nextSvr.Login(req)
}

// CheckCredential 检查当前操作用户凭证
func (svr *Server) CheckCredential(authCtx *authtypes.AcquireContext) error {
	return svr.nextSvr.CheckCredential(authCtx)
}

// GetUserHelper
func (svr *Server) GetUserHelper() authapi.UserHelper {
	return svr.nextSvr.GetUserHelper()
}

// CreateUsers 批量创建用户
func (svr *Server) CreateUsers(ctx context.Context, users []*apisecurity.User) *apiservice.BatchWriteResponse {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Create),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.CreateUsers),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.CreateUsers(authCtx.GetRequestContext(), users)
}

// UpdateUsers 更新用户信息
func (svr *Server) UpdateUsers(ctx context.Context, reqs []*apisecurity.User) *apiservice.BatchWriteResponse {
	rsp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)

	helper := svr.nextSvr.GetUserHelper()
	resources := make([]authtypes.ResourceEntry, 0, len(reqs))
	for _, req := range reqs {
		saveUser := helper.GetUserByID(ctx, req.GetId().GetValue())
		if saveUser == nil {
			api.Collect(rsp, api.NewUserResponse(apimodel.Code_NotFoundUser, req))
			continue
		}

		resources = append(resources, authtypes.ResourceEntry{
			ID:       req.GetId().GetValue(),
			Type:     apisecurity.ResourceType_Users,
			Metadata: saveUser.Metadata,
		})
	}

	if !api.IsSuccess(rsp) {
		return rsp
	}

	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Modify),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.UpdateUser),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_Users: resources,
		}),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.UpdateUsers(authCtx.GetRequestContext(), reqs)
}

// UpdateUserPassword 更新用户密码
func (svr *Server) UpdateUserPassword(ctx context.Context, req *apisecurity.ModifyUserPassword) *apiservice.Response {
	helper := svr.nextSvr.GetUserHelper()
	saveUser := helper.GetUserByID(ctx, req.GetId().GetValue())
	if saveUser == nil {
		return api.NewResponse(apimodel.Code_NotFoundUser)
	}

	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Modify),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.UpdateUserPassword),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_Users: {
				authtypes.ResourceEntry{
					ID:       req.GetId().GetValue(),
					Type:     apisecurity.ResourceType_Users,
					Metadata: saveUser.Metadata,
				},
			},
		}),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.UpdateUserPassword(authCtx.GetRequestContext(), req)
}

// DeleteUsers 批量删除用户
func (svr *Server) DeleteUsers(ctx context.Context, users []*apisecurity.User) *apiservice.BatchWriteResponse {
	helper := svr.nextSvr.GetUserHelper()
	resources := make([]authtypes.ResourceEntry, 0, len(users))
	for i := range users {
		saveUser := helper.GetUserByID(ctx, users[i].GetId().GetValue())
		if saveUser == nil {
			return api.NewBatchWriteResponse(apimodel.Code_NotFoundUser)
		}
		resources = append(resources, authtypes.ResourceEntry{
			ID:       users[i].GetId().GetValue(),
			Type:     apisecurity.ResourceType_Users,
			Metadata: saveUser.Metadata,
		})
	}
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Delete),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.DeleteUsers),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_Users: resources,
		}),
	)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.DeleteUsers(authCtx.GetRequestContext(), users)
}

// GetUsers 查询用户列表
func (svr *Server) GetUsers(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Read),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.DescribeUsers),
	)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = cachetypes.AppendUserPredicate(ctx, func(ctx context.Context, u *authtypes.User) bool {
		return svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     apisecurity.ResourceType_Users,
			ID:       u.ID,
			Metadata: u.Metadata,
		})
	})

	return svr.nextSvr.GetUsers(authCtx.GetRequestContext(), query)
}

// GetUserToken 获取用户的 token
func (svr *Server) GetUserToken(ctx context.Context, user *apisecurity.User) *apiservice.Response {
	helper := svr.nextSvr.GetUserHelper()
	saveUser := helper.GetUserByID(ctx, user.GetId().GetValue())
	if saveUser == nil {
		return api.NewResponse(apimodel.Code_NotFoundUser)
	}
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Read),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.DescribeUserToken),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_Users: {
				authtypes.ResourceEntry{
					ID:       user.GetId().GetValue(),
					Type:     apisecurity.ResourceType_Users,
					Metadata: saveUser.Metadata,
				},
			},
		}),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.GetUserToken(authCtx.GetRequestContext(), user)
}

// UpdateUserToken 禁止用户的token使用
func (svr *Server) EnableUserToken(ctx context.Context, user *apisecurity.User) *apiservice.Response {
	helper := svr.nextSvr.GetUserHelper()
	saveUser := helper.GetUserByID(ctx, user.GetId().GetValue())
	if saveUser == nil {
		return api.NewResponse(apimodel.Code_NotFoundUser)
	}
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Modify),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.EnableUserToken),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_Users: {
				authtypes.ResourceEntry{
					ID:       user.GetId().GetValue(),
					Type:     apisecurity.ResourceType_Users,
					Metadata: saveUser.Metadata,
				},
			},
		}),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.EnableUserToken(authCtx.GetRequestContext(), user)
}

// ResetUserToken 重置用户的token
func (svr *Server) ResetUserToken(ctx context.Context, user *apisecurity.User) *apiservice.Response {
	helper := svr.nextSvr.GetUserHelper()
	saveUser := helper.GetUserByID(ctx, user.GetId().GetValue())
	if saveUser == nil {
		return api.NewResponse(apimodel.Code_NotFoundUser)
	}
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Modify),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.ResetUserToken),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_Users: {
				authtypes.ResourceEntry{
					ID:       user.GetId().GetValue(),
					Type:     apisecurity.ResourceType_Users,
					Metadata: saveUser.Metadata,
				},
			},
		}),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponseWithMsg(authtypes.ConvertToErrCode(err), err.Error())
	}
	return svr.nextSvr.ResetUserToken(authCtx.GetRequestContext(), user)
}

// CreateGroup 创建用户组
func (svr *Server) CreateGroup(ctx context.Context, group *apisecurity.UserGroup) *apiservice.Response {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Create),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.CreateUserGroup),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.CreateGroup(authCtx.GetRequestContext(), group)
}

// UpdateGroups 更新用户组
func (svr *Server) UpdateGroups(ctx context.Context, groups []*apisecurity.UserGroup) *apiservice.BatchWriteResponse {
	helper := svr.nextSvr.GetUserHelper()
	resources := make([]authtypes.ResourceEntry, 0, len(groups))
	for i := range groups {
		saveGroup := helper.GetGroup(ctx, &apisecurity.UserGroup{Id: groups[i].GetId()})
		if saveGroup == nil {
			return api.NewBatchWriteResponse(apimodel.Code_NotFoundUserGroup)
		}
		resources = append(resources, authtypes.ResourceEntry{
			Type:     apisecurity.ResourceType_UserGroups,
			ID:       groups[i].GetId().GetValue(),
			Metadata: saveGroup.Metadata,
		})
	}

	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Modify),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.UpdateUserGroups),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_UserGroups: resources,
		}),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.UpdateGroups(authCtx.GetRequestContext(), groups)
}

// DeleteGroups 批量删除用户组
func (svr *Server) DeleteGroups(ctx context.Context, groups []*apisecurity.UserGroup) *apiservice.BatchWriteResponse {
	helper := svr.nextSvr.GetUserHelper()
	resources := make([]authtypes.ResourceEntry, 0, len(groups))
	for i := range groups {
		saveGroup := helper.GetGroup(ctx, &apisecurity.UserGroup{Id: groups[i].GetId()})
		if saveGroup == nil {
			return api.NewBatchWriteResponse(apimodel.Code_NotFoundUserGroup)
		}
		resources = append(resources, authtypes.ResourceEntry{
			ID: groups[i].GetId().GetValue(),
		})
	}

	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Modify),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.DeleteUserGroups),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_UserGroups: resources,
		}),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.DeleteGroups(authCtx.GetRequestContext(), groups)
}

// GetGroups 查询用户组列表（不带用户详细信息）
func (svr *Server) GetGroups(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Read),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.DescribeUserGroups),
	)
	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}
	ctx = authCtx.GetRequestContext()
	ctx = cachetypes.AppendUserGroupPredicate(ctx, func(ctx context.Context, u *authtypes.UserGroupDetail) bool {
		ok := svr.policySvr.GetAuthChecker().ResourcePredicate(authCtx, &authtypes.ResourceEntry{
			Type:     apisecurity.ResourceType_UserGroups,
			ID:       u.ID,
			Metadata: u.Metadata,
		})
		if ok {
			return true
		}
		// 兼容老版本的策略查询逻辑
		if compatible, _ := ctx.Value(utils.ContextKeyCompatible{}).(bool); compatible {
			_, exist := u.UserIds[utils.ParseUserID(ctx)]
			return exist
		}
		return false
	})
	authCtx.SetRequestContext(ctx)
	return svr.nextSvr.GetGroups(authCtx.GetRequestContext(), query)
}

// GetGroup 根据用户组信息，查询该用户组下的用户相信
func (svr *Server) GetGroup(ctx context.Context, req *apisecurity.UserGroup) *apiservice.Response {
	helper := svr.nextSvr.GetUserHelper()
	saveGroup := helper.GetGroup(ctx, req)
	if saveGroup == nil {
		return api.NewResponse(apimodel.Code_NotFoundUserGroup)
	}
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Read),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.DescribeUserGroupDetail),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_UserGroups: {
				authtypes.ResourceEntry{
					Type:     apisecurity.ResourceType_UserGroups,
					ID:       req.GetId().GetValue(),
					Metadata: saveGroup.Metadata,
				},
			},
		}),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.GetGroup(authCtx.GetRequestContext(), req)
}

// GetGroupToken 获取用户组的 token
func (svr *Server) GetGroupToken(ctx context.Context, group *apisecurity.UserGroup) *apiservice.Response {
	helper := svr.nextSvr.GetUserHelper()
	saveGroup := helper.GetGroup(ctx, group)
	if saveGroup == nil {
		return api.NewResponse(apimodel.Code_NotFoundUserGroup)
	}
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Read),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.DescribeUserGroupToken),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_UserGroups: {
				authtypes.ResourceEntry{
					ID:       group.GetId().GetValue(),
					Type:     apisecurity.ResourceType_UserGroups,
					Metadata: saveGroup.Metadata,
				},
			},
		}),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.GetGroupToken(authCtx.GetRequestContext(), group)
}

// EnableGroupToken 取消用户组的 token 使用
func (svr *Server) EnableGroupToken(ctx context.Context, group *apisecurity.UserGroup) *apiservice.Response {
	helper := svr.nextSvr.GetUserHelper()
	saveGroup := helper.GetGroup(ctx, group)
	if saveGroup == nil {
		return api.NewResponse(apimodel.Code_NotFoundUserGroup)
	}
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Modify),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.EnableUserGroupToken),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_UserGroups: {
				authtypes.ResourceEntry{
					ID:       group.GetId().GetValue(),
					Type:     apisecurity.ResourceType_UserGroups,
					Metadata: saveGroup.Metadata,
				},
			},
		}),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.EnableGroupToken(authCtx.GetRequestContext(), group)
}

// ResetGroupToken 重置用户组的 token
func (svr *Server) ResetGroupToken(ctx context.Context, group *apisecurity.UserGroup) *apiservice.Response {
	helper := svr.nextSvr.GetUserHelper()
	saveGroup := helper.GetGroup(ctx, group)
	if saveGroup == nil {
		return api.NewResponse(apimodel.Code_NotFoundUserGroup)
	}
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Modify),
		authtypes.WithModule(authtypes.AuthModule),
		authtypes.WithMethod(authtypes.ResetUserGroupToken),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_UserGroups: {
				authtypes.ResourceEntry{
					ID:       group.GetId().GetValue(),
					Type:     apisecurity.ResourceType_UserGroups,
					Metadata: saveGroup.Metadata,
				},
			},
		}),
	)

	if _, err := svr.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewResponse(authtypes.ConvertToErrCode(err))
	}
	return svr.nextSvr.ResetGroupToken(authCtx.GetRequestContext(), group)
}
