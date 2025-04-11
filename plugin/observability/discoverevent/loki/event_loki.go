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

package loki

import (
	"time"

	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/apis/observability/event"
)

const (
	PluginName       = "discoverEventLoki"
	defaultBatchSize = 512
	defaultQueueSize = 1024
)

type discoverEventLoki struct {
	eventCh  chan event.DiscoverEvent
	stopCh   chan struct{}
	eventLog *LokiLogger
}

// Name 插件名称
// @return string 返回插件名称
func (d *discoverEventLoki) Name() string {
	return PluginName
}

// Initialize 根据配置文件进行初始化插件 discoverEventLoki
// @param conf 配置文件内容
// @return error 初始化失败，返回 error 信息
func (d *discoverEventLoki) Initialize(conf *apis.ConfigEntry) error {
	var queueSize = defaultQueueSize
	if val, ok := conf.Option["queueSize"]; ok {
		queueSize, _ = val.(int)
	}
	lokiLogger, err := newLokiLogger(conf.Option)
	if err != nil {
		return err
	}
	d.eventLog = lokiLogger
	d.eventCh = make(chan event.DiscoverEvent, queueSize)
	d.stopCh = make(chan struct{}, 1)
	go d.Run()
	return nil
}

// Destroy 执行插件销毁
func (d *discoverEventLoki) Destroy() error {
	close(d.stopCh)
	return nil
}

// PublishEvent 发布一个服务事件
func (d *discoverEventLoki) PublishEvent(event event.DiscoverEvent) {
	select {
	case d.eventCh <- event:
		return
	default:
		// do nothing
	}
}

func (d *discoverEventLoki) Type() apis.PluginType {
	return apis.PluginTypeDiscoverEvent
}

// Run 执行主逻辑
func (d *discoverEventLoki) Run() {
	// 定时刷新事件到日志的定时器
	syncInterval := time.NewTicker(time.Duration(10) * time.Second)
	defer syncInterval.Stop()

	batch := make([]event.DiscoverEvent, 0, defaultBatchSize)
	batchSize := 0

	for {
		select {
		case e := <-d.eventCh:
			// 确保事件是顺序的
			batch = append(batch, e)
			batchSize++
			// 触发批量生产发送 log 阈值
			if batchSize == defaultBatchSize {
				d.eventLog.Log(batch[:batchSize])
				batch = make([]event.DiscoverEvent, 0, defaultBatchSize)
				batchSize = 0
			}
		case <-syncInterval.C:
			if batchSize > 0 {
				d.eventLog.Log(batch[:batchSize])
				batch = make([]event.DiscoverEvent, 0, defaultBatchSize)
				batchSize = 0
			}
		case <-d.stopCh:
			if batchSize > 0 {
				d.eventLog.Log(batch[:batchSize])
			}
			return
		}
	}
}
