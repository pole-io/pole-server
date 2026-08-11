package store

import (
	"context"
	"time"

	skilltypes "github.com/pole-io/pole-server/apis/pkg/types/skillmarketplace"
)

// BundleStore owns immutable content-addressed bundle bytes. Metadata stores
// reference digests only, so other adapters can replace MySQL BLOB storage.
type BundleStore interface {
	PutBundle(ctx context.Context, digest string, bundle []byte) error
	GetBundle(ctx context.Context, digest string) ([]byte, error)
	DeleteUnreferencedBundles(ctx context.Context, olderThan time.Time) (uint64, error)
}

// SkillMarketplaceStore is deliberately separate from Store until the public
// specification has a Skill resource enum. Concrete Store plugins may expose
// it through a type assertion without widening every generated Store mock.
type SkillMarketplaceStore interface {
	BundleStore
	UpsertSkillPublisher(context.Context, *skilltypes.Publisher) error
	GetSkillPublisher(context.Context, string) (*skilltypes.Publisher, error)
	AddSkillPublisherKey(context.Context, *skilltypes.PublisherKey) error
	GetSkillPublisherKey(context.Context, string, uint32) (*skilltypes.PublisherKey, error)
	RevokeSkillPublisherKey(context.Context, string, uint32, time.Time) error
	ListSkillPublisherMembers(context.Context, string) ([]*skilltypes.PublisherMember, error)
	ReplaceSkillPublisherMembers(context.Context, string, []*skilltypes.PublisherMember) error
	CreateSkillRelease(context.Context, *skilltypes.Skill, *skilltypes.Release) error
	GetSkill(context.Context, string, string) (*skilltypes.Skill, error)
	SearchSkills(context.Context, skilltypes.SearchQuery) ([]*skilltypes.Skill, uint32, error)
	ListSkillReleases(context.Context, string) ([]*skilltypes.Release, error)
	GetSkillRelease(context.Context, string, string, string) (*skilltypes.Release, error)
	SetSkillReleaseLifecycle(context.Context, string, string, bool, string) error
	CreateSkillReview(context.Context, *skilltypes.Review, skilltypes.ReleaseStatus, time.Time) error
	ListSkillReviews(context.Context, skilltypes.ReleaseStatus, uint32) ([]*skilltypes.ReviewQueueItem, error)
	ListSkillGrants(context.Context, string) ([]*skilltypes.Grant, error)
	ReplaceSkillGrants(context.Context, string, []*skilltypes.Grant) error
	CreateRegistrySource(context.Context, *skilltypes.RegistrySource) error
	UpdateRegistrySource(context.Context, *skilltypes.RegistrySource) error
	DeleteRegistrySource(context.Context, string) error
	GetRegistrySource(context.Context, string) (*skilltypes.RegistrySource, error)
	ListRegistrySources(context.Context) ([]*skilltypes.RegistrySource, error)
	ListDueRegistrySources(context.Context, time.Time, uint32) ([]*skilltypes.RegistrySource, error)
	MarkRegistrySyncStarted(context.Context, string, time.Time) error
	MarkRegistrySyncFinished(context.Context, string, string, string, time.Time, time.Time) error
	RecoverInterruptedRegistrySyncs(context.Context, time.Time) error
}
