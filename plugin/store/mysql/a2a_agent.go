/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 */

package sqldb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/apis/store"
)

const (
	labelCreateA2AAgent = "createA2AAgent"
	labelUpdateA2AAgent = "updateA2AAgent"
	labelDeleteA2AAgent = "deleteA2AAgent"
)

func newA2AID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

type a2aAgentStore struct {
	master *BaseDB
	slave  *BaseDB
}

func newA2AAgentStore(master, slave *BaseDB) *a2aAgentStore {
	return &a2aAgentStore{master: master, slave: slave}
}

func (s *a2aAgentStore) CreateA2AAgent(agent *aitypes.A2AAgent) error {
	if agent == nil || agent.Name == "" || agent.Namespace == "" {
		return store.NewStatusError(store.EmptyParamsErr, "create a2a agent missing name or namespace")
	}
	if agent.Id == "" {
		agent.Id = newA2AID()
	}
	err := RetryTransaction(labelCreateA2AAgent, func() error {
		return s.createA2AAgent(agent)
	})
	return store.Error(err)
}

func (s *a2aAgentStore) createA2AAgent(agent *aitypes.A2AAgent) error {
	tx, err := s.master.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.insertA2AAgent(tx, agent); err != nil {
		return err
	}
	if err := s.replaceA2AAgentInterfaces(tx, agent); err != nil {
		return err
	}
	if err := s.replaceA2AAgentSkills(tx, agent); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *a2aAgentStore) UpdateA2AAgent(agent *aitypes.A2AAgent) error {
	if agent == nil || agent.Id == "" {
		return store.NewStatusError(store.EmptyParamsErr, "update a2a agent missing id")
	}
	err := RetryTransaction(labelUpdateA2AAgent, func() error {
		return s.updateA2AAgent(agent)
	})
	return store.Error(err)
}

func (s *a2aAgentStore) updateA2AAgent(agent *aitypes.A2AAgent) error {
	tx, err := s.master.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	metadataJSON := marshalStringMap(agent.Metadata)
	sqlText := `UPDATE a2a_agent SET name = ?, namespace = ?, visibility = ?, description = ?,
		version = ?, protocol_version = ?, provider_organization = ?, provider_url = ?,
		documentation_url = ?, icon_url = ?, business = ?, department = ?, backend_type = ?,
		backend_service_namespace = ?, backend_service_name = ?, backend_address = ?,
		preferred_interface_url = ?, preferred_protocol_binding = ?, preferred_protocol_version = ?,
		streaming = ?, push_notifications = ?, extended_agent_card = ?, raw_card_json = ?,
		source_type = ?, source_url = ?, last_fetch_status = ?, last_fetch_time = ?, metadata = ?,
		flag = ?, mtime = sysdate() WHERE id = ?`
	if _, err := tx.Exec(sqlText,
		agent.Name, agent.Namespace, agent.Visibility, agent.Description, agent.Version, agent.ProtocolVersion,
		agent.ProviderOrganization, agent.ProviderUrl, agent.DocumentationUrl, agent.IconUrl, agent.Business,
		agent.Department, agent.BackendType, agent.BackendServiceNamespace, agent.BackendServiceName,
		agent.BackendAddress, agent.PreferredInterfaceUrl, agent.PreferredProtocolBinding,
		agent.PreferredProtocolVersion, agent.Streaming, agent.PushNotifications, agent.ExtendedAgentCard,
		agent.RawCardJson, agent.SourceType, agent.SourceUrl, agent.LastFetchStatus, agent.LastFetchTime,
		metadataJSON, agent.Flag, agent.Id); err != nil {
		return err
	}
	if err := s.replaceA2AAgentInterfaces(tx, agent); err != nil {
		return err
	}
	if err := s.replaceA2AAgentSkills(tx, agent); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *a2aAgentStore) DeleteA2AAgent(id string) error {
	if id == "" {
		return store.NewStatusError(store.EmptyParamsErr, "delete a2a agent missing id")
	}
	err := RetryTransaction(labelDeleteA2AAgent, func() error {
		_, err := s.master.Exec(`UPDATE a2a_agent SET flag = 1, mtime = sysdate() WHERE id = ?`, id)
		return err
	})
	return store.Error(err)
}

func (s *a2aAgentStore) GetA2AAgent(id string) (*aitypes.A2AAgent, error) {
	if id == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get a2a agent missing id")
	}
	rows, err := s.queryA2AAgentsFrom(s.master, `WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return s.fetchA2AAgentRow(rows)
}

func (s *a2aAgentStore) GetA2AAgentByName(name, namespace string) (*aitypes.A2AAgent, error) {
	if name == "" || namespace == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get a2a agent missing name or namespace")
	}
	rows, err := s.queryA2AAgentsFrom(s.master,
		`WHERE name = ? AND namespace = ? AND flag != 1`, name, namespace)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return s.fetchA2AAgentRow(rows)
}

func (s *a2aAgentStore) GetMoreA2AAgents(mtime time.Time, firstUpdate bool) ([]*aitypes.A2AAgent, error) {
	sqlText := `WHERE mtime > FROM_UNIXTIME(?)`
	args := []interface{}{mtime.Unix()}
	if firstUpdate {
		sqlText += ` AND flag != 1`
	}
	rows, err := s.queryA2AAgents(sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	agents := make([]*aitypes.A2AAgent, 0)
	for rows.Next() {
		agent, err := s.scanA2AAgent(rows)
		if err != nil {
			return nil, err
		}
		if err := s.fillA2AAgentChildren(agent); err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	return agents, rows.Err()
}

func (s *a2aAgentStore) HasA2AAgent(id string) (bool, error) {
	return s.checkA2AAgentExists(`SELECT id FROM a2a_agent WHERE id = ?`, id)
}

func (s *a2aAgentStore) HasA2AAgentByName(name, namespace string) (bool, error) {
	return s.checkA2AAgentExists(`SELECT id FROM a2a_agent WHERE name = ? AND namespace = ? AND flag != 1`, name, namespace)
}

func (s *a2aAgentStore) HasA2AAgentByNameExcludeId(name, namespace, id string) (bool, error) {
	return s.checkA2AAgentExists(
		`SELECT id FROM a2a_agent WHERE name = ? AND namespace = ? AND id != ? AND flag != 1`,
		name, namespace, id)
}

func (s *a2aAgentStore) GetA2AAgentSkills(agentID string) ([]*aitypes.A2AAgentSkill, error) {
	return s.getA2AAgentSkills(agentID)
}

func (s *a2aAgentStore) QueryA2AAgents(query *aitypes.A2AAgentQuery) (uint32, []*aitypes.A2AAgent, error) {
	if query == nil {
		query = &aitypes.A2AAgentQuery{}
	}
	if query.Limit == 0 {
		query.Limit = 100
	}

	where, args := buildA2AAgentWhere(query)
	countRow := s.slave.QueryRow("SELECT COUNT(*) FROM a2a_agent "+where, args...)
	var count int
	if err := countRow.Scan(&count); err != nil {
		return 0, nil, err
	}

	queryArgs := append(args, query.Offset, query.Limit)
	rows, err := s.queryA2AAgents(where+` ORDER BY mtime DESC LIMIT ?, ?`, queryArgs...)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = rows.Close() }()

	agents := make([]*aitypes.A2AAgent, 0)
	for rows.Next() {
		agent, err := s.scanA2AAgent(rows)
		if err != nil {
			return 0, nil, err
		}
		if err := s.fillA2AAgentChildren(agent); err != nil {
			return 0, nil, err
		}
		agents = append(agents, agent)
	}
	if err := rows.Err(); err != nil {
		return 0, nil, err
	}
	return uint32(count), agents, nil
}

func (s *a2aAgentStore) insertA2AAgent(tx *BaseTx, agent *aitypes.A2AAgent) error {
	metadataJSON := marshalStringMap(agent.Metadata)
	sqlText := `INSERT INTO a2a_agent(id, name, namespace, visibility, description, version,
		protocol_version, provider_organization, provider_url, documentation_url, icon_url,
		business, department, backend_type, backend_service_namespace, backend_service_name,
		backend_address, preferred_interface_url, preferred_protocol_binding, preferred_protocol_version,
		streaming, push_notifications, extended_agent_card, raw_card_json, source_type, source_url,
		last_fetch_status, last_fetch_time, metadata, flag, ctime, mtime)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, sysdate(), sysdate())
		ON DUPLICATE KEY UPDATE flag = VALUES(flag), mtime = sysdate()`
	_, err := tx.Exec(sqlText,
		agent.Id, agent.Name, agent.Namespace, agent.Visibility, agent.Description, agent.Version,
		agent.ProtocolVersion, agent.ProviderOrganization, agent.ProviderUrl, agent.DocumentationUrl,
		agent.IconUrl, agent.Business, agent.Department, agent.BackendType, agent.BackendServiceNamespace,
		agent.BackendServiceName, agent.BackendAddress, agent.PreferredInterfaceUrl,
		agent.PreferredProtocolBinding, agent.PreferredProtocolVersion, agent.Streaming,
		agent.PushNotifications, agent.ExtendedAgentCard, agent.RawCardJson, agent.SourceType,
		agent.SourceUrl, agent.LastFetchStatus, agent.LastFetchTime, metadataJSON, agent.Flag)
	return err
}

func (s *a2aAgentStore) replaceA2AAgentInterfaces(tx *BaseTx, agent *aitypes.A2AAgent) error {
	if _, err := tx.Exec(`UPDATE a2a_agent_interface SET flag = 1, mtime = sysdate() WHERE agent_id = ?`, agent.Id); err != nil {
		return err
	}
	for _, item := range agent.Interfaces {
		if item.Id == "" {
			item.Id = newA2AID()
		}
		item.AgentId = agent.Id
		if _, err := tx.Exec(`INSERT INTO a2a_agent_interface(id, agent_id, url, protocol_binding,
			protocol_version, tenant, flag, ctime, mtime) VALUES(?, ?, ?, ?, ?, ?, ?, sysdate(), sysdate())
			ON DUPLICATE KEY UPDATE agent_id = VALUES(agent_id), url = VALUES(url),
			protocol_binding = VALUES(protocol_binding), protocol_version = VALUES(protocol_version),
			tenant = VALUES(tenant), flag = VALUES(flag), mtime = sysdate()`,
			item.Id, item.AgentId, item.Url, item.ProtocolBinding, item.ProtocolVersion, item.Tenant, item.Flag); err != nil {
			return err
		}
	}
	return nil
}

func (s *a2aAgentStore) replaceA2AAgentSkills(tx *BaseTx, agent *aitypes.A2AAgent) error {
	if _, err := tx.Exec(`UPDATE a2a_agent_skill SET flag = 1, mtime = sysdate() WHERE agent_id = ?`, agent.Id); err != nil {
		return err
	}
	for _, item := range agent.Skills {
		if item.Id == "" {
			item.Id = newA2AID()
		}
		item.AgentId = agent.Id
		if _, err := tx.Exec(`INSERT INTO a2a_agent_skill(id, agent_id, skill_id, name, description,
			tags, examples, input_modes, output_modes, security_requirements, flag, ctime, mtime)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, sysdate(), sysdate())
			ON DUPLICATE KEY UPDATE agent_id = VALUES(agent_id), skill_id = VALUES(skill_id),
			name = VALUES(name), description = VALUES(description), tags = VALUES(tags),
			examples = VALUES(examples), input_modes = VALUES(input_modes), output_modes = VALUES(output_modes),
			security_requirements = VALUES(security_requirements), flag = VALUES(flag), mtime = sysdate()`,
			item.Id, item.AgentId, item.SkillId, item.Name, item.Description, marshalStringSlice(item.Tags),
			marshalStringSlice(item.Examples), marshalStringSlice(item.InputModes), marshalStringSlice(item.OutputModes),
			item.SecurityRequirementsJson, item.Flag); err != nil {
			return err
		}
	}
	return nil
}

func (s *a2aAgentStore) queryA2AAgents(where string, args ...interface{}) (*sql.Rows, error) {
	return s.queryA2AAgentsFrom(s.slave, where, args...)
}

func (s *a2aAgentStore) queryA2AAgentsFrom(db *BaseDB, where string, args ...interface{}) (*sql.Rows, error) {
	sqlText := `SELECT id, name, namespace, visibility, description, version, protocol_version,
		provider_organization, provider_url, documentation_url, icon_url, business, department,
		backend_type, backend_service_namespace, backend_service_name, backend_address,
		preferred_interface_url, preferred_protocol_binding, preferred_protocol_version, streaming,
		push_notifications, extended_agent_card, raw_card_json, source_type, source_url,
		last_fetch_status, last_fetch_time, metadata, flag, unix_timestamp(ctime), unix_timestamp(mtime)
		FROM a2a_agent ` + where
	return db.Query(sqlText, args...)
}

func (s *a2aAgentStore) fetchA2AAgentRow(rows *sql.Rows) (*aitypes.A2AAgent, error) {
	if !rows.Next() {
		return nil, nil
	}
	agent, err := s.scanA2AAgent(rows)
	if err != nil {
		return nil, err
	}
	if err := s.fillA2AAgentChildren(agent); err != nil {
		return nil, err
	}
	return agent, rows.Err()
}

func (s *a2aAgentStore) scanA2AAgent(rows *sql.Rows) (*aitypes.A2AAgent, error) {
	agent := &aitypes.A2AAgent{}
	var ctimeSec, mtimeSec int64
	var metadataJSON string
	err := rows.Scan(&agent.Id, &agent.Name, &agent.Namespace, &agent.Visibility, &agent.Description,
		&agent.Version, &agent.ProtocolVersion, &agent.ProviderOrganization, &agent.ProviderUrl,
		&agent.DocumentationUrl, &agent.IconUrl, &agent.Business, &agent.Department, &agent.BackendType,
		&agent.BackendServiceNamespace, &agent.BackendServiceName, &agent.BackendAddress,
		&agent.PreferredInterfaceUrl, &agent.PreferredProtocolBinding, &agent.PreferredProtocolVersion,
		&agent.Streaming, &agent.PushNotifications, &agent.ExtendedAgentCard, &agent.RawCardJson,
		&agent.SourceType, &agent.SourceUrl, &agent.LastFetchStatus, &agent.LastFetchTime,
		&metadataJSON, &agent.Flag, &ctimeSec, &mtimeSec)
	if err != nil {
		return nil, err
	}
	agent.Metadata = unmarshalStringMap(metadataJSON)
	agent.Ctime = time.Unix(ctimeSec, 0).Format("2006-01-02 15:04:05")
	agent.Mtime = time.Unix(mtimeSec, 0).Format("2006-01-02 15:04:05")
	return agent, nil
}

func (s *a2aAgentStore) fillA2AAgentChildren(agent *aitypes.A2AAgent) error {
	interfaces, err := s.getA2AAgentInterfaces(agent.Id)
	if err != nil {
		return err
	}
	skills, err := s.getA2AAgentSkills(agent.Id)
	if err != nil {
		return err
	}
	agent.Interfaces = interfaces
	agent.Skills = skills
	return nil
}

func (s *a2aAgentStore) getA2AAgentInterfaces(agentID string) ([]*aitypes.A2AAgentInterface, error) {
	rows, err := s.slave.Query(`SELECT id, agent_id, url, protocol_binding, protocol_version, tenant,
		flag, unix_timestamp(ctime), unix_timestamp(mtime) FROM a2a_agent_interface
		WHERE agent_id = ? AND flag != 1`, agentID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]*aitypes.A2AAgentInterface, 0)
	for rows.Next() {
		item := &aitypes.A2AAgentInterface{}
		var ctimeSec, mtimeSec int64
		if err := rows.Scan(&item.Id, &item.AgentId, &item.Url, &item.ProtocolBinding, &item.ProtocolVersion,
			&item.Tenant, &item.Flag, &ctimeSec, &mtimeSec); err != nil {
			return nil, err
		}
		item.Ctime = time.Unix(ctimeSec, 0).Format("2006-01-02 15:04:05")
		item.Mtime = time.Unix(mtimeSec, 0).Format("2006-01-02 15:04:05")
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *a2aAgentStore) getA2AAgentSkills(agentID string) ([]*aitypes.A2AAgentSkill, error) {
	rows, err := s.slave.Query(`SELECT id, agent_id, skill_id, name, description, tags, examples,
		input_modes, output_modes, security_requirements, flag, unix_timestamp(ctime), unix_timestamp(mtime)
		FROM a2a_agent_skill WHERE agent_id = ? AND flag != 1`, agentID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]*aitypes.A2AAgentSkill, 0)
	for rows.Next() {
		item := &aitypes.A2AAgentSkill{}
		var ctimeSec, mtimeSec int64
		var tags, examples, inputModes, outputModes string
		if err := rows.Scan(&item.Id, &item.AgentId, &item.SkillId, &item.Name, &item.Description,
			&tags, &examples, &inputModes, &outputModes, &item.SecurityRequirementsJson, &item.Flag,
			&ctimeSec, &mtimeSec); err != nil {
			return nil, err
		}
		item.Tags = unmarshalStringSlice(tags)
		item.Examples = unmarshalStringSlice(examples)
		item.InputModes = unmarshalStringSlice(inputModes)
		item.OutputModes = unmarshalStringSlice(outputModes)
		item.Ctime = time.Unix(ctimeSec, 0).Format("2006-01-02 15:04:05")
		item.Mtime = time.Unix(mtimeSec, 0).Format("2006-01-02 15:04:05")
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *a2aAgentStore) checkA2AAgentExists(query string, args ...interface{}) (bool, error) {
	rows, err := s.slave.Query(query, args...)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	return rows.Next(), nil
}

func buildA2AAgentWhere(query *aitypes.A2AAgentQuery) (string, []interface{}) {
	where := "WHERE flag != 1"
	args := make([]interface{}, 0)
	if query.Name != "" {
		where += " AND name = ?"
		args = append(args, query.Name)
	}
	if query.Namespace != "" {
		where += " AND namespace = ?"
		args = append(args, query.Namespace)
	}
	if query.Business != "" {
		where += " AND business = ?"
		args = append(args, query.Business)
	}
	if query.Department != "" {
		where += " AND department = ?"
		args = append(args, query.Department)
	}
	if query.ProtocolBinding != "" {
		where += " AND preferred_protocol_binding = ?"
		args = append(args, query.ProtocolBinding)
	}
	if query.BackendType != "" {
		where += " AND backend_type = ?"
		args = append(args, query.BackendType)
	}
	if query.BackendServiceNamespace != "" {
		where += " AND backend_service_namespace = ?"
		args = append(args, query.BackendServiceNamespace)
	}
	if query.BackendServiceName != "" {
		where += " AND backend_service_name = ?"
		args = append(args, query.BackendServiceName)
	}
	if query.Streaming != nil {
		where += " AND streaming = ?"
		args = append(args, *query.Streaming)
	}
	if query.PushNotifications != nil {
		where += " AND push_notifications = ?"
		args = append(args, *query.PushNotifications)
	}
	return where, args
}

func marshalStringSlice(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	data, _ := json.Marshal(values)
	return string(data)
}

func unmarshalStringSlice(value string) []string {
	if value == "" {
		return nil
	}
	var values []string
	if err := json.Unmarshal([]byte(value), &values); err != nil {
		return nil
	}
	return values
}

func marshalStringMap(values map[string]string) string {
	if len(values) == 0 {
		return "{}"
	}
	data, _ := json.Marshal(values)
	return string(data)
}

func unmarshalStringMap(value string) map[string]string {
	if value == "" {
		return nil
	}
	values := make(map[string]string)
	if err := json.Unmarshal([]byte(value), &values); err != nil {
		return nil
	}
	return values
}

func a2aAgentCardFromRaw(raw string) (map[string]interface{}, error) {
	if raw == "" {
		return map[string]interface{}{}, nil
	}
	var card map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &card); err != nil {
		return nil, fmt.Errorf("invalid a2a agent card json: %w", err)
	}
	return card, nil
}
