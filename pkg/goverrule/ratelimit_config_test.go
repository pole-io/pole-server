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

package goverrule_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	"github.com/pole-io/pole-server/pkg/cache"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

type CacheListener struct {
	onCreated      func(value interface{})
	onUpdated      func(value interface{})
	onDeleted      func(value interface{})
	onBatchCreated func(value interface{})
	onBatchUpdated func(value interface{})
	onBatchDeleted func(value interface{})
}

// OnCreated callback when cache value created
func (l *CacheListener) OnCreated(value interface{}) {
	if l.onCreated != nil {
		l.onCreated(value)
	}
}

// OnUpdated callback when cache value updated
func (l *CacheListener) OnUpdated(value interface{}) {
	if l.onUpdated != nil {
		l.onUpdated(value)
	}
}

// OnDeleted callback when cache value deleted
func (l *CacheListener) OnDeleted(value interface{}) {
	if l.onDeleted != nil {
		l.onDeleted(value)
	}
}

// OnBatchCreated callback when cache value created
func (l *CacheListener) OnBatchCreated(value interface{}) {
	if l.onBatchCreated != nil {
		l.onBatchCreated(value)
	}
}

// OnBatchUpdated callback when cache value updated
func (l *CacheListener) OnBatchUpdated(value interface{}) {
	if l.onBatchUpdated != nil {
		l.onBatchUpdated(value)
	}
}

// OnBatchDeleted callback when cache value deleted
func (l *CacheListener) OnBatchDeleted(value interface{}) {
	if l.onBatchDeleted != nil {
		l.onBatchDeleted(value)
	}
}

func Test_Echo(t *testing.T) {
	data, _ := json.Marshal(&apitraffic.Rule{
		Method: &apimodel.MatchString{
			Type:      apimodel.MatchString_EXACT,
			Value:     wrapperspb.String("*"),
			ValueType: apimodel.MatchString_TEXT,
		},
	})
	t.Logf("%s", string(data))
}

/**
 * @brief 测试创建限流规则
 */
func TestCreateRateLimit(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	service := &apiservice.Service{Name: protobuf.NewStringValue("testDestService"), Namespace: protobuf.NewStringValue("test")}

	defer discoverSuit.cleanServiceName(service.GetName().GetValue(), service.GetNamespace().GetValue())
	defer discoverSuit.cleanRateLimitRevision(service.GetName().GetValue(), service.GetNamespace().GetValue())

	t.Run("正常创建限流规则", func(t *testing.T) {
		_ = discoverSuit.CacheMgr().Clear()

		time.Sleep(5 * time.Second)

		rateLimitReq, rateLimitResp := discoverSuit.createCommonRateLimit(t, service, 3)
		defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())

		// 等待缓存更新
		_ = discoverSuit.GoverRuleServer().Cache().(*cache.CacheManager).TestUpdate()
		resp := discoverSuit.GoverRuleServer().GetRateLimitWithCache(context.Background(), service)
		checkRateLimit(t, rateLimitReq, resp.GetRateLimit().GetRules()[0])
	})

	t.Run("创建限流规则，删除，再创建，可以正常创建", func(t *testing.T) {
		_ = discoverSuit.CacheMgr().Clear()
		time.Sleep(5 * time.Second)

		rateLimitReq, rateLimitResp := discoverSuit.createCommonRateLimit(t, service, 3)
		defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
		discoverSuit.deleteRateLimit(t, rateLimitResp)
		if resp := discoverSuit.GoverRuleServer().CreateRateLimits(discoverSuit.DefaultCtx, []*apitraffic.Rule{rateLimitReq}); !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}

		// 等待缓存更新
		_ = discoverSuit.GoverRuleServer().Cache().(*cache.CacheManager).TestUpdate()
		resp := discoverSuit.GoverRuleServer().GetRateLimitWithCache(context.Background(), service)
		checkRateLimit(t, rateLimitReq, resp.GetRateLimit().GetRules()[0])
		discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
	})

	t.Run("重复创建限流规则，返回成功", func(t *testing.T) {
		rateLimitReq, rateLimitResp := discoverSuit.createCommonRateLimit(t, service, 3)
		defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
		if resp := discoverSuit.GoverRuleServer().CreateRateLimits(discoverSuit.DefaultCtx, []*apitraffic.Rule{rateLimitReq}); !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		} else {
			t.Log("pass")
		}
		discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
	})

	t.Run("创建限流规则时，没有传递token，返回失败", func(t *testing.T) {

		oldCtx := discoverSuit.DefaultCtx

		discoverSuit.DefaultCtx = context.Background()

		defer func() {
			discoverSuit.DefaultCtx = oldCtx
		}()

		rateLimit := &apitraffic.Rule{
			Service:   service.GetName(),
			Namespace: service.GetNamespace(),
			Labels:    map[string]*apimodel.MatchString{},
		}
		if resp := discoverSuit.GoverRuleServer().CreateRateLimits(discoverSuit.DefaultCtx, []*apitraffic.Rule{rateLimit}); !respSuccess(resp) {
			t.Logf("pass: %s", resp.GetInfo().GetValue())
		} else {
			t.Fatal("error")
		}
	})

	// t.Run("创建限流规则时，没有传递labels，返回失败", func(t *testing.T) {
	// 	rateLimit := &apitraffic.Rule{
	// 		Service:      serviceResp.GetName(),
	// 		Namespace:    serviceResp.GetNamespace(),
	// 		ServiceToken: serviceResp.GetToken(),
	// 	}
	// 	if resp := discoverSuit.GoverRuleServer().CreateRateLimits(discoverSuit.DefaultCtx, []*apitraffic.Rule{rateLimit}); !respSuccess(resp) {
	// 		t.Logf("pass: %s", resp.GetInfo().GetValue())
	// 	} else {
	// 		t.Fatalf("error")
	// 	}
	// })

	t.Run("创建限流规则时，amounts具有相同的duration，返回失败", func(t *testing.T) {
		rateLimit := &apitraffic.Rule{
			Service:   service.GetName(),
			Namespace: service.GetNamespace(),
			Labels:    map[string]*apimodel.MatchString{},
			Amounts: []*apitraffic.Amount{
				{
					MaxAmount: protobuf.NewUInt32Value(1),
					ValidDuration: &duration.Duration{
						Seconds: 10,
						Nanos:   10,
					},
				},
				{
					MaxAmount: protobuf.NewUInt32Value(2),
					ValidDuration: &duration.Duration{
						Seconds: 10,
						Nanos:   10,
					},
				},
			},
			ServiceToken: service.GetToken(),
		}
		if resp := discoverSuit.GoverRuleServer().CreateRateLimits(discoverSuit.DefaultCtx, []*apitraffic.Rule{rateLimit}); !respSuccess(resp) {
			t.Logf("pass: %s", resp.GetInfo().GetValue())
		} else {
			t.Fatalf("error")
		}
	})

	t.Run("并发创建同一服务的限流规则，可以正常创建", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 1; i <= 50; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				_, rateLimitResp := discoverSuit.createCommonRateLimit(t, service, index)
				defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
			}(i)
		}
		wg.Wait()
		t.Log("pass")
	})

	t.Run("并发创建不同服务的限流规则，可以正常创建", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 1; i <= 50; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				service := &apiservice.Service{Name: protobuf.NewStringValue("testDestService-" + strconv.Itoa(index)), Namespace: protobuf.NewStringValue("test")}

				defer discoverSuit.cleanServiceName(service.GetName().GetValue(), service.GetNamespace().GetValue())
				defer discoverSuit.cleanRateLimitRevision(service.GetName().GetValue(), service.GetNamespace().GetValue())
				_, rateLimitResp := discoverSuit.createCommonRateLimit(t, service, 3)
				defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
			}(i)
		}
		wg.Wait()
		t.Log("pass")
	})

	t.Run("为不存在的服务创建限流规则，返回成功", func(t *testing.T) {
		service := &apiservice.Service{Name: protobuf.NewStringValue("testDestService-noexist"), Namespace: protobuf.NewStringValue("test")}
		discoverSuit.cleanServiceName(service.GetName().GetValue(), service.GetNamespace().GetValue())
		rateLimit := &apitraffic.Rule{
			Service:      service.GetName(),
			Namespace:    service.GetNamespace(),
			Labels:       map[string]*apimodel.MatchString{},
			ServiceToken: service.GetToken(),
		}
		if resp := discoverSuit.GoverRuleServer().CreateRateLimits(discoverSuit.DefaultCtx, []*apitraffic.Rule{rateLimit}); respSuccess(resp) {
			t.Logf("pass: %s", resp.GetInfo().GetValue())
		} else {
			t.Fatalf("error : %s", resp.GetInfo().GetValue())
		}
	})
}

/**
 * @brief 测试删除限流规则
 */
func TestDeleteRateLimit(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	service := &apiservice.Service{Name: protobuf.NewStringValue("testDestService"), Namespace: protobuf.NewStringValue("test")}
	defer discoverSuit.cleanServiceName(service.GetName().GetValue(), service.GetNamespace().GetValue())
	defer discoverSuit.cleanRateLimitRevision(service.GetName().GetValue(), service.GetNamespace().GetValue())

	getRateLimits := func(t *testing.T, service *apiservice.Service, expectNum uint32) []*apitraffic.Rule {
		filters := map[string]string{
			"service":   service.GetName().GetValue(),
			"namespace": service.GetNamespace().GetValue(),
		}
		resp := discoverSuit.GoverRuleServer().GetRateLimits(discoverSuit.DefaultCtx, filters)
		if !respSuccess(resp) {
			t.Fatalf("error")
		}
		if resp.GetAmount().GetValue() != expectNum {
			t.Fatalf("error")
		}
		return resp.GetRateLimits()
	}

	t.Run("删除存在的限流规则，可以正常删除", func(t *testing.T) {
		_, rateLimitResp := discoverSuit.createCommonRateLimit(t, service, 3)
		defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
		discoverSuit.deleteRateLimit(t, rateLimitResp)
		getRateLimits(t, service, 0)
		t.Log("pass")
	})

	t.Run("删除不存在的限流规则，返回正常", func(t *testing.T) {
		_, rateLimitResp := discoverSuit.createCommonRateLimit(t, service, 3)
		discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
		discoverSuit.deleteRateLimit(t, rateLimitResp)
		getRateLimits(t, service, 0)
		t.Log("pass")
	})

	t.Run("删除限流规则时，没有传递token，返回失败", func(t *testing.T) {
		rateLimitReq, rateLimitResp := discoverSuit.createCommonRateLimit(t, service, 3)
		defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
		rateLimitReq.ServiceToken = protobuf.NewStringValue("")

		oldCtx := discoverSuit.DefaultCtx

		discoverSuit.DefaultCtx = context.Background()

		defer func() {
			discoverSuit.DefaultCtx = oldCtx
		}()

		resp := discoverSuit.GoverRuleServer().DeleteRateLimits(discoverSuit.DefaultCtx, []*apitraffic.Rule{rateLimitReq})
		assert.False(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
	})

	t.Run("并发删除限流规则，可以正常删除", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 1; i <= 50; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				service := &apiservice.Service{Name: protobuf.NewStringValue("testDestService-" + strconv.Itoa(index)), Namespace: protobuf.NewStringValue("test")}
				defer discoverSuit.cleanServiceName(service.GetName().GetValue(), service.GetNamespace().GetValue())
				defer discoverSuit.cleanRateLimitRevision(service.GetName().GetValue(), service.GetNamespace().GetValue())
				rateLimitReq, rateLimitResp := discoverSuit.createCommonRateLimit(t, service, 3)
				defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
				discoverSuit.deleteRateLimit(t, rateLimitReq)
			}(i)
		}
		wg.Wait()
		t.Log("pass")
	})
}

/**
 * @brief 测试更新限流规则
 */
func TestUpdateRateLimit(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	service := &apiservice.Service{Name: protobuf.NewStringValue("testDestService-" + strconv.Itoa(0)), Namespace: protobuf.NewStringValue("test")}
	defer discoverSuit.cleanServiceName(service.GetName().GetValue(), service.GetNamespace().GetValue())
	defer discoverSuit.cleanRateLimitRevision(service.GetName().GetValue(), service.GetNamespace().GetValue())

	_, rateLimitResp := discoverSuit.createCommonRateLimit(t, service, 1)
	defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())

	t.Run("01-更新单个限流规则，可以正常更新", func(t *testing.T) {
		updateRateLimitContent(rateLimitResp, 2)
		discoverSuit.updateRateLimit(t, rateLimitResp)
		filters := map[string]string{
			"service":   service.GetName().GetValue(),
			"namespace": service.GetNamespace().GetValue(),
		}
		resp := discoverSuit.GoverRuleServer().GetRateLimits(discoverSuit.DefaultCtx, filters)
		if !respSuccess(resp) {
			t.Fatalf("error")
		}
		assert.True(t, len(resp.GetRateLimits()) > 0)
		checkRateLimit(t, rateLimitResp, resp.GetRateLimits()[0])
	})

	t.Run("02-更新一个不存在的限流规则", func(t *testing.T) {
		discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
		if resp := discoverSuit.GoverRuleServer().UpdateRateLimits(discoverSuit.DefaultCtx, []*apitraffic.Rule{rateLimitResp}); !respSuccess(resp) {
			t.Logf("pass: %s", resp.GetInfo().GetValue())
		} else {
			t.Fatalf("error")
		}
	})

	t.Run("03-更新限流规则时，没有传递token，正常", func(t *testing.T) {
		oldCtx := discoverSuit.DefaultCtx
		discoverSuit.DefaultCtx = context.Background()

		defer func() {
			discoverSuit.DefaultCtx = oldCtx
		}()

		rateLimitResp.ServiceToken = protobuf.NewStringValue("")
		if resp := discoverSuit.GoverRuleServer().UpdateRateLimits(discoverSuit.DefaultCtx, []*apitraffic.Rule{rateLimitResp}); !respSuccess(resp) {
			t.Logf("pass: %s", resp.GetInfo().GetValue())
		} else {
			t.Fatalf("error")
		}
	})

	t.Run("04-并发更新限流规则时，可以正常更新", func(t *testing.T) {
		var wg sync.WaitGroup
		errs := make(chan error)

		lock := &sync.RWMutex{}
		waitDelSvcs := []*apiservice.Service{}
		waitDelRules := []*apitraffic.Rule{}

		t.Cleanup(func() {
			for i := range waitDelSvcs {
				serviceResp := waitDelSvcs[i]
				discoverSuit.cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
			}
			for i := range waitDelRules {
				rateLimitResp := waitDelRules[i]
				discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
			}
		})

		for i := 1; i <= 50; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				discoverSuit.cleanRateLimitRevision(service.GetName().GetValue(), service.GetNamespace().GetValue())
				_, rateLimitResp := discoverSuit.createCommonRateLimit(t, service, index)
				updateRateLimitContent(rateLimitResp, index+1)
				discoverSuit.updateRateLimit(t, rateLimitResp)

				func() {
					lock.Lock()
					defer lock.Unlock()

					waitDelSvcs = append(waitDelSvcs, service)
					waitDelRules = append(waitDelRules, rateLimitResp)
				}()

				_ = discoverSuit.CacheMgr().TestUpdate()

				filters := map[string]string{
					"service":   service.GetName().GetValue(),
					"namespace": service.GetNamespace().GetValue(),
				}
				resp := discoverSuit.GoverRuleServer().GetRateLimits(discoverSuit.DefaultCtx, filters)
				if !respSuccess(resp) {
					errs <- fmt.Errorf("error : %v", resp)
				}
				if len(resp.GetRateLimits()) == 0 {
					errs <- errors.New("ratelimit rule count is zero")
					return
				}
				checkRateLimit(t, rateLimitResp, resp.GetRateLimits()[0])
			}(i)
		}
		go func() {
			wg.Wait()
			close(errs)
		}()

		for err := range errs {
			if err != nil {
				t.Fatal(err)
			}
		}

		t.Log("pass")
	})
}

func TestDisableRateLimit(t *testing.T) {
	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	serviceResp := &apiservice.Service{Name: protobuf.NewStringValue("testDestService-" + strconv.Itoa(0)), Namespace: protobuf.NewStringValue("test")}
	defer discoverSuit.cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
	defer discoverSuit.cleanRateLimitRevision(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())

	_, rateLimitResp := discoverSuit.createCommonRateLimit(t, serviceResp, 1)
	defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())

	t.Run("反复启用禁止限流规则, 正常下发客户端", func(t *testing.T) {
		_, rateLimitResp := discoverSuit.createCommonRateLimit(t, serviceResp, 10000)
		defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
		delay := time.NewTimer(time.Second)
		t.Cleanup(func() {
			delay.Stop()
		})

		check := func(label string, disable bool) {
			ruleContents := []*apitraffic.Rule{
				{
					Id:      protobuf.NewStringValue(rateLimitResp.GetId().GetValue()),
					Disable: protobuf.NewBoolValue(disable),
				},
			}

			t.Logf("start run : %s", label)
			if resp := discoverSuit.GoverRuleServer().EnableRateLimits(discoverSuit.DefaultCtx, ruleContents); !respSuccess(resp) {
				t.Fatalf("error: %s", resp.GetInfo().GetValue())
			}
			filters := map[string]string{
				"id": rateLimitResp.GetId().GetValue(),
			}
			resp := discoverSuit.GoverRuleServer().GetRateLimits(discoverSuit.DefaultCtx, filters)
			if !respSuccess(resp) {
				t.Fatalf("error : %s", resp.GetInfo().GetValue())
			}
			assert.Equal(t, 1, len(resp.GetRateLimits()))

			data, _ := json.Marshal(resp.GetRateLimits())
			t.Logf("find target ratelimit rule from store : %s", string(data))

			assert.Equal(t, rateLimitResp.GetId().GetValue(), resp.GetRateLimits()[0].GetId().GetValue())
			assert.Equal(t, disable, resp.GetRateLimits()[0].GetDisable().GetValue())

			time.Sleep(10 * time.Second)

			var ok bool
			for i := 0; i < 3; i++ {
				discoverResp := discoverSuit.GoverRuleServer().GetRateLimitWithCache(discoverSuit.DefaultCtx, serviceResp)
				if !respSuccess(discoverResp) {
					t.Fatalf("error: %s", resp.GetInfo().GetValue())
				}

				assert.True(t, len(discoverResp.GetRateLimit().GetRules()) > 0)

				for i := range discoverResp.GetRateLimit().GetRules() {
					rule := discoverResp.GetRateLimit().GetRules()[i]
					if rule.GetId().GetValue() == rateLimitResp.GetId().GetValue() {
						data, _ := json.Marshal(rule)
						t.Logf("find target ratelimit rule from cache : %s", string(data))
						if disable == rule.GetDisable().GetValue() {
							ok = true
							break
						}
						time.Sleep(time.Second)
					}
				}
			}
			if !ok {
				t.Fatalf("%s match ratelimit disable status", label)
			} else {
				t.Logf("start run : success : %s %s", rateLimitResp.GetId().GetValue(), resp.GetRateLimits()[0].GetId().GetValue())
			}
		}

		check("禁用限流规则", true)
		time.Sleep(time.Second)
		check("启用限流规则", false)
		time.Sleep(time.Second)
		check("禁用限流规则", true)
		time.Sleep(time.Second)
		check("启用限流规则", false)
		time.Sleep(time.Second)

	})
}

/*
 * @brief 测试查询限流规则
 */
func TestGetRateLimit(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	serviceNum := 10
	rateLimitsNum := 30
	rateLimits := make([]*apitraffic.Rule, rateLimitsNum)
	serviceName := ""
	namespaceName := ""
	for i := 0; i < serviceNum; i++ {
		serviceName = fmt.Sprintf("ratelimit_service_%d", i)
		namespaceName = fmt.Sprintf("ratelimit_namespace_%d", i)
		defer discoverSuit.cleanRateLimitRevision(serviceName, namespaceName)
		for j := 0; j < rateLimitsNum/serviceNum; j++ {
			_, rateLimitResp := discoverSuit.createCommonRateLimit(t, &apiservice.Service{
				Name:      protobuf.NewStringValue(serviceName),
				Namespace: protobuf.NewStringValue(namespaceName),
			}, j)
			defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
			rateLimits = append(rateLimits, rateLimitResp)
		}
	}

	t.Run("查询限流规则，过滤条件为service", func(t *testing.T) {
		filters := map[string]string{
			"service": serviceName,
		}
		resp := discoverSuit.GoverRuleServer().GetRateLimits(discoverSuit.DefaultCtx, filters)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
		if resp.GetSize().GetValue() != uint32(rateLimitsNum/serviceNum) {
			t.Fatalf("expect num is %d, actual num is %d", rateLimitsNum/serviceNum, resp.GetSize().GetValue())
		}
		t.Logf("pass: num is %d", resp.GetSize().GetValue())
	})

	t.Run("查询限流规则，过滤条件为namespace", func(t *testing.T) {
		filters := map[string]string{
			"namespace": namespaceName,
		}
		resp := discoverSuit.GoverRuleServer().GetRateLimits(discoverSuit.DefaultCtx, filters)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
		if resp.GetSize().GetValue() != uint32(rateLimitsNum/serviceNum) {
			t.Fatalf("expect num is %d, actual num is %d", rateLimitsNum/serviceNum, resp.GetSize().GetValue())
		}
		t.Logf("pass: num is %d", resp.GetSize().GetValue())
	})

	t.Run("查询限流规则，过滤条件为不存在的namespace", func(t *testing.T) {
		filters := map[string]string{
			"namespace": "Development",
		}
		resp := discoverSuit.GoverRuleServer().GetRateLimits(discoverSuit.DefaultCtx, filters)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
		if resp.GetSize().GetValue() != 0 {
			t.Fatalf("expect num is 0, actual num is %d", resp.GetSize().GetValue())
		}
		t.Logf("pass: num is %d", resp.GetSize().GetValue())
	})

	t.Run("查询限流规则，过滤条件为namespace和service", func(t *testing.T) {
		filters := map[string]string{
			"service":   serviceName,
			"namespace": namespaceName,
		}
		resp := discoverSuit.GoverRuleServer().GetRateLimits(discoverSuit.DefaultCtx, filters)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
		if resp.GetSize().GetValue() != uint32(rateLimitsNum/serviceNum) {
			t.Fatalf("expect num is %d, actual num is %d", rateLimitsNum/serviceNum, resp.GetSize().GetValue())
		}
		t.Logf("pass: num is %d", resp.GetSize().GetValue())
	})

	t.Run("查询限流规则，过滤条件为offset和limit", func(t *testing.T) {
		filters := map[string]string{
			"offset": "1",
			"limit":  "5",
		}
		resp := discoverSuit.GoverRuleServer().GetRateLimits(discoverSuit.DefaultCtx, filters)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
		if resp.GetSize().GetValue() != 5 {
			t.Fatalf("expect num is 5, actual num is %d", resp.GetSize().GetValue())
		}
		t.Logf("pass: num is %d", resp.GetSize().GetValue())
	})

	t.Run("查询限流规则列表，过滤条件为name", func(t *testing.T) {
		filters := map[string]string{
			"name":  "rule_name_0",
			"brief": "true",
		}
		resp := discoverSuit.GoverRuleServer().GetRateLimits(discoverSuit.DefaultCtx, filters)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
		if resp.GetSize().GetValue() != uint32(serviceNum) {
			t.Fatalf("expect num is %d, actual num is %d", serviceNum, resp.GetSize().GetValue())
		}
	})

	t.Run("查询限流规则，offset为负数，返回错误", func(t *testing.T) {
		filters := map[string]string{
			"service":   serviceName,
			"namespace": namespaceName,
			"offset":    "-5",
		}
		resp := discoverSuit.GoverRuleServer().GetRateLimits(discoverSuit.DefaultCtx, filters)
		if !respSuccess(resp) {
			t.Logf("pass: %s", resp.GetInfo().GetValue())
		} else {
			t.Fatalf("error")
		}
	})

	t.Run("查询限流规则，limit为负数，返回错误", func(t *testing.T) {
		filters := map[string]string{
			"service":   serviceName,
			"namespace": namespaceName,
			"limit":     "-5",
		}
		resp := discoverSuit.GoverRuleServer().GetRateLimits(discoverSuit.DefaultCtx, filters)
		if !respSuccess(resp) {
			t.Logf("pass: %s", resp.GetInfo().GetValue())
		} else {
			t.Fatalf("error")
		}
	})
}

// test对ratelimit字段进行校验
func TestCheckRatelimitFieldLen(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	rateLimit := &apitraffic.Rule{
		Service:      protobuf.NewStringValue("test"),
		Namespace:    protobuf.NewStringValue("default"),
		Labels:       map[string]*apimodel.MatchString{},
		ServiceToken: protobuf.NewStringValue("test"),
	}
	t.Run("创建限流规则，服务名超长", func(t *testing.T) {
		str := genSpecialStr(129)
		oldName := rateLimit.Service
		rateLimit.Service = protobuf.NewStringValue(str)
		resp := discoverSuit.GoverRuleServer().CreateRateLimits(discoverSuit.DefaultCtx, []*apitraffic.Rule{rateLimit})
		rateLimit.Service = oldName
		if resp.Code.Value != api.InvalidServiceName {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("创建限流规则，命名空间超长", func(t *testing.T) {
		str := genSpecialStr(129)
		oldNamespace := rateLimit.Namespace
		rateLimit.Namespace = protobuf.NewStringValue(str)
		resp := discoverSuit.GoverRuleServer().CreateRateLimits(discoverSuit.DefaultCtx, []*apitraffic.Rule{rateLimit})
		rateLimit.Namespace = oldNamespace
		if resp.Code.Value != api.InvalidNamespaceName {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("创建限流规则，名称超长", func(t *testing.T) {
		str := genSpecialStr(2049)
		oldName := rateLimit.Name
		rateLimit.Name = protobuf.NewStringValue(str)
		resp := discoverSuit.GoverRuleServer().CreateRateLimits(discoverSuit.DefaultCtx, []*apitraffic.Rule{rateLimit})
		rateLimit.Name = oldName
		if resp.Code.Value != api.InvalidRateLimitName {
			t.Fatalf("%+v", resp)
		}
	})
}
