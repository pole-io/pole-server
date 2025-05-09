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

package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/metadata"

	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/pkg/common/utils"
)

// RegisterInstance 注册服务实例
func (c *Client) RegisterInstance(instance *apiservice.Instance) error {
	fmt.Printf("\nregister instance\n")

	md := metadata.Pairs("request-id", utils.NewUUID())
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	rsp, err := c.Worker.RegisterInstance(ctx, instance)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	fmt.Printf("%v\n", rsp)

	return nil
}

// DeregisterInstance 反注册服务实例
func (c *Client) DeregisterInstance(instance *apiservice.Instance) error {
	fmt.Printf("\nderegister instance\n")

	md := metadata.Pairs("request-id", utils.NewUUID())
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rsp, err := c.Worker.DeregisterInstance(ctx, instance)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	fmt.Printf("%v\n", rsp)

	return nil
}
