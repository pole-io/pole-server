# Skill HTTP API 接口设计

## 1. 概述

本文档定义 Skill 模块的 HTTP API 接口规范，包括 Skill 和 SkillGroup 的 CRUD 操作。

## 2. 目录结构

```
plugin/apiserver/httpserver/
├── skill/
│   ├── server.go           # HTTP 服务器入口
│   ├── skill_access.go     # Skill API 处理器
│   ├── skill_group_access.go # SkillGroup API 处理器
│   ├── proto.go            # 请求/响应类型定义
│   └── log.go              # 日志
└── docs/
    └── skill_apidoc.go     # API 文档
```

## 3. API 端点设计

### 3.1 Skill API

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/skills` | 创建 Skill |
| PUT | `/skills` | 更新 Skill |
| POST | `/skills/delete` | 删除 Skill |
| GET | `/skills` | 查询 Skill 列表 |
| GET | `/skills/all` | 获取所有 Skill |
| GET | `/skills/count` | 获取 Skill 数量 |

### 3.2 SkillGroup API

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/skill/groups` | 创建 SkillGroup |
| PUT | `/skill/groups` | 更新 SkillGroup |
| POST | `/skill/groups/delete` | 删除 SkillGroup |
| GET | `/skill/groups` | 查询 SkillGroup 列表 |

### 3.3 SkillVersion API (P1)

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/skill/versions` | 创建 SkillVersion |
| PUT | `/skill/versions/activate` | 激活 SkillVersion |
| GET | `/skill/versions` | 查询 SkillVersion 列表 |

### 3.4 SkillSubscription API (P1)

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/skill/subscriptions` | 创建订阅 |
| POST | `/skill/subscriptions/delete` | 删除订阅 |
| GET | `/skill/subscriptions` | 查询订阅列表 |

## 4. 请求/响应格式

### 4.1 Skill 数据结构

```go
// proto.go
type Skill struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Namespace   string            `json:"namespace"`
    Description string            `json:"description"`
    SkillType   string            `json:"skill_type"`   // function, tool, agent, workflow
    InputSchema string            `json:"input_schema"`
    OutputSchema string           `json:"output_schema"`
    Author      string            `json:"author"`
    Business    string            `json:"business"`
    Department  string            `json:"department"`
    Metadata    map[string]string `json:"metadata"`
    Tags        []string          `json:"tags"`
    Owner       string            `json:"owner"`
    Revision    string            `json:"revision"`
    CTime       time.Time         `json:"ctime"`
    MTime       time.Time         `json:"mtime"`
}
```

### 4.2 SkillGroup 数据结构

```go
type SkillGroup struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Namespace   string            `json:"namespace"`
    Description string            `json:"description"`
    Skills      []*SkillReference `json:"skills"`
    Metadata    map[string]string `json:"metadata"`
    Owner       string            `json:"owner"`
    CTime       time.Time         `json:"ctime"`
    MTime       time.Time         `json:"mtime"`
}

type SkillReference struct {
    SkillID   string `json:"skill_id"`
    SkillName string `json:"skill_name"`
    Version   uint64 `json:"version"` // 0 表示使用活跃版本
}
```

## 5. API 处理器实现

### 5.1 路由注册

```go
// server.go
func (h *HTTPServer) addSkillAccess(ws *restful.WebService) {
    // Skill API
    ws.Route(docs.EnrichCreateSkillsApiDocs(ws.POST("/skills").To(h.CreateSkills)))
    ws.Route(docs.EnrichUpdateSkillsApiDocs(ws.PUT("/skills").To(h.UpdateSkills)))
    ws.Route(docs.EnrichDeleteSkillsApiDocs(ws.POST("/skills/delete").To(h.DeleteSkills)))
    ws.Route(docs.EnrichGetSkillsApiDocs(ws.GET("/skills").To(h.GetSkills)))
    ws.Route(docs.EnrichGetAllSkillsApiDocs(ws.GET("/skills/all").To(h.GetAllSkills)))
    ws.Route(docs.EnrichGetSkillsCountApiDocs(ws.GET("/skills/count").To(h.GetSkillsCount)))

    // SkillGroup API
    ws.Route(docs.EnrichCreateSkillGroupsApiDocs(ws.POST("/skill/groups").To(h.CreateSkillGroups)))
    ws.Route(docs.EnrichUpdateSkillGroupsApiDocs(ws.PUT("/skill/groups").To(h.UpdateSkillGroups)))
    ws.Route(docs.EnrichDeleteSkillGroupsApiDocs(ws.POST("/skill/groups/delete").To(h.DeleteSkillGroups)))
    ws.Route(docs.EnrichGetSkillGroupsApiDocs(ws.GET("/skill/groups").To(h.GetSkillGroups)))
}
```

### 5.2 CreateSkills 实现

```go
// skill_access.go
func (h *HTTPServer) CreateSkills(req *restful.Request, rsp *restful.Response) {
    handler := &httpcommon.Handler{
        Request:  req,
        Response: rsp,
    }

    var skills SkillArr
    ctx, err := handler.ParseArray(func() proto.Message {
        msg := &Skill{}
        skills = append(skills, msg)
        return msg
    })
    if err != nil {
        handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(
            apimodel.Code_ParseException, err.Error()))
        return
    }

    // 调用 Skill 服务
    ret := h.skillServer.CreateSkills(ctx, skills)
    handler.WriteHeaderAndProto(ret)
}
```

### 5.3 GetSkills 实现

```go
// skill_access.go
func (h *HTTPServer) GetSkills(req *restful.Request, rsp *restful.Response) {
    handler := &httpcommon.Handler{
        Request:  req,
        Response: rsp,
    }

    queryParams := httpcommon.ParseQueryParams(req)
    ctx := handler.ParseHeaderContext()

    // 必填参数检查
    if queryParams["namespace"] == "" {
        handler.WriteHeaderAndProto(api.NewBatchQueryResponseWithMsg(
            apimodel.Code_BadRequest, "namespace is required"))
        return
    }

    ret := h.skillServer.GetSkills(ctx, queryParams)
    handler.WriteHeaderAndProto(ret)
}
```

### 5.4 DeleteSkills 实现

```go
// skill_access.go
func (h *HTTPServer) DeleteSkills(req *restful.Request, rsp *restful.Response) {
    handler := &httpcommon.Handler{
        Request:  req,
        Response: rsp,
    }

    var skills SkillArr
    ctx, err := handler.ParseArray(func() proto.Message {
        msg := &Skill{}
        skills = append(skills, msg)
        return msg
    })
    if err != nil {
        handler.WriteHeaderAndProto(api.NewBatchWriteResponseWithMsg(
            apimodel.Code_ParseException, err.Error()))
        return
    }

    ret := h.skillServer.DeleteSkills(ctx, skills)
    if code := api.CalcCode(ret); code != http.StatusOK {
        handler.WriteHeaderAndProto(ret)
        return
    }

    handler.WriteHeaderAndProto(ret)
}
```

## 6. 服务层接口

### 6.1 SkillServer 接口

```go
// pkg/skill/server.go
type SkillServer interface {
    // Skill CRUD
    CreateSkills(ctx context.Context, skills []*Skill) *apimodel.BatchWriteResponse
    UpdateSkills(ctx context.Context, skills []*Skill) *apimodel.BatchWriteResponse
    DeleteSkills(ctx context.Context, skills []*Skill) *apimodel.BatchWriteResponse
    GetSkills(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse
    GetAllSkills(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse
    GetSkillsCount(ctx context.Context) *apimodel.BatchQueryResponse

    // SkillGroup CRUD
    CreateSkillGroups(ctx context.Context, groups []*SkillGroup) *apimodel.BatchWriteResponse
    UpdateSkillGroups(ctx context.Context, groups []*SkillGroup) *apimodel.BatchWriteResponse
    DeleteSkillGroups(ctx context.Context, groups []*SkillGroup) *apimodel.BatchWriteResponse
    GetSkillGroups(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse
}
```

## 7. 查询参数

### 7.1 GetSkills 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| namespace | string | 是 | 命名空间 |
| name | string | 否 | Skill 名称（模糊匹配） |
| skill_type | string | 否 | Skill 类型 |
| owner | string | 否 | 所有者 |
| offset | int | 否 | 偏移量 |
| limit | int | 否 | 限制数量 |

### 7.2 GetSkillGroups 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| namespace | string | 是 | 命名空间 |
| name | string | 否 | Group 名称（模糊匹配） |
| owner | string | 否 | 所有者 |
| offset | int | 否 | 偏移量 |
| limit | int | 否 | 限制数量 |

## 8. 错误码

| 错误码 | 说明 |
|--------|------|
| 200000 | 执行成功 |
| 400000 | 请求参数错误 |
| 400001 | 数据解析失败 |
| 400100 | Skill 名称已存在 |
| 400101 | Skill 不存在 |
| 400102 | 命名空间不存在 |
| 400103 | 无效的 Skill 类型 |
| 401000 | 未授权 |
| 403000 | 权限不足 |
| 500000 | 服务器内部错误 |

## 9. 测试用例设计要求

### 9.1 API 测试

```go
// skill_access_test.go

// 测试场景：
// 1. TestCreateSkills_Success - 创建成功
// 2. TestCreateSkills_EmptyName - 名称为空
// 3. TestCreateSkills_EmptyNamespace - 命名空间为空
// 4. TestCreateSkills_DuplicateName - 名称重复
// 5. TestUpdateSkills_Success - 更新成功
// 6. TestUpdateSkills_NotFound - Skill 不存在
// 7. TestDeleteSkills_Success - 删除成功
// 8. TestDeleteSkills_NotFound - Skill 不存在
// 9. TestGetSkills_Success - 查询成功
// 10. TestGetSkills_WithFilter - 带过滤条件查询
// 11. TestGetSkills_EmptyNamespace - 命名空间为空
```

### 9.2 集成测试

- 完整 CRUD 流程测试
- 权限控制测试
- 并发操作测试

## 10. 实现优先级

### P0 - 必须实现

1. `plugin/apiserver/httpserver/skill/server.go`
2. `plugin/apiserver/httpserver/skill/skill_access.go`
3. `plugin/apiserver/httpserver/skill/proto.go`
4. `pkg/skill/server.go` - 服务层

### P1 - 尽快实现

1. `plugin/apiserver/httpserver/skill/skill_group_access.go`
2. `plugin/apiserver/httpserver/skill/skill_version_access.go`
3. `plugin/apiserver/httpserver/skill/skill_subscription_access.go`

## 11. 实现依赖

- `github.com/emicklei/go-restful/v3` - HTTP 框架
- `github.com/pole-io/specification` - API 规范
- `plugin/apiserver/httpserver/utils` - HTTP 工具
- `pkg/skill` - Skill 服务层
