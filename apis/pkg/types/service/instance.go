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

package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	"github.com/pole-io/pole-server/apis/pkg/utils"
)

// InstanceCount Service instance statistics
type InstanceCount struct {
	// IsolateInstanceCount 隔离状态的实例
	IsolateInstanceCount uint32
	// HealthyInstanceCount 健康实例数
	HealthyInstanceCount uint32
	// TotalInstanceCount 总实例数
	TotalInstanceCount uint32
	// VersionCounts 按照实例的版本进行统计计算
	VersionCounts map[string]*InstanceVersionCount
}

// InstanceVersionCount instance version metrics count
type InstanceVersionCount struct {
	// IsolateInstanceCount 隔离状态的实例
	IsolateInstanceCount uint32
	// HealthyInstanceCount 健康实例数
	HealthyInstanceCount uint32
	// TotalInstanceCount 总实例数
	TotalInstanceCount uint32
}

// NamespaceServiceCount Namespace service data
type NamespaceServiceCount struct {
	// ServiceCount 服务数量
	ServiceCount uint32
	// InstanceCnt 实例健康数/实例总数
	InstanceCnt *InstanceCount
}

const (
	MetadataInstanceLastHeartbeatTime   = "internal-lastheartbeat"
	MetadataServiceProtectThreshold     = "internal-service-protectthreshold"
	MetadataRegisterFrom                = "internal-register-from"
	MetadataInternalMetaHealthCheckPath = "internal-healthcheck_path"
	MetadataInternalMetaTraceSampling   = "internal-trace_sampling"
)

// Instance 组合了api的Instance对象
type Instance struct {
	Proto             *apiservice.Instance
	ServiceID         string
	ServicePlatformID string
	// Valid Whether it is deleted by logic
	Valid bool
	// ModifyTime Update time of instance
	ModifyTime time.Time
	// CreateTime Create time of instance
	CreateTime time.Time
}

// ID get id
func (i *Instance) ID() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetId().GetValue()
}

// Service get service
func (i *Instance) Service() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetService().GetValue()
}

// Namespace get namespace
func (i *Instance) Namespace() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetNamespace().GetValue()
}

// VpcID get vpcid
func (i *Instance) VpcID() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetVpcId().GetValue()
}

// Host get host
func (i *Instance) Host() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetHost().GetValue()
}

// Port get port
func (i *Instance) Port() uint32 {
	if i.Proto == nil {
		return 0
	}
	return i.Proto.GetPort().GetValue()
}

// Protocol get protocol
func (i *Instance) Protocol() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetProtocol().GetValue()
}

// Version get version
func (i *Instance) Version() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetVersion().GetValue()
}

// Priority gets priority
func (i *Instance) Priority() uint32 {
	if i.Proto == nil {
		return 0
	}
	return i.Proto.GetPriority().GetValue()
}

// Weight get weight
func (i *Instance) Weight() uint32 {
	if i.Proto == nil {
		return 0
	}
	return i.Proto.GetWeight().GetValue()
}

// EnableHealthCheck get enables health check
func (i *Instance) EnableHealthCheck() bool {
	if i.Proto == nil {
		return false
	}
	return i.Proto.GetEnableHealthCheck().GetValue()
}

// HealthCheck get health check
func (i *Instance) HealthCheck() *apiservice.HealthCheck {
	if i.Proto == nil {
		return nil
	}
	return i.Proto.GetHealthCheck()
}

// Healthy get healthy
func (i *Instance) Healthy() bool {
	if i.Proto == nil {
		return false
	}
	return i.Proto.GetHealthy().GetValue()
}

// Isolate get isolate
func (i *Instance) Isolate() bool {
	if i.Proto == nil {
		return false
	}
	return i.Proto.GetIsolate().GetValue()
}

// Location gets location
func (i *Instance) Location() *apimodel.Location {
	if i.Proto == nil {
		return nil
	}
	return i.Proto.GetLocation()
}

// Metadata get metadata
func (i *Instance) Metadata() map[string]string {
	if i.Proto == nil {
		return nil
	}
	return i.Proto.GetMetadata()
}

// LogicSet get logic set
func (i *Instance) LogicSet() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetLogicSet().GetValue()
}

// Ctime get ctime
func (i *Instance) Ctime() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetCtime().GetValue()
}

// Mtime get mtime
func (i *Instance) Mtime() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetMtime().GetValue()
}

// Revision get revision
func (i *Instance) Revision() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetRevision().GetValue()
}

// ServiceToken get service token
func (i *Instance) ServiceToken() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetServiceToken().GetValue()
}

// MallocProto malloc proto if proto is null
func (i *Instance) MallocProto() {
	if i.Proto == nil {
		i.Proto = &apiservice.Instance{}
	}
}

// InstanceStore 对应store层（database）的对象
type InstanceStore struct {
	ID                string
	ServiceID         string
	Host              string
	VpcID             string
	Port              uint32
	Protocol          string
	Version           string
	HealthStatus      int
	Isolate           int
	Weight            uint32
	EnableHealthCheck int
	CheckType         int32
	TTL               uint32
	Priority          uint32
	Revision          string
	LogicSet          string
	Region            string
	Zone              string
	Campus            string
	Meta              map[string]string
	Flag              int
	CreateTime        int64
	ModifyTime        int64
}

// ExpandInstanceStore 包含服务名的store信息
type ExpandInstanceStore struct {
	ServiceName       string
	Namespace         string
	ServiceToken      string
	ServicePlatformID string
	ServiceInstance   *InstanceStore
}

// Store2Instance store的数据转换为组合了api的数据结构
func Store2Instance(is *InstanceStore) *Instance {
	ins := &Instance{
		Proto: &apiservice.Instance{
			Id:                &wrappers.StringValue{Value: is.ID},
			VpcId:             &wrappers.StringValue{Value: is.VpcID},
			Host:              &wrappers.StringValue{Value: is.Host},
			Port:              &wrappers.UInt32Value{Value: is.Port},
			Protocol:          &wrappers.StringValue{Value: is.Protocol},
			Version:           &wrappers.StringValue{Value: is.Version},
			Priority:          &wrappers.UInt32Value{Value: is.Priority},
			Weight:            &wrappers.UInt32Value{Value: is.Weight},
			EnableHealthCheck: &wrappers.BoolValue{Value: utils.Int2bool(is.EnableHealthCheck)},
			Healthy:           &wrappers.BoolValue{Value: utils.Int2bool(is.HealthStatus)},
			Location: &apimodel.Location{
				Region: &wrappers.StringValue{Value: is.Region},
				Zone:   &wrappers.StringValue{Value: is.Zone},
				Campus: &wrappers.StringValue{Value: is.Campus},
			},
			Isolate:  &wrappers.BoolValue{Value: utils.Int2bool(is.Isolate)},
			Metadata: is.Meta,
			LogicSet: &wrappers.StringValue{Value: is.LogicSet},
			Ctime:    &wrappers.StringValue{Value: utils.Int64Time2String(is.CreateTime)},
			Mtime:    &wrappers.StringValue{Value: utils.Int64Time2String(is.ModifyTime)},
			Revision: &wrappers.StringValue{Value: is.Revision},
		},
		ServiceID:  is.ServiceID,
		Valid:      flag2valid(is.Flag),
		ModifyTime: time.Unix(is.ModifyTime, 0),
	}
	// 如果不存在checkType，即checkType==-1。HealthCheck置为nil
	if is.CheckType != -1 {
		ins.Proto.HealthCheck = &apiservice.HealthCheck{
			Type: apiservice.HealthCheck_HealthCheckType(is.CheckType),
			Heartbeat: &apiservice.HeartbeatHealthCheck{
				Ttl: &wrappers.UInt32Value{Value: is.TTL},
			},
		}
	}
	// 如果location不为空，那么填充一下location
	if is.Region != "" {
		ins.Proto.Location = &apimodel.Location{
			Region: &wrappers.StringValue{Value: is.Region},
			Zone:   &wrappers.StringValue{Value: is.Zone},
			Campus: &wrappers.StringValue{Value: is.Campus},
		}
	}

	return ins
}

// ExpandStore2Instance 扩展store转换
func ExpandStore2Instance(es *ExpandInstanceStore) *Instance {
	out := Store2Instance(es.ServiceInstance)
	out.Proto.Service = &wrappers.StringValue{Value: es.ServiceName}
	out.Proto.Namespace = &wrappers.StringValue{Value: es.Namespace}
	if es.ServiceToken != "" {
		out.Proto.ServiceToken = &wrappers.StringValue{Value: es.ServiceToken}
	}
	out.ServicePlatformID = es.ServicePlatformID
	return out
}

// CreateInstanceModel 创建存储层服务实例模型
func CreateInstanceModel(serviceID string, req *apiservice.Instance) *Instance {
	// 默认为健康的
	healthy := true
	if req.GetHealthy() != nil {
		healthy = req.GetHealthy().GetValue()
	}

	// 默认为不隔离的
	isolate := false
	if req.GetIsolate() != nil {
		isolate = req.GetIsolate().GetValue()
	}

	// 权重默认是100
	var weight uint32 = 100
	if req.GetWeight() != nil {
		weight = req.GetWeight().GetValue()
	}

	instance := &Instance{
		ServiceID: serviceID,
	}

	protoIns := &apiservice.Instance{
		Id:       req.GetId(),
		Host:     utils.NewStringValue(strings.TrimSpace(req.GetHost().GetValue())),
		VpcId:    req.GetVpcId(),
		Port:     req.GetPort(),
		Protocol: req.GetProtocol(),
		Version:  req.GetVersion(),
		Priority: req.GetPriority(),
		Weight:   utils.NewUInt32Value(weight),
		Healthy:  utils.NewBoolValue(healthy),
		Isolate:  utils.NewBoolValue(isolate),
		Location: req.Location,
		Metadata: req.Metadata,
		LogicSet: req.GetLogicSet(),
		Revision: utils.NewStringValue(utils.NewUUID()), // 更新版本号
	}

	// health Check，healthCheck不能为空，且没有显示把enable_health_check置为false
	// 如果create的时候，打开了healthCheck，那么实例模式是unhealthy，必须要一次心跳才会healthy
	if req.GetHealthCheck().GetHeartbeat() != nil &&
		(req.GetEnableHealthCheck() == nil || req.GetEnableHealthCheck().GetValue()) {
		protoIns.EnableHealthCheck = utils.NewBoolValue(true)
		protoIns.HealthCheck = req.HealthCheck
		protoIns.HealthCheck.Type = apiservice.HealthCheck_HEARTBEAT
		// ttl range: (0, 60]
		ttl := protoIns.GetHealthCheck().GetHeartbeat().GetTtl().GetValue()
		if ttl == 0 || ttl > 60 {
			if protoIns.HealthCheck.Heartbeat.Ttl == nil {
				protoIns.HealthCheck.Heartbeat.Ttl = utils.NewUInt32Value(5)
			}
			protoIns.HealthCheck.Heartbeat.Ttl.Value = 5
		}
	}

	instance.Proto = protoIns
	return instance
}

// InstanceEventType 探测事件类型
type InstanceEventType string

const (
	// EventDiscoverNone empty discover event
	EventDiscoverNone InstanceEventType = "EventDiscoverNone"
	// EventInstanceOnline instance becoming online
	EventInstanceOnline InstanceEventType = "InstanceOnline"
	// EventInstanceTurnUnHealth Instance becomes unhealthy
	EventInstanceTurnUnHealth InstanceEventType = "InstanceTurnUnHealth"
	// EventInstanceTurnHealth Instance becomes healthy
	EventInstanceTurnHealth InstanceEventType = "InstanceTurnHealth"
	// EventInstanceOpenIsolate Instance is in isolation
	EventInstanceOpenIsolate InstanceEventType = "InstanceOpenIsolate"
	// EventInstanceCloseIsolate Instance shutdown isolation state
	EventInstanceCloseIsolate InstanceEventType = "InstanceCloseIsolate"
	// EventInstanceOffline Instance offline
	EventInstanceOffline InstanceEventType = "InstanceOffline"
	// EventInstanceSendHeartbeat Instance send heartbeat package to server
	EventInstanceSendHeartbeat InstanceEventType = "InstanceSendHeartbeat"
	// EventInstanceUpdate Instance metadata and info update event
	EventInstanceUpdate InstanceEventType = "InstanceUpdate"
	// EventClientOffline .
	EventClientOffline InstanceEventType = "ClientOffline"
)

// CtxEventKeyMetadata 用于将metadata从Context中传入并取出
const CtxEventKeyMetadata = "ctx_event_metadata"

// InstanceEvent 服务实例事件
type InstanceEvent struct {
	Id         string
	SvcId      string
	Namespace  string
	Service    string
	Instance   *apiservice.Instance
	EType      InstanceEventType
	CreateTime time.Time
	MetaData   map[string]string
}

// InjectMetadata 从context中获取metadata并注入到事件对象
func (i *InstanceEvent) InjectMetadata(ctx context.Context) {
	value := ctx.Value(CtxEventKeyMetadata)
	if nil == value {
		return
	}
	i.MetaData = value.(map[string]string)
}

func (i *InstanceEvent) String() string {
	if nil == i {
		return "nil"
	}
	hostPortStr := fmt.Sprintf("%s:%d", i.Instance.GetHost().GetValue(), i.Instance.GetPort().GetValue())
	return fmt.Sprintf("InstanceEvent(id=%s, namespace=%s, svcId=%s, service=%s, type=%v, instance=%s, healthy=%v)",
		i.Id, i.Namespace, i.SvcId, i.Service, i.EType, hostPortStr, i.Instance.GetHealthy().GetValue())
}

type ClientEvent struct {
	EType InstanceEventType
	Id    string
}

type ServiceInstances struct {
	lock               sync.RWMutex
	instances          map[string]*Instance
	healthyInstances   map[string]*Instance
	unhealthyInstances map[string]*Instance
	protectInstances   map[string]*Instance
	protectThreshold   float32
}

func NewServiceInstances(protectThreshold float32) *ServiceInstances {
	return &ServiceInstances{
		instances:          make(map[string]*Instance, 128),
		healthyInstances:   make(map[string]*Instance, 128),
		unhealthyInstances: make(map[string]*Instance, 128),
		protectInstances:   make(map[string]*Instance, 128),
	}
}

func (si *ServiceInstances) TotalCount() int {
	si.lock.RLock()
	defer si.lock.RUnlock()

	return len(si.instances)
}

func (si *ServiceInstances) UpdateProtectThreshold(protectThreshold float32) {
	si.lock.Lock()
	defer si.lock.Unlock()

	si.protectThreshold = protectThreshold
}

func (si *ServiceInstances) UpsertInstance(ins *Instance) {
	si.lock.Lock()
	defer si.lock.Unlock()

	si.instances[ins.ID()] = ins
	if ins.Healthy() {
		si.healthyInstances[ins.ID()] = ins
	} else {
		si.unhealthyInstances[ins.ID()] = ins
	}
}

func (si *ServiceInstances) RemoveInstance(ins *Instance) {
	si.lock.Lock()
	defer si.lock.Unlock()

	delete(si.instances, ins.ID())
	delete(si.healthyInstances, ins.ID())
	delete(si.unhealthyInstances, ins.ID())
	delete(si.protectInstances, ins.ID())
}

func (si *ServiceInstances) Range(iterator func(id string, ins *Instance)) {
	si.lock.RLock()
	defer si.lock.RUnlock()

	for k, v := range si.instances {
		iterator(k, v)
	}
}

func (si *ServiceInstances) GetInstances(onlyHealthy bool) []*Instance {
	si.lock.RLock()
	defer si.lock.RUnlock()

	ret := make([]*Instance, 0, len(si.healthyInstances)+len(si.protectInstances))
	if !onlyHealthy {
		for k, v := range si.instances {
			protectIns, ok := si.protectInstances[k]
			if ok {
				ret = append(ret, protectIns)
			} else {
				ret = append(ret, v)
			}
		}
	} else {
		for _, v := range si.healthyInstances {
			ret = append(ret, v)
		}
		for _, v := range si.protectInstances {
			ret = append(ret, v)
		}
	}
	return ret
}

func (si *ServiceInstances) ReachHealthyProtect() bool {
	si.lock.RLock()
	defer si.lock.RUnlock()

	return len(si.protectInstances) > 0
}

func (si *ServiceInstances) RunHealthyProtect() {
	si.lock.Lock()
	defer si.lock.Unlock()

	lastBeat := int64(-1)

	curProportion := float32(len(si.healthyInstances)) / float32(len(si.instances))
	if curProportion > si.protectThreshold {
		// 不会触发, 并且清空当前保护状态的实例
		si.protectInstances = make(map[string]*Instance, 128)
		return
	}
	instanceLastBeatTimes := map[string]int64{}
	instances := si.unhealthyInstances
	for i := range instances {
		ins := instances[i]
		metadata := ins.Metadata()
		if len(metadata) == 0 {
			continue
		}
		val, ok := metadata[MetadataInstanceLastHeartbeatTime]
		if !ok {
			continue
		}
		beatTime, _ := strconv.ParseInt(val, 10, 64)
		if beatTime >= lastBeat {
			lastBeat = beatTime
		}
		instanceLastBeatTimes[ins.ID()] = beatTime
	}
	if lastBeat == -1 {
		return
	}
	for i := range instances {
		ins := instances[i]
		beatTime, ok := instanceLastBeatTimes[ins.ID()]
		if !ok {
			continue
		}
		needProtect := needZeroProtect(lastBeat, beatTime, int64(ins.HealthCheck().GetHeartbeat().GetTtl().GetValue()))
		if !needProtect {
			continue
		}
		si.protectInstances[ins.ID()] = ins
	}
}

// needZeroProtect .
func needZeroProtect(lastBeat, beatTime, ttl int64) bool {
	return lastBeat-3*ttl > beatTime
}

// store的flag转换为valid
// flag==1为无效，其他情况为有效
func flag2valid(flag int) bool {
	return flag != 1
}
