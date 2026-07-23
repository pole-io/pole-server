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

package rules

import (
	"encoding/json"
	"time"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	revisionapi "github.com/pole-io/pole-server/apis/pkg/utils/revision"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
)

// ServiceWithRateLimits 与服务绑定的路由规则数据
type ServiceWithRateLimits struct {
	Service  svctypes.ServiceKey
	Rules    *container.SyncMap[string, *RateLimitRelease]
	Revision string
}

func NewServiceWithRateLimits(svcKey svctypes.ServiceKey) *ServiceWithRateLimits {
	return &ServiceWithRateLimits{
		Service:  svcKey,
		Rules:    container.NewSyncMap[string, *RateLimitRelease](),
		Revision: "",
	}
}

func (s *ServiceWithRateLimits) AddRule(rule *RateLimitRelease) {
	s.Rules.Store(rule.Key(), rule)
}

func (s *ServiceWithRateLimits) DelRule(id string) {
	s.Rules.Delete(id)
}

func (s *ServiceWithRateLimits) GetRule(id string) (*RateLimitRelease, bool) {
	rule, ok := s.Rules.Load(id)
	return rule, ok
}

func (s *ServiceWithRateLimits) Reload() {
	revision := []uint64{}
	s.Rules.Range(func(key string, value *RateLimitRelease) {
		revision = append(revision, value.Version)
	})

	s.Revision, _ = revisionapi.ComputeRevisionUint64(revision)
}

// RateLimit 限流规则
type RateLimit struct {
	Proto     *apitraffic.RateLimit
	ID        string
	Namespace string
	ServiceID string
	Name      string
	Method    string
	// Labels for old compatible, will be removed later
	Labels     string
	Priority   uint32
	Rule       string
	Revision   string
	Disable    bool
	Valid      bool
	CreateTime time.Time
	ModifyTime time.Time
	EnableTime time.Time
	Metadata   map[string]string
}

func (r *RateLimit) CopyNoProto() *RateLimit {
	return &RateLimit{
		ID:         r.ID,
		Namespace:  r.Namespace,
		ServiceID:  r.ServiceID,
		Name:       r.Name,
		Method:     r.Method,
		Labels:     r.Labels,
		Proto:      r.Proto,
		Priority:   r.Priority,
		Rule:       r.Rule,
		Revision:   r.Revision,
		Disable:    r.Disable,
		Valid:      r.Valid,
		CreateTime: r.CreateTime,
		ModifyTime: r.ModifyTime,
		EnableTime: r.EnableTime,
		Metadata:   r.Metadata,
	}
}

func (r *RateLimit) GetId() string {
	return r.ID
}

func (r *RateLimit) GetMtime() time.Time {
	return r.ModifyTime
}

func (r *RateLimit) ToSpec() error {
	r.Proto = &apitraffic.RateLimit{}
	if len(r.Rule) == 0 {
		SetOwnerNamespaceOnProto(r.Proto, r.Namespace)
		return nil
	}
	// 反序列化rule
	if err := json.Unmarshal([]byte(r.Rule), r.Proto); err != nil {
		return err
	}
	r.Proto.Disable = r.Disable
	SetOwnerNamespaceOnProto(r.Proto, r.Namespace)
	return nil
}

const (
	LabelKeyPath          = "$path"
	LabelKeyMethod        = "$method"
	LabelKeyHeader        = "$header"
	LabelKeyQuery         = "$query"
	LabelKeyCallerService = "$caller_service"
	LabelKeyCallerIP      = "$caller_ip"
)

// Arguments2Labels 将参数列表适配成旧的标签模型
func Arguments2Labels(arguments []*apitraffic.MatchArgument) map[string]*apimodel.MatchString {
	if len(arguments) > 0 {
		var labels = make(map[string]*apimodel.MatchString)
		for _, argument := range arguments {
			key := BuildArgumentKey(argument.Type, argument.Key)
			labels[key] = argument.Value
		}
		return labels
	}
	return nil
}

func BuildArgumentKey(argumentType apitraffic.MatchArgument_Type, key string) string {
	switch argumentType {
	case apitraffic.MatchArgument_HEADER:
		return LabelKeyHeader + "." + key
	case apitraffic.MatchArgument_QUERY:
		return LabelKeyQuery + "." + key
	case apitraffic.MatchArgument_CALLER_SERVICE:
		return LabelKeyCallerService + "." + key
	case apitraffic.MatchArgument_CALLER_IP:
		return LabelKeyCallerIP
	case apitraffic.MatchArgument_CUSTOM:
		return key
	case apitraffic.MatchArgument_METHOD:
		return LabelKeyMethod
	default:
		return key
	}
}

// ExtendRateLimit 包含服务信息的限流规则
type ExtendRateLimit struct {
	ServiceName   string
	NamespaceName string
	RateLimit     *RateLimit
}

// RateLimitRevision 包含最新版本号的限流规则
type RateLimitRevision struct {
	ServiceID    string
	LastRevision string
	ModifyTime   time.Time
}
