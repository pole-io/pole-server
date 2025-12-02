/*
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
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/golang/protobuf/ptypes/wrappers"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
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

var (
	regFileName = regexp.MustCompile(`^[\dA-Za-z-./:_]+$`)
)

// CheckFileName 校验文件名
func CheckFileName(name *wrappers.StringValue) error {
	if name == nil {
		return errors.New(valid.NilErrString)
	}

	if name.GetValue() == "" {
		return errors.New(valid.EmptyErrString)
	}
	return nil
}

// ResolveFileType 解析文件类型
func ResolveFileType(name string) string {
	p := strings.LastIndex(name, ".")
	if p < 0 || p == len(name)-1 {
		return "txt"
	}
	ext := name[p+1:]
	if ext == "yaml" || ext == "yml" {
		return "yaml"
	}
	return ext
}

// CalMd5 计算md5值
func CalMd5(content string) string {
	h := md5.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

// CheckContentLength 校验文件内容长度
func CheckContentLength(content string, max int) error {
	if utf8.RuneCountInString(content) > max {
		return fmt.Errorf("content length too long. max length =%d", max)
	}

	return nil
}

func CompressConfigFiles(files []*conftypes.ConfigFile,
	fileID2Tags map[string][]*conftypes.ConfigFileTag, isExportGroup bool) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	defer w.Close()

	var configFileMetas = make(map[string]*conftypes.ConfigFileMeta)
	for _, file := range files {
		fileName := file.Name
		if isExportGroup {
			fileName = path.Join(file.Group, file.Name)
		}

		configFileMetas[fileName] = &conftypes.ConfigFileMeta{
			Tags:    make(map[string]string),
			Comment: file.Comment,
		}
		for _, tag := range fileID2Tags[file.Id] {
			configFileMetas[fileName].Tags[tag.Key] = tag.Value
		}
		f, err := w.Create(fileName)
		if err != nil {
			return nil, err
		}
		if _, err := f.Write([]byte(file.Content)); err != nil {
			return nil, err
		}
	}
	// 生成配置元文件
	f, err := w.Create(ConfigFileMetaFileName)
	if err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(configFileMetas, "", "\t")
	if err != nil {
		return nil, err
	}
	if _, err := f.Write(data); err != nil {
		return nil, err
	}
	return &buf, nil
}

// GetGrayConfigReaseKey 获取灰度资源key
func GetGrayConfigReaseKey(release *conftypes.SimpleConfigFileRelease) string {
	return fmt.Sprintf("%v@%v@%v@%v", rules.GrayModuleConfig, release.Namespace, release.Group, release.FileName)
}
