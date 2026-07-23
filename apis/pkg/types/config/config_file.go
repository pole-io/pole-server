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
	"time"

	"github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/pkg/utils"
)

/** ----------- DataObject ------------- */

// ConfigFileGroup 配置文件组数据持久化对象
type ConfigFileGroup struct {
	Id         string
	Name       string
	Namespace  string
	Comment    string
	Owner      string
	Business   string
	Department string
	Metadata   map[string]string
	CreateTime time.Time
	ModifyTime time.Time
	CreateBy   string
	ModifyBy   string
	Valid      bool
	Revision   string
}

type ConfigFileKey struct {
	Name      string
	Namespace string
	Group     string
}

func (c ConfigFileKey) String() string {
	return c.Namespace + "@" + c.Group + "@" + c.Name
}

// ConfigFile 配置文件数据持久化对象
type ConfigFile struct {
	Id        string
	Name      string
	Namespace string
	Group     string
	// OriginContent 最原始的配置文件内容数据
	OriginContent string
	Content       string
	Comment       string
	Format        string
	Flag          int
	Valid         bool
	Metadata      map[string]string
	Encrypt       bool
	EncryptAlgo   string
	Status        string
	CreateBy      string
	ModifyBy      string
	ReleaseBy     string
	CreateTime    time.Time
	ModifyTime    time.Time
	ReleaseTime   time.Time
}

func (s *ConfigFile) Key() *ConfigFileKey {
	return &ConfigFileKey{
		Name:      s.Name,
		Namespace: s.Namespace,
		Group:     s.Group,
	}
}

func (s *ConfigFile) KeyString() string {
	return s.Namespace + "@" + s.Group + "@" + s.Name
}

func (s *ConfigFile) GetEncryptDataKey() string {
	return s.Metadata[types.MetaKeyConfigFileDataKey]
}

func (s *ConfigFile) GetEncryptAlgo() string {
	if s.EncryptAlgo != "" {
		return s.EncryptAlgo
	}
	return s.Metadata[types.MetaKeyConfigFileEncryptAlgo]
}

func (s *ConfigFile) IsEncrypted() bool {
	return s.Encrypt || s.GetEncryptDataKey() != ""
}

func NewConfigFileRelease() *ConfigFileRelease {
	return &ConfigFileRelease{
		SimpleConfigFileRelease: &SimpleConfigFileRelease{
			ConfigFileReleaseKey: &ConfigFileReleaseKey{},
		},
	}
}

// ConfigFileRelease 配置文件发布数据持久化对象
type ConfigFileRelease struct {
	*SimpleConfigFileRelease
	Content string
}

type ConfigFileReleaseKey struct {
	Id          string
	Name        string
	Namespace   string
	Group       string
	FileName    string
	ReleaseType rules.ReleaseType
	Version     uint64
}

func (c ConfigFileReleaseKey) GetReleaseName() string {
	return c.Name
}

func (c ConfigFileReleaseKey) ToFileKey() *ConfigFileKey {
	return &ConfigFileKey{
		Name:      c.FileName,
		Group:     c.Group,
		Namespace: c.Namespace,
	}
}

func (c *ConfigFileReleaseKey) OwnerKey() string {
	return c.Namespace + "@" + c.Group
}

func (c ConfigFileReleaseKey) FileKey() string {
	return fmt.Sprintf("%v@%v@%v", c.Namespace, c.Group, c.FileName)
}

func (c ConfigFileReleaseKey) ActiveKey() string {
	if c.ReleaseType == ReleaseTypeGray && c.Name != "" {
		return fmt.Sprintf("%v@%v@%v@%v@%v", c.Namespace, c.Group, c.FileName, c.ReleaseType, c.Name)
	}
	return fmt.Sprintf("%v@%v@%v@%v", c.Namespace, c.Group, c.FileName, c.ReleaseType)
}

func (c ConfigFileReleaseKey) ReleaseKey() string {
	return fmt.Sprintf("%v@%v@%v@%v", c.Namespace, c.Group, c.FileName, c.Name)
}

// BuildKeyForClientConfigFileInfo 必须保证和 ConfigFileReleaseKey.FileKey 是一样的生成规则
func BuildKeyForClientConfigFileInfo(info *config_manage.ConfigFileRelease) string {
	key := info.GetNamespace() + "@" +
		info.GetGroup() + "@" + info.GetFileName()
	return key
}

// SimpleConfigFileRelease 配置文件发布数据持久化对象
type SimpleConfigFileRelease struct {
	*ConfigFileReleaseKey
	Version            uint64
	Comment            string
	Md5                string
	Flag               int
	Active             bool
	Valid              bool
	Format             string
	Metadata           map[string]string
	CreateTime         time.Time
	CreateBy           string
	ModifyTime         time.Time
	ModifyBy           string
	ReleaseDescription string
	BetaLabels         []*apimodel.ClientLabel
}

func (s *SimpleConfigFileRelease) GetGrayResource() string {
	if s.Name != "" {
		return fmt.Sprintf("%v@%v@%v@%v@%v", rules.GrayModuleConfig, s.Namespace, s.Group, s.FileName, s.Name)
	}
	return fmt.Sprintf("%v@%v@%v@%v", rules.GrayModuleConfig, s.Namespace, s.Group, s.FileName)
}

func (s *SimpleConfigFileRelease) GetClientLabels() []*apimodel.ClientLabel {
	return s.BetaLabels
}

func (s *SimpleConfigFileRelease) GetEncryptDataKey() string {
	return s.Metadata[types.MetaKeyConfigFileDataKey]
}

func (s *SimpleConfigFileRelease) GetEncryptAlgo() string {
	return s.Metadata[types.MetaKeyConfigFileEncryptAlgo]
}

func (s *SimpleConfigFileRelease) IsEncrypted() bool {
	return s.GetEncryptDataKey() != ""
}

func (s *SimpleConfigFileRelease) ToSpecNotifyClientRequest() *config_manage.ConfigFileRelease {
	return &config_manage.ConfigFileRelease{
		Namespace: s.Namespace,
		Group:     s.Group,
		FileName:  s.FileName,
		Name:      s.Name,
		Md5:       s.Md5,
		Version:   s.Version,
	}
}

// ConfigFileReleaseHistory 配置文件发布历史记录数据持久化对象
type ConfigFileReleaseHistory struct {
	Id                 string
	Name               string
	Namespace          string
	Group              string
	FileName           string
	Format             string
	Metadata           map[string]string
	Content            string
	Comment            string
	Version            uint64
	Md5                string
	Type               string
	Status             string
	CreateTime         time.Time
	CreateBy           string
	ModifyTime         time.Time
	ModifyBy           string
	Valid              bool
	Reason             string
	ReleaseDescription string
}

func (s ConfigFileReleaseHistory) GetEncryptDataKey() string {
	return s.Metadata[types.MetaKeyConfigFileDataKey]
}

func (s ConfigFileReleaseHistory) GetEncryptAlgo() string {
	return s.Metadata[types.MetaKeyConfigFileEncryptAlgo]
}

func (s ConfigFileReleaseHistory) IsEncrypted() bool {
	return s.GetEncryptDataKey() != ""
}

// ConfigFileTag 配置文件标签数据持久化对象
type ConfigFileTag struct {
	Id         uint64
	Key        string
	Value      string
	Namespace  string
	Group      string
	FileName   string
	CreateTime time.Time
	CreateBy   string
	ModifyTime time.Time
	ModifyBy   string
	Valid      bool
}

// ConfigFileTemplate config file template data object
type ConfigFileTemplate struct {
	Id         uint64
	Name       string
	Content    string
	Comment    string
	Format     string
	CreateTime time.Time
	CreateBy   string
	ModifyTime time.Time
	ModifyBy   string
}

func ToConfigFileStore(file *config_manage.ConfigFile) *ConfigFile {
	metadata := file.GetLabels()
	if file.GetEncryptAlgo() != "" {
		metadata[types.MetaKeyConfigFileEncryptAlgo] = file.GetEncryptAlgo()
	}

	return &ConfigFile{
		Name:        file.Name,
		Namespace:   file.Namespace,
		Group:       file.Group,
		Content:     file.Content,
		Comment:     file.Comment,
		Format:      file.Format,
		Encrypt:     file.GetEncrypted(),
		EncryptAlgo: file.GetEncryptAlgo(),
		Metadata:    metadata,
	}
}

func ToConfigFileAPI(file *ConfigFile) *config_manage.ConfigFile {
	if file == nil {
		return nil
	}
	return &config_manage.ConfigFile{
		Id:          file.Id,
		Name:        file.Name,
		Namespace:   file.Namespace,
		Group:       file.Group,
		Content:     file.Content,
		Comment:     file.Comment,
		Format:      file.Format,
		Status:      file.Status,
		Labels:      file.Metadata,
		EncryptAlgo: file.GetEncryptAlgo(),
		Encrypted:   file.IsEncrypted(),
		Ctime:       utils.Time2String(file.CreateTime),
		Mtime:       utils.Time2String(file.ModifyTime),
		Rtime:       utils.Time2String(file.ReleaseTime),
	}
}

// ToConfiogFileReleaseApi
func ToConfiogFileReleaseApi(release *ConfigFileRelease) *config_manage.ConfigFileRelease {
	if release == nil {
		return nil
	}

	return &config_manage.ConfigFileRelease{
		Id:                 release.Id,
		Name:               release.Name,
		Namespace:          release.Namespace,
		Group:              release.Group,
		FileName:           release.FileName,
		Format:             release.Format,
		Content:            release.Content,
		Comment:            release.Comment,
		Md5:                release.Md5,
		Version:            release.Version,
		Ctime:              utils.Time2String(release.CreateTime),
		CreateBy:           release.CreateBy,
		Mtime:              utils.Time2String(release.ModifyTime),
		ModifyBy:           release.ModifyBy,
		ReleaseDescription: release.ReleaseDescription,
		Labels:             release.Metadata,
		Active:             release.Active,
		ReleaseType:        string(release.ReleaseType),
		BetaLabels:         release.BetaLabels,
	}
}

// ToConfigFileReleaseStore
func ToConfigFileReleaseStore(release *config_manage.ConfigFileRelease) *ConfigFileRelease {
	if release == nil {
		return nil
	}

	return &ConfigFileRelease{
		SimpleConfigFileRelease: &SimpleConfigFileRelease{
			ConfigFileReleaseKey: &ConfigFileReleaseKey{
				Id:        release.Id,
				Namespace: release.Namespace,
				Group:     release.Group,
				FileName:  release.FileName,
			},
			Comment:  release.Comment,
			Md5:      release.Md5,
			Version:  release.Version,
			CreateBy: release.CreateBy,
			ModifyBy: release.ModifyBy,
		},
		Content: release.Content,
	}
}

func ToReleaseHistoryAPI(releaseHistory *ConfigFileReleaseHistory) *config_manage.ConfigFileReleaseHistory {
	if releaseHistory == nil {
		return nil
	}
	return &config_manage.ConfigFileReleaseHistory{
		Id:                 releaseHistory.Id,
		Name:               releaseHistory.Name,
		Namespace:          releaseHistory.Namespace,
		Group:              releaseHistory.Group,
		FileName:           releaseHistory.FileName,
		Content:            releaseHistory.Content,
		Comment:            releaseHistory.Comment,
		Format:             releaseHistory.Format,
		Md5:                releaseHistory.Md5,
		Type:               releaseHistory.Type,
		Status:             releaseHistory.Status,
		CreateBy:           releaseHistory.CreateBy,
		Ctime:              utils.Time2String(releaseHistory.CreateTime),
		ModifyBy:           releaseHistory.ModifyBy,
		Mtime:              utils.Time2String(releaseHistory.ModifyTime),
		ReleaseDescription: releaseHistory.ReleaseDescription,
		Reason:             releaseHistory.Reason,
		Labels:             releaseHistory.Metadata,
	}
}

func ToConfigGroupAPI(group *ConfigFileGroup) *config_manage.ConfigFileGroup {
	if group == nil {
		return nil
	}
	return &config_manage.ConfigFileGroup{
		Id:         group.Id,
		Name:       group.Name,
		Namespace:  group.Namespace,
		Comment:    group.Comment,
		Ctime:      utils.Time2String(group.CreateTime),
		Mtime:      utils.Time2String(group.ModifyTime),
		Business:   group.Business,
		Department: group.Department,
		Metadata:   group.Metadata,
		Editable:   true,
		Deleteable: true,
	}
}

func ToConfigGroupStore(group *config_manage.ConfigFileGroup) *ConfigFileGroup {
	return &ConfigFileGroup{
		Name:       group.GetName(),
		Namespace:  group.GetNamespace(),
		Comment:    group.Comment,
		Valid:      true,
		Business:   group.GetBusiness(),
		Department: group.GetDepartment(),
		Metadata:   group.GetMetadata(),
	}
}

func ToConfigFileTemplateAPI(template *ConfigFileTemplate) *config_manage.ConfigFileTemplate {
	return &config_manage.ConfigFileTemplate{
		Id:      template.Id,
		Name:    template.Name,
		Content: template.Content,
		Comment: template.Comment,
		Format:  template.Format,
		Ctime:   utils.Time2String(template.CreateTime),
		Mtime:   utils.Time2String(template.ModifyTime),
	}
}

func ToConfigFileTemplateStore(template *config_manage.ConfigFileTemplate) *ConfigFileTemplate {
	return &ConfigFileTemplate{
		Id:       template.Id,
		Name:     template.Name,
		Content:  template.Content,
		Comment:  template.Comment,
		Format:   template.Format,
		CreateBy: template.Ctime,
		ModifyBy: template.Mtime,
	}
}

type Subscriber struct {
	// 客户端 ID 信息
	ID string `json:"id"`
	// 客户端 Host 信息
	Host string `json:"host"`
	// 文件发布名称
	ReleaseName string `json:"release_name"`
	// 客户端版本
	Version string `json:"version"`
	// 客户端类型
	ClientType string `json:"client_type"`
}

// ConfigSubscribers 以文件视角的监听数据
type ConfigSubscribers struct {
	// key
	Key ConfigFileKey
	// VersionClients 版本对应的客户端
	VersionClients []*VersionClient `json:"clients"`
}

type VersionClient struct {
	Version     uint64        `json:"version"`
	Subscribers []*Subscriber `json:"subscribers"`
}

// FileReleaseSubscribeInfo 文件订阅信息
type FileReleaseSubscribeInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Group       string            `json:"group"`
	FileName    string            `json:"file_name"`
	ReleaseType rules.ReleaseType `json:"release_type"`
	Version     uint64            `json:"version"`
	ReleaseName string            `json:"release_name"`
}

// ClientSubscriber 以客户端视角的监听数据
type ClientSubscriber struct {
	Subscriber Subscriber `json:"subscriber"`
	// Files
	Files []FileReleaseSubscribeInfo `json:"files"`
}
