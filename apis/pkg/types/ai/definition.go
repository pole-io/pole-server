package ai

import "time"

// ResourceKind 区分逻辑定义对应的 AI Registry 资源类型。
type ResourceKind string

const (
	// ResourceKindMCPServer 表示 MCP Server 逻辑定义。
	ResourceKindMCPServer ResourceKind = "mcp_server"
	// ResourceKindA2AAgent 表示 A2A Agent 逻辑定义。
	ResourceKindA2AAgent ResourceKind = "a2a_agent"
)

// ResourceDefinition 是同一个 MCP Server 或 A2A Agent 在多个环境部署之间
// 共享的控制面身份。运行时注册、发现和调用均不使用该身份。
type ResourceDefinition struct {
	Kind        ResourceKind `json:"kind"`
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Owner       string       `json:"owner"`
	Business    string       `json:"business"`
	Department  string       `json:"department"`
	Revision    string       `json:"revision"`
	CreateTime  time.Time    `json:"ctime"`
	ModifyTime  time.Time    `json:"mtime"`
}

// EnvironmentBinding 将一个现有注册表环境实例显式关联到控制面定义。
// ResourceID 是权威关联键，Namespace 和 ResourceName 是管理视图快照。
type EnvironmentBinding struct {
	Kind         ResourceKind `json:"kind"`
	DefinitionID string       `json:"definition_id"`
	ResourceID   string       `json:"resource_id"`
	Namespace    string       `json:"namespace"`
	ResourceName string       `json:"resource_name"`
}

// ResourceDefinitionDeleteRequest 表示批量删除逻辑定义的请求。
type ResourceDefinitionDeleteRequest struct {
	DefinitionIDs []string `json:"definition_ids"`
}

// ResourceDefinitionListResponse 表示逻辑定义分页查询结果。
type ResourceDefinitionListResponse struct {
	Code   uint32                `json:"code"`
	Info   string                `json:"info"`
	Amount uint32                `json:"amount"`
	Size   uint32                `json:"size"`
	Data   []*ResourceDefinition `json:"data"`
}

// EnvironmentBindingListResponse 表示逻辑定义下的环境绑定列表。
type EnvironmentBindingListResponse struct {
	Code   uint32                `json:"code"`
	Info   string                `json:"info"`
	Amount uint32                `json:"amount"`
	Size   uint32                `json:"size"`
	Data   []*EnvironmentBinding `json:"data"`
}
