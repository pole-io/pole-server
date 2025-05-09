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

	"github.com/polarismesh/specification/source/go/api/v1/config_manage"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	"github.com/pole-io/pole-server/apis/pkg/utils"
)

type ReleaseType string

/** ----------- DataObject ------------- */

// ConfigFileGroup 配置文件组数据持久化对象
type ConfigFileGroup struct {
	Id         uint64
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
	Id        uint64
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
	Id          uint64
	Name        string
	Namespace   string
	Group       string
	FileName    string
	ReleaseType ReleaseType
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
	return fmt.Sprintf("%v@%v@%v@%v", c.Namespace, c.Group, c.FileName, c.ReleaseType)
}

func (c ConfigFileReleaseKey) ReleaseKey() string {
	return fmt.Sprintf("%v@%v@%v@%v", c.Namespace, c.Group, c.FileName, c.Name)
}

// BuildKeyForClientConfigFileInfo 必须保证和 ConfigFileReleaseKey.FileKey 是一样的生成规则
func BuildKeyForClientConfigFileInfo(info *config_manage.ClientConfigFileInfo) string {
	key := info.GetNamespace().GetValue() + "@" +
		info.GetGroup().GetValue() + "@" + info.GetFileName().GetValue()
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

func (s *SimpleConfigFileRelease) GetEncryptDataKey() string {
	return s.Metadata[types.MetaKeyConfigFileDataKey]
}

func (s *SimpleConfigFileRelease) GetEncryptAlgo() string {
	return s.Metadata[types.MetaKeyConfigFileEncryptAlgo]
}

func (s *SimpleConfigFileRelease) IsEncrypted() bool {
	return s.GetEncryptDataKey() != ""
}

func (s *SimpleConfigFileRelease) ToSpecNotifyClientRequest() *config_manage.ClientConfigFileInfo {
	return &config_manage.ClientConfigFileInfo{
		Namespace: protobuf.NewStringValue(s.Namespace),
		Group:     protobuf.NewStringValue(s.Group),
		FileName:  protobuf.NewStringValue(s.FileName),
		Name:      protobuf.NewStringValue(s.Name),
		Md5:       protobuf.NewStringValue(s.Md5),
		Version:   protobuf.NewUInt64Value(s.Version),
	}
}

// ConfigFileReleaseHistory 配置文件发布历史记录数据持久化对象
type ConfigFileReleaseHistory struct {
	Id                 uint64
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
	var comment string
	if file.Comment != nil {
		comment = file.Comment.Value
	}
	var createBy string
	if file.CreateBy != nil {
		createBy = file.CreateBy.Value
	}
	var content string
	if file.Content != nil {
		content = file.Content.Value
	}
	var format string
	if file.Format != nil {
		format = file.Format.Value
	}

	metadata := ToTagMap(file.GetTags())
	if file.GetEncryptAlgo().GetValue() != "" {
		metadata[types.MetaKeyConfigFileEncryptAlgo] = file.GetEncryptAlgo().GetValue()
	}

	return &ConfigFile{
		Name:        file.Name.GetValue(),
		Namespace:   file.Namespace.GetValue(),
		Group:       file.Group.GetValue(),
		Content:     content,
		Comment:     comment,
		Format:      format,
		CreateBy:    createBy,
		Encrypt:     file.GetEncrypted().GetValue(),
		EncryptAlgo: file.GetEncryptAlgo().GetValue(),
		Metadata:    metadata,
	}
}

func ToConfigFileAPI(file *ConfigFile) *config_manage.ConfigFile {
	if file == nil {
		return nil
	}
	return &config_manage.ConfigFile{
		Id:          protobuf.NewUInt64Value(file.Id),
		Name:        protobuf.NewStringValue(file.Name),
		Namespace:   protobuf.NewStringValue(file.Namespace),
		Group:       protobuf.NewStringValue(file.Group),
		Content:     protobuf.NewStringValue(file.Content),
		Comment:     protobuf.NewStringValue(file.Comment),
		Format:      protobuf.NewStringValue(file.Format),
		Status:      protobuf.NewStringValue(file.Status),
		Tags:        FromTagMap(file.Metadata),
		Encrypted:   protobuf.NewBoolValue(file.IsEncrypted()),
		EncryptAlgo: protobuf.NewStringValue(file.GetEncryptAlgo()),
		CreateBy:    protobuf.NewStringValue(file.CreateBy),
		ModifyBy:    protobuf.NewStringValue(file.ModifyBy),
		ReleaseBy:   protobuf.NewStringValue(file.ReleaseBy),
		CreateTime:  protobuf.NewStringValue(utils.Time2String(file.CreateTime)),
		ModifyTime:  protobuf.NewStringValue(utils.Time2String(file.ModifyTime)),
		ReleaseTime: protobuf.NewStringValue(utils.Time2String(file.ReleaseTime)),
	}
}

// ToConfiogFileReleaseApi
func ToConfiogFileReleaseApi(release *ConfigFileRelease) *config_manage.ConfigFileRelease {
	if release == nil {
		return nil
	}

	return &config_manage.ConfigFileRelease{
		Id:                 protobuf.NewUInt64Value(release.Id),
		Name:               protobuf.NewStringValue(release.Name),
		Namespace:          protobuf.NewStringValue(release.Namespace),
		Group:              protobuf.NewStringValue(release.Group),
		FileName:           protobuf.NewStringValue(release.FileName),
		Format:             protobuf.NewStringValue(release.Format),
		Content:            protobuf.NewStringValue(release.Content),
		Comment:            protobuf.NewStringValue(release.Comment),
		Md5:                protobuf.NewStringValue(release.Md5),
		Version:            protobuf.NewUInt64Value(release.Version),
		CreateBy:           protobuf.NewStringValue(release.CreateBy),
		CreateTime:         protobuf.NewStringValue(utils.Time2String(release.CreateTime)),
		ModifyBy:           protobuf.NewStringValue(release.ModifyBy),
		ModifyTime:         protobuf.NewStringValue(utils.Time2String(release.ModifyTime)),
		ReleaseDescription: protobuf.NewStringValue(release.ReleaseDescription),
		Tags:               FromTagMap(release.Metadata),
		Active:             protobuf.NewBoolValue(release.Active),
		ReleaseType:        protobuf.NewStringValue(string(release.ReleaseType)),
		BetaLabels:         release.BetaLabels,
	}
}

// ToConfigFileReleaseStore
func ToConfigFileReleaseStore(release *config_manage.ConfigFileRelease) *ConfigFileRelease {
	if release == nil {
		return nil
	}
	var comment string
	if release.Comment != nil {
		comment = release.Comment.Value
	}
	var content string
	if release.Content != nil {
		content = release.Content.Value
	}
	var md5 string
	if release.Md5 != nil {
		md5 = release.Md5.Value
	}
	var version uint64
	if release.Version != nil {
		version = release.Version.Value
	}
	var createBy string
	if release.CreateBy != nil {
		createBy = release.CreateBy.Value
	}
	var modifyBy string
	if release.ModifyBy != nil {
		createBy = release.ModifyBy.Value
	}
	var id uint64
	if release.Id != nil {
		id = release.Id.Value
	}

	return &ConfigFileRelease{
		SimpleConfigFileRelease: &SimpleConfigFileRelease{
			ConfigFileReleaseKey: &ConfigFileReleaseKey{
				Id:        id,
				Namespace: release.Namespace.GetValue(),
				Group:     release.Group.GetValue(),
				FileName:  release.FileName.GetValue(),
			},
			Comment:  comment,
			Md5:      md5,
			Version:  version,
			CreateBy: createBy,
			ModifyBy: modifyBy,
		},
		Content: content,
	}
}

func ToReleaseHistoryAPI(releaseHistory *ConfigFileReleaseHistory) *config_manage.ConfigFileReleaseHistory {
	if releaseHistory == nil {
		return nil
	}
	return &config_manage.ConfigFileReleaseHistory{
		Id:                 protobuf.NewUInt64Value(releaseHistory.Id),
		Name:               protobuf.NewStringValue(releaseHistory.Name),
		Namespace:          protobuf.NewStringValue(releaseHistory.Namespace),
		Group:              protobuf.NewStringValue(releaseHistory.Group),
		FileName:           protobuf.NewStringValue(releaseHistory.FileName),
		Content:            protobuf.NewStringValue(releaseHistory.Content),
		Comment:            protobuf.NewStringValue(releaseHistory.Comment),
		Format:             protobuf.NewStringValue(releaseHistory.Format),
		Tags:               FromTagMap(releaseHistory.Metadata),
		Md5:                protobuf.NewStringValue(releaseHistory.Md5),
		Type:               protobuf.NewStringValue(releaseHistory.Type),
		Status:             protobuf.NewStringValue(releaseHistory.Status),
		CreateBy:           protobuf.NewStringValue(releaseHistory.CreateBy),
		CreateTime:         protobuf.NewStringValue(utils.Time2String(releaseHistory.CreateTime)),
		ModifyBy:           protobuf.NewStringValue(releaseHistory.ModifyBy),
		ModifyTime:         protobuf.NewStringValue(utils.Time2String(releaseHistory.ModifyTime)),
		ReleaseDescription: protobuf.NewStringValue(releaseHistory.ReleaseDescription),
		Reason:             protobuf.NewStringValue(releaseHistory.Reason),
	}
}

type kv struct {
	Key   string
	Value string
}

// FromTagJson 从 Tags Json 字符串里反序列化出 Tags
func FromTagMap(kvs map[string]string) []*config_manage.ConfigFileTag {
	tags := make([]*config_manage.ConfigFileTag, 0, len(kvs))
	for k, v := range kvs {
		tags = append(tags, &config_manage.ConfigFileTag{
			Key:   protobuf.NewStringValue(k),
			Value: protobuf.NewStringValue(v),
		})
	}

	return tags
}

func ToTagMap(tags []*config_manage.ConfigFileTag) map[string]string {
	kvs := map[string]string{}
	for i := range tags {
		kvs[tags[i].GetKey().GetValue()] = tags[i].GetValue().GetValue()
	}
	return kvs
}

func ToConfigGroupAPI(group *ConfigFileGroup) *config_manage.ConfigFileGroup {
	if group == nil {
		return nil
	}
	return &config_manage.ConfigFileGroup{
		Id:         protobuf.NewUInt64Value(group.Id),
		Name:       protobuf.NewStringValue(group.Name),
		Namespace:  protobuf.NewStringValue(group.Namespace),
		Comment:    protobuf.NewStringValue(group.Comment),
		Owner:      protobuf.NewStringValue(group.Owner),
		CreateBy:   protobuf.NewStringValue(group.CreateBy),
		ModifyBy:   protobuf.NewStringValue(group.ModifyBy),
		CreateTime: protobuf.NewStringValue(utils.Time2String(group.CreateTime)),
		ModifyTime: protobuf.NewStringValue(utils.Time2String(group.ModifyTime)),
		Business:   protobuf.NewStringValue(group.Business),
		Department: protobuf.NewStringValue(group.Department),
		Metadata:   group.Metadata,
		Editable:   protobuf.NewBoolValue(true),
		Deleteable: protobuf.NewBoolValue(true),
	}
}

func ToConfigGroupStore(group *config_manage.ConfigFileGroup) *ConfigFileGroup {
	var comment string
	if group.Comment != nil {
		comment = group.Comment.Value
	}
	var createBy string
	if group.CreateBy != nil {
		createBy = group.CreateBy.Value
	}
	var groupOwner string
	if group.Owner != nil && group.Owner.GetValue() != "" {
		groupOwner = group.Owner.GetValue()
	} else {
		groupOwner = createBy
	}
	return &ConfigFileGroup{
		Name:       group.GetName().GetValue(),
		Namespace:  group.GetNamespace().GetValue(),
		Comment:    comment,
		CreateBy:   createBy,
		Valid:      true,
		Owner:      groupOwner,
		Business:   group.GetBusiness().GetValue(),
		Department: group.GetDepartment().GetValue(),
		Metadata:   group.GetMetadata(),
	}
}

func ToConfigFileTemplateAPI(template *ConfigFileTemplate) *config_manage.ConfigFileTemplate {
	return &config_manage.ConfigFileTemplate{
		Id:         protobuf.NewUInt64Value(template.Id),
		Name:       protobuf.NewStringValue(template.Name),
		Content:    protobuf.NewStringValue(template.Content),
		Comment:    protobuf.NewStringValue(template.Comment),
		Format:     protobuf.NewStringValue(template.Format),
		CreateBy:   protobuf.NewStringValue(template.CreateBy),
		CreateTime: protobuf.NewStringValue(utils.Time2String(template.CreateTime)),
		ModifyBy:   protobuf.NewStringValue(template.ModifyBy),
		ModifyTime: protobuf.NewStringValue(utils.Time2String(template.ModifyTime)),
	}
}

func ToConfigFileTemplateStore(template *config_manage.ConfigFileTemplate) *ConfigFileTemplate {
	return &ConfigFileTemplate{
		Id:       template.Id.GetValue(),
		Name:     template.Name.GetValue(),
		Content:  template.Content.GetValue(),
		Comment:  template.Comment.GetValue(),
		Format:   template.Format.GetValue(),
		CreateBy: template.CreateBy.GetValue(),
		ModifyBy: template.ModifyBy.GetValue(),
	}
}

type Subscriber struct {
	// 客户端 ID 信息
	ID string `json:"id"`
	// 客户端 Host 信息
	Host string `json:"host"`
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
	Versoin     uint64        `json:"versoin"`
	Subscribers []*Subscriber `json:"subscribers"`
}

// FileReleaseSubscribeInfo 文件订阅信息
type FileReleaseSubscribeInfo struct {
	Name        string      `json:"name"`
	Namespace   string      `json:"namespace"`
	Group       string      `json:"group"`
	FileName    string      `json:"file_name"`
	ReleaseType ReleaseType `json:"release_type"`
	Version     uint64      `json:"version"`
}

// ClientSubscriber 以客户端视角的监听数据
type ClientSubscriber struct {
	Subscriber Subscriber `json:"subscriber"`
	// Files
	Files []FileReleaseSubscribeInfo `json:"files"`
}
