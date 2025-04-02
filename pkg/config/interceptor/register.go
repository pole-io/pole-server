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

package config_chain

import (
	"github.com/pole-io/pole-server/apis/access_control/auth"
	"github.com/pole-io/pole-server/apis/store"
	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/pkg/config"
	config_auth "github.com/pole-io/pole-server/pkg/config/interceptor/auth"
	"github.com/pole-io/pole-server/pkg/config/interceptor/paramcheck"
)

func init() {
	err := config.RegisterServerProxy("paramcheck", func(cacheMgr cacheapi.CacheManager, s store.Store,
		next config.ConfigCenterServer, cfg config.Config) (config.ConfigCenterServer, error) {
		return paramcheck.New(next, cacheMgr, s, cfg), nil
	})
	if err != nil {
		panic(err)
	}

	err = config.RegisterServerProxy("auth", func(cacheMgr cacheapi.CacheManager, s store.Store,
		next config.ConfigCenterServer, cfg config.Config) (config.ConfigCenterServer, error) {
		userMgr, err := auth.GetUserServer()
		if err != nil {
			return nil, err
		}
		strategyMgr, err := auth.GetStrategyServer()
		if err != nil {
			return nil, err
		}

		return config_auth.New(next, cacheMgr, userMgr, strategyMgr), nil
	})
	if err != nil {
		panic(err)
	}
}
