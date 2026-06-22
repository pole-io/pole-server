package rules

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	commontime "github.com/pole-io/pole-server/pkg/common/utils/time"
)

// CircuitBreaker 熔断规则
type CircuitBreaker struct {
	ID         string
	Version    string
	Name       string
	Namespace  string
	Business   string
	Department string
	Comment    string
	Inbounds   string
	Outbounds  string
	Token      string
	Owner      string
	Revision   string
	Valid      bool
	CreateTime time.Time
	ModifyTime time.Time
}

func (r *CircuitBreaker) GetId() string {
	return r.ID
}

func (r *CircuitBreaker) GetMtime() time.Time {
	return r.ModifyTime
}

// ServiceWithCircuitBreaker 与服务关系绑定的熔断规则
type ServiceWithCircuitBreaker struct {
	ServiceID      string
	CircuitBreaker *CircuitBreaker
	Valid          bool
	CreateTime     time.Time
	ModifyTime     time.Time
}

// ServiceWithCircuitBreakerRules 与服务关系绑定的熔断规则
type ServiceWithCircuitBreakerRules struct {
	mutex               sync.RWMutex
	Service             service.ServiceKey
	circuitBreakerRules map[string]*CircuitBreakerRelease
	Revision            string
}

func NewServiceWithCircuitBreakerRules(svcKey service.ServiceKey) *ServiceWithCircuitBreakerRules {
	return &ServiceWithCircuitBreakerRules{
		Service:             svcKey,
		circuitBreakerRules: make(map[string]*CircuitBreakerRelease),
	}
}

func (s *ServiceWithCircuitBreakerRules) AddCircuitBreakerRule(rule *CircuitBreakerRelease) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.circuitBreakerRules[rule.Key()] = rule
}

func (s *ServiceWithCircuitBreakerRules) DelCircuitBreakerRule(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.circuitBreakerRules, id)
}

func (s *ServiceWithCircuitBreakerRules) IterateCircuitBreakerRules(callback func(*CircuitBreakerRelease)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for _, rule := range s.circuitBreakerRules {
		callback(rule)
	}
}

func (s *ServiceWithCircuitBreakerRules) CountCircuitBreakerRules() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.circuitBreakerRules)
}

func (s *ServiceWithCircuitBreakerRules) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.circuitBreakerRules = make(map[string]*CircuitBreakerRelease)
	s.Revision = ""
}

// ServiceWithFaultDetectRules 与服务关系绑定的探测规则
type ServiceWithFaultDetectRules struct {
	Service          service.ServiceKey
	faultDetectRules *container.SyncMap[string, *FaultDetectRelease]
	Revision         string
}

func NewServiceWithFaultDetectRules(svcKey service.ServiceKey) *ServiceWithFaultDetectRules {
	return &ServiceWithFaultDetectRules{
		Service:          svcKey,
		faultDetectRules: container.NewSyncMap[string, *FaultDetectRelease](),
	}
}

func (s *ServiceWithFaultDetectRules) AddFaultDetectRule(rule *FaultDetectRelease) {
	s.faultDetectRules.Store(rule.Key(), rule)
}

func (s *ServiceWithFaultDetectRules) DelFaultDetectRule(id string) {
	s.faultDetectRules.Delete(id)
}

func (s *ServiceWithFaultDetectRules) IterateFaultDetectRules(callback func(*FaultDetectRelease)) {
	s.faultDetectRules.Range(func(key string, rule *FaultDetectRelease) {
		callback(rule)
	})
}

func (s *ServiceWithFaultDetectRules) CountFaultDetectRules() int {
	return s.faultDetectRules.Len()
}

func (s *ServiceWithFaultDetectRules) Clear() {
	s.faultDetectRules = container.NewSyncMap[string, *FaultDetectRelease]()
	s.Revision = ""
}

// CircuitBreakerDetail 返回给控制台的熔断规则及服务数据
type CircuitBreakerDetail struct {
	Total               uint32
	CircuitBreakerInfos []*CircuitBreakerInfo
}

// CircuitBreakerInfo 熔断规则及绑定服务
type CircuitBreakerInfo struct {
	CircuitBreaker *CircuitBreaker
	Services       []*service.Service
}

// CircuitBreakerRule 熔断规则
type CircuitBreakerRule struct {
	Proto        *apifault.CircuitBreakerRule
	ID           string
	Name         string
	Namespace    string
	Description  string
	Level        int
	SrcService   string
	SrcNamespace string
	DstService   string
	DstNamespace string
	DstMethod    string
	Rule         string
	Revision     string
	Enable       bool
	Valid        bool
	CreateTime   time.Time
	ModifyTime   time.Time
	EnableTime   time.Time
}

func (r *CircuitBreakerRule) GetId() string {
	return r.ID
}

func (r *CircuitBreakerRule) GetMtime() time.Time {
	return r.ModifyTime
}

func (c *CircuitBreakerRule) IsServiceChange(other *CircuitBreakerRule) bool {
	srcSvcEqual := c.SrcService == other.SrcService && c.SrcNamespace == other.SrcNamespace
	dstSvcEqual := c.DstService == other.DstService && c.DstNamespace == other.DstNamespace
	return !srcSvcEqual || !dstSvcEqual
}

func (c *CircuitBreakerRule) ToSpec() (*apifault.CircuitBreakerRule, error) {
	if c == nil {
		return nil, nil
	}
	specData := &apifault.CircuitBreakerRule{}
	if len(c.Rule) > 0 {
		if err := json.Unmarshal([]byte(c.Rule), specData); err != nil {
			return nil, err
		}
	} else {
		// brief search, to display the services in list result
		specData.RuleMatcher = &apifault.RuleMatcher{
			Source: &apifault.RuleMatcher_SourceService{
				Service:   c.SrcService,
				Namespace: c.SrcNamespace,
			},
			Destination: &apifault.RuleMatcher_DestinationService{
				Service:   c.DstService,
				Namespace: c.DstNamespace,
				Method:    &apimodel.MatchString{Value: c.DstMethod},
			},
		}
	}
	specData.Id = c.ID
	specData.Name = c.Name
	specData.Description = c.Description
	specData.Level = apifault.Level(c.Level)
	specData.Enable = c.Enable
	specData.Revision = c.Revision
	specData.Ctime = commontime.Time2String(c.CreateTime)
	specData.Mtime = commontime.Time2String(c.ModifyTime)
	specData.Enable = c.Enable
	if c.EnableTime.Year() > 2000 {
		specData.Etime = commontime.Time2String(c.EnableTime)
	} else {
		specData.Etime = ""
	}
	return specData, nil
}

// FaultDetectRule 故障探测规则
type FaultDetectRule struct {
	ID           string
	Name         string
	Namespace    string
	Description  string
	DstService   string
	DstNamespace string
	DstMethod    string
	Rule         string
	Revision     string
	Metadata     map[string]string
	Valid        bool
	CreateTime   time.Time
	ModifyTime   time.Time
	Proto        *apifault.FaultDetectRule
}

func (r *FaultDetectRule) GetId() string {
	return r.ID
}

func (r *FaultDetectRule) GetMtime() time.Time {
	return r.ModifyTime
}

func (c *FaultDetectRule) IsServiceChange(other *FaultDetectRule) bool {
	dstSvcEqual := c.DstService == other.DstService && c.DstNamespace == other.DstNamespace
	return !dstSvcEqual
}

func (c *FaultDetectRule) ToSpec() (*apifault.FaultDetectRule, error) {
	if c == nil {
		return nil, nil
	}
	specData := &apifault.FaultDetectRule{}
	if len(c.Rule) > 0 {
		if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(c.Rule), specData); err != nil {
			if err = json.Unmarshal([]byte(c.Rule), specData); err != nil {
				return nil, err
			}
		}
	} else {
		// brief search, to display the services in list result
		specData.TargetService = &apifault.FaultDetectRule_DestinationService{
			Service:   c.DstService,
			Namespace: c.DstNamespace,
			Method:    &apimodel.MatchString{Value: c.DstMethod},
		}
	}
	normalizeFaultDetectRuleShape(specData, c.Rule)
	specData.Id = c.ID
	specData.Name = c.Name
	specData.Description = c.Description
	specData.Revision = c.Revision
	specData.Ctime = commontime.Time2String(c.CreateTime)
	specData.Mtime = commontime.Time2String(c.ModifyTime)
	specData.Editable = true
	specData.Deleteable = true
	return specData, nil
}

func normalizeFaultDetectRuleShape(specData *apifault.FaultDetectRule, rawRule string) {
	if specData == nil || rawRule == "" {
		return
	}
	var legacy map[string]json.RawMessage
	if err := json.Unmarshal([]byte(rawRule), &legacy); err != nil {
		return
	}
	if specData.GetTargetService() == nil {
		targetService := &apifault.FaultDetectRule_DestinationService{}
		if unmarshalFaultDetectProtoField(legacy["target_service"], targetService) {
			specData.TargetService = targetService
		}
	}
	if specData.GetTargetService() == nil {
		specData.TargetService = legacyFaultDetectTargetFromFirstSubRule(legacy["rules"])
	}
	if len(specData.GetRules()) > 0 {
		return
	}
	httpConfig := &apifault.HttpProtocolConfig{}
	if !unmarshalFaultDetectProtoField(legacy["http_config"], httpConfig) {
		httpConfig = nil
	}
	tcpConfig := &apifault.TcpProtocolConfig{}
	if !unmarshalFaultDetectProtoField(legacy["tcp_config"], tcpConfig) {
		tcpConfig = nil
	}
	udpConfig := &apifault.UdpProtocolConfig{}
	if !unmarshalFaultDetectProtoField(legacy["udp_config"], udpConfig) {
		udpConfig = nil
	}
	specData.Rules = []*apifault.FaultDetectSubRule{{
		Interval:   parseFaultDetectUint32(legacy["interval"]),
		Timeout:    parseFaultDetectUint32(legacy["timeout"]),
		Port:       parseFaultDetectUint32(legacy["port"]),
		Protocol:   parseFaultDetectProtocol(legacy["protocol"]),
		HttpConfig: httpConfig,
		TcpConfig:  tcpConfig,
		UdpConfig:  udpConfig,
	}}
}

func legacyFaultDetectTargetFromFirstSubRule(raw json.RawMessage) *apifault.FaultDetectRule_DestinationService {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var rules []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rules); err != nil || len(rules) == 0 {
		return nil
	}
	targetService := &apifault.FaultDetectRule_DestinationService{}
	if !unmarshalFaultDetectProtoField(rules[0]["target_service"], targetService) {
		return nil
	}
	return targetService
}

func unmarshalFaultDetectProtoField(raw json.RawMessage, msg proto.Message) bool {
	if len(raw) == 0 || string(raw) == "null" {
		return false
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(raw, msg); err == nil {
		return true
	}
	return json.Unmarshal(raw, msg) == nil
}

func parseFaultDetectUint32(raw json.RawMessage) uint32 {
	if len(raw) == 0 || string(raw) == "null" {
		return 0
	}
	var val uint32
	if err := json.Unmarshal(raw, &val); err == nil {
		return val
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		if parsed, err := strconv.ParseUint(strings.TrimSpace(text), 10, 32); err == nil {
			return uint32(parsed)
		}
	}
	return 0
}

func parseFaultDetectProtocol(raw json.RawMessage) apifault.FaultDetectRule_Protocol {
	if len(raw) == 0 || string(raw) == "null" {
		return apifault.FaultDetectRule_UNKNOWN
	}
	var name string
	if err := json.Unmarshal(raw, &name); err == nil {
		name = strings.TrimSpace(name)
		if val, ok := apifault.FaultDetectRule_Protocol_value[name]; ok {
			return apifault.FaultDetectRule_Protocol(val)
		}
		if val, err := strconv.ParseInt(name, 10, 32); err == nil {
			return apifault.FaultDetectRule_Protocol(val)
		}
		return apifault.FaultDetectRule_UNKNOWN
	}
	var val int32
	if err := json.Unmarshal(raw, &val); err == nil {
		return apifault.FaultDetectRule_Protocol(val)
	}
	return apifault.FaultDetectRule_UNKNOWN
}
