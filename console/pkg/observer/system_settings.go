package store

import (
	"context"
	"time"
)

// SystemConfigRevision is an immutable Console-owned system configuration
// document. Secrets are referenced by version and never embedded in Payload.
type SystemConfigRevision struct {
	ID              int64
	Component       string
	Domain          string
	Payload         []byte
	SecretVersionID int64
	State           string
	CreatedBy       string
	PublishedBy     string
	CreatedAt       time.Time
	PublishedAt     *time.Time
}

type SystemConfigDomain struct {
	Component      string
	Domain         string
	ActiveRevision *SystemConfigRevision
	DraftRevision  *SystemConfigRevision
}

type EncryptedSecretVersion struct {
	ID          int64
	Component   string
	Domain      string
	Purpose     string
	Ciphertext  []byte
	DataNonce   []byte
	WrappedDEK  []byte
	WrapNonce   []byte
	Fingerprint string
	CreatedBy   string
	CreatedAt   time.Time
}

// SystemSettingsRepository is deliberately narrow so product configuration is
// not coupled to observability SQL details.
type SystemSettingsRepository interface {
	GetSystemConfigDomain(ctx context.Context, component, domain string) (*SystemConfigDomain, error)
	SaveSystemConfigDraft(ctx context.Context, component, domain string, expectedDraftRevision int64,
		payload []byte, secretVersionID int64, actor string) (*SystemConfigRevision, error)
	PublishSystemConfigDraft(ctx context.Context, component, domain string, draftRevision int64,
		actor string) (*SystemConfigRevision, error)
	ListSystemConfigReleases(ctx context.Context, component, domain string, limit int) ([]*SystemConfigRevision, error)
	CreateSystemSecretVersion(ctx context.Context, version *EncryptedSecretVersion) (*EncryptedSecretVersion, error)
	GetSystemSecretVersion(ctx context.Context, id int64) (*EncryptedSecretVersion, error)
	DeleteSystemSecretVersion(ctx context.Context, id int64) error
}
