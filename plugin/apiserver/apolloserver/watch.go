package apolloserver

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/pole-io/specification/source/go/api/v1/config_manage"
	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/config"
)

func (a *ApolloServer) diffChangeFiles(ctx context.Context,
	listenCtx *config_manage.WatchConfigFileRequest) []*ApolloConfigNotification {
	clientLabels := map[string]string{
		types.ClientLabel_IP: utils.ParseClientIP(ctx),
	}
	changeKeys := make([]*ApolloConfigNotification, 0, 4)
	// quick get file and compare
	for _, item := range listenCtx.Files {
		namespace := item.Namespace
		group := item.Group
		filename := item.FileName
		mdval := item.Md5

		if beta := a.innerSvr.CacheManager().ConfigFile().GetActiveGrayRelease(namespace, group, filename); beta != nil {
			if a.innerSvr.CacheManager().Gray().HitGrayRule(beta.FileKey(), clientLabels) {
				notificationId, _ := strconv.ParseInt(beta.Id, 10, 64)
				changeKeys = append(changeKeys, &ApolloConfigNotification{
					NamespaceName:  filename,
					NotificationId: notificationId,
				})
				continue
			}
		}

		active := a.innerSvr.CacheManager().ConfigFile().GetActiveRelease(namespace, group, filename)
		if (active == nil && mdval != "") || (active != nil && active.Md5 != mdval) {
			notificationId, _ := strconv.ParseInt(active.Id, 10, 64)
			changeKeys = append(changeKeys, &ApolloConfigNotification{
				NamespaceName:  filename,
				NotificationId: notificationId,
			})
		}
	}
	return changeKeys
}

func (a *ApolloServer) BuildTimeoutWatchCtx(ctx context.Context, watchTimeOut time.Duration) config.WatchContextFactory {
	labels := map[string]string{}
	labels[types.ClientLabel_IP] = utils.ParseClientIP(ctx)

	return func(clientId string, matcher config.BetaReleaseMatcher) config.WatchContext {
		watchCtx := &ApolloWatchContext{
			clientId:         clientId,
			labels:           labels,
			finishTime:       time.Now().Add(watchTimeOut),
			finishChan:       make(chan *apiconfig.ConfigDiscoverResponse),
			watchConfigFiles: map[string]*config_manage.ConfigFile{},
			betaMatcher: func(clientLabels map[string]string, event *conftypes.SimpleConfigFileRelease) bool {
				return a.innerSvr.CacheManager().Gray().HitGrayRule(config.GetGrayConfigReaseKey(event), clientLabels)
			},
		}
		return watchCtx
	}
}

type ApolloWatchContext struct {
	lock             sync.RWMutex
	clientId         string
	labels           map[string]string
	once             sync.Once
	finishTime       time.Time
	finishChan       chan *apiconfig.ConfigDiscoverResponse
	watchConfigFiles map[string]*config_manage.ConfigFile
	betaMatcher      config.BetaReleaseMatcher
}

// ClientID 客户端发起的
func (w *ApolloWatchContext) ClientID() string {
	return w.clientId
}

// ClientLabels 客户端的标识，用于灰度发布要做标签的匹配判断
func (w *ApolloWatchContext) ClientLabels() map[string]string {
	return w.labels
}

// GetNotifieResult .
func (w *ApolloWatchContext) GetNotifieResult() *apiconfig.ConfigDiscoverResponse {
	return <-w.finishChan
}

// GetNotifieResultWithTime .
func (w *ApolloWatchContext) GetNotifieResultWithTime(timeout time.Duration) (*apiconfig.ConfigDiscoverResponse, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case ret := <-w.finishChan:
		return ret, nil
	case <-timer.C:
		return nil, context.DeadlineExceeded
	}
}

// AppendInterest 客户端增加订阅列表
func (w *ApolloWatchContext) AppendInterest(item *config_manage.ConfigFile) {
	// 这里可以实现对订阅列表的追加逻辑
	w.lock.Lock()
	defer w.lock.Unlock()
	if w.watchConfigFiles == nil {
		w.watchConfigFiles = make(map[string]*config_manage.ConfigFile)
	}
	key := item.Namespace + "@" + item.Group + "@" + item.Name
	w.watchConfigFiles[key] = item
}

// RemoveInterest 客户端删除订阅列表
func (w *ApolloWatchContext) RemoveInterest(item *config_manage.ConfigFile) {
	// 这里可以实现对订阅列表的删除逻辑
	w.lock.Lock()
	defer w.lock.Unlock()
	if w.watchConfigFiles == nil {
		return
	}
	key := item.Namespace + "@" + item.Group + "@" + item.Name
	delete(w.watchConfigFiles, key)
}

// ShouldNotify 判断是不是需要通知客户端某个配置变动了
func (w *ApolloWatchContext) ShouldNotify(event *conftypes.SimpleConfigFileRelease) bool {
	// 这里可以实现对配置变动的判断逻辑
	// 如果需要通知客户端，则返回 true
	// 否则返回 false
	w.lock.RLock()
	defer w.lock.RUnlock()
	if w.watchConfigFiles == nil {
		return false
	}
	key := event.FileKey()
	if watchFile, ok := w.watchConfigFiles[key]; ok {
		// 如果是更新操作，检查版本号是否有变化
		clientVersion, err := strconv.ParseUint(watchFile.Id, 10, 64)
		if err != nil {
			return false
		}
		return clientVersion < event.Version
	}
	return false
}

// Reply 真正的通知逻辑
func (w *ApolloWatchContext) Reply(rsp *apiconfig.ConfigDiscoverResponse) {
	// 这里可以实现对客户端的通知逻辑
	// 例如将 rsp 发送到客户端
	w.once.Do(func() {
		w.finishChan <- rsp
		close(w.finishChan)
	})
}

// Close .
func (w *ApolloWatchContext) Close() error {
	// 这里可以实现关闭逻辑
	// 例如清理资源、取消订阅等
	w.once.Do(func() {
		close(w.finishChan)
	})
	return nil
}

// ShouldExpire 是不是存在有效时间
func (w *ApolloWatchContext) ShouldExpire(now time.Time) bool {
	return now.After(w.finishTime)
}

// ListWatchFiles 列举出当前订阅的所有配置文件
func (w *ApolloWatchContext) ListWatchFiles() []*config_manage.ConfigFile {
	// 这里可以实现列举当前订阅的配置文件的逻辑
	// 返回一个包含所有订阅文件信息的切片
	w.lock.RLock()
	defer w.lock.RUnlock()

	ret := make([]*config_manage.ConfigFile, 0, len(w.watchConfigFiles))
	for _, v := range w.watchConfigFiles {
		ret = append(ret, v)
	}
	return ret
}

// CurWatchVersion 获取当前订阅的配置文件的版本
func (w *ApolloWatchContext) CurWatchVersion(k string) uint64 {
	// 这里可以实现获取当前订阅的配置文件版本的逻辑
	// 根据 k（可能是文件名或其他标识）返回对应的版本号
	w.lock.RLock()
	defer w.lock.RUnlock()

	if file, ok := w.watchConfigFiles[k]; ok {
		version, _ := strconv.ParseUint(file.Id, 10, 64)
		return version
	}
	return 0
}

// IsOnce 是不是只能被通知一次
func (w *ApolloWatchContext) IsOnce() bool {
	return true
}
