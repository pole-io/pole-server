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

package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/common/log"
)

type User struct {
	ID          string `json:"id"`
	AuthToken   string `json:"auth_token"`
	TokenEnable bool   `json:"token_enable"`
}

type GetUserTokenResponse struct {
	Code int    `json:"code"`
	Info string `json:"info"`
	User *User  `json:"data"`
}

// checkAuthoration 检查访问 token 是否合法
func checkAuthoration(ctx *gin.Context, conf *bootstrap.Config) bool {
	userId := ctx.Request.Header.Get("x-pole-user")
	accessToken := ctx.Request.Header.Get("Authorization")
	if userId == "" {
		log.Error("denied request to server because user-id is empty")
		return false
	}
	if accessToken == "" {
		log.Error("denied request to server because token is empty")
		return false
	}
	return true
}
