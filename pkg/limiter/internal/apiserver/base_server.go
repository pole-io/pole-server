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

package apiserver

import (
	"context"
	"fmt"
	"math"
)

// Config API服务器配置 配置文件
type Config struct {
	Name   string                 `yaml:"name"`
	Option map[string]interface{} `yaml:"option"`
}

// APIServer 是已经启动的 Limiter 协议服务。
type APIServer interface {
	Done() <-chan error
	Stop(context.Context) error
	GetProtocol() string
	GetPort() uint32
}

// ParseListenOption 解析通用 listener 配置。
func ParseListenOption(option map[string]interface{}) (string, uint32, error) {
	ip, _ := option["ip"].(string)
	if ip == "" {
		return "", 0, fmt.Errorf("listener ip is required")
	}
	port, err := parsePort(option["port"])
	if err != nil {
		return "", 0, err
	}
	return ip, port, nil
}

func parsePort(value interface{}) (uint32, error) {
	if value == nil {
		return 0, fmt.Errorf("listener port is required")
	}
	var port uint64
	valid := true
	switch typed := value.(type) {
	case int:
		if typed >= 0 {
			port = uint64(typed)
		} else {
			valid = false
		}
	case int64:
		if typed >= 0 {
			port = uint64(typed)
		} else {
			valid = false
		}
	case uint:
		port = uint64(typed)
	case uint32:
		port = uint64(typed)
	case uint64:
		port = typed
	case float64:
		if typed >= 0 && typed == math.Trunc(typed) {
			port = uint64(typed)
		} else {
			valid = false
		}
	default:
		valid = false
	}
	if !valid || port > math.MaxUint16 {
		return 0, fmt.Errorf("listener port must be between 0 and %d", math.MaxUint16)
	}
	return uint32(port), nil
}
