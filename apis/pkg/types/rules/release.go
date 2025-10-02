package rules

import (
	"time"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

type RuleRelease struct {
	Id          string
	ReleaseName string
	RuleId      string
	RuleName    string
	Description string
	ReleaseType ReleaseType
	Active      bool
	Version     uint64
	Valid       bool
	Ctime       time.Time
	Mtime       time.Time
}

func (l *RuleRelease) Key() string {
	return l.RuleName + "_" + string(l.ReleaseType)
}

func (r *RuleRelease) ToSpec() *apimodel.RuleRelease {
	return &apimodel.RuleRelease{
		Id:          r.Id,
		ReleaseName: r.ReleaseName,
		RuleId:      r.RuleId,
		RuleName:    r.RuleName,
		Description: r.Description,
		ReleaseType: string(r.ReleaseType),
		Version:     r.Version,
	}
}

func (r *RuleRelease) FromSpec(spec *apimodel.RuleRelease) {
	r.Id = spec.Id
	r.ReleaseName = spec.ReleaseName
	r.RuleId = spec.RuleId
	r.RuleName = spec.RuleName
	r.Description = spec.Description
	r.ReleaseType = ReleaseType(spec.ReleaseType)
	r.Version = spec.Version
}

func (r *RuleRelease) Clone() *RuleRelease {
	return &RuleRelease{
		Id:          r.Id,
		ReleaseName: r.ReleaseName,
		RuleId:      r.RuleId,
		RuleName:    r.RuleName,
		Description: r.Description,
		ReleaseType: r.ReleaseType,
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
