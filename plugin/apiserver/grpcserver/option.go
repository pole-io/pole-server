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

package grpcserver

import (
	authcommon "github.com/pole-io/pole-server/apis/pkg/types/auth"
	commonlog "github.com/pole-io/pole-server/pkg/common/log"
)

type InitOption func(svr *BaseGrpcServer)

// WithModule set bz module
func WithModule(bz authcommon.BzModule) InitOption {
	return func(svr *BaseGrpcServer) {
		svr.bz = bz
	}
}

// WithProtocol
func WithProtocol(protocol string) InitOption {
	return func(svr *BaseGrpcServer) {
		svr.protocol = protocol
	}
}

// WithProtobufCache
func WithProtobufCache(cache Cache) InitOption {
	return func(svr *BaseGrpcServer) {
		svr.cache = cache
	}
}

// WithMessageToCacheObject
func WithMessageToCacheObject(convert MessageToCache) InitOption {
	return func(svr *BaseGrpcServer) {
		svr.convert = convert
	}
}

// WithLogger
func WithLogger(log *commonlog.Scope) InitOption {
	return func(svr *BaseGrpcServer) {
		svr.log = log
	}
}
