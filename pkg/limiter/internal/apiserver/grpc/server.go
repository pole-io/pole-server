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
	"errors"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"

	apiv2 "github.com/pole-io/specification/source/go/api/v1/traffic_manage/ratelimiter"

	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/limiter/internal/apiserver"
	"github.com/pole-io/pole-server/pkg/limiter/internal/ratelimitv2"
	"github.com/pole-io/pole-server/pkg/limiter/internal/statistics"
)

// Server grpc server
type Server struct {
	IP                 string
	Port               uint32
	server             *grpc.Server
	rateLimitServiceV2 *RateLimitServiceV2
	done               chan error
	stopOnce           sync.Once
}

// GetProtocol 返回协议
func (g *Server) GetProtocol() string {
	return "grpc"
}

// GetPort 返回port
func (g *Server) GetPort() uint32 {
	return g.Port
}

// Start 同步绑定 listener 后启动 gRPC 服务。
func Start(option map[string]interface{}, core *ratelimitv2.Server, statics statistics.Statis) (*Server, error) {
	ip, port, err := apiserver.ParseListenOption(option)
	if err != nil {
		return nil, err
	}
	g := &Server{
		IP:   ip,
		Port: port,
		done: make(chan error, 1),
	}
	// 初始化grpc server监听listener
	address := fmt.Sprintf("%s:%d", g.IP, g.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen limiter grpc %s: %w", address, err)
	}
	g.Port = uint32(listener.Addr().(*net.TCPAddr).Port)
	g.rateLimitServiceV2 = &RateLimitServiceV2{
		coreServer: core,
		statics:    statics,
	}
	// 创建grpc server
	server := grpc.NewServer(
		grpc.UnaryInterceptor(g.unaryInterceptor),
		grpc.StreamInterceptor(g.streamInterceptor),
	)
	apiv2.RegisterRateLimitGRPCV2Server(server, g.rateLimitServiceV2)
	g.server = server

	serviceInfos := server.GetServiceInfo()
	for key, serviceInfo := range serviceInfos {
		log.Infof("register service key %s, info is %+v", key, serviceInfo)
	}

	// 绑定到连接中，并开始监听server
	go func() {
		err := g.server.Serve(listener)
		if errors.Is(err, grpc.ErrServerStopped) {
			err = nil
		}
		g.done <- err
		close(g.done)
	}()
	return g, nil
}

// Done 返回运行期退出结果。
func (g *Server) Done() <-chan error {
	return g.done
}

// Stop 优雅停止，超时后强制断开。
func (g *Server) Stop(ctx context.Context) error {
	g.stopOnce.Do(func() {
		stopped := make(chan struct{})
		go func() {
			g.server.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-ctx.Done():
			g.server.Stop()
		}
	})
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}
