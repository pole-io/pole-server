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
	_ "github.com/pole-io/pole-server/pkg/config/interceptor"
	_ "github.com/pole-io/pole-server/pkg/service/interceptor"
	_ "github.com/pole-io/pole-server/plugin/access_control/auth/policy"
	_ "github.com/pole-io/pole-server/plugin/access_control/auth/user"
	_ "github.com/pole-io/pole-server/plugin/access_control/ratelimit/token"
	_ "github.com/pole-io/pole-server/plugin/cmdb/memory"
	_ "github.com/pole-io/pole-server/plugin/crypto/aes"
	_ "github.com/pole-io/pole-server/plugin/healthchecker/heartbeat"
	_ "github.com/pole-io/pole-server/plugin/observability/discoverevent/local"
	_ "github.com/pole-io/pole-server/plugin/observability/history/logger"
	_ "github.com/pole-io/pole-server/plugin/observability/statis/logger"
	_ "github.com/pole-io/pole-server/plugin/observability/statis/prometheus"
	_ "github.com/pole-io/pole-server/plugin/store/boltdb"
	_ "github.com/pole-io/pole-server/plugin/store/mysql"
)
