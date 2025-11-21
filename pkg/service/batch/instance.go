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
	"fmt"
	"strconv"

	// 注释：移除golang/protobuf/ptypes/wrappers导入 - 不再使用wrapper类型

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/pkg/types"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/store"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/batchctrl"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

var (
	ErrorNotFoundService          = errors.New("not found service")
	ErrorSameRegIsInstanceRequest = errors.New("there is the same instance request")
	ErrorRegIsInstanceTimeout     = errors.New("polaris-sever regis instance busy")
)

// NewBatchRegisterCtrl 注册实例批量操作对象
func NewBatchRegisterCtrl(ctx context.Context, s store.Store, config *CtrlConfig) (*batchctrl.BatchController, error) {
	ctrl, err := newBatchCtrl(ctx, "instance_register", config, registerInstanceHandler(s))
	return ctrl, err
}

// NewBatchDeregisterCtrl 实例反注册的操作对象
func NewBatchDeregisterCtrl(ctx context.Context, s store.Store, config *CtrlConfig) (*batchctrl.BatchController, error) {
	ctrl, err := newBatchCtrl(ctx, "instance_deregister", config, deregisterInstanceHandler(s))
	return ctrl, err
}

// NewBatchHeartbeatCtrl 实例心跳的操作对象
func NewBatchHeartbeatCtrl(ctx context.Context, s store.Store, config *CtrlConfig) (*batchctrl.BatchController, error) {
	ctrl, err := newBatchCtrl(ctx, "instance_heartbeat", config, heartbeatInstanceHandler(s))
	return ctrl, err
}

// registerHandler 外部应该把鉴权完成
// 判断实例是否存在，也可以提前判断，减少batch复杂度
// 提前通过token判断，再进入batch操作
// batch操作，只是写操作
func registerInstanceHandler(s store.Store) func(futures []batchctrl.Future) {
	return func(futures []batchctrl.Future) {
		if len(futures) == 0 {
			log.Warn("[Batch] futures is empty")
			return
		}

		log.Infof("[Batch] Start batch creating instances count: %d", len(futures))
		remains := make(map[string]batchctrl.Future, len(futures))
		for i := range futures {
			entry := futures[i]
			param := entry.Param().(*InstanceFuture)
			entry.Attach("isRegis", true) // 标记为注册请求

			// 注释：实例ID获取改动 - GetId()返回类型从*wrapperspb.StringValue改为string，去掉.GetValue()调用
			if _, ok := remains[param.request.GetId()]; ok {
				entry.Reply(apimodel.Code_SameInstanceRequest, ErrorSameRegIsInstanceRequest)
				continue
			}
			remains[param.request.GetId()] = entry
		}

		// 统一判断实例是否存在，存在则需要更新部分数据
		if err := batchRestoreInstanceIsolate(s, remains); err != nil {
			log.Errorf("[Batch] batch check instances existed err: %s", err.Error())
		}

		// 判断入参数组是否为0
		if len(remains) == 0 {
			log.Info("[Batch] all instances is existed, return create instances process")
			return
		}
		// 构造model数据
		instances := make([]*svctypes.Instance, 0, len(remains))
		for _, entry := range remains {
			param := entry.Param().(*InstanceFuture)
			ins := svctypes.CreateInstanceModel(param.serviceId, param.request)
			instances = append(instances, ins)
			entry.Attach("instance_val", ins)
		}
		// 调用batch接口，创建实例
		if err := s.BatchAddInstances(instances); err != nil {
			sendReply(remains, storeapi.StoreCode2APICode(err), err)
			return
		}

		sendReply(remains, apimodel.Code_ExecuteSuccess, nil)
	}
}

// heartbeatHandler 心跳状态变更处理函数
func heartbeatInstanceHandler(s store.Store) func(futures []batchctrl.Future) {
	return func(futures []batchctrl.Future) {
		if len(futures) == 0 {
			return
		}
		log.Infof("[Batch] start batch heartbeat instances count: %d", len(futures))
		ids := make(map[string]bool, len(futures))
		statusToIds := map[bool]map[string]int64{
			true:  make(map[string]int64, len(futures)),
			false: make(map[string]int64, len(futures)),
		}
		for _, entry := range futures {
			param := entry.Param().(*InstanceFuture)
			// 多个记录，只有后面的一个生效
			// 注释：实例ID获取改动 - GetId()返回string而非*wrapperspb.StringValue，业务逻辑保持不变
			id := param.request.GetId()
			if _, ok := ids[id]; ok {
				values := statusToIds[!param.healthy]
				delete(values, id)
			}
			ids[id] = false
			statusToIds[param.healthy][id] = param.lastHeartbeatTimeSec
		}

		// 转为不健康的实例，需要添加 metadata
		appendMetaReqs := make([]*store.InstanceMetadataRequest, 0, len(statusToIds[false]))
		// 转为健康的实例，需要删除 metadata
		removeMetaReqs := make([]*store.InstanceMetadataRequest, 0, len(statusToIds[true]))
		revision := utils.NewUUID()
		for healthy, values := range statusToIds {
			if len(values) == 0 {
				continue
			}
			idValues := make([]any, 0, len(values))
			for id := range values {
				if healthy {
					removeMetaReqs = append(removeMetaReqs, &store.InstanceMetadataRequest{
						InstanceID: id,
						Revision:   revision,
						Keys:       []string{types.MetadataInstanceLastHeartbeatTime},
					})
				} else {
					appendMetaReqs = append(appendMetaReqs, &store.InstanceMetadataRequest{
						InstanceID: id,
						Revision:   revision,
						Metadata: map[string]string{
							types.MetadataInstanceLastHeartbeatTime: strconv.FormatInt(values[id], 10),
						},
					})
				}
				idValues = append(idValues, id)
			}
			err := s.BatchSetInstanceHealthStatus(idValues, utils.StatusBoolToInt(healthy), utils.NewUUID())
			if err != nil {
				log.Errorf("[Batch] batch healthy check instances err: %s", err.Error())
				sendReply(futures, storeapi.StoreCode2APICode(err), err)
				return
			}
			if err := s.BatchAppendInstanceMetadata(appendMetaReqs); err != nil {
				log.Errorf("[Batch] batch healthy check instances append metadata err: %s", err.Error())
				sendReply(futures, storeapi.StoreCode2APICode(err), err)
				return
			}
			if err := s.BatchRemoveInstanceMetadata(removeMetaReqs); err != nil {
				log.Errorf("[Batch] batch healthy check instances remove metadata err: %s", err.Error())
				sendReply(futures, storeapi.StoreCode2APICode(err), err)
				return
			}
		}
		sendReply(futures, apimodel.Code_ExecuteSuccess, nil)
	}
}

// deregisterHandler 反注册处理函数
// 步骤：
//   - 从数据库中批量读取实例ID对应的实例简要信息：
//     包括：ID，host，port，serviceName，serviceNamespace，serviceToken
//   - 对instance做存在与token的双重校验，较少与数据库的交互
//   - 对于不存在的token，返回notFoundResource
//   - 对于token校验失败的，返回校验失败
//   - 调用批量接口删除实例
func deregisterInstanceHandler(s store.Store) func(futures []batchctrl.Future) {
	return func(futures []batchctrl.Future) {
		if len(futures) == 0 {
			return
		}

		log.Infof("[Batch] Start batch deregister instances count: %d", len(futures))
		remains := make(map[string]batchctrl.Future, len(futures))
		ids := make(map[string]bool, len(futures))
		for _, entry := range futures {
			param := entry.Param().(*InstanceFuture)
			if _, ok := remains[param.request.GetId()]; ok {
				entry.Reply(apimodel.Code_SameInstanceRequest, ErrorSameRegIsInstanceRequest)
				continue
			}

			remains[param.request.GetId()] = entry
			ids[param.request.GetId()] = false
		}

		// 统一鉴权与判断是否存在
		instances, err := s.GetInstancesBrief(ids)
		if err != nil {
			log.Errorf("[Batch] get instances service token err: %s", err.Error())
			sendReply(remains, storeapi.StoreCode2APICode(err), err)
			return
		}
		for _, future := range futures {
			param := future.Param().(*InstanceFuture)
			instance, ok := instances[param.request.GetId()]
			if !ok {
				// 不存在，意味着不需要删除了
				future.Reply(apimodel.Code_NotFoundResource, fmt.Errorf("%s", api.Code2Info(api.NotFoundResource)))
				delete(remains, param.request.GetId())
				continue
			}

			future.Attach("instance_val", instance) // 这里保存instance的目的：方便上层使用model数据
		}

		if len(remains) == 0 {
			log.Infof("[Batch] deregister instances verify failed or instances is not existed, no remain any instances")
			return
		}

		// 调用storage batch接口，删除实例
		args := make([]interface{}, 0, len(remains))
		for _, entry := range remains {
			req := entry.Param().(*InstanceFuture)
			// 注释：实例ID获取改动 - GetId()返回string，去掉.GetValue()调用，功能不变
			args = append(args, req.request.GetId())
		}
		if err := s.BatchDeleteInstances(args); err != nil {
			log.Errorf("[Batch] batch delete instances err: %s", err.Error())
			sendReply(remains, storeapi.StoreCode2APICode(err), err)
			return
		}

		sendReply(remains, apimodel.Code_ExecuteSuccess, nil)
	}
}

// batchRestoreInstanceIsolate 批量恢复实例的隔离状态，以请求为准，请求如果不存在，就以数据库为准
func batchRestoreInstanceIsolate(s store.Store, futures map[string]batchctrl.Future) error {
	if len(futures) == 0 {
		return nil
	}

	// 初始化所有的id都是不存在的
	ids := make(map[string]bool, len(futures))
	for _, entry := range futures {
		param := entry.Param().(*InstanceFuture)
		ids[param.request.GetId()] = false
	}
	var id2Isolate map[string]bool
	var err error
	if id2Isolate, err = s.BatchGetInstanceIsolate(ids); err != nil {
		log.Errorf("[Batch] check instances existed storage err: %s", err.Error())
		sendReply(futures, storeapi.StoreCode2APICode(err), err)
		return err
	}

	if len(id2Isolate) == 0 {
		return nil
	}

	if len(id2Isolate) > 0 {
		for id, isolate := range id2Isolate {
			if future, ok := futures[id]; ok {
				req := future.Param().(*InstanceFuture)
				// 注释：字段类型改动 - Isolate从*wrapperspb.BoolValue改为bool，直接赋值而非包装对象
				req.request.Isolate = bool(isolate)
			}
		}
	}
	return nil
}
