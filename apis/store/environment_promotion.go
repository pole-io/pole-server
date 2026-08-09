package store

import (
	"errors"

	"github.com/pole-io/pole-server/apis/pkg/types"
)

var ErrEnvironmentPromotionRevisionConflict = errors.New("environment promotion topology draft revision conflict")

// EnvironmentPromotionStore persists the global topology draft and immutable revisions.
// It is intentionally separate from NamespaceStore: a Namespace is an environment
// resource, while topology is a relationship aggregate over multiple Namespaces.
type EnvironmentPromotionStore interface {
	GetEnvironmentPromotionTopology() (*types.EnvironmentPromotionTopology, error)
	SaveEnvironmentPromotionTopology(topology *types.EnvironmentPromotionTopology, expectedRevision uint64) error
	PublishEnvironmentPromotionTopology(expectedRevision uint64, createBy, comment string) (
		*types.EnvironmentPromotionTopologyRevision, error)
	ListEnvironmentPromotionTopologyRevisions() ([]*types.EnvironmentPromotionTopologyRevision, error)
}
