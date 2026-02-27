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
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/apis/store"
)

const (
	labelCreateSkill            = "createSkill"
	labelUpdateSkill            = "updateSkill"
	labelDeleteSkill            = "deleteSkill"
	labelCreateSkillGroup       = "createSkillGroup"
	labelUpdateSkillGroup       = "updateSkillGroup"
	labelDeleteSkillGroup       = "deleteSkillGroup"
	labelCreateSkillVersion     = "createSkillVersion"
	labelUpdateSkillVersion     = "updateSkillVersion"
	labelDeleteSkillVersion     = "deleteSkillVersion"
	labelCreateSkillSubscription = "createSkillSubscription"
	labelUpdateSkillSubscription = "updateSkillSubscription"
	labelDeleteSkillSubscription = "deleteSkillSubscription"
)

// skillStore implements store.SkillStore
type skillStore struct {
	master *BaseDB
	slave  *BaseDB
}

// newSkillStore creates a new skill store
func newSkillStore(master, slave *BaseDB) *skillStore {
	return &skillStore{
		master: master,
		slave:  slave,
	}
}

// CreateSkill creates a new skill
func (s *skillStore) CreateSkill(skill *ai.Skill) error {
	if skill.ID == "" {
		skill.ID = uuid.New().String()
	}
	if skill.Revision == "" {
		skill.Revision = uuid.New().String()
	}

	err := RetryTransaction(labelCreateSkill, func() error {
		return s.createSkill(skill)
	})
	return store.Error(err)
}

func (s *skillStore) createSkill(skill *ai.Skill) error {
	tx, err := s.master.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 插入 skill 表
	if err := s.insertSkillMain(tx, skill); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *skillStore) insertSkillMain(tx *BaseTx, skill *ai.Skill) error {
	sql := `INSERT INTO skill(id, name, namespace, description, input_schema, output_schema,
		skill_type, author, business, department, metadata, flag, revision, ctime, mtime)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, sysdate(), sysdate())`

	_, err := tx.Exec(sql,
		skill.ID,
		skill.Name,
		skill.Namespace,
		skill.Description,
		skill.InputSchema,
		skill.OutputSchema,
		skill.SkillType,
		skill.Author,
		skill.Business,
		skill.Department,
		nil2JSONString(skill.Metadata),
		skill.Flag,
		skill.Revision,
	)
	if err != nil {
		log.Errorf("[Store][database] insert skill err: %s", err.Error())
		return err
	}
	return nil
}

// UpdateSkill updates an existing skill
func (s *skillStore) UpdateSkill(skill *ai.Skill) error {
	if skill.ID == "" {
		return store.NewStatusError(store.EmptyParamsErr, "update skill missing id")
	}

	err := RetryTransaction(labelUpdateSkill, func() error {
		return s.updateSkill(skill)
	})
	return store.Error(err)
}

func (s *skillStore) updateSkill(skill *ai.Skill) error {
	sql := `UPDATE skill SET name = ?, namespace = ?, description = ?, input_schema = ?,
		output_schema = ?, skill_type = ?, author = ?, business = ?, department = ?,
		metadata = ?, revision = ?, mtime = sysdate() WHERE id = ?`

	_, err := s.master.Exec(sql,
		skill.Name,
		skill.Namespace,
		skill.Description,
		skill.InputSchema,
		skill.OutputSchema,
		skill.SkillType,
		skill.Author,
		skill.Business,
		skill.Department,
		nil2JSONString(skill.Metadata),
		skill.Revision,
		skill.ID,
	)
	if err != nil {
		log.Errorf("[Store][database] update skill err: %s", err.Error())
		return err
	}
	return nil
}

// DeleteSkill deletes a skill (logical delete)
func (s *skillStore) DeleteSkill(id string) error {
	if id == "" {
		return store.NewStatusError(store.EmptyParamsErr, "delete skill missing id")
	}

	err := RetryTransaction(labelDeleteSkill, func() error {
		return s.deleteSkill(id)
	})
	return store.Error(err)
}

func (s *skillStore) deleteSkill(id string) error {
	sql := `UPDATE skill SET flag = 1, mtime = sysdate() WHERE id = ?`
	_, err := s.master.Exec(sql, id)
	if err != nil {
		log.Errorf("[Store][database] logical delete skill err: %s", err.Error())
		return err
	}
	return nil
}

// GetSkill gets a skill by ID
func (s *skillStore) GetSkill(id string) (*ai.Skill, error) {
	if id == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get skill missing id")
	}

	rows, err := s.slave.Query(`SELECT id, name, namespace, description, input_schema, output_schema,
		skill_type, author, business, department, metadata, flag, revision,
		unix_timestamp(ctime), unix_timestamp(mtime)
		FROM skill WHERE id = ?`, id)
	if err != nil {
		log.Errorf("[Store][database] get skill query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return fetchSkillRow(rows)
}

// GetSkillByName gets a skill by name and namespace
func (s *skillStore) GetSkillByName(name, namespace string) (*ai.Skill, error) {
	if name == "" || namespace == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get skill missing name or namespace")
	}

	rows, err := s.slave.Query(`SELECT id, name, namespace, description, input_schema, output_schema,
		skill_type, author, business, department, metadata, flag, revision,
		unix_timestamp(ctime), unix_timestamp(mtime)
		FROM skill WHERE name = ? AND namespace = ? AND flag != 1`, name, namespace)
	if err != nil {
		log.Errorf("[Store][database] get skill by name query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return fetchSkillRow(rows)
}

// GetMoreSkills gets skills updated after mtime (for cache increment update)
func (s *skillStore) GetMoreSkills(mtime time.Time, firstUpdate bool) ([]*ai.Skill, error) {
	cacheSql := `SELECT id, name, namespace, description, input_schema, output_schema,
		skill_type, author, business, department, metadata, flag, revision,
		unix_timestamp(ctime), unix_timestamp(mtime)
		FROM skill WHERE mtime > FROM_UNIXTIME(?)`

	if firstUpdate {
		cacheSql += " AND flag != 1"
	}

	rows, err := s.slave.Query(cacheSql, timeToTimestamp(mtime))
	if err != nil {
		log.Errorf("[Store][database] get more skills query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var skills []*ai.Skill
	for rows.Next() {
		skill, err := fetchSkillRow(rows)
		if err != nil {
			return nil, err
		}
		if skill != nil {
			skills = append(skills, skill)
		}
	}

	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] get more skills rows err: %s", err.Error())
		return nil, err
	}

	return skills, nil
}

// HasSkill checks if a skill exists by ID
func (s *skillStore) HasSkill(id string) (bool, error) {
	return s.checkSkillExists(`SELECT id FROM skill WHERE id = ?`, id)
}

// HasSkillByName checks if a skill exists by name
func (s *skillStore) HasSkillByName(name, namespace string) (bool, error) {
	return s.checkSkillExists(`SELECT id FROM skill WHERE name = ? AND namespace = ? AND flag != 1`, name, namespace)
}

// HasSkillByNameExcludeId checks if a skill exists by name, excluding a specific ID
func (s *skillStore) HasSkillByNameExcludeId(name, namespace, id string) (bool, error) {
	return s.checkSkillExists(
		`SELECT id FROM skill WHERE name = ? AND namespace = ? AND id != ? AND flag != 1`,
		name, namespace, id)
}

func (s *skillStore) checkSkillExists(query string, args ...interface{}) (bool, error) {
	row := s.master.QueryRow(query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

// ===== Helper Functions =====

func fetchSkillRow(rows *sql.Rows) (*ai.Skill, error) {
	var skill ai.Skill
	var ctime, mtime int64
	var metadataStr string

	err := rows.Scan(
		&skill.ID,
		&skill.Name,
		&skill.Namespace,
		&skill.Description,
		&skill.InputSchema,
		&skill.OutputSchema,
		&skill.SkillType,
		&skill.Author,
		&skill.Business,
		&skill.Department,
		&metadataStr,
		&skill.Flag,
		&skill.Revision,
		&ctime,
		&mtime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorf("[Store][database] fetch skill row scan err: %s", err.Error())
		return nil, err
	}

	skill.CTime = time.Unix(ctime, 0)
	skill.MTime = time.Unix(mtime, 0)
	skill.Metadata = jsonString2Map(metadataStr)

	return &skill, nil
}

func (s *skillStore) hasSkillByName(tx *BaseTx, name, namespace string) (bool, error) {
	query := `SELECT id FROM skill WHERE name = ? AND namespace = ? AND flag != 1`
	row := tx.QueryRow(query, name, namespace)
	var count int
	if err := row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

// ===== Helper Functions =====

// nil2JSONString converts nil or empty map to "{}" for MySQL storage
func nil2JSONString(m map[string]string) string {
	if m == nil {
		return "{}"
	}
	data, _ := json.Marshal(m)
	return string(data)
}

// jsonString2Map converts JSON string from MySQL to map[string]string
func jsonString2Map(str string) map[string]string {
	if str == "" || str == "null" {
		return nil
	}
	m := make(map[string]string)
	_ = json.Unmarshal([]byte(str), &m)
	return m
}

// ===== SkillGroup Store Implementation =====

type skillGroupStore struct {
	master *BaseDB
	slave  *BaseDB
}

func newSkillGroupStore(master, slave *BaseDB) *skillGroupStore {
	return &skillGroupStore{
		master: master,
		slave:  slave,
	}
}

// CreateSkillGroup creates a new skill group
func (s *skillGroupStore) CreateSkillGroup(group *ai.SkillGroup) (*ai.SkillGroup, error) {
	if group.ID == "" {
		group.ID = uuid.New().String()
	}

	err := RetryTransaction(labelCreateSkillGroup, func() error {
		return s.createSkillGroup(group)
	})
	if err != nil {
		return nil, err
	}
	return s.GetSkillGroup(group.Namespace, group.Name)
}

func (s *skillGroupStore) createSkillGroup(group *ai.SkillGroup) error {
	tx, err := s.master.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Check if skill group already exists
	exists, err := s.hasSkillGroupByName(tx, group.Name, group.Namespace)
	if err != nil {
		return err
	}
	if exists {
		return store.NewStatusError(store.DuplicateEntryErr, fmt.Sprintf(
			"skill group %s/%s already exists", group.Namespace, group.Name))
	}

	sql := `INSERT INTO skill_group(id, name, namespace, comment, metadata, owner, business, department, flag, ctime, mtime)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, sysdate(), sysdate())`

	_, err = tx.Exec(sql,
		group.ID,
		group.Name,
		group.Namespace,
		group.Comment,
		nil2JSONString(group.Metadata),
		group.Owner,
		group.Business,
		group.Department,
		group.Flag,
	)
	if err != nil {
		log.Errorf("[Store][database] insert skill group err: %s", err.Error())
		return err
	}

	return tx.Commit()
}

func (s *skillGroupStore) hasSkillGroupByName(tx *BaseTx, name, namespace string) (bool, error) {
	query := `SELECT id FROM skill_group WHERE name = ? AND namespace = ? AND flag != 1`
	row := tx.QueryRow(query, name, namespace)
	var count int
	if err := row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

// UpdateSkillGroup updates an existing skill group
func (s *skillGroupStore) UpdateSkillGroup(group *ai.SkillGroup) error {
	if group.ID == "" {
		return store.NewStatusError(store.EmptyParamsErr, "update skill group missing id")
	}

	err := RetryTransaction(labelUpdateSkillGroup, func() error {
		return s.updateSkillGroup(group)
	})
	return store.Error(err)
}

func (s *skillGroupStore) updateSkillGroup(group *ai.SkillGroup) error {
	sql := `UPDATE skill_group SET name = ?, namespace = ?, comment = ?, metadata = ?,
		owner = ?, business = ?, department = ?, mtime = sysdate() WHERE id = ?`

	_, err := s.master.Exec(sql,
		group.Name,
		group.Namespace,
		group.Comment,
		nil2JSONString(group.Metadata),
		group.Owner,
		group.Business,
		group.Department,
		group.ID,
	)
	if err != nil {
		log.Errorf("[Store][database] update skill group err: %s", err.Error())
		return err
	}
	return nil
}

// DeleteSkillGroup deletes a skill group (logical delete)
func (s *skillGroupStore) DeleteSkillGroup(namespace, name string) error {
	err := RetryTransaction(labelDeleteSkillGroup, func() error {
		return s.deleteSkillGroup(namespace, name)
	})
	return store.Error(err)
}

func (s *skillGroupStore) deleteSkillGroup(namespace, name string) error {
	sql := `UPDATE skill_group SET flag = 1, mtime = sysdate() WHERE namespace = ? AND name = ?`
	_, err := s.master.Exec(sql, namespace, name)
	if err != nil {
		log.Errorf("[Store][database] logical delete skill group err: %s", err.Error())
		return err
	}
	return nil
}

// GetSkillGroup gets a skill group by namespace and name
func (s *skillGroupStore) GetSkillGroup(namespace, name string) (*ai.SkillGroup, error) {
	if namespace == "" || name == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get skill group missing namespace or name")
	}

	rows, err := s.slave.Query(`SELECT id, name, namespace, comment, metadata, owner, business, department, flag,
		unix_timestamp(ctime), unix_timestamp(mtime)
		FROM skill_group WHERE namespace = ? AND name = ? AND flag != 1`, namespace, name)
	if err != nil {
		log.Errorf("[Store][database] get skill group query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return fetchSkillGroupRow(rows)
}

// GetMoreSkillGroups gets skill groups updated after mtime (for cache increment update)
func (s *skillGroupStore) GetMoreSkillGroups(firstUpdate bool, mtime time.Time) ([]*ai.SkillGroup, error) {
	cacheSql := `SELECT id, name, namespace, comment, metadata, owner, business, department, flag,
		unix_timestamp(ctime), unix_timestamp(mtime)
		FROM skill_group WHERE mtime > FROM_UNIXTIME(?)`

	if firstUpdate {
		cacheSql += " AND flag != 1"
	}

	rows, err := s.slave.Query(cacheSql, timeToTimestamp(mtime))
	if err != nil {
		log.Errorf("[Store][database] get more skill groups query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var groups []*ai.SkillGroup
	for rows.Next() {
		group, err := fetchSkillGroupRow(rows)
		if err != nil {
			return nil, err
		}
		if group != nil {
			groups = append(groups, group)
		}
	}

	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] get more skill groups rows err: %s", err.Error())
		return nil, err
	}

	return groups, nil
}

// CountSkillGroups gets the count of skill groups in a namespace
func (s *skillGroupStore) CountSkillGroups(namespace string) (uint64, error) {
	if namespace == "" {
		return 0, store.NewStatusError(store.EmptyParamsErr, "count skill groups missing namespace")
	}

	var count uint64
	err := s.master.QueryRow(`SELECT COUNT(*) FROM skill_group WHERE namespace = ? AND flag != 1`, namespace).Scan(&count)
	if err != nil {
		log.Errorf("[Store][database] count skill groups err: %s", err.Error())
		return 0, err
	}
	return count, nil
}

// ===== Helper Functions for SkillGroup =====

func fetchSkillGroupRow(rows *sql.Rows) (*ai.SkillGroup, error) {
	var group ai.SkillGroup
	var ctime, mtime int64
	var metadataStr string

	err := rows.Scan(
		&group.ID,
		&group.Name,
		&group.Namespace,
		&group.Comment,
		&metadataStr,
		&group.Owner,
		&group.Business,
		&group.Department,
		&group.Flag,
		&ctime,
		&mtime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorf("[Store][database] fetch skill group row scan err: %s", err.Error())
		return nil, err
	}

	group.CTime = time.Unix(ctime, 0)
	group.MTime = time.Unix(mtime, 0)
	group.Metadata = jsonString2Map(metadataStr)

	return &group, nil
}

// ===== Skill Version Store =====

// skillVersionStore implements store.SkillVersionStore
type skillVersionStore struct {
	master *BaseDB
	slave  *BaseDB
}

// newSkillVersionStore creates a new skill version store
func newSkillVersionStore(master, slave *BaseDB) *skillVersionStore {
	return &skillVersionStore{
		master: master,
		slave:  slave,
	}
}

// CreateSkillVersion creates a new skill version
func (s *skillVersionStore) CreateSkillVersion(version *ai.SkillVersion) error {
	if version.ID == "" {
		version.ID = uuid.New().String()
	}

	err := RetryTransaction(labelCreateSkillVersion, func() error {
		return s.createSkillVersion(version)
	})
	return store.Error(err)
}

func (s *skillVersionStore) createSkillVersion(version *ai.SkillVersion) error {
	tx, err := s.master.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 插入 skill_version 表
	if err := s.insertSkillVersionMain(tx, version); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *skillVersionStore) insertSkillVersionMain(tx *BaseTx, version *ai.SkillVersion) error {
	sql := `INSERT INTO skill_version(id, skill_id, skill_name, namespace, version, comment,
		input_schema, output_schema, skill_type, metadata, active, flag, ctime, mtime)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, sysdate(), sysdate())`

	_, err := tx.Exec(sql,
		version.ID,
		version.SkillID,
		version.SkillName,
		version.Namespace,
		version.Version,
		version.Comment,
		version.InputSchema,
		version.OutputSchema,
		version.SkillType,
		nil2JSONStringForSkillVersion(version.Metadata),
		ifThenElse(version.Active, 1, 0),
		version.Flag,
	)
	if err != nil {
		log.Errorf("[Store][database] insert skill version err: %s", err.Error())
		return err
	}
	return nil
}

// UpdateSkillVersion updates an existing skill version
func (s *skillVersionStore) UpdateSkillVersion(version *ai.SkillVersion) error {
	if version.ID == "" {
		return store.NewStatusError(store.EmptyParamsErr, "update skill version missing id")
	}

	err := RetryTransaction(labelUpdateSkillVersion, func() error {
		return s.updateSkillVersion(version)
	})
	return store.Error(err)
}

func (s *skillVersionStore) updateSkillVersion(version *ai.SkillVersion) error {
	sql := `UPDATE skill_version SET version = ?, comment = ?, input_schema = ?,
		output_schema = ?, skill_type = ?, metadata = ?, active = ?, mtime = sysdate()
		WHERE id = ?`

	_, err := s.master.Exec(sql,
		version.Version,
		version.Comment,
		version.InputSchema,
		version.OutputSchema,
		version.SkillType,
		nil2JSONStringForSkillVersion(version.Metadata),
		ifThenElse(version.Active, 1, 0),
		version.Flag,
		version.ID,
	)
	if err != nil {
		log.Errorf("[Store][database] update skill version err: %s", err.Error())
		return err
	}
	return nil
}

// DeleteSkillVersion deletes a skill version (logical delete)
func (s *skillVersionStore) DeleteSkillVersion(id string) error {
	if id == "" {
		return store.NewStatusError(store.EmptyParamsErr, "delete skill version missing id")
	}

	err := RetryTransaction(labelDeleteSkillVersion, func() error {
		return s.deleteSkillVersion(id)
	})
	return store.Error(err)
}

func (s *skillVersionStore) deleteSkillVersion(id string) error {
	sql := `UPDATE skill_version SET flag = 1, mtime = sysdate() WHERE id = ?`
	_, err := s.master.Exec(sql, id)
	if err != nil {
		log.Errorf("[Store][database] logical delete skill version err: %s", err.Error())
		return err
	}
	return nil
}

// GetSkillVersion gets a skill version by ID
func (s *skillVersionStore) GetSkillVersion(id string) (*ai.SkillVersion, error) {
	if id == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get skill version missing id")
	}

	rows, err := s.slave.Query(`SELECT id, skill_id, skill_name, namespace, version, comment,
		input_schema, output_schema, skill_type, metadata, active, flag,
		unix_timestamp(ctime), unix_timestamp(mtime)
		FROM skill_version WHERE id = ?`, id)
	if err != nil {
		log.Errorf("[Store][database] get skill version query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return fetchSkillVersionRow(rows)
}

// GetSkillVersionByVersion gets a skill version by version number
func (s *skillVersionStore) GetSkillVersionByVersion(skillName, namespace string, version uint64) (*ai.SkillVersion, error) {
	if skillName == "" || namespace == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get skill version missing skill name or namespace")
	}

	rows, err := s.slave.Query(`SELECT id, skill_id, skill_name, namespace, version, comment,
		input_schema, output_schema, skill_type, metadata, active, flag,
		unix_timestamp(ctime), unix_timestamp(mtime)
		FROM skill_version WHERE skill_name = ? AND namespace = ? AND version = ? AND flag != 1`,
		skillName, namespace, version)
	if err != nil {
		log.Errorf("[Store][database] get skill version by version query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return fetchSkillVersionRow(rows)
}

// GetSkillVersionsBySkillID gets all versions of a skill
func (s *skillVersionStore) GetSkillVersionsBySkillID(skillID string) ([]*ai.SkillVersion, error) {
	if skillID == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get skill versions missing skill id")
	}

	rows, err := s.slave.Query(`SELECT id, skill_id, skill_name, namespace, version, comment,
		input_schema, output_schema, skill_type, metadata, active, flag,
		unix_timestamp(ctime), unix_timestamp(mtime)
		FROM skill_version WHERE skill_id = ? AND flag != 1 ORDER BY version DESC`, skillID)
	if err != nil {
		log.Errorf("[Store][database] get skill versions by skill id query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var versions []*ai.SkillVersion
	for rows.Next() {
		version, err := fetchSkillVersionRow(rows)
		if err != nil {
			return nil, err
		}
		if version != nil {
			versions = append(versions, version)
		}
	}

	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] get skill versions by skill id rows err: %s", err.Error())
		return nil, err
	}

	return versions, nil
}

// QuerySkillVersions queries skill versions with pagination
func (s *skillVersionStore) QuerySkillVersions(filter map[string]string, offset, limit uint32) (uint32, []*ai.SkillVersion, error) {
	// TODO: Implement query with filter support
	return 0, nil, nil
}

// GetActiveSkillVersion gets the active version of a skill
func (s *skillVersionStore) GetActiveSkillVersion(skillName, namespace string) (*ai.SkillVersion, error) {
	if skillName == "" || namespace == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get active skill version missing skill name or namespace")
	}

	rows, err := s.slave.Query(`SELECT id, skill_id, skill_name, namespace, version, comment,
		input_schema, output_schema, skill_type, metadata, active, flag,
		unix_timestamp(ctime), unix_timestamp(mtime)
		FROM skill_version WHERE skill_name = ? AND namespace = ? AND active = 1 AND flag != 1`,
		skillName, namespace)
	if err != nil {
		log.Errorf("[Store][database] get active skill version query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return fetchSkillVersionRow(rows)
}

// ActiveSkillVersion activates a skill version
func (s *skillVersionStore) ActiveSkillVersion(version *ai.SkillVersion) error {
	if version.ID == "" {
		return store.NewStatusError(store.EmptyParamsErr, "activate skill version missing id")
	}

	err := RetryTransaction("activeSkillVersion", func() error {
		return s.activeSkillVersion(version)
	})
	return store.Error(err)
}

func (s *skillVersionStore) activeSkillVersion(version *ai.SkillVersion) error {
	tx, err := s.master.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Deactivate other versions of the same skill
	_, err = tx.Exec(`UPDATE skill_version SET active = 0 WHERE skill_id = ?`,
		version.SkillID)
	if err != nil {
		return err
	}

	// Activate this version
	_, err = tx.Exec(`UPDATE skill_version SET active = 1, mtime = sysdate() WHERE id = ?`,
		version.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// InactiveSkillVersion deactivates a skill version
func (s *skillVersionStore) InactiveSkillVersion(version *ai.SkillVersion) error {
	if version.ID == "" {
		return store.NewStatusError(store.EmptyParamsErr, "inactive skill version missing id")
	}

	err := RetryTransaction("inactiveSkillVersion", func() error {
		return s.inactiveSkillVersion(version)
	})
	return store.Error(err)
}

func (s *skillVersionStore) inactiveSkillVersion(version *ai.SkillVersion) error {
	_, err := s.master.Exec(`UPDATE skill_version SET active = 0, mtime = sysdate() WHERE id = ?`,
		version.ID)
	if err != nil {
		log.Errorf("[Store][database] inactive skill version err: %s", err.Error())
		return err
	}
	return nil
}

// ===== Skill Subscription Store =====

// skillSubscriptionStore implements store.SkillSubscriptionStore
type skillSubscriptionStore struct {
	master *BaseDB
	slave  *BaseDB
}

// newSkillSubscriptionStore creates a new skill subscription store
func newSkillSubscriptionStore(master, slave *BaseDB) *skillSubscriptionStore {
	return &skillSubscriptionStore{
		master: master,
		slave:  slave,
	}
}

// CreateSkillSubscription creates a new skill subscription
func (s *skillSubscriptionStore) CreateSkillSubscription(sub *ai.SkillSubscription) error {
	if sub.ID == "" {
		sub.ID = uuid.New().String()
	}

	err := RetryTransaction(labelCreateSkillSubscription, func() error {
		return s.createSkillSubscription(sub)
	})
	return store.Error(err)
}

func (s *skillSubscriptionStore) createSkillSubscription(sub *ai.SkillSubscription) error {
	tx, err := s.master.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 插入 skill_subscription 表
	if err := s.insertSkillSubscriptionMain(tx, sub); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *skillSubscriptionStore) insertSkillSubscriptionMain(tx *BaseTx, sub *ai.SkillSubscription) error {
	sql := `INSERT INTO skill_subscription(id, skill_id, skill_name, namespace, client_id, client_host,
		client_type, version, active, ctime, mtime)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, sysdate(), sysdate())`

	_, err := tx.Exec(sql,
		sub.ID,
		sub.SkillID,
		sub.SkillName,
		sub.Namespace,
		sub.ClientID,
		sub.ClientHost,
		sub.ClientType,
		sub.Version,
		ifThenElse(sub.Active, 1, 0),
	)
	if err != nil {
		log.Errorf("[Store][database] insert skill subscription err: %s", err.Error())
		return err
	}
	return nil
}

// UpdateSkillSubscription updates an existing skill subscription
func (s *skillSubscriptionStore) UpdateSkillSubscription(sub *ai.SkillSubscription) error {
	if sub.ID == "" {
		return store.NewStatusError(store.EmptyParamsErr, "update skill subscription missing id")
	}

	err := RetryTransaction(labelUpdateSkillSubscription, func() error {
		return s.updateSkillSubscription(sub)
	})
	return store.Error(err)
}

func (s *skillSubscriptionStore) updateSkillSubscription(sub *ai.SkillSubscription) error {
	sql := `UPDATE skill_subscription SET version = ?, active = ?, mtime = sysdate()
		WHERE id = ?`

	_, err := s.master.Exec(sql,
		sub.Version,
		ifThenElse(sub.Active, 1, 0),
		sub.ID,
	)
	if err != nil {
		log.Errorf("[Store][database] update skill subscription err: %s", err.Error())
		return err
	}
	return nil
}

// DeleteSkillSubscription deletes a skill subscription (logical delete)
func (s *skillSubscriptionStore) DeleteSkillSubscription(id string) error {
	if id == "" {
		return store.NewStatusError(store.EmptyParamsErr, "delete skill subscription missing id")
	}

	err := RetryTransaction(labelDeleteSkillSubscription, func() error {
		return s.deleteSkillSubscription(id)
	})
	return store.Error(err)
}

func (s *skillSubscriptionStore) deleteSkillSubscription(id string) error {
	sql := `UPDATE skill_subscription SET flag = 1, mtime = sysdate() WHERE id = ?`
	_, err := s.master.Exec(sql, id)
	if err != nil {
		log.Errorf("[Store][database] logical delete skill subscription err: %s", err.Error())
		return err
	}
	return nil
}

// GetSkillSubscription gets a skill subscription by ID
func (s *skillSubscriptionStore) GetSkillSubscription(id string) (*ai.SkillSubscription, error) {
	if id == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get skill subscription missing id")
	}

	rows, err := s.slave.Query(`SELECT id, skill_id, skill_name, namespace, client_id, client_host,
		client_type, version, active, flag,
		unix_timestamp(ctime), unix_timestamp(mtime)
		FROM skill_subscription WHERE id = ?`, id)
	if err != nil {
		log.Errorf("[Store][database] get skill subscription query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return fetchSkillSubscriptionRow(rows)
}

// GetSkillSubscriptionByClient gets skill subscriptions by client ID
func (s *skillSubscriptionStore) GetSkillSubscriptionByClient(clientID string) ([]*ai.SkillSubscription, error) {
	if clientID == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get skill subscription by client missing client id")
	}

	rows, err := s.slave.Query(`SELECT id, skill_id, skill_name, namespace, client_id, client_host,
		client_type, version, active, flag,
		unix_timestamp(ctime), unix_timestamp(mtime)
		FROM skill_subscription WHERE client_id = ? AND flag != 1`, clientID)
	if err != nil {
		log.Errorf("[Store][database] get skill subscription by client query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var subs []*ai.SkillSubscription
	for rows.Next() {
		sub, err := fetchSkillSubscriptionRow(rows)
		if err != nil {
			return nil, err
		}
		if sub != nil {
			subs = append(subs, sub)
		}
	}

	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] get skill subscription by client rows err: %s", err.Error())
		return nil, err
	}

	return subs, nil
}

// GetSkillSubscriptionsBySkill gets all subscriptions of a skill
func (s *skillSubscriptionStore) GetSkillSubscriptionsBySkill(skillName, namespace string) ([]*ai.SkillSubscription, error) {
	if skillName == "" || namespace == "" {
		return nil, store.NewStatusError(store.EmptyParamsErr, "get skill subscriptions by skill missing skill name or namespace")
	}

	rows, err := s.slave.Query(`SELECT id, skill_id, skill_name, namespace, client_id, client_host,
		client_type, version, active, flag,
		unix_timestamp(ctime), unix_timestamp(mtime)
		FROM skill_subscription WHERE skill_name = ? AND namespace = ? AND flag != 1`, skillName, namespace)
	if err != nil {
		log.Errorf("[Store][database] get skill subscriptions by skill query err: %s", err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var subs []*ai.SkillSubscription
	for rows.Next() {
		sub, err := fetchSkillSubscriptionRow(rows)
		if err != nil {
			return nil, err
		}
		if sub != nil {
			subs = append(subs, sub)
		}
	}

	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] get skill subscriptions by skill rows err: %s", err.Error())
		return nil, err
	}

	return subs, nil
}

// QuerySkillSubscriptions queries skill subscriptions with pagination
func (s *skillSubscriptionStore) QuerySkillSubscriptions(filter map[string]string, offset, limit uint32) (uint32, []*ai.SkillSubscription, error) {
	// TODO: Implement query with filter support
	return 0, nil, nil
}

// UpdateSubscriptionVersion updates the version of a subscription
func (s *skillSubscriptionStore) UpdateSubscriptionVersion(clientID, skillName, namespace string, version uint64) error {
	if clientID == "" || skillName == "" || namespace == "" {
		return store.NewStatusError(store.EmptyParamsErr, "update subscription version missing params")
	}

	err := RetryTransaction("updateSubscriptionVersion", func() error {
		return s.updateSubscriptionVersion(clientID, skillName, namespace, version)
	})
	return store.Error(err)
}

func (s *skillSubscriptionStore) updateSubscriptionVersion(clientID, skillName, namespace string, version uint64) error {
	_, err := s.master.Exec(`UPDATE skill_subscription SET version = ?, mtime = sysdate()
		WHERE client_id = ? AND skill_name = ? AND namespace = ?`,
		version, clientID, skillName, namespace)
	if err != nil {
		log.Errorf("[Store][database] update subscription version err: %s", err.Error())
		return err
	}
	return nil
}

// DeactiveSubscription deactivates a subscription
func (s *skillSubscriptionStore) DeactiveSubscription(clientID, skillName, namespace string) error {
	if clientID == "" || skillName == "" || namespace == "" {
		return store.NewStatusError(store.EmptyParamsErr, "deactive subscription missing params")
	}

	err := RetryTransaction("deactiveSubscription", func() error {
		return s.deactiveSubscription(clientID, skillName, namespace)
	})
	return store.Error(err)
}

func (s *skillSubscriptionStore) deactiveSubscription(clientID, skillName, namespace string) error {
	_, err := s.master.Exec(`UPDATE skill_subscription SET active = 0, mtime = sysdate()
		WHERE client_id = ? AND skill_name = ? AND namespace = ?`,
		clientID, skillName, namespace)
	if err != nil {
		log.Errorf("[Store][database] deactive subscription err: %s", err.Error())
		return err
	}
	return nil
}

// ===== Helper Functions =====

func fetchSkillVersionRow(rows *sql.Rows) (*ai.SkillVersion, error) {
	var version ai.SkillVersion
	var ctime, mtime int64
	var metadataStr string
	var active int

	err := rows.Scan(
		&version.ID,
		&version.SkillID,
		&version.SkillName,
		&version.Namespace,
		&version.Version,
		&version.Comment,
		&version.InputSchema,
		&version.OutputSchema,
		&version.SkillType,
		&metadataStr,
		&active,
		&version.Flag,
		&ctime,
		&mtime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorf("[Store][database] fetch skill version row scan err: %s", err.Error())
		return nil, err
	}

	version.CTime = time.Unix(ctime, 0)
	version.MTime = time.Unix(mtime, 0)
	version.Metadata = jsonString2MapForSkillVersion(metadataStr)
	version.Active = active == 1

	return &version, nil
}

func fetchSkillSubscriptionRow(rows *sql.Rows) (*ai.SkillSubscription, error) {
	var sub ai.SkillSubscription
	var ctime, mtime int64
	var active int

	err := rows.Scan(
		&sub.ID,
		&sub.SkillID,
		&sub.SkillName,
		&sub.Namespace,
		&sub.ClientID,
		&sub.ClientHost,
		&sub.ClientType,
		&sub.Version,
		&active,
		&ctime,
		&mtime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorf("[Store][database] fetch skill subscription row scan err: %s", err.Error())
		return nil, err
	}

	sub.CTime = time.Unix(ctime, 0)
	sub.MTime = time.Unix(mtime, 0)
	sub.Active = active == 1

	return &sub, nil
}

// ===== Helper Functions =====

// nil2JSONStringForSkillVersion converts nil or empty map to "{}" for MySQL storage
func nil2JSONStringForSkillVersion(m map[string]string) string {
	if m == nil {
		return "{}"
	}
	data, _ := json.Marshal(m)
	return string(data)
}

// jsonString2MapForSkillVersion converts JSON string from MySQL to map[string]string
func jsonString2MapForSkillVersion(str string) map[string]string {
	if str == "" || str == "null" {
		return nil
	}
	m := make(map[string]string)
	_ = json.Unmarshal([]byte(str), &m)
	return m
}

// ifThenElse is a helper function to convert bool to int
func ifThenElse(condition bool, trueVal, falseVal int) int {
	if condition {
		return trueVal
	}
	return falseVal
}
