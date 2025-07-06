package ai

import "time"

// MCPServer 表结构
type MCPServer struct {
	ID          string    `json:"id"`          // mcp-server id
	Name        string    `json:"name"`        // mcp-server name, only under the namespace
	Namespace   string    `json:"namespace"`   // Namespace belongs to the mcp-server
	Ports       string    `json:"ports"`       // mcp-server will have a list of all port information
	Business    string    `json:"business"`    // mcp-server business information
	Department  string    `json:"department"`  // mcp-server department information
	Description string    `json:"description"` // Description information
	Revision    string    `json:"revision"`    // mcp-server version information
	Flag        int8      `json:"flag"`        // Logic delete flag, 0 means visible, 1 means logically deleted
	Reference   string    `json:"reference"`   // Actual service name pointed out
	Protocol    string    `json:"protocol"`    // mcp-server protocol
	CTime       time.Time `json:"ctime"`       // Create time
	MTime       time.Time `json:"mtime"`       // Last updated time
	ExportTo    string    `json:"export_to"`   // service export to some namespace
}

// TableName 设置表名
func (MCPServer) TableName() string {
	return "mcp_server"
}

// MCPServerTool 表结构
type MCPServerTool struct {
	ID           string    `json:"id"`            // mcp-server id
	MCPServerID  string    `json:"mcp_server_id"` // mcp-server id
	Name         string    `json:"name"`          // mcp-server name, only under the namespace
	Description  string    `json:"description"`   // Description information
	InputSchema  string    `json:"input_schema"`  // Input schema information
	OutputSchema string    `json:"output_schema"` // Output schema information
	Annotations  string    `json:"annotations"`   // Annotations information
	Flag         int8      `json:"flag"`          // Logic delete flag
	CTime        time.Time `json:"ctime"`         // Create time
	MTime        time.Time `json:"mtime"`         // Last updated time
}

// TableName 设置表名
func (MCPServerTool) TableName() string {
	return "mcp_server_tools"
}
