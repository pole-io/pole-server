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
	"context"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/access_control/auth"
	"github.com/pole-io/pole-server/apis/crypto"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/cache"
	"github.com/pole-io/pole-server/pkg/namespace"
)

// Initialize 初始化配置中心模块
func TestInitialize(ctx context.Context, config Config, s store.Store, cacheMgn *cache.CacheManager,
	namespaceOperator namespace.NamespaceOperateServer, userMgn auth.UserServer,
	strategyMgn auth.StrategyServer) (ConfigCenterServer, *Server, error) {
	mockServer := &Server{
		initialized: true,
	}

	log.Info("Config.TestInitialize", zap.Any("entries", testConfigCacheEntries))
	_ = cacheMgn.OpenResourceCache(testConfigCacheEntries...)
	if err := mockServer.initialize(ctx, config, s, namespaceOperator, cacheMgn); err != nil {
		return nil, nil, err
	}

	var proxySvr ConfigCenterServer
	proxySvr = mockServer
	// 需要返回包装代理的 ConfigCenterServer
	order := config.Interceptors
	for i := range order {
		factory, exist := serverProxyFactories[order[i]]
		if !exist {
			return nil, nil, fmt.Errorf("name(%s) not exist in serverProxyFactories", order[i])
		}

		tmpSvr, err := factory(cacheMgn, s, proxySvr, config)
		if err != nil {
			return nil, nil, err
		}
		proxySvr = tmpSvr
	}
	return proxySvr, mockServer, nil
}

func (s *Server) TestCheckClientConfigFile(ctx context.Context, files []*apiconfig.ConfigFile,
	compartor CompareFunction) (*apiconfig.ConfigDiscoverResponse, bool) {
	if len(files) == 0 {
		return &apiconfig.ConfigDiscoverResponse{
			Code: uint32(apimodel.Code_BadRequest),
			Info: apimodel.Code_BadRequest.String(),
			Type: apiconfig.ConfigDiscoverResponse_UNKNOWN,
		}, false
	}
	for _, configFile := range files {
		namespace := configFile.GetNamespace()
		group := configFile.GetGroup()
		fileName := configFile.GetName()

		if namespace == "" || group == "" || fileName == "" {
			return &apiconfig.ConfigDiscoverResponse{
				Code: uint32(apimodel.Code_BadRequest),
				Info: "namespace & group & fileName can not be empty",
				Type: apiconfig.ConfigDiscoverResponse_UNKNOWN,
			}, false
		}
		// 从缓存中获取最新的配置文件信息
		release := s.fileCache.GetActiveRelease(namespace, group, fileName)
		if release != nil && compartor(configFile, release) {
			return &apiconfig.ConfigDiscoverResponse{
				Code: uint32(apimodel.Code_ExecuteSuccess),
				Info: apimodel.Code_ExecuteSuccess.String(),
				Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
				File: &apiconfig.ConfigFileRelease{
					Namespace: namespace,
					Group:     group,
					Name:      fileName,
					Version:   release.Version,
					Md5:       release.Md5,
				},
			}, false
		}
	}
	return &apiconfig.ConfigDiscoverResponse{
		Code: uint32(apimodel.Code_DataNoChange),
		Info: apimodel.Code_DataNoChange.String(),
		Type: apiconfig.ConfigDiscoverResponse_UNKNOWN,
	}, true
}

func TestCompareByVersion(clientInfo *apiconfig.ConfigFile, file *conftypes.ConfigFileRelease) bool {
	clientVersion, err := strconv.ParseUint(clientInfo.GetId(), 10, 64)
	if err != nil {
		return false
	}
	return clientVersion < file.Version
}

// TestDecryptConfigFile 解密配置文件
func (s *Server) TestDecryptConfigFile(ctx context.Context, configFile *conftypes.ConfigFile) (err error) {
	for i := range s.chains.chains {
		chain := s.chains.chains[i]
		if val, ok := chain.(*CryptoConfigFileChain); ok {
			if _, err := val.AfterGetFile(ctx, configFile); err != nil {
				return err
			}
		}
	}
	return nil
}

// TestEncryptConfigFile 解密配置文件
func (s *Server) TestEncryptConfigFile(ctx context.Context,
	configFile *conftypes.ConfigFile, algorithm string, dataKey string) error {
	for i := range s.chains.chains {
		chain := s.chains.chains[i]
		if val, ok := chain.(*CryptoConfigFileChain); ok {
			return val.encryptConfigFile(ctx, configFile, algorithm, dataKey)
		}
	}
	return nil
}

// TestMockStore
func (s *Server) TestMockStore(ms store.Store) {
	s.storage = ms
}

// TestMockCryptoManager 获取加密管理
func (s *Server) TestMockCryptoManager(mgr crypto.CryptoManager) {
	s.cryptoManager = mgr
}
