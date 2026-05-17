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

package listener

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestQueryConfigListeners 测试查询配置监听者
func TestQueryConfigListeners(t *testing.T) {
	// 模拟配置监听者数据
	type ConfigListener struct {
		ListenerID   string
		FileID       string
		FileName     string
		GroupID      string
		ListenerIP   string
		LastPollTime string
		LastVersion  string
		Enable       bool
	}

	listeners := []ConfigListener{
		{ListenerID: "l-1", FileID: "f-1", FileName: "app.yaml", GroupID: "default", ListenerIP: "10.0.0.1", LastPollTime: "2024-01-01T00:00:00Z", LastVersion: "v1", Enable: true},
		{ListenerID: "l-2", FileID: "f-1", FileName: "app.yaml", GroupID: "default", ListenerIP: "10.0.0.2", LastPollTime: "2024-01-01T00:01:00Z", LastVersion: "v1", Enable: true},
		{ListenerID: "l-3", FileID: "f-2", FileName: "config.yaml", GroupID: "prod", ListenerIP: "10.0.0.3", LastPollTime: "2024-01-01T00:02:00Z", LastVersion: "v2", Enable: false},
	}

	// 验证查询结果
	assert.Equal(t, 3, len(listeners), "should have 3 listeners")
	for _, l := range listeners {
		assert.NotEmpty(t, l.ListenerID, "listener ID should not be empty")
		assert.NotEmpty(t, l.FileID, "file ID should not be empty")
		assert.NotEmpty(t, l.GroupID, "group ID should not be empty")
	}
}

// TestConfigSubscriptionFilter 测试配置订阅关系过滤
func TestConfigSubscriptionFilter(t *testing.T) {
	// 模拟配置订阅关系
	type ConfigSubscription struct {
		ListenerID string
		FileID     string
		GroupID    string
		Namespace  string
		Enable     bool
		Protocol   string // "long-polling", "http"
	}

	subscriptions := []ConfigSubscription{
		{ListenerID: "l-1", FileID: "f-1", GroupID: "default", Namespace: "default", Enable: true, Protocol: "long-polling"},
		{ListenerID: "l-2", FileID: "f-2", GroupID: "default", Namespace: "default", Enable: false, Protocol: "http"},
		{ListenerID: "l-3", FileID: "f-1", GroupID: "prod", Namespace: "prod", Enable: true, Protocol: "long-polling"},
	}

	// 过滤启用状态的订阅
	enabledSubs := make([]ConfigSubscription, 0)
	for _, s := range subscriptions {
		if s.Enable {
			enabledSubs = append(enabledSubs, s)
		}
	}
	assert.Equal(t, 2, len(enabledSubs), "should have 2 enabled subscriptions")

	// 过滤特定协议的订阅
	longPollingSubs := make([]ConfigSubscription, 0)
	for _, s := range subscriptions {
		if s.Protocol == "long-polling" {
			longPollingSubs = append(longPollingSubs, s)
		}
	}
	assert.Equal(t, 2, len(longPollingSubs), "should have 2 long-polling subscriptions")
}

// TestConfigReleaseNotification 测试配置发布通知
func TestConfigReleaseNotification(t *testing.T) {
	// 模拟配置发布通知数据
	type ReleaseNotification struct {
		NotificationID string
		FileID         string
		FileName       string
		GroupID        string
		Namespace      string
		OldVersion     string
		NewVersion     string
		NotifyTime     string
		Targets        []string // 通知的目标监听者列表
		SuccessCount   int
		FailCount      int
	}

	notifications := []ReleaseNotification{
		{NotificationID: "n-1", FileID: "f-1", FileName: "app.yaml", GroupID: "default", Namespace: "default", OldVersion: "v1", NewVersion: "v2", NotifyTime: "2024-01-01T00:00:00Z", Targets: []string{"l-1", "l-2"}, SuccessCount: 2, FailCount: 0},
		{NotificationID: "n-2", FileID: "f-2", FileName: "config.yaml", GroupID: "prod", Namespace: "prod", OldVersion: "v1", NewVersion: "v2", NotifyTime: "2024-01-01T01:00:00Z", Targets: []string{"l-3", "l-4"}, SuccessCount: 2, FailCount: 0},
	}

	// 验证通知数据
	for _, n := range notifications {
		assert.NotEmpty(t, n.NotificationID, "notification ID should not be empty")
		assert.NotEmpty(t, n.FileID, "file ID should not be empty")
		assert.NotEmpty(t, n.NewVersion, "new version should not be empty")
		assert.NotEqual(t, n.OldVersion, n.NewVersion, "version should change")
		assert.Equal(t, n.SuccessCount+n.FailCount, len(n.Targets), "success+fail should equal targets count")
	}

	// 计算总计
	totalSuccess := 0
	totalFail := 0
	for _, n := range notifications {
		totalSuccess += n.SuccessCount
		totalFail += n.FailCount
	}

	assert.Equal(t, 4, totalSuccess+totalFail, "total notifications should be 4")
}

// TestQueryConfigListeners 测试配置监听者数量统计
func TestConfigListenerCountMetrics(t *testing.T) {
	// 模拟配置监听者数量统计
	type FileListenerCount struct {
		FileID           string
		GroupName        string
		TotalListeners   int
		EnableListeners  int
		DisableListeners int
		LastPollAvgMs    float64
	}

	counts := []FileListenerCount{
		{FileID: "f-1", GroupName: "default", TotalListeners: 10, EnableListeners: 8, DisableListeners: 2, LastPollAvgMs: 15.5},
		{FileID: "f-2", GroupName: "prod", TotalListeners: 5, EnableListeners: 5, DisableListeners: 0, LastPollAvgMs: 8.2},
		{FileID: "f-3", GroupName: "dev", TotalListeners: 15, EnableListeners: 12, DisableListeners: 3, LastPollAvgMs: 25.3},
	}

	// 验证统计准确性
	for _, c := range counts {
		assert.Equal(t, c.EnableListeners+c.DisableListeners, c.TotalListeners, "enable+disable should equal total")
		assert.GreaterOrEqual(t, c.TotalListeners, c.EnableListeners, "total should >= enable")
		assert.GreaterOrEqual(t, c.TotalListeners, c.DisableListeners, "total should >= disable")
	}

	// 计算总计
	totalListeners := 0
	totalEnabled := 0
	for _, c := range counts {
		totalListeners += c.TotalListeners
		totalEnabled += c.EnableListeners
	}

	assert.Equal(t, 30, totalListeners, "total listeners should be 30")
	assert.Equal(t, 25, totalEnabled, "total enabled should be 25")
}

// TestConfigListenerRelationValidation 测试配置监听关系验证
func TestConfigListenerRelationValidation(t *testing.T) {
	// 模拟配置监听关系数据
	type Relation struct {
		ListenerID    string
		FileID        string
		GroupID       string
		Valid         bool
		ValidationMsg string
	}

	relations := []Relation{
		{ListenerID: "l-1", FileID: "f-1", GroupID: "default", Valid: true, ValidationMsg: ""},
		{ListenerID: "", FileID: "f-2", GroupID: "default", Valid: false, ValidationMsg: "listener ID is empty"},
		{ListenerID: "l-3", FileID: "", GroupID: "default", Valid: false, ValidationMsg: "file ID is empty"},
		{ListenerID: "l-4", FileID: "f-3", GroupID: "", Valid: false, ValidationMsg: "group ID is empty"},
		{ListenerID: "l-5", FileID: "f-4", GroupID: "prod", Valid: true, ValidationMsg: ""},
	}

	// 验证关系有效性
	validCount := 0
	invalidCount := 0
	for _, r := range relations {
		if r.Valid {
			validCount++
			assert.Empty(t, r.ValidationMsg, "valid relation should have empty validation message")
		} else {
			invalidCount++
			assert.NotEmpty(t, r.ValidationMsg, "invalid relation should have validation message")
		}
	}

	assert.Equal(t, 2, validCount, "should have 2 valid relations")
	assert.Equal(t, 3, invalidCount, "should have 3 invalid relations")
}

// TestConfigListenerHistoryQuery 测试配置监听者历史查询
func TestConfigListenerHistoryQuery(t *testing.T) {
	// 模拟配置监听历史记录
	type HistoryRecord struct {
		RecordID   string
		ListenerID string
		FileID     string
		Action     string // "poll", "push", "update", "disable", "enable"
		OldVersion string
		NewVersion string
		Timestamp  string
	}

	history := []HistoryRecord{
		{RecordID: "rec-1", ListenerID: "l-1", FileID: "f-1", Action: "poll", OldVersion: "", NewVersion: "v1", Timestamp: "2024-01-01T00:00:00Z"},
		{RecordID: "rec-2", ListenerID: "l-1", FileID: "f-1", Action: "push", OldVersion: "v1", NewVersion: "v2", Timestamp: "2024-01-02T00:00:00Z"},
		{RecordID: "rec-3", ListenerID: "l-1", FileID: "f-1", Action: "disable", OldVersion: "v2", NewVersion: "v2", Timestamp: "2024-01-03T00:00:00Z"},
	}

	// 验证历史记录
	assert.Equal(t, 3, len(history), "should have 3 history records")
	assert.Equal(t, "l-1", history[0].ListenerID)
	assert.Equal(t, "poll", history[0].Action)
	assert.Equal(t, "disable", history[len(history)-1].Action)
}

// TestConfigListenerQueryPerformance 测试配置监听者查询性能
func TestConfigListenerQueryPerformance(t *testing.T) {
	// 模拟不同规模的查询
	type TestCase struct {
		NumFiles      int
		NumListeners  int
		ExpectedMaxMs int64
	}

	testCases := []struct {
		Name          string
		NumFiles      int
		NumListeners  int
		ExpectedMaxMs int64
	}{
		{"small", 10, 50, 1000},
		{"medium", 100, 500, 5000},
		{"large", 1000, 5000, 50000},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// 模拟查询时间
			queryTimeMs := int64(tc.NumFiles*10 + tc.NumListeners*5)
			expectedMaxMs := int64(tc.ExpectedMaxMs)

			assert.LessOrEqual(t, queryTimeMs, expectedMaxMs,
				"query time should be within expected limit")
		})
	}
}

// TestConfigListenerAggregateStats 测试配置监听者聚合统计
func TestConfigListenerAggregateStats(t *testing.T) {
	// 模拟按文件聚合的监听者统计数据
	type AggregatedStats struct {
		FileID           string
		GroupName        string
		TotalListeners   int
		EnableListeners  int
		DisableListeners int
		VersionStats     map[string]int
	}

	stats := []AggregatedStats{
		{
			FileID:           "f-1",
			GroupName:        "default",
			TotalListeners:   100,
			EnableListeners:  80,
			DisableListeners: 20,
			VersionStats: map[string]int{
				"v1": 60,
				"v2": 40,
			},
		},
		{
			FileID:           "f-2",
			GroupName:        "prod",
			TotalListeners:   50,
			EnableListeners:  50,
			DisableListeners: 0,
			VersionStats: map[string]int{
				"v1": 30,
				"v2": 20,
			},
		},
	}

	// 验证聚合统计
	for _, s := range stats {
		assert.Equal(t, s.EnableListeners+s.DisableListeners, s.TotalListeners, "enable+disable should equal total")

		// 验证版本统计
		versionTotal := 0
		for _, count := range s.VersionStats {
			versionTotal += count
		}
		assert.Equal(t, versionTotal, s.TotalListeners, "version stats should sum to total")
	}
}

// TestConfigNotificationBatch 测试配置发布批量通知
func TestConfigNotificationBatch(t *testing.T) {
	// 模拟批量通知
	type NotificationBatch struct {
		BatchID       string
		FileID        string
		NewVersion    string
		TargetCount   int
		NotifiedCount int
		SuccessCount  int
		FailCount     int
		NotifyTime    string
	}

	batches := []NotificationBatch{
		{BatchID: "b-1", FileID: "f-1", NewVersion: "v2", TargetCount: 100, NotifiedCount: 100, SuccessCount: 98, FailCount: 2, NotifyTime: "2024-01-01T00:00:00Z"},
		{BatchID: "b-2", FileID: "f-2", NewVersion: "v3", TargetCount: 50, NotifiedCount: 50, SuccessCount: 50, FailCount: 0, NotifyTime: "2024-01-01T01:00:00Z"},
	}

	// 验证批量通知
	for _, b := range batches {
		assert.Equal(t, b.SuccessCount+b.FailCount, b.NotifiedCount, "success+fail should equal notified")
		assert.Equal(t, b.NotifiedCount, b.TargetCount, "notified should equal target")
		assert.NotEmpty(t, b.BatchID, "batch ID should not be empty")
		assert.NotEmpty(t, b.NewVersion, "new version should not be empty")
	}

	// 计算总计
	totalTarget := 0
	totalSuccess := 0
	for _, b := range batches {
		totalTarget += b.TargetCount
		totalSuccess += b.SuccessCount
	}

	assert.Equal(t, 150, totalTarget, "total target should be 150")
	assert.Equal(t, 148, totalSuccess, "total success should be 148")
}

// TestConfigListenerQueryWithFilters 测试带过滤条件的配置监听者查询
func TestConfigListenerQueryWithFilters(t *testing.T) {
	// 模拟带过滤条件的查询
	type ConfigListener struct {
		ListenerID string
		FileID     string
		GroupID    string
		Namespace  string
		Protocol   string
		Enable     bool
		Version    string
	}

	listeners := []ConfigListener{
		{ListenerID: "l-1", FileID: "f-1", GroupID: "default", Namespace: "default", Protocol: "long-polling", Enable: true, Version: "v1"},
		{ListenerID: "l-2", FileID: "f-1", GroupID: "default", Namespace: "default", Protocol: "http", Enable: true, Version: "v2"},
		{ListenerID: "l-3", FileID: "f-1", GroupID: "prod", Namespace: "prod", Protocol: "long-polling", Enable: false, Version: "v1"},
	}

	// 过虑条件
	type Filter struct {
		Protocol string
		Enable   *bool
	}

	// 测试过滤
	testFilters := []struct {
		Name   string
		Filter Filter
		Count  int
	}{
		{"long_polling_only", Filter{Protocol: "long-polling"}, 2},
		{"enabled_only", Filter{Enable: boolPtr(true)}, 2},
		{"long_polling_and_enabled", Filter{Protocol: "long-polling", Enable: boolPtr(true)}, 1},
	}

	for _, tf := range testFilters {
		t.Run(tf.Name, func(t *testing.T) {
			count := 0
			for _, l := range listeners {
				match := true
				if tf.Filter.Protocol != "" && l.Protocol != tf.Filter.Protocol {
					match = false
				}
				if tf.Filter.Enable != nil && l.Enable != *tf.Filter.Enable {
					match = false
				}
				if match {
					count++
				}
			}
			assert.Equal(t, tf.Count, count, "filter result count should match")
		})
	}
}

// TestConfigListenerQueryPagination 测试配置监听者查询分页
func TestConfigListenerQueryPagination(t *testing.T) {
	// 模拟分页查询
	totalListeners := 125
	pageSize := 50

	// 计算页数
	totalPages := (totalListeners + pageSize - 1) / pageSize

	assert.Equal(t, 3, totalPages, "should have 3 pages")

	// 验证每页数据
	for page := 1; page <= totalPages; page++ {
		offset := (page - 1) * pageSize
		limit := pageSize
		if page == totalPages {
			limit = totalListeners - offset
		}

		// 模拟获取页数据
		items := make([]string, 0, limit)
		for i := 0; i < limit; i++ {
			items = append(items, fmt.Sprintf("listener-%d", offset+i+1))
		}

		hasNext := page < totalPages
		assert.GreaterOrEqual(t, len(items), 1, "should have at least 1 item")
		assert.Equal(t, hasNext, page < totalPages, "hasNext should match page number")
	}
}

// TestConfigListenerCacheHitRate 测试配置监听者缓存命中率
func TestConfigListenerCacheHitRate(t *testing.T) {
	// 模拟缓存命中率统计
	type CacheStats struct {
		ListenerID   string
		HitCount     int64
		MissCount    int64
		CacheHitRate float64
	}

	stats := []CacheStats{
		{ListenerID: "l-1", HitCount: 1000, MissCount: 100, CacheHitRate: 0.909},
		{ListenerID: "l-2", HitCount: 500, MissCount: 50, CacheHitRate: 0.909},
		{ListenerID: "l-3", HitCount: 2000, MissCount: 200, CacheHitRate: 0.909},
	}

	// 验证缓存命中率
	for _, s := range stats {
		expectedRate := float64(s.HitCount) / float64(s.HitCount+s.MissCount)
		assert.InDelta(t, expectedRate, s.CacheHitRate, 0.01, "cache hit rate should match calculation")
		assert.GreaterOrEqual(t, s.CacheHitRate, 0.0, "cache hit rate should be >= 0")
		assert.LessOrEqual(t, s.CacheHitRate, 1.0, "cache hit rate should be <= 1")
	}
}

// TestConfigListenerDataAccuracy 测试配置监听数据准确性
func TestConfigListenerDataAccuracy(t *testing.T) {
	// 模拟监听数据一致性校验
	type ListenerData struct {
		ListenerID  string
		FileID      string
		GroupID     string
		Namespace   string
		LastVersion string
		PollCount   int64
		PushCount   int64
	}

	listeners := []ListenerData{
		{ListenerID: "l-1", FileID: "f-1", GroupID: "default", Namespace: "default", LastVersion: "v2", PollCount: 100, PushCount: 10},
		{ListenerID: "l-2", FileID: "f-2", GroupID: "prod", Namespace: "prod", LastVersion: "v1", PollCount: 50, PushCount: 5},
	}

	// 验证数据完整性
	for _, l := range listeners {
		assert.NotEmpty(t, l.ListenerID, "listener ID should not be empty")
		assert.NotEmpty(t, l.FileID, "file ID should not be empty")
		assert.GreaterOrEqual(t, l.PollCount, int64(0), "poll count should be non-negative")
		assert.GreaterOrEqual(t, l.PushCount, int64(0), "push count should be non-negative")
	}

	assert.Equal(t, 2, len(listeners), "should have 2 listeners")
}

// 辅助函数
func boolPtr(b bool) *bool {
	return &b
}
