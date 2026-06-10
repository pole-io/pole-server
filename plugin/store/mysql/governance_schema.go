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

type governanceRuleTable struct {
	name string
	ddl  string
}

func ensureGovernanceRuleSchema(db *BaseDB) error {
	for _, table := range governanceRuleTables() {
		if err := ensureGovernanceRuleTable(db, table); err != nil {
			return err
		}
	}
	return nil
}

func governanceRuleTables() []governanceRuleTable {
	return []governanceRuleTable{
		{
			name: "governance_rule",
			ddl: `CREATE TABLE IF NOT EXISTS governance_rule (
	id varchar(128) NOT NULL COMMENT 'rule id',
	rule_type varchar(64) NOT NULL COMMENT 'governance rule type',
	namespace varchar(64) NOT NULL DEFAULT '' COMMENT 'namespace',
	name varchar(128) NOT NULL DEFAULT '' COMMENT 'rule name',
	service_id varchar(128) NOT NULL DEFAULT '' COMMENT 'service id',
	service varchar(128) NOT NULL DEFAULT '' COMMENT 'service name',
	method varchar(512) NOT NULL DEFAULT '' COMMENT 'method',
	priority int NOT NULL DEFAULT 0 COMMENT 'rule priority',
	enable tinyint(4) NOT NULL DEFAULT 1 COMMENT 'enable flag',
	disable tinyint(4) NOT NULL DEFAULT 0 COMMENT 'disable flag',
	level varchar(32) NOT NULL DEFAULT '' COMMENT 'breaker level',
	src_service varchar(128) NOT NULL DEFAULT '' COMMENT 'source service',
	src_namespace varchar(64) NOT NULL DEFAULT '' COMMENT 'source namespace',
	dst_service varchar(128) NOT NULL DEFAULT '' COMMENT 'destination service',
	dst_namespace varchar(64) NOT NULL DEFAULT '' COMMENT 'destination namespace',
	dst_method varchar(512) NOT NULL DEFAULT '' COMMENT 'destination method',
	labels text COMMENT 'labels json',
	policy varchar(64) NOT NULL DEFAULT '' COMMENT 'route policy',
	config text COMMENT 'config json',
	rule mediumtext COMMENT 'rule json',
	revision varchar(128) NOT NULL DEFAULT '' COMMENT 'rule revision',
	description varchar(1024) NOT NULL DEFAULT '' COMMENT 'description',
	metadata text COMMENT 'metadata json',
	flag tinyint(4) NOT NULL DEFAULT 0 COMMENT 'delete flag',
	ctime timestamp NOT NULL DEFAULT current_timestamp COMMENT 'create time',
	etime timestamp NOT NULL DEFAULT '1980-01-01 00:00:01' COMMENT 'enable time',
	mtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp COMMENT 'modify time',
	PRIMARY KEY (id),
	KEY idx_rule_type_name (rule_type, name),
	KEY idx_rule_type_namespace_name (rule_type, namespace, name),
	KEY idx_rule_type_mtime (rule_type, mtime),
	KEY idx_rule_type_service (rule_type, namespace, service)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='governance rule unified table'`,
		},
		{
			name: "governance_rule_release",
			ddl: `CREATE TABLE IF NOT EXISTS governance_rule_release (
	id varchar(128) NOT NULL COMMENT 'release id',
	rule_type varchar(64) NOT NULL COMMENT 'governance rule type',
	name varchar(128) NOT NULL DEFAULT '' COMMENT 'release name',
	rule_id varchar(128) NOT NULL DEFAULT '' COMMENT 'rule id',
	rule_name varchar(128) NOT NULL DEFAULT '' COMMENT 'rule name',
	namespace varchar(64) NOT NULL DEFAULT '' COMMENT 'namespace',
	service varchar(128) NOT NULL DEFAULT '' COMMENT 'service name',
	rule mediumtext COMMENT 'released rule json',
	version bigint unsigned NOT NULL DEFAULT 0 COMMENT 'release version',
	active tinyint(4) NOT NULL DEFAULT 0 COMMENT 'active flag',
	description varchar(1024) NOT NULL DEFAULT '' COMMENT 'description',
	release_type varchar(64) NOT NULL DEFAULT '' COMMENT 'release type',
	client_labels text COMMENT 'gray client labels json',
	metadata text COMMENT 'metadata json',
	flag tinyint(4) NOT NULL DEFAULT 0 COMMENT 'delete flag',
	ctime timestamp NOT NULL DEFAULT current_timestamp COMMENT 'create time',
	mtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp COMMENT 'modify time',
	PRIMARY KEY (id),
	KEY idx_rule_type_rule_id (rule_type, rule_id),
	KEY idx_rule_type_rule_name (rule_type, rule_name),
	KEY idx_rule_type_release (rule_type, rule_id, name, release_type),
	KEY idx_rule_type_active (rule_type, active, release_type),
	KEY idx_rule_type_mtime (rule_type, mtime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='governance rule release unified table'`,
		},
	}
}

func ensureGovernanceRuleTable(db *BaseDB, table governanceRuleTable) error {
	ok, err := hasTable(db, table.name)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	_, err = db.Exec(table.ddl)
	return err
}

func hasTable(db *BaseDB, tableName string) (bool, error) {
	var count uint32
	err := db.QueryRow(`
		select count(*)
		from information_schema.tables
		where table_schema = database()
		  and table_name = ?`, tableName).Scan(&count)
	return count > 0, err
}
