/*
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 */

package sqldb

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/store"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

type configTemplateReleaseStore struct {
	master *BaseDB
	slave  *BaseDB
}

func (s *configTemplateReleaseStore) CreateConfigTemplateRelease(release *conftypes.ConfigTemplateRelease) error {
	_, err := s.master.Exec(`INSERT INTO config_template_release (
		id, template_id, name, content, format, parameter_schema, engine, engine_version,
		version, content_sha256, comment, create_by, ctime
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, sysdate())`,
		release.ID, release.TemplateID, release.Name, release.Content, release.Format,
		release.ParameterSchema, release.Engine, release.EngineVersion, release.Version,
		release.ContentSHA256, release.Comment, release.CreateBy)
	return store.Error(err)
}

func (s *configTemplateReleaseStore) GetConfigTemplateRelease(id string) (*conftypes.ConfigTemplateRelease, error) {
	row := s.slave.QueryRow(`SELECT id, template_id, name, content, format,
		IFNULL(parameter_schema, ''), engine, engine_version, version, IFNULL(content_sha256, ''),
		IFNULL(comment, ''),
		IFNULL(create_by, ''), UNIX_TIMESTAMP(ctime)
		FROM config_template_release WHERE id = ?`, id)
	release, err := scanConfigTemplateRelease(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, store.Error(err)
	}
	return release, nil
}

func (s *configTemplateReleaseStore) ListConfigTemplateReleases(
	templateID uint64) ([]*conftypes.ConfigTemplateRelease, error) {
	rows, err := s.slave.Query(`SELECT id, template_id, name, content, format,
		IFNULL(parameter_schema, ''), engine, engine_version, version, IFNULL(content_sha256, ''),
		IFNULL(comment, ''),
		IFNULL(create_by, ''), UNIX_TIMESTAMP(ctime)
		FROM config_template_release WHERE template_id = ? ORDER BY version DESC`, templateID)
	if err != nil {
		return nil, store.Error(err)
	}
	defer rows.Close()

	var releases []*conftypes.ConfigTemplateRelease
	for rows.Next() {
		release, scanErr := scanConfigTemplateRelease(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		releases = append(releases, release)
	}
	if err := rows.Err(); err != nil {
		return nil, store.Error(err)
	}
	return releases, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanConfigTemplateRelease(row rowScanner) (*conftypes.ConfigTemplateRelease, error) {
	release := &conftypes.ConfigTemplateRelease{}
	var ctime int64
	if err := row.Scan(&release.ID, &release.TemplateID, &release.Name, &release.Content,
		&release.Format, &release.ParameterSchema, &release.Engine, &release.EngineVersion,
		&release.Version, &release.ContentSHA256, &release.Comment, &release.CreateBy, &ctime); err != nil {
		return nil, err
	}
	release.CreateTime = time.Unix(ctime, 0)
	return release, nil
}

type namespaceTemplateValuesStore struct {
	master *BaseDB
	slave  *BaseDB
}

func (s *namespaceTemplateValuesStore) SaveNamespaceTemplateValues(
	values *conftypes.NamespaceTemplateValues) error {
	_, err := s.master.Exec(`INSERT INTO namespace_template_values (
		id, namespace, template_id, values_content, revision, create_by, modify_by, ctime, mtime
	) VALUES (?, ?, ?, ?, ?, ?, ?, sysdate(), sysdate())
	ON DUPLICATE KEY UPDATE values_content = VALUES(values_content), revision = VALUES(revision),
		modify_by = VALUES(modify_by), mtime = sysdate()`,
		values.ID, values.Namespace, values.TemplateID, values.Values, values.Revision,
		values.CreateBy, values.ModifyBy)
	return store.Error(err)
}

func (s *namespaceTemplateValuesStore) GetNamespaceTemplateValues(
	namespace string, templateID uint64) (*conftypes.NamespaceTemplateValues, error) {
	row := s.slave.QueryRow(`SELECT id, namespace, template_id, values_content, revision,
		IFNULL(create_by, ''), IFNULL(modify_by, ''), UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime)
		FROM namespace_template_values WHERE namespace = ? AND template_id = ?`,
		namespace, templateID)
	values, err := scanNamespaceTemplateValues(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, store.Error(err)
	}
	return values, nil
}

func scanNamespaceTemplateValues(row rowScanner) (*conftypes.NamespaceTemplateValues, error) {
	values := &conftypes.NamespaceTemplateValues{}
	var ctime, mtime int64
	if err := row.Scan(&values.ID, &values.Namespace, &values.TemplateID, &values.Values,
		&values.Revision, &values.CreateBy, &values.ModifyBy, &ctime, &mtime); err != nil {
		return nil, err
	}
	values.CreateTime = time.Unix(ctime, 0)
	values.ModifyTime = time.Unix(mtime, 0)
	return values, nil
}

func (s *namespaceTemplateValuesStore) CreateNamespaceTemplateValueRelease(
	release *conftypes.NamespaceTemplateValueRelease) error {
	labels, err := json.Marshal(release.BetaLabels)
	if err != nil {
		return err
	}

	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()

	if release.Active && release.ReleaseType == conftypes.TemplateValueReleaseTypeNormal {
		if _, err := tx.Exec(`UPDATE namespace_template_value_release
			SET active = 0, mtime = sysdate()
			WHERE namespace = ? AND template_id = ? AND release_type = ? AND active = 1`,
			release.Namespace, release.TemplateID, conftypes.TemplateValueReleaseTypeNormal); err != nil {
			return store.Error(err)
		}
	}
	if _, err := tx.Exec(`INSERT INTO namespace_template_value_release (
		id, values_id, namespace, template_id, template_release_id, values_content,
		release_type, beta_labels, priority, active, version, revision, comment,
		create_by, modify_by, ctime, mtime
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, sysdate(), sysdate())`,
		release.ID, release.ValuesID, release.Namespace, release.TemplateID,
		release.TemplateReleaseID, release.Values, release.ReleaseType, string(labels),
		release.Priority, release.Active, release.Version, release.Revision, release.Comment,
		release.CreateBy, release.ModifyBy); err != nil {
		return store.Error(err)
	}
	return store.Error(tx.Commit())
}

func (s *namespaceTemplateValuesStore) GetNamespaceTemplateValueRelease(
	id string) (*conftypes.NamespaceTemplateValueRelease, error) {
	row := s.slave.QueryRow(namespaceTemplateValueReleaseSelect+` WHERE id = ?`, id)
	release, err := scanNamespaceTemplateValueRelease(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, store.Error(err)
	}
	return release, nil
}

func (s *namespaceTemplateValuesStore) ListNamespaceTemplateValueReleases(
	namespace string, templateID uint64) ([]*conftypes.NamespaceTemplateValueRelease, error) {
	rows, err := s.slave.Query(namespaceTemplateValueReleaseSelect+
		` WHERE namespace = ? AND template_id = ?
		ORDER BY priority ASC, version DESC, mtime DESC`, namespace, templateID)
	if err != nil {
		return nil, store.Error(err)
	}
	defer rows.Close()

	var releases []*conftypes.NamespaceTemplateValueRelease
	for rows.Next() {
		release, scanErr := scanNamespaceTemplateValueRelease(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		releases = append(releases, release)
	}
	if err := rows.Err(); err != nil {
		return nil, store.Error(err)
	}
	return releases, nil
}

func (s *namespaceTemplateValuesStore) SetNamespaceTemplateValueReleaseActive(id string, active bool) error {
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()

	release, err := scanNamespaceTemplateValueRelease(tx.QueryRow(
		namespaceTemplateValueReleaseSelect+` WHERE id = ? FOR UPDATE`, id))
	if err != nil {
		return store.Error(err)
	}
	if active && release.ReleaseType == conftypes.TemplateValueReleaseTypeNormal {
		if _, err := tx.Exec(`UPDATE namespace_template_value_release
			SET active = 0, mtime = sysdate()
			WHERE namespace = ? AND template_id = ? AND release_type = ? AND active = 1`,
			release.Namespace, release.TemplateID, conftypes.TemplateValueReleaseTypeNormal); err != nil {
			return store.Error(err)
		}
	}
	if _, err := tx.Exec(`UPDATE namespace_template_value_release
		SET active = ?, mtime = sysdate() WHERE id = ?`, active, id); err != nil {
		return store.Error(err)
	}
	return store.Error(tx.Commit())
}

const namespaceTemplateValueReleaseSelect = `SELECT id, values_id, namespace, template_id,
	template_release_id, values_content, release_type, IFNULL(beta_labels, '[]'), priority,
	active, version, IFNULL(revision, ''), IFNULL(comment, ''), IFNULL(create_by, ''), IFNULL(modify_by, ''),
	UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime)
	FROM namespace_template_value_release`

func scanNamespaceTemplateValueRelease(row rowScanner) (*conftypes.NamespaceTemplateValueRelease, error) {
	release := &conftypes.NamespaceTemplateValueRelease{}
	var labels string
	var ctime, mtime int64
	if err := row.Scan(&release.ID, &release.ValuesID, &release.Namespace, &release.TemplateID,
		&release.TemplateReleaseID, &release.Values, &release.ReleaseType, &labels,
		&release.Priority, &release.Active, &release.Version, &release.Revision, &release.Comment,
		&release.CreateBy, &release.ModifyBy, &ctime, &mtime); err != nil {
		return nil, err
	}
	release.BetaLabels = []*apimodel.ClientLabel{}
	if err := json.Unmarshal([]byte(labels), &release.BetaLabels); err != nil {
		return nil, err
	}
	release.CreateTime = time.Unix(ctime, 0)
	release.ModifyTime = time.Unix(mtime, 0)
	return release, nil
}

type configTemplateBindingStore struct {
	master *BaseDB
	slave  *BaseDB
}

func (s *configTemplateBindingStore) CreateConfigTemplateBinding(
	binding *conftypes.ConfigTemplateBinding) error {
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.createConfigTemplateBindingTx(tx, binding); err != nil {
		return err
	}
	return store.Error(tx.Commit())
}

func (s *configTemplateBindingStore) CreateConfigTemplateBindingTx(
	tx store.Tx, binding *conftypes.ConfigTemplateBinding) error {
	if tx == nil {
		return ErrTxIsNil
	}
	return s.createConfigTemplateBindingTx(tx.GetDelegateTx().(*BaseTx), binding)
}

func (s *configTemplateBindingStore) createConfigTemplateBindingTx(
	tx interface {
		Exec(query string, args ...any) (sql.Result, error)
	}, binding *conftypes.ConfigTemplateBinding) error {
	if binding.Active {
		if _, err := tx.Exec(`UPDATE config_file_template_binding_release
			SET active = 0, mtime = sysdate()
			WHERE namespace = ? AND config_group = ? AND file_name = ? AND active = 1`,
			binding.Namespace, binding.Group, binding.FileName); err != nil {
			return store.Error(err)
		}
	}
	if _, err := tx.Exec(`INSERT INTO config_file_template_binding_release (
		binding_release_id, namespace, config_group, file_name, template_id,
		template_release_id, active, version, comment, create_by, modify_by, ctime, mtime
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, sysdate(), sysdate())`,
		binding.BindingReleaseID, binding.Namespace, binding.Group, binding.FileName,
		binding.TemplateID, binding.TemplateReleaseID, binding.Active, binding.Version,
		binding.Comment, binding.CreateBy, binding.ModifyBy); err != nil {
		return store.Error(err)
	}
	return nil
}

func (s *configTemplateBindingStore) GetActiveConfigTemplateBinding(
	file *conftypes.ConfigFileKey) (*conftypes.ConfigTemplateBinding, error) {
	row := s.slave.QueryRow(configTemplateBindingSelect+
		` WHERE namespace = ? AND config_group = ? AND file_name = ? AND active = 1
		ORDER BY version DESC LIMIT 1`, file.Namespace, file.Group, file.Name)
	binding, err := scanConfigTemplateBinding(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, store.Error(err)
	}
	return binding, nil
}

func (s *configTemplateBindingStore) ListConfigTemplateBindings(
	file *conftypes.ConfigFileKey) ([]*conftypes.ConfigTemplateBinding, error) {
	rows, err := s.slave.Query(configTemplateBindingSelect+
		` WHERE namespace = ? AND config_group = ? AND file_name = ? ORDER BY version DESC`,
		file.Namespace, file.Group, file.Name)
	if err != nil {
		return nil, store.Error(err)
	}
	defer rows.Close()

	var bindings []*conftypes.ConfigTemplateBinding
	for rows.Next() {
		binding, scanErr := scanConfigTemplateBinding(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		bindings = append(bindings, binding)
	}
	if err := rows.Err(); err != nil {
		return nil, store.Error(err)
	}
	return bindings, nil
}

func (s *configTemplateBindingStore) SetConfigTemplateBindingActive(
	bindingReleaseID string, active bool) error {
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.setConfigTemplateBindingActiveTx(tx, bindingReleaseID, active); err != nil {
		return err
	}
	return store.Error(tx.Commit())
}

func (s *configTemplateBindingStore) SetConfigTemplateBindingActiveTx(
	tx store.Tx, bindingReleaseID string, active bool) error {
	if tx == nil {
		return ErrTxIsNil
	}
	return s.setConfigTemplateBindingActiveTx(tx.GetDelegateTx().(*BaseTx), bindingReleaseID, active)
}

func (s *configTemplateBindingStore) setConfigTemplateBindingActiveTx(
	tx interface {
		Exec(query string, args ...any) (sql.Result, error)
		QueryRow(query string, args ...any) *sql.Row
	}, bindingReleaseID string, active bool) error {
	binding, err := scanConfigTemplateBinding(tx.QueryRow(
		configTemplateBindingSelect+` WHERE binding_release_id = ? FOR UPDATE`, bindingReleaseID))
	if err != nil {
		return store.Error(err)
	}
	if active {
		if _, err := tx.Exec(`UPDATE config_file_template_binding_release
			SET active = 0, mtime = sysdate()
			WHERE namespace = ? AND config_group = ? AND file_name = ? AND active = 1`,
			binding.Namespace, binding.Group, binding.FileName); err != nil {
			return store.Error(err)
		}
	}
	if _, err := tx.Exec(`UPDATE config_file_template_binding_release
		SET active = ?, mtime = sysdate() WHERE binding_release_id = ?`,
		active, bindingReleaseID); err != nil {
		return store.Error(err)
	}
	return nil
}

const configTemplateBindingSelect = `SELECT binding_release_id, namespace, config_group,
	file_name, template_id, template_release_id, active, version, IFNULL(comment, ''),
	IFNULL(create_by, ''), IFNULL(modify_by, ''), UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime)
	FROM config_file_template_binding_release`

func scanConfigTemplateBinding(row rowScanner) (*conftypes.ConfigTemplateBinding, error) {
	binding := &conftypes.ConfigTemplateBinding{}
	var ctime, mtime int64
	if err := row.Scan(&binding.BindingReleaseID, &binding.Namespace, &binding.Group,
		&binding.FileName, &binding.TemplateID, &binding.TemplateReleaseID, &binding.Active,
		&binding.Version, &binding.Comment, &binding.CreateBy, &binding.ModifyBy,
		&ctime, &mtime); err != nil {
		return nil, err
	}
	binding.CreateTime = time.Unix(ctime, 0)
	binding.ModifyTime = time.Unix(mtime, 0)
	return binding, nil
}
