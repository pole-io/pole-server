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
	"strconv"

	"go.uber.org/zap"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	cachetypes "github.com/pole-io/pole-server/apis/cache"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/valid"
)

var (
	// UserFilterAttributes 查询用户所能允许的参数查询列表
	UserFilterAttributes = map[string]bool{
		"id":         true,
		"name":       true,
		"owner":      true,
		"source":     true,
		"offset":     true,
		"group_id":   true,
		"limit":      true,
		"hide_admin": true,
	}
	// UserGroupAttributes is the user link group attributes
	UserGroupAttributes = map[string]struct{}{
		"id":        {},
		"user_id":   {},
		"user_name": {},
		"group_id":  {},
		"name":      {},
		"offset":    {},
		"limit":     {},
		"owner":     {},
	}
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
func (svr *Server) CreateUsers(ctx context.Context, reqs []*apisecurity.User) *apiservice.BatchWriteResponse {
	rsp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, req := range reqs {
		if checkErrResp := checkCreateUser(ctx, req); checkErrResp != nil {
			api.Collect(rsp, checkErrResp)
		}
	}
	if !api.IsSuccess(rsp) {
		return rsp
	}

	return svr.nextSvr.CreateUsers(ctx, reqs)
}

// UpdateUsers 更新用户信息
func (svr *Server) UpdateUsers(ctx context.Context, reqs []*apisecurity.User) *apiservice.BatchWriteResponse {
	rsp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, req := range reqs {
		if checkErrResp := checkUpdateUser(req); checkErrResp != nil {
			api.Collect(rsp, checkErrResp)
		}
	}
	if !api.IsSuccess(rsp) {
		return rsp
	}

	return svr.nextSvr.UpdateUsers(ctx, reqs)
}

// UpdateUserPassword 更新用户密码
func (svr *Server) UpdateUserPassword(ctx context.Context, req *apisecurity.ModifyUserPassword) *apiservice.Response {
	
	return svr.nextSvr.UpdateUserPassword(ctx, req)
}

// DeleteUsers 批量删除用户
func (svr *Server) DeleteUsers(ctx context.Context, users []*apisecurity.User) *apiservice.BatchWriteResponse {
	return svr.nextSvr.DeleteUsers(ctx, users)
}

// GetUsers 查询用户列表
func (svr *Server) GetUsers(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse {
	log.Debug("[Auth][User] origin get users query params", utils.RequestID(ctx), zap.Any("query", query))
	var (
		offset, limit uint32
		err           error
		searchFilters = make(map[string]string, len(query)+1)
	)

	for key, value := range query {
		if _, ok := UserFilterAttributes[key]; !ok {
			log.Error("[Auth][User] attribute it not allowed", utils.RequestID(ctx), zap.String("key", key))
			return api.NewAuthBatchQueryResponseWithMsg(apimodel.Code_InvalidParameter, key+" is not allowed")
		}
		searchFilters[key] = value
	}

	offset, limit, err = valid.ParseOffsetAndLimit(searchFilters)
	if err != nil {
		return api.NewAuthBatchQueryResponse(apimodel.Code_InvalidParameter)
	}
	searchFilters["offset"] = strconv.FormatUint(uint64(offset), 10)
	searchFilters["limit"] = strconv.FormatUint(uint64(limit), 10)
	return svr.nextSvr.GetUsers(ctx, query)
}

// GetUserToken 获取用户的 token
func (svr *Server) GetUserToken(ctx context.Context, req *apisecurity.User) *apiservice.Response {
	if rsp := checkUpdateUser(req); rsp != nil {
		return rsp
	}
	return svr.nextSvr.GetUserToken(ctx, req)
}

// EnableUserToken 禁止用户的token使用
func (svr *Server) EnableUserToken(ctx context.Context, user *apisecurity.User) *apiservice.Response {
	helper := svr.nextSvr.GetUserHelper()
	saveUser := helper.GetUserByID(ctx, user.GetId().GetValue())
	if saveUser == nil {
		return api.NewResponse(apimodel.Code_NotFoundUser)
	}
	if saveUser.GetUserType().GetValue() != strconv.Itoa(int(authtypes.SubAccountUserRole)) {
		return api.NewUserResponseWithMsg(apimodel.Code_NotAllowedAccess, "only disable sub-account token", user)
	}
	return svr.nextSvr.EnableUserToken(ctx, user)
}

// ResetUserToken 重置用户的token
func (svr *Server) ResetUserToken(ctx context.Context, req *apisecurity.User) *apiservice.Response {
	if rsp := checkUpdateUser(req); rsp != nil {
		return rsp
	}
	return svr.nextSvr.ResetUserToken(ctx, req)
}

// CreateGroup 创建用户组
func (svr *Server) CreateGroup(ctx context.Context, group *apisecurity.UserGroup) *apiservice.Response {
	return svr.nextSvr.CreateGroup(ctx, group)
}

// UpdateGroups 更新用户组
func (svr *Server) UpdateGroups(ctx context.Context, groups []*apisecurity.UserGroup) *apiservice.BatchWriteResponse {
	return svr.nextSvr.UpdateGroups(ctx, groups)
}

// DeleteGroups 批量删除用户组
func (svr *Server) DeleteGroups(ctx context.Context, groups []*apisecurity.UserGroup) *apiservice.BatchWriteResponse {
	return svr.nextSvr.DeleteGroups(ctx, groups)
}

// GetGroups 查询用户组列表（不带用户详细信息）
func (svr *Server) GetGroups(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse {
	log.Info("[Auth][Group] origin get groups query params",
		utils.RequestID(ctx), zap.Any("query", query))

	offset, limit, err := valid.ParseOffsetAndLimit(query)
	if err != nil {
		return api.NewAuthBatchQueryResponse(apimodel.Code_InvalidParameter)
	}

	searchFilters := make(map[string]string, len(query))
	for key, value := range query {
		if _, ok := UserGroupAttributes[key]; !ok {
			log.Error("[Auth][Group] get groups attribute it not allowed", utils.RequestID(ctx), zap.String("key", key))
			return api.NewAuthBatchQueryResponseWithMsg(apimodel.Code_InvalidParameter, key+" is not allowed")
		}
		searchFilters[key] = value
	}

	searchFilters["offset"] = strconv.FormatUint(uint64(offset), 10)
	searchFilters["limit"] = strconv.FormatUint(uint64(limit), 10)
	return svr.nextSvr.GetGroups(ctx, query)
}

// GetGroup 根据用户组信息，查询该用户组下的用户相信
func (svr *Server) GetGroup(ctx context.Context, req *apisecurity.UserGroup) *apiservice.Response {
	return svr.nextSvr.GetGroup(ctx, req)
}

// GetGroupToken 获取用户组的 token
func (svr *Server) GetGroupToken(ctx context.Context, group *apisecurity.UserGroup) *apiservice.Response {
	return svr.nextSvr.GetGroupToken(ctx, group)
}

// UpdateGroupToken 取消用户组的 token 使用
func (svr *Server) EnableGroupToken(ctx context.Context, group *apisecurity.UserGroup) *apiservice.Response {
	return svr.nextSvr.EnableGroupToken(ctx, group)
}

// ResetGroupToken 重置用户组的 token
func (svr *Server) ResetGroupToken(ctx context.Context, group *apisecurity.UserGroup) *apiservice.Response {
	return svr.nextSvr.ResetGroupToken(ctx, group)
}

// checkCreateUser 检查创建用户的请求
func checkCreateUser(ctx context.Context, req *apisecurity.User) *apiservice.Response {
	if req == nil {
		return api.NewUserResponse(apimodel.Code_EmptyRequest, req)
	}

	if err := CheckName(req.Name); err != nil {
		return api.NewUserResponse(apimodel.Code_InvalidUserName, req)
	}

	if err := CheckPassword(req.Password); err != nil {
		return api.NewUserResponse(apimodel.Code_InvalidUserPassword, req)
	}
	return nil
}

// checkUpdateUser 检查用户更新请求
func checkUpdateUser(req *apisecurity.User) *apiservice.Response {
	if req == nil {
		return api.NewUserResponse(apimodel.Code_EmptyRequest, req)
	}

	if req.GetId() == nil || req.GetId().GetValue() == "" {
		return api.NewUserResponse(apimodel.Code_BadRequest, req)
	}
	return nil
}
