package rules

import (
	"time"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
	"google.golang.org/protobuf/proto"

	commontime "github.com/pole-io/pole-server/pkg/common/utils/time"
)

type TrafficGovernanceRuleKind string

const (
	TrafficGovernanceRuleKindSecurity TrafficGovernanceRuleKind = "traffic-security"
	TrafficGovernanceRuleKindMirror   TrafficGovernanceRuleKind = "traffic-mirror"
	TrafficGovernanceRuleKindMock     TrafficGovernanceRuleKind = "traffic-mock"
)

// TrafficGovernanceRule wraps the three traffic governance protobuf rules that
// share the same storage, cache, release, and query lifecycle.
type TrafficGovernanceRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Service     string `json:"service"`
	Enable      bool
	Priority    uint32
	Valid       bool
	CTime       time.Time
	MTime       time.Time
	Metadata    map[string]string
	Revision    string
	Description string
	Kind        TrafficGovernanceRuleKind
	Proto       proto.Message
}

func NewTrafficSecurityRule(spec *apisecurity.TrafficSecurityRule) *TrafficGovernanceRule {
	r := &TrafficGovernanceRule{Kind: TrafficGovernanceRuleKindSecurity}
	r.FromTrafficSecuritySpec(spec)
	return r
}

func NewTrafficMirrorRule(spec *apitraffic.TrafficMirror) *TrafficGovernanceRule {
	r := &TrafficGovernanceRule{Kind: TrafficGovernanceRuleKindMirror}
	r.FromTrafficMirrorSpec(spec)
	return r
}

func NewTrafficMockRule(spec *apitraffic.TrafficMock) *TrafficGovernanceRule {
	r := &TrafficGovernanceRule{Kind: TrafficGovernanceRuleKindMock}
	r.FromTrafficMockSpec(spec)
	return r
}

func (r *TrafficGovernanceRule) FromTrafficSecuritySpec(spec *apisecurity.TrafficSecurityRule) {
	if r == nil || spec == nil {
		return
	}
	r.ID = spec.Id
	r.Name = spec.Name
	r.Enable = spec.Enable
	r.Priority = spec.Priority
	r.Metadata = spec.Metadata
	r.Revision = spec.Revision
	r.Description = spec.Description
	r.Valid = true
	r.Kind = TrafficGovernanceRuleKindSecurity
	r.Proto = spec
}

func (r *TrafficGovernanceRule) FromTrafficMirrorSpec(spec *apitraffic.TrafficMirror) {
	if r == nil || spec == nil {
		return
	}
	r.ID = spec.Id
	r.Name = spec.Name
	r.Namespace, r.Service = mirrorRuleServiceScope(spec)
	r.Enable = spec.Enable
	r.Priority = spec.Priority
	r.Metadata = spec.Metadata
	r.Revision = spec.Revision
	r.Description = spec.Description
	r.Valid = true
	r.Kind = TrafficGovernanceRuleKindMirror
	r.Proto = spec
}

func (r *TrafficGovernanceRule) FromTrafficMockSpec(spec *apitraffic.TrafficMock) {
	if r == nil || spec == nil {
		return
	}
	r.ID = spec.Id
	r.Name = spec.Name
	r.Namespace, r.Service = mockRuleServiceScope(spec)
	r.Enable = spec.Enable
	r.Priority = spec.Priority
	r.Metadata = spec.Metadata
	r.Revision = spec.Revision
	r.Description = spec.Description
	r.Valid = true
	r.Kind = TrafficGovernanceRuleKindMock
	r.Proto = spec
}

func (r *TrafficGovernanceRule) ToTrafficSecuritySpec() *apisecurity.TrafficSecurityRule {
	if r == nil {
		return nil
	}
	spec, _ := r.Proto.(*apisecurity.TrafficSecurityRule)
	if spec == nil {
		spec = &apisecurity.TrafficSecurityRule{}
	} else {
		spec = proto.Clone(spec).(*apisecurity.TrafficSecurityRule)
	}
	r.applyTrafficSecurityFields(spec)
	return spec
}

func (r *TrafficGovernanceRule) ToTrafficMirrorSpec() *apitraffic.TrafficMirror {
	if r == nil {
		return nil
	}
	spec, _ := r.Proto.(*apitraffic.TrafficMirror)
	if spec == nil {
		spec = &apitraffic.TrafficMirror{}
	} else {
		spec = proto.Clone(spec).(*apitraffic.TrafficMirror)
	}
	r.applyTrafficMirrorFields(spec)
	return spec
}

func (r *TrafficGovernanceRule) ToTrafficMockSpec() *apitraffic.TrafficMock {
	if r == nil {
		return nil
	}
	spec, _ := r.Proto.(*apitraffic.TrafficMock)
	if spec == nil {
		spec = &apitraffic.TrafficMock{}
	} else {
		spec = proto.Clone(spec).(*apitraffic.TrafficMock)
	}
	r.applyTrafficMockFields(spec)
	return spec
}

func (r *TrafficGovernanceRule) applyTrafficSecurityFields(spec *apisecurity.TrafficSecurityRule) {
	spec.Id = r.ID
	spec.Name = r.Name
	spec.Enable = r.Enable
	spec.Priority = r.Priority
	spec.Metadata = r.Metadata
	spec.Revision = r.Revision
	spec.Description = r.Description
	spec.Ctime = commontime.Time2String(r.CTime)
	spec.Mtime = commontime.Time2String(r.MTime)
}

func (r *TrafficGovernanceRule) applyTrafficMirrorFields(spec *apitraffic.TrafficMirror) {
	spec.Id = r.ID
	spec.Name = r.Name
	spec.Enable = r.Enable
	spec.Priority = r.Priority
	spec.Metadata = r.Metadata
	spec.Revision = r.Revision
	spec.Description = r.Description
	spec.Ctime = commontime.Time2String(r.CTime)
	spec.Mtime = commontime.Time2String(r.MTime)
}

func (r *TrafficGovernanceRule) applyTrafficMockFields(spec *apitraffic.TrafficMock) {
	spec.Id = r.ID
	spec.Name = r.Name
	spec.Enable = r.Enable
	spec.Priority = r.Priority
	spec.Metadata = r.Metadata
	spec.Revision = r.Revision
	spec.Description = r.Description
	spec.Ctime = commontime.Time2String(r.CTime)
	spec.Mtime = commontime.Time2String(r.MTime)
}

func (r *TrafficGovernanceRule) GetId() string {
	return r.ID
}

func (r *TrafficGovernanceRule) GetMtime() time.Time {
	return r.MTime
}

func mirrorRuleServiceScope(spec *apitraffic.TrafficMirror) (string, string) {
	if spec == nil {
		return "", ""
	}
	for _, rule := range spec.GetRules() {
		if source := rule.GetSource(); source != nil {
			return source.GetNamespace(), source.GetService()
		}
	}
	return "", ""
}

func mockRuleServiceScope(spec *apitraffic.TrafficMock) (string, string) {
	if spec == nil {
		return "", ""
	}
	for _, rule := range spec.GetRules() {
		if source := rule.GetSource(); source != nil {
			return source.GetNamespace(), source.GetService()
		}
	}
	return "", ""
}

type TrafficGovernanceRuleRelease struct {
	RuleRelease
	Rule *TrafficGovernanceRule
}

func (r *TrafficGovernanceRuleRelease) ActiveKey() string {
	if r == nil || r.Rule == nil {
		return ""
	}
	return string(r.ReleaseType) + "/" + r.Rule.Namespace + "/" + r.Rule.Service
}
