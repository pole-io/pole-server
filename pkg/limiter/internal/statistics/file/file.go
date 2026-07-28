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

package file

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/limiter/internal/statistics"
)

// StaticsWorker 智研上报处理器
type StaticsWorker struct {
	rateLimitCurveReporter *RateLimitCurveReporter
	serverCurveReporter    *ServerCurveReporter
	eventLogReporter       *EventLogReporter
	logStatHandler         *LogStatHandler
	ctx                    context.Context
	cancel                 context.CancelFunc
	// 曲线上报的时间间隔
	interval time.Duration
	// 精度上报的时间间隔
	precisionInterval time.Duration
	reportHandler     ReportHandler
}

// RuntimeIdentity 描述统计记录中的当前 Limiter 实例身份。
type RuntimeIdentity struct {
	ServerAddress    string
	LimitServiceName string
}

// Name 获取统计插件名称
func (s *StaticsWorker) Name() string {
	return "file"
}

// Initialize 初始化统计插件
func (s *StaticsWorker) Initialize(conf *statistics.ConfigEntry) error {
	return s.InitializeContext(context.Background(), conf)
}

// InitializeContext 使用宿主生命周期初始化统计输出。
func (s *StaticsWorker) InitializeContext(ctx context.Context, conf *statistics.ConfigEntry) error {
	return s.InitializeContextWithIdentity(ctx, conf, RuntimeIdentity{
		ServerAddress: "127.0.0.1",
	})
}

// InitializeContextWithIdentity 使用实例级身份初始化统计输出，避免多个嵌入实例共享全局状态。
func (s *StaticsWorker) InitializeContextWithIdentity(
	ctx context.Context, conf *statistics.ConfigEntry, identity RuntimeIdentity,
) error {
	reportConf, err := decodeReportConfig(conf.Option)
	if err != nil {
		return fmt.Errorf("invalid plugin %s option: %w", s.Name(), err)
	}
	s.rateLimitCurveReporter = NewRateLimitCurveReporter(reportConf, identity.LimitServiceName)
	s.serverCurveReporter = NewServerCurveReporter(reportConf, identity.LimitServiceName)
	s.eventLogReporter = NewEventLogReporter(reportConf)
	s.reportHandler = NewReportHandler(reportConf, identity.ServerAddress)
	s.logStatHandler = NewLogStatHandler(reportConf)
	s.interval = time.Duration(reportConf.LogInterval) * time.Second
	s.precisionInterval = time.Duration(reportConf.PrecisionLogInterval) * time.Second
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.Start(s.ctx)
	return nil
}

func decodeReportConfig(option map[string]interface{}) (*ReportConfig, error) {
	text, err := json.Marshal(option)
	if nil != err {
		return nil, fmt.Errorf("marshal option: %w", err)
	}
	reportConf := &ReportConfig{}
	decoder := json.NewDecoder(bytes.NewReader(text))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(reportConf); err != nil {
		return nil, fmt.Errorf("decode option: %w", err)
	}
	if err = reportConf.Validate(); err != nil {
		return nil, fmt.Errorf("validate option: %w", err)
	}
	return reportConf, nil
}

// Start 启动调度
func (s *StaticsWorker) Start(ctx context.Context) {
	// 启动上报协程
	go func() {
		ticker := time.NewTicker(s.interval)
		precisionTicker := time.NewTicker(s.precisionInterval)
		defer func() {
			ticker.Stop()
			precisionTicker.Stop()
		}()
		for {
			select {
			case <-ctx.Done():
				log.Infof("file statis loop timer routine exists")
				return
			case <-precisionTicker.C:
				startTime := time.Now()
				statValues := s.rateLimitCurveReporter.MergeAllStatValues(false)
				total := s.logStatHandler.LogPrecisionRecord(statValues)
				totalTime := time.Since(startTime)
				if total > 0 && totalTime >= 800*time.Millisecond {
					// 因为每秒数量比较多，为避免刷屏，只在存在精度数据时候进行打印
					log.Infof("time consume for log precision is %v, item count is %d", totalTime, total)
				}
			case <-ticker.C:
				startTime := time.Now()
				srvRecord := s.serverCurveReporter.BuildReportRecord()
				if srvRecord.HasTags() {
					s.reportHandler.Report(srvRecord)
				}
				rateLimitRecord := s.rateLimitCurveReporter.BuildReportRecord()
				if rateLimitRecord.HasTags() {
					s.reportHandler.Report(rateLimitRecord)
				}
				totalItemCount := len(srvRecord.Tags) + len(rateLimitRecord.Tags)
				totalTime := time.Since(startTime)
				log.Infof("time consume for report is %v, item count is %d", totalTime, totalItemCount)

				startTime = time.Now()
				total := s.eventLogReporter.LogAllEvents()
				totalTime = time.Since(startTime)
				log.Infof("time consume for log event is %v, item count is %d", totalTime, total)

			}
		}
	}()
	log.Infof("file statis loop has started")
}

// Destroy 销毁统计插件
func (s *StaticsWorker) Destroy() error {
	if nil != s.cancel {
		s.cancel()
	}
	return nil
}

// AddAPICall 服务方法调用结果反馈，含有规则的计算周期
func (s *StaticsWorker) AddAPICall(value statistics.APICallStatValue) {
	s.serverCurveReporter.AddIncrement(value)
}

// CreateRateLimitStatCollectorV1 创建采集器V1，每个stream上来后获取一次
func (s *StaticsWorker) CreateRateLimitStatCollectorV1() *statistics.RateLimitStatCollectorV1 {
	return s.rateLimitCurveReporter.CreateCollectorV1()
}

// CreateRateLimitStatCollectorV2 创建采集器V2，每个stream上来后获取一次
func (s *StaticsWorker) CreateRateLimitStatCollectorV2() *statistics.RateLimitStatCollectorV2 {
	return s.rateLimitCurveReporter.CreateCollectorV2()
}

// DropRateLimitStatCollector 归还采集器
func (s *StaticsWorker) DropRateLimitStatCollector(value statistics.RateLimitStatCollector) {
	s.rateLimitCurveReporter.DropCollector(value)
}

// AddEventToLog 添加事件
func (s *StaticsWorker) AddEventToLog(value statistics.EventToLog) {
	s.eventLogReporter.AddEvent(value)
}
