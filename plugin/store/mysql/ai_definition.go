package sqldb

import (
	"database/sql"
	"errors"
	"time"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/apis/store"
)

type aiDefinitionTables struct {
	definition  string
	environment string
}

type aiDefinitionStore struct {
	master *BaseDB
	slave  *BaseDB
}

func newAIResourceDefinitionStore(master, slave *BaseDB) *aiDefinitionStore {
	return &aiDefinitionStore{master: master, slave: slave}
}

func aiTables(kind aitypes.ResourceKind) (aiDefinitionTables, error) {
	switch kind {
	case aitypes.ResourceKindMCPServer:
		return aiDefinitionTables{definition: "mcp_server_definition", environment: "mcp_server"}, nil
	case aitypes.ResourceKindA2AAgent:
		return aiDefinitionTables{definition: "a2a_agent_definition", environment: "a2a_agent"}, nil
	default:
		return aiDefinitionTables{}, store.NewStatusError(store.EmptyParamsErr, "invalid ai resource kind")
	}
}

func (s *aiDefinitionStore) CreateAIResourceDefinition(definition *aitypes.ResourceDefinition) error {
	if definition == nil || definition.ID == "" || definition.Name == "" || definition.Revision == "" {
		return store.NewStatusError(store.EmptyParamsErr, "create ai resource definition missing parameters")
	}
	tables, err := aiTables(definition.Kind)
	if err != nil {
		return err
	}
	_, err = s.master.Exec(`INSERT INTO `+tables.definition+`
		(id, name, description, owner, business, department, revision, flag, ctime, mtime)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0, sysdate(), sysdate())`,
		definition.ID, definition.Name, definition.Description, definition.Owner,
		definition.Business, definition.Department, definition.Revision)
	return store.Error(err)
}

func (s *aiDefinitionStore) UpdateAIResourceDefinition(
	definition *aitypes.ResourceDefinition, previousRevision string) error {
	if definition == nil || definition.ID == "" || definition.Name == "" ||
		definition.Revision == "" || previousRevision == "" {
		return store.NewStatusError(store.EmptyParamsErr, "update ai resource definition missing parameters")
	}
	tables, err := aiTables(definition.Kind)
	if err != nil {
		return err
	}
	result, err := s.master.Exec(`UPDATE `+tables.definition+` SET name = ?, description = ?, owner = ?,
		business = ?, department = ?, revision = ?, mtime = sysdate()
		WHERE id = ? AND flag = 0 AND revision = ?`,
		definition.Name, definition.Description, definition.Owner, definition.Business,
		definition.Department, definition.Revision, definition.ID, previousRevision)
	if err != nil {
		return store.Error(err)
	}
	return requireAIDefinitionAffected(result, "ai resource definition revision conflict")
}

func (s *aiDefinitionStore) DeleteAIResourceDefinition(kind aitypes.ResourceKind, id string) error {
	if id == "" {
		return store.NewStatusError(store.EmptyParamsErr, "delete ai resource definition missing id")
	}
	tables, err := aiTables(kind)
	if err != nil {
		return err
	}
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()

	var lockedID string
	if err := tx.QueryRow(`SELECT id FROM `+tables.definition+
		` WHERE id = ? AND flag = 0 FOR UPDATE`, id).Scan(&lockedID); errors.Is(err, sql.ErrNoRows) {
		return store.NewStatusError(store.NotFoundService, "ai resource definition not found")
	} else if err != nil {
		return store.Error(err)
	}
	var count uint32
	if _, err := tx.Exec(`UPDATE `+tables.environment+
		` SET definition_id = NULL WHERE definition_id = ? AND flag = 1`, id); err != nil {
		return store.Error(err)
	}
	if err := tx.QueryRow(`SELECT COUNT(*) FROM `+tables.environment+
		` WHERE definition_id = ? AND flag != 1`, id).Scan(&count); err != nil {
		return store.Error(err)
	}
	if count > 0 {
		return store.NewStatusError(store.DataConflictErr, "ai resource definition still has environment bindings")
	}
	if _, err := tx.Exec(`UPDATE `+tables.definition+
		` SET flag = 1, mtime = sysdate() WHERE id = ? AND flag = 0`, id); err != nil {
		return store.Error(err)
	}
	return store.Error(tx.Commit())
}

func (s *aiDefinitionStore) GetAIResourceDefinition(
	kind aitypes.ResourceKind, id string) (*aitypes.ResourceDefinition, error) {
	if id == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get ai resource definition missing id")
	}
	tables, err := aiTables(kind)
	if err != nil {
		return nil, err
	}
	return scanAIResourceDefinition(kind, s.slave.QueryRow(aiDefinitionSelect(tables)+
		` WHERE id = ? AND flag = 0`, id))
}

func (s *aiDefinitionStore) ListAIResourceDefinitions(
	kind aitypes.ResourceKind, name string, offset, limit uint32) (
	uint32, []*aitypes.ResourceDefinition, error) {
	tables, err := aiTables(kind)
	if err != nil {
		return 0, nil, err
	}
	if limit == 0 {
		limit = 100
	}
	where := ` WHERE flag = 0`
	args := []any{}
	if name != "" {
		where += ` AND name LIKE ?`
		args = append(args, "%"+name+"%")
	}
	var total uint32
	if err := s.slave.QueryRow(`SELECT COUNT(*) FROM `+tables.definition+where, args...).Scan(&total); err != nil {
		return 0, nil, store.Error(err)
	}
	args = append(args, offset, limit)
	rows, err := s.slave.Query(aiDefinitionSelect(tables)+where+
		` ORDER BY mtime DESC, id LIMIT ?, ?`, args...)
	if err != nil {
		return 0, nil, store.Error(err)
	}
	defer rows.Close()
	definitions := make([]*aitypes.ResourceDefinition, 0)
	for rows.Next() {
		definition, err := scanAIResourceDefinition(kind, rows)
		if err != nil {
			return 0, nil, err
		}
		definitions = append(definitions, definition)
	}
	return total, definitions, store.Error(rows.Err())
}

func (s *aiDefinitionStore) BindAIResourceEnvironment(
	binding *aitypes.EnvironmentBinding, revision string) error {
	if binding == nil || binding.DefinitionID == "" || binding.ResourceID == "" || revision == "" {
		return store.NewStatusError(store.EmptyParamsErr, "bind ai resource environment missing parameters")
	}
	tables, err := aiTables(binding.Kind)
	if err != nil {
		return err
	}
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()

	var definitionID string
	if err := tx.QueryRow(`SELECT id FROM `+tables.definition+
		` WHERE id = ? AND flag = 0 FOR UPDATE`, binding.DefinitionID).Scan(&definitionID); errors.Is(err, sql.ErrNoRows) {
		return store.NewStatusError(store.NotFoundService, "ai resource definition not found")
	} else if err != nil {
		return store.Error(err)
	}
	var namespace, resourceName string
	var existingDefinitionID sql.NullString
	if err := tx.QueryRow(`SELECT namespace, name, definition_id FROM `+tables.environment+
		` WHERE id = ? AND flag != 1 FOR UPDATE`, binding.ResourceID).
		Scan(&namespace, &resourceName, &existingDefinitionID); errors.Is(err, sql.ErrNoRows) {
		return store.NewStatusError(store.NotFoundService, "ai environment resource not found")
	} else if err != nil {
		return store.Error(err)
	}
	if existingDefinitionID.Valid {
		if existingDefinitionID.String == binding.DefinitionID {
			binding.Namespace = namespace
			binding.ResourceName = resourceName
			return store.Error(tx.Commit())
		}
		return store.NewStatusError(store.DataConflictErr, "ai environment resource is already bound")
	}
	result, err := tx.Exec(`UPDATE `+tables.environment+
		` SET definition_id = ?, mtime = sysdate() WHERE id = ? AND flag != 1 AND definition_id IS NULL`,
		binding.DefinitionID, binding.ResourceID)
	if err != nil {
		if store.Code(store.Error(err)) == store.DuplicateEntryErr {
			return store.NewStatusError(store.DataConflictErr,
				"ai resource definition already has an environment binding")
		}
		return store.Error(err)
	}
	if err := requireAIDefinitionAffected(result, "ai environment binding conflict"); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE `+tables.definition+
		` SET revision = ?, mtime = sysdate() WHERE id = ? AND flag = 0`,
		revision, binding.DefinitionID); err != nil {
		return store.Error(err)
	}
	binding.Namespace = namespace
	binding.ResourceName = resourceName
	return store.Error(tx.Commit())
}

func (s *aiDefinitionStore) UnbindAIResourceEnvironment(
	kind aitypes.ResourceKind, definitionID, resourceID, revision string) error {
	if definitionID == "" || resourceID == "" || revision == "" {
		return store.NewStatusError(store.EmptyParamsErr, "unbind ai resource environment missing parameters")
	}
	tables, err := aiTables(kind)
	if err != nil {
		return err
	}
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()

	var lockedID string
	if err := tx.QueryRow(`SELECT id FROM `+tables.definition+
		` WHERE id = ? AND flag = 0 FOR UPDATE`, definitionID).Scan(&lockedID); errors.Is(err, sql.ErrNoRows) {
		return store.NewStatusError(store.NotFoundService, "ai resource definition not found")
	} else if err != nil {
		return store.Error(err)
	}
	var actualDefinitionID sql.NullString
	if err := tx.QueryRow(`SELECT definition_id FROM `+tables.environment+
		` WHERE id = ? FOR UPDATE`, resourceID).Scan(&actualDefinitionID); errors.Is(err, sql.ErrNoRows) {
		return store.NewStatusError(store.NotFoundService, "ai environment binding not found")
	} else if err != nil {
		return store.Error(err)
	}
	if !actualDefinitionID.Valid {
		return store.NewStatusError(store.NotFoundService, "ai environment binding not found")
	}
	if actualDefinitionID.String != definitionID {
		return store.NewStatusError(store.DataConflictErr, "ai environment binding belongs to another definition")
	}
	if _, err := tx.Exec(`UPDATE `+tables.environment+
		` SET definition_id = NULL, mtime = sysdate() WHERE id = ? AND definition_id = ?`,
		resourceID, definitionID); err != nil {
		return store.Error(err)
	}
	if _, err := tx.Exec(`UPDATE `+tables.definition+
		` SET revision = ?, mtime = sysdate() WHERE id = ? AND flag = 0`,
		revision, definitionID); err != nil {
		return store.Error(err)
	}
	return store.Error(tx.Commit())
}

func (s *aiDefinitionStore) ListAIResourceEnvironmentBindings(
	kind aitypes.ResourceKind, definitionID string) ([]*aitypes.EnvironmentBinding, error) {
	tables, err := aiTables(kind)
	if err != nil {
		return nil, err
	}
	rows, err := s.slave.Query(`SELECT definition_id, id, namespace, name FROM `+
		tables.environment+` WHERE definition_id = ? AND flag != 1 ORDER BY namespace, name`, definitionID)
	if err != nil {
		return nil, store.Error(err)
	}
	defer rows.Close()
	return scanAIEnvironmentBindings(kind, rows)
}

func (s *aiDefinitionStore) GetAIResourceEnvironmentBinding(
	kind aitypes.ResourceKind, resourceID string) (*aitypes.EnvironmentBinding, error) {
	tables, err := aiTables(kind)
	if err != nil {
		return nil, err
	}
	binding := &aitypes.EnvironmentBinding{Kind: kind}
	err = s.slave.QueryRow(`SELECT definition_id, id, namespace, name FROM `+
		tables.environment+` WHERE id = ? AND definition_id IS NOT NULL AND flag != 1`, resourceID).
		Scan(&binding.DefinitionID, &binding.ResourceID, &binding.Namespace, &binding.ResourceName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, store.Error(err)
	}
	return binding, nil
}

func (s *aiDefinitionStore) CountAIResourcesByNamespace(namespace string) (uint32, error) {
	if namespace == "" {
		return 0, store.NewStatusError(store.EmptyParamsErr, "count ai resources missing namespace")
	}
	var count uint32
	err := s.master.QueryRow(`SELECT
		(SELECT COUNT(*) FROM mcp_server WHERE namespace = ? AND flag != 1) +
		(SELECT COUNT(*) FROM a2a_agent WHERE namespace = ? AND flag != 1)`,
		namespace, namespace).Scan(&count)
	return count, store.Error(err)
}

func (s *aiDefinitionStore) CountAIBackendReferences(serviceID string) (uint32, error) {
	if serviceID == "" {
		return 0, store.NewStatusError(store.EmptyParamsErr, "count ai backend references missing service id")
	}
	var count uint32
	err := s.master.QueryRow(`SELECT
		(SELECT COUNT(*) FROM mcp_server AS ai
			INNER JOIN service AS backend ON backend.id = ?
			WHERE ai.flag != 1 AND (
				ai.backend_service_id = backend.id OR (
					ai.backend_service_id IS NULL
					AND ai.backend_type = 'service'
					AND ai.namespace = backend.namespace
					AND ai.backend_service_namespace = backend.namespace
					AND ai.backend_service_name = backend.name
					AND IFNULL(backend.reference, '') = ''
				)
			)) +
		(SELECT COUNT(*) FROM a2a_agent AS ai
			INNER JOIN service AS backend ON backend.id = ?
			WHERE ai.flag != 1 AND (
				ai.backend_service_id = backend.id OR (
					ai.backend_service_id IS NULL
					AND ai.backend_type = 'service'
					AND ai.namespace = backend.namespace
					AND ai.backend_service_namespace = backend.namespace
					AND ai.backend_service_name = backend.name
					AND IFNULL(backend.reference, '') = ''
				)
			))`,
		serviceID, serviceID).Scan(&count)
	return count, store.Error(err)
}

func aiDefinitionSelect(tables aiDefinitionTables) string {
	return `SELECT id, name, IFNULL(description, ''), IFNULL(owner, ''),
		IFNULL(business, ''), IFNULL(department, ''), revision,
		UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime) FROM ` + tables.definition
}

func scanAIResourceDefinition(
	kind aitypes.ResourceKind, row rowScanner) (*aitypes.ResourceDefinition, error) {
	definition := &aitypes.ResourceDefinition{Kind: kind}
	var ctime, mtime int64
	if err := row.Scan(&definition.ID, &definition.Name, &definition.Description, &definition.Owner,
		&definition.Business, &definition.Department, &definition.Revision, &ctime, &mtime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, store.Error(err)
	}
	definition.CreateTime = time.Unix(ctime, 0)
	definition.ModifyTime = time.Unix(mtime, 0)
	return definition, nil
}

func scanAIEnvironmentBindings(
	kind aitypes.ResourceKind, rows *sql.Rows) ([]*aitypes.EnvironmentBinding, error) {
	bindings := make([]*aitypes.EnvironmentBinding, 0)
	for rows.Next() {
		binding := &aitypes.EnvironmentBinding{Kind: kind}
		if err := rows.Scan(&binding.DefinitionID, &binding.ResourceID,
			&binding.Namespace, &binding.ResourceName); err != nil {
			return nil, store.Error(err)
		}
		bindings = append(bindings, binding)
	}
	return bindings, store.Error(rows.Err())
}

func requireAIDefinitionAffected(result sql.Result, message string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return store.Error(err)
	}
	if affected == 0 {
		return store.NewStatusError(store.DataConflictErr, message)
	}
	return nil
}
