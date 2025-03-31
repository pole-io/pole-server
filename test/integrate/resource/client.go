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

package resource

import (
	"fmt"

	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/pkg/common/utils"
)

const (
	clientHost = "%v.%v.%v.%v"
)

/**
 * @brief 创建客户端
 */
func CreateClient(index uint32) *apiservice.Client {
	return &apiservice.Client{
		Host:    utils.NewStringValue(fmt.Sprintf(clientHost, index, index, index, index)),
		Type:    apiservice.Client_SDK,
		Version: utils.NewStringValue("8.8.8"),
	}
}
