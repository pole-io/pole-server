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

package statis

import (
	"fmt"
	"sync"

	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/apis/pkg/types/metrics"
)

var (
	statisOnce sync.Once
	_statis    Statis
)

// Statis Statistical plugin interface
type Statis interface {
	apis.Plugin
	// ReportCallMetrics report call metrics info
	ReportCallMetrics(metric metrics.CallMetric)
	// ReportDiscoveryMetrics report discovery metrics
	ReportDiscoveryMetrics(metric ...metrics.DiscoveryMetric)
	// ReportConfigMetrics report config_center metrics
	ReportConfigMetrics(metric ...metrics.ConfigMetrics)
	// ReportDiscoverCall report discover service times
	ReportDiscoverCall(metric metrics.ClientDiscoverMetric)
}

// compositeStatis is used to receive discover events from the agent
type compositeStatis struct {
	chain   []Statis
	options []apis.ConfigEntry
}

func (c *compositeStatis) Name() string {
	return "compositeStatis"
}

func (c *compositeStatis) Initialize(config *apis.ConfigEntry) error {
	for i := range c.options {
		entry := c.options[i]
		item, exist := apis.GetPlugin(apis.PluginTypeStatis, entry.Name)
		if !exist {
			return fmt.Errorf("plugin Statis not found target: %s", entry.Name)
		}

		statis, ok := item.(Statis)
		if !ok {
			return fmt.Errorf("plugin target: %s not Statis", entry.Name)
		}

		if err := statis.Initialize(&entry); err != nil {
			return err
		}
		c.chain = append(c.chain, statis)
	}
	return nil
}

func (c *compositeStatis) Type() apis.PluginType {
	return apis.PluginTypeStatis
}

func (c *compositeStatis) Destroy() error {
	for i := range c.chain {
		if err := c.chain[i].Destroy(); err != nil {
			return err
		}
	}
	return nil
}

// ReportCallMetrics report call metrics info
func (c *compositeStatis) ReportCallMetrics(metric metrics.CallMetric) {
	for i := range c.chain {
		c.chain[i].ReportCallMetrics(metric)
	}
}

// ReportDiscoveryMetrics report discovery metrics
func (c *compositeStatis) ReportDiscoveryMetrics(metric ...metrics.DiscoveryMetric) {
	for i := range c.chain {
		c.chain[i].ReportDiscoveryMetrics(metric...)
	}
}

// ReportConfigMetrics report config_center metrics
func (c *compositeStatis) ReportConfigMetrics(metric ...metrics.ConfigMetrics) {
	for i := range c.chain {
		c.chain[i].ReportConfigMetrics(metric...)
	}
}

// ReportDiscoverCall report discover service times
func (c *compositeStatis) ReportDiscoverCall(metric metrics.ClientDiscoverMetric) {
	for i := range c.chain {
		c.chain[i].ReportDiscoverCall(metric)
	}
}

// GetStatis Get statistical plugin
func GetStatis() Statis {
	if _statis != nil {
		return _statis
	}

	statisOnce.Do(func() {
		var (
			entries        []apis.ConfigEntry
			defaultEntries = []apis.ConfigEntry{
				{
					Name: "local",
				},
				{
					Name: "prometheus",
				},
			}
		)

		if len(apis.GetPluginConfig().Statis.Entries) != 0 {
			entries = append(entries, apis.GetPluginConfig().Statis.Entries...)
		} else {
			if apis.GetPluginConfig().Statis.Name == "local" {
				entries = defaultEntries
			} else {
				entries = append(entries, apis.ConfigEntry{
					Name:   apis.GetPluginConfig().Statis.Name,
					Option: apis.GetPluginConfig().Statis.Option,
				})
			}
		}

		_statis = &compositeStatis{
			chain:   []Statis{},
			options: entries,
		}
		if err := _statis.Initialize(nil); err != nil {
			panic(fmt.Sprintf("Statis plugin init err: %s", err.Error()))
		}
	})

	return _statis
}
