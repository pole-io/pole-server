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

package local

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/apis/observability/event"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	commonlog "github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"go.uber.org/zap"
)

const (
	PluginName        = "local"
	defaultBufferSize = 1024
)

var log = commonlog.RegisterScope(PluginName+"event", "", 0)

func init() {
	d := &discoverEventLocal{}
	apis.RegisterPlugin(d.Name(), d)
}

type eventBufferHolder struct {
	writeCursor int
	readCursor  int
	size        int
	buffer      []event.DiscoverEvent
}

func newEventBufferHolder(cap int) *eventBufferHolder {
	return &eventBufferHolder{
		writeCursor: 0,
		readCursor:  0,
		size:        0,
		buffer:      make([]event.DiscoverEvent, cap),
	}
}

// Reset 重置 eventBufferHolder，使之可以复用
func (holder *eventBufferHolder) Reset() {
	holder.writeCursor = 0
	holder.readCursor = 0
	holder.size = 0
}

// Put 放入一个 model.DiscoverEvent
func (holder *eventBufferHolder) Put(event event.DiscoverEvent) {
	holder.buffer[holder.writeCursor] = event
	holder.size++
	holder.writeCursor++
}

// HasNext 判断是否还有下一个元素
func (holder *eventBufferHolder) HasNext() bool {
	return holder.readCursor < holder.size
}

// Next 返回下一个元素
//
//	@return model.DiscoverEvent 元素
//	@return bool 是否还有下一个元素可以继续读取
func (holder *eventBufferHolder) Next() event.DiscoverEvent {
	event := holder.buffer[holder.readCursor]
	holder.readCursor++

	return event
}

// Size 当前所存储的有效元素的个数
func (holder *eventBufferHolder) Size() int {
	return holder.size
}

type discoverEventLocal struct {
	eventCh        chan event.DiscoverEvent
	bufferPool     sync.Pool
	curEventBuffer *eventBufferHolder
	cursor         int
	syncLock       sync.Mutex
	eventHandler   func(eventHolder *eventBufferHolder)
	cancel         context.CancelFunc
}

// Name 插件名称
// @return string 返回插件名称
func (el *discoverEventLocal) Name() string {
	return PluginName
}

// Initialize 根据配置文件进行初始化插件 discoverEventLocal
// @param conf 配置文件内容
// @return error 初始化失败，返回 error 信息
func (el *discoverEventLocal) Initialize(conf *apis.ConfigEntry) error {
	contentBytes, err := json.Marshal(conf.Option)
	if err != nil {
		return err
	}

	config := DefaultDiscoverEventConfig()
	if err := json.Unmarshal(contentBytes, config); err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}

	el.eventCh = make(chan event.DiscoverEvent, config.QueueSize)
	el.eventHandler = el.writeToFile
	el.bufferPool = sync.Pool{
		New: func() interface{} {
			return newEventBufferHolder(defaultBufferSize)
		},
	}

	el.switchEventBuffer()
	ctx, cancel := context.WithCancel(context.Background())
	go el.Run(ctx)

	el.cancel = cancel
	return nil
}

func (el *discoverEventLocal) Type() apis.PluginType {
	return apis.PluginTypeDiscoverEvent
}

// Destroy 执行插件销毁
func (el *discoverEventLocal) Destroy() error {
	if el.cancel != nil {
		el.cancel()
	}
	return nil
}

// PublishEvent 发布一个服务事件
func (el *discoverEventLocal) PublishEvent(event event.DiscoverEvent) {
	select {
	case el.eventCh <- event:
		return
	default:
		// do nothing
	}
}

var (
	subscribeEvents = map[string]struct{}{
		string(svctypes.EventInstanceCloseIsolate): {},
		string(svctypes.EventInstanceOpenIsolate):  {},
		string(svctypes.EventInstanceOffline):      {},
		string(svctypes.EventInstanceOnline):       {},
		string(svctypes.EventInstanceTurnHealth):   {},
		string(svctypes.EventInstanceTurnUnHealth): {},
	}
)

// Run 执行主逻辑
func (el *discoverEventLocal) Run(ctx context.Context) {
	// 定时刷新事件到日志的定时器
	syncInterval := time.NewTicker(time.Duration(10) * time.Second)
	defer syncInterval.Stop()

	for {
		select {
		case event := <-el.eventCh:
			if _, ok := subscribeEvents[event.Event()]; !ok {
				break
			}

			el.curEventBuffer.Put(event)

			// 触发持久化到 log 阈值
			if el.curEventBuffer.Size() == defaultBufferSize {
				go el.eventHandler(el.curEventBuffer)
				el.switchEventBuffer()
			}
		case <-syncInterval.C:
			go el.eventHandler(el.curEventBuffer)
			el.switchEventBuffer()
		case <-ctx.Done():
			return
		}
	}
}

// switchEventBuffer 换一个新的 buffer 实例继续使用
func (el *discoverEventLocal) switchEventBuffer() {
	el.curEventBuffer = el.bufferPool.Get().(*eventBufferHolder)
}

// writeToFile 事件落盘
func (el *discoverEventLocal) writeToFile(eventHolder *eventBufferHolder) {
	el.syncLock.Lock()
	defer func() {
		el.syncLock.Unlock()
		eventHolder.Reset()
		el.bufferPool.Put(eventHolder)
	}()

	for eventHolder.HasNext() {
		event := eventHolder.Next()
		log.Info("",
			zap.String("id", event.ID()),
			zap.String("resource", event.Resource()),
			zap.String("happen_time", event.HappenTime().Format(time.DateTime)),
			zap.String("server", utils.LocalHost))
	}
}
