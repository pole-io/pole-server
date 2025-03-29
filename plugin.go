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

package main

import (
	_ "github.com/pole-io/pole-server/admin/interceptor"
	_ "github.com/pole-io/pole-server/apiserver/eurekaserver"
	_ "github.com/pole-io/pole-server/apiserver/grpcserver/config"
	_ "github.com/pole-io/pole-server/apiserver/grpcserver/discover"
	_ "github.com/pole-io/pole-server/apiserver/httpserver"
	_ "github.com/pole-io/pole-server/apiserver/l5pbserver"
	_ "github.com/pole-io/pole-server/apiserver/nacosserver"
	_ "github.com/pole-io/pole-server/apiserver/xdsserverv3"
	_ "github.com/pole-io/pole-server/auth/policy"
	_ "github.com/pole-io/pole-server/auth/user"
	_ "github.com/pole-io/pole-server/cache"
	_ "github.com/pole-io/pole-server/cache/auth"
	_ "github.com/pole-io/pole-server/cache/client"
	_ "github.com/pole-io/pole-server/cache/config"
	_ "github.com/pole-io/pole-server/cache/namespace"
	_ "github.com/pole-io/pole-server/cache/service"
	_ "github.com/pole-io/pole-server/config/interceptor"
	_ "github.com/pole-io/pole-server/namespace/interceptor"
	_ "github.com/pole-io/pole-server/plugin/cmdb/memory"
	_ "github.com/pole-io/pole-server/plugin/crypto/aes"
	_ "github.com/pole-io/pole-server/plugin/discoverevent/local"
	_ "github.com/pole-io/pole-server/plugin/healthchecker/p2p"
	_ "github.com/pole-io/pole-server/plugin/healthchecker/redis"
	_ "github.com/pole-io/pole-server/plugin/history/logger"
	_ "github.com/pole-io/pole-server/plugin/password"
	_ "github.com/pole-io/pole-server/plugin/ratelimit/token"
	_ "github.com/pole-io/pole-server/plugin/statis/logger"
	_ "github.com/pole-io/pole-server/plugin/statis/prometheus"
	_ "github.com/pole-io/pole-server/plugin/whitelist"
	_ "github.com/pole-io/pole-server/service/interceptor"
	_ "github.com/pole-io/pole-server/store/boltdb"
	_ "github.com/pole-io/pole-server/store/mysql"
)
