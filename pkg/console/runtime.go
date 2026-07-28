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

package console

import (
	"fmt"

	"github.com/gin-gonic/gin"
	consoleconfig "github.com/pole-io/pole-server/pkg/console/config"
	"github.com/pole-io/pole-server/pkg/console/internal/common/log"
	store "github.com/pole-io/pole-server/pkg/console/internal/observer"
)

// initialize configures Console-owned runtime dependencies.
func initialize(config *consoleconfig.Config) error {
	err := config.Logger.SetOutputLevel(log.DefaultScopeName, config.Logger.Level)
	if err != nil {
		return fmt.Errorf("set console log level: %w", err)
	}

	config.Logger.SetStackTraceLevel(log.DefaultScopeName, "none")
	config.Logger.SetLogCallers(log.DefaultScopeName, true)
	err = log.Configure(&config.Logger)
	if err != nil {
		return fmt.Errorf("configure console logger: %w", err)
	}

	store.SetStoreConfig(&config.Store)
	if _, err := store.GetStore(); err != nil {
		return fmt.Errorf("initialize console store: %w", err)
	}
	return nil
}

func setMode(config *consoleconfig.Config) {
	// 判断模式
	mode := config.WebServer.Mode
	if mode == "release" || mode == "" {
		gin.SetMode(gin.ReleaseMode)
	}
}
