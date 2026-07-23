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
	"encoding/json"
	"errors"

	"golang.org/x/crypto/bcrypt"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	cachetypes "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/observability/history"
	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// AuthConfig 鉴权配置
type AuthConfig struct {
	// Salt 相关密码、token加密的salt
	Salt string `json:"salt" xml:"salt"`
}

// Verify 检查配置是否合法
func (cfg *AuthConfig) Verify() error {
	k := len(cfg.Salt)
	switch k {
	case 16, 24, 32:
		break
	default:
		return errors.New("[Auth][Config] salt len must 16 | 24 | 32")
	}

	return nil
}

// DefaultUserConfig 返回一个默认的鉴权配置
func DefaultUserConfig() *AuthConfig {
	return &AuthConfig{
		// Salt token 加密 key
		Salt: "polarismesh@2021",
	}
}

type Server struct {
	authOpt   *AuthConfig
	storage   store.Store
	policySvr authapi.StrategyServer
	cacheMgr  cachetypes.CacheManager
	helper    authapi.UserHelper
}

// Name of the user operator plugin
func (svr *Server) Name() string {
	return authapi.DefaultUserMgnPluginName
}

func (svr *Server) Initialize(authOpt *authapi.Config, storage store.Store, policySvr authapi.StrategyServer,
	cacheMgr cachetypes.CacheManager) error {
	svr.cacheMgr = cacheMgr
	svr.storage = storage
	svr.policySvr = policySvr
	if err := svr.parseOptions(authOpt); err != nil {
		return err
	}

	_ = cacheMgr.OpenResourceCache(cachetypes.ConfigEntry{
		Name: cachetypes.UsersName,
	})
	svr.helper = &DefaultUserHelper{svr: svr}
	return nil
}

func (svr *Server) parseOptions(options *authapi.Config) error {
	// 新版本鉴权策略配置均从auth.Option中迁移至auth.user.option及auth.strategy.option中
	var (
		userContentBytes []byte
		authContentBytes []byte
		err              error
	)

	cfg := DefaultUserConfig()

	// 一旦设置了auth.user.option或auth.strategy.option，将不会继续读取auth.option
	if len(options.User.Option) > 0 {
		// 判断auth.option是否还有值，有则不兼容
		if len(options.Option) > 0 {
			log.Warn("auth.user.option or auth.strategy.option has set, auth.option will ignore")
		}
		userContentBytes, err = json.Marshal(options.User.Option)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(userContentBytes, cfg); err != nil {
			return err
		}
	} else {
		log.Warn("[Auth][Checker] auth.option has deprecated, use auth.user.option and auth.strategy.option instead.")
		authContentBytes, err = json.Marshal(options.Option)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(authContentBytes, cfg); err != nil {
			return err
		}
	}
	if err := cfg.Verify(); err != nil {
		return err
	}
	svr.authOpt = cfg
	return nil
}

// Login 登录动作
func (svr *Server) Login(req *apisecurity.LoginRequest) *apimodel.Response {
	username := req.GetName()
	user := svr.cacheMgr.User().GetUserByName(username)
	if user == nil {
		return api.NewAuthResponse(apimodel.Code_NotFoundUser)
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.GetPassword()))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return api.NewAuthResponseWithMsg(
				apimodel.Code_NotAllowedAccess, authtypes.ErrorWrongUsernameOrPassword.Error())
		}
		return api.NewAuthResponseWithMsg(apimodel.Code_ExecuteException, authtypes.ErrorWrongUsernameOrPassword.Error())
	}

	token, resp := svr.loginToken(user)
	if resp != nil {
		return resp
	}

	return api.NewLoginResponse(apimodel.Code_ExecuteSuccess, &apisecurity.LoginResponse{
		UserId: user.ID,
		Token:  token,
		Name:   user.Name,
		Role: resolveLoginRole(
			user,
			svr.cacheMgr.Role().GetPrincipalRoles,
			svr.cacheMgr.User().GetUserLinkGroupIds,
		),
	})
}

func resolveLoginRole(user *authtypes.User,
	rolesFor func(authtypes.Principal) []*authtypes.Role,
	groupsFor func(string) []string) string {
	if user == nil {
		return ""
	}
	if user.Type == authtypes.OwnerUserRole {
		return authtypes.UserRoleNames[user.Type]
	}
	hasAdminRole := func(principal authtypes.Principal) bool {
		for _, role := range rolesFor(principal) {
			if role != nil && role.ID == authtypes.SystemRoleAdminID && role.Type == authtypes.RoleTypeSystem {
				return true
			}
		}
		return false
	}
	if hasAdminRole(authtypes.Principal{PrincipalID: user.ID, PrincipalType: authtypes.PrincipalUser}) {
		return "admin"
	}
	for _, groupID := range groupsFor(user.ID) {
		if hasAdminRole(authtypes.Principal{PrincipalID: groupID, PrincipalType: authtypes.PrincipalGroup}) {
			return "admin"
		}
	}
	return authtypes.UserRoleNames[user.Type]
}

func (svr *Server) loginToken(user *authtypes.User) (string, *apimodel.Response) {
	if svr.isUserTokenUsable(user.ID, user.Token) {
		return user.Token, nil
	}

	token, err := createUserToken(user.ID, svr.authOpt.Salt)
	if err != nil {
		return "", api.NewAuthResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
	}

	nextUser := *user
	nextUser.Token = token
	if err := svr.storage.UpdateUser(&nextUser); err != nil {
		return "", api.NewAuthResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
	}
	user.Token = token
	_ = svr.cacheMgr.User().Update()
	return token, nil
}

func (svr *Server) isUserTokenUsable(userID string, token string) bool {
	operator, err := svr.decodeToken(token)
	if err != nil {
		return false
	}
	return operator.IsUserToken && operator.OperatorID == userID
}

// RecordHistory Server对外提供history插件的简单封装
func (svr *Server) RecordHistory(entry *types.RecordEntry) {
	history.GetHistory().Record(entry)
}

func (svr *Server) GetUserHelper() authapi.UserHelper {
	return svr.helper
}
