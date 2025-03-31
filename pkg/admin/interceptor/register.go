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

package inteceptor

import (
	"context"

	"github.com/pole-io/pole-server/pkg/admin"
	admin_auth "github.com/pole-io/pole-server/pkg/admin/interceptor/auth"
	"github.com/pole-io/pole-server/pkg/auth"
)

func init() {
	err := admin.RegisterServerProxy("auth", func(ctx context.Context,
		pre admin.AdminOperateServer) (admin.AdminOperateServer, error) {

		userSvr, err := auth.GetUserServerContext(ctx)
		if err != nil {
			return nil, err
		}

		policySvr, err := auth.GetStrategyServerContext(ctx)
		if err != nil {
			return nil, err
		}
		return admin_auth.NewServer(pre, userSvr, policySvr), nil
	})
	if err != nil {
		panic(err)
	}
}
