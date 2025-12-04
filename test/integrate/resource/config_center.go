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

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

const (
	totalGroups        = 2
	confiGroupNameTemp = "fileGroup-%s-%d"
	configFileNameTemp = "file-%s-%d"
)

func MockConfigGroups(ns *apimodel.Namespace) []*apiconfig.ConfigFileGroup {
	ret := make([]*apiconfig.ConfigFileGroup, 0, totalGroups)

	for i := 0; i < totalGroups; i++ {

		ret = append(ret, &apiconfig.ConfigFileGroup{
			Name:      fmt.Sprintf(confiGroupNameTemp, utils.NewUUID(), i),
			Namespace: ns.GetName(),
			Comment:   "",
		})

	}

	return ret
}

func MockConfigFiles(group *apiconfig.ConfigFileGroup) []*apiconfig.ConfigFile {
	ret := make([]*apiconfig.ConfigFile, 0, totalGroups)

	for i := 0; i < totalGroups; i++ {
		name := fmt.Sprintf(configFileNameTemp, utils.NewUUID(), i)
		if i%2 == 0 {
			name = fmt.Sprintf("dir%d/", i) + name
		}
		ret = append(ret, &apiconfig.ConfigFile{
			Name:      name,
			Namespace: group.Namespace,
			Group:     group.Name,
			Content:   `name: polarismesh`,
			Format:    "yaml",
			Status:    conftypes.ReleaseStatusToRelease,
			Labels:    map[string]string{},
		})

	}

	return ret
}
