package store

import (
	"errors"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
)

var ErrNamespaceConfigTemplateDraftConflict = errors.New("namespace config template draft version conflict")

// NamespaceConfigTemplateDraftStore persists environment-scoped editable definitions.
type NamespaceConfigTemplateDraftStore interface {
	GetNamespaceConfigTemplateDraft(namespace string, templateID uint64) (
		*conftypes.NamespaceConfigTemplateDraft, error)
	SaveNamespaceConfigTemplateDraft(draft *conftypes.NamespaceConfigTemplateDraft, expectedVersion uint64) error
}
