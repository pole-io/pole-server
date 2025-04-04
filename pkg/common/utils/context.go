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

package utils

import (
	"context"
	"strings"

	"github.com/pole-io/pole-server/apis/pkg/types"
)

type (
	// ContextKeyAutoCreateNamespace .
	ContextKeyAutoCreateNamespace struct{}
	// ContextKeyAutoCreateService .
	ContextKeyAutoCreateService struct{}
	// ContextKeyCompatible .
	ContextKeyCompatible struct{}
)

const (
	// PolarisCode polaris code
	PolarisCode = "X-Polaris-Code"
	// PolarisMessage polaris message
	PolarisMessage = "X-Polaris-Message"
	// PolarisRequestID request_id
	PolarisRequestID = "Request-Id"
)

var (
	// LocalHost local host
	LocalHost = "127.0.0.1"
	// LocalPort default listen port
	LocalPort = 8091
	// ConfDir default config dir
	ConfDir = "conf/"
)

type (
	// localhostCtx is a context key that carries localhost info.
	localhostCtx struct{}
	// ContextAPIServerSlot
	ContextAPIServerSlot struct{}
	// WatchTimeoutCtx .
	WatchTimeoutCtx struct{}
)

// WithLocalhost 存储localhost
func WithLocalhost(ctx context.Context, localhost string) context.Context {
	return context.WithValue(ctx, localhostCtx{}, localhost)
}

// ValueLocalhost 获取localhost
func ValueLocalhost(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	value, ok := ctx.Value(localhostCtx{}).(string)
	if !ok {
		return ""
	}

	return value
}

// ParseRequestID 从ctx中获取Request-ID
func ParseRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	rid, _ := ctx.Value(types.ContextRequestId).(string)
	return rid
}

// ParseClientAddress 从ctx中获取客户端地址
func ParseClientAddress(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	rid, _ := ctx.Value(types.ContextClientAddress).(string)
	return rid
}

// ParseClientIP .
func ParseClientIP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	rid, _ := ctx.Value(types.ContextClientAddress).(string)
	if strings.Contains(rid, ":") {
		return strings.Split(rid, ":")[0]
	}
	return rid
}

// ParseAuthToken 从ctx中获取token
func ParseAuthToken(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	token, _ := ctx.Value(types.ContextAuthTokenKey).(string)
	return token
}

// ParseIsOwner 从ctx中获取token
func ParseIsOwner(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	isOwner, _ := ctx.Value(types.ContextIsOwnerKey).(bool)
	return isOwner
}

// ParseUserID 从ctx中解析用户ID
func ParseUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	userID, _ := ctx.Value(types.ContextUserIDKey).(string)
	return userID
}

// ParseUserName 从ctx解析用户名称
func ParseUserName(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	userName, _ := ctx.Value(types.ContextUserNameKey).(string)
	if userName == "" {
		return ParseOperator(ctx)
	}
	return userName
}

// ParseOwnerID 从ctx解析Owner ID
func ParseOwnerID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	ownerID, _ := ctx.Value(types.ContextOwnerIDKey).(string)
	return ownerID
}

// ParseToken 从ctx中获取token
func ParseToken(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	token, _ := ctx.Value(types.ContextPoleToken).(string)
	return token
}

// ParseOperator 从ctx中获取operator
func ParseOperator(ctx context.Context) string {
	defaultOperator := "Pole"
	if ctx == nil {
		return defaultOperator
	}

	if operator, _ := ctx.Value(types.ContextOperator).(string); operator != "" {
		return operator
	}

	return defaultOperator
}
