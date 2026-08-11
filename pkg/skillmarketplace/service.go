package skillmarketplace

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/auth"
	skilltypes "github.com/pole-io/pole-server/apis/pkg/types/skillmarketplace"
	"github.com/pole-io/pole-server/apis/pkg/utils"
	"github.com/pole-io/pole-server/apis/store"
)

var (
	ErrNotFound           = errors.New("skill marketplace resource not found")
	ErrForbidden          = errors.New("skill marketplace access denied")
	ErrVersionExists      = errors.New("skill release version is immutable and already exists")
	ErrDigestConflict     = errors.New("upstream reused a version with a different digest")
	ErrManualPublicImport = errors.New("public Git import requires detached-signature upload")
	slugPattern           = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9._-]{0,62}[a-z0-9])?$`)
)

type Principal struct {
	Type     string
	ID       string
	Admin    bool
	Expanded []PrincipalRef
}

type PrincipalRef struct{ Type, ID string }

type PublishRequest struct {
	Publisher    string
	Name         string
	Description  string
	Version      string
	Visibility   skilltypes.Visibility
	Bundle       []byte
	Signature    []byte
	SignedAt     time.Time
	Actor        Principal
	SourceID     string
	SourceDigest string
	SourceTrust  skilltypes.TrustLevel
	ManualImport bool
}

type GitImportItemResult struct {
	Path        string              `json:"path"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Version     string              `json:"version"`
	Digest      string              `json:"digest"`
	Entries     int                 `json:"entries"`
	Size        int64               `json:"size"`
	Status      string              `json:"status"`
	Release     *skilltypes.Release `json:"release,omitempty"`
	Error       string              `json:"error,omitempty"`
}

type GitImportResult struct {
	RepositoryURL string                `json:"repository_url"`
	Reference     string                `json:"reference"`
	Tag           string                `json:"tag"`
	CommitSHA     string                `json:"commit_sha"`
	RootPath      string                `json:"root_path"`
	Version       string                `json:"version"`
	Items         []GitImportItemResult `json:"items"`
	Succeeded     int                   `json:"succeeded"`
	Skipped       int                   `json:"skipped"`
	Failed        int                   `json:"failed"`
}

type Service struct {
	store store.SkillMarketplaceStore
	now   func() time.Time
}

func NewService(storage store.SkillMarketplaceStore) *Service {
	return &Service{store: storage, now: time.Now}
}

func (s *Service) Store() store.SkillMarketplaceStore { return s.store }

func (s *Service) CanPublishPublisher(ctx context.Context, name string, actor Principal) bool {
	publisher, err := s.store.GetSkillPublisher(ctx, name)
	return err == nil && (actor.Admin || s.canPublish(ctx, publisher, actor))
}

func (s *Service) UpsertPublisher(ctx context.Context, publisher *skilltypes.Publisher, actor Principal) error {
	if publisher == nil || !slugPattern.MatchString(publisher.Name) {
		return fmt.Errorf("invalid publisher name")
	}
	if len(publisher.PublicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("publisher Ed25519 public key must be %d bytes", ed25519.PublicKeySize)
	}
	if actor.ID == "" {
		return ErrForbidden
	}
	existing, err := s.store.GetSkillPublisher(ctx, publisher.Name)
	if err == nil && existing != nil && !actor.Admin && existing.OwnerID != actor.ID {
		return ErrForbidden
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if existing != nil {
		publisher.ID = existing.ID
		if !actor.Admin {
			publisher.Trusted = existing.Trusted
		}
	}
	if publisher.ID == "" {
		publisher.ID = utils.NewUUID()
	}
	if publisher.OwnerID == "" || !actor.Admin {
		publisher.OwnerID = actor.ID
	}
	if existing != nil && publisher.PublicKeyVersion <= existing.PublicKeyVersion {
		publisher.PublicKeyVersion = existing.PublicKeyVersion + 1
	}
	if publisher.PublicKeyVersion == 0 {
		publisher.PublicKeyVersion = 1
	}
	if len(publisher.PublicKey) > 0 {
		if err := s.store.AddSkillPublisherKey(ctx, &skilltypes.PublisherKey{
			PublisherID: publisher.ID, Version: publisher.PublicKeyVersion, PublicKey: publisher.PublicKey,
		}); err != nil {
			return err
		}
	}
	return s.store.UpsertSkillPublisher(ctx, publisher)
}

func (s *Service) Publish(ctx context.Context, request PublishRequest) (*skilltypes.Release, error) {
	if !slugPattern.MatchString(request.Publisher) || !slugPattern.MatchString(request.Name) {
		return nil, fmt.Errorf("publisher and skill name must be lowercase slugs")
	}
	if err := ValidateVersion(request.Version); err != nil {
		return nil, err
	}
	if request.Visibility != skilltypes.VisibilityPrivate && request.Visibility != skilltypes.VisibilityPublic {
		return nil, fmt.Errorf("visibility must be private or public")
	}
	if request.ManualImport && request.Visibility == skilltypes.VisibilityPublic {
		return nil, ErrManualPublicImport
	}
	publisher, err := s.store.GetSkillPublisher(ctx, request.Publisher)
	if err != nil {
		return nil, err
	}
	if (request.SourceID == "" || request.ManualImport) && !request.Actor.Admin && !s.canPublish(ctx, publisher, request.Actor) {
		return nil, ErrForbidden
	}
	bundle, err := ValidateAndNormalizeBundle(request.Bundle)
	if err != nil {
		return nil, err
	}
	if bundle.Name != request.Name {
		return nil, fmt.Errorf("SKILL.md name %q does not match release name %q", bundle.Name, request.Name)
	}
	if request.SourceDigest != "" {
		sourceSum := sha256.Sum256(request.Bundle)
		if !strings.EqualFold(request.SourceDigest, hex.EncodeToString(sourceSum[:])) {
			return nil, fmt.Errorf("source digest does not match downloaded bundle bytes")
		}
	}
	existing, getErr := s.store.GetSkillRelease(ctx, request.Publisher, request.Name, request.Version)
	if getErr == nil && existing != nil {
		if existing.Digest != bundle.Digest ||
			(request.SourceDigest != "" && existing.SourceDigest != "" &&
				!strings.EqualFold(existing.SourceDigest, request.SourceDigest)) {
			return nil, ErrDigestConflict
		}
		return nil, ErrVersionExists
	}
	if getErr != nil && !errors.Is(getErr, ErrNotFound) {
		return nil, getErr
	}
	if request.SourceID == "" && request.Visibility == skilltypes.VisibilityPublic {
		key, err := s.store.GetSkillPublisherKey(ctx, publisher.ID, publisher.PublicKeyVersion)
		if err != nil {
			return nil, err
		}
		if publisher.KeyRevoked || key.Revoked {
			return nil, fmt.Errorf("publisher signing key is revoked")
		}
		if err := VerifyDetachedSignature(key.PublicKey, request.Signature,
			request.Publisher, request.Name, request.Version, bundle.Digest, request.SignedAt); err != nil {
			return nil, fmt.Errorf("verify signature for canonical bundle digest %s: %w", bundle.Digest, err)
		}
	}
	now := s.now().UTC()
	scanStatus, scanEvidence, err := ScanBundle(bundle.Bytes)
	if err != nil {
		return nil, fmt.Errorf("scan normalized bundle: %w", err)
	}
	status := skilltypes.ReleasePublished
	if scanStatus != "passed" {
		status = skilltypes.ReleaseQuarantined
	} else if request.ManualImport && request.Visibility == skilltypes.VisibilityPrivate {
		status = skilltypes.ReleasePublished
	} else if request.SourceID != "" && request.SourceTrust != skilltypes.TrustTrusted {
		status = skilltypes.ReleaseQuarantined
	} else if request.Visibility == skilltypes.VisibilityPublic && !publisher.Trusted {
		status = skilltypes.ReleasePendingReview
	}
	skill := &skilltypes.Skill{
		ID: utils.NewUUID(), PublisherID: publisher.ID, Publisher: publisher.Name,
		Name: request.Name, Description: request.Description, Visibility: request.Visibility,
		OwnerID: publisher.OwnerID, RegistrySourceID: request.SourceID,
	}
	if skill.Description == "" {
		skill.Description = bundle.Description
	}
	release := &skilltypes.Release{
		ID: utils.NewUUID(), SkillID: skill.ID, Publisher: publisher.Name, Name: request.Name,
		Version: request.Version, Digest: bundle.Digest, BundleSize: int64(len(bundle.Bytes)),
		Status: status, Signature: request.Signature, SignerKeyVersion: publisher.PublicKeyVersion,
		SignedAt: request.SignedAt.UTC(), RegistrySourceID: request.SourceID,
		SourceDigest: request.SourceDigest, CreateBy: request.Actor.ID, CreateTime: now,
		ScanStatus: scanStatus, ScanEvidence: scanEvidence,
	}
	if status == skilltypes.ReleasePublished {
		release.PublishedAt = now
	}
	if err := s.store.PutBundle(ctx, bundle.Digest, bundle.Bytes); err != nil {
		return nil, err
	}
	if err := s.store.CreateSkillRelease(ctx, skill, release); err != nil {
		return nil, err
	}
	return release, nil
}

func (s *Service) ImportGitDiscovery(ctx context.Context, discovery *GitSkillDiscovery, publisher string,
	visibility skilltypes.Visibility, actor Principal) *GitImportResult {
	result := &GitImportResult{
		RepositoryURL: discovery.RepositoryURL, Reference: discovery.Reference, Tag: discovery.Tag,
		CommitSHA: discovery.CommitSHA, RootPath: discovery.RootPath, Version: discovery.Version,
		Items: make([]GitImportItemResult, 0, len(discovery.Items)),
	}
	for _, item := range discovery.Items {
		entry := GitImportItemResult{
			Path: item.Path, Name: item.Name, Description: item.Description, Version: discovery.Version,
			Digest: item.Digest, Entries: item.Entries, Size: item.Size,
		}
		if item.Error != "" {
			entry.Status, entry.Error = "failed", item.Error
			result.Failed++
			result.Items = append(result.Items, entry)
			continue
		}
		release, err := s.Publish(ctx, PublishRequest{
			Publisher: publisher, Name: item.Name, Version: discovery.Version,
			Visibility: visibility, Bundle: item.Bundle, Actor: actor,
			SourceID: "git-import", SourceDigest: item.Digest,
			SourceTrust: skilltypes.TrustUntrusted, ManualImport: true,
		})
		switch {
		case err == nil:
			entry.Status, entry.Release = "succeeded", release
			result.Succeeded++
		case errors.Is(err, ErrVersionExists):
			entry.Status, entry.Error = "skipped", err.Error()
			result.Skipped++
		default:
			entry.Status, entry.Error = "failed", err.Error()
			result.Failed++
		}
		result.Items = append(result.Items, entry)
	}
	return result
}

func (s *Service) Search(ctx context.Context, query string, viewer Principal, offset, limit uint32) ([]*skilltypes.Skill, uint32, error) {
	if limit == 0 || limit > 100 {
		limit = 50
	}
	principals := make([]skilltypes.Grant, 0, 1+len(viewer.Expanded))
	principals = append(principals, skilltypes.Grant{PrincipalType: viewer.Type, PrincipalID: viewer.ID})
	for _, principal := range viewer.Expanded {
		principals = append(principals, skilltypes.Grant{PrincipalType: principal.Type, PrincipalID: principal.ID})
	}
	return s.store.SearchSkills(ctx, skilltypes.SearchQuery{
		Query: strings.TrimSpace(query), ViewerType: viewer.Type, ViewerID: viewer.ID,
		Principals: principals, Offset: offset, Limit: limit,
	})
}

func (s *Service) Get(ctx context.Context, publisher, name string, viewer Principal) (*skilltypes.Skill, error) {
	skill, err := s.store.GetSkill(ctx, publisher, name)
	if err != nil {
		return nil, err
	}
	if !s.canRead(ctx, skill, viewer) {
		return nil, ErrNotFound
	}
	releases, err := s.store.ListSkillReleases(ctx, skill.ID)
	if err != nil {
		return nil, err
	}
	visible := releases[:0]
	for _, release := range releases {
		if skill.Visibility == skilltypes.VisibilityPublic && !viewer.Admin && viewer.ID != skill.OwnerID && release.Status != skilltypes.ReleasePublished {
			continue
		}
		visible = append(visible, release)
	}
	sort.Slice(visible, func(i, j int) bool { return CompareVersions(visible[i].Version, visible[j].Version) > 0 })
	skill.Releases = visible
	return skill, nil
}

func (s *Service) Download(ctx context.Context, publisher, name, version string, viewer Principal) ([]byte, *skilltypes.Release, error) {
	skill, err := s.store.GetSkill(ctx, publisher, name)
	if err != nil || !s.canRead(ctx, skill, viewer) {
		return nil, nil, ErrNotFound
	}
	release, err := s.store.GetSkillRelease(ctx, publisher, name, version)
	if err != nil {
		return nil, nil, err
	}
	if !viewer.Admin && release.Status != skilltypes.ReleasePublished {
		return nil, nil, ErrNotFound
	}
	bundle, err := s.store.GetBundle(ctx, release.Digest)
	return bundle, release, err
}

func (s *Service) Review(ctx context.Context, releaseID string, decision skilltypes.ReviewDecision, comment string, reviewer Principal) error {
	if !reviewer.Admin {
		return ErrForbidden
	}
	status := skilltypes.ReleaseRejected
	var publishedAt time.Time
	if decision == skilltypes.ReviewApproved {
		status = skilltypes.ReleasePublished
		publishedAt = s.now().UTC()
	} else if decision != skilltypes.ReviewRejected {
		return fmt.Errorf("invalid review decision")
	}
	return s.store.CreateSkillReview(ctx, &skilltypes.Review{
		ID: utils.NewUUID(), ReleaseID: releaseID, ReviewerID: reviewer.ID,
		Decision: decision, Comment: comment,
	}, status, publishedAt)
}

func (s *Service) SetLifecycle(ctx context.Context, publisher, name, version string, yanked bool, deprecated string, actor Principal) error {
	skill, err := s.store.GetSkill(ctx, publisher, name)
	if err != nil {
		return err
	}
	if !actor.Admin && (actor.ID == "" || actor.ID != skill.OwnerID) {
		return ErrForbidden
	}
	return s.store.SetSkillReleaseLifecycle(ctx, skill.ID, version, yanked, deprecated)
}

func (s *Service) canRead(ctx context.Context, skill *skilltypes.Skill, viewer Principal) bool {
	if skill == nil {
		return false
	}
	if skill.Visibility == skilltypes.VisibilityPublic || viewer.Admin || (viewer.ID != "" && skill.OwnerID == viewer.ID) {
		return true
	}
	grants, err := s.store.ListSkillGrants(ctx, skill.ID)
	if err != nil {
		return false
	}
	principals := append([]PrincipalRef{{Type: viewer.Type, ID: viewer.ID}}, viewer.Expanded...)
	for _, grant := range grants {
		for _, principal := range principals {
			if grant.PrincipalType == principal.Type && grant.PrincipalID == principal.ID {
				return true
			}
		}
	}
	return false
}

func (s *Service) canPublish(ctx context.Context, publisher *skilltypes.Publisher, actor Principal) bool {
	if publisher.OwnerID == actor.ID {
		return true
	}
	members, err := s.store.ListSkillPublisherMembers(ctx, publisher.ID)
	if err != nil {
		return false
	}
	principals := append([]PrincipalRef{{Type: actor.Type, ID: actor.ID}}, actor.Expanded...)
	for _, member := range members {
		if member.Role != "owner" && member.Role != "publisher" {
			continue
		}
		for _, principal := range principals {
			if member.PrincipalType == principal.Type && member.PrincipalID == principal.ID {
				return true
			}
		}
	}
	return false
}

func PrincipalFromAcquireContext(authCtx *auth.AcquireContext) Principal {
	if authCtx == nil {
		return Principal{}
	}
	value, ok := authCtx.GetAttachment(auth.PrincipalKey)
	if !ok {
		return Principal{}
	}
	principal, ok := value.(auth.Principal)
	if !ok {
		return Principal{}
	}
	return Principal{Type: auth.PrincipalNames[principal.PrincipalType], ID: principal.PrincipalID}
}
