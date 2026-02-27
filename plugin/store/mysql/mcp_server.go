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

package sqldb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/apis/store"
)

const (
	labelCreateMCPServer    = "createMCPServer"
	labelUpdateMCPServer    = "updateMCPServer"
	labelDeleteMCPServer    = "deleteMCPServer"
	labelCreateMCPServerTool = "createMCPServerTool"
	labelUpdateMCPServerTool = "updateMCPServerTool"
	labelDeleteMCPServerTool = "deleteMCPServerTool"
)

// mcpServerStore 实现 MCP Server 存储接口
type mcpServerStore struct {
	master *BaseDB
	slave  *BaseDB
}

// newMCPServerStore 创建 MCP Server 存储
func newMCPServerStore(master, slave *BaseDB) *mcpServerStore {
	return &mcpServerStore{
		master: master,
		slave:  slave,
	}
}

// ===== MCP Server 存储接口 =====

// CreateMCPServer 创建 MCP Server
func (m *mcpServerStore) CreateMCPServer(server *ai.MCPServer) error {
	if server.ID == "" || server.Name == "" || server.Namespace == "" {
		return store.NewStatusError(store.EmptyParamsErr, fmt.Sprintf(
			"create mcp server missing some params, id is %s, name is %s, namespace is %s",
			server.ID, server.Name, server.Namespace))
	}

	// 如果没有提供 ID，生成新的 UUID
	if server.ID == "" {
		server.ID = uuid.New().String()
	}
	// 如果没有提供 Revision，生成新的版本号
	if server.Revision == "" {
		server.Revision = uuid.New().String()
	}

	err := RetryTransaction(labelCreateMCPServer, func() error {
		return m.createMCPServer(server)
	})
	return store.Error(err)
}

func (m *mcpServerStore) createMCPServer(server *ai.MCPServer) error {
	tx, err := m.master.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 插入 mcp_server 表
	if err := m.insertMCPServerMain(tx, server); err != nil {
		return err
	}

	return tx.Commit()
}

func (m *mcpServerStore) insertMCPServerMain(tx *BaseTx, server *ai.MCPServer) error {
	sql := `INSERT INTO mcp_server(id, name, namespace, ports, business, department, description,
		revision, flag, reference, protocol, ctime, mtime, export_to)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, sysdate(), sysdate(), ?)`

	_, err := tx.Exec(sql,
		server.ID,
		server.Name,
		server.Namespace,
		server.Ports,
		server.Business,
		server.Department,
		server.Description,
		server.Revision,
		server.Flag,
		server.Reference,
		server.Protocol,
		server.ExportTo,
	)
	if err != nil {
		log.Errorf("[Store][database] insert mcp server err: %s", err.Error())
		return err
	}
	return nil
}

// UpdateMCPServer 更新 MCP Server
func (m *mcpServerStore) UpdateMCPServer(server *ai.MCPServer) error {
	if server.ID == "" {
		return store.NewStatusError(store.EmptyParamsErr, "update mcp server missing id")
	}

	err := RetryTransaction(labelUpdateMCPServer, func() error {
		return m.updateMCPServer(server)
	})
	return store.Error(err)
}

func (m *mcpServerStore) updateMCPServer(server *ai.MCPServer) error {
	if server.Revision == "" {
		server.Revision = uuid.New().String()
	}

	sql := `UPDATE mcp_server SET name = ?, namespace = ?, ports = ?, business = ?,
		department = ?, description = ?, revision = ?, reference = ?, protocol = ?,
		mtime = sysdate(), export_to = ? WHERE id = ?`

	_, err := m.master.Exec(sql,
		server.Name,
		server.Namespace,
		server.Ports,
		server.Business,
		server.Department,
		server.Description,
		server.Revision,
		server.Reference,
		server.Protocol,
		server.ExportTo,
		server.ID,
	)
	if err != nil {
		log.Errorf("[Store][database] update mcp server err: %s", err.Error())
		return err
	}
	return nil
}

// DeleteMCPServer 删除 MCP Server (逻辑删除)
func (m *mcpServerStore) DeleteMCPServer(id string) error {
	if id == "" {
		return store.NewStatusError(store.EmptyParamsErr, "delete mcp server missing id")
	}

	err := RetryTransaction(labelDeleteMCPServer, func() error {
		return m.deleteMCPServer(id)
	})
	return store.Error(err)
}

func (m *mcpServerStore) deleteMCPServer(id string) error {
	sql := `UPDATE mcp_server SET flag = 1, mtime = sysdate() WHERE id = ?`
	_, err := m.master.Exec(sql, id)
	if err != nil {
		log.Errorf("[Store][database] logical delete mcp server err: %s", err.Error())
		return err
	}
	return nil
}

// GetMCPServer 获取 MCP Server
func (m *mcpServerStore) GetMCPServer(id string) (*ai.MCPServer, error) {
	if id == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get mcp server missing id")
	}

	rows, err := m.slave.Query(`SELECT id, name, namespace, ports, business, department, description,
		revision, flag, reference, protocol, unix_timestamp(ctime), unix_timestamp(mtime), export_to
		FROM mcp_server WHERE id = ?`, id)
	if err != nil {
		log.Errorf("[Store][database] get mcp server query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return fetchMCPServerRow(rows)
}

// GetMCPServerByName 获取 MCP Server by name and namespace
func (m *mcpServerStore) GetMCPServerByName(name, namespace string) (*ai.MCPServer, error) {
	if name == "" || namespace == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get mcp server missing name or namespace")
	}

	rows, err := m.slave.Query(`SELECT id, name, namespace, ports, business, department, description,
		revision, flag, reference, protocol, unix_timestamp(ctime), unix_timestamp(mtime), export_to
		FROM mcp_server WHERE name = ? AND namespace = ? AND flag != 1`, name, namespace)
	if err != nil {
		log.Errorf("[Store][database] get mcp server by name query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return fetchMCPServerRow(rows)
}

// GetMoreMCPServers 增量获取 MCP Servers (供 cache 使用)
func (m *mcpServerStore) GetMoreMCPServers(mtime time.Time, firstUpdate bool) ([]*ai.MCPServer, error) {
	cacheSql := `SELECT id, name, namespace, ports, business, department, description,
		revision, flag, reference, protocol, unix_timestamp(ctime), unix_timestamp(mtime), export_to
		FROM mcp_server WHERE mtime > FROM_UNIXTIME(?)`

	if firstUpdate {
		cacheSql += " AND flag != 1"
	}

	rows, err := m.slave.Query(cacheSql, timeToTimestamp(mtime))
	if err != nil {
		log.Errorf("[Store][database] get more mcp servers query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var servers []*ai.MCPServer
	for rows.Next() {
		server, err := fetchMCPServerRow(rows)
		if err != nil {
			return nil, err
		}
		if server != nil {
			servers = append(servers, server)
		}
	}

	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] get more mcp servers rows err: %s", err.Error())
		return nil, err
	}

	return servers, nil
}

// HasMCPServer 检查 MCP Server 是否存在
func (m *mcpServerStore) HasMCPServer(id string) (bool, error) {
	return m.checkMCPServerExists(`SELECT id FROM mcp_server WHERE id = ?`, id)
}

// HasMCPServerByName 检查 MCP Server 是否存在 by name
func (m *mcpServerStore) HasMCPServerByName(name, namespace string) (bool, error) {
	return m.checkMCPServerExists(`SELECT id FROM mcp_server WHERE name = ? AND namespace = ? AND flag != 1`, name, namespace)
}

// HasMCPServerByNameExcludeId 检查 MCP Server 是否存在 by name exclude id
func (m *mcpServerStore) HasMCPServerByNameExcludeId(name, namespace, id string) (bool, error) {
	return m.checkMCPServerExists(
		`SELECT id FROM mcp_server WHERE name = ? AND namespace = ? AND id != ? AND flag != 1`,
		name, namespace, id)
}

func (m *mcpServerStore) checkMCPServerExists(query string, args ...interface{}) (bool, error) {
	row := m.master.QueryRow(query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

// ===== MCP Server Tool 存储接口 =====

// CreateMCPServerTool 创建 MCP Server Tool
func (m *mcpServerStore) CreateMCPServerTool(tool *ai.MCPServerTool) error {
	if tool.MCPServerID == "" || tool.Name == "" {
		return store.NewStatusError(store.EmptyParamsErr, "create mcp server tool missing mcp_server_id or name")
	}

	// 如果没有提供 ID，生成新的 UUID
	if tool.ID == "" {
		tool.ID = uuid.New().String()
	}

	sql := `INSERT INTO mcp_server_tools(id, mcp_server_id, name, description, input_schema, output_schema, annotations, flag, ctime, mtime)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, sysdate(), sysdate())`

	_, err := m.master.Exec(sql,
		tool.ID,
		tool.MCPServerID,
		tool.Name,
		tool.Description,
		tool.InputSchema,
		tool.OutputSchema,
		tool.Annotations,
		tool.Flag,
	)
	if err != nil {
		log.Errorf("[Store][database] insert mcp server tool err: %s", err.Error())
		return err
	}
	return nil
}

// UpdateMCPServerTool 更新 MCP Server Tool
func (m *mcpServerStore) UpdateMCPServerTool(tool *ai.MCPServerTool) error {
	if tool.ID == "" {
		return store.NewStatusError(store.EmptyParamsErr, "update mcp server tool missing id")
	}

	sql := `UPDATE mcp_server_tools SET mcp_server_id = ?, name = ?, description = ?,
		input_schema = ?, output_schema = ?, annotations = ?, mtime = sysdate() WHERE id = ?`

	_, err := m.master.Exec(sql,
		tool.MCPServerID,
		tool.Name,
		tool.Description,
		tool.InputSchema,
		tool.OutputSchema,
		tool.Annotations,
		tool.ID,
	)
	if err != nil {
		log.Errorf("[Store][database] update mcp server tool err: %s", err.Error())
		return err
	}
	return nil
}

// DeleteMCPServerTool 删除 MCP Server Tool (逻辑删除)
func (m *mcpServerStore) DeleteMCPServerTool(id string) error {
	if id == "" {
		return store.NewStatusError(store.EmptyParamsErr, "delete mcp server tool missing id")
	}

	sql := `UPDATE mcp_server_tools SET flag = 1, mtime = sysdate() WHERE id = ?`
	_, err := m.master.Exec(sql, id)
	if err != nil {
		log.Errorf("[Store][database] logical delete mcp server tool err: %s", err.Error())
		return err
	}
	return nil
}

// GetMCPServerTool 获取 MCP Server Tool
func (m *mcpServerStore) GetMCPServerTool(id string) (*ai.MCPServerTool, error) {
	if id == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get mcp server tool missing id")
	}

	rows, err := m.slave.Query(`SELECT id, mcp_server_id, name, description, input_schema, output_schema, annotations, flag, unix_timestamp(ctime), unix_timestamp(mtime)
		FROM mcp_server_tools WHERE id = ?`, id)
	if err != nil {
		log.Errorf("[Store][database] get mcp server tool query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return fetchMCPServerToolRow(rows)
}

// GetMCPServerToolsByServerID 获取 MCP Server 的所有 Tools
func (m *mcpServerStore) GetMCPServerToolsByServerID(serverID string) ([]*ai.MCPServerTool, error) {
	if serverID == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get mcp server tools missing server id")
	}

	rows, err := m.slave.Query(`SELECT id, mcp_server_id, name, description, input_schema, output_schema, annotations, flag, unix_timestamp(ctime), unix_timestamp(mtime)
		FROM mcp_server_tools WHERE mcp_server_id = ? AND flag != 1`, serverID)
	if err != nil {
		log.Errorf("[Store][database] get mcp server tools query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tools []*ai.MCPServerTool
	for rows.Next() {
		tool, err := fetchMCPServerToolRow(rows)
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] get mcp server tools rows err: %s", err.Error())
		return nil, err
	}

	return tools, nil
}

// GetMCPServerTools 增量获取 MCP Server Tools (供 cache 使用)
func (m *mcpServerStore) GetMCPServerTools(mtime time.Time, firstUpdate bool) ([]*ai.MCPServerTool, error) {
	cacheSql := `SELECT id, mcp_server_id, name, description, input_schema, output_schema, annotations, flag, unix_timestamp(ctime), unix_timestamp(mtime)
		FROM mcp_server_tools WHERE mtime > FROM_UNIXTIME(?)`

	if firstUpdate {
		cacheSql += " AND flag != 1"
	}

	rows, err := m.slave.Query(cacheSql, timeToTimestamp(mtime))
	if err != nil {
		log.Errorf("[Store][database] get more mcp server tools query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tools []*ai.MCPServerTool
	for rows.Next() {
		tool, err := fetchMCPServerToolRow(rows)
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] get more mcp server tools rows err: %s", err.Error())
		return nil, err
	}

	return tools, nil
}

// ===== Helper Functions =====

func fetchMCPServerRow(rows *sql.Rows) (*ai.MCPServer, error) {
	var server ai.MCPServer
	var ctime, mtime int64
	var exportTo sql.NullString

	err := rows.Scan(
		&server.ID,
		&server.Name,
		&server.Namespace,
		&server.Ports,
		&server.Business,
		&server.Department,
		&server.Description,
		&server.Revision,
		&server.Flag,
		&server.Reference,
		&server.Protocol,
		&ctime,
		&mtime,
		&exportTo,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorf("[Store][database] fetch mcp server row scan err: %s", err.Error())
		return nil, err
	}

	server.CTime = time.Unix(ctime, 0)
	server.MTime = time.Unix(mtime, 0)
	if exportTo.Valid {
		server.ExportTo = exportTo.String
	}

	return &server, nil
}

func fetchMCPServerToolRow(rows *sql.Rows) (*ai.MCPServerTool, error) {
	var tool ai.MCPServerTool
	var ctime, mtime int64

	err := rows.Scan(
		&tool.ID,
		&tool.MCPServerID,
		&tool.Name,
		&tool.Description,
		&tool.InputSchema,
		&tool.OutputSchema,
		&tool.Annotations,
		&tool.Flag,
		&ctime,
		&mtime,
	)
	if err != nil {
		log.Errorf("[Store][database] fetch mcp server tool row scan err: %s", err.Error())
		return nil, err
	}

	tool.CTime = time.Unix(ctime, 0)
	tool.MTime = time.Unix(mtime, 0)

	return &tool, nil
}
