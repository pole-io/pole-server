package rules

import (
	"time"

	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
	"google.golang.org/protobuf/proto"

	commontime "github.com/pole-io/pole-server/pkg/common/utils/time"
)

const (
	losslessNamespaceMetadata   = "pole.io/lossless/namespace"
	losslessServiceMetadata     = "pole.io/lossless/service"
	losslessDescriptionMetadata = "pole.io/lossless/description"
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
	out := r.Proto
	if out == nil {
		out = &apitraffic.LosslessRule{}
	} else {
		out = proto.Clone(out).(*apitraffic.LosslessRule)
	}
	out.Id = r.ID
	out.Namespace = r.Namespace
	out.Metadata = cloneLosslessMetadata(r.Metadata)
	out.Metadata[losslessNamespaceMetadata] = r.Namespace
	out.Metadata[losslessServiceMetadata] = r.Service
	out.Metadata[losslessDescriptionMetadata] = r.Description
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
	r.Metadata = cloneLosslessMetadata(spec.Metadata)
	r.Namespace = spec.GetNamespace()
	if r.Namespace == "" {
		r.Namespace = r.Metadata[losslessNamespaceMetadata]
	}
	r.Service = r.Metadata[losslessServiceMetadata]
	r.Description = r.Metadata[losslessDescriptionMetadata]
	delete(r.Metadata, losslessNamespaceMetadata)
	delete(r.Metadata, losslessServiceMetadata)
	delete(r.Metadata, losslessDescriptionMetadata)
	r.Revision = spec.Revision
	r.Valid = true
	r.Proto = proto.Clone(spec).(*apitraffic.LosslessRule)
	r.Proto.Metadata = cloneLosslessMetadata(r.Metadata)
}

func cloneLosslessMetadata(metadata map[string]string) map[string]string {
	out := make(map[string]string, len(metadata)+3)
	for key, value := range metadata {
		out[key] = value
	}
	return out
}

func (r *LosslessRule) GetId() string {
	return r.ID
}

func (r *LosslessRule) GetMtime() time.Time {
	return r.MTime
}
