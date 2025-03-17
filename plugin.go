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
	_ "github.com/GovernSea/sergo-server/admin/interceptor"
	_ "github.com/GovernSea/sergo-server/apiserver/eurekaserver"
	_ "github.com/GovernSea/sergo-server/apiserver/grpcserver/config"
	_ "github.com/GovernSea/sergo-server/apiserver/grpcserver/discover"
	_ "github.com/GovernSea/sergo-server/apiserver/httpserver"
	_ "github.com/GovernSea/sergo-server/apiserver/l5pbserver"
	_ "github.com/GovernSea/sergo-server/apiserver/nacosserver"
	_ "github.com/GovernSea/sergo-server/apiserver/xdsserverv3"
	_ "github.com/GovernSea/sergo-server/auth/policy"
	_ "github.com/GovernSea/sergo-server/auth/user"
	_ "github.com/GovernSea/sergo-server/cache"
	_ "github.com/GovernSea/sergo-server/cache/auth"
	_ "github.com/GovernSea/sergo-server/cache/client"
	_ "github.com/GovernSea/sergo-server/cache/config"
	_ "github.com/GovernSea/sergo-server/cache/namespace"
	_ "github.com/GovernSea/sergo-server/cache/service"
	_ "github.com/GovernSea/sergo-server/config/interceptor"
	_ "github.com/GovernSea/sergo-server/namespace/interceptor"
	_ "github.com/GovernSea/sergo-server/plugin/cmdb/memory"
	_ "github.com/GovernSea/sergo-server/plugin/crypto/aes"
	_ "github.com/GovernSea/sergo-server/plugin/discoverevent/local"
	_ "github.com/GovernSea/sergo-server/plugin/healthchecker/leader"
	_ "github.com/GovernSea/sergo-server/plugin/healthchecker/memory"
	_ "github.com/GovernSea/sergo-server/plugin/healthchecker/redis"
	_ "github.com/GovernSea/sergo-server/plugin/history/logger"
	_ "github.com/GovernSea/sergo-server/plugin/password"
	_ "github.com/GovernSea/sergo-server/plugin/ratelimit/token"
	_ "github.com/GovernSea/sergo-server/plugin/statis/logger"
	_ "github.com/GovernSea/sergo-server/plugin/statis/prometheus"
	_ "github.com/GovernSea/sergo-server/plugin/whitelist"
	_ "github.com/GovernSea/sergo-server/service/interceptor"
	_ "github.com/GovernSea/sergo-server/store/boltdb"
	_ "github.com/GovernSea/sergo-server/store/mysql"
)
