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
	"context"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/pkg/types"
)

type (
	// WatchCallback 监听回调函数
	WatchCallback func() *apiconfig.ConfigDiscoverResponse
)

const (
	// MaxPageSize 最大分页大小
	MaxPageSize = 100
)

// ConfigFileGroupOperate 配置文件组接口
type ConfigFileGroupOperate interface {
	// CreateConfigFileGroups 创建配置文件组
	CreateConfigFileGroups(ctx context.Context, reqs []*apiconfig.ConfigFileGroup) *apimodel.BatchWriteResponse
	// UpdateConfigFileGroups 更新配置文件组
	UpdateConfigFileGroups(ctx context.Context, reqs []*apiconfig.ConfigFileGroup) *apimodel.BatchWriteResponse
	// DeleteConfigFileGroups 删除配置文件组
	DeleteConfigFileGroups(ctx context.Context, reqs []*apiconfig.ConfigFileGroup) *apimodel.BatchWriteResponse
	// QueryConfigFileGroups 查询配置文件组
	QueryConfigFileGroups(ctx context.Context, filter map[string]string) *apimodel.BatchQueryResponse
}

// ConfigFileOperate 配置文件接口
type ConfigFileOperate interface {
	// CreateConfigFiles 创建配置文件
	CreateConfigFiles(ctx context.Context, reqs []*apiconfig.ConfigFile) *apimodel.BatchWriteResponse
	// UpdateConfigFile 更新配置文件
	UpdateConfigFiles(ctx context.Context, reqs []*apiconfig.ConfigFile) *apimodel.BatchWriteResponse
	// DeleteConfigFiles 批量删除配置文件
	DeleteConfigFiles(ctx context.Context, req []*apiconfig.ConfigFile) *apimodel.BatchWriteResponse
	// GetConfigFileRichInfo 获取单个配置文件基础信息，包含发布状态等信息
	GetConfigFileRichInfo(ctx context.Context, req *apiconfig.ConfigFile) *apimodel.Response
	// SearchConfigFiles 按 group 和 name 模糊搜索配置文件
	SearchConfigFiles(ctx context.Context, filter map[string]string) *apimodel.BatchQueryResponse
	// ExportConfigFile 导出配置文件
	ExportConfigFile(ctx context.Context,
		configFileExport *apiconfig.ConfigFileExportRequest) *apimodel.Response
	// ImportConfigFile 导入配置文件
	ImportConfigFile(ctx context.Context,
		configFiles []*apiconfig.ConfigFile, conflictHandling string) *apimodel.Response
	// GetAllConfigEncryptAlgorithms 获取配置加密算法
	GetAllConfigEncryptAlgorithms(ctx context.Context) *apimodel.Response
	// GetClientSubscribers 获取客户端订阅者
	GetClientSubscribers(ctx context.Context, filter map[string]string) *types.CommonResponse
	// GetConfigSubscribers 获取配置订阅者
	GetConfigSubscribers(ctx context.Context, filter map[string]string) *types.CommonResponse
}

// ConfigFileReleaseOperate 配置文件发布接口
type ConfigFileReleaseOperate interface {
	// PublishConfigFile 发布配置文件
	PublishConfigFile(ctx context.Context, configFileRelease *apiconfig.ConfigFileRelease) *apimodel.Response
	// GetConfigFileRelease 获取配置文件发布
	GetConfigFileRelease(ctx context.Context, req *apiconfig.ConfigFileRelease) *apimodel.Response
	// DeleteConfigFileReleases 批量删除配置文件发布内容
	DeleteConfigFileReleases(ctx context.Context, reqs []*apiconfig.ConfigFileRelease) *apimodel.BatchWriteResponse
	// RollbackConfigFileReleases 批量回滚配置到指定版本
	RollbackConfigFileReleases(ctx context.Context, releases []*apiconfig.ConfigFileRelease) *apimodel.BatchWriteResponse
	// GetConfigFileReleases 查询所有的配置发布版本信息
	GetConfigFileReleases(ctx context.Context, filters map[string]string) *apimodel.BatchQueryResponse
	// GetConfigFileReleaseVersions 查询所有的配置发布版本信息
	GetConfigFileReleaseVersions(ctx context.Context, filters map[string]string) *apimodel.BatchQueryResponse
	// GetConfigFileReleaseHistories 获取配置文件的发布历史
	GetConfigFileReleaseHistories(ctx context.Context, filter map[string]string) *apimodel.BatchQueryResponse
	// UpsertAndReleaseConfigFile 创建/更新配置文件并发布
	UpsertAndReleaseConfigFile(ctx context.Context, req *apiconfig.ConfigFilePublishInfo) *apimodel.Response
	// StopGrayConfigFileReleases 停止所有的灰度发布配置
	StopGrayConfigFileReleases(ctx context.Context, reqs []*apiconfig.ConfigFileRelease) *apimodel.BatchWriteResponse
	// PromoteGrayConfigFileReleaseToDraft 将灰度发布版本提交为正式草稿
	PromoteGrayConfigFileReleaseToDraft(ctx context.Context, req *apiconfig.ConfigFileRelease) *apimodel.Response
}

// ConfigFileClientOperate 给客户端提供服务接口，不同的上层协议抽象的公共服务逻辑
type ConfigFileClientOperate interface {
	// CreateConfigFileFromClient 调用config_file的方法创建配置文件
	CreateConfigFileFromClient(ctx context.Context, req *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse
	// UpdateConfigFileFromClient 调用config_file的方法更新配置文件
	UpdateConfigFileFromClient(ctx context.Context, req *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse
	// DeleteConfigFileFromClient 调用config_file的方法更新配置文件
	DeleteConfigFileFromClient(ctx context.Context, req *apiconfig.ConfigFile) *apimodel.Response
	// PublishConfigFileFromClient 调用config_file_release的方法发布配置文件
	PublishConfigFileFromClient(ctx context.Context, req *apiconfig.ConfigFileRelease) *apiconfig.ConfigDiscoverResponse
	// UpsertAndReleaseConfigFile 创建/更新配置文件并发布
	UpsertAndReleaseConfigFileFromClient(ctx context.Context, req *apiconfig.ConfigFilePublishInfo) *apimodel.Response
	// LongPullWatchFile 客户端监听配置文件
	LongPullWatchFile(ctx context.Context, req *apiconfig.WatchConfigFileRequest) (WatchCallback, error)
	// GetConfigFileNamesWithCache 获取某个配置分组下的配置文件
	GetConfigFileNamesWithCache(ctx context.Context,
		req *apiconfig.ConfigFileGroupRequest) *apiconfig.ConfigDiscoverResponse
	// GetConfigFileWithCache 获取配置文件
	GetConfigFileWithCache(ctx context.Context, req *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse
	// GetConfigGroupsWithCache 获取某个命名空间下的配置分组列表
	GetConfigGroupsWithCache(ctx context.Context, req *apiconfig.ConfigFile) *apiconfig.ConfigDiscoverResponse
}

// ConfigFileTemplateOperate config file template operate
type ConfigFileTemplateOperate interface {
	// PreviewConfigTemplate performs the server reference render for validation only.
	PreviewConfigTemplate(ctx context.Context, req *apiconfig.RenderPreviewRequest) *apiconfig.RenderPreview
	PublishConfigTemplateRelease(ctx context.Context, req *apiconfig.ConfigTemplateRelease) *apimodel.Response
	GetConfigTemplateLabels(ctx context.Context, templateID uint64) *apimodel.Response
	SaveConfigTemplateLabels(ctx context.Context, templateID uint64, labels map[string]string) *apimodel.Response
	SaveNamespaceTemplateValues(ctx context.Context, req *apiconfig.NamespaceTemplateValues) *apimodel.Response
	PublishNamespaceTemplateValueRelease(
		ctx context.Context, req *apiconfig.NamespaceTemplateValueRelease) *apimodel.Response
	BindConfigFileTemplate(ctx context.Context, file *apiconfig.ConfigFile) *apimodel.Response
	ListConfigTemplateReleases(ctx context.Context, templateID uint64) *apimodel.BatchQueryResponse
	GetNamespaceTemplateValues(ctx context.Context, namespace string, templateID uint64) *apimodel.Response
	ListNamespaceTemplateValueReleases(
		ctx context.Context, namespace string, templateID uint64) *apimodel.BatchQueryResponse
	ListConfigTemplateBindings(
		ctx context.Context, namespace, group, fileName string) *apimodel.BatchQueryResponse
	// GetAllConfigFileTemplates get all config file templates
	GetAllConfigFileTemplates(ctx context.Context) *apimodel.BatchQueryResponse
	// CreateConfigFileTemplates create config file template
	CreateConfigFileTemplates(ctx context.Context, template []*apiconfig.ConfigFileTemplate) *apimodel.Response
	// UpdateConfigFileTemplates create config file template
	UpdateConfigFileTemplates(ctx context.Context, template []*apiconfig.ConfigFileTemplate) *apimodel.Response
	// GetConfigFileTemplate get config file template
	GetConfigFileTemplate(ctx context.Context, name string) *apimodel.Response
}

// ConfigCenterServer 配置中心server
type ConfigCenterServer interface {
	ConfigFileGroupOperate
	ConfigFileOperate
	ConfigFileReleaseOperate
	ConfigFileClientOperate
	ConfigFileTemplateOperate
}

// ResourceHook The listener is placed before and after the resource operation, only normal flow
type ResourceHook interface {
	// Before
	Before(ctx context.Context, resourceType types.Resource)
	// After
	After(ctx context.Context, resourceType types.Resource, res *ResourceEvent) error
}

// ResourceEvent 资源事件
type ResourceEvent struct {
	ConfigGroup *apiconfig.ConfigFileGroup
}
