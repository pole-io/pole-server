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
	"testing"

	"github.com/stretchr/testify/assert"

	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	"github.com/pole-io/pole-server/pkg/cache"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// 测试discover circuitbreaker
func TestDiscoverCircuitBreaker(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	t.Run("熔断规则测试", func(t *testing.T) {
		rules, resp := createCircuitBreakerRules(discoverSuit, 5)
		defer cleanCircuitBreakerRules(discoverSuit, resp)
		service := &apiservice.Service{Name: protobuf.NewStringValue("testDestService"), Namespace: protobuf.NewStringValue("test")}
		t.Run("正常获取熔断规则", func(t *testing.T) {
			_ = discoverSuit.GoverRuleServer().Cache().(*cache.CacheManager).TestUpdate()
			out := discoverSuit.GoverRuleServer().GetCircuitBreakerWithCache(discoverSuit.DefaultCtx, service)
			assert.True(t, respSuccess(out))
			assert.Equal(t, len(out.GetCircuitBreaker().GetRules()), len(rules))
			t.Logf("pass: out is %+v", out)

			// 再次请求
			out = discoverSuit.GoverRuleServer().GetCircuitBreakerWithCache(discoverSuit.DefaultCtx, out.GetService())
			assert.True(t, respSuccess(out))
			assert.Equal(t, out.GetCode().GetValue(), api.DataNoChange)
			t.Logf("pass: out is %+v", out)
		})
	})
}

// 测试discover circuitbreaker
func TestDiscoverCircuitBreaker2(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	t.Run("熔断规则异常测试", func(t *testing.T) {
		_, resp := createCircuitBreakerRules(discoverSuit, 1)
		defer cleanCircuitBreakerRules(discoverSuit, resp)
		service := &apiservice.Service{Name: protobuf.NewStringValue("testDestService"), Namespace: protobuf.NewStringValue("default")}
		t.Run("熔断规则不存在", func(t *testing.T) {
			_ = discoverSuit.GoverRuleServer().Cache().(*cache.CacheManager).TestUpdate()
			out := discoverSuit.GoverRuleServer().GetCircuitBreakerWithCache(discoverSuit.DefaultCtx, service)
			assert.True(t, respSuccess(out))
			assert.Equal(t, 0, len(out.GetCircuitBreaker().GetRules()))
			t.Logf("pass: out is %+v", out)
		})
	})
}

// 测试discover ratelimit
func TestDiscoverRateLimits(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	t.Run("限流规则测试", func(t *testing.T) {
		service := &apiservice.Service{Name: protobuf.NewStringValue("testDestService"), Namespace: protobuf.NewStringValue("test")}
		defer discoverSuit.cleanServiceName(service.GetName().GetValue(), service.GetNamespace().GetValue())
		_, rateLimitResp := discoverSuit.createCommonRateLimit(t, service, 1)
		defer discoverSuit.cleanRateLimit(rateLimitResp.GetId().GetValue())
		defer discoverSuit.cleanRateLimitRevision(service.GetName().GetValue(), service.GetNamespace().GetValue())
		t.Run("正常获取限流规则", func(t *testing.T) {
			_ = discoverSuit.GoverRuleServer().Cache().(*cache.CacheManager).TestUpdate()
			out := discoverSuit.GoverRuleServer().GetRateLimitWithCache(discoverSuit.DefaultCtx, service)
			assert.True(t, respSuccess(out))
			assert.Equal(t, len(out.GetRateLimit().GetRules()), 1)
			checkRateLimit(t, rateLimitResp, out.GetRateLimit().GetRules()[0])
			t.Logf("pass: out is %+v", out)
			// 再次请求
			out = discoverSuit.GoverRuleServer().GetRateLimitWithCache(discoverSuit.DefaultCtx, out.GetService())
			assert.True(t, respSuccess(out))
			assert.Equal(t, out.GetCode().GetValue(), api.DataNoChange)
			t.Logf("pass: out is %+v", out)
		})
		t.Run("限流规则已删除", func(t *testing.T) {
			discoverSuit.deleteRateLimit(t, rateLimitResp)
			_ = discoverSuit.GoverRuleServer().Cache().(*cache.CacheManager).TestUpdate()
			out := discoverSuit.GoverRuleServer().GetRateLimitWithCache(discoverSuit.DefaultCtx, service)
			assert.True(t, respSuccess(out))
			assert.Equal(t, len(out.GetRateLimit().GetRules()), 0)
			t.Logf("pass: out is %+v", out)
		})
	})
}

// 测试discover ratelimit
func TestDiscoverRateLimits2(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	service := &apiservice.Service{Name: protobuf.NewStringValue("testDestService"), Namespace: protobuf.NewStringValue("test")}

	t.Run("限流规则异常测试", func(t *testing.T) {
		t.Run("限流规则不存在", func(t *testing.T) {
			_ = discoverSuit.GoverRuleServer().Cache().(*cache.CacheManager).TestUpdate()
			out := discoverSuit.GoverRuleServer().GetRateLimitWithCache(discoverSuit.DefaultCtx, service)
			assert.True(t, respSuccess(out))
			assert.Nil(t, out.GetRateLimit())
			t.Logf("pass: out is %+v", out)
		})
		t.Run("服务不存在", func(t *testing.T) {
			_ = discoverSuit.GoverRuleServer().Cache().(*cache.CacheManager).TestUpdate()
			out := discoverSuit.GoverRuleServer().GetRateLimitWithCache(discoverSuit.DefaultCtx, &apiservice.Service{
				Name:      protobuf.NewStringValue("not_exist_service"),
				Namespace: protobuf.NewStringValue("not_exist_namespace"),
			})
			assert.True(t, respSuccess(out))
			t.Logf("pass: out is %+v", out)
		})
	})
}
