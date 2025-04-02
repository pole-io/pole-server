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

package boltdb

import (
	"errors"
	"time"

	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
)

var _ store.RouterRuleConfigStore = (*routerRuleStore)(nil)

var (
	// ErrMultipleRoutingFound 多个路由配置
	ErrMultipleRoutingFound = errors.New("multiple routing  found")
)

const (
	tblNameRouting = "routing_config_"

	routingFieldID          = "ID"
	routingFieldName        = "Name"
	routingFieldNamespace   = "Namespace"
	routingFieldPolicy      = "Policy"
	routingFieldConfig      = "Config"
	routingFieldEnable      = "Enable"
	routingFieldRevision    = "Revision"
	routingFieldCreateTime  = "CreateTime"
	routingFieldModifyTime  = "ModifyTime"
	routingFieldEnableTime  = "EnableTime"
	routingFieldValid       = "Valid"
	routingFieldPriority    = "Priority"
	routingFieldDescription = "Description"
)

type routerRuleStore struct {
	handler BoltHandler
}

// CreateRoutingConfig 新增一个路由配置
func (r *routerRuleStore) CreateRoutingConfig(conf *rules.RouterConfig) error {
	if conf.ID == "" || conf.Revision == "" {
		log.Errorf("[Store][boltdb] create routing config  missing id or revision")
		return store.NewStatusError(store.EmptyParamsErr, "missing id or revision")
	}
	if conf.Policy == "" || conf.Config == "" {
		log.Errorf("[Store][boltdb] create routing config  missing params")
		return store.NewStatusError(store.EmptyParamsErr, "missing some params")
	}

	return r.handler.Execute(true, func(tx *bolt.Tx) error {
		return r.createRoutingConfig(tx, conf)
	})
}

// cleanRoutingConfig 从数据库彻底清理路由配置
func (r *routerRuleStore) cleanRoutingConfig(tx *bolt.Tx, ruleID string) error {
	err := deleteValues(tx, tblNameRouting, []string{ruleID})
	if err != nil {
		log.Errorf("[Store][boltdb] delete invalid route config  error, %v", err)
		return err
	}
	return nil
}

func (r *routerRuleStore) CreateRoutingConfigTx(tx store.Tx, conf *rules.RouterConfig) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}

	dbTx := tx.GetDelegateTx().(*bolt.Tx)
	return r.createRoutingConfig(dbTx, conf)
}

func (r *routerRuleStore) createRoutingConfig(tx *bolt.Tx, conf *rules.RouterConfig) error {
	if err := r.cleanRoutingConfig(tx, conf.ID); err != nil {
		return err
	}

	currTime := time.Now()
	conf.CreateTime = currTime
	conf.ModifyTime = currTime
	conf.EnableTime = time.Time{}
	conf.Valid = true

	if conf.Enable {
		conf.EnableTime = time.Now()
	} else {
		conf.EnableTime = time.Time{}
	}

	err := saveValue(tx, tblNameRouting, conf.ID, conf)
	if err != nil {
		log.Errorf("[Store][boltdb] add routing config  to kv error, %v", err)
		return err
	}
	return nil
}

// UpdateRoutingConfig 更新一个路由配置
func (r *routerRuleStore) UpdateRoutingConfig(conf *rules.RouterConfig) error {
	if conf.ID == "" || conf.Revision == "" {
		log.Errorf("[Store][boltdb] update routing config  missing id or revision")
		return store.NewStatusError(store.EmptyParamsErr, "missing id or revision")
	}
	if conf.Policy == "" || conf.Config == "" {
		log.Errorf("[Store][boltdb] create routing config  missing params")
		return store.NewStatusError(store.EmptyParamsErr, "missing some params")
	}

	return r.handler.Execute(true, func(tx *bolt.Tx) error {
		return r.updateRoutingConfigTx(tx, conf)
	})
}

func (r *routerRuleStore) UpdateRoutingConfigTx(tx store.Tx, conf *rules.RouterConfig) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	dbTx := tx.GetDelegateTx().(*bolt.Tx)
	return r.updateRoutingConfigTx(dbTx, conf)
}

func (r *routerRuleStore) updateRoutingConfigTx(tx *bolt.Tx, conf *rules.RouterConfig) error {
	properties := make(map[string]interface{})
	properties[routingFieldEnable] = conf.Enable
	properties[routingFieldName] = conf.Name
	properties[routingFieldPolicy] = conf.Policy
	properties[routingFieldConfig] = conf.Config
	properties[routingFieldPriority] = conf.Priority
	properties[routingFieldRevision] = conf.Revision
	properties[routingFieldDescription] = conf.Description
	properties[routingFieldModifyTime] = time.Now()

	err := updateValue(tx, tblNameRouting, conf.ID, properties)
	if err != nil {
		log.Errorf("[Store][boltdb] update route config  to kv error, %v", err)
		return err
	}
	return nil
}

// EnableRouting
func (r *routerRuleStore) EnableRouting(conf *rules.RouterConfig) error {
	if conf.ID == "" || conf.Revision == "" {
		return errors.New("[Store][database] enable routing config  missing some params")
	}

	if conf.Enable {
		conf.EnableTime = time.Now()
	} else {
		conf.EnableTime = time.Time{}
	}

	properties := make(map[string]interface{})
	properties[routingFieldEnable] = conf.Enable
	properties[routingFieldEnableTime] = conf.EnableTime
	properties[routingFieldRevision] = conf.Revision
	properties[routingFieldModifyTime] = time.Now()

	err := r.handler.UpdateValue(tblNameRouting, conf.ID, properties)
	if err != nil {
		log.Errorf("[Store][boltdb] enable route config  to kv error, %v", err)
		return err
	}
	return nil
}

// DeleteRoutingConfig 删除一个路由配置
func (r *routerRuleStore) DeleteRoutingConfig(ruleID string) error {
	if ruleID == "" {
		log.Errorf("[Store][boltdb] update routing config  missing id")
		return store.NewStatusError(store.EmptyParamsErr, "missing id")
	}
	properties := make(map[string]interface{})
	properties[routingFieldValid] = false
	properties[routingFieldModifyTime] = time.Now()

	err := r.handler.UpdateValue(tblNameRouting, ruleID, properties)
	if err != nil {
		log.Errorf("[Store][boltdb] update route config  to kv error, %v", err)
		return err
	}
	return nil
}

// GetRoutingConfigsForCache 通过mtime拉取增量的路由配置信息
// 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
func (r *routerRuleStore) GetRoutingConfigsForCache(mtime time.Time, firstUpdate bool) ([]*rules.RouterConfig, error) {
	if firstUpdate {
		mtime = time.Time{}
	}

	fields := []string{routingFieldModifyTime}

	routes, err := r.handler.LoadValuesByFilter(tblNameRouting, fields, &rules.RouterConfig{},
		func(m map[string]interface{}) bool {
			rMtime, ok := m[routingFieldModifyTime]
			if !ok {
				return false
			}
			routeMtime := rMtime.(time.Time)
			return !routeMtime.Before(mtime)
		})
	if err != nil {
		log.Errorf("[Store][boltdb] load route config  from kv error, %v", err)
		return nil, err
	}

	return toRouteConf(routes), nil
}

func toRouteConf(m map[string]interface{}) []*rules.RouterConfig {
	var routeConf []*rules.RouterConfig
	for _, r := range m {
		routeConf = append(routeConf, r.(*rules.RouterConfig))
	}

	return routeConf
}

// GetRoutingConfigWithID 根据服务ID拉取路由配置
func (r *routerRuleStore) GetRoutingConfigWithID(id string) (*rules.RouterConfig, error) {
	tx, err := r.handler.StartTx()
	if err != nil {
		return nil, err
	}

	boldTx := tx.GetDelegateTx().(*bolt.Tx)
	defer func() {
		_ = boldTx.Rollback()
	}()

	return r.getRoutingConfigWithIDTx(boldTx, id)
}

// GetRoutingConfigWithIDTx 根据服务ID拉取路由配置
func (r *routerRuleStore) GetRoutingConfigWithIDTx(tx store.Tx, id string) (*rules.RouterConfig, error) {

	if tx == nil {
		return nil, errors.New("tx is nil")
	}

	boldTx := tx.GetDelegateTx().(*bolt.Tx)
	return r.getRoutingConfigWithIDTx(boldTx, id)
}

func (r *routerRuleStore) getRoutingConfigWithIDTx(tx *bolt.Tx, id string) (*rules.RouterConfig, error) {
	ret := make(map[string]interface{})
	if err := loadValues(tx, tblNameRouting, []string{id}, &rules.RouterConfig{}, ret); err != nil {
		log.Error("[Store][boltdb] load route config  from kv", zap.String("routing-id", id), zap.Error(err))
		return nil, err
	}

	if len(ret) == 0 {
		return nil, nil
	}

	if len(ret) > 1 {
		return nil, ErrMultipleRoutingFound
	}

	val := ret[id].(*rules.RouterConfig)
	if !val.Valid {
		return nil, nil
	}

	return val, nil
}
