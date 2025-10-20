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

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
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
	return i.Proto.GetId()
}

// Service get service
func (i *Instance) Service() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetService()
}

// Namespace get namespace
func (i *Instance) Namespace() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetNamespace()
}

// Host get host
func (i *Instance) Host() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetHost()
}

// Port get port
func (i *Instance) Port() uint32 {
	if i.Proto == nil {
		return 0
	}
	return i.Proto.GetPort()
}

// Protocol get protocol
func (i *Instance) Protocol() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetProtocol()
}

// Version get version
func (i *Instance) Version() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetVersion()
}

// Priority gets priority
func (i *Instance) Priority() uint32 {
	if i.Proto == nil {
		return 0
	}
	return i.Proto.GetPriority()
}

// Weight get weight
func (i *Instance) Weight() uint32 {
	if i.Proto == nil {
		return 0
	}
	return i.Proto.GetWeight()
}

// EnableHealthCheck get enables health check
func (i *Instance) EnableHealthCheck() bool {
	if i.Proto == nil {
		return false
	}
	return i.Proto.GetEnableHealthCheck()
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
	return i.Proto.GetHealthy()
}

// Isolate get isolate
func (i *Instance) Isolate() bool {
	if i.Proto == nil {
		return false
	}
	return i.Proto.GetIsolate()
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

// Ctime get ctime
func (i *Instance) Ctime() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetCtime()
}

// Mtime get mtime
func (i *Instance) Mtime() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetMtime()
}

// Revision get revision
func (i *Instance) Revision() string {
	if i.Proto == nil {
		return ""
	}
	return i.Proto.GetRevision()
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
			Id:                is.ID,
			Host:              is.Host,
			Port:              is.Port,
			Protocol:          is.Protocol,
			Version:           is.Version,
			Priority:          is.Priority,
			Weight:            is.Weight,
			EnableHealthCheck: utils.Int2bool(is.EnableHealthCheck),
			Healthy:           utils.Int2bool(is.HealthStatus),
			Location: &apimodel.Location{
				Region: is.Region,
				Zone:   is.Zone,
				Campus: is.Campus,
			},
			Isolate:  utils.Int2bool(is.Isolate),
			Metadata: is.Meta,
			Ctime:    utils.Int64Time2String(is.CreateTime),
			Mtime:    utils.Int64Time2String(is.ModifyTime),
			Revision: is.Revision,
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
				Ttl: is.TTL,
			},
		}
	}
	// 如果location不为空，那么填充一下location
	if is.Region != "" {
		ins.Proto.Location = &apimodel.Location{
			Region: is.Region,
			Zone:   is.Zone,
			Campus: is.Campus,
		}
	}

	return ins
}

// ExpandStore2Instance 扩展store转换
func ExpandStore2Instance(es *ExpandInstanceStore) *Instance {
	out := Store2Instance(es.ServiceInstance)
	out.Proto.Service = es.ServiceName
	out.Proto.Namespace = es.Namespace
	out.ServicePlatformID = es.ServicePlatformID
	return out
}

// CreateInstanceModel 创建存储层服务实例模型
func CreateInstanceModel(serviceID string, req *apiservice.Instance) *Instance {
	// 权重默认是100
	var weight uint32 = 100
	if req.GetWeight() != 0 {
		weight = req.GetWeight()
	}

	instance := &Instance{
		ServiceID: serviceID,
	}

	protoIns := &apiservice.Instance{
		Id:       req.GetId(),
		Host:     strings.TrimSpace(req.GetHost()),
		Port:     req.GetPort(),
		Protocol: req.GetProtocol(),
		Version:  req.GetVersion(),
		Priority: req.GetPriority(),
		Weight:   weight,
		Healthy:  req.GetHealthy(),
		Isolate:  req.GetIsolate(),
		Location: req.Location,
		Metadata: req.Metadata,
		Revision: utils.NewUUID(), // 更新版本号
	}

	// health Check，healthCheck不能为空，且没有显示把enable_health_check置为false
	// 如果create的时候，打开了healthCheck，那么实例模式是unhealthy，必须要一次心跳才会healthy
	if req.GetHealthCheck().GetHeartbeat() != nil && req.GetEnableHealthCheck() {
		protoIns.EnableHealthCheck = true
		protoIns.HealthCheck = req.HealthCheck
		protoIns.HealthCheck.Type = apiservice.HealthCheck_HEARTBEAT
		// ttl range: (0, 60]
		ttl := protoIns.GetHealthCheck().GetHeartbeat().GetTtl()
		if ttl == 0 || ttl > 60 {
			if protoIns.HealthCheck.Heartbeat.Ttl == 0 {
				protoIns.HealthCheck.Heartbeat.Ttl = uint32(5)
			}
		}
	}

	instance.Proto = protoIns
	return instance
}

// InstanceEventType 探测事件类型
type (
	InstanceEventType string
	ServiceEventType  string
)

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

const (
	// EventServiceCloseEmptyPushProtect 服务实例推空保护开启
	EventServiceOpenEmptyPushProtect ServiceEventType = "ServiceOpenEmptyPushProtect"
	// EventServiceCloseEmptyPushProtect 服务实例推空保护关闭
	EventServiceCloseEmptyPushProtect ServiceEventType = "ServiceCloseEmptyPushProtect"
	// EventServiceExpireEmptyPushProtect 服务实例推空保护开关过期
	EventServiceExpireEmptyPushProtect ServiceEventType = "ServiceExpireEmptyPushProtect"
)

// CtxEventKeyMetadata 用于将metadata从Context中传入并取出
const CtxEventKeyMetadata = "ctx_event_metadata"

// ServiceEvent 服务事件
type ServiceEvent struct {
	Id         string
	Namespace  string
	Service    string
	EType      ServiceEventType
	CreateTime time.Time
	MetaData   map[string]string
}

// 资源 ID
func (i *ServiceEvent) ID() string {
	return i.Id
}

// 事件类型
func (i *ServiceEvent) Event() string {
	return string(i.EType)
}

// 资源信息
func (i *ServiceEvent) Resource() string {
	return fmt.Sprintf("%s/%s", i.Namespace, i.Service)
}

// 发生时间
func (i *ServiceEvent) HappenTime() time.Time {
	return i.CreateTime
}

func (i *ServiceEvent) String() string {
	if nil == i {
		return "nil"
	}
	return fmt.Sprintf("ServiceEvent(id=%s, namespace=%s, service=%s, type=%v)", i.Id, i.Namespace, i.Service, i.EType)
}

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

// 资源 ID
func (i *InstanceEvent) ID() string {
	return i.Id
}

// 事件类型
func (i *InstanceEvent) Event() string {
	return string(i.EType)
}

// 资源信息
func (i *InstanceEvent) Resource() string {
	hostPortStr := fmt.Sprintf("%s:%d", i.Instance.GetHost(), i.Instance.GetPort())
	return fmt.Sprintf("%s/%s/%s", i.Namespace, i.Service, hostPortStr)
}

// 发生时间
func (i *InstanceEvent) HappenTime() time.Time {
	return i.CreateTime
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
	hostPortStr := fmt.Sprintf("%s:%d", i.Instance.GetHost(), i.Instance.GetPort())
	return fmt.Sprintf("InstanceEvent(id=%s, namespace=%s, svcId=%s, service=%s, type=%v, instance=%s, healthy=%v)",
		i.Id, i.Namespace, i.SvcId, i.Service, i.EType, hostPortStr, i.Instance.GetHealthy())
}

type ClientEvent struct {
	EType InstanceEventType
	Id    string
}

type ServiceInstances struct {
	lock sync.RWMutex
	// instances 全部实例
	instances map[string]*Instance
	// healthyInstances 健康的实例
	healthyInstances map[string]struct{}
	// unhealthyInstances 不健康的实例
	unhealthyInstances map[string]struct{}
	// protectInstances 被健康实例阈值保护而修改状态的实例
	protectInstances map[string]*Instance
	// protectThreshold 健康实例保护阈值
	protectThreshold float32
	// emptyPushProtectThreshold 推空保护时间
	emptyPushProtectThreshold time.Time
	// openEmptyProtect 开启实例推空保护
	openEmptyPushProtect bool
}

func NewServiceInstances(protectThreshold float32) *ServiceInstances {
	return &ServiceInstances{
		instances:          make(map[string]*Instance, 128),
		healthyInstances:   make(map[string]struct{}, 128),
		unhealthyInstances: make(map[string]struct{}, 128),
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
		si.healthyInstances[ins.ID()] = struct{}{}
	} else {
		si.unhealthyInstances[ins.ID()] = struct{}{}
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

func (si *ServiceInstances) GetInstances(onlyHealthy bool, consumer func(*Instance)) {
	si.lock.RLock()
	defer si.lock.RUnlock()

	if !onlyHealthy {
		for k, v := range si.instances {
			protectIns, ok := si.protectInstances[k]
			if ok {
				consumer(protectIns)
			} else {
				consumer(v)
			}
		}
	} else {
		for k := range si.healthyInstances {
			consumer(si.instances[k])
		}
		for _, v := range si.protectInstances {
			consumer(v)
		}
	}
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
	for i := range si.unhealthyInstances {
		ins := si.instances[i]
		metadata := ins.Metadata()
		if len(metadata) == 0 {
			continue
		}
		val, ok := metadata[types.MetadataInstanceLastHeartbeatTime]
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
	for i := range si.unhealthyInstances {
		ins := si.instances[i]
		beatTime, ok := instanceLastBeatTimes[ins.ID()]
		if !ok {
			continue
		}
		needProtect := needZeroProtect(lastBeat, beatTime, int64(ins.HealthCheck().GetHeartbeat().GetTtl()))
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
