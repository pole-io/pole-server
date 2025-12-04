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

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"

	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/specification/source/go/api/v1/model"
	"github.com/pole-io/specification/source/go/api/v1/service_manage"
)

var (
	// ServerAddr pole-server服务端接入地址IP，默认为 127.0.0.1
	ServerAddr = os.Getenv("SERVER_IP")
	// GRPCPort pole-server服务端接入地址 GRPC 协议端口，默认为 8091
	GRPCPort, _ = strconv.ParseInt(os.Getenv("SERVER_PORT"), 10, 64)
	// HttpPort pole-server服务端接入地址 HTTP 协议端口，默认为 8090
	HttpPort, _ = strconv.ParseInt(os.Getenv("SERVER_PORT"), 10, 64)
	// RunMode 运行模式，内容为 VERIFY(验证模式)/BENCHMARK(压测模式)/ALL(同时执行验证模式+压测模式)
	RunMode = os.Getenv("RUN_MODE")
	// Service 服务名
	Service = os.Getenv("SERVICE")
	// Namespace 命名空间
	Namespace = os.Getenv("NAMESPACE")
	// BasePort 端口起始
	BasePort, _ = strconv.ParseInt(os.Getenv("BASE_PORT"), 10, 64)
	// PortNum 单个 POD 注册多少个端口
	PortNum, _ = strconv.ParseInt(os.Getenv("PORT_NUM"), 10, 64)
	// BeatInterval 心跳默认周期, 单位为秒
	BeatInterval, _ = strconv.ParseInt(os.Getenv("BEAT_INTERVAL"), 10, 64)
	// CheckInterval 检查任务执行周期
	CheckInterval, _ = time.ParseDuration(os.Getenv("CHECK_INTERVAL"))
	// PodIP 实例注册 IP
	PodIP = os.Getenv("POD_IP")
)

const (
	defaultSeverIP       = "127.0.0.1"
	defaultGrpcPort      = 8091
	defaultHttpPort      = 8090
	defaultBeatInterval  = 5
	defaultBasePort      = 8080
	defaultPortNum       = 1
	defaultService       = "benchmark-heartbeat"
	defaultNamesapce     = "benchmark"
	metricsPort          = 9090
	defaultCheckInterval = time.Minute
	defaultPorIP         = "172.0.0.1"
)

func setDefault() {
	if ServerAddr == "" {
		ServerAddr = defaultSeverIP
	}
	if GRPCPort == 0 {
		GRPCPort = defaultGrpcPort
	}
	if HttpPort == 0 {
		HttpPort = defaultHttpPort
	}
	if Service == "" {
		Service = defaultService
	}
	if Namespace == "" {
		Namespace = defaultNamesapce
	}
	if BasePort == 0 {
		BasePort = defaultBasePort
	}
	if PortNum == 0 {
		PortNum = 1
	}
	if CheckInterval == 0 {
		CheckInterval = defaultCheckInterval
	}
	if PodIP == "" {
		PodIP = defaultPorIP
	}
	if BeatInterval == 0 {
		BeatInterval = 1
	}
	log.Printf("run_mode(%s)", RunMode)
	log.Printf("server_addr(%s)", ServerAddr)
	log.Printf("grpc_port(%d)", GRPCPort)
	log.Printf("http_port(%d)", HttpPort)
	log.Printf("namespace(%s)", Namespace)
	log.Printf("service(%s)", Service)
	log.Printf("base_port(%d)", BasePort)
	log.Printf("pod_ip(%s)", PodIP)
	log.Printf("port_num(%d)", PortNum)
	log.Printf("beat_interval(%+v)", BeatInterval)
	log.Printf("check_interval(%v)", CheckInterval)
}

func setMetrics() {
}

func main() {
	f, err := os.Create("./health_check.log")
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)
	setDefault()
	setMetrics()
	switch strings.ToLower(RunMode) {
	case "verify":
		go runVerifyMode()
	case "benchmark":
		go runBenchmarkMode()
	case "all":
		go runVerifyMode()
		go runBenchmarkMode()
	default:
		panic("unknown run mode, please export RUN_MODE=verify or RUN_MODE=benchmark or RUN_MODE=all")
	}
	go func() {
	}()
	mainLoop()
}

func runVerifyMode() {
	log.Printf("[INFO] start run verify task")
	ticker := time.NewTicker(CheckInterval)

	checkInstanceHealth := func() {
		req := &service_manage.DiscoverRequest{
			Type: service_manage.DiscoverRequest_INSTANCE,
			Service: &service_manage.Service{
				Namespace: Namespace,
				Name:      Service,
			},
		}

		marshaler := jsonpb.Marshaler{}
		body, _ := marshaler.MarshalToString(req)
		rsp, err := http.Post(fmt.Sprintf("http://%s:%d/v1/Discover", ServerAddr, HttpPort), "application/json", bytes.NewBufferString(body))
		if err != nil {
			log.Printf("[ERROR] send discover to server fail: %s", err.Error())
			return
		}

		defer func() {
			_ = rsp.Body.Close()
		}()
		data, _ := io.ReadAll(rsp.Body)
		discoverRsp := &service_manage.DiscoverResponse{}
		if err := jsonpb.Unmarshal(bytes.NewBuffer(data), discoverRsp); err != nil {
			log.Printf("[ERROR] unmarshaler discover resp fail: %s", err.Error())
			return
		}
		if discoverRsp.GetCode() != uint32(model.Code_ExecuteSuccess) {
			log.Printf("[ERROR] receive discover resp fail: %s", discoverRsp.GetInfo())
			return
		}
		unHealthCount := 0
		// 检查实例健康状态
		instances := discoverRsp.GetInstances()
		for i := range instances {
			isHealth := instances[i].GetHealthy()
			if !isHealth {
				unHealthCount++
			}
		}
		if unHealthCount > 0 {
			log.Printf("[ERROR] total instance unhealthy: %d", unHealthCount)
		} else {
			log.Printf("[INFO] all instance is healthy, you are luckly")
		}
	}

	for range ticker.C {
		checkInstanceHealth()
	}
}

func runBenchmarkMode() {
	log.Printf("[INFO] start run benchmark task")
	// 先注册
	for i := 0; i < int(PortNum); i++ {
		// 每个 Port 对应一个 Grpc Connection
		conn, err := grpc.DialContext(context.Background(), fmt.Sprintf("%s:%d", ServerAddr, GRPCPort),
			grpc.WithBlock(),
			grpc.WithInsecure(),
		)
		if err != nil {
			panic(err)
		}

		client := service_manage.NewDiscoverGRPCClient(conn)

		instance := &service_manage.Instance{
			Namespace:         Namespace,
			Service:           Service,
			Host:              PodIP,
			Port:              uint32(int(BasePort) + i),
			EnableHealthCheck: true,
			HealthCheck: &service_manage.HealthCheck{
				Type: service_manage.HealthCheck_HEARTBEAT,
				Heartbeat: &service_manage.HeartbeatHealthCheck{
					Ttl: uint32(BeatInterval),
				},
			},
		}

		// 计算实例 ID
		instanceID, err := utils.CalculateInstanceID(Namespace, Service, "", PodIP, uint32(int(BasePort)+i))
		if err != nil {
			panic(fmt.Sprintf("calculate instance id failed: %v", err))
		}
		instance.Id = instanceID

		resp, err := client.RegisterInstance(context.Background(), instance)
		if err != nil {
			panic(err)
		}
		if resp.GetCode() != uint32(model.Code_ExecuteSuccess) {
			panic(resp.GetInfo())
		}
		log.Printf("[INFO] instance register success id: %s", instance.GetId())
		go func(instance *service_manage.Instance) {
			ticker := time.NewTicker(time.Duration(BeatInterval) * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				// 使用 HTTP API 上报心跳
				heartbeatReq := &service_manage.Instance{
					Id:        instance.GetId(),
					Namespace: instance.GetNamespace(),
					Service:   instance.GetService(),
					Host:      instance.GetHost(),
					Port:      instance.GetPort(),
				}

				marshaler := jsonpb.Marshaler{}
				body, err := marshaler.MarshalToString(heartbeatReq)
				if err != nil {
					log.Printf("[ERROR] instance(%s) marshal heartbeat request fail: %s", instance.GetId(), err.Error())
					continue
				}

				httpResp, err := http.Post(
					fmt.Sprintf("http://%s:%d/v1/Heartbeat", ServerAddr, HttpPort),
					"application/json",
					bytes.NewBufferString(body),
				)
				if err != nil {
					log.Printf("[ERROR] instance(%s) beat fail error: %s", instance.GetId(), err.Error())
					continue
				}

				data, _ := io.ReadAll(httpResp.Body)
				_ = httpResp.Body.Close()

				heartbeatResp := &model.Response{}
				if err := jsonpb.Unmarshal(bytes.NewBuffer(data), heartbeatResp); err != nil {
					log.Printf("[ERROR] instance(%s) unmarshal heartbeat response fail: %s", instance.GetId(), err.Error())
					continue
				}

				if heartbeatResp.GetCode() != uint32(model.Code_ExecuteSuccess) {
					log.Printf("[ERROR] instance(%s) beat fail info: %s", instance.GetId(), heartbeatResp.GetInfo())
				}
			}
		}(instance)
	}

	// 发起任务开始定期 心跳上报
}

// mainLoop 等待信号量执行退出
func mainLoop() {
	ch := make(chan os.Signal, 1)

	// 监听信号量
	signal.Notify(ch, []os.Signal{
		syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGSEGV,
	}...)

	for range ch {
		log.Printf("[INFO] catch signal, stop benchmark server")
		return
	}
}
