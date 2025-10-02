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

package event

import (
	"fmt"
	"sync"
	"time"

	"github.com/pole-io/pole-server/apis"
)

var (
	discoverEventOnce sync.Once
	_discoverChannel  DiscoverChannel
)

type DiscoverEvent interface {
	// 资源 ID
	ID() string
	// 事件类型
	Event() string
	// 资源信息
	Resource() string
	// 发生时间
	HappenTime() time.Time
}

// DiscoverChannel is used to receive discover events from the agent
type DiscoverChannel interface {
	apis.Plugin
	// PublishEvent Release a service event
	PublishEvent(event DiscoverEvent)
}

// GetDiscoverEvent Get service discovery event plug -in
func GetDiscoverEvent() DiscoverChannel {
	if _discoverChannel != nil {
		return _discoverChannel
	}

	discoverEventOnce.Do(func() {
		var (
			entries []apis.ConfigEntry
		)

		entries = append(entries, apis.GetPluginConfig().DiscoverEvent.Entries...)

		_discoverChannel = newCompositeDiscoverChannel(entries)
		if err := _discoverChannel.Initialize(nil); err != nil {
			panic(fmt.Errorf("DiscoverChannel plugin init err: %s", err.Error()))
		}
	})

	return _discoverChannel
}

// newCompositeDiscoverChannel creates Composite DiscoverChannel
func newCompositeDiscoverChannel(options []apis.ConfigEntry) *compositeDiscoverChannel {
	return &compositeDiscoverChannel{
		chain:   make([]DiscoverChannel, 0, len(options)),
		options: options,
	}
}

// compositeDiscoverChannel is used to receive discover events from the agent
type compositeDiscoverChannel struct {
	chain   []DiscoverChannel
	options []apis.ConfigEntry
}

func (c *compositeDiscoverChannel) Name() string {
	return "CompositeDiscoverChannel"
}

func (c *compositeDiscoverChannel) Initialize(config *apis.ConfigEntry) error {
	for i := range c.options {
		entry := c.options[i]
		item, exist := apis.GetPlugin(apis.PluginTypeDiscoverEvent, entry.Name)
		if !exist {
			panic(fmt.Errorf("plugin DiscoverChannel not found target: %s", entry.Name))
		}

		discoverChannel, ok := item.(DiscoverChannel)
		if !ok {
			panic(fmt.Errorf("plugin target: %s not DiscoverChannel", entry.Name))
		}

		if err := discoverChannel.Initialize(&entry); err != nil {
			return err
		}
		c.chain = append(c.chain, discoverChannel)
	}
	return nil
}

func (c *compositeDiscoverChannel) Destroy() error {
	for i := range c.chain {
		if err := c.chain[i].Destroy(); err != nil {
			return err
		}
	}
	return nil
}

// PublishEvent Release a service event
func (c *compositeDiscoverChannel) PublishEvent(event DiscoverEvent) {
	for i := range c.chain {
		c.chain[i].PublishEvent(event)
	}
}

func (c *compositeDiscoverChannel) Type() apis.PluginType {
	return apis.PluginTypeDiscoverEvent
}
