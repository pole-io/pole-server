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

package policy

import (
	"fmt"
	golog "log"

	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	cachetypes "github.com/pole-io/pole-server/apis/cache"
	policy_auth "github.com/pole-io/pole-server/plugin/access_control/auth/policy/inteceptor/auth"
	"github.com/pole-io/pole-server/plugin/access_control/auth/policy/inteceptor/paramcheck"
)

type ServerProxyFactory func(svr *Server, pre authapi.StrategyServer) (authapi.StrategyServer, error)

var (
	// serverProxyFactories authapi.StrategyServer API 代理工厂
	serverProxyFactories = map[string]ServerProxyFactory{}
)

// RegisterServerProxy .
func RegisterServerProxy(name string, factor ServerProxyFactory) {
	if _, ok := serverProxyFactories[name]; ok {
		golog.Printf("duplicate ServerProxyFactory, name(%s)", name)
		return
	}
	serverProxyFactories[name] = factor
}

func init() {
	_, nextSvr, err := BuildServer()
	if err != nil {
		panic(err)
	}
	_ = authapi.RegisterStrategyServer(nextSvr)
}

func loadInteceptors() {
	RegisterServerProxy("auth", func(svr *Server, pre authapi.StrategyServer) (authapi.StrategyServer, error) {
		return policy_auth.NewServer(pre), nil
	})
	RegisterServerProxy("paramcheck", func(svr *Server, pre authapi.StrategyServer) (authapi.StrategyServer, error) {
		return paramcheck.NewServer(pre), nil
	})
}

func BuildServer() (*Server, authapi.StrategyServer, error) {
	loadInteceptors()
	svr := &Server{}
	var nextSvr authapi.StrategyServer
	nextSvr = svr
	// 需要返回包装代理的 DiscoverServer
	order := GetChainOrder()
	for i := range order {
		factory, exist := serverProxyFactories[order[i]]
		if !exist {
			return nil, nil, fmt.Errorf("name(%s) not exist in serverProxyFactories", order[i])
		}

		proxySvr, err := factory(svr, nextSvr)
		if err != nil {
			panic(err)
		}
		nextSvr = proxySvr
	}
	return svr, nextSvr, nil
}

func GetChainOrder() []string {
	return []string{
		"auth",
		"paramcheck",
	}
}

var (
	_cacheEntries = []cachetypes.ConfigEntry{
		{
			Name: cachetypes.StrategyRuleName,
		},
		{
			Name: cachetypes.RolesName,
		},
	}
)
