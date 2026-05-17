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

package discover

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
	"github.com/pole-io/pole-server/pkg/service"
	"github.com/pole-io/pole-server/plugin/apiserver/nacosserver/core"
	"github.com/pole-io/pole-server/plugin/apiserver/nacosserver/model"
)

func (n *DiscoverServer) handleRegister(ctx context.Context, namespace, serviceName string, ins *model.Instance) error {
	specIns := model.PrepareSpecInstance(namespace, serviceName, ins)
	resp := n.discoverSvr.RegisterInstance(ctx, specIns)
	if apimodel.Code(resp.GetCode()) != apimodel.Code_ExecuteSuccess {
		return &model.NacosError{
			ErrCode: int32(model.ExceptionCode_ServerError),
			ErrMsg:  resp.GetInfo(),
		}
	}
	return nil
}

func (n *DiscoverServer) handleUpdate(ctx context.Context, namespace, serviceName string, ins *model.Instance) error {
	specIns := model.PrepareSpecInstance(namespace, serviceName, ins)
	if specIns.GetId() == "" {
		insId, errRsp := valid.CheckInstanceTetrad(specIns)
		if errRsp != nil {
			return &model.NacosError{
				ErrCode: int32(model.ExceptionCode_ServerError),
				ErrMsg:  errRsp.GetInfo(),
			}
		}
		specIns.Id = insId
	}
	svr := n.discoverSvr.(*service.Server)
	saveIns, err := svr.Store().GetInstance(specIns.GetId())
	if err != nil {
		return &model.NacosError{
			ErrCode: int32(model.ExceptionCode_ServerError),
			ErrMsg:  err.Error(),
		}
	}
	specIns = mergeUpdateInstanceInfo(specIns, saveIns)
	resp := n.discoverSvr.UpdateInstance(ctx, specIns)
	if apimodel.Code(resp.GetCode()) != apimodel.Code_ExecuteSuccess {
		return &model.NacosError{
			ErrCode: int32(model.ExceptionCode_ServerError),
			ErrMsg:  resp.GetInfo(),
		}
	}
	return nil
}

func (n *DiscoverServer) handleDeregister(ctx context.Context, namespace, svcName string, ins *model.Instance) error {
	specIns := model.PrepareSpecInstance(namespace, svcName, ins)
	resp := n.discoverSvr.DeregisterInstance(ctx, specIns)
	if apimodel.Code(resp.GetCode()) != apimodel.Code_ExecuteSuccess {
		return &model.NacosError{
			ErrCode: int32(model.ExceptionCode_ServerError),
			ErrMsg:  resp.GetInfo(),
		}
	}
	return nil
}

// handleBeat com.alibaba.nacos.naming.core.InstanceOperatorClientImpl#handleBeat
func (n *DiscoverServer) handleBeat(ctx context.Context, namespace, svcName string,
	clientBeat *model.ClientBeat) (map[string]interface{}, error) {
	svcName = model.ReplaceNacosService(svcName)
	svc := n.discoverSvr.Cache().Service().GetServiceByName(svcName, namespace)
	if svc == nil {
		return nil, &model.NacosError{
			ErrCode: int32(model.ExceptionCode_ServerError),
			ErrMsg:  "service not found: " + svcName + "@" + namespace,
		}
	}

	resp := n.healthSvr.Report(ctx, &apiservice.Instance{
		Service:   model.ReplaceNacosService(svcName),
		Namespace: namespace,
		Host:      clientBeat.Ip,
		Port:      uint32(clientBeat.Port),
	})
	rspCode := apimodel.Code(resp.GetCode())

	if rspCode == apimodel.Code_ExecuteSuccess {
		return map[string]interface{}{
			"code":               10200,
			"clientBeatInterval": model.ClientBeatIntervalMill,
			"lightBeatEnabled":   true,
		}, nil
	}

	if rspCode == apimodel.Code_NotFoundResource {
		return map[string]interface{}{
			"code":               20404,
			"clientBeatInterval": model.ClientBeatIntervalMill,
			"lightBeatEnabled":   true,
		}, nil
	}

	return nil, &model.NacosError{
		ErrCode: int32(model.ExceptionCode_ServerError),
		ErrMsg:  resp.GetInfo(),
	}

}

// handleQueryInstances com.alibaba.nacos.naming.controllers.InstanceController#list
func (n *DiscoverServer) handleQueryInstances(ctx context.Context, params map[string]string) (interface{}, error) {
	namespace := params[model.ParamNamespaceID]
	group := model.GetGroupName(params[model.ParamServiceName])
	svcName := model.GetServiceName(params[model.ParamServiceName])
	clusters := params["clusters"]
	clientIP := params["clientIP"]
	udpPort, _ := strconv.ParseInt(params["udpPort"], 10, 32)
	healthyOnly, _ := strconv.ParseBool(params["healthyOnly"])

	if n.pushCenter != nil && udpPort > 0 {
		n.pushCenter.AddSubscriber(core.Subscriber{
			Key:         fmt.Sprintf("%s:%d", clientIP, udpPort),
			App:         utils.DefaultString(params["app"], "unknown"),
			AddrStr:     clientIP,
			Ip:          clientIP,
			Port:        int(udpPort),
			NamespaceId: namespace,
			Group:       group,
			Service:     svcName,
			Cluster:     clusters,
			Type:        core.UDPCPush,
		})
	}

	filterCtx := &core.FilterContext{
		Service:     core.ToNacosService(n.discoverSvr.Cache(), namespace, svcName, group),
		Clusters:    strings.Split(clusters, ","),
		EnableOnly:  true,
		HealthyOnly: healthyOnly,
	}
	// 默认只下发 enable 的实例
	result := n.store.ListInstances(filterCtx, core.SelectInstancesWithHealthyProtection)
	// adapt for nacos v1.x SDK
	result.Name = fmt.Sprintf("%s%s%s", result.GroupName, model.DefaultNacosGroupConnectStr, result.Name)
	result.Namespace = model.ToNacosNamespace(namespace)
	return result, nil
}

func mergeUpdateInstanceInfo(req *apiservice.Instance, saveVal *svctypes.Instance) *apiservice.Instance {
	return req
}
