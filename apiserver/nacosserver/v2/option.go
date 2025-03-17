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

package v2

import (
	"github.com/GovernSea/sergo-server/auth"
	connlimit "github.com/GovernSea/sergo-server/common/conn/limit"
	"github.com/GovernSea/sergo-server/common/secure"
	"github.com/GovernSea/sergo-server/config"
	"github.com/GovernSea/sergo-server/namespace"
	"github.com/GovernSea/sergo-server/service"
	"github.com/GovernSea/sergo-server/service/healthcheck"
)

type option func(svr *NacosV2Server)

func WithTLS(tlsInfo *secure.TLSInfo) option {
	return func(svr *NacosV2Server) {
		svr.tlsInfo = tlsInfo
	}
}

func WithConnLimitConfig(connLimitConfig *connlimit.Config) option {
	return func(svr *NacosV2Server) {
		svr.connLimitConfig = connLimitConfig
	}
}

func WithNamespaceSvr(namespaceSvr namespace.NamespaceOperateServer) option {
	return func(svr *NacosV2Server) {
		svr.discoverOpt.NamespaceSvr = namespaceSvr
		svr.configOpt.NamespaceSvr = namespaceSvr
	}
}

func WithDiscoverSvr(discoverSvr service.DiscoverServer, originDiscoverSvr service.DiscoverServer,
	healthSvr *healthcheck.Server) option {
	return func(svr *NacosV2Server) {
		svr.discoverOpt.DiscoverSvr = discoverSvr
		svr.discoverOpt.OriginDiscoverSvr = originDiscoverSvr
		svr.discoverOpt.HealthSvr = healthSvr
	}
}

func WithConfigSvr(configSvr config.ConfigCenterServer, originSvr config.ConfigCenterServer) option {
	return func(svr *NacosV2Server) {
		svr.configOpt.ConfigSvr = configSvr
		svr.configOpt.OriginConfigSvr = originSvr
	}
}

// WithAuthSvr 设置鉴权 Server
func WithAuthSvr(userSvr auth.UserServer) option {
	return func(svr *NacosV2Server) {
		svr.discoverOpt.UserSvr = userSvr
		svr.configOpt.UserSvr = userSvr
	}
}
