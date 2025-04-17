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
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
)

var _ store.RouterRuleConfigStore = (*routerRuleStore)(nil)

// routerRuleStore impl
type routerRuleStore struct {
	master *BaseDB
	slave  *BaseDB
}

// CreateRoutingConfig Add a new routing configuration
func (r *routerRuleStore) CreateRoutingConfig(conf *rules.RouterConfig) error {
	if conf.ID == "" || conf.Revision == "" {
		log.Errorf("[Store][boltdb] create routing config  missing id or revision")
		return store.NewStatusError(store.EmptyParamsErr, "missing id or revision")
	}
	if conf.Policy == "" || conf.Config == "" {
		log.Errorf("[Store][boltdb] create routing config  missing params")
		return store.NewStatusError(store.EmptyParamsErr, "missing some params")
	}

	err := RetryTransaction("CreateRoutingConfig", func() error {
		tx, err := r.master.Begin()
		if err != nil {
			return err
		}

		defer func() {
			_ = tx.Rollback()
		}()
		if err := r.createRoutingConfigTx(tx, conf); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			log.Errorf("[Store][database] create routing config (%+v) commit: %s", conf, err.Error())
			return store.Error(err)
		}

		return nil
	})

	return store.Error(err)
}

func (r *routerRuleStore) CreateRoutingConfigTx(tx store.Tx, conf *rules.RouterConfig) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	dbTx := tx.GetDelegateTx().(*BaseTx)
	return r.createRoutingConfigTx(dbTx, conf)
}

func (r *routerRuleStore) createRoutingConfigTx(tx *BaseTx, conf *rules.RouterConfig) error {
	// 删除无效的数据
	if _, err := tx.Exec("DELETE FROM router_rule WHERE id = ? AND flag = 1", conf.ID); err != nil {
		log.Errorf("[Store][database] create routing (%+v) err: %s", conf, err.Error())
		return store.Error(err)
	}

	insertSQL := "INSERT INTO router_rule(id, namespace, name, policy, config, enable, " +
		" priority, revision, description, ctime, mtime, etime) VALUES (?,?,?,?,?,?,?,?,?,sysdate(),sysdate(),%s)"

	var enable int
	if conf.Enable {
		enable = 1
		insertSQL = fmt.Sprintf(insertSQL, "sysdate()")
	} else {
		enable = 0
		insertSQL = fmt.Sprintf(insertSQL, emptyEnableTime)
	}

	log.Debug("[Store][database] create routing ", zap.String("sql", insertSQL))

	if _, err := tx.Exec(insertSQL, conf.ID, conf.Namespace, conf.Name, conf.Policy,
		conf.Config, enable, conf.Priority, conf.Revision, conf.Description); err != nil {
		log.Errorf("[Store][database] create routing (%+v) err: %s", conf, err.Error())
		return store.Error(err)
	}
	return nil
}

// UpdateRoutingConfig Update a routing configuration
func (r *routerRuleStore) UpdateRoutingConfig(conf *rules.RouterConfig) error {

	tx, err := r.master.Begin()
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := r.updateRoutingConfigTx(tx, conf); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Errorf("[Store][database] update routing config (%+v) commit: %s", conf, err.Error())
		return store.Error(err)
	}

	return nil
}

func (r *routerRuleStore) UpdateRoutingConfigTx(tx store.Tx, conf *rules.RouterConfig) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	dbTx := tx.GetDelegateTx().(*BaseTx)
	return r.updateRoutingConfigTx(dbTx, conf)
}

func (r *routerRuleStore) updateRoutingConfigTx(tx *BaseTx, conf *rules.RouterConfig) error {
	if conf.ID == "" || conf.Revision == "" {
		log.Errorf("[Store][database] update routing config  missing id or revision")
		return store.NewStatusError(store.EmptyParamsErr, "missing id or revision")
	}
	if conf.Policy == "" || conf.Config == "" {
		log.Errorf("[Store][boltdb] create routing config  missing params")
		return store.NewStatusError(store.EmptyParamsErr, "missing some params")
	}

	str := "update router_rule set name = ?, policy = ?, config = ?, revision = ?, priority = ?, " +
		" description = ?, mtime = sysdate() where id = ?"
	if _, err := tx.Exec(str, conf.Name, conf.Policy, conf.Config, conf.Revision, conf.Priority, conf.Description,
		conf.ID); err != nil {
		log.Errorf("[Store][database] update routing config (%+v) exec err: %s", conf, err.Error())
		return store.Error(err)
	}
	return nil
}

// EnableRateLimit Enable current limit rules
func (r *routerRuleStore) EnableRouting(conf *rules.RouterConfig) error {
	if conf.ID == "" || conf.Revision == "" {
		return errors.New("[Store][database] enable routing config  missing some params")
	}

	err := RetryTransaction("EnableRouting", func() error {
		var (
			enable   int
			etimeStr string
		)
		if conf.Enable {
			enable = 1
			etimeStr = "sysdate()"
		} else {
			enable = 0
			etimeStr = emptyEnableTime
		}
		str := fmt.Sprintf(
			`update router_rule set enable = ?, revision = ?, mtime = sysdate(), etime=%s where id = ?`, etimeStr)
		if _, err := r.master.Exec(str, enable, conf.Revision, conf.ID); err != nil {
			log.Errorf("[Store][database] update outing config (%+v), sql %s, err: %s", conf, str, err)
			return err
		}

		return nil
	})

	return store.Error(err)
}

// DeleteRoutingConfig Delete a routing configuration
func (r *routerRuleStore) DeleteRoutingConfig(ruleID string) error {

	if ruleID == "" {
		log.Errorf("[Store][database] delete routing config  missing service id")
		return store.NewStatusError(store.EmptyParamsErr, "missing service id")
	}

	str := `update router_rule set flag = 1, mtime = sysdate() where id = ?`
	if _, err := r.master.Exec(str, ruleID); err != nil {
		log.Errorf("[Store][database] delete routing config (%s) err: %s", ruleID, err.Error())
		return store.Error(err)
	}

	return nil
}

// GetRoutingConfigsForCache Pull the incremental routing configuration information through mtime
func (r *routerRuleStore) GetRoutingConfigsForCache(
	mtime time.Time, firstUpdate bool) ([]*rules.RouterConfig, error) {
	str := `select id, name, policy, config, enable, revision, flag, priority, description,
	unix_timestamp(ctime), unix_timestamp(mtime), unix_timestamp(etime)  
	from router_rule where mtime > FROM_UNIXTIME(?) `

	if firstUpdate {
		str += " and flag != 1"
	}
	rows, err := r.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		log.Errorf("[Store][database] query routing configs  with mtime err: %s", err.Error())
		return nil, err
	}
	out, err := fetchRoutingConfigRows(rows)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// GetRoutingConfigWithID Pull the routing configuration according to the rules ID
func (r *routerRuleStore) GetRoutingConfigWithID(ruleID string) (*rules.RouterConfig, error) {

	tx, err := r.master.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback()
	}()
	return r.getRoutingConfigWithIDTx(tx, ruleID)
}

// GetRoutingConfigWithIDTx Pull the routing configuration according to the rules ID
func (r *routerRuleStore) GetRoutingConfigWithIDTx(tx store.Tx, ruleID string) (*rules.RouterConfig, error) {

	if tx == nil {
		return nil, errors.New("transaction is nil")
	}

	dbTx := tx.GetDelegateTx().(*BaseTx)
	return r.getRoutingConfigWithIDTx(dbTx, ruleID)
}

func (r *routerRuleStore) getRoutingConfigWithIDTx(tx *BaseTx, ruleID string) (*rules.RouterConfig, error) {

	str := `select id, name, policy, config, enable, revision, flag, priority, description,
	unix_timestamp(ctime), unix_timestamp(mtime), unix_timestamp(etime)
	from router_rule 
	where id = ? and flag = 0`
	rows, err := tx.Query(str, ruleID)
	if err != nil {
		log.Errorf("[Store][database] query routing  with id(%s) err: %s", ruleID, err.Error())
		return nil, err
	}

	out, err := fetchRoutingConfigRows(rows)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, nil
	}

	return out[0], nil
}

// ActiveRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) ActiveRouterRule(tx store.Tx, name string) error {
	panic("unimplemented")
}

// GetActiveRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) GetActiveRouterRule(tx store.Tx, name string) (*rules.RouterConfig, error) {
	panic("unimplemented")
}

// InactiveRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) InactiveRouterRule(tx store.Tx, name string) error {
	panic("unimplemented")
}

// LockRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) LockRouterRule(tx store.Tx, name string) (*rules.RouterConfig, error) {
	panic("unimplemented")
}

// PublishRouterRule implements store.RouterRuleConfigStore.
func (r *routerRuleStore) PublishRouterRule(tx store.Tx, rule *rules.RouterConfig) error {
	panic("unimplemented")
}

// fetchRoutingConfigRows Read the data of the database and release ROWS
func fetchRoutingConfigRows(rows *sql.Rows) ([]*rules.RouterConfig, error) {
	defer rows.Close()
	var out []*rules.RouterConfig
	for rows.Next() {
		var (
			entry               rules.RouterConfig
			flag, enable        int
			ctime, mtime, etime int64
		)

		err := rows.Scan(&entry.ID, &entry.Name, &entry.Policy, &entry.Config, &enable, &entry.Revision,
			&flag, &entry.Priority, &entry.Description, &ctime, &mtime, &etime)
		if err != nil {
			log.Errorf("[database][store] fetch routing config  scan err: %s", err.Error())
			return nil, err
		}

		entry.CreateTime = time.Unix(ctime, 0)
		entry.ModifyTime = time.Unix(mtime, 0)
		entry.EnableTime = time.Unix(etime, 0)
		entry.Valid = true
		if flag == 1 {
			entry.Valid = false
		}
		entry.Enable = enable == 1

		out = append(out, &entry)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("[database][store] fetch routing config  next err: %s", err.Error())
		return nil, err
	}

	return out, nil
}
