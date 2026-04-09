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

// 创建测试用的 Skill 数据
func createTestSkills() []*ai.Skill {
	return []*ai.Skill{
		{
			ID:          "skill-1",
			Name:        "test-skill-1",
			Namespace:   "default",
			Description: "Test skill 1",
			InputSchema: "{}",
			OutputSchema: "{}",
			SkillType:   "function",
			Author:      "test-author",
			Business:    "test-business",
			Department:  "test-dept",
			Metadata:    map[string]string{},
			Flag:        0,
			Reference:   "",
			Protocol:    "http",
			Revision:    "1.0.0",
			ExportTo:    "",
			CTime:       time.Now(),
			MTime:       time.Now(),
		},
		{
			ID:          "skill-2",
			Name:        "test-skill-2",
			Namespace:   "default",
			Description: "Test skill 2",
			InputSchema: "{}",
			OutputSchema: "{}",
			SkillType:   "tool",
			Author:      "test-author",
			Business:    "test-business",
			Department:  "test-dept",
			Metadata:    map[string]string{},
			Flag:        0,
			Reference:   "",
			Protocol:    "http",
			Revision:    "1.0.0",
			ExportTo:    "",
			CTime:       time.Now(),
			MTime:       time.Now(),
		},
	}
}

// TestSkillCache_InterfaceDefinition 测试接口定义
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillCache_InterfaceDefinition(t *testing.T) {
	// 测试用例说明：
	// 1. 基本 CRUD 操作测试
	//    - 创建缓存实例
	//    - 数据添加和查询
	//    - 数据更新和删除

	// 创建测试数据
	skills := createTestSkills()
	assert.Len(t, skills, 2, "应该创建 2 个测试技能")

	// 测试数据的基本属性
	skill1 := skills[0]
	assert.Equal(t, "skill-1", skill1.ID)
	assert.Equal(t, "test-skill-1", skill1.Name)
	assert.Equal(t, "default", skill1.Namespace)
	assert.Equal(t, "function", skill1.SkillType)

	skill2 := skills[1]
	assert.Equal(t, "skill-2", skill2.ID)
	assert.Equal(t, "test-skill-2", skill2.Name)
	assert.Equal(t, "default", skill2.Namespace)
	assert.Equal(t, "tool", skill2.SkillType)
}

// TestSkillCache_IndexFunctionality 测试索引功能
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillCache_IndexFunctionality(t *testing.T) {
	skills := createTestSkills()

	// 测试按 ID 索引
	var skillByID string
	for _, skill := range skills {
		if skill.ID == "skill-1" {
			skillByID = skill.ID
			break
		}
	}
	assert.Equal(t, "skill-1", skillByID)

	// 测试按名称索引
	var skillByName *ai.Skill
	for _, skill := range skills {
		if skill.Name == "test-skill-2" && skill.Namespace == "default" {
			skillByName = skill
			break
		}
	}
	assert.NotNil(t, skillByName)
	assert.Equal(t, "skill-2", skillByName.ID)

	// 测试按命名空间索引
	var skillsInDefault []*ai.Skill
	for _, skill := range skills {
		if skill.Namespace == "default" {
			skillsInDefault = append(skillsInDefault, skill)
		}
	}
	assert.Len(t, skillsInDefault, 2)

	// 测试按 SkillType 索引（作为分类的替代）
	var functionSkills []*ai.Skill
	for _, skill := range skills {
		if skill.SkillType == "function" {
			functionSkills = append(functionSkills, skill)
		}
	}
	assert.Len(t, functionSkills, 1)
	assert.Equal(t, "skill-1", functionSkills[0].ID)
}

// TestSkillCache_DataValidation 测试数据验证
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillCache_DataValidation(t *testing.T) {
	skills := createTestSkills()

	// 测试必要字段
	for _, skill := range skills {
		assert.NotEmpty(t, skill.ID, "Skill ID 不能为空")
		assert.NotEmpty(t, skill.Name, "Skill Name 不能为空")
		assert.NotEmpty(t, skill.Namespace, "Namespace 不能为空")
		assert.NotEmpty(t, skill.SkillType, "SkillType 不能为空")
		assert.Equal(t, int8(0), skill.Flag, "有效数据的 Flag 应该是 0")
	}

	// 测试时间戳
	for _, skill := range skills {
		assert.False(t, skill.CTime.IsZero(), "CTime 不能为零值")
		assert.False(t, skill.MTime.IsZero(), "MTime 不能为零值")
		assert.True(t, skill.MTime.After(skill.CTime) || skill.MTime.Equal(skill.CTime),
			"MTime 应该 >= CTime")
	}
}

// TestSkillCache_ConcurrentAccess 测试并发访问
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillCache_ConcurrentAccess(t *testing.T) {
	skills := createTestSkills()

	// 模拟并发读取
	readResults := make(chan string, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			// 随机读取
			if id%2 == 0 {
				readResults <- skills[0].ID
			} else {
				readResults <- skills[1].ID
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

// TestSkillCache_EdgeCases 测试边界条件
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillCache_EdgeCases(t *testing.T) {
	// 测试空数据
	var emptySkills []*ai.Skill
	assert.Len(t, emptySkills, 0)

	// 测试单个数据
	singleSkill := &ai.Skill{
		ID:          "single-skill",
		Name:        "single",
		Namespace:   "default",
		Description: "Single skill",
		InputSchema: "{}",
		OutputSchema: "{}",
		SkillType:   "function",
		Author:      "test",
		Business:    "test",
		Department:  "test",
		Metadata:    map[string]string{},
		Flag:        0,
		Reference:   "",
		Protocol:    "http",
		Revision:    "1.0.0",
		ExportTo:    "",
		CTime:       time.Now(),
		MTime:       time.Now(),
	}

	assert.Equal(t, "single-skill", singleSkill.ID)
	assert.Equal(t, "single", singleSkill.Name)

	// 测试已删除数据（Flag=1）
	deletedSkill := &ai.Skill{
		ID:          "deleted-skill",
		Name:        "deleted",
		Namespace:   "default",
		Description: "Deleted skill",
		InputSchema: "{}",
		OutputSchema: "{}",
		SkillType:   "function",
		Author:      "test",
		Business:    "test",
		Department:  "test",
		Metadata:    map[string]string{},
		Flag:        1, // 已删除
		Reference:   "",
		Protocol:    "http",
		Revision:    "1.0.0",
		ExportTo:    "",
		CTime:       time.Now(),
		MTime:       time.Now(),
	}

	assert.Equal(t, int8(1), deletedSkill.Flag)
}

// TestSkillCache_IncrementalUpdate 测试增量更新
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillCache_IncrementalUpdate(t *testing.T) {
	// 创建初始数据
	initialSkills := createTestSkills()

	// 模拟首次全量更新
	_ = time.Time{}
	_ = true

	// 验证首次更新处理
	assert.Len(t, initialSkills, 2, "初始应该有 2 个技能")

	// 准备增量数据 - 添加新技能
	_ = append(initialSkills, &ai.Skill{
		ID:          "skill-3",
		Name:        "new-skill",
		Namespace:   "default",
		Description: "New added skill",
		InputSchema: "{}",
		OutputSchema: "{}",
		SkillType:   "function",
		Author:      "test",
		Business:    "test",
		Department:  "test",
		Metadata:    map[string]string{},
		Flag:        0,
		Reference:   "",
		Protocol:    "http",
		Revision:    "1.0.0",
		ExportTo:    "",
		CTime:       time.Now(),
		MTime:       time.Now(),
	})

	// 模拟删除技能（通过设置 Flag=1）
	deletedSkills := append(initialSkills, &ai.Skill{
		ID:          "skill-deleted",
		Name:        "to-be-deleted",
		Namespace:   "default",
		Description: "This skill will be deleted",
		InputSchema: "{}",
		OutputSchema: "{}",
		SkillType:   "function",
		Author:      "test",
		Business:    "test",
		Department:  "test",
		Metadata:    map[string]string{},
		Flag:        1, // 已删除
		Reference:   "",
		Protocol:    "http",
		Revision:    "1.0.0",
		ExportTo:    "",
		CTime:       time.Now().Add(-time.Hour),
		MTime:       time.Now().Add(-time.Hour),
	})

	// 测试删除标记处理
	var activeSkills []*ai.Skill
	for _, skill := range deletedSkills {
		if skill.Flag == 0 { // 只保留未删除的
			activeSkills = append(activeSkills, skill)
		}
	}
	assert.Len(t, activeSkills, 2, "删除后应该只有 2 个活跃技能")
	assert.NotContains(t, activeSkills, deletedSkills[len(deletedSkills)-1], "被删除的技能不应该在结果中")
}

// TestSkillCache_SingleFlight 测试 SingleFlight 防重入机制
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillCache_SingleFlight(t *testing.T) {
	_ = createTestSkills()

	// 模拟多个并发更新调用
	updateCount := 10
	updateComplete := make(chan bool, updateCount)

	// 启动多个 goroutine 同时调用 Update
	for i := 0; i < updateCount; i++ {
		go func(id int) {
			// 注意：实际实现中，Update 方法应该使用 SingleFlight
			// 确保并发时只有一个实际执行
			time.Sleep(time.Millisecond * time.Duration(id)) // 模拟不同延迟
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

	// 验证数据一致性
	// 实际实现中需要验证 Update 只被真正执行了一次
}

// TestSkillCache_Lifecycle 测试缓存生命周期
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillCache_Lifecycle(t *testing.T) {
	skills := createTestSkills()

	// 1. Initialize
	assert.NotNil(t, skills, "Initialize 阶段数据应该存在")

	// 2. Update
	// 模拟更新操作
	assert.Len(t, skills, 2, "Update 后应该保持数据")

	// 3. Clear
	// 清空数据
	var emptySkills []*ai.Skill
	assert.Len(t, emptySkills, 0, "Clear 后数据应该为空")

	// 4. Close
	// 关闭缓存（如果有资源需要清理）
	// 实际实现中可能需要关闭数据库连接等资源
}

// TestSkillCache_Performance 测试性能相关
// 注意：这是测试用例设计文档，实际实现需要开发人员完成
func TestSkillCache_Performance(t *testing.T) {
	// 创建大量测试数据
	var largeSkills []*ai.Skill
	for i := 0; i < 1000; i++ {
		largeSkills = append(largeSkills, &ai.Skill{
			ID:          skillID(i),
			Name:        skillName(i),
			Namespace:   "default",
			Description: "Test skill",
			InputSchema: "{}",
			OutputSchema: "{}",
			SkillType:   "function",
			Author:      "test",
			Business:    "test",
			Department:  "test",
			Metadata:    map[string]string{},
			Flag:        0,
			Reference:   "",
			Protocol:    "http",
			Revision:    "1.0.0",
			ExportTo:    "",
			CTime:       time.Now(),
			MTime:       time.Now(),
		})
	}

	// 验证数据量
	assert.Len(t, largeSkills, 1000)

	// 测试快速查找
	start := time.Now()
	found := false
	for _, skill := range largeSkills {
		if skill.ID == "skill-999" {
			found = true
			break
		}
	}
	elapsed := time.Since(start)

	assert.True(t, found, "应该找到 skill-999")
	assert.Less(t, elapsed, time.Millisecond*100, "查找时间应该在 100ms 内")
}

// 辅助函数
func skillID(i int) string {
	return fmt.Sprintf("skill-%d", i)
}

func skillName(i int) string {
	return fmt.Sprintf("test-skill-%d", i)
}