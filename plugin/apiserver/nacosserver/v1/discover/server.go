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

package discover

import (
	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	"github.com/pole-io/pole-server/pkg/namespace"
	"github.com/pole-io/pole-server/pkg/service"
	"github.com/pole-io/pole-server/pkg/service/healthcheck"
	"github.com/pole-io/pole-server/plugin/apiserver/nacosserver/core"
	"github.com/pole-io/pole-server/plugin/apiserver/nacosserver/v2/remote"
)

type ServerOption struct {
	// nacos
	ConnectionManager *remote.ConnectionManager
	Store             *core.NacosDataStorage

	// polaris
	UserSvr           authapi.UserServer
	NamespaceSvr      namespace.NamespaceOperateServer
	DiscoverSvr       service.DiscoverServer
	OriginDiscoverSvr service.DiscoverServer
	HealthSvr         *healthcheck.Server
}

type DiscoverServer struct {
	pushCenter core.PushCenter
	store      *core.NacosDataStorage

	userSvr           authapi.UserServer
	checkerSvr        authapi.StrategyServer
	namespaceSvr      namespace.NamespaceOperateServer
	discoverSvr       service.DiscoverServer
	originDiscoverSvr service.DiscoverServer
	healthSvr         *healthcheck.Server
}

func (h *DiscoverServer) Initialize(opt *ServerOption) error {
	udpPush, err := NewUDPPushCenter(opt.Store)
	if err != nil {
		return err
	}
	h.pushCenter = udpPush
	h.store = opt.Store
	h.namespaceSvr = opt.NamespaceSvr
	h.discoverSvr = opt.DiscoverSvr
	h.originDiscoverSvr = opt.OriginDiscoverSvr
	h.healthSvr = opt.HealthSvr
	h.userSvr = opt.UserSvr
	return nil
}
