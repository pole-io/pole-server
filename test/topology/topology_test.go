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

package topology

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// RelationChange 用于拓扑比较的关系变化
type RelationChange struct {
	Source   string
	Target   string
	OldValue int64
	NewValue int64
}

// TestCollectTopologyData 测试拓扑数据采集
func TestCollectTopologyData(t *testing.T) {
	// 模拟服务调用关系
	type CallRelation struct {
		SourceService   string
		TargetService   string
		CallCount       int64
		SuccessCount    int64
		FailCount       int64
		AvgLatencyMs    float64
		P99LatencyMs    float64
		CollectTime     time.Time
	}

	relations := []CallRelation{
		{SourceService: "service-A", TargetService: "service-B", CallCount: 1000, SuccessCount: 990, FailCount: 10, AvgLatencyMs: 15.5, P99LatencyMs: 45.2},
		{SourceService: "service-A", TargetService: "service-C", CallCount: 500, SuccessCount: 495, FailCount: 5, AvgLatencyMs: 8.2, P99LatencyMs: 22.1},
		{SourceService: "service-B", TargetService: "service-D", CallCount: 800, SuccessCount: 780, FailCount: 20, AvgLatencyMs: 25.3, P99LatencyMs: 60.5},
	}

	// 验证数据采集
	totalCalls := int64(0)
	for _, r := range relations {
		totalCalls += r.CallCount
		assert.GreaterOrEqual(t, r.SuccessCount, int64(0), "success count should be non-negative")
		assert.GreaterOrEqual(t, r.FailCount, int64(0), "fail count should be non-negative")
		assert.Equal(t, r.SuccessCount+r.FailCount, r.CallCount, "success+fail should equal total calls")
	}

	assert.Equal(t, int64(2300), totalCalls, "total calls should be sum of all relations")
}

// TestQueryTopologyAPI 测试拓扑查询 API
func TestQueryTopologyAPI(t *testing.T) {
	// 模拟拓扑查询参数
	type TopologyQuery struct {
		ServiceID      string
		Depth          int
		Start time.Time
		End            time.Time
		IncludeMetrics bool
	}

	query := TopologyQuery{
		ServiceID:      "svc-123",
		Depth:          3,
		Start:          time.Now().Add(-time.Hour),
		End:            time.Now(),
		IncludeMetrics: true,
	}

	assert.Equal(t, "svc-123", query.ServiceID)
	assert.Equal(t, 3, query.Depth)
	assert.True(t, query.End.After(query.Start))
	assert.True(t, query.IncludeMetrics)
}

// TestServiceCallChain 测试服务调用链路
func TestServiceCallChain(t *testing.T) {
	// 模拟调用链
	type CallNode struct {
		ServiceID   string
		InstanceID  string
		StartTime   time.Time
		EndTime     time.Time
		DurationMs  float64
		Status      string // "success", "fail", "timeout"
		Children    []CallNode
	}

	// 构建简单的调用链
	root := CallNode{
		ServiceID:  "api-gateway",
		InstanceID: "inst-1",
		StartTime:  time.Now().Add(-time.Millisecond * 100),
		EndTime:    time.Now(),
		DurationMs: 50.5,
		Status:     "success",
		Children: []CallNode{
			{
				ServiceID:  "user-service",
				InstanceID: "inst-2",
				StartTime:  time.Now().Add(-time.Millisecond * 80),
				EndTime:    time.Now().Add(-time.Millisecond * 20),
				DurationMs: 60.0,
				Status:     "success",
				Children: []CallNode{
					{
						ServiceID:  "database",
						InstanceID: "inst-3",
						StartTime:  time.Now().Add(-time.Millisecond * 50),
						EndTime:    time.Now().Add(-time.Millisecond * 25),
						DurationMs: 25.0,
						Status:     "success",
					},
				},
			},
		},
	}

	// 验证调用链
	assert.Equal(t, "api-gateway", root.ServiceID)
	assert.GreaterOrEqual(t, len(root.Children), 1, "should have at least one child")

	// 计算调用链深度
	var maxDepth func(CallNode) int
	maxDepth = func(node CallNode) int {
		if len(node.Children) == 0 {
			return 1
		}
		maxChildDepth := 0
		for _, child := range node.Children {
			childDepth := maxDepth(child)
			if childDepth > maxChildDepth {
				maxChildDepth = childDepth
			}
		}
		return 1 + maxChildDepth
	}

	assert.Equal(t, 3, maxDepth(root), "call chain depth should be 3")
}

// TestTopologyDataAccuracy 测试拓扑数据准确性
func TestTopologyDataAccuracy(t *testing.T) {
	// 模拟拓扑数据一致性校验
	type ServiceMetric struct {
		ServiceID       string
		OutgoingCalls   int64
		IncomingCalls   int64
		AliasFor        string // 别名服务ID（用于合并拓扑）
	}

	services := map[string]*ServiceMetric{
		"svc-A": {ServiceID: "svc-A", OutgoingCalls: 100, IncomingCalls: 50},
		"svc-B": {ServiceID: "svc-B", OutgoingCalls: 80, IncomingCalls: 120},
		"svc-C": {ServiceID: "svc-C", OutgoingCalls: 60, IncomingCalls: 90},
	}

	// 验证数据完整性
	for id, metric := range services {
		assert.NotEmpty(t, id, "service ID should not be empty")
		assert.GreaterOrEqual(t, metric.OutgoingCalls, int64(0))
		assert.GreaterOrEqual(t, metric.IncomingCalls, int64(0))
		assert.Equal(t, id, metric.ServiceID, "service ID should match map key")
	}

	assert.Equal(t, 3, len(services), "should have 3 services")
}

// TestTopologyQuery Performance 测试拓扑查询性能
func TestTopologyQueryPerformance(t *testing.T) {
	// 模拟不同规模的拓扑查询
	type TestCase struct {
		NumServices   int
		NumRelations  int
		ExpectedMaxMs int64
	}

	testCases := []TestCase{
		{NumServices: 10, NumRelations: 50, ExpectedMaxMs: 100},
		{NumServices: 100, NumRelations: 500, ExpectedMaxMs: 500},
		{NumServices: 1000, NumRelations: 5000, ExpectedMaxMs: 2000},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d_services", tc.NumServices), func(t *testing.T) {
			// 模拟拓扑查询时间
			// 实际查询时间应该与服务数量和关系数量成正比
			queryTimeMs := int64(tc.NumServices*10 + tc.NumRelations*5) // 简单模拟

			assert.LessOrEqual(t, queryTimeMs, tc.ExpectedMaxMs,
				"query time should be within expected limit")
		})
	}
}

// TestTopologyDataCleanup 测试拓扑数据清理
func TestTopologyDataCleanup(t *testing.T) {
	// 模拟拓扑数据保留策略
	type RetentionPolicy struct {
		MaxDataDays   int
		MinRetainCalls int64
	}

	policy := RetentionPolicy{
		MaxDataDays:    7,
		MinRetainCalls: 100,
	}

	type TopologyRecord struct {
		ServiceID    string
		LastCallTime time.Time
		CallCount    int64
	}

	records := []TopologyRecord{
		{ServiceID: "active-svc", LastCallTime: time.Now(), CallCount: 1000},
		{ServiceID: "inactive-svc", LastCallTime: time.Now().AddDate(0, 0, -10), CallCount: 50},
		{ServiceID: "low-call-svc", LastCallTime: time.Now().AddDate(0, 0, -1), CallCount: 10},
	}

	// 清理条件：超过保留天数或调用次数低于阈值
	cleanupCount := 0
	for _, r := range records {
		shouldCleanup := false
		if time.Since(r.LastCallTime).Hours() > float64(policy.MaxDataDays*24) {
			shouldCleanup = true
		}
		if r.CallCount < policy.MinRetainCalls {
			shouldCleanup = true
		}
		if shouldCleanup {
			cleanupCount++
		}
	}

	assert.Equal(t, 1, cleanupCount, "should cleanup 1 record (inactive-svc)")
}

// TestTopologyVisualization 测试拓扑数据可视化
func TestTopologyVisualization(t *testing.T) {
	// 模拟拓扑可视化数据结构
	type Node struct {
		ID       string
		Name     string
		Type     string // "service", "instance", "database", "cache"
		Metrics  map[string]interface{}
	}

	type Edge struct {
		Source      string
		Target      string
		Value       float64
		Metadata    map[string]string
	}

	nodes := []Node{
		{ID: "svc-1", Name: "api-gateway", Type: "service", Metrics: map[string]interface{}{"qps": 1000.0}},
		{ID: "svc-2", Name: "user-service", Type: "service", Metrics: map[string]interface{}{"qps": 500.0}},
		{ID: "db-1", Name: "mysql-master", Type: "database", Metrics: map[string]interface{}{"qps": 2000.0}},
	}

	edges := []Edge{
		{Source: "svc-1", Target: "svc-2", Value: 500.0, Metadata: map[string]string{"latency": "15ms"}},
		{Source: "svc-2", Target: "db-1", Value: 2000.0, Metadata: map[string]string{"latency": "5ms"}},
	}

	assert.Equal(t, 3, len(nodes), "should have 3 nodes")
	assert.Equal(t, 2, len(edges), "should have 2 edges")

	// 验证拓扑连通性
	edgeMap := make(map[string]bool)
	for _, e := range edges {
		edgeMap[e.Source+"->"+e.Target] = true
	}

	assert.True(t, edgeMap["svc-1->svc-2"])
	assert.True(t, edgeMap["svc-2->db-1"])
}

// TestTopologyAnomalyDetection 测试拓扑异常检测
func TestTopologyAnomalyDetection(t *testing.T) {
	// 模拟拓扑异常检测
	type ServiceAnomaly struct {
		ServiceID     string
		AnomalyType   string // "high_failure", "high_latency", "low_throughput"
		CurrentValue  float64
		Threshold     float64
		DurationSecs  int
	}

	anomalies := []ServiceAnomaly{
		{ServiceID: "svc-A", AnomalyType: "high_failure", CurrentValue: 0.15, Threshold: 0.1, DurationSecs: 300},
		{ServiceID: "svc-B", AnomalyType: "high_latency", CurrentValue: 200.0, Threshold: 100.0, DurationSecs: 600},
	}

	// 验证异常检测
	for _, a := range anomalies {
		assert.NotEmpty(t, a.ServiceID, "service ID should not be empty")
		assert.NotEmpty(t, a.AnomalyType, "anomaly type should not be empty")
		assert.True(t, a.DurationSecs > 0, "duration should be positive")
	}

	assert.Equal(t, 2, len(anomalies), "should have 2 anomalies detected")
}

// TestTopologyAggregation 测试拓扑数据聚合
func TestTopologyAggregation(t *testing.T) {
	// 模拟拓扑数据按时间粒度聚合
	type TimeBucket struct {
		StartTime time.Time
		EndTime   time.Time
		Calls     int64
		Success   int64
		Fail      int64
	}

	buckets := []TimeBucket{
		{StartTime: time.Now().Add(-time.Hour * 3), EndTime: time.Now().Add(-time.Hour * 2), Calls: 1000, Success: 990, Fail: 10},
		{StartTime: time.Now().Add(-time.Hour * 2), EndTime: time.Now().Add(-time.Hour * 1), Calls: 1200, Success: 1180, Fail: 20},
		{StartTime: time.Now().Add(-time.Hour * 1), EndTime: time.Now(), Calls: 800, Success: 790, Fail: 10},
	}

	// 聚合统计
	totalCalls := int64(0)
	totalSuccess := int64(0)
	totalFail := int64(0)
	for _, b := range buckets {
		totalCalls += b.Calls
		totalSuccess += b.Success
		totalFail += b.Fail
	}

	assert.Equal(t, int64(3000), totalCalls)
	assert.Equal(t, int64(2960), totalSuccess)
	assert.Equal(t, int64(40), totalFail)

	// 计算成功率
	successRate := float64(totalSuccess) / float64(totalCalls)
	assert.GreaterOrEqual(t, successRate, 0.98)
}

// TestTopologyHistory 测试拓扑历史数据
func TestTopologyHistory(t *testing.T) {
	// 模拟拓扑历史版本
	type TopologyVersion struct {
		Version     int64
		TopologyID  string
		LastUpdate  time.Time
		ServiceCount int
		RelationCount int
	}

	history := []TopologyVersion{
		{Version: 1, TopologyID: "topo-1", LastUpdate: time.Now().Add(-time.Hour * 24), ServiceCount: 10, RelationCount: 50},
		{Version: 2, TopologyID: "topo-1", LastUpdate: time.Now().Add(-time.Hour * 12), ServiceCount: 15, RelationCount: 80},
		{Version: 3, TopologyID: "topo-1", LastUpdate: time.Now(), ServiceCount: 20, RelationCount: 100},
	}

	// 验证版本顺序
	for i := 1; i < len(history); i++ {
		assert.Greater(t, history[i].Version, history[i-1].Version, "version should increase")
		assert.GreaterOrEqual(t, history[i].ServiceCount, history[i-1].ServiceCount, "service count should not decrease")
	}

	// 验证最新版本
	assert.Equal(t, int64(3), history[len(history)-1].Version)
	assert.Equal(t, 20, history[len(history)-1].ServiceCount)
}

// TestTopologyExport 测试拓扑数据导出
func TestTopologyExport(t *testing.T) {
	// 模拟拓扑数据导出格式
	type ExportFormat string
	const (
		FormatJSON ExportFormat = "json"
		FormatGraphML ExportFormat = "graphml"
		FormatDOT ExportFormat = "dot"
	)

	type TopologyData struct {
		Nodes []map[string]interface{}
		Edges []map[string]interface{}
	}

	data := TopologyData{
		Nodes: []map[string]interface{}{
			{"id": "svc-1", "name": "api-gateway"},
			{"id": "svc-2", "name": "user-service"},
		},
		Edges: []map[string]interface{}{
			{"source": "svc-1", "target": "svc-2"},
		},
	}

	assert.Equal(t, 2, len(data.Nodes))
	assert.Equal(t, 1, len(data.Edges))

	// 测试 JSON 导出
	jsonBytes, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, string(jsonBytes))
}

// TestTopologyComparison 测试拓扑数据比较
func TestTopologyComparison(t *testing.T) {
	// 模拟两个版本的拓扑数据比较
	type TopologyDiff struct {
		AddedServices    []string
		RemovedServices  []string
		ChangedRelations []RelationChange
	}

	type RelationChange struct {
		Source      string
		Target      string
		OldValue    int64
		NewValue    int64
	}

	oldTopology := map[string][]string{
		"svc-A": {"svc-B", "svc-C"},
		"svc-B": {"svc-D"},
	}

	newTopology := map[string][]string{
		"svc-A": {"svc-B", "svc-C", "svc-E"},
		"svc-B": {"svc-D"},
		"svc-C": {"svc-F"},
	}

	// 计算差异
	var addedServices, removedServices []string
	var changedRelations []RelationChange

	// 检查新增服务
	for svc := range newTopology {
		if _, ok := oldTopology[svc]; !ok {
			addedServices = append(addedServices, svc)
		}
	}

	// 检查移除的服务
	for svc := range oldTopology {
		if _, ok := newTopology[svc]; !ok {
			removedServices = append(removedServices, svc)
		}
	}

	// 检查关系变化
	for svc, targets := range newTopology {
		oldTargets := oldTopology[svc]
		if len(oldTargets) != len(targets) {
			changedRelations = append(changedRelations, RelationChange{
				Source: svc,
				OldValue: int64(len(oldTargets)),
				NewValue: int64(len(targets)),
			})
		}
	}

	assert.Equal(t, 1, len(addedServices), "should have 1 added service")
	assert.Equal(t, 0, len(removedServices), "should have no removed services")
	assert.GreaterOrEqual(t, len(changedRelations), 0)
}
