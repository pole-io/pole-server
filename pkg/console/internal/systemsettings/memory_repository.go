package systemsettings

import (
	"context"
	"errors"
	"sync"
	"time"

	store "github.com/pole-io/pole-server/pkg/console/internal/observer"
)

// MemoryRepository is used only when unit-test configurations intentionally do
// not configure the Console database. Production initialization always uses
// the configured Pole MySQL repository.
type MemoryRepository struct {
	mu        sync.Mutex
	nextID    int64
	domain    *store.SystemConfigDomain
	revisions map[int64]*store.SystemConfigRevision
	secrets   map[int64]*store.EncryptedSecretVersion
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		nextID: 1, domain: &store.SystemConfigDomain{Component: AgentComponent, Domain: AgentDomain},
		revisions: map[int64]*store.SystemConfigRevision{}, secrets: map[int64]*store.EncryptedSecretVersion{},
	}
}

func (r *MemoryRepository) GetSystemConfigDomain(context.Context, string, string) (*store.SystemConfigDomain, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := *r.domain
	result.ActiveRevision = cloneRevision(r.domain.ActiveRevision)
	result.DraftRevision = cloneRevision(r.domain.DraftRevision)
	return &result, nil
}

func (r *MemoryRepository) SaveSystemConfigDraft(_ context.Context, component, domain string,
	expected int64, payload []byte, secretID int64, actor string) (*store.SystemConfigRevision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	current := int64(0)
	if r.domain.DraftRevision != nil {
		current = r.domain.DraftRevision.ID
	}
	if current != expected {
		return nil, errors.New("system configuration revision conflict")
	}
	revision := &store.SystemConfigRevision{
		ID: r.nextID, Component: component, Domain: domain, Payload: append([]byte(nil), payload...),
		SecretVersionID: secretID, State: "draft", CreatedBy: actor, CreatedAt: time.Now().UTC(),
	}
	r.nextID++
	r.revisions[revision.ID] = revision
	r.domain.DraftRevision = revision
	return cloneRevision(revision), nil
}

func (r *MemoryRepository) PublishSystemConfigDraft(_ context.Context, _, _ string,
	draftID int64, actor string) (*store.SystemConfigRevision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.domain.DraftRevision == nil || r.domain.DraftRevision.ID != draftID {
		return nil, errors.New("system configuration revision conflict")
	}
	now := time.Now().UTC()
	revision := r.domain.DraftRevision
	revision.State, revision.PublishedBy, revision.PublishedAt = "published", actor, &now
	r.domain.ActiveRevision, r.domain.DraftRevision = revision, nil
	return cloneRevision(revision), nil
}

func (r *MemoryRepository) ListSystemConfigReleases(context.Context, string, string, int) ([]*store.SystemConfigRevision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*store.SystemConfigRevision, 0)
	for id := r.nextID - 1; id > 0; id-- {
		if revision := r.revisions[id]; revision != nil && revision.State == "published" {
			result = append(result, cloneRevision(revision))
		}
	}
	return result, nil
}

func (r *MemoryRepository) CreateSystemSecretVersion(_ context.Context,
	version *store.EncryptedSecretVersion) (*store.EncryptedSecretVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := *version
	clone.ID, clone.CreatedAt = r.nextID, time.Now().UTC()
	r.nextID++
	r.secrets[clone.ID] = &clone
	return cloneSecret(&clone), nil
}

func (r *MemoryRepository) GetSystemSecretVersion(_ context.Context, id int64) (*store.EncryptedSecretVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	version := r.secrets[id]
	if version == nil {
		return nil, errors.New("system secret version not found")
	}
	return cloneSecret(version), nil
}

func (r *MemoryRepository) DeleteSystemSecretVersion(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, revision := range r.revisions {
		if revision != nil && revision.SecretVersionID == id {
			return nil
		}
	}
	delete(r.secrets, id)
	return nil
}

func cloneRevision(value *store.SystemConfigRevision) *store.SystemConfigRevision {
	if value == nil {
		return nil
	}
	result := *value
	result.Payload = append([]byte(nil), value.Payload...)
	return &result
}

func cloneSecret(value *store.EncryptedSecretVersion) *store.EncryptedSecretVersion {
	result := *value
	result.Ciphertext = append([]byte(nil), value.Ciphertext...)
	result.DataNonce = append([]byte(nil), value.DataNonce...)
	result.WrappedDEK = append([]byte(nil), value.WrappedDEK...)
	result.WrapNonce = append([]byte(nil), value.WrapNonce...)
	return &result
}
