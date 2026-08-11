package skillmarketplace

import (
	"context"
	"fmt"
	"testing"

	skilltypes "github.com/pole-io/pole-server/apis/pkg/types/skillmarketplace"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/stretchr/testify/require"
)

type memorySkillStore struct {
	store.SkillMarketplaceStore
	publishers map[string]*skilltypes.Publisher
	keys       map[string]*skilltypes.PublisherKey
	skills     map[string]*skilltypes.Skill
	releases   map[string]*skilltypes.Release
	bundles    map[string][]byte
	grants     map[string][]*skilltypes.Grant
}

func newMemorySkillStore() *memorySkillStore {
	return &memorySkillStore{publishers: map[string]*skilltypes.Publisher{}, keys: map[string]*skilltypes.PublisherKey{}, skills: map[string]*skilltypes.Skill{}, releases: map[string]*skilltypes.Release{}, bundles: map[string][]byte{}, grants: map[string][]*skilltypes.Grant{}}
}
func identity(p, n string) string { return p + "/" + n }
func (m *memorySkillStore) GetSkillPublisher(_ context.Context, name string) (*skilltypes.Publisher, error) {
	value := m.publishers[name]
	if value == nil {
		return nil, ErrNotFound
	}
	copyValue := *value
	return &copyValue, nil
}
func (m *memorySkillStore) GetSkillPublisherKey(_ context.Context, id string, version uint32) (*skilltypes.PublisherKey, error) {
	value := m.keys[fmt.Sprintf("%s/%d", id, version)]
	if value == nil {
		return nil, ErrNotFound
	}
	return value, nil
}
func (m *memorySkillStore) GetSkillRelease(_ context.Context, publisher, name, version string) (*skilltypes.Release, error) {
	value := m.releases[identity(publisher, name)+"@"+version]
	if value == nil {
		return nil, ErrNotFound
	}
	copyValue := *value
	return &copyValue, nil
}
func (m *memorySkillStore) PutBundle(_ context.Context, digest string, bundle []byte) error {
	m.bundles[digest] = append([]byte(nil), bundle...)
	return nil
}
func (m *memorySkillStore) GetBundle(_ context.Context, digest string) ([]byte, error) {
	value := m.bundles[digest]
	if value == nil {
		return nil, ErrNotFound
	}
	return value, nil
}
func (m *memorySkillStore) CreateSkillRelease(_ context.Context, skill *skilltypes.Skill, release *skilltypes.Release) error {
	key := identity(skill.Publisher, skill.Name)
	if current := m.skills[key]; current != nil {
		release.SkillID = current.ID
	} else {
		copied := *skill
		m.skills[key] = &copied
	}
	copiedRelease := *release
	m.releases[key+"@"+release.Version] = &copiedRelease
	return nil
}
func (m *memorySkillStore) GetSkill(_ context.Context, publisher, name string) (*skilltypes.Skill, error) {
	value := m.skills[identity(publisher, name)]
	if value == nil {
		return nil, ErrNotFound
	}
	copied := *value
	return &copied, nil
}
func (m *memorySkillStore) ListSkillGrants(_ context.Context, skillID string) ([]*skilltypes.Grant, error) {
	return m.grants[skillID], nil
}
func (m *memorySkillStore) SearchSkills(_ context.Context, query skilltypes.SearchQuery) ([]*skilltypes.Skill, uint32, error) {
	items := make([]*skilltypes.Skill, 0)
	for _, skill := range m.skills {
		allowed := skill.Visibility == skilltypes.VisibilityPublic || skill.OwnerID == query.ViewerID
		for _, principal := range query.Principals {
			for _, grant := range m.grants[skill.ID] {
				allowed = allowed || principal.PrincipalType == grant.PrincipalType && principal.PrincipalID == grant.PrincipalID
			}
		}
		if allowed {
			copied := *skill
			items = append(items, &copied)
		}
	}
	return items, uint32(len(items)), nil
}

func TestPublishNewVersionReusesExistingSkillID(t *testing.T) {
	storage := newMemorySkillStore()
	storage.publishers["pole"] = &skilltypes.Publisher{ID: "publisher-1", Name: "pole", OwnerID: "user-1"}
	service := NewService(storage)
	bundle := testZIP(t, zipTestEntry{name: "SKILL.md", content: "---\nname: demo-skill\ndescription: demo\n---\n", mode: 0o644})
	first, err := service.Publish(context.Background(), PublishRequest{Publisher: "pole", Name: "demo-skill", Version: "1.0.0", Visibility: skilltypes.VisibilityPrivate, Bundle: bundle, Actor: Principal{Type: "user", ID: "user-1"}})
	require.NoError(t, err)
	second, err := service.Publish(context.Background(), PublishRequest{Publisher: "pole", Name: "demo-skill", Version: "1.1.0", Visibility: skilltypes.VisibilityPrivate, Bundle: bundle, Actor: Principal{Type: "user", ID: "user-1"}})
	require.NoError(t, err)
	require.Equal(t, first.SkillID, second.SkillID)
	require.Len(t, storage.skills, 1)
	require.Len(t, storage.releases, 2)
}

func TestSearchUsesExpandedGroupAndRolePrincipals(t *testing.T) {
	storage := newMemorySkillStore()
	storage.skills["pole/private-skill"] = &skilltypes.Skill{ID: "skill-1", Publisher: "pole", Name: "private-skill", Visibility: skilltypes.VisibilityPrivate, OwnerID: "owner"}
	storage.grants["skill-1"] = []*skilltypes.Grant{{SkillID: "skill-1", PrincipalType: "group", PrincipalID: "group-1"}}
	items, total, err := NewService(storage).Search(context.Background(), "", Principal{Type: "user", ID: "user-2", Expanded: []PrincipalRef{{Type: "group", ID: "group-1"}, {Type: "role", ID: "role-1"}}}, 0, 50)
	require.NoError(t, err)
	require.Equal(t, uint32(1), total)
	require.Len(t, items, 1)
}

func TestManualGitImportCannotBypassPublicSignature(t *testing.T) {
	storage := newMemorySkillStore()
	storage.publishers["pole"] = &skilltypes.Publisher{ID: "publisher-1", Name: "pole", OwnerID: "user-1", Trusted: true}
	_, err := NewService(storage).Publish(context.Background(), PublishRequest{Publisher: "pole", Name: "demo-skill", Version: "1.0.0", Visibility: skilltypes.VisibilityPublic, ManualImport: true, SourceID: "git-import", SourceTrust: skilltypes.TrustTrusted, Actor: Principal{Type: "user", ID: "user-1"}})
	require.ErrorIs(t, err, ErrManualPublicImport)
	require.Empty(t, storage.releases)
}

func TestDownloadBlocksPrivateQuarantinedBundleForOwnerAndGrantee(t *testing.T) {
	storage := newMemorySkillStore()
	storage.skills["pole/private-skill"] = &skilltypes.Skill{ID: "skill-1", Publisher: "pole", Name: "private-skill", Visibility: skilltypes.VisibilityPrivate, OwnerID: "owner"}
	storage.releases["pole/private-skill@1.0.0"] = &skilltypes.Release{ID: "release-1", SkillID: "skill-1", Version: "1.0.0", Digest: "digest", Status: skilltypes.ReleaseQuarantined}
	storage.bundles["digest"] = []byte("unsafe")
	storage.grants["skill-1"] = []*skilltypes.Grant{{SkillID: "skill-1", PrincipalType: "user", PrincipalID: "grantee"}}
	service := NewService(storage)
	_, _, err := service.Download(context.Background(), "pole", "private-skill", "1.0.0", Principal{Type: "user", ID: "owner"})
	require.ErrorIs(t, err, ErrNotFound)
	_, _, err = service.Download(context.Background(), "pole", "private-skill", "1.0.0", Principal{Type: "user", ID: "grantee"})
	require.ErrorIs(t, err, ErrNotFound)
	bundle, _, err := service.Download(context.Background(), "pole", "private-skill", "1.0.0", Principal{Type: "user", ID: "admin", Admin: true})
	require.NoError(t, err)
	require.Equal(t, []byte("unsafe"), bundle)
}
