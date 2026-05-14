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

package skill

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/pole-io/pole-server/apis/pkg/types/ai"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSkillServer implements skill.SkillServer interface for testing
type MockSkillServer struct {
	mock.Mock
}

func (m *MockSkillServer) CreateSkill(ctx context.Context, skill *ai.Skill) error {
	args := m.Called(ctx, skill)
	return args.Error(0)
}

func (m *MockSkillServer) UpdateSkill(ctx context.Context, skill *ai.Skill) error {
	args := m.Called(ctx, skill)
	return args.Error(0)
}

func (m *MockSkillServer) DeleteSkill(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSkillServer) GetSkill(ctx context.Context, id string) (*ai.Skill, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ai.Skill), args.Error(1)
}

func (m *MockSkillServer) GetSkillByName(ctx context.Context, name, namespace string) (*ai.Skill, error) {
	args := m.Called(ctx, name, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ai.Skill), args.Error(1)
}

func (m *MockSkillServer) CreateSkillGroup(ctx context.Context, group *ai.SkillGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockSkillServer) UpdateSkillGroup(ctx context.Context, group *ai.SkillGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockSkillServer) DeleteSkillGroup(ctx context.Context, namespace, name string) error {
	args := m.Called(ctx, namespace, name)
	return args.Error(0)
}

func (m *MockSkillServer) GetSkillGroup(ctx context.Context, namespace, name string) (*ai.SkillGroup, error) {
	args := m.Called(ctx, namespace, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ai.SkillGroup), args.Error(1)
}

func (m *MockSkillServer) CreateSkillVersion(ctx context.Context, version *ai.SkillVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockSkillServer) ActivateSkillVersion(ctx context.Context, versionID string) error {
	args := m.Called(ctx, versionID)
	return args.Error(0)
}

func (m *MockSkillServer) CreateSkillSubscription(ctx context.Context, sub *ai.SkillSubscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSkillServer) DeleteSkillSubscription(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSkillServer) GetSkillSubscriptionsBySkill(ctx context.Context, skillName, namespace string) ([]*ai.SkillSubscription, error) {
	args := m.Called(ctx, skillName, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ai.SkillSubscription), args.Error(1)
}

func (m *MockSkillServer) GetSkillSubscriptionsByClient(ctx context.Context, clientID string) ([]*ai.SkillSubscription, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ai.SkillSubscription), args.Error(1)
}

// Batch operations
func (m *MockSkillServer) CreateSkills(ctx context.Context, skills []*ai.Skill) *apimodel.BatchWriteResponse {
	args := m.Called(ctx, skills)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchWriteResponse)
}

func (m *MockSkillServer) UpdateSkills(ctx context.Context, skills []*ai.Skill) *apimodel.BatchWriteResponse {
	args := m.Called(ctx, skills)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchWriteResponse)
}

func (m *MockSkillServer) DeleteSkills(ctx context.Context, skills []*ai.Skill) *apimodel.BatchWriteResponse {
	args := m.Called(ctx, skills)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchWriteResponse)
}

func (m *MockSkillServer) GetSkills(ctx context.Context, filters map[string]string) *apimodel.BatchQueryResponse {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchQueryResponse)
}

func (m *MockSkillServer) GetAllSkills(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchQueryResponse)
}

func (m *MockSkillServer) GetSkillsCount(ctx context.Context) *apimodel.BatchQueryResponse {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchQueryResponse)
}

func (m *MockSkillServer) CreateSkillGroups(ctx context.Context, groups []*ai.SkillGroup) *apimodel.BatchWriteResponse {
	args := m.Called(ctx, groups)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchWriteResponse)
}

func (m *MockSkillServer) UpdateSkillGroups(ctx context.Context, groups []*ai.SkillGroup) *apimodel.BatchWriteResponse {
	args := m.Called(ctx, groups)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchWriteResponse)
}

func (m *MockSkillServer) DeleteSkillGroups(ctx context.Context, groups []*ai.SkillGroup) *apimodel.BatchWriteResponse {
	args := m.Called(ctx, groups)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchWriteResponse)
}

func (m *MockSkillServer) GetSkillGroups(ctx context.Context, filters map[string]string) *apimodel.BatchQueryResponse {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchQueryResponse)
}

func (m *MockSkillServer) GetAllSkillGroups(ctx context.Context) *apimodel.BatchQueryResponse {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchQueryResponse)
}

func (m *MockSkillServer) GetSkillGroupsCount(ctx context.Context) *apimodel.BatchQueryResponse {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchQueryResponse)
}

// SkillVersion batch operations
func (m *MockSkillServer) CreateSkillVersions(ctx context.Context, versions []*ai.SkillVersion) *apimodel.BatchWriteResponse {
	args := m.Called(ctx, versions)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchWriteResponse)
}

func (m *MockSkillServer) DeleteSkillVersions(ctx context.Context, ids []string) *apimodel.BatchWriteResponse {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchWriteResponse)
}

func (m *MockSkillServer) GetSkillVersions(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchQueryResponse)
}

// SkillSubscription batch operations
func (m *MockSkillServer) CreateSkillSubscriptions(ctx context.Context, subs []*ai.SkillSubscription) *apimodel.BatchWriteResponse {
	args := m.Called(ctx, subs)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchWriteResponse)
}

func (m *MockSkillServer) DeleteSkillSubscriptions(ctx context.Context, ids []string) *apimodel.BatchWriteResponse {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchWriteResponse)
}

func (m *MockSkillServer) GetSkillSubscriptions(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*apimodel.BatchQueryResponse)
}

// MockResponse is a mock response type
type MockResponse struct {
	Code uint32 `json:"code"`
	Info string `json:"info"`
	Size uint32 `json:"size"`
}

func (m *MockResponse) Reset()         { *m = MockResponse{} }
func (m *MockResponse) String() string { return "" }
func (m *MockResponse) ProtoMessage()  {}

// MockQueryResponse is a mock query response type
type MockQueryResponse struct {
	Code   uint32        `json:"code"`
	Info   string        `json:"info"`
	Amount uint32        `json:"amount"`
	Size   uint32        `json:"size"`
	Data   []interface{} `json:"data"`
}

func (m *MockQueryResponse) Reset()         { *m = MockQueryResponse{} }
func (m *MockQueryResponse) String() string { return "" }
func (m *MockQueryResponse) ProtoMessage()  {}

// TestCreateSkillGroups_Success 测试创建 SkillGroup 成功
func TestCreateSkillGroups_Success(t *testing.T) {
	mockServer := &MockSkillServer{}
	handler := NewHTTPServer(mockServer)

	group := &ai.SkillGroup{
		ID:        "group-1",
		Name:      "test-group",
		Namespace: "default",
	}

	reqBody := map[string][]*ai.SkillGroup{
		"groups": {group},
	}
	jsonBody, _ := json.Marshal(reqBody)

	// 模拟服务层成功响应
	mockServer.On("CreateSkillGroups", mock.Anything, []*ai.SkillGroup{group}).Return(
		&MockResponse{
			Code: 200000,
			Info: "execute success",
			Size: 1,
		},
	)

	req := httptest.NewRequest("POST", "/skill/groups", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handler.CreateSkillGroups(restful.NewRequest(req), &restful.Response{ResponseWriter: resp})

	assert.Equal(t, http.StatusOK, resp.Code)

	var response MockResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint32(200000), response.Code)
	assert.Equal(t, "execute success", response.Info)
	assert.Equal(t, uint32(1), response.Size)

	mockServer.AssertExpectations(t)
}

// TestCreateSkillGroups_EmptyName 测试创建 SkillGroup - 名称为空
func TestCreateSkillGroups_EmptyName(t *testing.T) {
	handler := NewHTTPServer(&MockSkillServer{})

	group := &ai.SkillGroup{
		ID:        "group-1",
		Name:      "", // 空名称
		Namespace: "default",
	}

	reqBody := map[string][]*ai.SkillGroup{
		"groups": {group},
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/skill/groups", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handler.CreateSkillGroups(restful.NewRequest(req), &restful.Response{ResponseWriter: resp})

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var response MockResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint32(400000), response.Code)
	assert.Contains(t, response.Info, "name")
}

// TestCreateSkillGroups_EmptyNamespace 测试创建 SkillGroup - 命名空间为空
func TestCreateSkillGroups_EmptyNamespace(t *testing.T) {
	handler := NewHTTPServer(&MockSkillServer{})

	group := &ai.SkillGroup{
		ID:        "group-1",
		Name:      "test-group",
		Namespace: "", // 空命名空间
	}

	reqBody := map[string][]*ai.SkillGroup{
		"groups": {group},
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/skill/groups", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handler.CreateSkillGroups(restful.NewRequest(req), &restful.Response{ResponseWriter: resp})

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var response MockResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint32(400000), response.Code)
	assert.Contains(t, response.Info, "namespace")
}

// TestGetSkillGroups_Success 测试查询 SkillGroup 列表成功
func TestGetSkillGroups_Success(t *testing.T) {
	mockServer := &MockSkillServer{}
	handler := NewHTTPServer(mockServer)

	query := map[string]string{
		"namespace": "default",
		"name":      "test*",
		"offset":    "0",
		"limit":     "10",
	}

	// 模拟服务层成功响应
	mockServer.On("GetSkillGroups", mock.Anything, query).Return(
		&MockQueryResponse{
			Code:   200000,
			Info:   "execute success",
			Amount: 2,
			Size:   1,
			Data:   nil,
		},
	)

	req := httptest.NewRequest("GET", "/skill/groups?namespace=default&name=test*&offset=0&limit=10", nil)
	resp := httptest.NewRecorder()

	handler.GetSkillGroups(restful.NewRequest(req), &restful.Response{ResponseWriter: resp})

	assert.Equal(t, http.StatusOK, resp.Code)

	var response MockQueryResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint32(200000), response.Code)
	assert.Equal(t, "execute success", response.Info)
	assert.Equal(t, uint32(2), response.Amount)
	assert.Equal(t, uint32(1), response.Size)

	mockServer.AssertExpectations(t)
}

// TestGetSkillGroups_WithPagination 测试分页查询 SkillGroup
func TestGetSkillGroups_WithPagination(t *testing.T) {
	mockServer := &MockSkillServer{}
	handler := NewHTTPServer(mockServer)

	query := map[string]string{
		"namespace": "default",
		"offset":    "50",
		"limit":     "20",
	}

	// 模拟服务层成功响应
	mockServer.On("GetSkillGroups", mock.Anything, query).Return(
		&MockQueryResponse{
			Code:   200000,
			Info:   "execute success",
			Amount: 100,
			Size:   0,
			Data:   nil,
		},
	)

	req := httptest.NewRequest("GET", "/skill/groups?namespace=default&offset=50&limit=20", nil)
	resp := httptest.NewRecorder()

	handler.GetSkillGroups(restful.NewRequest(req), &restful.Response{ResponseWriter: resp})

	assert.Equal(t, http.StatusOK, resp.Code)

	var response MockQueryResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint32(200000), response.Code)
	assert.Equal(t, "execute success", response.Info)
	assert.Equal(t, uint32(100), response.Amount)
	assert.Equal(t, uint32(0), response.Size)

	mockServer.AssertExpectations(t)
}

// TestUpdateSkillGroups_Success 测试更新 SkillGroup 成功
func TestUpdateSkillGroups_Success(t *testing.T) {
	mockServer := &MockSkillServer{}
	handler := NewHTTPServer(mockServer)

	updatedGroup := &ai.SkillGroup{
		ID:        "group-1",
		Name:      "updated-group",
		Namespace: "default",
	}

	reqBody := map[string][]*ai.SkillGroup{
		"groups": {updatedGroup},
	}
	jsonBody, _ := json.Marshal(reqBody)

	// 模拟服务层成功响应
	mockServer.On("UpdateSkillGroups", mock.Anything, []*ai.SkillGroup{updatedGroup}).Return(
		&MockResponse{
			Code: 200000,
			Info: "execute success",
			Size: 1,
		},
	)

	req := httptest.NewRequest("PUT", "/skill/groups", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handler.UpdateSkillGroups(restful.NewRequest(req), &restful.Response{ResponseWriter: resp})

	assert.Equal(t, http.StatusOK, resp.Code)

	var response MockResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint32(200000), response.Code)
	assert.Equal(t, "execute success", response.Info)
	assert.Equal(t, uint32(1), response.Size)

	mockServer.AssertExpectations(t)
}

// TestUpdateSkillGroups_NotFound 测试更新不存在的 SkillGroup
func TestUpdateSkillGroups_NotFound(t *testing.T) {
	mockServer := &MockSkillServer{}
	handler := NewHTTPServer(mockServer)

	group := &ai.SkillGroup{
		ID:        "non-existent-group",
		Name:      "non-existent",
		Namespace: "default",
	}

	// 模拟服务层返回不存在的错误
	mockServer.On("UpdateSkillGroups", mock.Anything, []*ai.SkillGroup{group}).Return(
		&MockResponse{
			Code: 400101,
			Info: "skill group not found",
		},
	)

	reqBody := map[string][]*ai.SkillGroup{
		"groups": {group},
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/skill/groups", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handler.UpdateSkillGroups(restful.NewRequest(req), &restful.Response{ResponseWriter: resp})

	assert.Equal(t, http.StatusOK, resp.Code)

	var response MockResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint32(400101), response.Code)
	assert.Equal(t, "skill group not found", response.Info)

	mockServer.AssertExpectations(t)
}

// TestDeleteSkillGroups_Success 测试删除 SkillGroup 成功
func TestDeleteSkillGroups_Success(t *testing.T) {
	mockServer := &MockSkillServer{}
	handler := NewHTTPServer(mockServer)

	groups := []*ai.SkillGroup{
		{ID: "group-1", Name: "group-1", Namespace: "default"},
		{ID: "group-2", Name: "group-2", Namespace: "default"},
	}

	reqBody := map[string][]*ai.SkillGroup{
		"groups": groups,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// 模拟服务层成功响应
	mockServer.On("DeleteSkillGroups", mock.Anything, groups).Return(
		&MockResponse{
			Code: 200000,
			Info: "execute success",
			Size: 2,
		},
	)

	req := httptest.NewRequest("POST", "/skill/groups/delete", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handler.DeleteSkillGroups(restful.NewRequest(req), &restful.Response{ResponseWriter: resp})

	assert.Equal(t, http.StatusOK, resp.Code)

	var response MockResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint32(200000), response.Code)
	assert.Equal(t, "execute success", response.Info)
	assert.Equal(t, uint32(2), response.Size)

	mockServer.AssertExpectations(t)
}

// TestDeleteSkillGroups_NotFound 测试删除不存在的 SkillGroup
func TestDeleteSkillGroups_NotFound(t *testing.T) {
	mockServer := &MockSkillServer{}
	handler := NewHTTPServer(mockServer)

	groups := []*ai.SkillGroup{
		{ID: "non-existent", Name: "non-existent", Namespace: "default"},
	}

	// 模拟服务层返回不存在的错误
	mockServer.On("DeleteSkillGroups", mock.Anything, groups).Return(
		&MockResponse{
			Code: 400101,
			Info: "skill group not found",
		},
	)

	reqBody := map[string][]*ai.SkillGroup{
		"groups": groups,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/skill/groups/delete", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handler.DeleteSkillGroups(restful.NewRequest(req), &restful.Response{ResponseWriter: resp})

	assert.Equal(t, http.StatusOK, resp.Code)

	var response MockResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint32(400101), response.Code)
	assert.Equal(t, "skill group not found", response.Info)

	mockServer.AssertExpectations(t)
}

// TestSkillGroupsIntegration_CrudFlow 测试完整的 SkillGroup CRUD 流程
func TestSkillGroupsIntegration_CrudFlow(t *testing.T) {
	mockServer := &MockSkillServer{}
	handler := NewHTTPServer(mockServer)

	// 1. 创建 SkillGroup
	group := &ai.SkillGroup{
		ID:        "integration-group",
		Name:      "integration-test",
		Namespace: "default",
	}

	reqBody := map[string][]*ai.SkillGroup{
		"groups": {group},
	}
	jsonBody, _ := json.Marshal(reqBody)

	mockServer.On("CreateSkillGroups", mock.Anything, []*ai.SkillGroup{group}).Return(
		&MockResponse{
			Code: 200000,
			Info: "execute success",
			Size: 1,
		},
	)

	// 2. 查询 SkillGroup
	query := map[string]string{
		"namespace": "default",
		"name":      "integration-test",
	}
	mockServer.On("GetSkillGroups", mock.Anything, query).Return(
		&MockQueryResponse{
			Code:   200000,
			Info:   "execute success",
			Amount: 1,
			Size:   1,
			Data:   nil,
		},
	)

	// 3. 更新 SkillGroup
	updatedGroup := &ai.SkillGroup{
		ID:        "integration-group",
		Name:      "integration-test",
		Namespace: "default",
	}
	mockServer.On("UpdateSkillGroups", mock.Anything, []*ai.SkillGroup{updatedGroup}).Return(
		&MockResponse{
			Code: 200000,
			Info: "execute success",
			Size: 1,
		},
	)

	// 4. 删除 SkillGroup
	mockServer.On("DeleteSkillGroups", mock.Anything, []*ai.SkillGroup{group}).Return(
		&MockResponse{
			Code: 200000,
			Info: "execute success",
			Size: 1,
		},
	)

	// 执行测试步骤
	// Step 1: Create
	createReq := httptest.NewRequest("POST", "/skill/groups", strings.NewReader(string(jsonBody)))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	handler.CreateSkillGroups(restful.NewRequest(createReq), &restful.Response{ResponseWriter: createResp})
	assert.Equal(t, http.StatusOK, createResp.Code)

	// Step 2: Get
	getReq := httptest.NewRequest("GET", "/skill/groups?namespace=default&name=integration-test", nil)
	getResp := httptest.NewRecorder()
	handler.GetSkillGroups(restful.NewRequest(getReq), &restful.Response{ResponseWriter: getResp})
	assert.Equal(t, http.StatusOK, getResp.Code)

	// Step 3: Update
	updateReqBody, _ := json.Marshal(map[string][]*ai.SkillGroup{
		"groups": {updatedGroup},
	})
	updateReq := httptest.NewRequest("PUT", "/skill/groups", strings.NewReader(string(updateReqBody)))
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp := httptest.NewRecorder()
	handler.UpdateSkillGroups(restful.NewRequest(updateReq), &restful.Response{ResponseWriter: updateResp})
	assert.Equal(t, http.StatusOK, updateResp.Code)

	// Step 4: Delete
	deleteReqBody, _ := json.Marshal(map[string][]*ai.SkillGroup{
		"groups": {group},
	})
	deleteReq := httptest.NewRequest("POST", "/skill/groups/delete", strings.NewReader(string(deleteReqBody)))
	deleteReq.Header.Set("Content-Type", "application/json")
	deleteResp := httptest.NewRecorder()
	handler.DeleteSkillGroups(restful.NewRequest(deleteReq), &restful.Response{ResponseWriter: deleteResp})
	assert.Equal(t, http.StatusOK, deleteResp.Code)

	// 验证所有调用
	mockServer.AssertExpectations(t)
}

// TestSkillGroupsErrorHandling_InvalidSkills 测试包含无效技能的 SkillGroup
func TestSkillGroupsErrorHandling_InvalidSkills(t *testing.T) {
	mockServer := &MockSkillServer{}
	handler := NewHTTPServer(mockServer)

	group := &ai.SkillGroup{
		ID:        "invalid-skills-group",
		Name:      "invalid-skills",
		Namespace: "default",
	}

	// 模拟部分验证失败
	mockServer.On("CreateSkillGroups", mock.Anything, []*ai.SkillGroup{group}).Return(
		&MockResponse{
			Code: 400102,
			Info: "invalid skills in group",
		},
	)

	reqBody := map[string][]*ai.SkillGroup{
		"groups": {group},
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/skill/groups", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handler.CreateSkillGroups(restful.NewRequest(req), &restful.Response{ResponseWriter: resp})

	assert.Equal(t, http.StatusOK, resp.Code)

	var response MockResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint32(400102), response.Code)
	assert.Contains(t, response.Info, "invalid skills in group")

	mockServer.AssertExpectations(t)
}

// TestSkillGroupsEmptyResponse 测试空响应情况
func TestSkillGroupsEmptyResponse(t *testing.T) {
	mockServer := &MockSkillServer{}
	handler := NewHTTPServer(mockServer)

	query := map[string]string{
		"namespace": "empty-namespace",
	}

	// 模拟服务层返回空响应
	mockServer.On("GetSkillGroups", mock.Anything, query).Return(
		&MockQueryResponse{
			Code:   200000,
			Info:   "execute success",
			Amount: 0,
			Size:   0,
			Data:   nil,
		},
	)

	req := httptest.NewRequest("GET", "/skill/groups?namespace=empty-namespace", nil)
	resp := httptest.NewRecorder()

	handler.GetSkillGroups(restful.NewRequest(req), &restful.Response{ResponseWriter: resp})

	assert.Equal(t, http.StatusOK, resp.Code)

	var response MockQueryResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint32(200000), response.Code)
	assert.Equal(t, uint32(0), response.Amount)
	assert.Equal(t, uint32(0), response.Size)

	mockServer.AssertExpectations(t)
}