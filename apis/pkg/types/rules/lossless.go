package rules

import (
	"time"

	commontime "github.com/pole-io/pole-server/pkg/common/utils/time"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

// LosslessRule 无损规则
type LosslessRule struct {
	ID          string `json:"id"`
	Namespace   string `json:"namespace"`
	Service     string `json:"service"`
	Valid       bool
	CTime       time.Time
	MTime       time.Time
	Metadata    map[string]string
	Revision    string
	Description string
	Proto       *apitraffic.LosslessRule
}

func (r *LosslessRule) ToSpec() *apitraffic.LosslessRule {
	if r == nil {
		return nil
	}
	out := &apitraffic.LosslessRule{}
	proto := r.Proto
	out = proto
	out.Id = r.ID
	out.Namespace = r.Namespace
	out.Service = r.Service
	out.Metadata = r.Metadata
	out.Revision = r.Revision
	out.Ctime = commontime.Time2String(r.CTime)
	out.Mtime = commontime.Time2String(r.MTime)
	return out
}

func (r *LosslessRule) FromSpec(spec *apitraffic.LosslessRule) {
	if r == nil || spec == nil {
		return
	}
	r.ID = spec.Id
	r.Namespace = spec.Namespace
	r.Service = spec.Service
	r.Metadata = spec.Metadata
	r.Revision = spec.Revision
	r.Valid = true
	r.Proto = spec
}

func (r *LosslessRule) GetId() string {
	return r.ID
}

func (r *LosslessRule) GetMtime() time.Time {
	return r.MTime
}
