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

func ensureServiceSubscribeGraphSchema(db *BaseDB) error {
	if ok, err := hasTableColumn(db, "service_subscribe_graph", "ctime"); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`
			alter table service_subscribe_graph
			add column ctime timestamp not null default current_timestamp comment 'Create time'`); err != nil {
			return err
		}
	}

	if ok, err := hasTableColumn(db, "service_subscribe_graph", "mtime"); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`
			alter table service_subscribe_graph
			add column mtime timestamp not null default current_timestamp on update current_timestamp comment 'Last updated time'`); err != nil {
			return err
		}
	}

	if ok, err := hasTableIndex(db, "service_subscribe_graph", "mtime"); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`alter table service_subscribe_graph add key mtime (mtime)`); err != nil {
			return err
		}
	}

	return nil
}

func hasTableColumn(db *BaseDB, tableName string, columnName string) (bool, error) {
	var count uint32
	err := db.QueryRow(`
		select count(*)
		from information_schema.columns
		where table_schema = database()
		  and table_name = ?
		  and column_name = ?`, tableName, columnName).Scan(&count)
	return count > 0, err
}

func hasTableIndex(db *BaseDB, tableName string, indexName string) (bool, error) {
	var count uint32
	err := db.QueryRow(`
		select count(*)
		from information_schema.statistics
		where table_schema = database()
		  and table_name = ?
		  and index_name = ?`, tableName, indexName).Scan(&count)
	return count > 0, err
}
