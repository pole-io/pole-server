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

	"go.uber.org/zap"

	apiconfig "github.com/polarismesh/specification/source/go/api/v1/config_manage"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/access_control/auth"
	"github.com/pole-io/pole-server/apis/crypto"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/cache"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
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

func (s *Server) TestCheckClientConfigFile(ctx context.Context, files []*apiconfig.ClientConfigFileInfo,
	compartor CompareFunction) (*apiconfig.ConfigClientResponse, bool) {
	if len(files) == 0 {
		return api.NewConfigClientResponse(apimodel.Code_InvalidWatchConfigFileFormat, nil), false
	}
	for _, configFile := range files {
		namespace := configFile.GetNamespace().GetValue()
		group := configFile.GetGroup().GetValue()
		fileName := configFile.GetFileName().GetValue()

		if namespace == "" || group == "" || fileName == "" {
			return api.NewConfigClientResponseWithInfo(apimodel.Code_BadRequest,
				"namespace & group & fileName can not be empty"), false
		}
		// 从缓存中获取最新的配置文件信息
		release := s.fileCache.GetActiveRelease(namespace, group, fileName)
		if release != nil && compartor(configFile, release) {
			ret := &apiconfig.ClientConfigFileInfo{
				Namespace: protobuf.NewStringValue(namespace),
				Group:     protobuf.NewStringValue(group),
				FileName:  protobuf.NewStringValue(fileName),
				Version:   protobuf.NewUInt64Value(release.Version),
				Md5:       protobuf.NewStringValue(release.Md5),
			}
			return api.NewConfigClientResponse(apimodel.Code_ExecuteSuccess, ret), false
		}
	}
	return api.NewConfigClientResponse(apimodel.Code_DataNoChange, nil), true
}

func TestCompareByVersion(clientInfo *apiconfig.ClientConfigFileInfo, file *conftypes.ConfigFileRelease) bool {
	return clientInfo.GetVersion().GetValue() < file.Version
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
