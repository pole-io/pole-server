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

package defaultuser

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	cachetypes "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	commontime "github.com/pole-io/pole-server/pkg/common/time"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/valid"
)

type (
	// User2Api convert user to api.User
	User2Api func(user *authtypes.User) *apisecurity.User
)

// CreateUsers 批量创建用户
func (svr *Server) CreateUsers(ctx context.Context, req []*apisecurity.User) *apiservice.BatchWriteResponse {
	batchResp := api.NewAuthBatchWriteResponse(apimodel.Code_ExecuteSuccess)

	for i := range req {
		resp := svr.CreateUser(ctx, req[i])
		api.Collect(batchResp, resp)
	}

	return batchResp
}

// CreateUser 创建用户
func (svr *Server) CreateUser(ctx context.Context, req *apisecurity.User) *apiservice.Response {
	ownerID := utils.ParseOwnerID(ctx)
	req.Owner = protobuf.NewStringValue(ownerID)

	if ownerID != "" {
		owner, err := svr.storage.GetUser(ownerID)
		if err != nil {
			log.Error("[Auth][User] get owner user", utils.RequestID(ctx), zap.Error(err), zap.String("owner", ownerID))
			return api.NewUserResponse(storeapi.StoreCode2APICode(err), req)
		}

		if owner.Name == req.Name.GetValue() {
			log.Error("[Auth][User] create user name is equal owner", utils.RequestID(ctx),
				zap.Error(err), zap.String("name", req.GetName().GetValue()))
			return api.NewUserResponse(apimodel.Code_UserExisted, req)
		}
	}

	// 只有通过 owner + username 才能唯一确定一个用户
	user, err := svr.storage.GetUserByName(req.Name.GetValue())
	if err != nil {
		log.Error("[Auth][User] get user by name and owner", utils.RequestID(ctx),
			zap.Error(err), zap.String("owner", ownerID), zap.String("name", req.GetName().GetValue()))
		return api.NewUserResponse(storeapi.StoreCode2APICode(err), req)
	}
	if user != nil {
		return api.NewUserResponse(apimodel.Code_UserExisted, req)
	}

	return svr.createUser(ctx, req)
}

func (svr *Server) createUser(ctx context.Context, req *apisecurity.User) *apiservice.Response {
	data, err := svr.createUserModel(req, authtypes.SubAccountUserRole)
	if err != nil {
		log.Error("[Auth][User] create user model", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponse(apimodel.Code_ExecuteException)
	}

	tx, err := svr.storage.StartTx()
	if err != nil {
		log.Error("[Auth][User] create user begion storage tx", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponse(apimodel.Code_ExecuteException)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := svr.storage.AddUser(tx, data); err != nil {
		log.Error("[Auth][User] add user into store", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}

	if err := svr.policySvr.PolicyHelper().CreatePrincipalPolicy(ctx, tx, authtypes.Principal{
		PrincipalID:   data.ID,
		PrincipalType: authtypes.PrincipalUser,
		Owner:         "",
		Name:          data.Name,
	}); err != nil {
		log.Error("[Auth][User] add user default policy rule", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}

	if err := tx.Commit(); err != nil {
		log.Error("[Auth][User] create user commit storage tx", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponse(apimodel.Code_ExecuteException)
	}

	log.Info("[Auth][User] create user", utils.RequestID(ctx), zap.String("name", req.GetName().GetValue()))
	svr.RecordHistory(userRecordEntry(ctx, req, data, types.OCreate))

	// 去除 owner 信息
	req.Owner = protobuf.NewStringValue("")
	req.Id = protobuf.NewStringValue(data.ID)
	return api.NewUserResponse(apimodel.Code_ExecuteSuccess, req)
}

// UpdateUsers 更新用户信息，仅能修改 comment 以及账户密码
func (svr *Server) UpdateUsers(ctx context.Context, reqs []*apisecurity.User) *apiservice.BatchWriteResponse {
	batchResp := api.NewAuthBatchWriteResponse(apimodel.Code_ExecuteSuccess)

	for i := range reqs {
		resp := svr.UpdateUser(ctx, reqs[i])
		api.Collect(batchResp, resp)
	}

	return batchResp
}

// UpdateUser 更新用户信息，仅能修改 comment 以及账户密码
func (svr *Server) UpdateUser(ctx context.Context, req *apisecurity.User) *apiservice.Response {
	user, err := svr.storage.GetUser(req.Id.GetValue())
	if err != nil {
		log.Error("[Auth][User] get user", utils.RequestID(ctx), zap.String("user-id", req.GetId().GetValue()), zap.Error(err))
		return api.NewUserResponse(storeapi.StoreCode2APICode(err), req)
	}
	if user == nil {
		return api.NewUserResponse(apimodel.Code_NotFoundUser, req)
	}

	data, needUpdate, err := updateUserAttribute(user, req)
	if err != nil {
		return api.NewAuthResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
	}

	if !needUpdate {
		log.Info("[Auth][User] update user data no change, no need update", utils.RequestID(ctx), zap.String("user", req.String()))
		return api.NewUserResponse(apimodel.Code_NoNeedUpdate, req)
	}

	if err := svr.storage.UpdateUser(data); err != nil {
		log.Error("[Auth][User] update user from store", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error())
	}

	log.Info("[Auth][User] update user", utils.RequestID(ctx), zap.String("name", req.GetName().GetValue()))
	svr.RecordHistory(userRecordEntry(ctx, req, user, types.OUpdate))

	return api.NewUserResponse(apimodel.Code_ExecuteSuccess, req)
}

// UpdateUserPassword 更新用户密码信息
func (svr *Server) UpdateUserPassword(ctx context.Context, req *apisecurity.ModifyUserPassword) *apiservice.Response {
	user, err := svr.storage.GetUser(req.Id.GetValue())
	if err != nil {
		log.Error("[Auth][User] get user", utils.RequestID(ctx),
			zap.String("user-id", req.Id.GetValue()), zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}
	if user == nil {
		return api.NewAuthResponse(apimodel.Code_NotFoundUser)
	}

	ignoreOrigin := authtypes.ParseUserRole(ctx) == authtypes.OwnerUserRole
	data, needUpdate, err := updateUserPasswordAttribute(ignoreOrigin, user, req)
	if err != nil {
		log.Error("[Auth][User] compute user update attribute", zap.Error(err),
			zap.String("user", req.GetId().GetValue()))
		return api.NewAuthResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
	}

	if !needUpdate {
		log.Info("[Auth][User] update user password no change, no need update",
			utils.RequestID(ctx), zap.String("user", req.GetId().GetValue()))
		return api.NewAuthResponse(apimodel.Code_NoNeedUpdate)
	}

	if err := svr.storage.UpdateUser(data); err != nil {
		log.Error("[Auth][User] update user from store", utils.RequestID(ctx),
			zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}

	log.Info("[Auth][User] update user", utils.RequestID(ctx), zap.String("user-id", req.Id.GetValue()))

	return api.NewAuthResponse(apimodel.Code_ExecuteSuccess)
}

// DeleteUsers 批量删除用户
func (svr *Server) DeleteUsers(ctx context.Context, reqs []*apisecurity.User) *apiservice.BatchWriteResponse {
	resp := api.NewAuthBatchWriteResponse(apimodel.Code_ExecuteSuccess)

	for index := range reqs {
		ret := svr.DeleteUser(ctx, reqs[index])
		api.Collect(resp, ret)
	}

	return resp
}

// DeleteUser 删除用户
// Case 1. 删除主账户，主账户不能自己删除自己
// Case 2. 删除主账户，如果主账户下还存在子账户，必须先删除子账户，才能删除主账户
// Case 3. 主账户角色下，只能删除自己创建的子账户
// Case 4. 超级账户角色下，可以删除任意账户
func (svr *Server) DeleteUser(ctx context.Context, req *apisecurity.User) *apiservice.Response {
	user, err := svr.storage.GetUser(req.Id.GetValue())
	if err != nil {
		log.Error("[Auth][User] get user from store", utils.RequestID(ctx), zap.Error(err))
		return api.NewUserResponse(storeapi.StoreCode2APICode(err), req)
	}
	if user == nil {
		return api.NewUserResponse(apimodel.Code_ExecuteSuccess, req)
	}

	if user.ID == utils.ParseOwnerID(ctx) {
		log.Error("[Auth][User] delete user forbidden, can't delete when self is owner",
			utils.RequestID(ctx), zap.String("name", req.Name.GetValue()))
		return api.NewUserResponse(apimodel.Code_NotAllowedAccess, req)
	}
	if user.Type == authtypes.OwnerUserRole {
		count, err := svr.storage.GetSubCount(user)
		if err != nil {
			log.Error("[Auth][User] get user sub-account", zap.String("owner", user.ID),
				utils.RequestID(ctx), zap.Error(err))
			return api.NewUserResponse(storeapi.StoreCode2APICode(err), req)
		}
		if count != 0 {
			log.Error("[Auth][User] delete user but some sub-account existed", zap.String("owner", user.ID))
			return api.NewUserResponse(apimodel.Code_SubAccountExisted, req)
		}
	}
	tx, err := svr.storage.StartTx()
	if err != nil {
		log.Error("[Auth][User] delete user begion storage tx", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponse(apimodel.Code_ExecuteException)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := svr.storage.DeleteUser(tx, user); err != nil {
		log.Error("[Auth][User] delete user from store", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}
	if err := svr.policySvr.PolicyHelper().CleanPrincipal(ctx, tx, authtypes.Principal{
		PrincipalID:   user.ID,
		PrincipalType: authtypes.PrincipalUser,
	}); err != nil {
		log.Error("[Auth][User] delete user from policy server", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponse(storeapi.StoreCode2APICode(err))
	}
	if err := tx.Commit(); err != nil {
		log.Error("[Auth][User] delete user commit storage tx", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponse(apimodel.Code_ExecuteException)
	}

	log.Info("[Auth][User] delete user", utils.RequestID(ctx), zap.String("name", req.Name.GetValue()))
	svr.RecordHistory(userRecordEntry(ctx, req, user, types.ODelete))

	return api.NewUserResponse(apimodel.Code_ExecuteSuccess, req)
}

// GetUsers 查询用户列表
func (svr *Server) GetUsers(ctx context.Context, filters map[string]string) *apiservice.BatchQueryResponse {
	offset, limit, _ := valid.ParseOffsetAndLimit(filters)

	total, users, err := svr.cacheMgr.User().QueryUsers(ctx, cachetypes.UserSearchArgs{
		Filters: filters,
		Offset:  offset,
		Limit:   limit,
	})
	if err != nil {
		log.Error("[Auth][User] get user from store", utils.RequestID(ctx), zap.Any("req", filters),
			zap.Error(err))
		return api.NewAuthBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}

	resp := api.NewAuthBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = protobuf.NewUInt32Value(total)
	resp.Size = protobuf.NewUInt32Value(uint32(len(users)))
	resp.Users = enhancedUsers2Api(users, user2Api)
	return resp
}

// GetUserToken 获取用户 token
func (svr *Server) GetUserToken(ctx context.Context, req *apisecurity.User) *apiservice.Response {
	var user *authtypes.User
	if req.GetId().GetValue() != "" {
		user = svr.cacheMgr.User().GetUserByID(req.GetId().GetValue())
	} else if req.GetName().GetValue() != "" {
		user = svr.cacheMgr.User().GetUserByName(req.GetName().GetValue())
	} else {
		return api.NewAuthResponse(apimodel.Code_InvalidParameter)
	}

	if user == nil {
		return api.NewUserResponse(apimodel.Code_NotFoundUser, req)
	}

	out := &apisecurity.User{
		Id:          protobuf.NewStringValue(user.ID),
		Name:        protobuf.NewStringValue(user.Name),
		AuthToken:   protobuf.NewStringValue(user.Token),
		TokenEnable: protobuf.NewBoolValue(user.TokenEnable),
	}

	return api.NewUserResponse(apimodel.Code_ExecuteSuccess, out)
}

// EnableUserToken 更新用户 token
func (svr *Server) EnableUserToken(ctx context.Context, req *apisecurity.User) *apiservice.Response {
	user, err := svr.storage.GetUser(req.GetId().GetValue())
	if err != nil {
		log.Error("[Auth][User] get user from store", utils.RequestID(ctx), zap.Error(err))
		return api.NewUserResponse(storeapi.StoreCode2APICode(err), req)
	}
	if user == nil {
		return api.NewUserResponse(apimodel.Code_NotFoundUser, req)
	}

	user.TokenEnable = req.GetTokenEnable().GetValue()

	if err := svr.storage.UpdateUser(user); err != nil {
		log.Error("[Auth][User] update user token into store", utils.RequestID(ctx), zap.Error(err))
		return api.NewAuthResponseWithMsg(storeapi.StoreCode2APICode(err), err.Error())
	}

	log.Info("[Auth][User] update user token", utils.RequestID(ctx),
		zap.String("id", req.GetId().GetValue()), zap.Bool("enable", req.GetTokenEnable().GetValue()))
	svr.RecordHistory(userRecordEntry(ctx, req, user, types.OUpdateToken))

	return api.NewUserResponse(apimodel.Code_ExecuteSuccess, req)
}

// ResetUserToken 重置用户 token
func (svr *Server) ResetUserToken(ctx context.Context, req *apisecurity.User) *apiservice.Response {
	user, err := svr.storage.GetUser(req.Id.GetValue())
	if err != nil {
		log.Error("[Auth][User] get user from store", utils.RequestID(ctx), zap.Error(err))
		return api.NewUserResponse(storeapi.StoreCode2APICode(err), req)
	}
	if user == nil {
		return api.NewUserResponse(apimodel.Code_NotFoundUser, req)
	}

	newToken, err := createUserToken(user.ID, svr.authOpt.Salt)
	if err != nil {
		log.Error("[Auth][User] update user token", utils.RequestID(ctx), zap.Error(err))
		return api.NewUserResponse(apimodel.Code_ExecuteException, req)
	}

	user.Token = newToken

	if err := svr.storage.UpdateUser(user); err != nil {
		log.Error("[Auth][User] update user token into store", utils.RequestID(ctx), zap.Error(err))
		return api.NewUserResponse(storeapi.StoreCode2APICode(err), req)
	}

	log.Info("[Auth][User] reset user token", utils.RequestID(ctx), zap.String("id", req.GetId().GetValue()))
	svr.RecordHistory(userRecordEntry(ctx, req, user, types.OUpdateToken))

	req.AuthToken = protobuf.NewStringValue(user.Token)

	return api.NewUserResponse(apimodel.Code_ExecuteSuccess, req)
}

// VerifyCredential 对 token 进行检查验证，并将 verify 过程中解析出的数据注入到 model.AcquireContext 中
// step 1. 首先对 token 进行解析，获取相关的数据信息，注入到整个的 AcquireContext 中
// step 2. 最后对 token 进行一些验证步骤的执行
// step 3. 兜底措施：如果开启了鉴权的非严格模式，则根据错误的类型，判断是否转为匿名用户进行访问
//   - 如果是访问权限控制相关模块（用户、用户组、权限策略），不得转为匿名用户
func (svr *Server) CheckCredential(authCtx *authtypes.AcquireContext) error {
	// 如果已经存在了解析出的 token 信息，则直接返回
	if _, ok := authCtx.GetAttachment(authtypes.TokenDetailInfoKey); ok {
		return nil
	}

	checkErr := func() error {
		if authCtx.GetRequestContext() == nil {
			return authtypes.ErrorTokenInvalid
		}

		authHeaders := []string{types.HeaderAuthorizationKey, "X-Polaris-Token"}
		headers := types.GetRequestHeader(authCtx.GetRequestContext())
		var authToken string
		for _, k := range authHeaders {
			if v := headers[k]; len(v) > 0 {
				authToken = v[0]
				break
			}
		}
		operator, err := svr.decodeToken(authToken)
		if err != nil {
			log.Error("[Auth][Checker] decode token", utils.RequestID(authCtx.GetRequestContext()), zap.String("token", authToken), zap.Error(err))
			return authtypes.ErrorTokenInvalid
		}

		ownerId, isOwner, err := svr.checkToken(&operator)
		if err != nil {
			log.Error("[Auth][Checker] check token", utils.RequestID(authCtx.GetRequestContext()), zap.String("token", authToken), zap.Error(err))
			return err
		}

		operator.OwnerID = ownerId
		ctx := authCtx.GetRequestContext()
		ctx = context.WithValue(ctx, types.ContextIsOwnerKey, isOwner)
		ctx = context.WithValue(ctx, types.ContextUserIDKey, operator.OperatorID)
		ctx = context.WithValue(ctx, types.ContextOwnerIDKey, ownerId)
		authCtx.SetRequestContext(ctx)
		svr.parseOperatorInfo(operator, authCtx)
		if operator.Disable {
			log.Error("[Auth][Checker] token has been set disable", utils.RequestID(authCtx.GetRequestContext()),
				zap.String("operator", operator.String()))
			return authtypes.ErrorTokenDisabled
		}
		return nil
	}()

	if checkErr != nil {
		if !canDowngradeAnonymous(authCtx, checkErr) {
			return checkErr
		}
		log.Warn("[Auth][Checker] parse operator info, downgrade to anonymous", utils.RequestID(authCtx.GetRequestContext()),
			zap.Error(checkErr))
		// 操作者信息解析失败，降级为匿名用户
		authCtx.SetAttachment(authtypes.PrincipalKey, authtypes.NewAnonymousPrincipal())
		authCtx.SetAttachment(authtypes.TokenDetailInfoKey, authapi.NewAnonymousOperatorInfo())
	}
	return nil
}

func (svr *Server) parseOperatorInfo(operator authapi.OperatorInfo, authCtx *authtypes.AcquireContext) {
	ctx := authCtx.GetRequestContext()
	if operator.IsUserToken {
		user := svr.cacheMgr.User().GetUserByID(operator.OperatorID)
		if user != nil {
			operator.Role = user.Type
			ctx = context.WithValue(ctx, types.ContextOperator, user.Name)
			ctx = context.WithValue(ctx, types.ContextUserNameKey, user.Name)
			ctx = context.WithValue(ctx, types.ContextUserRoleIDKey, user.Type)
		}
	} else {
		userGroup := svr.cacheMgr.User().GetGroup(operator.OperatorID)
		if userGroup != nil {
			ctx = context.WithValue(ctx, types.ContextOperator, userGroup.Name)
			ctx = context.WithValue(ctx, types.ContextUserNameKey, userGroup.Name)
		}
	}

	authCtx.SetAttachment(authtypes.PrincipalKey, authtypes.Principal{
		PrincipalID: operator.OperatorID,
		PrincipalType: func() authtypes.PrincipalType {
			if operator.IsUserToken {
				return authtypes.PrincipalUser
			}
			return authtypes.PrincipalGroup
		}(),
	})
	authCtx.SetAttachment(authtypes.OperatorRoleKey, operator.Role)
	authCtx.SetAttachment(authtypes.OperatorIDKey, operator.OperatorID)
	authCtx.SetAttachment(authtypes.OperatorOwnerKey, operator)
	authCtx.SetAttachment(authtypes.TokenDetailInfoKey, operator)

	authCtx.SetRequestContext(ctx)
}

func canDowngradeAnonymous(authCtx *authtypes.AcquireContext, err error) bool {
	if authCtx.GetModule() == authtypes.AuthModule || authCtx.GetModule() == authtypes.MaintainModule {
		return false
	}
	if !authCtx.IsAllowAnonymous() {
		return false
	}
	if errors.Is(err, authtypes.ErrorTokenInvalid) {
		return true
	}
	if errors.Is(err, authtypes.ErrorTokenNotExist) {
		return true
	}
	return false
}

// user 数组转为[]*apisecurity.User
func enhancedUsers2Api(users []*authtypes.User, handler User2Api) []*apisecurity.User {
	out := make([]*apisecurity.User, 0, len(users))
	for _, entry := range users {
		outUser := handler(entry)
		out = append(out, outUser)
	}

	return out
}

// model.Service 转为 api.Service
func user2Api(user *authtypes.User) *apisecurity.User {
	if user == nil {
		return nil
	}

	// note: 不包括token，token比较特殊
	out := &apisecurity.User{
		Id:          protobuf.NewStringValue(user.ID),
		Name:        protobuf.NewStringValue(user.Name),
		Source:      protobuf.NewStringValue(user.Source),
		TokenEnable: protobuf.NewBoolValue(user.TokenEnable),
		Comment:     protobuf.NewStringValue(user.Comment),
		Ctime:       protobuf.NewStringValue(commontime.Time2String(user.CreateTime)),
		Mtime:       protobuf.NewStringValue(commontime.Time2String(user.ModifyTime)),
		UserType:    protobuf.NewStringValue(authtypes.UserRoleNames[user.Type]),
		Metadata:    user.Metadata,
	}

	return out
}

// 生成用户的记录entry
func userRecordEntry(ctx context.Context, req *apisecurity.User, md *authtypes.User,
	operationType types.OperationType) *types.RecordEntry {

	marshaler := jsonpb.Marshaler{}
	detail, _ := marshaler.MarshalToString(req)

	entry := &types.RecordEntry{
		ResourceType:  types.RUser,
		ResourceName:  fmt.Sprintf("%s(%s)", md.Name, md.ID),
		OperationType: operationType,
		Operator:      utils.ParseOperator(ctx),
		Detail:        detail,
		HappenTime:    time.Now(),
	}

	return entry
}

// updateUserAttribute 更新用户属性
func updateUserAttribute(old *authtypes.User, newUser *apisecurity.User) (*authtypes.User, bool, error) {
	var needUpdate = true

	if newUser.Comment != nil && old.Comment != newUser.Comment.GetValue() {
		old.Comment = newUser.Comment.GetValue()
		needUpdate = true
	}
	if !maps.Equal(old.Metadata, newUser.Metadata) {
		old.Metadata = newUser.Metadata
		needUpdate = true

	}
	return old, needUpdate, nil
}

// updateUserAttribute 更新用户密码信息，如果用户的密码被更新
func updateUserPasswordAttribute(
	isAdmin bool, user *authtypes.User, req *apisecurity.ModifyUserPassword) (*authtypes.User, bool, error) {
	needUpdate := false

	if err := CheckPassword(req.NewPassword); err != nil {
		return nil, false, err
	}

	if !isAdmin {
		if req.GetOldPassword().GetValue() == "" {
			return nil, false, errors.New("original password is empty")
		}

		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.GetOldPassword().GetValue()))
		if err != nil {
			return nil, false, errors.New("original password match failed")
		}
	}

	if req.GetNewPassword().GetValue() != "" {
		pwd, err := bcrypt.GenerateFromPassword([]byte(req.GetNewPassword().GetValue()), bcrypt.DefaultCost)
		if err != nil {
			return nil, false, err
		}
		needUpdate = true
		user.Password = string(pwd)
	}
	return user, needUpdate, nil
}

// createUserModel 创建用户模型
func (svr *Server) createUserModel(req *apisecurity.User, role authtypes.UserRoleType) (*authtypes.User, error) {
	pwd, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword().GetValue()), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	id := utils.NewUUID()
	if req.GetId().GetValue() != "" {
		id = req.GetId().GetValue()
	}

	user := &authtypes.User{
		ID:          id,
		Name:        req.GetName().GetValue(),
		Password:    string(pwd),
		Source:      req.GetSource().GetValue(),
		Valid:       true,
		Type:        role,
		Comment:     req.GetComment().GetValue(),
		CreateTime:  time.Now(),
		ModifyTime:  time.Now(),
		TokenEnable: true,
	}

	newToken, err := createUserToken(user.ID, svr.authOpt.Salt)
	if err != nil {
		return nil, err
	}

	user.Token = newToken

	return user, nil
}
