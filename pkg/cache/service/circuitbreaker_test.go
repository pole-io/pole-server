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
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/pole-io/pole-server/pkg/common/model"
	"github.com/pole-io/pole-server/pkg/store/mock"
)

/**
 * @brief 创建一个测试mock circuitBreakerCache
 */
func newTestCircuitBreakerCache(t *testing.T) (*gomock.Controller, *mock.MockStore, *circuitBreakerCache) {
	ctl := gomock.NewController(t)

	storage := mock.NewMockStore(ctl)
	rlc := NewCircuitBreakerCache(storage, nil)
	storage.EXPECT().GetUnixSecond(gomock.Any()).AnyTimes().Return(time.Now().Unix(), nil)
	var opt map[string]interface{}
	_ = rlc.Initialize(opt)
	return ctl, storage, rlc.(*circuitBreakerCache)
}

/**
 * @brief 生成熔断规则测试数据
 */
func genModelCircuitBreakers(beginNum, total int) []*model.CircuitBreakerRule {
	out := make([]*model.CircuitBreakerRule, 0, total)

	for i := beginNum; i < total+beginNum; i++ {
		item := &model.CircuitBreakerRule{
			ID:           fmt.Sprintf("id-%d", i),
			Name:         fmt.Sprintf("rule-%d", i),
			DstService:   fmt.Sprintf("svc-%d", i),
			DstNamespace: "test",
			Valid:        true,
			ModifyTime:   time.Unix(int64(i), 0),
		}
		out = append(out, item)
	}
	return out
}

/**
 * @brief 统计缓存中的熔断数据
 */
func getCircuitBreakerCount(cbc *circuitBreakerCache) int {
	return cbc.GetCircuitBreakerCount()
}

/**
 * TestCircuitBreakersUpdate 生成熔断规则测试数据
 */
func TestCircuitBreakersUpdate(t *testing.T) {
	ctl, storage, cbc := newTestCircuitBreakerCache(t)
	defer ctl.Finish()

	total := 10
	serviceWithCircuitBreakers := genModelCircuitBreakers(0, total)

	t.Run("正常更新缓存，可以获取到数据", func(t *testing.T) {
		_ = cbc.Clear()

		storage.EXPECT().GetCircuitBreakerRulesForCache(gomock.Any(), cbc.IsFirstUpdate()).
			Return(serviceWithCircuitBreakers, nil)
		if err := cbc.Update(); err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		// 检查数目是否一致
		if getCircuitBreakerCount(cbc) == total {
			t.Log("pass")
		} else {
			t.Fatalf("actual count is %d", getCircuitBreakerCount(cbc))
		}
	})

	t.Run("缓存数据为空", func(t *testing.T) {
		_ = cbc.Clear()

		storage.EXPECT().GetCircuitBreakerRulesForCache(gomock.Any(), cbc.IsFirstUpdate()).
			Return(nil, nil)
		if err := cbc.Update(); err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		if getCircuitBreakerCount(cbc) == 0 {
			t.Log("pass")
		} else {
			t.Fatalf("actual count is %d", getCircuitBreakerCount(cbc))
		}
	})

	t.Run("lastMtime正确更新", func(t *testing.T) {
		_ = cbc.Clear()

		currentTime := time.Now()
		serviceWithCircuitBreakers[0].ModifyTime = currentTime
		storage.EXPECT().GetUnixSecond(gomock.Any()).Return(currentTime.Unix(), nil).AnyTimes()
		storage.EXPECT().GetCircuitBreakerRulesForCache(gomock.Any(), cbc.IsFirstUpdate()).
			Return(serviceWithCircuitBreakers, nil)
		if err := cbc.Update(); err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		if cbc.OriginLastFetchTime().Unix() == currentTime.Unix() {
			t.Log("pass")
		} else {
			t.Fatalf("last mtime error")
		}
	})

	t.Run("数据库返回错误, update错误", func(t *testing.T) {
		storage.EXPECT().GetCircuitBreakerRulesForCache(gomock.Any(), cbc.IsFirstUpdate()).
			Return(nil, fmt.Errorf("storage error"))
		if err := cbc.Update(); err != nil {
			t.Log("pass")
		} else {
			t.Fatalf("error")
		}
	})
}

/**
 * TestCircuitBreakerUpdate2 统计缓存中的熔断规则数据
 */
func TestCircuitBreakerUpdate2(t *testing.T) {
	ctl, storage, cbc := newTestCircuitBreakerCache(t)
	defer ctl.Finish()

	total := 10

	t.Run("更新缓存后，增加部分数据，缓存正常更新", func(t *testing.T) {
		_ = cbc.Clear()

		serviceWithCircuitBreakers := genModelCircuitBreakers(0, total)
		storage.EXPECT().GetCircuitBreakerRulesForCache(gomock.Any(), cbc.IsFirstUpdate()).
			Return(serviceWithCircuitBreakers, nil)
		if err := cbc.Update(); err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		serviceWithCircuitBreakers = genModelCircuitBreakers(10, total)
		storage.EXPECT().GetCircuitBreakerRulesForCache(gomock.Any(), cbc.IsFirstUpdate()).
			Return(serviceWithCircuitBreakers, nil)
		if err := cbc.Update(); err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		if getCircuitBreakerCount(cbc) == total*2 {
			t.Log("pass")
		} else {
			t.Fatalf("actual count is %d", getCircuitBreakerCount(cbc))
		}
	})

	t.Run("更新缓存后，删除部分数据，缓存正常更新", func(t *testing.T) {
		_ = cbc.Clear()

		serviceWithCircuitBreakers := genModelCircuitBreakers(0, total)
		storage.EXPECT().GetCircuitBreakerRulesForCache(gomock.Any(), cbc.IsFirstUpdate()).
			Return(serviceWithCircuitBreakers, nil)
		if err := cbc.Update(); err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		for i := 0; i < total; i += 2 {
			serviceWithCircuitBreakers[i].Valid = false
		}

		storage.EXPECT().GetCircuitBreakerRulesForCache(gomock.Any(), cbc.IsFirstUpdate()).
			Return(serviceWithCircuitBreakers, nil)
		if err := cbc.Update(); err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		if getCircuitBreakerCount(cbc) == total/2 {
			t.Log("pass")
		} else {
			t.Fatalf("actual count is %d", getCircuitBreakerCount(cbc))
		}
	})
}
