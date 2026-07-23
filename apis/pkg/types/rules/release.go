package rules

import (
	"time"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	commontime "github.com/pole-io/pole-server/pkg/common/utils/time"
)

type RuleRelease struct {
	Id           string
	Namespace    string
	ReleaseName  string
	RuleId       string
	RuleName     string
	Description  string
	Resource     apimodel.RuleRelease_RuleType
	ReleaseType  ReleaseType
	ClientLabels []*apimodel.ClientLabel
	Active       bool
	Version      uint64
	Valid        bool
	Ctime        time.Time
	Mtime        time.Time
}

func (l *RuleRelease) GetGrayResource() string {
	return l.Resource.String() + "/" + l.RuleId
}

func (l *RuleRelease) GetClientLabels() []*apimodel.ClientLabel {
	return l.ClientLabels
}

func (l *RuleRelease) Key() string {
	return l.RuleName + "_" + string(l.ReleaseType)
}

func (r *RuleRelease) ToSpec() *apimodel.RuleRelease {
	out := &apimodel.RuleRelease{
		Id:          r.Id,
		ReleaseName: r.ReleaseName,
		RuleId:      r.RuleId,
		RuleName:    r.RuleName,
		Description: r.Description,
		Resource:    r.Resource,
		ReleaseType: string(r.ReleaseType),
		Version:     r.Version,
		Active:      r.Active,
		Ctime:       commontime.Time2String(r.Ctime),
		Mtime:       commontime.Time2String(r.Mtime),
	}
	SetOwnerNamespaceOnProto(out, r.Namespace)
	return out
}

func (r *RuleRelease) FromSpec(spec *apimodel.RuleRelease) {
	r.Id = spec.Id
	r.Namespace = OwnerNamespaceFromProto(spec)
	r.ReleaseName = spec.ReleaseName
	r.RuleId = spec.RuleId
	r.RuleName = spec.RuleName
	r.Description = spec.Description
	r.ReleaseType = ReleaseType(spec.ReleaseType)
	r.Version = spec.Version
	r.Active = spec.Active
	r.Resource = spec.Resource
}

func (r *RuleRelease) Clone() *RuleRelease {
	return &RuleRelease{
		Id:          r.Id,
		Namespace:   r.Namespace,
		ReleaseName: r.ReleaseName,
		RuleId:      r.RuleId,
		RuleName:    r.RuleName,
		Description: r.Description,
		ReleaseType: r.ReleaseType,
		Resource:    r.Resource,
		Active:      r.Active,
		Ctime:       r.Ctime,
		Mtime:       r.Mtime,
		Version:     r.Version,
		Valid:       r.Valid,
	}
}

type LaneGroupRelease struct {
	RuleRelease
	Rule *LaneGroupProto
}

type RouterRuleRelease struct {
	RuleRelease
	Rule *ExtendRouterConfig
}

type FaultDetectRelease struct {
	RuleRelease
	Rule *FaultDetectRule
}

type RateLimitRelease struct {
	RuleRelease
	Rule *RateLimit
}

type CircuitBreakerRelease struct {
	RuleRelease
	Rule *CircuitBreakerRule
}

type LosslessRuleRelease struct {
	RuleRelease
	Rule *LosslessRule
}

func (l *LosslessRuleRelease) ActiveKey() string {
	return string(l.ReleaseType) + "/" + l.Rule.Namespace + "/" + l.Rule.Service
}
