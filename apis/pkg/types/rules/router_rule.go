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
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	protoV2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apitraffic "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"

	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	"github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/pkg/utils"
)

type TrafficDirection string

const (
	TrafficDirection_INBOUND  TrafficDirection = "TrafficDirection_INBOUND"
	TrafficDirection_OUTBOUND TrafficDirection = "TrafficDirection_OUTBOUND"
)

const (
	// V2RuleIDKey v2 版本的规则路由 ID
	V2RuleIDKey = "__routing_v2_id__"
	// V2RuleIDPriority v2 版本的规则路由优先级
	V2RuleIDPriority = "__routing_v2_priority__"
	// V1RuleInRoute inBound 类型
	V1RuleInRoute = "in"
	// V1RuleOutRoute outBound 类型
	V1RuleOutRoute = "out"
)

var (
	// RuleRoutingTypeUrl 记录 anypb.Any 中关于 RuleRoutingConfig 的 url 信息
	RuleRoutingTypeUrl string
	// MetaRoutingTypeUrl 记录 anypb.Any 中关于 MetadataRoutingConfig 的 url 信息
	MetaRoutingTypeUrl string
	// NearbyRoutingTypeUrl 记录 anypb.Any 中关于 NearbyRoutingConfig 的 url 信息
	NearbyRoutingTypeUrl string
)

func init() {
	ruleAny, _ := ptypes.MarshalAny(&apitraffic.RuleRoutingConfig{})
	metaAny, _ := ptypes.MarshalAny(&apitraffic.MetadataRoutingConfig{})
	nearbyAny, _ := ptypes.MarshalAny(&apitraffic.NearbyRoutingConfig{})

	RuleRoutingTypeUrl = ruleAny.GetTypeUrl()
	MetaRoutingTypeUrl = metaAny.GetTypeUrl()
	NearbyRoutingTypeUrl = nearbyAny.GetTypeUrl()
}

/*
 * RoutingConfig 路由配置
 */
type RoutingConfig struct {
	ID         string
	InBounds   string
	OutBounds  string
	Revision   string
	Valid      bool
	CreateTime time.Time
	ModifyTime time.Time
}

// ExtendRoutingConfig 路由配置的扩展结构体
type ExtendRoutingConfig struct {
	ServiceName   string
	NamespaceName string
	Config        *RoutingConfig
}

// ExtendRouterConfig 路由信息的扩展
type ExtendRouterConfig struct {
	*RouterConfig
	// MetadataRouting 元数据路由配置
	MetadataRouting *apitraffic.MetadataRoutingConfig
	// RuleRouting 规则路由配置
	RuleRouting *RuleRoutingConfigWrapper
	// NearbyRouting 就近路由规则数据
	NearbyRouting *apitraffic.NearbyRoutingConfig
	// Metadata .
	Metadata map[string]string
}

// ToApi Turn to API object
func (r *ExtendRouterConfig) ToApi() (*apitraffic.RouteRule, error) {
	var (
		anyValue *anypb.Any
		err      error
	)

	switch r.GetRoutingPolicy() {
	case apitraffic.RoutingPolicy_NearbyPolicy:
		anyValue, err = ptypes.MarshalAny(r.NearbyRouting)
		if err != nil {
			return nil, err
		}
	case apitraffic.RoutingPolicy_MetadataPolicy:
		anyValue, err = ptypes.MarshalAny(r.MetadataRouting)
		if err != nil {
			return nil, err
		}
	default:
		anyValue, err = ptypes.MarshalAny(r.RuleRouting.RuleRouting)
		if err != nil {
			return nil, err
		}
	}

	rule := &apitraffic.RouteRule{
		Id:            r.ID,
		Name:          r.Name,
		Namespace:     r.Namespace,
		Enable:        r.Enable,
		RoutingPolicy: r.GetRoutingPolicy(),
		RoutingConfig: anyValue,
		Revision:      r.Revision,
		Ctime:         utils.Time2String(r.CreateTime),
		Mtime:         utils.Time2String(r.ModifyTime),
		Etime:         utils.Time2String(r.EnableTime),
		Priority:      r.Priority,
		Description:   r.Description,
		Editable:      true,
		Deleteable:    true,
	}
	if r.EnableTime.Year() > 2000 {
		rule.Etime = utils.Time2String(r.EnableTime)
	} else {
		rule.Etime = ""
	}
	return rule, nil
}

type RuleRoutingConfigWrapper struct {
	Caller service.ServiceKey
	Callee service.ServiceKey
	// RuleRouting 规则路由配置
	RuleRouting *apitraffic.RuleRoutingConfig
}

// RouterConfig Routing rules
type RouterConfig struct {
	// ID The unique id of the rules
	ID string `json:"id"`
	// namespace router config owner namespace
	Namespace string `json:"namespace"`
	// name router config name
	Name string `json:"name"`
	// policy Rules
	Policy string `json:"policy"`
	// config Specific routing rules content
	Config string `json:"config"`
	// enable Whether the routing rules are enabled
	Enable bool `json:"enable"`
	// priority Rules priority
	Priority uint32 `json:"priority"`
	// revision Edition information of routing rules
	Revision string `json:"revision"`
	// Description Simple description of rules
	Description string `json:"description"`
	// valid Whether the routing rules are valid and have not been deleted by logic
	Valid bool `json:"flag"`
	// createtime Rules creation time
	CreateTime time.Time `json:"ctime"`
	// modifytime Rules modify time
	ModifyTime time.Time `json:"mtime"`
	// enabletime The last time the rules enabled
	EnableTime time.Time `json:"etime"`
	// Metadata.
	Metadata map[string]string `json:"metadata"`
}

// GetRoutingPolicy Query routing rules type
func (r *RouterConfig) GetRoutingPolicy() apitraffic.RoutingPolicy {
	v, ok := apitraffic.RoutingPolicy_value[r.Policy]

	if !ok {
		return apitraffic.RoutingPolicy(-1)
	}

	return apitraffic.RoutingPolicy(v)
}

// ToExpendRoutingConfig Converted to an expansion object, serialize the corresponding PB Struct in advance
func (r *RouterConfig) ToExpendRoutingConfig() (*ExtendRouterConfig, error) {
	ret := &ExtendRouterConfig{
		RouterConfig: r,
	}

	configText := r.Config
	if len(configText) == 0 {
		return ret, nil
	}
	policy := r.GetRoutingPolicy()
	var err error
	if strings.HasPrefix(configText, "{") {
		// process with json
		switch policy {
		case apitraffic.RoutingPolicy_RulePolicy:
			rule := &apitraffic.RuleRoutingConfig{}
			if err = utils.UnmarshalFromJsonString(rule, configText); nil != err {
				return nil, err
			}
			ret.RuleRouting = parseSubRouteRule(rule)
			break
		case apitraffic.RoutingPolicy_MetadataPolicy:
			rule := &apitraffic.MetadataRoutingConfig{}
			if err = utils.UnmarshalFromJsonString(rule, configText); nil != err {
				return nil, err
			}
			ret.MetadataRouting = rule
			break
		case apitraffic.RoutingPolicy_NearbyPolicy:
			rule := &apitraffic.NearbyRoutingConfig{}
			if err = utils.UnmarshalFromJsonString(rule, configText); nil != err {
				return nil, err
			}
			ret.NearbyRouting = rule
			break
		}
		return ret, nil
	}

	if err := r.parseBinaryAnyMessage(policy, ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (r *RouterConfig) parseBinaryAnyMessage(
	policy apitraffic.RoutingPolicy, ret *ExtendRouterConfig) error {
	// parse v1 binary
	switch policy {
	case apitraffic.RoutingPolicy_RulePolicy:
		rule := &apitraffic.RuleRoutingConfig{}
		anyMsg := &anypb.Any{
			TypeUrl: RuleRoutingTypeUrl,
			Value:   []byte(r.Config),
		}
		if err := unmarshalToAny(anyMsg, rule); nil != err {
			return err
		}
		ret.RuleRouting = parseSubRouteRule(rule)
	case apitraffic.RoutingPolicy_MetadataPolicy:
		rule := &apitraffic.MetadataRoutingConfig{}
		anyMsg := &anypb.Any{
			TypeUrl: MetaRoutingTypeUrl,
			Value:   []byte(r.Config),
		}
		if err := unmarshalToAny(anyMsg, rule); nil != err {
			return err
		}
		ret.MetadataRouting = rule
	case apitraffic.RoutingPolicy_NearbyPolicy:
		rule := &apitraffic.NearbyRoutingConfig{}
		anyMsg := &anypb.Any{
			TypeUrl: NearbyRoutingTypeUrl,
			Value:   []byte(r.Config),
		}
		if err := unmarshalToAny(anyMsg, rule); nil != err {
			return err
		}
		ret.NearbyRouting = rule
	}
	return nil
}

// ParseRouteRuleFromAPI Convert an internal object from the API object
func (r *RouterConfig) ParseRouteRuleFromAPI(routing *apitraffic.RouteRule) error {
	ruleMessage, err := ParseRouteRuleAnyToMessage(routing.RoutingPolicy, routing.RoutingConfig)
	if nil != err {
		return err
	}

	if r.Config, err = utils.MarshalToJsonString(ruleMessage); nil != err {
		return err
	}
	r.ID = routing.Id
	r.Revision = routing.Revision
	r.Name = routing.Name
	r.Namespace = routing.Namespace
	r.Enable = routing.Enable
	r.Policy = routing.GetRoutingPolicy().String()
	r.Priority = routing.Priority
	r.Description = routing.Description
	r.Metadata = routing.Metadata

	// Priority range range [0, 10]
	if r.Priority > 10 {
		r.Priority = 10
	}

	return nil
}

func unmarshalToAny(anyMessage *anypb.Any, message proto.Message) error {
	return anypb.UnmarshalTo(anyMessage, proto.MessageV2(message),
		protoV2.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true})
}

// ParseRouteRuleAnyToMessage convert the any routing proto to message object
func ParseRouteRuleAnyToMessage(policy apitraffic.RoutingPolicy, anyMessage *anypb.Any) (proto.Message, error) {
	var rule proto.Message
	switch policy {
	case apitraffic.RoutingPolicy_RulePolicy:
		rule = &apitraffic.RuleRoutingConfig{}
		if err := unmarshalToAny(anyMessage, rule); err != nil {
			return nil, err
		}
		ruleRouting := rule.(*apitraffic.RuleRoutingConfig)
		parseSubRouteRule(ruleRouting)
		break
	case apitraffic.RoutingPolicy_MetadataPolicy:
		rule = &apitraffic.MetadataRoutingConfig{}
		if err := unmarshalToAny(anyMessage, rule); err != nil {
			return nil, err
		}
		break
	case apitraffic.RoutingPolicy_NearbyPolicy:
		rule = &apitraffic.NearbyRoutingConfig{}
		if err := unmarshalToAny(anyMessage, rule); err != nil {
			return nil, err
		}
		break
	}
	return rule, nil
}

func parseSubRouteRule(ruleRouting *apitraffic.RuleRoutingConfig) *RuleRoutingConfigWrapper {
	if len(ruleRouting.Rules) == 0 {
		subRule := &apitraffic.SubRuleRouting{
			Name:         "",
			Sources:      ruleRouting.GetSources(),
			Destinations: ruleRouting.GetDestinations(),
		}
		ruleRouting.Rules = []*apitraffic.SubRuleRouting{
			subRule,
		}
	} else {
		for i := range ruleRouting.Rules {
			subRule := ruleRouting.Rules[i]
			if len(subRule.Sources) == 0 {
				subRule.Sources = ruleRouting.GetSources()
			}
			if len(subRule.Destinations) == 0 {
				subRule.Destinations = ruleRouting.GetDestinations()
			}
		}
		// Abandon the value of the old field
		ruleRouting.Destinations = nil
		ruleRouting.Sources = nil
	}

	wrapper := &RuleRoutingConfigWrapper{
		RuleRouting: ruleRouting,
	}

	for i := range ruleRouting.Rules {
		item := ruleRouting.Rules[i]
		if len(item.Sources) != 0 {
			source := item.Sources[0]
			wrapper.Caller = service.ServiceKey{
				Namespace: source.Namespace,
				Name:      source.Service,
			}
		}
		if len(item.Destinations) != 0 {
			destination := item.Destinations[0]
			wrapper.Callee = service.ServiceKey{
				Namespace: destination.Namespace,
				Name:      destination.Service,
			}
		}
	}
	return wrapper
}

const (
	_labelKeyPath     = "$path"
	_labelKeyMethod   = "$method"
	_labelKeyHeader   = "$header"
	_labelKeyQuery    = "$query"
	_labelKeyCallerIP = "$caller_ip"
	_labelKeyCookie   = "$cookie"

	MatchAll = "*"
)

// RoutingLabels2Arguments Adapting the old label model into a list of parameters
func RoutingLabels2Arguments(labels map[string]*apimodel.MatchString) []*apitraffic.SourceMatch {
	if len(labels) == 0 {
		return []*apitraffic.SourceMatch{}
	}

	arguments := make([]*apitraffic.SourceMatch, 0, 4)
	for index := range labels {
		arguments = append(arguments, &apitraffic.SourceMatch{
			Type: apitraffic.SourceMatch_CUSTOM,
			Key:  index,
			Value: &apimodel.MatchString{
				Type:      labels[index].GetType(),
				Value:     labels[index].GetValue(),
				ValueType: labels[index].GetValueType(),
			},
		})
	}

	return arguments
}

// RoutingArguments2Labels Adapt the parameter list to the old label model
func RoutingArguments2Labels(args []*apitraffic.SourceMatch) map[string]*apimodel.MatchString {
	labels := make(map[string]*apimodel.MatchString)
	for i := range args {
		argument := args[i]
		var key string
		switch argument.Type {
		case apitraffic.SourceMatch_CUSTOM:
			key = argument.Key
		case apitraffic.SourceMatch_METHOD:
			key = _labelKeyMethod
		case apitraffic.SourceMatch_HEADER:
			key = _labelKeyHeader + "." + argument.Key
		case apitraffic.SourceMatch_QUERY:
			key = _labelKeyQuery + "." + argument.Key
		case apitraffic.SourceMatch_CALLER_IP:
			key = _labelKeyCallerIP
		case apitraffic.SourceMatch_COOKIE:
			key = _labelKeyCookie + "." + argument.Key
		case apitraffic.SourceMatch_PATH:
			key = _labelKeyPath
		default:
			continue
		}

		labels[key] = &apimodel.MatchString{
			Type:      argument.GetValue().GetType(),
			Value:     argument.GetValue().GetValue(),
			ValueType: argument.GetValue().GetValueType(),
		}
	}

	return labels
}

// CompareRoutingV2 Compare the priority of two routing.
func CompareRoutingV2(a, b *ExtendRouterConfig) bool {
	if a.Priority != b.Priority {
		return a.Priority < b.Priority
	}
	// 如果优先级相同，则比较规则 ID
	return a.ID < b.ID
}

// CompareRoutingV1 Compare the priority of two routing.
func CompareRoutingV1(a, b *apitraffic.Route) bool {
	ap := a.ExtendInfo[V2RuleIDPriority]
	bp := b.ExtendInfo[V2RuleIDPriority]
	return ap < bp
}

func BuildRoutes(item *ExtendRouterConfig, direction TrafficDirection) []*apitraffic.Route {
	switch direction {
	case TrafficDirection_INBOUND:
		return BuildInBoundsRoute(item)
	default:
		return BuildOutBoundsRoutes(item)
	}
}

// BuildInBoundsRoute Convert the routing rules of V2 to the inbounds in the routing rule of V1
func BuildInBoundsRoute(item *ExtendRouterConfig) []*apitraffic.Route {
	if item.GetRoutingPolicy() != apitraffic.RoutingPolicy_RulePolicy {
		return []*apitraffic.Route{}
	}

	routes := make([]*apitraffic.Route, 0, 8)

	specRules := item.RuleRouting.RuleRouting.Rules

	for i := range specRules {
		subRule := specRules[i]
		destinations := specRules[i].Destinations
		v1destinations := make([]*apitraffic.Destination, 0, len(destinations))
		for i := range destinations {
			name := fmt.Sprintf("%s.%s.%s", item.Name, subRule.Name, destinations[i].Name)
			entry := &apitraffic.Destination{
				Name:      protobuf.NewStringValue(name),
				Service:   protobuf.NewStringValue(item.RuleRouting.Callee.Name),
				Namespace: protobuf.NewStringValue(item.RuleRouting.Callee.Namespace),
				Priority:  protobuf.NewUInt32Value(destinations[i].GetPriority()),
				Weight:    protobuf.NewUInt32Value(destinations[i].GetWeight()),
				Transfer:  protobuf.NewStringValue(destinations[i].GetTransfer()),
				Isolate:   protobuf.NewBoolValue(destinations[i].GetIsolate()),
			}

			v1labels := make(map[string]*apimodel.MatchString)
			v2labels := destinations[i].GetLabels()
			for index := range v2labels {
				v1labels[index] = &apimodel.MatchString{
					Type:      v2labels[index].GetType(),
					Value:     v2labels[index].GetValue(),
					ValueType: v2labels[index].GetValueType(),
				}
			}

			entry.Metadata = v1labels
			v1destinations = append(v1destinations, entry)
		}

		sources := specRules[i].Sources
		v1sources := make([]*apitraffic.Source, 0, len(sources))
		for i := range sources {
			entry := &apitraffic.Source{
				Service:   protobuf.NewStringValue(sources[i].Service),
				Namespace: protobuf.NewStringValue(sources[i].Namespace),
			}

			entry.Metadata = RoutingArguments2Labels(sources[i].GetArguments())
			v1sources = append(v1sources, entry)
		}

		routes = append(routes, &apitraffic.Route{
			Sources:      v1sources,
			Destinations: v1destinations,
			ExtendInfo: map[string]string{
				V2RuleIDKey:      item.ID,
				V2RuleIDPriority: fmt.Sprintf("%04d", item.Priority),
			},
		})
	}

	return routes
}

// BuildOutBoundsRoutes According to the routing rules of the V2 version, it is adapted to the
// outbounds in the routing rule of V1 version
func BuildOutBoundsRoutes(item *ExtendRouterConfig) []*apitraffic.Route {
	if item.GetRoutingPolicy() != apitraffic.RoutingPolicy_RulePolicy {
		return []*apitraffic.Route{}
	}

	routes := make([]*apitraffic.Route, 0, 8)

	specRules := item.RuleRouting.RuleRouting.Rules

	for i := range specRules {
		subRule := specRules[i]
		sources := specRules[i].Sources
		v1sources := make([]*apitraffic.Source, 0, len(sources))
		for i := range sources {
			entry := &apitraffic.Source{
				Service:   protobuf.NewStringValue(item.RuleRouting.Caller.Name),
				Namespace: protobuf.NewStringValue(item.RuleRouting.Caller.Namespace),
			}
			entry.Metadata = RoutingArguments2Labels(sources[i].GetArguments())
			v1sources = append(v1sources, entry)
		}

		destinations := specRules[i].Destinations
		v1destinations := make([]*apitraffic.Destination, 0, len(destinations))
		for i := range destinations {
			name := fmt.Sprintf("%s.%s.%s", item.Name, subRule.Name, destinations[i].Name)
			entry := &apitraffic.Destination{
				Name:      protobuf.NewStringValue(name),
				Service:   protobuf.NewStringValue(destinations[i].Service),
				Namespace: protobuf.NewStringValue(destinations[i].Namespace),
				Priority:  protobuf.NewUInt32Value(destinations[i].GetPriority()),
				Weight:    protobuf.NewUInt32Value(destinations[i].GetWeight()),
				Transfer:  protobuf.NewStringValue(destinations[i].GetTransfer()),
				Isolate:   protobuf.NewBoolValue(destinations[i].GetIsolate()),
			}

			v1labels := make(map[string]*apimodel.MatchString)
			v2labels := destinations[i].GetLabels()
			for index := range v2labels {
				v1labels[index] = &apimodel.MatchString{
					Type:      v2labels[index].GetType(),
					Value:     v2labels[index].GetValue(),
					ValueType: v2labels[index].GetValueType(),
				}
			}

			entry.Metadata = v1labels
			v1destinations = append(v1destinations, entry)
		}

		routes = append(routes, &apitraffic.Route{
			Sources:      v1sources,
			Destinations: v1destinations,
			ExtendInfo: map[string]string{
				V2RuleIDKey: item.ID,
			},
		})
	}

	return routes
}
