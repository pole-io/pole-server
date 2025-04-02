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

package base

import (
	"runtime"
	"sync"
	"time"

	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/common/metrics"

	cachetypes "github.com/pole-io/pole-server/apis/cache"
)

// BaseCache 对于 Cache 中的一些 func 做统一实现，避免重复逻辑
type BaseCache struct {
	lock sync.RWMutex
	//
	timeDiff time.Duration
	// firstUpdate Whether the cache is loaded for the first time
	// this field can only make value on exec initialize/clean, and set it to false on exec update
	firstUpdate           bool
	s                     store.Store
	lastFetchTime         int64
	lastMtimes            map[string]time.Time
	CacheMgr              cachetypes.CacheManager
	reportMetrics         func()
	lastReportMetricsTime time.Time
}

func NewBaseCache(s store.Store, cacheMgr cachetypes.CacheManager) *BaseCache {
	c := &BaseCache{
		s:        s,
		CacheMgr: cacheMgr,
	}

	c.initialize()
	return c
}

func NewBaseCacheWithRepoerMetrics(s store.Store, cacheMgr cachetypes.CacheManager, reportMetrics func()) *BaseCache {
	c := &BaseCache{
		s:             s,
		CacheMgr:      cacheMgr,
		reportMetrics: reportMetrics,
	}

	c.initialize()
	return c
}

func (bc *BaseCache) initialize() {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	bc.lastFetchTime = 1
	bc.firstUpdate = true
	bc.lastMtimes = map[string]time.Time{}
}

var (
	zeroTime = time.Unix(0, 0)
)

func (bc *BaseCache) Store() store.Store {
	return bc.s
}

func (bc *BaseCache) ResetLastMtime(label string) {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	bc.lastMtimes[label] = time.Unix(0, 0)
}

func (bc *BaseCache) ResetLastFetchTime() {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	bc.lastFetchTime = 1
}

func (bc *BaseCache) LastMtime(label string) time.Time {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	v, ok := bc.lastMtimes[label]
	if ok {
		return v
	}

	return time.Unix(0, 0)
}

func (bc *BaseCache) LastFetchTime() time.Time {
	lastTime := time.Unix(bc.lastFetchTime, 0)
	tmp := lastTime.Add(bc.timeDiff)
	if zeroTime.After(tmp) {
		return lastTime
	}
	lastTime = tmp
	return lastTime
}

// OriginLastFetchTime only for test
func (bc *BaseCache) OriginLastFetchTime() time.Time {
	lastTime := time.Unix(bc.lastFetchTime, 0)
	return lastTime
}

func (bc *BaseCache) IsFirstUpdate() bool {
	return bc.firstUpdate
}

// update
func (bc *BaseCache) DoCacheUpdate(name string, executor func() (map[string]time.Time, int64, error)) error {
	if bc.IsFirstUpdate() {
		log.Infof("[Cache][%s] begin run cache update work", name)
	}

	curStoreTime, err := bc.s.GetUnixSecond(0)
	if err != nil {
		curStoreTime = bc.lastFetchTime
		log.Warnf("[Cache][%s] get store timestamp fail, skip update lastMtime, err : %v", name, err)
	}
	defer func() {
		if err := recover(); err != nil {
			var buf [4086]byte
			n := runtime.Stack(buf[:], false)
			log.Errorf("[Cache][%s] run cache update panic: %+v, stack\n%s\n", name, err, string(buf[:n]))
		} else {
			bc.lastFetchTime = curStoreTime
		}
	}()

	start := time.Now()
	lastMtimes, total, err := executor()
	if err != nil {
		return err
	}

	bc.lock.Lock()
	defer bc.lock.Unlock()
	if len(lastMtimes) != 0 {
		if len(bc.lastMtimes) != 0 {
			for label, lastMtime := range lastMtimes {
				preLastMtime := bc.lastMtimes[label]
				log.Infof("[Cache][%s] lastFetchTime %s, lastMtime update from %s to %s",
					label, time.Unix(bc.lastFetchTime, 0), preLastMtime, lastMtime)
			}
		}
		bc.lastMtimes = lastMtimes
	}

	if total >= 0 {
		metrics.RecordCacheUpdateCost(time.Since(start), name, total)
	}
	if bc.reportMetrics != nil {
		if time.Since(bc.lastReportMetricsTime) >= bc.CacheMgr.GetReportInterval() {
			bc.reportMetrics()
			bc.lastReportMetricsTime = start
		}
	}
	bc.firstUpdate = false
	return nil
}

func (bc *BaseCache) Clear() {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	bc.lastMtimes = make(map[string]time.Time)
	bc.lastFetchTime = 1
	bc.firstUpdate = true
}

func (bc *BaseCache) Close() error {
	return nil
}

func (bc *BaseCache) RefreshInterval() time.Duration {
	return time.Second
}
