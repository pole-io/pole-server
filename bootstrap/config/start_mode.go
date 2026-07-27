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
	// StartModeAll 启动 Control Plane 和 Console。
	StartModeAll = "all"
	// StartModeConsole 仅启动 Console。
	StartModeConsole = "console"
	// StartModeControlPlane 仅启动 Control Plane。
	StartModeControlPlane = "control-plane"
	// StartModeLimiterServer 仅启动 Limiter Server。
	StartModeLimiterServer = "limiter-server"
	// StartModeFull 同时启动 Control Plane、Limiter Server 和 Console。
	StartModeFull = "full"

	// StartModeServer 是 StartModeControlPlane 的弃用兼容别名。
	StartModeServer = "server"
)

// StartProfile 描述启动模式选中的静态模块集合。
type StartProfile struct {
	Mode          string
	ControlPlane  bool
	Console       bool
	LimiterServer bool
}

// ResolveStartMode 解析启动模式，CLI 参数优先于配置文件。
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
	case StartModeServer:
		return StartModeControlPlane, nil
	case StartModeAll, StartModeConsole, StartModeControlPlane, StartModeLimiterServer, StartModeFull:
		return mode, nil
	default:
		return "", fmt.Errorf("invalid start mode %q, expected one of: %s, %s, %s, %s, %s (or deprecated alias %s)",
			mode, StartModeAll, StartModeConsole, StartModeControlPlane, StartModeLimiterServer, StartModeFull,
			StartModeServer)
	}
}

// ResolveStartProfile 解析启动模式并返回静态模块 Profile。
func ResolveStartProfile(configMode, cliMode string) (StartProfile, error) {
	mode, err := ResolveStartMode(configMode, cliMode)
	if err != nil {
		return StartProfile{}, err
	}

	profile := StartProfile{Mode: mode}
	switch mode {
	case StartModeConsole:
		profile.Console = true
	case StartModeControlPlane:
		profile.ControlPlane = true
	case StartModeLimiterServer:
		profile.LimiterServer = true
	case StartModeAll:
		profile.ControlPlane = true
		profile.Console = true
	case StartModeFull:
		profile.ControlPlane = true
		profile.Console = true
		profile.LimiterServer = true
	}
	return profile, nil
}
