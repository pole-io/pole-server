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

package testsuit

import (
	_ "github.com/pole-io/pole-server/auth/policy"
	_ "github.com/pole-io/pole-server/auth/user"
	_ "github.com/pole-io/pole-server/config/interceptor"
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
	_ "github.com/pole-io/pole-server/service/interceptor"
	_ "github.com/pole-io/pole-server/store/boltdb"
)
