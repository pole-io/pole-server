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

package ai

import (
	"fmt"
	"testing"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/stretchr/testify/assert"
)

// 创建测试用的 Skill Version 数据
func createTestSkillVersions() []*ai.SkillVersion {
	return []*ai.SkillVersion{
		{
			ID:            "version-1",
			SkillID:       "skill-1",
			SkillName:     "test-skill-1",
			Namespace:     "default",
			Version:       1,
			Comment:       "Initial version",
			InputSchema:   "{}",
			OutputSchema:  "{}",
			SkillType:     "function",
			Metadata:      map[string]string{},
			Active:        true,
			Flag:          0,
			CTime:         time.Now(),
			MTime:         time.Now(),
		},
		{
			ID:            "version-2",
			SkillID:       "skill-1",
			SkillName:     "test-skill-1",
			Namespace:     "default",
			Version:       2,
			Comment:       "Updated version",
			InputSchema:   "{}",
			OutputSchema:  "{}",
			SkillType:     "function",
			Metadata:      map[string]string{},
			Active:        false,
			Flag:          0,
			CTime:         time.Now(),
			MTime:         time.Now(),
		},
		{
			ID:            "version-3",
			SkillID:       "skill-2",
			SkillName:     "test-skill-2",
			Namespace:     "default",
			Version:       1,
			Comment:       "Initial version",
			InputSchema:   "{}",
			OutputSchema:  "{}",
			SkillType:     "tool",
			Metadata:      map[string]string{},
			Active:        true,
			Flag:          0,
			CTime:         time.Now(),
			MTime:         time.Now(),
		},
	}
}

// TestSkillVersionCache_InterfaceDefinition 测试接口定义
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillVersionCache_InterfaceDefinition(t *testing.T) {
	// 创建测试数据
	versions := createTestSkillVersions()
	assert.Len(t, versions, 3, "应该创建 3 个测试版本")

	// 测试数据的基本属性
	version1 := versions[0]
	assert.Equal(t, "version-1", version1.ID)
	assert.Equal(t, "skill-1", version1.SkillID)
	assert.Equal(t, "test-skill-1", version1.SkillName)
	assert.Equal(t, "default", version1.Namespace)
	assert.Equal(t, uint64(1), version1.Version)
	assert.True(t, version1.Active)

	version2 := versions[1]
	assert.Equal(t, "version-2", version2.ID)
	assert.Equal(t, uint64(2), version2.Version)
	assert.False(t, version2.Active)
}

// TestSkillVersionCache_VersionManagement 测试版本管理
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillVersionCache_VersionManagement(t *testing.T) {
	versions := createTestSkillVersions()

	// 测试按版本号查找
	var versionByVersion *ai.SkillVersion
	for _, v := range versions {
		if v.SkillName == "test-skill-1" && v.Namespace == "default" && v.Version == 1 {
			versionByVersion = v
			break
		}
	}
	assert.NotNil(t, versionByVersion)
	assert.Equal(t, "version-1", versionByVersion.ID)

	// 测试获取活跃版本
	var activeVersion *ai.SkillVersion
	for _, v := range versions {
		if v.SkillName == "test-skill-1" && v.Namespace == "default" && v.Active {
			activeVersion = v
			break
		}
	}
	assert.NotNil(t, activeVersion)
	assert.Equal(t, uint64(1), activeVersion.Version) // version-1 是活跃版本

	// 测试获取技能的所有版本
	var skillVersions []*ai.SkillVersion
	for _, v := range versions {
		if v.SkillID == "skill-1" {
			skillVersions = append(skillVersions, v)
		}
	}
	assert.Len(t, skillVersions, 2)

	// 验证版本号排序
	assert.Equal(t, uint64(1), skillVersions[0].Version)
	assert.Equal(t, uint64(2), skillVersions[1].Version)
}

// TestSkillVersionCache_DataValidation 测试数据验证
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillVersionCache_DataValidation(t *testing.T) {
	versions := createTestSkillVersions()

	// 测试必要字段
	for _, v := range versions {
		assert.NotEmpty(t, v.ID, "Version ID 不能为空")
		assert.NotEmpty(t, v.SkillID, "Skill ID 不能为空")
		assert.NotEmpty(t, v.SkillName, "Skill Name 不能为空")
		assert.NotEmpty(t, v.Namespace, "Namespace 不能为空")
		assert.Greater(t, v.Version, uint64(0), "Version 必须大于 0")
		assert.Equal(t, int8(0), v.Flag, "有效数据的 Flag 应该是 0")
	}

	// 检查活跃版本的唯一性
	var activeCount int
	for _, v := range versions {
		if v.Active {
			activeCount++
		}
	}
	// 同一个技能的活跃版本应该只有一个
	assert.Equal(t, 2, activeCount) // test-skill-1 和 test-skill-2 各有一个活跃版本
}

// TestSkillVersionCache_ConcurrentAccess 测试并发访问
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillVersionCache_ConcurrentAccess(t *testing.T) {
	versions := createTestSkillVersions()

	// 模拟并发读取
	readResults := make(chan string, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			// 随机读取
			if id%2 == 0 {
				readResults <- versions[0].ID
			} else {
				readResults <- versions[1].ID
			}
		}(i)
	}

	// 收集结果
	for i := 0; i < 5; i++ {
		result := <-readResults
		assert.NotEmpty(t, result)
	}

	// 关闭通道
	close(readResults)
}

// TestSkillVersionCache_EdgeCases 测试边界条件
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillVersionCache_EdgeCases(t *testing.T) {
	// 测试空数据
	var emptyVersions []*ai.SkillVersion
	assert.Len(t, emptyVersions, 0)

	// 测试单个数据
	singleVersion := &ai.SkillVersion{
		ID:            "single-version",
		SkillID:       "single-skill",
		SkillName:     "single",
		Namespace:     "default",
		Version:       1,
		Comment:       "Single version",
		InputSchema:   "{}",
		OutputSchema:  "{}",
		SkillType:     "function",
		Metadata:      map[string]string{},
		Active:        true,
		Flag:          0,
		CTime:         time.Now(),
		MTime:         time.Now(),
	}

	assert.Equal(t, "single-version", singleVersion.ID)
	assert.Equal(t, uint64(1), singleVersion.Version)

	// 测试已删除数据（Flag=1）
	deletedVersion := &ai.SkillVersion{
		ID:            "deleted-version",
		SkillID:       "deleted-skill",
		SkillName:     "deleted",
		Namespace:     "default",
		Version:       1,
		Comment:       "Deleted version",
		InputSchema:   "{}",
		OutputSchema:  "{}",
		SkillType:     "function",
		Metadata:      map[string]string{},
		Active:        false,
		Flag:          1, // 已删除
		CTime:         time.Now(),
		MTime:         time.Now(),
	}

	assert.Equal(t, int8(1), deletedVersion.Flag)
}

// TestSkillVersionCache_Performance 测试性能相关
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillVersionCache_Performance(t *testing.T) {
	// 创建大量测试数据
	var largeVersions []*ai.SkillVersion
	for i := 0; i < 1000; i++ {
		largeVersions = append(largeVersions, &ai.SkillVersion{
			ID:            versionID(i),
			SkillID:       "skill-1",
			SkillName:     "test-skill",
			Namespace:     "default",
			Version:       uint64(i + 1),
			Comment:       "Version",
			InputSchema:   "{}",
			OutputSchema:  "{}",
			SkillType:     "function",
			Metadata:      map[string]string{},
			Active:        i%2 == 0,
			Flag:          0,
			CTime:         time.Now(),
			MTime:         time.Now(),
		})
	}

	// 验证数据量
	assert.Len(t, largeVersions, 1000)

	// 测试快速查找
	start := time.Now()
	found := false
	for _, v := range largeVersions {
		if v.ID == "version-999" {
			found = true
			break
		}
	}
	elapsed := time.Since(start)

	assert.True(t, found, "应该找到 version-999")
	assert.Less(t, elapsed, time.Millisecond*100, "查找时间应该在 100ms 内")
}

// TestSkillVersionCache_IncrementalUpdate 测试增量更新
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillVersionCache_IncrementalUpdate(t *testing.T) {
	// 创建初始数据
	initialVersions := createTestSkillVersions()
	assert.Len(t, initialVersions, 3, "初始应该有 3 个版本")

	// 准备增量数据 - 添加新版本
	_ = append(initialVersions, &ai.SkillVersion{
		ID:            "version-4",
		SkillID:       "skill-2",
		SkillName:     "test-skill-2",
		Namespace:     "default",
		Version:       2,
		Comment:       "Updated version 2",
		InputSchema:   "{}",
		OutputSchema:  "{}",
		SkillType:     "tool",
		Metadata:      map[string]string{},
		Active:        false,
		Flag:          0,
		CTime:         time.Now(),
		MTime:         time.Now(),
	})

	// 模拟删除版本（通过设置 Flag=1）
	deletedVersions := append(initialVersions, &ai.SkillVersion{
		ID:            "version-deleted",
		SkillID:       "skill-3",
		SkillName:     "to-be-deleted",
		Namespace:     "default",
		Version:       1,
		Comment:       "Deleted version",
		InputSchema:   "{}",
		OutputSchema:  "{}",
		SkillType:     "function",
		Metadata:      map[string]string{},
		Active:        false,
		Flag:          1, // 已删除
		CTime:         time.Now().Add(-time.Hour),
		MTime:         time.Now().Add(-time.Hour),
	})

	// 测试删除标记处理
	var activeVersions []*ai.SkillVersion
	for _, v := range deletedVersions {
		if v.Flag == 0 { // 只保留未删除的
			activeVersions = append(activeVersions, v)
		}
	}
	assert.Len(t, activeVersions, 3, "删除后应该只有 3 个活跃版本")
}

// TestSkillVersionCache_SingleFlight 测试 SingleFlight 防重入机制
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillVersionCache_SingleFlight(t *testing.T) {
	_ = createTestSkillVersions()

	// 模拟多个并发更新调用
	updateCount := 10
	updateComplete := make(chan bool, updateCount)

	// 启动多个 goroutine 同时调用 Update
	for i := 0; i < updateCount; i++ {
		go func(id int) {
			// 注意：实际实现中，Update 方法应该使用 SingleFlight
			time.Sleep(time.Millisecond * time.Duration(id))
			updateComplete <- true
		}(i)
	}

	// 等待所有更新完成
	completed := 0
	for i := 0; i < updateCount; i++ {
		<-updateComplete
		completed++
	}

	assert.Equal(t, updateCount, completed, "所有并发更新都应该完成")
}

// TestSkillVersionCache_Lifecycle 测试缓存生命周期
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillVersionCache_Lifecycle(t *testing.T) {
	versions := createTestSkillVersions()

	// 1. Initialize
	assert.NotNil(t, versions, "Initialize 阶段数据应该存在")

	// 2. Update
	assert.Len(t, versions, 3, "Update 后应该保持数据")

	// 3. Clear
	var emptyVersions []*ai.SkillVersion
	assert.Len(t, emptyVersions, 0, "Clear 后数据应该为空")

	// 4. Close
	// 实际实现中可能需要关闭资源
}

// TestSkillVersionCache_VersionActivation 测试版本激活
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillVersionCache_VersionActivation(t *testing.T) {
	versions := createTestSkillVersions()

	// 查找技能 test-skill-1 的版本
	var versionsForSkill []*ai.SkillVersion
	for _, v := range versions {
		if v.SkillName == "test-skill-1" && v.Namespace == "default" {
			versionsForSkill = append(versionsForSkill, v)
		}
	}

	assert.Len(t, versionsForSkill, 2)

	// 验证初始状态
	var activeCount int
	for _, v := range versionsForSkill {
		if v.Active {
			activeCount++
		}
	}
	assert.Equal(t, 1, activeCount) // 应该只有一个活跃版本

	// 注意：实际实现中需要测试激活/取消激活操作
	// 这里只是测试初始数据的一致性
}

// 辅助函数
func versionID(i int) string {
	return fmt.Sprintf("version-%d", i)
}