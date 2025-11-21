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

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	"github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/store"
	storeapi "github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/common/batchctrl"
)

// NewBatchRegisterClientCtrl 注册客户端批量操作对象
func NewBatchRegisterClientCtrl(ctx context.Context, s store.Store, config *CtrlConfig) (*batchctrl.BatchController, error) {
	ctrl, err := newBatchCtrl(ctx, "client_register", config, registerClientHandler(s))
	return ctrl, err
}

// NewBatchDeregisterClientCtrl 注册客户端批量操作对象
func NewBatchDeregisterClientCtrl(ctx context.Context, s store.Store, config *CtrlConfig) (*batchctrl.BatchController, error) {
	ctrl, err := newBatchCtrl(ctx, "client_deregister", config, deregisterClientHandler(s))
	return ctrl, err
}

// registerHandler 外部应该把鉴权完成
// 判断实例是否存在，也可以提前判断，减少batch复杂度
// 提前通过token判断，再进入batch操作
// batch操作，只是写操作
func registerClientHandler(s store.Store) func(futures []batchctrl.Future) {
	return func(futures []batchctrl.Future) {
		if len(futures) == 0 {
			return
		}

		log.Infof("[Batch] Start batch creating clients count: %d", len(futures))

		// 调用batch接口，创建实例
		clients := make([]*types.Client, 0, len(futures))
		for _, entry := range futures {
			clients = append(clients, types.NewClient(entry.Param().(*service_manage.Client)))
		}
		if err := s.BatchAddClients(clients); err != nil {
			sendReply(futures, storeapi.StoreCode2APICode(err), err)
			return
		}

		sendReply(futures, apimodel.Code_ExecuteSuccess, nil)
	}
}

// deregisterHandler 外部应该把鉴权完成
// 判断实例是否存在，也可以提前判断，减少batch复杂度
// 提前通过token判断，再进入batch操作
// batch操作，只是写操作
func deregisterClientHandler(s store.Store) func(futures []batchctrl.Future) {
	return func(futures []batchctrl.Future) {
		if len(futures) == 0 {
			return
		}

		log.Infof("[Batch] Start batch deleting clients count: %d", len(futures))

		// 调用batch接口，创建实例
		clients := make([]string, 0, len(futures))
		for _, entry := range futures {
			// 注释：方法调用改动 - GetId()返回类型从*wrapperspb.StringValue改为string，去掉.GetValue()调用
			id := entry.Param().(*service_manage.Client).GetId()
			clients = append(clients, id)
		}
		if err := s.BatchDeleteClients(clients); err != nil {
			sendReply(futures, storeapi.StoreCode2APICode(err), err)
			return
		}

		sendReply(futures, apimodel.Code_ExecuteSuccess, nil)
	}
}
