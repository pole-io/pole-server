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
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"

	"github.com/pole-io/pole-server/apis/observability/statis"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/types/metrics"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	commontime "github.com/pole-io/pole-server/pkg/common/utils/time"
	"github.com/pole-io/pole-server/pkg/config"
	nacosmodel "github.com/pole-io/pole-server/plugin/apiserver/nacosserver/model"
	nacospb "github.com/pole-io/pole-server/plugin/apiserver/nacosserver/v2/pb"
	"github.com/pole-io/pole-server/plugin/apiserver/nacosserver/v2/remote"
)

type ConnectionClientManager struct {
	configSvr *config.Server
	watchCtx  *eventhub.SubscribtionContext
}

func NewConnectionClientManager(configSvr *config.Server) (*ConnectionClientManager, error) {
	mgr := &ConnectionClientManager{
		configSvr: configSvr,
	}
	subCtx, err := eventhub.Subscribe(remote.ClientConnectionEvent, mgr)
	if err != nil {
		return nil, err
	}
	mgr.watchCtx = subCtx
	return mgr, nil
}

// PreProcess do preprocess logic for event
func (cm *ConnectionClientManager) PreProcess(_ context.Context, a any) any {
	return a
}

// OnEvent event process logic
func (c *ConnectionClientManager) OnEvent(ctx context.Context, a any) error {
	event, ok := a.(*remote.ConnectionEvent)
	if !ok {
		return nil
	}
	switch event.EventType {
	case remote.EventClientConnected:
		// do nothing
	case remote.EventClientDisConnected:
		c.configSvr.WatchCenter().RemoveAllWatcher(event.ConnID)
	}

	return nil
}

type StreamWatchContext struct {
	clientId         string
	labels           map[string]string
	connMgr          *remote.ConnectionManager
	watchConfigFiles *container.SyncMap[string, *apiconfig.ConfigFile]
	betaMatcher      config.BetaReleaseMatcher
}

func (c *StreamWatchContext) ClientLabels() map[string]string {
	return c.labels
}

// IsOnce
func (c *StreamWatchContext) IsOnce() bool {
	return false
}

func (c *StreamWatchContext) ShouldExpire(now time.Time) bool {
	return false
}

// ClientID .
func (c *StreamWatchContext) ClientID() string {
	return c.clientId
}

// ShouldNotify .
func (c *StreamWatchContext) ShouldNotify(event *conftypes.SimpleConfigFileRelease) bool {
	if event.ReleaseType == conftypes.ReleaseTypeGray && !c.betaMatcher(c.ClientLabels(), event) {
		return false
	}
	key := event.FileKey()
	watchFile, ok := c.watchConfigFiles.Load(key)
	if !ok {
		return false
	}
	// 删除操作，直接通知
	if !event.Valid {
		return true
	}
	// ConfigFile没有GetMd5方法，使用Tags中的md5值进行比较
	// 如果Tags中没有md5或为空，认为有变化需要通知
	watchFileMd5 := ""
	if watchFile.GetLabels() != nil {
		watchFileMd5 = watchFile.GetLabels()["md5"]
	}
	// 与原逻辑保持一致：比较MD5值是否不同
	isChange := watchFileMd5 != event.Md5
	return isChange
}

// ListWatchFiles .
func (c *StreamWatchContext) ListWatchFiles() []*apiconfig.ConfigFile {
	return c.watchConfigFiles.Values()
}

func (c *StreamWatchContext) CurWatchVersion(k string) uint64 {
	watchFile, ok := c.watchConfigFiles.Load(k)
	if !ok {
		return 0
	}
	// 尝试从Labels中获取版本信息，如果没有则使用ID作为版本
	if watchFile.GetLabels() != nil {
		if version := watchFile.GetLabels()["version"]; version != "" {
			if v, err := strconv.ParseUint(version, 10, 64); err == nil {
				return v
			}
		}
	}
	// 使用ID字段作为版本号
	if version, err := strconv.ParseUint(watchFile.GetId(), 10, 64); err == nil {
		return version
	}
	return 0
}

// AppendInterest .
func (c *StreamWatchContext) AppendInterest(item *apiconfig.ConfigFile) {
	// 使用自定义的key生成方式
	key := fmt.Sprintf("%s@%s@%s", item.GetNamespace(), item.GetGroup(), item.GetName())
	c.watchConfigFiles.Store(key, item)
}

// RemoveInterest .
func (c *StreamWatchContext) RemoveInterest(item *apiconfig.ConfigFile) {
	// 使用自定义的key生成方式
	key := fmt.Sprintf("%s@%s@%s", item.GetNamespace(), item.GetGroup(), item.GetName())
	c.watchConfigFiles.Delete(key)
}

// Close .
func (c *StreamWatchContext) Close() error {
	return nil
}

// Reply .
func (c *StreamWatchContext) Reply(rsp *apiconfig.ConfigDiscoverResponse) {
	// 从 ConfigDiscoverResponse 中获取配置文件信息
	var viewConfig *apiconfig.ConfigFileRelease
	if rsp.GetFile() != nil {
		viewConfig = rsp.GetFile()
	} else if len(rsp.GetFileNames()) > 0 {
		viewConfig = rsp.GetFileNames()[0]
	} else {
		return
	}

	notifyRequest := nacospb.NewConfigChangeNotifyRequest()
	notifyRequest.Tenant = nacosmodel.ToNacosConfigNamespace(viewConfig.GetNamespace())
	notifyRequest.Group = viewConfig.GetGroup()
	notifyRequest.DataId = viewConfig.GetFileName()

	success := false
	startTime := commontime.CurrentMillisecond()
	defer func() {
		statis.GetStatis().ReportDiscoverCall(metrics.ClientDiscoverMetric{
			Action:    nacosmodel.ActionGrpcPushConfigFile,
			ClientIP:  c.ClientID(),
			Namespace: notifyRequest.Tenant,
			Resource:  metrics.ResourceOfConfigFile(notifyRequest.Group, notifyRequest.DataId),
			Timestamp: startTime,
			CostTime:  commontime.CurrentMillisecond() - startTime,
			Revision:  viewConfig.GetMd5(),
			Success:   success,
		})
	}()

	remoteClient, ok := c.connMgr.GetClient(c.clientId)
	if !ok {
		nacoslog.Error("[NACOS-V2][Config][Push] send ConfigChangeNotifyRequest not found remoteClient",
			zap.String("clientId", c.ClientID()))
		return
	}
	stream, ok := remoteClient.LoadStream()
	if !ok {
		nacoslog.Error("[NACOS-V2][Config][Push] send ConfigChangeNotifyRequest not stream",
			zap.String("clientId", c.ClientID()))
		return
	}
	clientResp, err := remote.MarshalPayload(notifyRequest)
	if err != nil {
		nacoslog.Error("[NACOS-V2][Config][Push] send ConfigChangeNotifyRequest marshal payload",
			zap.String("clientId", c.ClientID()), zap.Error(err))
		return
	}
	if err := stream.SendMsg(clientResp); err != nil {
		nacoslog.Error("[NACOS-V2][Config][Push] send ConfigChangeNotifyRequest fail",
			zap.String("clientId", c.ClientID()), zap.Error(err))
	}
	success = true
}
