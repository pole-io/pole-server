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

package eurekaserver

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"

	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/cache"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/valid"
	"github.com/pole-io/pole-server/pkg/service"
	storeplugin "github.com/pole-io/pole-server/plugin/store"
	"github.com/pole-io/pole-server/plugin/store/mock"
	testsuit "github.com/pole-io/pole-server/test/suit"
)

func TestEurekaServer_renew(t *testing.T) {
	eventhub.InitEventHub()
	ins := &svctypes.Instance{
		ServiceID: utils.NewUUID(),
		Proto: &apiservice.Instance{
			Service:   protobuf.NewStringValue("echo"),
			Namespace: protobuf.NewStringValue("default"),
			Host:      protobuf.NewStringValue("127.0.0.1"),
			Port:      protobuf.NewUInt32Value(8080),
			HealthCheck: &apiservice.HealthCheck{
				Type: apiservice.HealthCheck_HEARTBEAT,
				Heartbeat: &apiservice.HeartbeatHealthCheck{
					Ttl: &wrapperspb.UInt32Value{
						Value: 5,
					},
				},
			},
		},
		Valid: true,
	}

	insId, resp := valid.CheckInstanceTetrad(ins.Proto)
	if resp != nil {
		t.Fatal(resp.GetInfo().GetValue())
		return
	}

	ins.Proto.Id = protobuf.NewStringValue(insId)

	disableBeatIns := &svctypes.Instance{
		ServiceID: utils.NewUUID(),
		Proto: &apiservice.Instance{
			Service:   protobuf.NewStringValue("echo"),
			Namespace: protobuf.NewStringValue("default"),
			Host:      protobuf.NewStringValue("127.0.0.2"),
			Port:      protobuf.NewUInt32Value(8081),
			HealthCheck: &apiservice.HealthCheck{
				Type: apiservice.HealthCheck_HEARTBEAT,
				Heartbeat: &apiservice.HeartbeatHealthCheck{
					Ttl: &wrapperspb.UInt32Value{
						Value: 5,
					},
				},
			},
		},
		Valid: true,
	}

	disableBeatInsId, resp := valid.CheckInstanceTetrad(disableBeatIns.Proto)
	if resp != nil {
		t.Fatal(resp.GetInfo().GetValue())
		return
	}

	disableBeatIns.Proto.Id = protobuf.NewStringValue(disableBeatInsId)

	ctrl := gomock.NewController(t)

	mockTx := mock.NewMockTx(ctrl)
	mockTx.EXPECT().Commit().Return(nil).AnyTimes()
	mockTx.EXPECT().Rollback().Return(nil).AnyTimes()
	mockTx.EXPECT().CreateReadView().Return(nil).AnyTimes()

	mockStore := mock.NewMockStore(ctrl)
	mockStore.EXPECT().
		GetMoreInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(map[string]*svctypes.Instance{
			insId:            ins,
			disableBeatInsId: disableBeatIns,
		}, nil)
	mockStore.EXPECT().StartReadTx().Return(mockTx, nil).AnyTimes()
	mockStore.EXPECT().
		GetMoreServices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(map[string]*svctypes.Service{
			ins.ServiceID: {
				ID:        ins.ServiceID,
				Name:      ins.Proto.GetService().GetValue(),
				Namespace: ins.Proto.GetNamespace().GetValue(),
			},
		}, nil)

	mockStore.EXPECT().GetMoreServiceContracts(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockStore.EXPECT().GetMoreClients(gomock.Any(), gomock.Any()).Return(map[string]*types.Client{}, nil).AnyTimes()
	mockStore.EXPECT().GetMoreGrayResouces(gomock.Any(), gomock.Any()).Return([]*rules.GrayResource{}, nil).AnyTimes()
	mockStore.EXPECT().GetInstancesCountTx(gomock.Any()).AnyTimes().Return(uint32(1), nil)
	mockStore.EXPECT().GetUnixSecond(gomock.Any()).AnyTimes().Return(time.Now().Unix(), nil)
	mockStore.EXPECT().GetServicesCount().Return(uint32(1), nil).AnyTimes()
	mockStore.EXPECT().StartLeaderElection(gomock.Any()).AnyTimes()
	mockStore.EXPECT().GetMoreNamespaces(gomock.Any()).Return(nil, nil).AnyTimes()
	mockStore.EXPECT().Destroy().Return(nil)
	mockStore.EXPECT().Initialize(gomock.Any()).Return(nil).AnyTimes()
	mockStore.EXPECT().Name().Return("eureka_store_test").AnyTimes()

	eurekaSuit := newEurekaTestSuit()
	eurekaSuit.ReplaceStore(func() store.Store {
		storeplugin.TestGetStore()
		store.StoreSlots["eureka_store_test"] = mockStore
		return mockStore
	})
	eurekaSuit.Initialize(func(conf *testsuit.TestConfig) {
		conf.DisableAuth = true
		conf.Cache = cache.Config{}
		conf.DisableConfig = true
		conf.ServiceCacheEntries = service.GetRegisterCaches()
		store.SetStoreConfig(&store.Config{
			Name: "eureka_store_test",
		})
	})

	defer eurekaSuit.Destroy()

	t.Run("eureka客户端心跳上报-实例正常且开启心跳", func(t *testing.T) {
		svr := &EurekaServer{
			healthCheckServer: eurekaSuit.HealthCheckServer(),
		}
		code := svr.renew(context.Background(), ins.Namespace(), "", insId, false)
		assert.Equalf(t, api.ExecuteSuccess, code, "code need success, actual : %d", code)
	})

	t.Run("eureka客户端心跳上报-实例未开启心跳", func(t *testing.T) {
		svr := &EurekaServer{
			healthCheckServer: eurekaSuit.HealthCheckServer(),
		}
		code := svr.renew(context.Background(), ins.Namespace(), "", disableBeatInsId, false)
		assert.Equalf(t, api.ExecuteSuccess, code, "code need success, actual : %d", code)
	})

	t.Run("eureka客户端心跳上报-实例不存在", func(t *testing.T) {
		svr := &EurekaServer{
			healthCheckServer: eurekaSuit.HealthCheckServer(),
		}
		instId := utils.NewUUID()
		var code uint32
		for i := 0; i < 5; i++ {
			code = svr.renew(context.Background(), ins.Namespace(), "", instId, false)
			time.Sleep(time.Second)
		}
		assert.Equalf(t, api.NotFoundResource, code, "code need notfound, actual : %d", code)
	})

}
