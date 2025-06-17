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

package batch

import (
	"context"
	"errors"
	"time"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/cache"
	"github.com/pole-io/pole-server/pkg/common/batchctrl"
)

// Controller 批量控制器
type Controller struct {
	svcWatch         *batchctrl.BatchController
	register         *batchctrl.BatchController
	deregister       *batchctrl.BatchController
	heartbeat        *batchctrl.BatchController
	clientRegister   *batchctrl.BatchController
	clientDeregister *batchctrl.BatchController
}

// NewBatchCtrlWithConfig 根据配置文件创建一个批量控制器
func NewBatchCtrlWithConfig(ctx context.Context, storage store.Store, cacheMgn *cache.CacheManager, config *Config) (*Controller, error) {
	if config == nil {
		return nil, nil
	}

	bc := &Controller{}

	if config.Register.Open {
		register, err := NewBatchRegisterCtrl(ctx, storage, config.Register)
		if err != nil {
			log.Errorf("[Batch] new batch register instance ctrl err: %s", err.Error())
			return nil, err
		}
		bc.register = register
	}

	if config.Deregister.Open {
		deregister, err := NewBatchDeregisterCtrl(ctx, storage, config.Deregister)
		if err != nil {
			log.Errorf("[Batch] new batch deregister instance ctrl err: %s", err.Error())
			return nil, err
		}
		bc.deregister = deregister
	}

	if config.Heartbeat.Open {
		heartbeat, err := NewBatchHeartbeatCtrl(ctx, storage, config.Heartbeat)
		if err != nil {
			log.Errorf("[Batch] new batch heartbeat instance ctrl err: %s", err.Error())
			return nil, err
		}
		bc.heartbeat = heartbeat
	}

	if config.ClientRegister.Open {
		clientRegister, err := NewBatchRegisterClientCtrl(ctx, storage, config.ClientRegister)
		if err != nil {
			log.Errorf("[Batch] new batch client register ctrl err: %s", err.Error())
			return nil, err
		}
		bc.clientRegister = clientRegister
	}

	if config.ClientDeregister.Open {
		clientDeregister, err := NewBatchDeregisterClientCtrl(ctx, storage, config.ClientDeregister)
		if err != nil {
			log.Errorf("[Batch] new batch client deregister ctrl err: %s", err.Error())
			return nil, err
		}
		bc.clientDeregister = clientDeregister
	}

	svcWatch, err := NewBatchServiceSubscribersCtrl(ctx, storage, &CtrlConfig{
		Open:          true,
		WaitTime:      "64ms",
		Concurrency:   32,
		QueueSize:     16384,
		MaxBatchCount: 64,
	})
	if err != nil {
		log.Errorf("[Batch] new batch service subscriber ctrl err: %s", err.Error())
		return nil, err
	}
	bc.svcWatch = svcWatch

	return bc, nil
}

// CreateInstanceOpen 创建是否开启
func (bc *Controller) CreateInstanceOpen() bool {
	return bc.register != nil
}

// DeleteInstanceOpen 删除实例是否开启
func (bc *Controller) DeleteInstanceOpen() bool {
	return bc.deregister != nil
}

// HeartbeatOpen 心跳是否开启
func (bc *Controller) HeartbeatOpen() bool {
	return bc.heartbeat != nil
}

// ClientRegisterOpen 添加客户端是否开启
func (bc *Controller) ClientRegisterOpen() bool {
	return bc.clientRegister != nil
}

// ClientDeregisterOpen 删除客户端是否开启
func (bc *Controller) ClientDeregisterOpen() bool {
	return bc.clientDeregister != nil
}

// AsyncCreateInstance 异步创建实例，返回一个future，根据future获取创建结果
func (bc *Controller) AsyncCreateInstance(svcId string, instance *apiservice.Instance, needWait bool) batchctrl.Future {
	// 发送到注册请求队列
	return bc.register.Submit(&InstanceFuture{
		serviceId: svcId,
		needWait:  needWait,
		request:   instance,
		begin:     time.Now(),
	})
}

// AsyncDeleteInstance 异步合并反注册
func (bc *Controller) AsyncDeleteInstance(instance *apiservice.Instance, needWait bool) batchctrl.Future {
	return bc.deregister.Submit(&InstanceFuture{
		request:  instance,
		result:   make(chan error, 1),
		needWait: true,
	})
}

// AsyncHeartbeat 异步心跳
func (bc *Controller) AsyncHeartbeat(instance *apiservice.Instance, healthy bool, lastBeatTime int64) batchctrl.Future {
	return bc.heartbeat.Submit(&InstanceFuture{
		request:              instance,
		result:               make(chan error, 1),
		healthy:              healthy,
		needWait:             true,
		lastHeartbeatTimeSec: lastBeatTime,
	})
}

// AsyncRegisterClient 异步合并反注册
func (bc *Controller) AsyncRegisterClient(client *apiservice.Client) batchctrl.Future {
	return bc.clientRegister.Submit(client)
}

// AsyncDeregisterClient 异步合并反注册
func (bc *Controller) AsyncDeregisterClient(client *apiservice.Client) batchctrl.Future {
	return bc.clientDeregister.Submit(client)
}

// AsyncRecordServiceSubscriberGraph 异步合并反注册
func (bc *Controller) AsyncRecordServiceSubscriberGraph(d *svctypes.ServiceSubscriber) batchctrl.Future {
	return bc.svcWatch.Submit(d)
}

// sendReply 批量答复futures
func sendReply(futures any, code apimodel.Code, result error) {
	switch futureType := futures.(type) {
	case []batchctrl.Future:
		for _, entry := range futureType {
			entry.Reply(code, result)
		}
	case map[string]batchctrl.Future:
		for _, entry := range futureType {
			entry.Reply(code, result)
		}
	default:
		log.Errorf("[Controller] not found reply futures type: %T", futures)
	}
}

// newBatchClientCtrl 注册客户端批量操作对象
func newBatchCtrl(ctx context.Context, label string, config *CtrlConfig, handler func([]batchctrl.Future)) (*batchctrl.BatchController, error) {
	duration, err := time.ParseDuration(config.WaitTime)
	if err != nil {
		log.Errorf("[Batch] parse waitTime(%s) err: %s", config.WaitTime, err.Error())
		return nil, err
	}
	if duration == 0 {
		log.Errorf("[Batch] config waitTime is invalid")
		return nil, errors.New("config waitTime is invalid")
	}
	return batchctrl.NewBatchController(
		ctx,
		batchctrl.CtrlConfig{
			Label:         label,
			Concurrency:   uint32(config.Concurrency),
			QueueSize:     uint32(config.QueueSize),
			WaitTime:      duration,
			MaxBatchCount: uint32(config.MaxBatchCount),
			Handler:       handler,
		},
	), nil
}
