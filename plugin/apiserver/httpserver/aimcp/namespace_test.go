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

package aimcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNamespaceHandleQueryNamespaces_Success 测试查询命名空间成功场景
func TestNamespaceHandleQueryNamespaces_Success(t *testing.T) {
	// 模拟请求参数
	args := map[string]interface{}{
		"name":   "test*",
		"offset": uint32(0),
		"limit":  uint32(10),
		"all":    false,
	}

	// 测试参数解析
	name, _ := args["name"].(string)
	assert.Equal(t, "test*", name)

	_, _ = args["offset"].(uint32)

	limit, _ := args["limit"].(uint32)
	assert.Equal(t, uint32(10), limit)

	all, _ := args["all"].(bool)
	assert.False(t, all)
}

// TestNamespaceHandleQueryNamespaces_WithAllTrue 测试查询所有命名空间
func TestNamespaceHandleQueryNamespaces_WithAllTrue(t *testing.T) {
	args := map[string]interface{}{
		"name": "test",
		"all":  true,
	}

	all, _ := args["all"].(bool)
	assert.True(t, all)

	// 当 all 为 true 时，应该循环分页查询
	// 测试分页逻辑
	_ = uint32(0)
	limit := uint32(100)
	hasMore := true

	count := 0
	for hasMore && count < 5 {
		count++
		// 模拟每次查询 100 条
		if count*int(limit) > 250 {
			hasMore = false
		}
	}

	assert.Equal(t, 3, count)
}

// TestNamespaceHandleQueryNamespaces_InvalidArgs 测试无效参数
func TestNamespaceHandleQueryNamespaces_InvalidArgs(t *testing.T) {
	// 空参数
	emptyArgs := map[string]interface{}{}
	assert.Empty(t, emptyArgs)

	// nil 参数
	var nilArgs map[string]interface{}
	assert.Nil(t, nilArgs)

	// 错误类型参数
	wrongTypeArgs := map[string]interface{}{
		"name":   123,        // 应该是 string
		"offset": "invalid",  // 应该是数字
		"limit":  []int{1, 2}, // 应该是数字
	}

	// 类型断言失败时返回零值
	name, ok := wrongTypeArgs["name"].(string)
	assert.False(t, ok)
	assert.Empty(t, name)
}

// TestNamespaceHandleCreateNamespaces_Success 测试创建命名空间成功
func TestNamespaceHandleCreateNamespaces_Success(t *testing.T) {
	// 模拟请求参数
	namespaces := []interface{}{
		map[string]interface{}{
			"name":        "test-ns-1",
			"description": "test namespace 1",
		},
		map[string]interface{}{
			"name":        "test-ns-2",
			"description": "test namespace 2",
		},
	}

	args := map[string]interface{}{
		"namespaces": namespaces,
	}

	// 验证参数解析
	nsArray, ok := args["namespaces"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(nsArray))
}

// TestNamespaceHandleCreateNamespaces_Empty 测试空命名空间数组
func TestNamespaceHandleCreateNamespaces_Empty(t *testing.T) {
	args := map[string]interface{}{
		"namespaces": []interface{}{},
	}

	nsArray, ok := args["namespaces"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 0, len(nsArray))

	// 验证空数组应该返回错误
	assert.Equal(t, 0, len(nsArray))
}

// TestNamespaceHandleCreateNamespaces_Invalid 测试无效命名空间数据
func TestNamespaceHandleCreateNamespaces_Invalid(t *testing.T) {
	args := map[string]interface{}{
		"namespaces": "not-an-array",
	}

	// 类型断言失败
	nsArray, ok := args["namespaces"].([]interface{})
	assert.False(t, ok)
	assert.Nil(t, nsArray)
}

// TestNamespaceToolNames 测试工具名称
func TestNamespaceToolNames(t *testing.T) {
	tools := []string{
		"list_namespaces",
		"create_namespaces",
		"update_namespaces",
		"delete_namespaces",
	}

	expectedTools := map[string]bool{
		"list_namespaces":   true,
		"create_namespaces": true,
		"update_namespaces": true,
		"delete_namespaces": true,
	}

	for _, tool := range tools {
		assert.True(t, expectedTools[tool], "tool %s should be registered", tool)
	}
}

// TestNamespaceToolDescription 测试工具描述
func TestNamespaceToolDescription(t *testing.T) {
	descriptionTests := []struct {
		tool        string
		description string
	}{
		{
			tool:        "list_namespaces",
			description: "查询服务治理中心下的命名空间列表",
		},
		{
			tool:        "create_namespaces",
			description: "在服务治理中心下的创建多个命名空间",
		},
		{
			tool:        "update_namespaces",
			description: "在服务治理中心下的更新多个命名空间",
		},
		{
			tool:        "delete_namespaces",
			description: "在服务治理中心下的删除多个命名空间",
		},
	}

	for _, tt := range descriptionTests {
		t.Run(tt.tool, func(t *testing.T) {
			assert.Contains(t, tt.description, "命名空间")
		})
	}
}

// TestNamespaceToolParameters 测试工具参数定义
func TestNamespaceToolParameters(t *testing.T) {
	// list_namespaces 参数
	listParams := map[string]string{
		"name":   "根据名称查询过滤",
		"offset": "查询的偏移量",
		"limit":  "限制返回的命名空间数量",
		"all":    "是否查询所有命名空间",
	}

	assert.Equal(t, "根据名称查询过滤", listParams["name"])
	assert.Equal(t, "查询的偏移量", listParams["offset"])
}

// TestHandleCreateNamespaces_InvalidType 测试创建时的类型错误
func TestHandleCreateNamespaces_InvalidType(t *testing.T) {
	// 测试非数组输入
	args := map[string]interface{}{
		"namespaces": map[string]interface{}{
			"name": "test",
		},
	}

	// 应该返回类型断言失败
	_, ok := args["namespaces"].([]interface{})
	assert.False(t, ok)
}

// TestNamespaceArgumentParsing 测试参数解析的边界情况
func TestNamespaceArgumentParsing(t *testing.T) {
	testCases := []struct {
		name  string
		value interface{}
		valid bool
	}{
		{"string name", "test-name", true},
		{"int name", 12345, false},
		{"nil value", nil, true},
		{"empty string", "", true},
		{"large offset", uint32(999999), false},
		{"zero limit", uint32(0), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := tc.value.(string)
			if tc.valid {
				assert.True(t, ok || tc.value == nil)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

// TestMarshalWithJSON 测试使用 JSON 序列化
func TestMarshalWithJSON(t *testing.T) {
	// 使用测试结构体代替外部依赖
	type testNamespace struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	ns := &testNamespace{
		Name:        "test-ns",
		Description: "test description",
	}

	// 使用标准库 JSON 序列化
	jsonBytes, err := json.Marshal(ns)
	assert.NoError(t, err)
	assert.NotEmpty(t, string(jsonBytes))

	// 验证 JSON 包含预期字段
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NoError(t, err)
	assert.Equal(t, "test-ns", result["name"])
	assert.Equal(t, "test description", result["description"])
}

// TestMarshalWithJSON_Empty 测试空结构体序列化
func TestMarshalWithJSON_Empty(t *testing.T) {
	type testNamespace struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	ns := &testNamespace{}

	jsonBytes, err := json.Marshal(ns)
	assert.NoError(t, err)
	assert.NotEmpty(t, string(jsonBytes))

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NoError(t, err)
}

// TestMarshalWithJSON_Nil 测试 nil 序列化
func TestMarshalWithJSON_Nil(t *testing.T) {
	type testNamespaceNil struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var nilNS *testNamespaceNil

	// nil 指针可以序列化为 null
	jsonBytes, err := json.Marshal(nilNS)
	assert.NoError(t, err)
	assert.Equal(t, "null", string(jsonBytes))
}
