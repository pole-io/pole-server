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
	"encoding/json"
	"time"

	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	commontime "github.com/pole-io/pole-server/pkg/common/time"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

const (
	tblClient string = "client"

	ClientFieldHost       string = "Host"
	ClientFieldType       string = "Type"
	ClientFieldVersion    string = "Version"
	ClientFieldLocation   string = "Location"
	ClientFieldId         string = "Id"
	ClientFieldStatArrStr string = "StatArrStr"
	ClientFieldCtime      string = "Ctime"
	ClientFieldMtime      string = "Mtime"
	ClientFieldValid      string = "Valid"
)

type clientObject struct {
	Host       string
	Type       string
	Version    string
	Location   map[string]string
	Id         string
	Ctime      time.Time
	Mtime      time.Time
	StatArrStr string
	Valid      bool
}

type clientStore struct {
	handler BoltHandler
}

// BatchAddClients insert the client info
func (cs *clientStore) BatchAddClients(clients []*types.Client) error {
	if err := cs.handler.Execute(true, func(tx *bolt.Tx) error {
		for i := range clients {
			client := clients[i]
			saveVal, err := convertToClientObject(client)
			if err != nil {
				return err
			}

			if err := saveValue(tx, tblClient, saveVal.Id, saveVal); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Error("[Client] batch add clients", zap.Error(err))
		return err
	}

	return nil
}

// BatchDeleteClients delete the client info
func (cs *clientStore) BatchDeleteClients(ids []string) error {
	err := cs.handler.Execute(true, func(tx *bolt.Tx) error {
		for i := range ids {
			properties := make(map[string]interface{})
			properties[ClientFieldValid] = false
			properties[ClientFieldMtime] = time.Now()

			if err := updateValue(tx, tblClient, ids[i], properties); err != nil {
				log.Error("[Client] batch delete clients", zap.Error(err))
				return err
			}
		}

		return nil
	})

	return err
}

// GetMoreClients 根据mtime获取增量clients，返回所有store的变更信息
func (cs *clientStore) GetMoreClients(mtime time.Time, firstUpdate bool) (map[string]*types.Client, error) {
	fields := []string{ClientFieldMtime, ClientFieldValid}
	if firstUpdate {
		mtime = time.Time{}
	}
	ret, err := cs.handler.LoadValuesByFilter(tblClient, fields, &clientObject{}, func(m map[string]interface{}) bool {
		if firstUpdate {
			// 首次更新，那么就只看 valid 状态
			valid, _ := m[ClientFieldValid].(bool)
			return valid
		}
		return !m[ClientFieldMtime].(time.Time).Before(mtime)
	})

	if err != nil {
		log.Error("[Client] get more clients for cache", zap.Error(err))
		return nil, err
	}

	clients := make(map[string]*types.Client, len(ret))
	for k, v := range ret {
		client, err := convertToModelClient(v.(*clientObject))
		if err != nil {
			log.Error("[Client] convert clientObject to types.Client", zap.Error(err))
			return nil, err
		}

		clients[k] = client
	}

	return clients, nil
}

func convertToClientObject(client *types.Client) (*clientObject, error) {
	stat := client.Proto().Stat
	data, err := json.Marshal(stat)
	if err != nil {
		return nil, err
	}
	tn := time.Now()
	return &clientObject{
		Host:    client.Proto().GetHost().GetValue(),
		Type:    client.Proto().GetType().String(),
		Version: client.Proto().GetVersion().GetValue(),
		Location: map[string]string{
			"region": client.Proto().GetLocation().GetRegion().GetValue(),
			"zone":   client.Proto().GetLocation().GetZone().GetValue(),
			"campus": client.Proto().GetLocation().GetCampus().GetValue(),
		},
		Id:         client.Proto().GetId().GetValue(),
		Ctime:      tn,
		Mtime:      tn,
		StatArrStr: string(data),
		Valid:      true,
	}, nil
}

func convertToModelClient(client *clientObject) (*types.Client, error) {
	stat := make([]*apiservice.StatInfo, 0, 4)
	err := json.Unmarshal([]byte(client.StatArrStr), &stat)
	if err != nil {
		return nil, err
	}

	c := &apiservice.Client{
		Id:      utils.NewStringValue(client.Id),
		Host:    utils.NewStringValue(client.Host),
		Type:    apiservice.Client_ClientType(apiservice.Client_ClientType_value[client.Type]),
		Version: utils.NewStringValue(client.Version),
		Ctime:   utils.NewStringValue(commontime.Time2String(client.Ctime)),
		Mtime:   utils.NewStringValue(commontime.Time2String(client.Mtime)),
		Location: &apimodel.Location{
			Region: utils.NewStringValue(client.Location["region"]),
			Zone:   utils.NewStringValue(client.Location["zone"]),
			Campus: utils.NewStringValue(client.Location["campus"]),
		},
		Stat: stat,
	}

	mc := types.NewClient(c)
	mc.SetValid(client.Valid)
	return mc, nil
}
