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

package heartbeat

import (
	otelmetric "go.opentelemetry.io/otel/metric"

	"github.com/pole-io/pole-server/pkg/common/otel"
)

var (
	beatRecordCost otelmetric.Float64Histogram
)

const (
	labelAction = "action"
	labelCode   = "code"
)

func registerMetrics() {
	var err error
	beatRecordCost, err = otel.Meter().Float64Histogram("p2p_checker_heartbeat_op",
		otelmetric.WithExplicitBucketBoundaries([]float64{5, 10, 15, 20, 30, 50, 100, 500, 1000, 5000}...))
	if err != nil {
		panic("register heartbeat record cost metric failed: " + err.Error())
	}
}
