package rules

import (
	"sync"
	"time"

	apifault "github.com/polarismesh/specification/source/go/api/v1/fault_tolerance"

	"github.com/pole-io/pole-server/apis/pkg/types/service"
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
	circuitBreakerRules map[string]*CircuitBreakerRule
	Revision            string
}

func NewServiceWithCircuitBreakerRules(svcKey service.ServiceKey) *ServiceWithCircuitBreakerRules {
	return &ServiceWithCircuitBreakerRules{
		Service:             svcKey,
		circuitBreakerRules: make(map[string]*CircuitBreakerRule),
	}
}

func (s *ServiceWithCircuitBreakerRules) AddCircuitBreakerRule(rule *CircuitBreakerRule) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.circuitBreakerRules[rule.ID] = rule
}

func (s *ServiceWithCircuitBreakerRules) DelCircuitBreakerRule(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.circuitBreakerRules, id)
}

func (s *ServiceWithCircuitBreakerRules) IterateCircuitBreakerRules(callback func(*CircuitBreakerRule)) {
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
	s.circuitBreakerRules = make(map[string]*CircuitBreakerRule)
	s.Revision = ""
}

// ServiceWithFaultDetectRules 与服务关系绑定的探测规则
type ServiceWithFaultDetectRules struct {
	mutex            sync.RWMutex
	Service          service.ServiceKey
	faultDetectRules map[string]*FaultDetectRule
	Revision         string
}

func NewServiceWithFaultDetectRules(svcKey service.ServiceKey) *ServiceWithFaultDetectRules {
	return &ServiceWithFaultDetectRules{
		Service:          svcKey,
		faultDetectRules: make(map[string]*FaultDetectRule),
	}
}

func (s *ServiceWithFaultDetectRules) AddFaultDetectRule(rule *FaultDetectRule) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.faultDetectRules[rule.ID] = rule
}

func (s *ServiceWithFaultDetectRules) DelFaultDetectRule(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.faultDetectRules, id)
}

func (s *ServiceWithFaultDetectRules) IterateFaultDetectRules(callback func(*FaultDetectRule)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for _, rule := range s.faultDetectRules {
		callback(rule)
	}
}

func (s *ServiceWithFaultDetectRules) CountFaultDetectRules() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.faultDetectRules)
}

func (s *ServiceWithFaultDetectRules) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.faultDetectRules = make(map[string]*FaultDetectRule)
	s.Revision = ""
}

// CircuitBreakerRelation 熔断规则绑定关系
type CircuitBreakerRelation struct {
	ServiceID   string
	RuleID      string
	RuleVersion string
	Valid       bool
	CreateTime  time.Time
	ModifyTime  time.Time
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

func (c *CircuitBreakerRule) IsServiceChange(other *CircuitBreakerRule) bool {
	srcSvcEqual := c.SrcService == other.SrcService && c.SrcNamespace == other.SrcNamespace
	dstSvcEqual := c.DstService == other.DstService && c.DstNamespace == other.DstNamespace
	return !srcSvcEqual || !dstSvcEqual
}

// FaultDetectRule 故障探测规则
type FaultDetectRule struct {
	Proto        *apifault.FaultDetectRule
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
}

func (c *FaultDetectRule) IsServiceChange(other *FaultDetectRule) bool {
	dstSvcEqual := c.DstService == other.DstService && c.DstNamespace == other.DstNamespace
	return !dstSvcEqual
}
