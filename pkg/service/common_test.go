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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	bolt "go.etcd.io/bbolt"

	apifault "github.com/polarismesh/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	_ "github.com/pole-io/pole-server/pkg/cache"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/service"
	sqldb "github.com/pole-io/pole-server/plugin/store/mysql"
	testsuit "github.com/pole-io/pole-server/test/suit"
)

const (
	tblNameNamespace          = "namespace"
	tblNameInstance           = "instance"
	tblNameService            = "service"
	tblNameRouting            = "routing"
	tblRateLimitConfig        = "ratelimit_config"
	tblRateLimitRevision      = "ratelimit_revision"
	tblCircuitBreaker         = "circuitbreaker_rule"
	tblCircuitBreakerRelation = "circuitbreaker_rule_relation"
	tblNameL5                 = "l5"
	tblNameRoutingV2          = "routing_config_v2"
	tblClient                 = "client"
)

type DiscoverTestSuit struct {
	testsuit.DiscoverTestSuit
}

// 从数据库彻底删除服务名对应的服务
func (d *DiscoverTestSuit) cleanServiceName(name string, namespace string) {
	// log.Infof("clean service %s, %s", name, namespace)
	d.GetTestDataClean().CleanService(name, namespace)
}

// 从数据库彻底删除实例
func (d *DiscoverTestSuit) cleanInstance(instanceID string) {
	d.GetTestDataClean().CleanInstance(instanceID)
}

// 增加一个服务
func (d *DiscoverTestSuit) createCommonService(t *testing.T, id int) (*apiservice.Service, *apiservice.Service) {
	serviceReq := genMainService(id)
	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("key-%d-%d", id, i)
		v := fmt.Sprintf("value-%d-%d", id, i)
		serviceReq.Metadata[k] = v
	}

	d.cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())

	resp := d.DiscoverServer().CreateServices(d.DefaultCtx, []*apiservice.Service{serviceReq})
	if !respSuccess(resp) {
		t.Fatalf("error: %s", resp.GetInfo().GetValue())
	}

	return serviceReq, resp.Responses[0].GetService()
}

func (d *DiscoverTestSuit) HeartBeat(t *testing.T, service *apiservice.Service, instanceID string) {
	req := &apiservice.Instance{
		ServiceToken: protobuf.NewStringValue(service.GetToken().GetValue()),
		Id:           protobuf.NewStringValue(instanceID),
	}

	resp := d.HealthCheckServer().Report(d.DefaultCtx, req)
	if !respSuccess(resp) {
		t.Fatalf("error: %s", resp.GetInfo().GetValue())
	}
}

func (d *DiscoverTestSuit) GetLastHeartBeat(t *testing.T, service *apiservice.Service,
	instanceID string) *apiservice.Response {
	req := &apiservice.Instance{
		ServiceToken: protobuf.NewStringValue(service.GetToken().GetValue()),
		Id:           protobuf.NewStringValue(instanceID),
	}

	return d.HealthCheckServer().GetLastHeartbeat(req)
}

// 生成服务的主要数据
func genMainService(id int) *apiservice.Service {
	return &apiservice.Service{
		Name:       protobuf.NewStringValue(fmt.Sprintf("test-service-%d", id)),
		Namespace:  protobuf.NewStringValue(service.DefaultNamespace),
		Metadata:   make(map[string]string),
		Ports:      protobuf.NewStringValue(fmt.Sprintf("ports-%d", id)),
		Business:   protobuf.NewStringValue(fmt.Sprintf("business-%d", id)),
		Department: protobuf.NewStringValue(fmt.Sprintf("department-%d", id)),
		CmdbMod1:   protobuf.NewStringValue(fmt.Sprintf("cmdb-mod1-%d", id)),
		CmdbMod2:   protobuf.NewStringValue(fmt.Sprintf("cmdb-mod2-%d", id)),
		CmdbMod3:   protobuf.NewStringValue(fmt.Sprintf("cmdb-mod2-%d", id)),
		Comment:    protobuf.NewStringValue(fmt.Sprintf("service-comment-%d", id)),
		Owners:     protobuf.NewStringValue(fmt.Sprintf("service-owner-%d", id)),
	}
}

// removeCommonService
func (d *DiscoverTestSuit) removeCommonServices(t *testing.T, req []*apiservice.Service) {
	if resp := d.DiscoverServer().DeleteServices(d.DefaultCtx, req); !respSuccess(resp) {
		t.Fatalf("error: %s", resp.GetInfo().GetValue())
	}
}

// removeCommonService
func (d *DiscoverTestSuit) removeCommonServiceAliases(t *testing.T, req []*apiservice.ServiceAlias) {
	if resp := d.DiscoverServer().DeleteServiceAliases(d.DefaultCtx, req); !respSuccess(resp) {
		t.Fatalf("error: %s", resp.GetInfo().GetValue())
	}
}

// 新增一个实例ById
func (d *DiscoverTestSuit) createCommonInstanceById(t *testing.T, svc *apiservice.Service, count int, instanceID string) (
	*apiservice.Instance, *apiservice.Instance) {
	instanceReq := &apiservice.Instance{
		ServiceToken: protobuf.NewStringValue(svc.GetToken().GetValue()),
		Service:      protobuf.NewStringValue(svc.GetName().GetValue()),
		Namespace:    protobuf.NewStringValue(svc.GetNamespace().GetValue()),
		VpcId:        protobuf.NewStringValue(fmt.Sprintf("vpcid-%d", count)),
		Host:         protobuf.NewStringValue(fmt.Sprintf("9.9.9.%d", count)),
		Port:         protobuf.NewUInt32Value(8000 + uint32(count)),
		Protocol:     protobuf.NewStringValue(fmt.Sprintf("protocol-%d", count)),
		Version:      protobuf.NewStringValue(fmt.Sprintf("version-%d", count)),
		Priority:     protobuf.NewUInt32Value(1 + uint32(count)%10),
		Weight:       protobuf.NewUInt32Value(1 + uint32(count)%1000),
		HealthCheck: &apiservice.HealthCheck{
			Type: apiservice.HealthCheck_HEARTBEAT,
			Heartbeat: &apiservice.HeartbeatHealthCheck{
				Ttl: protobuf.NewUInt32Value(3),
			},
		},
		Healthy:  protobuf.NewBoolValue(false), // 默认是非健康，因为打开了healthCheck
		Isolate:  protobuf.NewBoolValue(false),
		LogicSet: protobuf.NewStringValue(fmt.Sprintf("logic-set-%d", count)),
		Metadata: map[string]string{
			"internal-personal-xxx":        fmt.Sprintf("internal-personal-xxx_%d", count),
			"2my-meta":                     fmt.Sprintf("my-meta-%d", count),
			"my-meta-a1":                   "1111",
			"smy-xmeta-h2":                 "2222",
			"my-1meta-o3":                  "2222",
			"my-2meta-4c":                  "2222",
			"my-3meta-d5":                  "2222",
			"dmy-meta-6p":                  "2222",
			"1my-pmeta-d7":                 "2222",
			"my-dmeta-8c":                  "2222",
			"my-xmeta-9p":                  "2222",
			"other-meta-x":                 "xxx",
			"other-meta-1":                 "xx11",
			"amy-instance":                 "my-instance",
			"very-long-key-data-xxxxxxxxx": "Y",
			"very-long-key-data-uuuuuuuuu": "P",
		},
	}
	if len(instanceID) > 0 {
		instanceReq.Id = protobuf.NewStringValue(instanceID)
	}

	resp := d.DiscoverServer().CreateInstances(d.DefaultCtx, []*apiservice.Instance{instanceReq})
	if respSuccess(resp) {
		return instanceReq, resp.Responses[0].GetInstance()
	}

	if resp.GetCode().GetValue() != api.ExistedResource {
		t.Fatalf("error: %s", resp.GetInfo().GetValue())
	}

	if len(instanceID) == 0 {
		instanceID, _ = utils.CalculateInstanceID(
			instanceReq.GetNamespace().GetValue(), instanceReq.GetService().GetValue(),
			instanceReq.GetVpcId().GetValue(), instanceReq.GetHost().GetValue(), instanceReq.GetPort().GetValue())
	}
	// repeated
	d.cleanInstance(instanceID)
	t.Logf("repeatd create instance(%s)", instanceID)
	resp = d.DiscoverServer().CreateInstances(d.DefaultCtx, []*apiservice.Instance{instanceReq})
	if !respSuccess(resp) {
		t.Fatalf("error: %s", resp.GetInfo().GetValue())
	}

	return instanceReq, resp.Responses[0].GetInstance()
}

// 新增一个实例
func (d *DiscoverTestSuit) createCommonInstance(t *testing.T, svc *apiservice.Service, count int) (
	*apiservice.Instance, *apiservice.Instance) {
	return d.createCommonInstanceById(t, svc, count, "")
}

// 指定 IP 和端口为一个服务创建实例
func (d *DiscoverTestSuit) addHostPortInstance(t *testing.T, service *apiservice.Service, host string, port uint32) (
	*apiservice.Instance, *apiservice.Instance) {
	instanceReq := &apiservice.Instance{
		ServiceToken: protobuf.NewStringValue(service.GetToken().GetValue()),
		Service:      protobuf.NewStringValue(service.GetName().GetValue()),
		Namespace:    protobuf.NewStringValue(service.GetNamespace().GetValue()),
		Host:         protobuf.NewStringValue(host),
		Port:         protobuf.NewUInt32Value(port),
		Healthy:      protobuf.NewBoolValue(true),
		Isolate:      protobuf.NewBoolValue(false),
	}
	resp := d.DiscoverServer().CreateInstances(d.DefaultCtx, []*apiservice.Instance{instanceReq})
	if respSuccess(resp) {
		return instanceReq, resp.Responses[0].GetInstance()
	}

	if resp.GetCode().GetValue() != api.ExistedResource {
		t.Fatalf("error: %s", resp.GetInfo().GetValue())
	}
	return instanceReq, resp.Responses[0].GetInstance()
}

// 添加一个实例
func (d *DiscoverTestSuit) addInstance(t *testing.T, ins *apiservice.Instance) (
	*apiservice.Instance, *apiservice.Instance) {
	resp := d.DiscoverServer().CreateInstances(d.DefaultCtx, []*apiservice.Instance{ins})
	if !respSuccess(resp) {
		if resp.GetCode().GetValue() == api.ExistedResource {
			id, _ := utils.CalculateInstanceID(ins.GetNamespace().GetValue(), ins.GetService().GetValue(),
				ins.GetHost().GetValue(), ins.GetHost().GetValue(), ins.GetPort().GetValue())
			d.cleanInstance(id)
		}
	} else {
		return ins, resp.Responses[0].GetInstance()
	}

	resp = d.DiscoverServer().CreateInstances(d.DefaultCtx, []*apiservice.Instance{ins})
	if !respSuccess(resp) {
		t.Fatalf("error: %s", resp.GetInfo().GetValue())
	}

	return ins, resp.Responses[0].GetInstance()
}

// 删除一个实例
func (d *DiscoverTestSuit) removeCommonInstance(t *testing.T, service *apiservice.Service, instanceID string) {
	req := &apiservice.Instance{
		ServiceToken: protobuf.NewStringValue(service.GetToken().GetValue()),
		Id:           protobuf.NewStringValue(instanceID),
	}

	resp := d.DiscoverServer().DeleteInstances(d.DefaultCtx, []*apiservice.Instance{req})
	if !respSuccess(resp) {
		t.Fatalf("error: %s", resp.GetInfo().GetValue())
	}

}

// 通过四元组或者五元组删除实例
func (d *DiscoverTestSuit) removeInstanceWithAttrs(
	t *testing.T, service *apiservice.Service, instance *apiservice.Instance) {
	req := &apiservice.Instance{
		ServiceToken: protobuf.NewStringValue(service.GetToken().GetValue()),
		Service:      protobuf.NewStringValue(service.GetName().GetValue()),
		Namespace:    protobuf.NewStringValue(service.GetNamespace().GetValue()),
		VpcId:        protobuf.NewStringValue(instance.GetVpcId().GetValue()),
		Host:         protobuf.NewStringValue(instance.GetHost().GetValue()),
		Port:         protobuf.NewUInt32Value(instance.GetPort().GetValue()),
	}
	if resp := d.DiscoverServer().DeleteInstances(d.DefaultCtx, []*apiservice.Instance{req}); !respSuccess(resp) {
		t.Fatalf("error: %s", resp.GetInfo().GetValue())
	}
}

func mockRoutingV1(serviceName, serviceNamespace string, inCount int) *apitraffic.Routing {
	inBounds := make([]*apitraffic.Route, 0, inCount)
	for i := 0; i < inCount; i++ {
		matchString := &apimodel.MatchString{
			Type:  apimodel.MatchString_EXACT,
			Value: protobuf.NewStringValue(fmt.Sprintf("in-meta-value-%d", i)),
		}
		source := &apitraffic.Source{
			Service:   protobuf.NewStringValue(fmt.Sprintf("in-source-service-%d", i)),
			Namespace: protobuf.NewStringValue(fmt.Sprintf("in-source-service-%d", i)),
			Metadata: map[string]*apimodel.MatchString{
				fmt.Sprintf("in-metadata-%d", i): matchString,
			},
		}
		destination := &apitraffic.Destination{
			Service:   protobuf.NewStringValue(serviceName),
			Namespace: protobuf.NewStringValue(serviceNamespace),
			Metadata: map[string]*apimodel.MatchString{
				fmt.Sprintf("in-metadata-%d", i): matchString,
			},
			Priority: protobuf.NewUInt32Value(120),
			Weight:   protobuf.NewUInt32Value(100),
			Transfer: protobuf.NewStringValue("abcdefg"),
		}

		entry := &apitraffic.Route{
			Sources:      []*apitraffic.Source{source},
			Destinations: []*apitraffic.Destination{destination},
		}
		inBounds = append(inBounds, entry)
	}

	conf := &apitraffic.Routing{
		Service:   protobuf.NewStringValue(serviceName),
		Namespace: protobuf.NewStringValue(serviceNamespace),
		Inbounds:  inBounds,
	}

	return conf
}

// 创建一个路由配置
func (d *DiscoverTestSuit) createCommonRoutingConfigV2(t *testing.T, cnt int32) []*apitraffic.RouteRule {
	rules := testsuit.MockRoutingV2(t, cnt)

	return d.createCommonRoutingConfigV2WithReq(t, rules)
}

// 创建一个路由配置
func (d *DiscoverTestSuit) createCommonRoutingConfigV2WithReq(
	t *testing.T, rules []*apitraffic.RouteRule) []*apitraffic.RouteRule {
	resp := d.DiscoverServer().CreateRoutingConfigs(d.DefaultCtx, rules)
	if !respSuccess(resp) {
		t.Fatalf("error: %+v", resp)
	}

	if len(rules) != len(resp.GetResponses()) {
		t.Fatal("error: create v2 routings not equal resp")
	}

	ret := []*apitraffic.RouteRule{}
	for i := range resp.GetResponses() {
		item := resp.GetResponses()[i]
		msg := &apitraffic.RouteRule{}

		if err := ptypes.UnmarshalAny(item.GetData(), msg); err != nil {
			t.Fatal(err)
			return nil
		}

		ret = append(ret, msg)
	}

	return ret
}

// 删除一个路由配置
func (d *DiscoverTestSuit) deleteCommonRoutingConfigV2(t *testing.T, req *apitraffic.RouteRule) {
	resp := d.DiscoverServer().DeleteRoutingConfigs(d.DefaultCtx, []*apitraffic.RouteRule{req})
	if !respSuccess(resp) {
		t.Fatalf("%s", resp.GetInfo())
	}
}

// 彻底删除一个路由配置
func (d *DiscoverTestSuit) cleanCommonRoutingConfig(service string, namespace string) {
	d.GetTestDataClean().CleanCommonRoutingConfig(service, namespace)
}

func (d *DiscoverTestSuit) truncateCommonRoutingConfigV2() {
	d.GetTestDataClean().TruncateCommonRoutingConfigV2()
}

// 彻底删除一个路由配置
func (d *DiscoverTestSuit) cleanCommonRoutingConfigV2(rules []*apitraffic.RouteRule) {
	d.GetTestDataClean().CleanCommonRoutingConfigV2(rules)
}

func (d *DiscoverTestSuit) CheckGetService(
	t *testing.T, expectReqs []*apiservice.Service, actualReqs []*apiservice.Service) {
	if len(expectReqs) != len(actualReqs) {
		t.Fatalf("error: %d %d", len(expectReqs), len(actualReqs))
	}

	for _, expect := range expectReqs {
		found := false
		for _, actual := range actualReqs {
			if expect.GetName().GetValue() != actual.GetName().GetValue() ||
				expect.GetNamespace().GetValue() != actual.GetNamespace().GetValue() {
				continue
			}

			found = true

			if expect.GetPorts().GetValue() != actual.GetPorts().GetValue() ||
				expect.GetOwners().GetValue() != actual.GetOwners().GetValue() ||
				expect.GetComment().GetValue() != actual.GetComment().GetValue() ||
				actual.GetToken().GetValue() != "" || actual.GetRevision().GetValue() == "" {
				t.Fatalf("error: %+v, %+v", expect, actual)
			}

			if len(expect.Metadata) != len(actual.Metadata) {
				t.Fatalf("error: %d, %d", len(expect.Metadata), len(actual.Metadata))
			}
			for key, value := range expect.Metadata {
				match, ok := actual.Metadata[key]
				if !ok {
					t.Fatalf("error")
				}
				if value != match {
					t.Fatalf("error")
				}
			}
		}
		if !found {
			t.Fatalf("error: %s, %s", expect.GetName().GetValue(), expect.GetNamespace().GetValue())
		}

	}
}

// 检查服务发现的字段是否一致
func (d *DiscoverTestSuit) discoveryCheck(t *testing.T, req *apiservice.Service, resp *apiservice.DiscoverResponse) {
	if resp == nil {
		t.Fatalf("error")
	}

	if resp.GetService().GetName().GetValue() != req.GetName().GetValue() ||
		resp.GetService().GetNamespace().GetValue() != req.GetNamespace().GetValue() ||
		resp.GetService().GetRevision().GetValue() == "" {
		t.Fatalf("error: %+v", resp)
	}

	if resp.Service == nil {
		t.Fatalf("error")
	}
	// t.Logf("%+v", resp.Service)

	if resp.Service.GetName().GetValue() != req.GetName().GetValue() ||
		resp.Service.GetNamespace().GetValue() != req.GetNamespace().GetValue() {
		t.Fatalf("error: %+v", resp.Service)
	}
}

// 实例校验
func instanceCheck(t *testing.T, expect *apiservice.Instance, actual *apiservice.Instance) {
	// #lizard forgives
	switch {
	case expect.GetService().GetValue() != actual.GetService().GetValue():
		t.Fatalf("error %s---%s", expect.GetService().GetValue(), actual.GetService().GetValue())
	case expect.GetNamespace().GetValue() != actual.GetNamespace().GetValue():
		t.Fatalf("error")
	case expect.GetPort().GetValue() != actual.GetPort().GetValue():
		t.Fatalf("error")
	case expect.GetHost().GetValue() != actual.GetHost().GetValue():
		t.Fatalf("error")
	case expect.GetVpcId().GetValue() != actual.GetVpcId().GetValue():
		t.Fatalf("error")
	case expect.GetProtocol().GetValue() != actual.GetProtocol().GetValue():
		t.Fatalf("error")
	case expect.GetVersion().GetValue() != actual.GetVersion().GetValue():
		t.Fatalf("error")
	case expect.GetWeight().GetValue() != actual.GetWeight().GetValue():
		t.Fatalf("error")
	case expect.GetHealthy().GetValue() != actual.GetHealthy().GetValue():
		t.Fatalf("error")
	case expect.GetIsolate().GetValue() != actual.GetIsolate().GetValue():
		t.Fatalf("error")
	case expect.GetLogicSet().GetValue() != actual.GetLogicSet().GetValue():
		t.Fatalf("error")
	default:
		break

		// 实例创建，无法指定cmdb信息
		/*case expect.GetCmdbRegion().GetValue() != actual.GetCmdbRegion().GetValue():
		  	t.Fatalf("error")
		  case expect.GetCmdbCampus().GetValue() != actual.GetCmdbRegion().GetValue():
		  	t.Fatalf("error")
		  case expect.GetCmdbZone().GetValue() != actual.GetCmdbZone().GetValue():
		  	t.Fatalf("error")*/

	}
	for key, value := range expect.GetMetadata() {
		actualValue := actual.GetMetadata()[key]
		if value != actualValue {
			t.Fatalf("error %+v, %+v", expect.Metadata, actual.Metadata)
		}
	}

	if expect.GetHealthCheck().GetType() != actual.GetHealthCheck().GetType() {
		t.Fatalf("error")
	}
	if expect.GetHealthCheck().GetHeartbeat().GetTtl().GetValue() !=
		actual.GetHealthCheck().GetHeartbeat().GetTtl().GetValue() {
		t.Fatalf("error")
	}
}

// 完整对比service的各个属性
func serviceCheck(t *testing.T, expect *apiservice.Service, actual *apiservice.Service) {
	switch {
	case expect.GetName().GetValue() != actual.GetName().GetValue():
		t.Fatalf("error")
	case expect.GetNamespace().GetValue() != actual.GetNamespace().GetValue():
		t.Fatalf("error")
	case expect.GetPorts().GetValue() != actual.GetPorts().GetValue():
		t.Fatalf("error")
	case expect.GetBusiness().GetValue() != actual.GetBusiness().GetValue():
		t.Fatalf("error")
	case expect.GetDepartment().GetValue() != actual.GetDepartment().GetValue():
		t.Fatalf("error")
	case expect.GetCmdbMod1().GetValue() != actual.GetCmdbMod1().GetValue():
		t.Fatalf("error")
	case expect.GetCmdbMod2().GetValue() != actual.GetCmdbMod2().GetValue():
		t.Fatalf("error")
	case expect.GetCmdbMod3().GetValue() != actual.GetCmdbMod3().GetValue():
		t.Fatalf("error")
	case expect.GetComment().GetValue() != actual.GetComment().GetValue():
		t.Fatalf("error")
	case expect.GetOwners().GetValue() != actual.GetOwners().GetValue():
		t.Fatalf("error")
	default:
		break
	}

	for key, value := range expect.GetMetadata() {
		actualValue := actual.GetMetadata()[key]
		if actualValue != value {
			t.Fatalf("error")
		}
	}
}

// 创建限流规则
func (d *DiscoverTestSuit) createCommonRateLimit(
	t *testing.T, service *apiservice.Service, index int) (*apitraffic.Rule, *apitraffic.Rule) {
	// 先不考虑Cluster
	rateLimit := &apitraffic.Rule{
		Name:      &wrappers.StringValue{Value: fmt.Sprintf("rule_name_%d", index)},
		Service:   service.GetName(),
		Namespace: service.GetNamespace(),
		Priority:  protobuf.NewUInt32Value(uint32(index)),
		Resource:  apitraffic.Rule_QPS,
		Type:      apitraffic.Rule_GLOBAL,
		Arguments: []*apitraffic.MatchArgument{
			{
				Type: apitraffic.MatchArgument_CUSTOM,
				Key:  fmt.Sprintf("name-%d", index),
				Value: &apimodel.MatchString{
					Type:  apimodel.MatchString_EXACT,
					Value: protobuf.NewStringValue(fmt.Sprintf("value-%d", index)),
				},
			},
			{
				Type: apitraffic.MatchArgument_CUSTOM,
				Key:  fmt.Sprintf("name-%d", index+1),
				Value: &apimodel.MatchString{
					Type:  apimodel.MatchString_EXACT,
					Value: protobuf.NewStringValue(fmt.Sprintf("value-%d", index+1)),
				},
			},
		},
		Amounts: []*apitraffic.Amount{
			{
				MaxAmount: protobuf.NewUInt32Value(uint32(10 * index)),
				ValidDuration: &duration.Duration{
					Seconds: int64(index),
					Nanos:   int32(index),
				},
			},
		},
		Action:  protobuf.NewStringValue(fmt.Sprintf("behavior-%d", index)),
		Disable: protobuf.NewBoolValue(false),
		Report: &apitraffic.Report{
			Interval: &duration.Duration{
				Seconds: int64(index),
			},
			AmountPercent: protobuf.NewUInt32Value(uint32(index)),
		},
	}

	resp := d.DiscoverServer().CreateRateLimits(d.DefaultCtx, []*apitraffic.Rule{rateLimit})
	if !respSuccess(resp) {
		t.Fatalf("error: %+v", resp)
	}
	return rateLimit, resp.Responses[0].GetRateLimit()
}

// 删除限流规则
func (d *DiscoverTestSuit) deleteRateLimit(t *testing.T, rateLimit *apitraffic.Rule) {
	if resp := d.DiscoverServer().DeleteRateLimits(d.DefaultCtx, []*apitraffic.Rule{rateLimit}); !respSuccess(resp) {
		t.Fatalf("%s", resp.GetInfo().GetValue())
	}
}

// 更新单个限流规则
func (d *DiscoverTestSuit) updateRateLimit(t *testing.T, rateLimit *apitraffic.Rule) {
	if resp := d.DiscoverServer().UpdateRateLimits(d.DefaultCtx, []*apitraffic.Rule{rateLimit}); !respSuccess(resp) {
		t.Fatalf("%s", resp.GetInfo().GetValue())
	}
}

// 彻底删除限流规则
func (d *DiscoverTestSuit) cleanRateLimit(id string) {
	d.GetTestDataClean().CleanRateLimit(id)
}

// 彻底删除限流规则版本号
func (d *DiscoverTestSuit) cleanRateLimitRevision(service, namespace string) {
}

// 更新限流规则内容
func updateRateLimitContent(rateLimit *apitraffic.Rule, index int) {
	rateLimit.Priority = protobuf.NewUInt32Value(uint32(index))
	rateLimit.Resource = apitraffic.Rule_CONCURRENCY
	rateLimit.Type = apitraffic.Rule_LOCAL
	rateLimit.Labels = map[string]*apimodel.MatchString{
		fmt.Sprintf("name-%d", index): {
			Type:  apimodel.MatchString_EXACT,
			Value: protobuf.NewStringValue(fmt.Sprintf("value-%d", index)),
		},
		fmt.Sprintf("name-%d", index+1): {
			Type:  apimodel.MatchString_REGEX,
			Value: protobuf.NewStringValue(fmt.Sprintf("value-%d", index+1)),
		},
	}
	rateLimit.Amounts = []*apitraffic.Amount{
		{
			MaxAmount: protobuf.NewUInt32Value(uint32(index)),
			ValidDuration: &duration.Duration{
				Seconds: int64(index),
			},
		},
	}
	rateLimit.Action = protobuf.NewStringValue(fmt.Sprintf("value-%d", index))
	rateLimit.Disable = protobuf.NewBoolValue(true)
	rateLimit.Report = &apitraffic.Report{
		Interval: &duration.Duration{
			Seconds: int64(index),
		},
		AmountPercent: protobuf.NewUInt32Value(uint32(index)),
	}
}

/*
 * @brief 对比限流规则的各个属性
 */
func checkRateLimit(t *testing.T, expect *apitraffic.Rule, actual *apitraffic.Rule) {
	switch {
	case expect.GetId().GetValue() != actual.GetId().GetValue():
		t.Fatalf("error id, expect %s, actual %s", expect.GetId().GetValue(), actual.GetId().GetValue())
	case expect.GetService().GetValue() != actual.GetService().GetValue():
		t.Fatalf(
			"error service, expect %s, actual %s",
			expect.GetService().GetValue(), actual.GetService().GetValue())
	case expect.GetNamespace().GetValue() != actual.GetNamespace().GetValue():
		t.Fatalf("error namespace, expect %s, actual %s",
			expect.GetNamespace().GetValue(), actual.GetNamespace().GetValue())
	case expect.GetPriority().GetValue() != actual.GetPriority().GetValue():
		t.Fatalf("error priority, expect %v, actual %v",
			expect.GetPriority().GetValue(), actual.GetPriority().GetValue())
	case expect.GetResource() != actual.GetResource():
		t.Fatalf("error resource, expect %v, actual %v", expect.GetResource(), actual.GetResource())
	case expect.GetType() != actual.GetType():
		t.Fatalf("error type, expect %v, actual %v", expect.GetType(), actual.GetType())
	case expect.GetDisable().GetValue() != actual.GetDisable().GetValue():
		t.Fatalf("error disable, expect %v, actual %v",
			expect.GetDisable().GetValue(), actual.GetDisable().GetValue())
	case expect.GetAction().GetValue() != actual.GetAction().GetValue():
		t.Fatalf("error action, expect %s, actual %s",
			expect.GetAction().GetValue(), actual.GetAction().GetValue())
	default:
		break
	}

	expectSubset, err := json.Marshal(expect.GetSubset())
	if err != nil {
		panic(err)
	}
	actualSubset, err := json.Marshal(actual.GetSubset())
	if err != nil {
		panic(err)
	}
	if string(expectSubset) != string(actualSubset) {
		t.Fatal("error subset")
	}

	expectLabels, err := json.Marshal(expect.GetArguments())
	if err != nil {
		panic(err)
	}
	actualLabels, err := json.Marshal(actual.GetArguments())
	if err != nil {
		panic(err)
	}
	if string(expectLabels) != string(actualLabels) {
		t.Fatal("error labels")
	}

	expectAmounts, err := json.Marshal(expect.GetAmounts())
	if err != nil {
		panic(err)
	}
	actualAmounts, err := json.Marshal(actual.GetAmounts())
	if err != nil {
		panic(err)
	}
	if string(expectAmounts) != string(actualAmounts) {
		t.Fatal("error amounts")
	}
}

// 对比熔断规则的各个属性
func checkCircuitBreaker(
	t *testing.T, expect, expectMaster *apifault.CircuitBreaker, actual *apifault.CircuitBreaker) {
	switch {
	case expectMaster.GetId().GetValue() != actual.GetId().GetValue():
		t.Fatal("error id")
	case expect.GetVersion().GetValue() != actual.GetVersion().GetValue():
		t.Fatal("error version")
	case expectMaster.GetName().GetValue() != actual.GetName().GetValue():
		t.Fatal("error name")
	case expectMaster.GetNamespace().GetValue() != actual.GetNamespace().GetValue():
		t.Fatal("error namespace")
	case expectMaster.GetOwners().GetValue() != actual.GetOwners().GetValue():
		t.Fatal("error owners")
	case expectMaster.GetComment().GetValue() != actual.GetComment().GetValue():
		t.Fatal("error comment")
	case expectMaster.GetBusiness().GetValue() != actual.GetBusiness().GetValue():
		t.Fatal("error business")
	case expectMaster.GetDepartment().GetValue() != actual.GetDepartment().GetValue():
		t.Fatal("error department")
	default:
		break
	}

	expectInbounds, err := json.Marshal(expect.GetInbounds())
	if err != nil {
		panic(err)
	}
	inbounds, err := json.Marshal(actual.GetInbounds())
	if err != nil {
		panic(err)
	}
	if string(expectInbounds) != string(inbounds) {
		t.Fatal("error inbounds")
	}

	expectOutbounds, err := json.Marshal(expect.GetOutbounds())
	if err != nil {
		panic(err)
	}
	outbounds, err := json.Marshal(actual.GetOutbounds())
	if err != nil {
		panic(err)
	}
	if string(expectOutbounds) != string(outbounds) {
		t.Fatal("error inbounds")
	}
}

func buildCircuitBreakerKey(id, version string) string {
	return fmt.Sprintf("%s_%s", id, version)
}

// 彻底删除熔断规则
func (d *DiscoverTestSuit) cleanCircuitBreaker(id, version string) {
	d.GetTestDataClean().CleanCircuitBreaker(id, version)
}

// 彻底删除熔断规则发布记录
func (d *DiscoverTestSuit) cleanCircuitBreakerRelation(name, namespace, ruleID, ruleVersion string) {
	d.GetTestDataClean().CleanCircuitBreakerRelation(name, namespace, ruleID, ruleVersion)
}

func (d *DiscoverTestSuit) cleanReportClient() {
	d.GetTestDataClean().CleanReportClient()
}

func (d *DiscoverTestSuit) cleanServices(services []*apiservice.Service) {
	d.GetTestDataClean().CleanServices(services)
}

func (d *DiscoverTestSuit) cleanNamespace(n string) {
	d.GetTestDataClean().CleanNamespace(n)
}

func (d *DiscoverTestSuit) cleanAllService() {
	d.GetTestDataClean().CleanAllService()
}

// 获取指定长度str
func genSpecialStr(n int) string {
	str := ""
	for i := 0; i < n; i++ {
		str += "a"
	}
	return str
}

// 解析字符串sid为modID和cmdID
func parseStr2Sid(sid string) (uint32, uint32) {
	items := strings.Split(sid, ":")
	if len(items) != 2 {
		return 0, 0
	}

	mod, _ := strconv.ParseUint(items[0], 10, 32)
	cmd, _ := strconv.ParseUint(items[1], 10, 32)
	return uint32(mod), uint32(cmd)
}

// 判断一个resp是否执行成功
func respSuccess(resp api.ResponseMessage) bool {

	ret := api.CalcCode(resp) == 200

	return ret
}

func respNotFound(resp api.ResponseMessage) bool {
	res := apimodel.Code(resp.GetCode().GetValue()) == apimodel.Code_NotFoundResource

	return res
}

func rollbackDbTx(dbTx *sqldb.BaseTx) {
	if err := dbTx.Rollback(); err != nil {
		log.Errorf("fail to rollback db tx, err %v", err)
	}
}

func commitDbTx(dbTx *sqldb.BaseTx) {
	if err := dbTx.Commit(); err != nil {
		log.Errorf("fail to commit db tx, err %v", err)
	}
}

func rollbackBoltTx(tx *bolt.Tx) {
	if err := tx.Rollback(); err != nil {
		log.Errorf("fail to rollback bolt tx, err %v", err)
	}
}

func commitBoltTx(tx *bolt.Tx) {
	if err := tx.Commit(); err != nil {
		log.Errorf("fail to commit bolt tx, err %v", err)
	}
}
