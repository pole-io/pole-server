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
	"errors"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// namespaceStore 实现了NamespaceStore
type namespaceStore struct {
	master *BaseDB // 大部分操作都用主数据库
	slave  *BaseDB // 缓存相关的读取，请求到slave
}

// AddNamespace 添加命名空间
func (ns *namespaceStore) AddNamespace(data *types.Namespace) error {
	if data.Name == "" {
		return errors.New("store add namespace name is empty")
	}
	return RetryTransaction("addNamespace", func() error {
		return ns.master.processWithTransaction("addNamespace", func(tx *BaseTx) error {
			// 先删除无效数据，再添加新数据
			if err := cleanNamespace(tx, data.Name); err != nil {
				return err
			}

			str := `
			INSERT INTO namespace (name, comment, token, owner, ctime
				, mtime, service_export_to, metadata)
			VALUES (?, ?, ?, ?, sysdate()
				, sysdate(), ?)
			`
			args := []interface{}{data.Name, data.Comment, data.Token, data.Owner,
				utils.MustJson(data.ServiceExportTo), utils.MustJson(data.Metadata)}
			if _, err := tx.Exec(str, args...); err != nil {
				return store.Error(err)
			}

			if err := tx.Commit(); err != nil {
				log.Errorf("[Store][database] batch delete instance commit tx err: %s", err.Error())
				return err
			}

			return nil
		})
	})
}

// UpdateNamespace 更新命名空间，目前只更新owner
func (ns *namespaceStore) UpdateNamespace(data *types.Namespace) error {
	if data.Name == "" {
		return errors.New("store update namespace name is empty")
	}
	return RetryTransaction("updateNamespace", func() error {
		return ns.master.processWithTransaction("updateNamespace", func(tx *BaseTx) error {
			str := "update namespace set owner = ?, comment = ?, service_export_to = ?, mtime = sysdate(), metadata = ? where name = ?"
			args := []interface{}{data.Owner, data.Comment, utils.MustJson(data.ServiceExportTo), utils.MustJson(data.Metadata), data.Name}
			if _, err := tx.Exec(str, args...); err != nil {
				return store.Error(err)
			}

			if err := tx.Commit(); err != nil {
				log.Errorf("[Store][database] batch delete instance commit tx err: %s", err.Error())
				return err
			}

			return nil
		})
	})
}

// GetNamespace 根据名字获取命名空间详情，只返回有效的
func (ns *namespaceStore) GetNamespace(name string) (*types.Namespace, error) {
	namespace, err := ns.getNamespace(name)
	if err != nil {
		return nil, err
	}

	if namespace != nil && !namespace.Valid {
		return nil, nil
	}

	return namespace, nil
}

// GetNamespaces 根据过滤条件查询命名空间及数目
func (ns *namespaceStore) GetNamespaces(filter map[string][]string, offset, limit int) (
	[]*types.Namespace, uint32, error) {
	// 只查询有效数据
	filter["flag"] = []string{"0"}

	num, err := ns.getNamespacesCount(filter)
	if err != nil {
		return nil, 0, err
	}

	out, err := ns.getNamespaces(filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return out, num, nil
}

// GetMoreNamespaces 根据mtime获取命名空间
func (ns *namespaceStore) GetMoreNamespaces(mtime time.Time) ([]*types.Namespace, error) {
	str := genNamespaceSelectSQL() + " where mtime >= FROM_UNIXTIME(?)"
	rows, err := ns.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		log.Errorf("[Store][database] get more namespace query err: %s", err.Error())
		return nil, err
	}

	return namespaceFetchRows(rows)
}

// getNamespacesCount根据相关条件查询对应命名空间数目
func (ns *namespaceStore) getNamespacesCount(filter map[string][]string) (uint32, error) {
	str := `select count(*) from namespace `
	str, args := genNamespaceWhereSQLAndArgs(str, filter, nil, 0, 1)

	var count uint32
	err := ns.master.QueryRow(str, args...).Scan(&count)
	switch {
	case err == sql.ErrNoRows:
		log.Errorf("[Store][database] no row with this namespace filter")
		return count, err
	case err != nil:
		log.Errorf("[Store][database] get namespace count by filter err: %s", err.Error())
		return count, err
	default:
		return count, err
	}
}

// getNamespaces 根据相关条件查询对应命名空间
func (ns *namespaceStore) getNamespaces(filter map[string][]string, offset, limit int) ([]*types.Namespace, error) {
	str := genNamespaceSelectSQL()
	order := &Order{"mtime", "desc"}
	str, args := genNamespaceWhereSQLAndArgs(str, filter, order, offset, limit)

	rows, err := ns.master.Query(str, args...)
	if err != nil {
		log.Errorf("[Store][database] get namespaces by filter query err: %s", err.Error())
		return nil, err
	}

	return namespaceFetchRows(rows)
}

// getNamespace 获取namespace的内部函数，从数据库中拉取数据
func (ns *namespaceStore) getNamespace(name string) (*types.Namespace, error) {
	if name == "" {
		return nil, errors.New("store get namespace name is empty")
	}

	str := genNamespaceSelectSQL() + " where name = ?"
	rows, err := ns.master.Query(str, name)
	if err != nil {
		log.Errorf("[Store][database] get namespace query err: %s", err.Error())
		return nil, err
	}

	out, err := namespaceFetchRows(rows)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, nil
	}
	return out[0], nil
}

// clean真实的数据，只有flag=1的数据才可以清除
func cleanNamespace(tx *BaseTx, name string) error {
	str := "delete from namespace where name = ? and flag = 1"
	// 必须打印日志说明
	log.Infof("[Store][database] clean namespace(%s)", name)
	if _, err := tx.Exec(str, name); err != nil {
		log.Infof("[Store][database] clean namespace(%s) err: %s", name, err.Error())
		return err
	}

	return nil
}

// rlockNamespace rlock namespace
func rlockNamespace(queryRow func(query string, args ...interface{}) *sql.Row, namespace string) (
	string, error) {
	str := "select name from namespace where name = ? and flag != 1 lock in share mode"

	var name string
	err := queryRow(str, namespace).Scan(&name)
	switch {
	case err == sql.ErrNoRows:
		return "", nil
	case err != nil:
		return "", err
	default:
		return name, nil
	}
}

// genNamespaceSelectSQL 生成namespace的查询语句
func genNamespaceSelectSQL() string {
	str := `
	SELECT name, IFNULL(comment, ''), token
	, owner, flag, UNIX_TIMESTAMP(ctime)
	, UNIX_TIMESTAMP(mtime)
	, IFNULL(service_export_to, '{}')
	, IFNULL(metadata, '{}')
FROM namespace
	`
	return str
}

// namespaceFetchRows 取出rows的数据
func namespaceFetchRows(rows *sql.Rows) ([]*types.Namespace, error) {
	if rows == nil {
		return nil, nil
	}
	defer rows.Close()

	var out []*types.Namespace
	var ctime, mtime int64
	var flag int
	var serviceExportTo, metadata string

	for rows.Next() {
		space := &types.Namespace{}
		err := rows.Scan(
			&space.Name, &space.Comment, &space.Token, &space.Owner, &flag, &ctime,
			&mtime, &serviceExportTo, &metadata,
		)
		if err != nil {
			log.Errorf("[Store][database] fetch namespace rows scan err: %s", err.Error())
			return nil, err
		}

		space.CreateTime = time.Unix(ctime, 0)
		space.ModifyTime = time.Unix(mtime, 0)
		space.ServiceExportTo = map[string]struct{}{}
		_ = json.Unmarshal([]byte(serviceExportTo), &space.ServiceExportTo)
		space.Metadata = map[string]string{}
		_ = json.Unmarshal([]byte(metadata), &space.Metadata)
		space.Valid = true
		if flag == 1 {
			space.Valid = false
		}

		out = append(out, space)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("[Store][database] fetch namespace rows next err: %s", err.Error())
		return nil, err
	}

	return out, nil
}
