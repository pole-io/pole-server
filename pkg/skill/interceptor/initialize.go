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

package interceptor

import (
	cacheapi "github.com/pole-io/pole-server/apis/cache"
	accessauth "github.com/pole-io/pole-server/apis/access_control/auth"
	"github.com/pole-io/pole-server/pkg/skill"
	"github.com/pole-io/pole-server/pkg/skill/interceptor/auth"
	"github.com/pole-io/pole-server/pkg/skill/interceptor/paramcheck"
)

// Initialize 初始化拦截器链
// 装饰器模式: origin -> paramcheck -> auth -> final
func Initialize(originSvr skill.SkillServer,
	userSvr accessauth.UserServer,
	policySvr accessauth.StrategyServer,
	cacheMgr cacheapi.CacheManager,
	options map[string]any) skill.SkillServer {

	// 构建 interceptor 链
	// 顺序: paramcheck(最外层) -> auth -> origin(最内层)

	// 1. 创建带认证的 server (内层)
	authSvr := auth.NewServer(originSvr, userSvr, policySvr, cacheMgr)

	// 2. 创建带参数校验的 server (外层)
	paramcheckSvr := paramcheck.NewServer(authSvr)

	return paramcheckSvr
}
