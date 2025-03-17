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

package config

import (
	"github.com/GovernSea/sergo-server/apiserver/nacosserver/core"
	nacospb "github.com/GovernSea/sergo-server/apiserver/nacosserver/v2/pb"
	"github.com/GovernSea/sergo-server/apiserver/nacosserver/v2/remote"
	"github.com/GovernSea/sergo-server/auth"
	cachetypes "github.com/GovernSea/sergo-server/cache/api"
	"github.com/GovernSea/sergo-server/config"
	"github.com/GovernSea/sergo-server/namespace"
)

type ServerOption struct {
	// nacos
	ConnectionManager *remote.ConnectionManager
	Store             *core.NacosDataStorage

	// polaris
	UserSvr         auth.UserServer
	CheckerSvr      auth.StrategyServer
	NamespaceSvr    namespace.NamespaceOperateServer
	ConfigSvr       config.ConfigCenterServer
	OriginConfigSvr config.ConfigCenterServer
}

type ConfigServer struct {
	connMgr             *remote.ConnectionManager
	connectionClientMgr *ConnectionClientManager

	userSvr         auth.UserServer
	checkerSvr      auth.StrategyServer
	namespaceSvr    namespace.NamespaceOperateServer
	configSvr       config.ConfigCenterServer
	originConfigSvr config.ConfigCenterServer
	cacheSvr        cachetypes.CacheManager
	handleRegistry  map[string]*remote.RequestHandlerWarrper
}

func (h *ConfigServer) Initialize(opt *ServerOption) error {
	var err error
	h.userSvr = opt.UserSvr
	h.checkerSvr = opt.CheckerSvr
	h.namespaceSvr = opt.NamespaceSvr
	h.configSvr = opt.ConfigSvr
	h.originConfigSvr = opt.OriginConfigSvr
	h.cacheSvr = opt.Store.Cache()
	h.handleRegistry = make(map[string]*remote.RequestHandlerWarrper)
	h.connMgr = opt.ConnectionManager
	h.connectionClientMgr, err = NewConnectionClientManager(h.originConfigSvr.(*config.Server))
	if err != nil {
		return err
	}
	h.initGRPCHandlers()
	return nil
}

func (h *ConfigServer) ListGRPCHandlers() map[string]*remote.RequestHandlerWarrper {
	return h.handleRegistry
}

func (h *ConfigServer) initGRPCHandlers() {
	h.handleRegistry = map[string]*remote.RequestHandlerWarrper{
		// Request
		nacospb.TypeConfigPublishRequest: {
			Handler: h.handlePublishConfigRequest,
			PayloadBuilder: func() nacospb.CustomerPayload {
				return nacospb.NewConfigPublishRequest()
			},
		},
		nacospb.TypeConfigQueryRequest: {
			Handler: h.handleGetConfigRequest,
			PayloadBuilder: func() nacospb.CustomerPayload {
				return nacospb.NewConfigQueryRequest()
			},
		},
		nacospb.TypeConfigRemoveRequest: {
			Handler: h.handleDeleteConfigRequest,
			PayloadBuilder: func() nacospb.CustomerPayload {
				return nacospb.NewConfigRemoveRequest()
			},
		},
		nacospb.TypeConfigBatchListenRequest: {
			Handler: h.handleWatchConfigRequest,
			PayloadBuilder: func() nacospb.CustomerPayload {
				return nacospb.NewConfigBatchListenRequest()
			},
		},
		// RequestBiStream
		nacospb.TypeConfigChangeNotifyResponse: {
			PayloadBuilder: func() nacospb.CustomerPayload {
				return &nacospb.ConfigChangeNotifyResponse{}
			},
		},
	}
}
