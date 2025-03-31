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

package whitelist

import (
	"sync"

	"github.com/pole-io/pole-server/apis"
)

var (
	whitelistOnce sync.Once
)

// Whitelist White list interface
type Whitelist interface {
	apis.Plugin

	Contain(entry interface{}) bool
}

// GetWhitelist Get the whitelist plugin
func GetWhitelist() Whitelist {
	return nil
}
