package rules

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

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
				Method:    &apimodel.MatchString{Value: &wrappers.StringValue{Value: c.DstMethod}},
			},
		}
	}
	specData.Id = c.ID
	specData.Name = c.Name
	specData.Namespace = c.Namespace
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
		if err := json.Unmarshal([]byte(c.Rule), specData); err != nil {
			return nil, err
		}
	} else {
		// brief search, to display the services in list result
		specData.TargetService = &apifault.FaultDetectRule_DestinationService{
			Service:   c.DstService,
			Namespace: c.DstNamespace,
			Method:    &apimodel.MatchString{Value: &wrappers.StringValue{Value: c.DstMethod}},
		}
	}
	specData.Id = c.ID
	specData.Name = c.Name
	specData.Namespace = c.Namespace
	specData.Description = c.Description
	specData.Revision = c.Revision
	specData.Ctime = commontime.Time2String(c.CreateTime)
	specData.Mtime = commontime.Time2String(c.ModifyTime)
	specData.Editable = true
	specData.Deleteable = true
	return specData, nil
}
