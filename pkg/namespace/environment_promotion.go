package namespace

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

const (
	conflictPolicyBlock = "BLOCK"
)

var ErrEnvironmentPromotionStoreUnavailable = errors.New("environment promotion store is unavailable")

func (s *Server) GetEnvironmentPromotionTopology(
	_ context.Context) (*types.EnvironmentPromotionTopology, error) {
	storage, err := s.environmentPromotionStorage()
	if err != nil {
		return nil, err
	}
	return storage.GetEnvironmentPromotionTopology()
}

func (s *Server) SaveEnvironmentPromotionTopology(ctx context.Context,
	topology *types.EnvironmentPromotionTopology) ([]types.TopologyValidationIssue, error) {
	issues := s.ValidateEnvironmentPromotionTopology(topology)
	if len(issues) != 0 {
		return issues, nil
	}
	storage, err := s.environmentPromotionStorage()
	if err != nil {
		return nil, err
	}
	expectedRevision := topology.DraftRevision
	topology.ModifyBy = string(utils.ParseOwnerID(ctx))
	if err := storage.SaveEnvironmentPromotionTopology(topology, expectedRevision); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Server) PublishEnvironmentPromotionTopology(ctx context.Context,
	expectedRevision uint64, comment string) (*types.EnvironmentPromotionTopologyRevision,
	[]types.TopologyValidationIssue, error) {
	storage, err := s.environmentPromotionStorage()
	if err != nil {
		return nil, nil, err
	}
	topology, err := storage.GetEnvironmentPromotionTopology()
	if err != nil {
		return nil, nil, err
	}
	issues := s.ValidateEnvironmentPromotionTopology(topology)
	if len(issues) != 0 {
		return nil, issues, nil
	}
	revision, err := storage.PublishEnvironmentPromotionTopology(
		expectedRevision, string(utils.ParseOwnerID(ctx)), strings.TrimSpace(comment))
	return revision, nil, err
}

func (s *Server) ListEnvironmentPromotionTopologyRevisions(
	_ context.Context) ([]*types.EnvironmentPromotionTopologyRevision, error) {
	storage, err := s.environmentPromotionStorage()
	if err != nil {
		return nil, err
	}
	return storage.ListEnvironmentPromotionTopologyRevisions()
}

func (s *Server) ValidateEnvironmentPromotionTopology(
	topology *types.EnvironmentPromotionTopology) []types.TopologyValidationIssue {
	issues := ValidateEnvironmentPromotionTopology(topology)
	if topology == nil {
		return issues
	}
	names := make(map[string]struct{})
	for _, edge := range topology.Edges {
		names[strings.TrimSpace(edge.Source)] = struct{}{}
		names[strings.TrimSpace(edge.Target)] = struct{}{}
	}
	for _, binding := range topology.LaneBaseBindings {
		names[strings.TrimSpace(binding.Lane)] = struct{}{}
		names[strings.TrimSpace(binding.Base)] = struct{}{}
	}
	for name := range names {
		if name == "" {
			continue
		}
		namespace, err := s.storage.GetNamespace(name)
		switch {
		case err != nil:
			issues = append(issues, types.TopologyValidationIssue{
				Code: "NAMESPACE_LOOKUP_FAILED", Path: name, Message: "环境空间查询失败",
			})
		case namespace == nil:
			issues = append(issues, types.TopologyValidationIssue{
				Code: "NAMESPACE_NOT_FOUND", Path: name, Message: "环境空间不存在",
			})
		case isSystemNamespace(namespace):
			issues = append(issues, types.TopologyValidationIssue{
				Code: "SYSTEM_NAMESPACE_UNSUPPORTED", Path: name, Message: "系统空间不能加入晋升拓扑",
			})
		}
	}
	return issues
}

func (s *Server) environmentPromotionStorage() (store.EnvironmentPromotionStore, error) {
	storage, ok := s.storage.(store.EnvironmentPromotionStore)
	if !ok || storage == nil {
		return nil, ErrEnvironmentPromotionStoreUnavailable
	}
	return storage, nil
}

// ValidateEnvironmentPromotionTopology validates graph invariants independent
// from persistence. Namespace existence is checked by the service boundary.
func ValidateEnvironmentPromotionTopology(topology *types.EnvironmentPromotionTopology) []types.TopologyValidationIssue {
	if topology == nil {
		return []types.TopologyValidationIssue{{Code: "TOPOLOGY_REQUIRED", Message: "晋升拓扑不能为空"}}
	}

	issues := make([]types.TopologyValidationIssue, 0)
	edgeIDs := make(map[string]struct{}, len(topology.Edges))
	baselineNodes := make(map[string]struct{})
	adjacency := make(map[string][]string)
	for index := range topology.Edges {
		edge := &topology.Edges[index]
		path := fmt.Sprintf("edges[%d]", index)
		edge.ID = strings.TrimSpace(edge.ID)
		edge.Source = strings.TrimSpace(edge.Source)
		edge.Target = strings.TrimSpace(edge.Target)
		if edge.ID == "" {
			issues = append(issues, types.TopologyValidationIssue{Code: "EDGE_ID_REQUIRED", Path: path + ".id", Message: "晋升边 ID 不能为空"})
		} else if _, exists := edgeIDs[edge.ID]; exists {
			issues = append(issues, types.TopologyValidationIssue{Code: "EDGE_ID_DUPLICATED", Path: path + ".id", Message: "晋升边 ID 不能重复"})
		} else {
			edgeIDs[edge.ID] = struct{}{}
		}
		if edge.Source == "" || edge.Target == "" {
			issues = append(issues, types.TopologyValidationIssue{Code: "EDGE_ENDPOINT_REQUIRED", Path: path, Message: "晋升边必须指定来源和目标环境"})
			continue
		}
		if edge.Source == edge.Target {
			issues = append(issues, types.TopologyValidationIssue{Code: "EDGE_SELF_LOOP", Path: path, Message: "环境不能晋升到自身"})
			continue
		}
		if edge.ConflictPolicy == "" {
			edge.ConflictPolicy = conflictPolicyBlock
		}
		if edge.ConflictPolicy != conflictPolicyBlock {
			issues = append(issues, types.TopologyValidationIssue{Code: "CONFLICT_POLICY_UNSUPPORTED", Path: path + ".conflict_policy", Message: "当前仅支持冲突阻断策略 BLOCK"})
		}
		baselineNodes[edge.Source] = struct{}{}
		baselineNodes[edge.Target] = struct{}{}
		adjacency[edge.Source] = append(adjacency[edge.Source], edge.Target)
	}

	bindings := make(map[string]struct{}, len(topology.LaneBaseBindings))
	for index := range topology.LaneBaseBindings {
		binding := &topology.LaneBaseBindings[index]
		path := fmt.Sprintf("lane_base_bindings[%d]", index)
		binding.Lane = strings.TrimSpace(binding.Lane)
		binding.Base = strings.TrimSpace(binding.Base)
		if binding.Lane == "" || binding.Base == "" {
			issues = append(issues, types.TopologyValidationIssue{Code: "LANE_BINDING_REQUIRED", Path: path, Message: "泳道绑定必须指定泳道和基线环境"})
			continue
		}
		if binding.Lane == binding.Base {
			issues = append(issues, types.TopologyValidationIssue{Code: "LANE_BASE_SAME", Path: path, Message: "泳道不能绑定自身为基线"})
		}
		if _, exists := bindings[binding.Lane]; exists {
			issues = append(issues, types.TopologyValidationIssue{Code: "LANE_BINDING_DUPLICATED", Path: path + ".lane", Message: "一个泳道只能绑定一个基线环境"})
		} else {
			bindings[binding.Lane] = struct{}{}
		}
		if _, exists := baselineNodes[binding.Lane]; exists {
			issues = append(issues, types.TopologyValidationIssue{Code: "LANE_IN_BASELINE_DAG", Path: path + ".lane", Message: "泳道不能作为基线晋升 DAG 节点"})
		}
	}

	if hasEnvironmentPromotionCycle(adjacency) {
		issues = append(issues, types.TopologyValidationIssue{Code: "BASELINE_DAG_CYCLE", Path: "edges", Message: "基线环境晋升关系不能形成环"})
	}
	return issues
}

func hasEnvironmentPromotionCycle(adjacency map[string][]string) bool {
	const (
		unvisited = iota
		visiting
		visited
	)
	states := make(map[string]int, len(adjacency))
	var visit func(string) bool
	visit = func(node string) bool {
		if states[node] == visiting {
			return true
		}
		if states[node] == visited {
			return false
		}
		states[node] = visiting
		for _, target := range adjacency[node] {
			if visit(target) {
				return true
			}
		}
		states[node] = visited
		return false
	}
	for node := range adjacency {
		if visit(node) {
			return true
		}
	}
	return false
}
