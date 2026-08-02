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

package history

import (
	"fmt"

	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/pluginapi"
)

var historyRuntime pluginapi.RuntimeValue[History]

// History 历史记录插件
type History interface {
	apis.Plugin
	Record(entry *types.RecordEntry)
}

// GetHistory Get the historical record plugin
func GetHistory() History {
	history, err := historyRuntime.Load(func() (History, error) {
		entries := apis.GetPluginConfig().History.Entries
		composite := &CompositeHistory{
			chain:   make([]History, 0, len(entries)),
			options: entries,
		}

		if err := composite.Initialize(nil); err != nil {
			return nil, err
		}
		return composite, nil
	})
	if err != nil {
		panic(fmt.Errorf("History plugin init err: %s", err.Error()))
	}
	return history
}

type CompositeHistory struct {
	chain   []History
	options []apis.ConfigEntry
}

func (c *CompositeHistory) Name() string {
	return "CompositeHistory"
}

func (c *CompositeHistory) Initialize(config *apis.ConfigEntry) error {
	for i := range c.options {
		entry := c.options[i]
		item, err := apis.ResolvePlugin(apis.PluginTypeHistory, entry.Name)
		if err != nil {
			return fmt.Errorf("resolve History plugin %q: %w", entry.Name, err)
		}

		history, ok := item.(History)
		if !ok {
			return fmt.Errorf("plugin %s not implement History interface", entry.Name)
		}

		if err := history.Initialize(&entry); err != nil {
			return err
		}
		c.chain = append(c.chain, history)
	}
	return nil
}

func (c *CompositeHistory) Destroy() error {
	for i := range c.chain {
		if err := c.chain[i].Destroy(); err != nil {
			return err
		}
	}
	return nil
}

func (c *CompositeHistory) Type() apis.PluginType {
	return apis.PluginTypeHistory
}

func (c *CompositeHistory) Record(entry *types.RecordEntry) {
	for i := range c.chain {
		c.chain[i].Record(entry)
	}
}
