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

package service_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

const defaultAliasNs = "Production"

// 创建一个服务别名
func (d *DiscoverTestSuit) createCommonAlias(service *apiservice.Service, alias string, aliasNamespace string, typ apiservice.AliasType) *apiservice.Response {
	req := &apiservice.ServiceAlias{
		Service:        service.Name,
		Namespace:      service.Namespace,
		Alias:          protobuf.NewStringValue(alias),
		AliasNamespace: protobuf.NewStringValue(aliasNamespace),
		Type:           typ,
		Owners:         protobuf.NewStringValue("pole"),
	}
	return d.DiscoverServer().CreateServiceAlias(d.DefaultCtx, req)
}

// 创建别名，并检查
func (d *DiscoverTestSuit) createCommonAliasCheck(
	t *testing.T, service *apiservice.Service, alias string, aliasNamespace string, typ apiservice.AliasType) *apiservice.Response {
	resp := d.createCommonAlias(service, alias, aliasNamespace, typ)
	if !respSuccess(resp) {
		t.Fatalf("error : %s", resp.GetInfo().GetValue())
	}
	return resp
}

// 检查一个服务别名是否是sid
func isSid(alias string) bool {
	items := strings.Split(alias, ":")
	if len(items) != 2 {
		return false
	}

	for _, it := range items {
		if ok, _ := regexp.MatchString("^[0-9]+$", it); !ok {
			return false
		}
	}

	return true
}

// 正常场景测试
func TestCreateServiceAlias(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	_, serviceResp := discoverSuit.createCommonService(t, 123)
	defer discoverSuit.cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())

	t.Run("正常创建非Sid的别名", func(t *testing.T) {
		alias := fmt.Sprintf("alias.%d", time.Now().Unix())
		resp := discoverSuit.createCommonAlias(serviceResp, alias, serviceResp.GetNamespace().GetValue(), apiservice.AliasType_DEFAULT)
		defer discoverSuit.cleanServiceName(alias, serviceResp.GetNamespace().GetValue())
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
		assert.Equal(t, resp.Alias.Alias.Value, alias)
	})

	t.Run("正常创建Sid别名", func(t *testing.T) {
		resp := discoverSuit.createCommonAlias(serviceResp, "", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_CL5SID)
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
		defer discoverSuit.cleanServiceName(resp.Alias.Alias.Value, serviceResp.GetNamespace().GetValue())
		assert.True(t, isSid(resp.Alias.Alias.Value))
		t.Logf("alias sid: %s", resp.Alias.Alias.Value)
	})

	t.Run("使用ctx带上的token可以创建成功", func(t *testing.T) {
		req := &apiservice.ServiceAlias{
			Service:        serviceResp.Name,
			Namespace:      serviceResp.Namespace,
			AliasNamespace: serviceResp.Namespace,
			Type:           apiservice.AliasType_CL5SID,
		}
		ctx := context.WithValue(discoverSuit.DefaultCtx, types.ContextPoleToken,
			serviceResp.GetToken().GetValue())
		resp := discoverSuit.DiscoverServer().CreateServiceAlias(ctx, req)
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
		discoverSuit.cleanServiceName(resp.Alias.Alias.Value, serviceResp.GetNamespace().GetValue())

		// 带上系统token，也可以成功
		ctx = context.WithValue(discoverSuit.DefaultCtx, types.ContextPoleToken,
			"polaris@12345678")
		resp = discoverSuit.DiscoverServer().CreateServiceAlias(ctx, req)
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
		discoverSuit.cleanServiceName(resp.Alias.Alias.Value, serviceResp.GetNamespace().GetValue())
	})
	t.Run("不允许为别名创建别名", func(t *testing.T) {
		resp := discoverSuit.NamespaceServer().CreateNamespace(discoverSuit.DefaultCtx, &apimodel.Namespace{
			Name: &wrapperspb.StringValue{Value: defaultAliasNs},
		})
		if !respSuccess(resp) {
			t.Fatalf("error : %s", resp.GetInfo().GetValue())
		}

		resp = discoverSuit.createCommonAliasCheck(t, serviceResp, "", defaultAliasNs, apiservice.AliasType_CL5SID)
		defer discoverSuit.cleanServiceName(resp.Alias.Alias.Value, serviceResp.Namespace.Value)

		service := &apiservice.Service{
			Name:      resp.Alias.Alias,
			Namespace: serviceResp.Namespace,
			Token:     serviceResp.Token,
		}
		repeatedResp := discoverSuit.createCommonAlias(service, "", defaultAliasNs, apiservice.AliasType_CL5SID)
		if respSuccess(repeatedResp) {
			t.Fatalf("error: %+v", repeatedResp)
		}
		t.Logf("%+v", repeatedResp)
	})
}

// 并发测试
func TestConcurrencyCreateSid(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	_, serviceResp := discoverSuit.createCommonService(t, 234)
	defer discoverSuit.cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())

	t.Run("并发创建sid别名，sid不会重复", func(t *testing.T) {
		c := 20
		var wg sync.WaitGroup
		resultCh := make(chan *apiservice.Response, 1)
		results := make([]*apiservice.Response, 0, 200)
		shutdown := make(chan struct{})

		go func() {
			for {
				select {
				case result := <-resultCh:
					results = append(results, result)
				case <-shutdown:
					t.Log("[Alias] concurrency function exit")
					return
				}
			}
		}()

		for i := 0; i < c; i++ {
			wg.Add(1)
			go func(index int) {
				defer func() {
					t.Logf("[Alias] finish creating alias sid func index(%d)", index)
					wg.Done()
				}()
				resp := discoverSuit.createCommonAlias(
					serviceResp, "", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_CL5SID)
				resultCh <- resp
			}(i)
		}

		wg.Wait()
		time.Sleep(time.Second)
		close(shutdown)

		repeated := make(map[string]bool)
		for i := 0; i < c; i++ {
			resp := results[i]
			assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
			defer discoverSuit.cleanServiceName(resp.Alias.Alias.Value, serviceResp.GetNamespace().GetValue())
			assert.True(t, isSid(resp.Alias.Alias.Value))

			repeated[resp.Alias.Alias.Value] = true
		}
		// 检查是否重复，必须是200个
		assert.Equal(t, len(repeated), c)
	})
}

// 异常测试
func TestExceptCreateAlias(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	_, serviceResp := discoverSuit.createCommonService(t, 345)
	defer discoverSuit.cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())

	t.Run("参数缺失，报错", func(t *testing.T) {
		noService := &apiservice.Service{}
		resp := discoverSuit.createCommonAlias(
			noService, "x1.x2.x3", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_DEFAULT)
		assert.False(t, respSuccess(resp), resp.GetInfo().GetValue())

		noService.Name = protobuf.NewStringValue("123")
		resp = discoverSuit.createCommonAlias(
			noService, "x1.x2.x3", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_DEFAULT)
		assert.False(t, respSuccess(resp), resp.GetInfo().GetValue())

		noService.Namespace = protobuf.NewStringValue("456")
		resp = discoverSuit.createCommonAlias(
			noService, "x1.x2.x3", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_DEFAULT)
		assert.False(t, respSuccess(resp), resp.GetInfo().GetValue())

		noService.Token = protobuf.NewStringValue("567")
		resp = discoverSuit.createCommonAlias(noService, "", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_DEFAULT)
		assert.False(t, respSuccess(resp), resp.GetInfo().GetValue())
		t.Logf("return code: %d", resp.Code.Value)
	})

	t.Run("不存在的源服务，报错", func(t *testing.T) {
		noService := &apiservice.Service{
			Name:      protobuf.NewStringValue("my.service.2020.02.19"),
			Namespace: protobuf.NewStringValue("123123"),
			Token:     protobuf.NewStringValue("aaa"),
		}
		resp := discoverSuit.createCommonAlias(noService, "x1.x2.x3", noService.Namespace.GetValue(), apiservice.AliasType_DEFAULT)
		assert.False(t, respSuccess(resp), resp.GetInfo().GetValue())
		t.Logf("return code: %d", resp.Code.Value)
		assert.Equal(t, resp.GetCode().GetValue(), api.NotFoundService)
	})

	t.Run("同名alias，报错", func(t *testing.T) {
		resp := discoverSuit.createCommonAlias(
			serviceResp, "x1.x2.x3", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_DEFAULT)
		assert.True(t, respSuccess(resp), resp.GetInfo().GetValue())

		defer discoverSuit.cleanServiceName(resp.Alias.Alias.Value, serviceResp.GetNamespace().GetValue())

		resp = discoverSuit.createCommonAlias(
			serviceResp, "x1.x2.x3", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_DEFAULT)
		assert.False(t, respSuccess(resp), resp.GetInfo().GetValue())
		t.Logf("same alias return code: %d", resp.Code.Value)
	})

	t.Run("目标服务已经是一个别名", func(t *testing.T) {
		resp := discoverSuit.createCommonAlias(
			serviceResp, "x1.x2.x3.x4", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_DEFAULT)
		assert.True(t, respSuccess(resp), resp.GetInfo().GetValue())

		defer discoverSuit.cleanServiceName(resp.Alias.Alias.Value, serviceResp.GetNamespace().GetValue())

		resp = discoverSuit.createCommonAlias(
			&apiservice.Service{
				Name:      protobuf.NewStringValue("x1.x2.x3.x4"),
				Namespace: serviceResp.GetNamespace(),
			}, "x1.x2.x3.x5", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_DEFAULT)
		assert.False(t, respSuccess(resp), resp.GetInfo().GetValue())
		assert.Equal(t, apimodel.Code_NotAllowCreateAliasForAlias, apimodel.Code(resp.GetCode().GetValue()))
		t.Logf("same alias return code: %d", resp.Code.Value)
	})

	t.Run("鉴权失败，报错", func(t *testing.T) {
		service := &apiservice.Service{
			Name:      serviceResp.Name,
			Namespace: serviceResp.Namespace,
			Token:     protobuf.NewStringValue("123123123"),
		}

		oldCtx := discoverSuit.DefaultCtx

		discoverSuit.DefaultCtx = context.Background()

		defer func() {
			discoverSuit.DefaultCtx = oldCtx
		}()

		_ = discoverSuit.CacheMgr().TestUpdate()

		resp := discoverSuit.createCommonAlias(service, "x1.x2.x3", service.Namespace.GetValue(), apiservice.AliasType_DEFAULT)
		assert.False(t, respSuccess(resp), resp.GetInfo().GetValue())
		t.Logf("error token, return code: %d", resp.Code.Value)
	})

	t.Run("指向的服务不存在（新接口）", func(t *testing.T) {
		_, serviceResp2 := discoverSuit.createCommonService(t, 2)
		discoverSuit.cleanServiceName(serviceResp2.GetName().GetValue(), serviceResp2.GetNamespace().GetValue())
		resp := discoverSuit.createCommonAlias(serviceResp2, "", serviceResp2.GetNamespace().GetValue(), apiservice.AliasType_CL5SID)
		if respSuccess(resp) {
			t.Fatalf("error: %+v", resp)
		}
		t.Logf("%+v", resp)
	})
}

// 别名修改的测试
func TestUpdateServiceAlias(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	_, serviceResp := discoverSuit.createCommonService(t, 3)
	defer discoverSuit.cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
	t.Run("修改别名负责人", func(t *testing.T) {
		resp := discoverSuit.createCommonAlias(serviceResp, "", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_CL5SID)
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
		defer discoverSuit.cleanServiceName(resp.GetAlias().GetAlias().GetValue(), serviceResp.GetNamespace().GetValue())

		// 修改别名负责人
		req := &apiservice.ServiceAlias{
			Service:        resp.GetAlias().GetService(),
			Namespace:      resp.GetAlias().GetNamespace(),
			Alias:          resp.GetAlias().GetAlias(),
			AliasNamespace: resp.GetAlias().GetNamespace(),
			Owners:         protobuf.NewStringValue("alias-owner-new"),
			ServiceToken:   resp.GetAlias().GetServiceToken(),
		}

		repeatedResp := discoverSuit.DiscoverServer().UpdateServiceAlias(discoverSuit.DefaultCtx, req)
		assert.True(t, api.IsSuccess(repeatedResp), resp.GetInfo().GetValue())

		query := map[string]string{
			"alias":     req.GetAlias().GetValue(),
			"namespace": req.GetNamespace().GetValue(),
		}
		aliasResponse := discoverSuit.DiscoverServer().GetServiceAliases(discoverSuit.DefaultCtx, query)
		// 判断负责人是否一致
		assert.Equal(t, aliasResponse.GetAliases()[0].GetOwners().GetValue(), "alias-owner-new")
		t.Logf("pass, owner is %v", aliasResponse.GetAliases()[0].GetOwners().GetValue())
	})

	t.Run("修改指向服务", func(t *testing.T) {
		resp := discoverSuit.createCommonAlias(serviceResp, "", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_CL5SID)
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
		defer discoverSuit.cleanServiceName(resp.GetAlias().GetAlias().GetValue(), serviceResp.GetNamespace().GetValue())

		// 创建新的服务
		_, serviceResp2 := discoverSuit.createCommonService(t, 4)
		defer discoverSuit.cleanServiceName(serviceResp2.GetName().GetValue(), serviceResp2.GetNamespace().GetValue())

		// 修改别名指向
		req := &apiservice.ServiceAlias{
			Service:        serviceResp2.GetName(),
			Namespace:      serviceResp2.GetNamespace(),
			Alias:          resp.GetAlias().GetAlias(),
			AliasNamespace: serviceResp2.GetNamespace(),
			Owners:         resp.GetAlias().GetOwners(),
			Comment:        resp.GetAlias().GetComment(),
			ServiceToken:   resp.GetAlias().GetServiceToken(),
		}

		repeatedResp := discoverSuit.DiscoverServer().UpdateServiceAlias(discoverSuit.DefaultCtx, req)
		assert.True(t, api.IsSuccess(repeatedResp), resp.GetInfo().GetValue())

		query := map[string]string{
			"alias":     req.GetAlias().GetValue(),
			"namespace": req.GetNamespace().GetValue(),
		}
		aliasResponse := discoverSuit.DiscoverServer().GetServiceAliases(discoverSuit.DefaultCtx, query)
		// 判断指向服务是否一致
		assert.Equal(t, aliasResponse.GetAliases()[0].GetService().GetValue(), serviceResp2.GetName().GetValue())
		t.Logf("pass, service is %v", aliasResponse.GetAliases()[0].GetService().GetValue())
	})

	t.Run("要指向的服务不存在", func(t *testing.T) {
		resp := discoverSuit.createCommonAlias(serviceResp, "", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_CL5SID)
		assert.True(t, respSuccess(resp), resp.GetInfo().GetValue())

		defer discoverSuit.cleanServiceName(resp.GetAlias().GetAlias().GetValue(), serviceResp.GetNamespace().GetValue())

		// 创建新的服务并删除
		_, serviceResp2 := discoverSuit.createCommonService(t, 4)
		discoverSuit.cleanServiceName(serviceResp2.GetName().GetValue(), serviceResp2.GetNamespace().GetValue())

		// 修改别名指向
		req := &apiservice.ServiceAlias{
			Service:        serviceResp2.GetName(),
			Namespace:      serviceResp2.GetNamespace(),
			Alias:          resp.GetAlias().GetAlias(),
			AliasNamespace: resp.GetAlias().GetNamespace(),
			Owners:         resp.GetAlias().GetOwners(),
			Comment:        resp.GetAlias().GetComment(),
			ServiceToken:   resp.GetAlias().GetServiceToken(),
		}
		repeatedResp := discoverSuit.DiscoverServer().UpdateServiceAlias(discoverSuit.DefaultCtx, req)
		if respSuccess(repeatedResp) {
			t.Fatalf("error: %+v", repeatedResp)
		}
		t.Logf("%+v", repeatedResp)
	})

	t.Run("鉴权失败", func(t *testing.T) {
		resp := discoverSuit.createCommonAlias(serviceResp, "", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_CL5SID)
		assert.True(t, respSuccess(resp), resp.GetInfo().GetValue())
		defer discoverSuit.cleanServiceName(resp.GetAlias().GetAlias().GetValue(), serviceResp.GetNamespace().GetValue())

		_ = discoverSuit.CacheMgr().TestUpdate()

		// 修改service token
		req := resp.GetAlias()
		req.ServiceToken = protobuf.NewStringValue("")

		repeatedResp := discoverSuit.DiscoverServer().UpdateServiceAlias(context.Background(), req)

		if respSuccess(repeatedResp) {
			t.Fatalf("error: %+v", repeatedResp)
		}
		t.Logf("%+v", repeatedResp)
	})
}

// 别名删除
func TestDeleteServiceAlias(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	_, serviceResp := discoverSuit.createCommonService(t, 201)
	defer discoverSuit.cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
	t.Run("通过服务别名删除接口可以直接删除别名", func(t *testing.T) {
		resp := discoverSuit.createCommonAlias(serviceResp, serviceResp.Name.GetValue()+"_alias", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_DEFAULT)
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())

		defer discoverSuit.cleanServiceName(resp.Alias.Alias.Value, resp.Alias.AliasNamespace.Value)
		discoverSuit.removeCommonServiceAliases(t, []*apiservice.ServiceAlias{resp.Alias})

		query := map[string]string{"name": resp.Alias.Alias.Value}
		queryResp := discoverSuit.DiscoverServer().GetServices(discoverSuit.DefaultCtx, query)
		assert.True(t, api.IsSuccess(queryResp), queryResp.GetInfo().GetValue())
		assert.Equal(t, len(queryResp.Services), 0)
	})

	t.Run("通过ctx带上token，可以删除别名成功", func(t *testing.T) {
		resp := discoverSuit.createCommonAlias(serviceResp, "", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_CL5SID)
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())

		defer discoverSuit.cleanServiceName(resp.Alias.Alias.Value, serviceResp.Namespace.Value)

		ctx := context.WithValue(discoverSuit.DefaultCtx, types.ContextPoleToken,
			"polaris@12345678")
		batchResp := discoverSuit.DiscoverServer().DeleteServiceAliases(ctx, []*apiservice.ServiceAlias{resp.Alias})
		assert.True(t, api.IsSuccess(batchResp), batchResp.GetInfo().GetValue())
	})

}

// 服务实例与服务路由关联测试
func TestServiceAliasRelated(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	_, serviceResp := discoverSuit.createCommonService(t, 202)
	defer discoverSuit.cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
	resp := discoverSuit.createCommonAlias(serviceResp, "", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_CL5SID)
	if !respSuccess(resp) {
		t.Fatalf("errror")
	}
	defer discoverSuit.cleanServiceName(resp.Alias.Alias.Value, serviceResp.Namespace.Value)
	t.Run("实例新建，不允许为别名新建实例", func(t *testing.T) {
		instance := &apiservice.Instance{
			Service:      resp.Alias.Alias,
			Namespace:    serviceResp.Namespace,
			ServiceToken: serviceResp.Token,
			Host:         protobuf.NewStringValue("1.12.123.132"),
			Port:         protobuf.NewUInt32Value(8080),
		}
		instanceResp := discoverSuit.DiscoverServer().CreateInstances(discoverSuit.DefaultCtx, []*apiservice.Instance{instance})
		assert.False(t, api.IsSuccess(instanceResp), instanceResp.GetInfo().GetValue())

		t.Logf("alias create instance ret code(%d), msg(%s)",
			instanceResp.Code.Value, instanceResp.Info.Value)
	})
	t.Run("实例Discover，别名查询实例，返回源服务的实例信息", func(t *testing.T) {
		_, instanceResp := discoverSuit.createCommonInstance(t, serviceResp, 123)
		defer discoverSuit.cleanInstance(instanceResp.GetId().GetValue())
		_ = discoverSuit.CacheMgr().TestUpdate()
		service := &apiservice.Service{Name: resp.Alias.Alias, Namespace: resp.Alias.Namespace}
		disResp := discoverSuit.DiscoverServer().ServiceInstancesCache(discoverSuit.DefaultCtx, &apiservice.DiscoverFilter{}, service)
		assert.True(t, api.IsSuccess(disResp), disResp.GetInfo().GetValue())
		assert.Equal(t, len(disResp.Instances), 1)
	})
	// Convey("路由Discover，别名查询路由，返回源服务的路由信息", t, func() {
	// 	discoverSuit.createCommonRoutingConfig(t, serviceResp, 1, 0) // in=1, out=0
	// 	defer discoverSuit.cleanCommonRoutingConfig(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())

	// 	time.Sleep(discoverSuit.updateCacheInterval)
	// 	service := &apiservice.Service{Name: resp.Alias.Alias, Namespace: resp.Alias.Namespace}
	// 	disResp := discoverSuit.DiscoverServer().GetRoutingConfigWithCache(discoverSuit.DefaultCtx, service)
	// 	So(respSuccess(disResp), ShouldEqual, true)
	// 	So(len(disResp.Routing.Inbounds), ShouldEqual, 1)
	// 	So(len(disResp.Routing.Outbounds), ShouldEqual, 0)
	// })
}

// 测试获取别名列表
func TestGetServiceAliases(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	_, serviceResp := discoverSuit.createCommonService(t, 203)
	t.Cleanup(func() {
		discoverSuit.cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
		discoverSuit.Destroy()
	})

	var aliases []*apiservice.Response
	count := 5
	for i := 0; i < count; i++ {
		resp := discoverSuit.createCommonAlias(serviceResp, "", serviceResp.GetNamespace().GetValue(), apiservice.AliasType_CL5SID)
		if !respSuccess(resp) {
			t.Fatalf("error: %+v", resp)
		}
		t.Cleanup(func() {
			discoverSuit.cleanServiceName(resp.Alias.Alias.Value, serviceResp.Namespace.Value)
		})
		aliases = append(aliases, resp)
	}

	t.Run("可以查询到全量别名", func(t *testing.T) {
		resp := discoverSuit.DiscoverServer().GetServiceAliases(discoverSuit.DefaultCtx, nil)
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
		assert.True(t, len(resp.Aliases) >= count)
		assert.True(t, int(resp.Amount.Value) >= count)
	})
	t.Run("offset,limit测试", func(t *testing.T) {
		query := map[string]string{"offset": "0", "limit": "100"}
		resp := discoverSuit.DiscoverServer().GetServiceAliases(discoverSuit.DefaultCtx, query)
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
		assert.True(t, len(resp.Aliases) >= count)
		assert.True(t, int(resp.Amount.Value) >= count)

		query["limit"] = "0"
		resp = discoverSuit.DiscoverServer().GetServiceAliases(discoverSuit.DefaultCtx, query)
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
		assert.True(t, len(resp.Aliases) == 0, fmt.Sprintf("actual: %d, expect: %d", len(resp.Aliases), 0))
		assert.True(t, int(resp.Amount.Value) == count, fmt.Sprintf("actual: %d, expect: %d", len(resp.Aliases), count))
	})
	t.Run("不合法的过滤条件", func(t *testing.T) {
		query := map[string]string{"xxx": "1", "limit": "100"}
		resp := discoverSuit.DiscoverServer().GetServiceAliases(discoverSuit.DefaultCtx, query)
		assert.False(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
	})
	t.Run("过滤条件可以生效", func(t *testing.T) {
		query := map[string]string{
			"alias":     aliases[2].Alias.Alias.Value,
			"service":   serviceResp.Name.Value,
			"namespace": serviceResp.Namespace.Value,
		}
		resp := discoverSuit.DiscoverServer().GetServiceAliases(discoverSuit.DefaultCtx, query)
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
		assert.True(t, len(resp.Aliases) == 1)
		assert.True(t, int(resp.Amount.Value) == 1)
	})
	t.Run("找不到别名", func(t *testing.T) {
		query := map[string]string{"alias": "x1.1.x2.x3"}
		resp := discoverSuit.DiscoverServer().GetServiceAliases(discoverSuit.DefaultCtx, query)
		assert.True(t, api.IsSuccess(resp), resp.GetInfo().GetValue())
		assert.True(t, len(resp.Aliases) == 0)
		assert.True(t, int(resp.Amount.Value) == 0)
	})
	// Convey("支持owner过滤", t, func() {
	// 	query := map[string]string{"owner": "service-owner-203"}
	// 	resp := discoverSuit.DiscoverServer().GetServiceAliases(discoverSuit.DefaultCtx, query)
	// 	So(respSuccess(resp), ShouldEqual, true)
	// 	So(len(resp.Aliases), ShouldEqual, count)
	// 	So(resp.Amount.Value, ShouldEqual, count)
	// })
}

// test对serviceAlias字段进行校验
func TestCheckServiceAliasFieldLen(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	serviceAlias := &apiservice.ServiceAlias{
		Service:        protobuf.NewStringValue("test-123"),
		Namespace:      protobuf.NewStringValue("Production"),
		Alias:          protobuf.NewStringValue("0"),
		AliasNamespace: protobuf.NewStringValue("Production"),
		Type:           apiservice.AliasType_DEFAULT,
		Owners:         protobuf.NewStringValue("alias-owner"),
		Comment:        protobuf.NewStringValue("comment"),
	}
	t.Run("服务名超长", func(t *testing.T) {
		str := genSpecialStr(129)
		oldService := serviceAlias.Service
		serviceAlias.Service = protobuf.NewStringValue(str)
		resp := discoverSuit.DiscoverServer().CreateServiceAlias(discoverSuit.DefaultCtx, serviceAlias)
		serviceAlias.Service = oldService
		if resp.Code.Value != api.InvalidServiceName {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("命名空间超长", func(t *testing.T) {
		str := genSpecialStr(129)
		oldNamespace := serviceAlias.Namespace
		serviceAlias.Namespace = protobuf.NewStringValue(str)
		resp := discoverSuit.DiscoverServer().CreateServiceAlias(discoverSuit.DefaultCtx, serviceAlias)
		serviceAlias.Namespace = oldNamespace
		if resp.Code.Value != api.InvalidNamespaceName {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("别名超长", func(t *testing.T) {
		str := genSpecialStr(129)
		oldAlias := serviceAlias.Alias
		serviceAlias.Alias = protobuf.NewStringValue(str)
		resp := discoverSuit.DiscoverServer().CreateServiceAlias(discoverSuit.DefaultCtx, serviceAlias)
		serviceAlias.Alias = oldAlias
		if resp.Code.Value != api.InvalidServiceAlias {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("服务别名comment超长", func(t *testing.T) {
		str := genSpecialStr(1025)
		oldComment := serviceAlias.Comment
		serviceAlias.Comment = protobuf.NewStringValue(str)
		resp := discoverSuit.DiscoverServer().CreateServiceAlias(discoverSuit.DefaultCtx, serviceAlias)
		serviceAlias.Comment = oldComment
		if resp.Code.Value != api.InvalidServiceAliasComment {
			t.Fatalf("%+v", resp)
		}
	})
	// t.Run("服务owner超长", func(t *testing.T) {
	// 	str := genSpecialStr(1025)
	// 	oldOwner := serviceAlias.Owners
	// 	serviceAlias.Owners = protobuf.NewStringValue(str)
	// 	resp := discoverSuit.DiscoverServer().CreateServiceAlias(discoverSuit.DefaultCtx, serviceAlias)
	// 	serviceAlias.Owners = oldOwner
	// 	if resp.Code.Value != api.InvalidServiceAliasOwners {
	// 		t.Fatalf("%+v", resp)
	// 	}
	// })
}

// test测试别名的命名空间与服务名不一样
func TestServiceAliasDifferentNamespace(t *testing.T) {

	discoverSuit := &DiscoverTestSuit{}
	if err := discoverSuit.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer discoverSuit.Destroy()

	_, serviceResp := discoverSuit.createCommonService(t, 203)
	defer discoverSuit.cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
	t.Run("正常创建不一样命名空间的非Sid的别名", func(t *testing.T) {
		alias := fmt.Sprintf("alias.%d", time.Now().Unix())
		resp := discoverSuit.createCommonAlias(serviceResp, alias, defaultAliasNs, apiservice.AliasType_DEFAULT)
		defer discoverSuit.cleanServiceName(alias, defaultAliasNs)
		assert.True(t, respSuccess(resp), resp.GetInfo().GetValue())
		assert.Equal(t, resp.Alias.Alias.Value, alias)
	})
}
