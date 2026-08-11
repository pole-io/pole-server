package skillmarketplace

import "time"

type Visibility string

const (
	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
)

type ReleaseStatus string

const (
	ReleasePendingReview ReleaseStatus = "pending_review"
	ReleasePublished     ReleaseStatus = "published"
	ReleaseRejected      ReleaseStatus = "rejected"
	ReleaseQuarantined   ReleaseStatus = "quarantined"
)

type ReviewDecision string

const (
	ReviewApproved ReviewDecision = "approved"
	ReviewRejected ReviewDecision = "rejected"
)

type RegistryType string

const (
	RegistryPole RegistryType = "pole"
	RegistryHTTP RegistryType = "http_index"
	RegistryGit  RegistryType = "git"
)

type TrustLevel string

const (
	TrustUntrusted TrustLevel = "untrusted"
	TrustTrusted   TrustLevel = "trusted"
)

type Publisher struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	DisplayName      string    `json:"displayName,omitempty"`
	OwnerID          string    `json:"ownerId,omitempty"`
	PublicKey        []byte    `json:"-"`
	PublicKeyVersion uint32    `json:"publicKeyVersion"`
	KeyRevoked       bool      `json:"keyRevoked"`
	Trusted          bool      `json:"trusted"`
	CreateTime       time.Time `json:"createTime"`
	ModifyTime       time.Time `json:"modifyTime"`
}

type PublisherKey struct {
	PublisherID string    `json:"publisherId"`
	Version     uint32    `json:"version"`
	PublicKey   []byte    `json:"-"`
	Revoked     bool      `json:"revoked"`
	RevokedAt   time.Time `json:"revokedAt,omitempty"`
	CreateTime  time.Time `json:"createTime"`
}

type PublisherMember struct {
	PublisherID   string `json:"publisherId"`
	PrincipalType string `json:"principalType"`
	PrincipalID   string `json:"principalId"`
	Role          string `json:"role"`
}

type Skill struct {
	ID               string            `json:"id"`
	PublisherID      string            `json:"publisherId,omitempty"`
	Publisher        string            `json:"publisher"`
	Name             string            `json:"name"`
	Description      string            `json:"description,omitempty"`
	Visibility       Visibility        `json:"visibility"`
	OwnerID          string            `json:"ownerId,omitempty"`
	LatestVersion    string            `json:"latestVersion,omitempty"`
	LatestDigest     string            `json:"latestDigest,omitempty"`
	RegistrySourceID string            `json:"registrySourceId,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	CreateTime       time.Time         `json:"createTime"`
	ModifyTime       time.Time         `json:"modifyTime"`
	Releases         []*Release        `json:"releases,omitempty"`
}

type Release struct {
	ID               string        `json:"id"`
	SkillID          string        `json:"skillId,omitempty"`
	Publisher        string        `json:"publisher,omitempty"`
	Name             string        `json:"name,omitempty"`
	Version          string        `json:"version"`
	Digest           string        `json:"digest"`
	BundleSize       int64         `json:"bundleSize"`
	Status           ReleaseStatus `json:"status"`
	Yanked           bool          `json:"yanked"`
	Deprecated       string        `json:"deprecated,omitempty"`
	Signature        []byte        `json:"-"`
	SignerKeyVersion uint32        `json:"signerKeyVersion"`
	SignedAt         time.Time     `json:"signedAt"`
	PublishedAt      time.Time     `json:"publishedAt,omitempty"`
	RegistrySourceID string        `json:"registrySourceId,omitempty"`
	SourceDigest     string        `json:"sourceDigest,omitempty"`
	ScanStatus       string        `json:"scanStatus"`
	ScanEvidence     string        `json:"scanEvidence,omitempty"`
	CreateBy         string        `json:"createBy,omitempty"`
	CreateTime       time.Time     `json:"createTime"`
}

type Review struct {
	ID         string         `json:"id"`
	ReleaseID  string         `json:"releaseId"`
	ReviewerID string         `json:"reviewerId"`
	Decision   ReviewDecision `json:"decision"`
	Comment    string         `json:"comment,omitempty"`
	CreateTime time.Time      `json:"createTime"`
}

type ReviewQueueItem struct {
	ID              string        `json:"id"`
	ReleaseID       string        `json:"releaseId"`
	Publisher       string        `json:"publisher"`
	Name            string        `json:"name"`
	Version         string        `json:"version"`
	Digest          string        `json:"digest"`
	Status          ReleaseStatus `json:"status"`
	SignatureStatus string        `json:"signatureStatus"`
	ScanStatus      string        `json:"scanStatus"`
	ScanEvidence    string        `json:"scanEvidence,omitempty"`
	RequestedAt     time.Time     `json:"requestedAt"`
}

type RegistrySource struct {
	ID                  string       `json:"id"`
	Name                string       `json:"name"`
	Type                RegistryType `json:"type"`
	URL                 string       `json:"url"`
	Enabled             bool         `json:"enabled"`
	TrustLevel          TrustLevel   `json:"trustLevel"`
	SyncIntervalSeconds uint32       `json:"syncIntervalSeconds"`
	Cursor              string       `json:"cursor,omitempty"`
	SyncStatus          string       `json:"syncStatus"`
	LastSyncError       string       `json:"lastSyncError,omitempty"`
	LastSyncTime        time.Time    `json:"lastSyncTime,omitempty"`
	NextSyncTime        time.Time    `json:"nextSyncTime,omitempty"`
	CreateBy            string       `json:"createBy,omitempty"`
	CreateTime          time.Time    `json:"createTime"`
	ModifyTime          time.Time    `json:"modifyTime"`
}

type BundleBlob struct {
	Digest     string    `json:"digest"`
	Size       int64     `json:"size"`
	RefCount   uint64    `json:"refCount"`
	CreateTime time.Time `json:"createTime"`
}

type Grant struct {
	SkillID       string `json:"skillId"`
	PrincipalType string `json:"principalType"`
	PrincipalID   string `json:"principalId"`
}

type SearchQuery struct {
	Query      string
	ViewerType string
	ViewerID   string
	Principals []Grant
	Offset     uint32
	Limit      uint32
}

type RemoteRelease struct {
	Publisher   string            `json:"publisher"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Version     string            `json:"version"`
	Digest      string            `json:"digest"`
	BundleURL   string            `json:"bundleUrl"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
