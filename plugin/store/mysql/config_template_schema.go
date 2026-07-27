/*
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 */

package sqldb

func ensureConfigTemplateSchema(db *BaseDB) error {
	for _, table := range configTemplateTables() {
		if err := ensureGovernanceRuleTable(db, table); err != nil {
			return err
		}
	}
	return ensureConfigTemplateDraftColumns(db)
}

type configTemplateColumn struct {
	name string
	ddl  string
}

func configTemplateDraftColumns() []configTemplateColumn {
	return []configTemplateColumn{
		{name: "engine", ddl: `ALTER TABLE config_file_template
			ADD COLUMN engine varchar(32) COLLATE utf8_bin NOT NULL DEFAULT 'pole-mustache'
			COMMENT 'template engine'`},
		{name: "engine_version", ddl: `ALTER TABLE config_file_template
			ADD COLUMN engine_version varchar(32) COLLATE utf8_bin NOT NULL DEFAULT 'v1'
			COMMENT 'template engine version'`},
		{name: "parameter_schema", ddl: `ALTER TABLE config_file_template
			ADD COLUMN parameter_schema longtext COLLATE utf8_bin COMMENT 'parameter schema json'`},
		{name: "revision", ddl: `ALTER TABLE config_file_template
			ADD COLUMN revision varchar(128) COLLATE utf8_bin NOT NULL DEFAULT ''
			COMMENT 'draft revision'`},
	}
}

func ensureConfigTemplateDraftColumns(db *BaseDB) error {
	for _, column := range configTemplateDraftColumns() {
		var count uint32
		if err := db.QueryRow(`
			SELECT COUNT(*)
			FROM information_schema.columns
			WHERE table_schema = database()
			  AND table_name = 'config_file_template'
			  AND column_name = ?`, column.name).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if _, err := db.Exec(column.ddl); err != nil {
			return err
		}
	}
	return nil
}

func configTemplateTables() []governanceRuleTable {
	return []governanceRuleTable{
		{
			name: "config_template_release",
			ddl: `CREATE TABLE IF NOT EXISTS config_template_release (
	id varchar(128) NOT NULL COMMENT 'template release id',
	template_id bigint unsigned NOT NULL COMMENT 'config template id',
	name varchar(128) COLLATE utf8_bin NOT NULL COMMENT 'release name',
	content longtext COLLATE utf8_bin NOT NULL COMMENT 'immutable template content',
	format varchar(16) COLLATE utf8_bin NOT NULL DEFAULT 'text' COMMENT 'rendered config format',
	parameter_schema longtext COLLATE utf8_bin COMMENT 'parameter schema json',
	engine varchar(32) COLLATE utf8_bin NOT NULL DEFAULT 'pole-mustache' COMMENT 'template engine',
	engine_version varchar(32) COLLATE utf8_bin NOT NULL DEFAULT 'v1' COMMENT 'template engine version',
	version bigint unsigned NOT NULL COMMENT 'template release version',
	content_sha256 varchar(64) COLLATE utf8_bin NOT NULL COMMENT 'template source SHA-256',
	comment varchar(512) COLLATE utf8_bin DEFAULT NULL COMMENT 'release description',
	create_by varchar(32) COLLATE utf8_bin DEFAULT NULL COMMENT 'creator',
	ctime timestamp NOT NULL DEFAULT current_timestamp COMMENT 'create time',
	PRIMARY KEY (id),
	UNIQUE KEY uk_template_version (template_id, version),
	KEY idx_template_ctime (template_id, ctime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='immutable config template releases'`,
		},
		{
			name: "namespace_template_values",
			ddl: `CREATE TABLE IF NOT EXISTS namespace_template_values (
	id varchar(128) NOT NULL COMMENT 'values aggregate id',
	namespace varchar(64) COLLATE utf8_bin NOT NULL COMMENT 'namespace',
	template_id bigint unsigned NOT NULL COMMENT 'config template id',
	values_content longtext COLLATE utf8_bin NOT NULL COMMENT 'draft values json',
	revision varchar(128) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT 'draft revision',
	create_by varchar(32) COLLATE utf8_bin DEFAULT NULL COMMENT 'creator',
	modify_by varchar(32) COLLATE utf8_bin DEFAULT NULL COMMENT 'modifier',
	ctime timestamp NOT NULL DEFAULT current_timestamp COMMENT 'create time',
	mtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp COMMENT 'modify time',
	PRIMARY KEY (id),
	UNIQUE KEY uk_namespace_template (namespace, template_id),
	KEY idx_values_mtime (mtime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='Namespace and template scoped Value drafts'`,
		},
		{
			name: "namespace_template_value_release",
			ddl: `CREATE TABLE IF NOT EXISTS namespace_template_value_release (
	id varchar(128) NOT NULL COMMENT 'value release id',
	values_id varchar(128) NOT NULL COMMENT 'values aggregate id',
	namespace varchar(64) COLLATE utf8_bin NOT NULL COMMENT 'namespace',
	template_id bigint unsigned NOT NULL COMMENT 'config template id',
	template_release_id varchar(128) NOT NULL COMMENT 'validated template release id',
	values_content longtext COLLATE utf8_bin NOT NULL COMMENT 'immutable values json',
	release_type varchar(16) COLLATE utf8_bin NOT NULL COMMENT 'normal or gray',
	beta_labels text COLLATE utf8_bin COMMENT 'gray client labels json',
	priority int NOT NULL DEFAULT 0 COMMENT 'gray match priority',
	active tinyint(4) NOT NULL DEFAULT 0 COMMENT 'active flag',
	version bigint unsigned NOT NULL COMMENT 'value release version',
	revision varchar(128) COLLATE utf8_bin NOT NULL COMMENT 'immutable Value release revision',
	comment varchar(512) COLLATE utf8_bin DEFAULT NULL COMMENT 'release description',
	create_by varchar(32) COLLATE utf8_bin DEFAULT NULL COMMENT 'creator',
	modify_by varchar(32) COLLATE utf8_bin DEFAULT NULL COMMENT 'modifier',
	ctime timestamp NOT NULL DEFAULT current_timestamp COMMENT 'create time',
	mtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp COMMENT 'modify time',
	PRIMARY KEY (id),
	UNIQUE KEY uk_values_version (values_id, version),
	KEY idx_value_release_match (namespace, template_id, active, release_type, priority),
	KEY idx_value_release_mtime (mtime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='immutable Namespace template Value releases'`,
		},
		{
			name: "config_file_template_binding_release",
			ddl: `CREATE TABLE IF NOT EXISTS config_file_template_binding_release (
	binding_release_id varchar(128) NOT NULL COMMENT 'binding release id',
	namespace varchar(64) COLLATE utf8_bin NOT NULL COMMENT 'config namespace',
	config_group varchar(128) COLLATE utf8_bin NOT NULL COMMENT 'config group',
	file_name varchar(128) COLLATE utf8_bin NOT NULL COMMENT 'config file name',
	template_id bigint unsigned NOT NULL COMMENT 'config template id',
	template_release_id varchar(128) NOT NULL COMMENT 'pinned template release id',
	active tinyint(4) NOT NULL DEFAULT 0 COMMENT 'active flag',
	version bigint unsigned NOT NULL COMMENT 'binding release version',
	comment varchar(512) COLLATE utf8_bin DEFAULT NULL COMMENT 'binding description',
	create_by varchar(32) COLLATE utf8_bin DEFAULT NULL COMMENT 'creator',
	modify_by varchar(32) COLLATE utf8_bin DEFAULT NULL COMMENT 'modifier',
	ctime timestamp NOT NULL DEFAULT current_timestamp COMMENT 'create time',
	mtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp COMMENT 'modify time',
	PRIMARY KEY (binding_release_id),
	UNIQUE KEY uk_config_binding_version (namespace, config_group, file_name, version),
	KEY idx_config_binding_active (namespace, config_group, file_name, active),
	KEY idx_config_binding_template (template_id, template_release_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='explicit config file template binding releases'`,
		},
	}
}
