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

package metrics

import (
	"context"

	"go.opentelemetry.io/otel/metric"

	"github.com/pole-io/pole-server/pkg/common/otel"
)

func registerClientMetrics() {
	var err error

	// discoveryConnTotal 服务发现客户端链接数量
	discoveryConnTotal, err = otel.Meter().Int64Gauge("discovery_conn_total",
		metric.WithDescription("polaris discovery client connection total"),
	)
	if err != nil {
		panic("register discovery conn total metric failed: " + err.Error())
	}

	// configurationConnTotal 配置中心客户端链接数量
	configurationConnTotal, err = otel.Meter().Int64Gauge("config_conn_total",
		metric.WithDescription("polaris configuration client connection total"),
	)
	if err != nil {
		panic("register configuration conn total metric failed: " + err.Error())
	}

	// sdkClientTotal 客户端链接数量
	sdkClientTotal, err = otel.Meter().Int64Gauge("sdk_client_total",
		metric.WithDescription("polaris client connection total"),
	)
	if err != nil {
		panic("register sdk client total metric failed: " + err.Error())
	}
}

// ReportDiscoveryClientConn report discovery client connection number
func ReportDiscoveryClientConn(v int64) {
	discoveryConnTotal.Record(context.Background(), v)
}

// ResetDiscoveryClientConn reset discovery client connection number
func ResetDiscoveryClientConn() {
	discoveryConnTotal.Record(context.Background(), 0)
}

// ReportConfigurationClientConn report configuration client connection number
func ReportConfigurationClientConn(v int64) {
	configurationConnTotal.Record(context.Background(), v)
}

// ResetConfigurationClientConn reset configuration client connection number
func ResetConfigurationClientConn() {
	configurationConnTotal.Record(context.Background(), 0)
}

// ReportSDKClientConn report SDK client connection number
func ReportSDKClientConn(v int64) {
	sdkClientTotal.Record(context.Background(), v)
}

// Conn reset client connection number
func ResetSDKClientConn() {
	sdkClientTotal.Record(context.Background(), 0)
}
