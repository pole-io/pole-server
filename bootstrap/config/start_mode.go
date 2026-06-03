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
	"fmt"
	"strings"
)

const (
	// StartModeAll starts server and console together.
	StartModeAll = "all"
	// StartModeServer starts only the control-plane server.
	StartModeServer = "server"
	// StartModeConsole starts only the console process.
	StartModeConsole = "console"
)

// ResolveStartMode resolves startup mode, with CLI override taking precedence over config.
func ResolveStartMode(configMode, cliMode string) (string, error) {
	mode := strings.TrimSpace(cliMode)
	if mode == "" {
		mode = strings.TrimSpace(configMode)
	}
	if mode == "" {
		mode = StartModeAll
	}
	mode = strings.ToLower(mode)

	switch mode {
	case StartModeAll, StartModeServer, StartModeConsole:
		return mode, nil
	default:
		return "", fmt.Errorf("invalid start mode %q, expected one of: %s, %s, %s",
			mode, StartModeAll, StartModeServer, StartModeConsole)
	}
}
