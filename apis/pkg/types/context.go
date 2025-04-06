/**
 * Tencent is pleased to support the open source community by making Pole available.
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

package types

import "context"

type (
	StringContext string
)

const (
	// PoleCode Pole code
	PoleCode = "X-Pole-Code"
	// PoleMessage Pole message
	PoleMessage = "X-Pole-Message"
	// PoleRequestID request_id
	PoleRequestID = "Request-Id"
)

const (
	// HeaderAuthorizationKey auth token key
	HeaderAuthorizationKey string = "Authorization"
	// HeaderIsOwnerKey is owner key
	HeaderIsOwnerKey string = "X-Is-Owner"
	// HeaderUserIDKey user id key
	HeaderUserIDKey string = "X-User-ID"
	// HeaderOwnerIDKey owner id key
	HeaderOwnerIDKey string = "X-Owner-ID"
	// HeaderUserRoleKey user role key
	HeaderUserRoleKey string = "X-Pole-User-Role"
	// HeaderRequestId request-id
	HeaderRequestId string = "request-id"
	// HeaderUserAgent user agent
	HeaderUserAgent string = "user-agent"

	// ContextAuthTokenKey auth token key
	ContextAuthTokenKey = StringContext(HeaderAuthorizationKey)
	// ContextIsOwnerKey is owner key
	ContextIsOwnerKey = StringContext(HeaderIsOwnerKey)
	// ContextUserIDKey user id key
	ContextUserIDKey = StringContext(HeaderUserIDKey)
	// ContextOwnerIDKey owner id key
	ContextOwnerIDKey = StringContext(HeaderOwnerIDKey)
	// ContextUserRoleIDKey user role key
	ContextUserRoleIDKey = StringContext(HeaderUserRoleKey)
	// ContextAuthContextKey auth context key
	ContextAuthContextKey = StringContext("X-Pole-AuthContext")
	// ContextUserNameKey users name key
	ContextUserNameKey = StringContext("X-User-Name")
	// ContextClientAddress client address key
	ContextClientAddress = StringContext("client-address")
	// ContextOpenAsyncRegis open async register key
	ContextOpenAsyncRegis = StringContext("client-asyncRegis")
	// ContextGrpcHeader grpc header key
	ContextGrpcHeader = StringContext("grpc-header")
	// ContextIsFromClient is from client
	ContextIsFromClient = StringContext("from-client")
	// ContextIsFromSystem is from Pole system
	ContextIsFromSystem = StringContext("from-system")
	// ContextOperator operator info
	ContextOperator = StringContext("operator")
	// ContextRequestHeaders request headers, save value type is map[string][]string
	ContextRequestHeaders = StringContext("request-headers")
	// ContextRequestHeaderKey request header key
	ContextRequestId = StringContext(HeaderRequestId)
	// ContextPoleToken Pole token
	ContextPoleToken = StringContext("Pole-token")
	// ContextClientIP client ip
	ContextClientIP = StringContext("client-ip")
	// ContextUserAgent user agent
	ContextUserAgent = StringContext(HeaderUserAgent)
	// ContextKeyConditions key conditions
	ContextKeyConditions = StringContext("key-conditions")
)

func AppendRequestHeader(ctx context.Context, headers map[string][]string) context.Context {
	return context.WithValue(ctx, ContextRequestHeaders, headers)
}

func GetRequestHeader(ctx context.Context) map[string][]string {
	if headers, ok := ctx.Value(ContextRequestHeaders).(map[string][]string); ok {
		return headers
	}
	return map[string][]string{}
}

func AppendContextValue(ctx context.Context, key, value any) context.Context {
	return context.WithValue(ctx, key, value)
}
