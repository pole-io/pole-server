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

package eurekaserver

import (
	_ "github.com/GovernSea/sergo-server/plugin/cmdb/memory"
	_ "github.com/GovernSea/sergo-server/plugin/discoverevent/local"
	_ "github.com/GovernSea/sergo-server/plugin/healthchecker/p2p"
	_ "github.com/GovernSea/sergo-server/plugin/healthchecker/redis"
	_ "github.com/GovernSea/sergo-server/plugin/history/logger"
	_ "github.com/GovernSea/sergo-server/plugin/password"
	_ "github.com/GovernSea/sergo-server/plugin/ratelimit/token"
	_ "github.com/GovernSea/sergo-server/plugin/statis/logger"
	_ "github.com/GovernSea/sergo-server/plugin/statis/prometheus"
	testsuit "github.com/GovernSea/sergo-server/test/suit"
)

type EurekaTestSuit struct {
	testsuit.DiscoverTestSuit
}

func newEurekaTestSuit() *EurekaTestSuit {
	return &EurekaTestSuit{}
}
