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

package utils

import (
	"strings"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
)

const (
	// ConfigFileFormKey 配置文件表单键
	ConfigFileFormKey = "config"
	// ConfigFileMetaFileName 配置文件元数据文件名
	ConfigFileMetaFileName = "META"
	// ConfigFileImportConflictSkip 导入配置文件发生冲突跳过
	ConfigFileImportConflictSkip = "skip"
	// ConfigFileImportConflictOverwrite 导入配置文件发生冲突覆盖原配置文件
	ConfigFileImportConflictOverwrite = "overwrite"
)

// GenFileId 生成文件 Id
func GenFileId(namespace, group, fileName string) string {
	return namespace + conftypes.FileIdSeparator + group + conftypes.FileIdSeparator + fileName
}

// ParseFileId 解析文件 Id
func ParseFileId(fileId string) (namespace, group, fileName string) {
	fileInfo := strings.Split(fileId, conftypes.FileIdSeparator)
	return fileInfo[0], fileInfo[1], fileInfo[2]
}
