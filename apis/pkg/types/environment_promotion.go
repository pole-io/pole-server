package types

import "time"

// EnvironmentPromotionTopology is the editable global environment promotion graph.
// Only baseline environments participate in Edges. Lanes return changes through
// LaneBaseBindings so that the baseline graph remains acyclic.
type EnvironmentPromotionTopology struct {
	DraftRevision     uint64                     `json:"draft_revision"`
	PublishedRevision uint64                     `json:"published_revision"`
	Edges             []EnvironmentPromotionEdge `json:"edges"`
	LaneBaseBindings  []LaneBaseBinding          `json:"lane_base_bindings"`
	ModifyBy          string                     `json:"modify_by,omitempty"`
	ModifyTime        time.Time                  `json:"modify_time,omitempty"`
}

// EnvironmentPromotionEdge defines one allowed baseline promotion direction.
type EnvironmentPromotionEdge struct {
	ID              string   `json:"id"`
	Source          string   `json:"source"`
	Target          string   `json:"target"`
	RequireFormal   bool     `json:"require_formal_release"`
	RequireApproval bool     `json:"require_approval"`
	ValidationGates []string `json:"validation_gates,omitempty"`
	AllowedDomains  []string `json:"allowed_resource_domains,omitempty"`
	ConflictPolicy  string   `json:"conflict_policy,omitempty"`
}

// LaneBaseBinding binds a lane to exactly one baseline environment.
type LaneBaseBinding struct {
	Lane string `json:"lane"`
	Base string `json:"base"`
}

// EnvironmentPromotionTopologyRevision is an immutable published topology snapshot.
type EnvironmentPromotionTopologyRevision struct {
	Revision   uint64                       `json:"revision"`
	Topology   EnvironmentPromotionTopology `json:"topology"`
	Comment    string                       `json:"comment,omitempty"`
	CreateBy   string                       `json:"create_by,omitempty"`
	CreateTime time.Time                    `json:"create_time,omitempty"`
}

// TopologyValidationIssue is stable enough for both API and Console field feedback.
type TopologyValidationIssue struct {
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}
