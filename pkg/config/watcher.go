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
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
)

const (
	defaultLongPollingTimeout = 30000 * time.Millisecond
	QueueSize                 = 10240
)

var (
	notModifiedResponse = &apiconfig.ConfigDiscoverResponse{
		Code: uint32(apimodel.Code_DataNoChange),
		Info: apimodel.Code_DataNoChange.String(),
		Type: apiconfig.ConfigDiscoverResponse_UNKNOWN,
	}
)

type (
	BetaReleaseMatcher func(clientLabels map[string]string, event *conftypes.SimpleConfigFileRelease) bool

	FileReleaseCallback func(clientId string, rsp *apiconfig.ConfigDiscoverResponse) bool

	ConfigSnapshotResolver func(labels map[string]string,
		file *apiconfig.ConfigFileRelease) *apiconfig.ConfigDiscoverResponse

	WatchContextFactory func(clientId string, matcher BetaReleaseMatcher) WatchContext

	WatchContext interface {
		// ClientID 客户端发起的
		ClientID() string
		// ClientLabels 客户端的标识，用于灰度发布要做标签的匹配判断
		ClientLabels() map[string]string
		// AppendInterest 客户端增加订阅列表
		AppendInterest(item *apiconfig.ConfigFileRelease)
		// RemoveInterest 客户端删除订阅列表
		RemoveInterest(item *apiconfig.ConfigFileRelease)
		// ShouldNotify 判断是不是需要通知客户端某个配置变动了
		ShouldNotify(event *conftypes.SimpleConfigFileRelease) bool
		// Reply 真正的通知逻辑
		Reply(rsp *apiconfig.ConfigDiscoverResponse)
		// Close .
		Close() error
		// ShouldExpire 是不是存在有效时间
		ShouldExpire(now time.Time) bool
		// ListWatchFiles 列举出当前订阅的所有配置文件
		ListWatchFiles() []*apiconfig.ConfigFileRelease
		// CurWatchVersion 获取当前订阅的配置文件的版本
		CurWatchVersion(k string) uint64
		// IsOnce 是不是只能被通知一次
		IsOnce() bool
	}
)

type LongPollWatchContext struct {
	lock             sync.RWMutex
	clientId         string
	labels           map[string]string
	once             sync.Once
	finishTime       time.Time
	finishChan       chan *apiconfig.ConfigDiscoverResponse
	watchConfigFiles map[string]*apiconfig.ConfigFileRelease
	betaMatcher      BetaReleaseMatcher
}

func (c *LongPollWatchContext) ClientLabels() map[string]string {
	return c.labels
}

// IsOnce
func (c *LongPollWatchContext) IsOnce() bool {
	return true
}

func (c *LongPollWatchContext) GetNotifieResult() *apiconfig.ConfigDiscoverResponse {
	return <-c.finishChan
}

func (c *LongPollWatchContext) GetNotifieResultWithTime(timeout time.Duration) (*apiconfig.ConfigDiscoverResponse, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case ret := <-c.finishChan:
		return ret, nil
	case <-timer.C:
		return nil, context.DeadlineExceeded
	}
}

func (c *LongPollWatchContext) ShouldExpire(now time.Time) bool {
	return now.After(c.finishTime)
}

// ClientID .
func (c *LongPollWatchContext) ClientID() string {
	return c.clientId
}

func (c *LongPollWatchContext) ShouldNotify(event *conftypes.SimpleConfigFileRelease) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if event.ReleaseType == conftypes.ReleaseTypeGray && !c.betaMatcher(c.ClientLabels(), event) {
		return false
	}

	key := GenFileId(event.Namespace, event.Group, event.FileName)
	watchFile, ok := c.watchConfigFiles[key]
	if !ok {
		return false
	}
	clientVersion, err := strconv.ParseUint(watchFile.GetId(), 10, 64)
	if err != nil {
		return false
	}
	return clientVersion < event.Version
}

func (c *LongPollWatchContext) ListWatchFiles() []*apiconfig.ConfigFileRelease {
	c.lock.RLock()
	defer c.lock.RUnlock()

	ret := make([]*apiconfig.ConfigFileRelease, 0, len(c.watchConfigFiles))
	for _, v := range c.watchConfigFiles {
		ret = append(ret, v)
	}
	return ret
}

func (c *LongPollWatchContext) CurWatchVersion(k string) uint64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if file, ok := c.watchConfigFiles[k]; ok {
		version, _ := strconv.ParseUint(file.GetId(), 10, 64)
		return version
	}
	return 0
}

// AppendInterest .
func (c *LongPollWatchContext) AppendInterest(item *apiconfig.ConfigFileRelease) {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := GenFileId(item.GetNamespace(), item.GetGroup(), item.GetName())
	c.watchConfigFiles[key] = item
}

// RemoveInterest .
func (c *LongPollWatchContext) RemoveInterest(item *apiconfig.ConfigFileRelease) {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := GenFileId(item.GetNamespace(), item.GetGroup(), item.GetName())
	delete(c.watchConfigFiles, key)
}

// Close .
func (c *LongPollWatchContext) Close() error {
	return nil
}

func (c *LongPollWatchContext) Reply(rsp *apiconfig.ConfigDiscoverResponse) {
	c.once.Do(func() {
		c.finishChan <- rsp
		close(c.finishChan)
	})
}

// watchCenter 处理客户端订阅配置请求，监听配置文件发布事件通知客户端
type watchCenter struct {
	subCtx *eventhub.SubscribtionContext
	lock   sync.Mutex
	// clientId -> watchContext
	clients *container.SyncMap[string, WatchContext]
	// fileId -> []clientId
	watchers *container.SyncMap[string, *container.SyncSet[string]]
	// fileCache
	fileCache        cacheapi.ConfigFileCache
	cacheMgr         cacheapi.CacheManager
	snapshotResolver ConfigSnapshotResolver
	cancel           context.CancelFunc
}

// NewWatchCenter 创建一个客户端监听配置发布的处理中心
func NewWatchCenter(cacheMgr cacheapi.CacheManager, snapshotResolver ConfigSnapshotResolver) (*watchCenter, error) {
	ctx, cancel := context.WithCancel(context.Background())

	wc := &watchCenter{
		clients:          container.NewSyncMap[string, WatchContext](),
		watchers:         container.NewSyncMap[string, *container.SyncSet[string]](),
		fileCache:        cacheMgr.ConfigFile(),
		cacheMgr:         cacheMgr,
		snapshotResolver: snapshotResolver,
		cancel:           cancel,
	}

	var err error
	wc.subCtx, err = eventhub.Subscribe(eventhub.ConfigFilePublishTopic, wc, eventhub.WithQueueSize(QueueSize))
	if err != nil {
		return nil, err
	}
	go wc.startHandleTimeoutRequestWorker(ctx)
	return wc, nil
}

// PreProcess do preprocess logic for event
func (wc *watchCenter) PreProcess(_ context.Context, e any) any {
	return e
}

// OnEvent event process logic
func (wc *watchCenter) OnEvent(ctx context.Context, arg any) error {
	switch event := arg.(type) {
	case *eventhub.PublishConfigFileEvent:
		wc.notifyToWatchers(event.Message)
	case *eventhub.ConfigTemplateSnapshotChangedEvent:
		wc.notifyTemplateSnapshotChanged(event)
	default:
		log.Warn("[Config][Watcher] receive invalid event type")
	}
	return nil
}

func (wc *watchCenter) CheckQuickResponseClient(watchCtx WatchContext) *apiconfig.ConfigDiscoverResponse {
	buildRet := func(release *conftypes.ConfigFileRelease) *apiconfig.ConfigDiscoverResponse {
		return &apiconfig.ConfigDiscoverResponse{
			Code: uint32(apimodel.Code_ExecuteSuccess),
			Info: apimodel.Code_ExecuteSuccess.String(),
			Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
			File: &apiconfig.ConfigFileRelease{
				Namespace: release.Namespace,
				Group:     release.Group,
				Name:      release.FileName,
				Version:   release.Version,
				Md5:       release.Md5,
			},
		}
	}

	for _, configFile := range watchCtx.ListWatchFiles() {
		namespace := configFile.GetNamespace()
		group := configFile.GetGroup()
		fileName := configFile.GetName()
		if isSnapshotRevision(configFile.GetId()) {
			response := wc.resolveSnapshot(watchCtx, configFile)
			if response != nil && response.GetCode() == uint32(apimodel.Code_ExecuteSuccess) &&
				response.GetRevision() != "" && response.GetRevision() != configFile.GetId() {
				return response
			}
			continue
		}
		// 从缓存中获取灰度文件
		if len(watchCtx.ClientLabels()) > 0 {
			release := selectMatchedGrayRelease(wc.fileCache.GetActiveGrayReleases(namespace, group, fileName),
				func(release *conftypes.SimpleConfigFileRelease) bool {
					return watchCtx.ShouldNotify(release)
				})
			if release != nil {
				return buildRet(release)
			}
		}
		release := wc.fileCache.GetActiveRelease(namespace, group, fileName)
		// 从缓存中获取最新的配置文件信息
		if release != nil && watchCtx.ShouldNotify(release.SimpleConfigFileRelease) {
			return buildRet(release)
		}
	}
	return nil
}

func isSnapshotRevision(revision string) bool {
	if revision == "" {
		return false
	}
	_, err := strconv.ParseUint(revision, 10, 64)
	return err != nil
}

func (wc *watchCenter) resolveSnapshot(watchCtx WatchContext,
	file *apiconfig.ConfigFileRelease) *apiconfig.ConfigDiscoverResponse {
	if wc.snapshotResolver == nil {
		return nil
	}
	labels := make(map[string]string, len(watchCtx.ClientLabels())+len(file.GetLabels()))
	for key, value := range watchCtx.ClientLabels() {
		labels[key] = value
	}
	for key, value := range file.GetLabels() {
		labels[key] = value
	}
	return wc.snapshotResolver(labels, file)
}

// GetWatchContext .
func (wc *watchCenter) GetWatchContext(clientId string) (WatchContext, bool) {
	return wc.clients.Load(clientId)
}

// DelWatchContext .
func (wc *watchCenter) DelWatchContext(clientId string) (WatchContext, bool) {
	return wc.clients.Delete(clientId)
}

// AddWatcher 新增订阅者
func (wc *watchCenter) AddWatcher(clientId string,
	watchFiles []*apiconfig.ConfigFileRelease, factory WatchContextFactory) WatchContext {
	watchCtx, _ := wc.clients.ComputeIfAbsent(clientId, func(k string) WatchContext {
		return factory(clientId, wc.MatchBetaReleaseFile)
	})

	for _, file := range watchFiles {
		fileKey := GenFileId(file.GetNamespace(), file.GetGroup(), file.GetName())

		watchCtx.AppendInterest(file)
		clientIds, _ := wc.watchers.ComputeIfAbsent(fileKey, func(k string) *container.SyncSet[string] {
			return container.NewSyncSet[string]()
		})
		clientIds.Add(clientId)
	}
	return watchCtx
}

// RemoveAllWatcher 删除订阅者
func (wc *watchCenter) RemoveAllWatcher(clientId string) {
	oldVal, exist := wc.clients.Delete(clientId)
	if !exist {
		return
	}
	_ = oldVal.Close()
	for _, file := range oldVal.ListWatchFiles() {
		watchFileId := GenFileId(file.Namespace, file.Group, file.Name)
		watchers, ok := wc.watchers.Load(watchFileId)
		if !ok {
			continue
		}
		watchers.Remove(clientId)
	}
	wc.clients.Delete(clientId)
}

// RemoveWatcher 删除订阅者
func (wc *watchCenter) RemoveWatcher(clientId string, watchConfigFiles []*apiconfig.ConfigFile) {
	oldVal, exist := wc.clients.Delete(clientId)
	if exist {
		_ = oldVal.Close()
	}
	if len(watchConfigFiles) == 0 {
		return
	}

	for _, file := range watchConfigFiles {
		watchFileId := GenFileId(file.Namespace, file.Group, file.GetName())
		watchers, ok := wc.watchers.Load(watchFileId)
		if !ok {
			continue
		}
		watchers.Remove(clientId)
	}
}

func (wc *watchCenter) notifyToWatchers(publishConfigFile *conftypes.SimpleConfigFileRelease) {
	watchFileId := GenFileId(publishConfigFile.Namespace, publishConfigFile.Group, publishConfigFile.FileName)
	clientIds, ok := wc.watchers.Load(watchFileId)
	if !ok {
		return
	}
	// 构建ConfigDiscoverResponse
	response := &apiconfig.ConfigDiscoverResponse{
		Code: uint32(apimodel.Code_ExecuteSuccess),
		Info: apimodel.Code_ExecuteSuccess.String(),
		Type: apiconfig.ConfigDiscoverResponse_CONFIG_FILE,
		File: &apiconfig.ConfigFileRelease{
			Namespace: publishConfigFile.Namespace,
			Group:     publishConfigFile.Group,
			Name:      publishConfigFile.FileName,
			Version:   publishConfigFile.Version,
			Md5:       publishConfigFile.Md5,
		},
	}

	notifyCnt := 0
	clientIds.Range(func(clientId string) {
		watchCtx, ok := wc.clients.Load(clientId)
		if !ok {
			log.Warn("[Config][Watcher] not found client when do notify.", zap.String("clientId", clientId),
				zap.String("file", watchFileId))
			return
		}

		clientResponse := response
		shouldNotify := watchCtx.ShouldNotify(publishConfigFile)
		if !shouldNotify {
			for _, file := range watchCtx.ListWatchFiles() {
				if GenFileId(file.GetNamespace(), file.GetGroup(), file.GetName()) != watchFileId ||
					!isSnapshotRevision(file.GetId()) {
					continue
				}
				resolved := wc.resolveSnapshot(watchCtx, file)
				if resolved != nil && resolved.GetCode() == uint32(apimodel.Code_ExecuteSuccess) &&
					resolved.GetRevision() != "" && resolved.GetRevision() != file.GetId() {
					clientResponse = resolved
					shouldNotify = true
				}
				break
			}
		}
		if shouldNotify {
			watchCtx.Reply(clientResponse)
			notifyCnt++
			// 只能用一次，通知完就要立马清理掉这个 WatchContext
			if watchCtx.IsOnce() {
				wc.RemoveAllWatcher(watchCtx.ClientID())
			}
		}
	})

	log.Info("[Config][Watcher] received config file release event.", zap.String("file", watchFileId),
		zap.Uint64("version", publishConfigFile.Version), zap.Int("clients", clientIds.Len()),
		zap.Int("notify", notifyCnt))
}

func (wc *watchCenter) notifyTemplateSnapshotChanged(event *eventhub.ConfigTemplateSnapshotChangedEvent) {
	if event == nil || event.Namespace == "" || event.TemplateID == 0 {
		return
	}
	notified := 0
	wc.clients.Range(func(_ string, watchCtx WatchContext) {
		for _, file := range watchCtx.ListWatchFiles() {
			if file.GetNamespace() != event.Namespace || !isSnapshotRevision(file.GetId()) {
				continue
			}
			response := wc.resolveSnapshot(watchCtx, file)
			if response == nil || response.GetCode() != uint32(apimodel.Code_ExecuteSuccess) ||
				response.GetRevision() == "" || response.GetRevision() == file.GetId() ||
				response.GetRenderSnapshot().GetTemplateBinding().GetTemplateId() != event.TemplateID {
				continue
			}
			watchCtx.Reply(response)
			notified++
			if watchCtx.IsOnce() {
				wc.RemoveAllWatcher(watchCtx.ClientID())
			}
			break
		}
	})
	log.Info("[Config][Watcher] received template snapshot change event.",
		zap.String("namespace", event.Namespace), zap.Uint64("template-id", event.TemplateID),
		zap.Int("notify", notified))
}

func (wc *watchCenter) MatchBetaReleaseFile(clientLabels map[string]string, event *conftypes.SimpleConfigFileRelease) bool {
	return wc.cacheMgr.Gray().HitGrayRule(GetGrayConfigReaseKey(event), clientLabels)
}

func (wc *watchCenter) Close() {
	wc.cancel()
	wc.subCtx.Cancel()
}

func (wc *watchCenter) startHandleTimeoutRequestWorker(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			tNow := time.Now()
			waitRemove := make([]WatchContext, 0, 32)
			wc.clients.Range(func(client string, watchCtx WatchContext) {
				if watchCtx.ShouldExpire(tNow) {
					waitRemove = append(waitRemove, watchCtx)
				}
			})
			if len(waitRemove) > 0 {
				log.Info("remove expire watch context", zap.Any("client-ids", waitRemove))
			}

			for i := range waitRemove {
				watchCtx := waitRemove[i]
				watchCtx.Reply(notModifiedResponse)
				wc.RemoveAllWatcher(watchCtx.ClientID())
			}
		}
	}
}
