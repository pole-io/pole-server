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
	"github.com/golang/protobuf/proto"

	apiconfig "github.com/polarismesh/specification/source/go/api/v1/config_manage"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
)

// ConfigReleaseArr 发布规则数组定义
type ConfigReleaseArr []*apiservice.ConfigRelease

// Reset reset initialization
func (m *ConfigReleaseArr) Reset() { *m = ConfigReleaseArr{} }

// String return string
func (m *ConfigReleaseArr) String() string { return proto.CompactTextString(m) }

// ProtoMessage return proto message
func (*ConfigReleaseArr) ProtoMessage() {}

// ConfigFileArr 配置文件数组定义
type ConfigFileArr []*apiconfig.ConfigFile

// Reset reset initialization
func (m *ConfigFileArr) Reset() { *m = ConfigFileArr{} }

// String return string
func (m *ConfigFileArr) String() string { return proto.CompactTextString(m) }

// ProtoMessage proto message
func (*ConfigFileArr) ProtoMessage() {}
