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

package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	cachetypes "github.com/pole-io/pole-server/apis/cache"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

func NewServiceContractCache(storage store.Store, cacheMgr cachetypes.CacheManager) cachetypes.ServiceContractCache {
	return &ServiceContractCache{
		BaseCache: cachebase.NewBaseCache(storage, cacheMgr),
	}
}

type ServiceContractCache struct {
	*cachebase.BaseCache
	// data namespace/service/type/protocol/version -> *svctypes.EnrichServiceContract
	data *container.SyncMap[string, *svctypes.EnrichServiceContract]
	// valueCache save ConfigFileRelease.Content into local file to reduce memory use
	valueCache  *bbolt.DB
	singleGroup singleflight.Group
}

// Initialize
func (sc *ServiceContractCache) Initialize(c map[string]interface{}) error {
	valueCache, err := sc.openBoltCache(c)
	if err != nil {
		return err
	}
	sc.valueCache = valueCache
	sc.data = container.NewSyncMap[string, *svctypes.EnrichServiceContract]()
	return nil
}

func (fc *ServiceContractCache) openBoltCache(opt map[string]interface{}) (*bbolt.DB, error) {
	path, _ := opt["cachePath"].(string)
	if path == "" {
		path = "./data/cache/service_contract"
	}
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, err
	}
	dbFile := filepath.Join(path, "service_contract.bolt")
	_ = os.Remove(dbFile)
	valueCache, err := bbolt.Open(dbFile, os.ModePerm, &bbolt.Options{})
	if err != nil {
		return nil, err
	}
	return valueCache, nil
}

// Update
func (sc *ServiceContractCache) Update() error {
	err, _ := sc.singleUpdate()
	return err
}

func (sc *ServiceContractCache) singleUpdate() (error, bool) {
	// 多个线程竞争，只有一个线程进行更新
	_, err, shared := sc.singleGroup.Do(sc.Name(), func() (interface{}, error) {
		return nil, sc.DoCacheUpdate(sc.Name(), sc.realUpdate)
	})
	return err, shared
}

func (sc *ServiceContractCache) realUpdate() (map[string]time.Time, int64, error) {
	start := time.Now()
	values, err := sc.Store().GetMoreServiceContracts(sc.IsFirstUpdate(), sc.LastFetchTime())
	if err != nil {
		log.Error("[Cache][ServiceContract] update service_contract", zap.Error(err))
		return nil, 0, err
	}

	lastMtimes, update, del := sc.setContracts(values)
	costTime := time.Since(start)
	log.Info("[Cache][ServiceContract] get more service_contract", zap.Int("total", len(values)),
		zap.Int("upsert", update), zap.Int("delete", del), zap.Time("last", sc.LastMtime(sc.Name())),
		zap.Duration("used", costTime))
	return lastMtimes, int64(len(values)), err
}

func (sc *ServiceContractCache) setContracts(values []*svctypes.EnrichServiceContract) (map[string]time.Time, int, int) {
	var (
		upsert, del int
		lastMtime   time.Time
	)
	for i := range values {
		item := values[i]
		if !item.Valid {
			del++
			_ = sc.upsertValueCache(item, true)
			continue
		}
		upsert++
		_ = sc.upsertValueCache(item, false)
	}
	return map[string]time.Time{
		sc.Name(): lastMtime,
	}, upsert, del
}

// Clear
func (sc *ServiceContractCache) Clear() error {
	sc.data = container.NewSyncMap[string, *svctypes.EnrichServiceContract]()
	return nil
}

// Name
func (sc *ServiceContractCache) Name() string {
	return cachetypes.ServiceContractName
}

func (sc *ServiceContractCache) Get(ctx context.Context, req *svctypes.ServiceContract) *svctypes.EnrichServiceContract {
	ret, _ := sc.loadValueCache(req)
	return ret
}

func (fc *ServiceContractCache) upsertValueCache(item *svctypes.EnrichServiceContract, del bool) error {
	return fc.valueCache.Update(func(tx *bbolt.Tx) error {
		if del {
			return tx.DeleteBucket([]byte(item.GetCacheKey()))
		}
		bucket, err := tx.CreateBucketIfNotExists([]byte(item.GetCacheKey()))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(item.GetCacheKey()), []byte(utils.MustJson(item)))
	})
}

func (fc *ServiceContractCache) loadValueCache(release *svctypes.ServiceContract) (*svctypes.EnrichServiceContract, error) {
	ret := &svctypes.EnrichServiceContract{}
	found := false
	err := fc.valueCache.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(release.GetCacheKey()))
		if bucket == nil {
			return nil
		}
		found = true
		val := bucket.Get([]byte(release.GetCacheKey()))
		return json.Unmarshal(val, ret)
	})
	if !found {
		ret.ServiceContract = &svctypes.ServiceContract{}
	}
	return ret, err
}
