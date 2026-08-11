package sqldb

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	skilltypes "github.com/pole-io/pole-server/apis/pkg/types/skillmarketplace"
	"github.com/pole-io/pole-server/apis/store"
	market "github.com/pole-io/pole-server/pkg/skillmarketplace"
)

type skillMarketplaceStore struct {
	master *BaseDB
	slave  *BaseDB
}

func newSkillMarketplaceStore(master, slave *BaseDB) *skillMarketplaceStore {
	return &skillMarketplaceStore{master: master, slave: slave}
}

func (s *skillMarketplaceStore) PutBundle(_ context.Context, digest string, bundle []byte) error {
	if len(digest) != 64 || len(bundle) == 0 {
		return store.NewStatusError(store.EmptyParamsErr, "bundle digest and bytes are required")
	}
	result, err := s.master.Exec(`INSERT IGNORE INTO skill_bundle_blob(digest, size, bundle) VALUES (?, ?, ?)`, digest, len(bundle), bundle)
	if err != nil {
		return store.Error(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		var size int64
		var existing []byte
		if err := s.master.QueryRow(`SELECT size, bundle FROM skill_bundle_blob WHERE digest = ?`, digest).Scan(&size, &existing); err != nil {
			return store.Error(err)
		}
		if size != int64(len(bundle)) || !bytes.Equal(existing, bundle) {
			return fmt.Errorf("bundle digest collision: existing size %d, incoming size %d", size, len(bundle))
		}
	}
	return nil
}

func (s *skillMarketplaceStore) GetBundle(_ context.Context, digest string) ([]byte, error) {
	var bundle []byte
	if err := s.slave.QueryRow(`SELECT bundle FROM skill_bundle_blob WHERE digest = ?`, digest).Scan(&bundle); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, market.ErrNotFound
		}
		return nil, store.Error(err)
	}
	return bundle, nil
}

func (s *skillMarketplaceStore) DeleteUnreferencedBundles(_ context.Context, olderThan time.Time) (uint64, error) {
	result, err := s.master.Exec(`DELETE b FROM skill_bundle_blob b
		LEFT JOIN skill_release r ON r.digest = b.digest
		WHERE r.id IS NULL AND b.ctime < ?`, olderThan.UTC())
	if err != nil {
		return 0, store.Error(err)
	}
	count, err := result.RowsAffected()
	return uint64(count), store.Error(err)
}

func (s *skillMarketplaceStore) UpsertSkillPublisher(_ context.Context, publisher *skilltypes.Publisher) error {
	_, err := s.master.Exec(`INSERT INTO skill_publisher
		(id, name, display_name, owner_id, public_key, public_key_version, key_revoked, trusted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE display_name = VALUES(display_name), owner_id = VALUES(owner_id),
		public_key = VALUES(public_key), public_key_version = VALUES(public_key_version),
		key_revoked = VALUES(key_revoked), trusted = VALUES(trusted), mtime = sysdate()`,
		publisher.ID, publisher.Name, publisher.DisplayName, publisher.OwnerID, publisher.PublicKey,
		publisher.PublicKeyVersion, publisher.KeyRevoked, publisher.Trusted)
	return store.Error(err)
}

func (s *skillMarketplaceStore) GetSkillPublisher(_ context.Context, name string) (*skilltypes.Publisher, error) {
	row := s.slave.QueryRow(`SELECT id, name, display_name, owner_id, IFNULL(public_key, ''),
		public_key_version, key_revoked, trusted, ctime, mtime FROM skill_publisher WHERE name = ?`, name)
	publisher := &skilltypes.Publisher{}
	if err := row.Scan(&publisher.ID, &publisher.Name, &publisher.DisplayName, &publisher.OwnerID,
		&publisher.PublicKey, &publisher.PublicKeyVersion, &publisher.KeyRevoked, &publisher.Trusted,
		&publisher.CreateTime, &publisher.ModifyTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, market.ErrNotFound
		}
		return nil, store.Error(err)
	}
	return publisher, nil
}

func (s *skillMarketplaceStore) AddSkillPublisherKey(_ context.Context, key *skilltypes.PublisherKey) error {
	_, err := s.master.Exec(`INSERT INTO skill_publisher_key(publisher_id, key_version, public_key, revoked, revoked_at)
		VALUES (?, ?, ?, ?, ?)`, key.PublisherID, key.Version, key.PublicKey, key.Revoked, nullableTime(key.RevokedAt))
	return store.Error(err)
}

func (s *skillMarketplaceStore) GetSkillPublisherKey(_ context.Context, publisherID string, version uint32) (*skilltypes.PublisherKey, error) {
	key := &skilltypes.PublisherKey{}
	var revokedAt sql.NullTime
	err := s.slave.QueryRow(`SELECT publisher_id, key_version, public_key, revoked, revoked_at, ctime
		FROM skill_publisher_key WHERE publisher_id = ? AND key_version = ?`, publisherID, version).Scan(
		&key.PublisherID, &key.Version, &key.PublicKey, &key.Revoked, &revokedAt, &key.CreateTime)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, market.ErrNotFound
	}
	if err != nil {
		return nil, store.Error(err)
	}
	if revokedAt.Valid {
		key.RevokedAt = revokedAt.Time
	}
	return key, nil
}

func (s *skillMarketplaceStore) RevokeSkillPublisherKey(_ context.Context, publisherID string, version uint32, revokedAt time.Time) error {
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.Exec(`UPDATE skill_publisher_key SET revoked = 1, revoked_at = ?
		WHERE publisher_id = ? AND key_version = ? AND revoked = 0`, revokedAt.UTC(), publisherID, version)
	if err != nil {
		return store.Error(err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return market.ErrNotFound
	}
	if _, err := tx.Exec(`UPDATE skill_publisher SET key_revoked = 1 WHERE id = ? AND public_key_version = ?`, publisherID, version); err != nil {
		return store.Error(err)
	}
	return tx.Commit()
}

func (s *skillMarketplaceStore) ListSkillPublisherMembers(_ context.Context, publisherID string) ([]*skilltypes.PublisherMember, error) {
	rows, err := s.slave.Query(`SELECT publisher_id, principal_type, principal_id, member_role
		FROM skill_publisher_member WHERE publisher_id = ?`, publisherID)
	if err != nil {
		return nil, store.Error(err)
	}
	defer rows.Close()
	members := make([]*skilltypes.PublisherMember, 0)
	for rows.Next() {
		member := &skilltypes.PublisherMember{}
		if err := rows.Scan(&member.PublisherID,
			&member.PrincipalType, &member.PrincipalID, &member.Role); err != nil {
			return nil, store.Error(err)
		}
		members = append(members, member)
	}
	return members, store.Error(rows.Err())
}

func (s *skillMarketplaceStore) ReplaceSkillPublisherMembers(_ context.Context, publisherID string,
	members []*skilltypes.PublisherMember) error {
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM skill_publisher_member WHERE publisher_id = ?`, publisherID); err != nil {
		return store.Error(err)
	}
	for _, member := range members {
		if member == nil || member.PrincipalID == "" {
			continue
		}
		if _, err := tx.Exec(
			`INSERT INTO skill_publisher_member(publisher_id, principal_type, principal_id, member_role) VALUES (?, ?, ?, ?)`,
			publisherID, member.PrincipalType, member.PrincipalID, member.Role); err != nil {
			return store.Error(err)
		}
	}
	return tx.Commit()
}

func (s *skillMarketplaceStore) CreateSkillRelease(_ context.Context, skill *skilltypes.Skill, release *skilltypes.Release) error {
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()
	var skillID, visibility, ownerID string
	err = tx.QueryRow(`SELECT id, visibility, owner_id FROM skill_marketplace_skill
		WHERE publisher_id = ? AND name = ? FOR UPDATE`, skill.PublisherID, skill.Name).Scan(&skillID, &visibility, &ownerID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		metadata, marshalErr := json.Marshal(skill.Metadata)
		if marshalErr != nil {
			return marshalErr
		}
		if len(metadata) == 0 || string(metadata) == "null" {
			metadata = []byte("{}")
		}
		_, err = tx.Exec(`INSERT INTO skill_marketplace_skill
			(id, publisher_id, name, description, visibility, owner_id, registry_source_id, metadata_json)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, skill.ID, skill.PublisherID, skill.Name,
			skill.Description, skill.Visibility, skill.OwnerID, skill.RegistrySourceID, metadata)
		skillID = skill.ID
	case err != nil:
		return store.Error(err)
	case visibility != string(skill.Visibility) || ownerID != skill.OwnerID:
		return market.ErrForbidden
	default:
		_, err = tx.Exec(`UPDATE skill_marketplace_skill SET description = CASE WHEN description = '' THEN ? ELSE description END,
			mtime = sysdate() WHERE id = ?`, skill.Description, skillID)
	}
	if err != nil {
		return store.Error(err)
	}
	release.SkillID = skillID
	var bundleExists int
	if err := tx.QueryRow(`SELECT 1 FROM skill_bundle_blob WHERE digest = ?`, release.Digest).Scan(&bundleExists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("release bundle blob does not exist")
		}
		return store.Error(err)
	}
	_, err = tx.Exec(`INSERT INTO skill_release
		(id, skill_id, version, digest, bundle_size, status, yanked, deprecated, signature,
		signer_key_version, signed_at, published_at, registry_source_id, source_digest, scan_status, scan_evidence, create_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, release.ID, release.SkillID,
		release.Version, release.Digest, release.BundleSize, release.Status, release.Yanked,
		release.Deprecated, release.Signature, release.SignerKeyVersion, nullableTime(release.SignedAt),
		nullableTime(release.PublishedAt), release.RegistrySourceID, release.SourceDigest,
		release.ScanStatus, release.ScanEvidence, release.CreateBy)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return market.ErrVersionExists
		}
		return store.Error(err)
	}
	return tx.Commit()
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC()
}

func (s *skillMarketplaceStore) GetSkill(_ context.Context, publisher, name string) (*skilltypes.Skill, error) {
	row := s.slave.QueryRow(`SELECT s.id, s.publisher_id, p.name, s.name, s.description, s.visibility,
		s.owner_id, s.registry_source_id, s.metadata_json, s.ctime, s.mtime
		FROM skill_marketplace_skill s JOIN skill_publisher p ON p.id = s.publisher_id
		WHERE p.name = ? AND s.name = ?`, publisher, name)
	return scanSkill(row)
}

type skillRowScanner interface{ Scan(...any) error }

func scanSkill(row skillRowScanner) (*skilltypes.Skill, error) {
	skill := &skilltypes.Skill{}
	var metadata []byte
	if err := row.Scan(&skill.ID, &skill.PublisherID, &skill.Publisher, &skill.Name,
		&skill.Description, &skill.Visibility, &skill.OwnerID, &skill.RegistrySourceID,
		&metadata, &skill.CreateTime, &skill.ModifyTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, market.ErrNotFound
		}
		return nil, store.Error(err)
	}
	_ = json.Unmarshal(metadata, &skill.Metadata)
	return skill, nil
}

func (s *skillMarketplaceStore) SearchSkills(ctx context.Context, query skilltypes.SearchQuery) ([]*skilltypes.Skill, uint32, error) {
	like := "%" + query.Query + "%"
	access := `(s.visibility = 'public' AND EXISTS (SELECT 1 FROM skill_release pr WHERE pr.skill_id = s.id AND pr.status = 'published')) OR s.owner_id = ?`
	filter := `(? = '' OR p.name LIKE ? OR s.name LIKE ? OR s.description LIKE ?)`
	args := []any{query.ViewerID}
	for _, principal := range query.Principals {
		if principal.PrincipalID == "" {
			continue
		}
		access += ` OR EXISTS (SELECT 1 FROM skill_access_grant g WHERE g.skill_id = s.id AND g.principal_type = ? AND g.principal_id = ?)`
		args = append(args, principal.PrincipalType, principal.PrincipalID)
	}
	args = append(args, query.Query, like, like, like)
	var total uint32
	if err := s.slave.QueryRow(`SELECT COUNT(*) FROM skill_marketplace_skill s JOIN skill_publisher p ON p.id = s.publisher_id
		WHERE (`+access+`) AND `+filter, args...).Scan(&total); err != nil {
		return nil, 0, store.Error(err)
	}
	rows, err := s.slave.Query(`SELECT s.id, s.publisher_id, p.name, s.name, s.description, s.visibility,
		s.owner_id, s.registry_source_id, s.metadata_json, s.ctime, s.mtime
		FROM skill_marketplace_skill s JOIN skill_publisher p ON p.id = s.publisher_id
		WHERE (`+access+`) AND `+filter+` ORDER BY p.name, s.name LIMIT ? OFFSET ?`,
		append(args, query.Limit, query.Offset)...)
	if err != nil {
		return nil, 0, store.Error(err)
	}
	defer rows.Close()
	items := make([]*skilltypes.Skill, 0)
	for rows.Next() {
		skill, err := scanSkill(rows)
		if err != nil {
			return nil, 0, err
		}
		releases, err := s.ListSkillReleases(ctx, skill.ID)
		if err != nil {
			return nil, 0, err
		}
		for _, release := range releases {
			if release.Status == skilltypes.ReleasePublished && !release.Yanked &&
				(skill.LatestVersion == "" || market.CompareVersions(release.Version, skill.LatestVersion) > 0) {
				skill.LatestVersion, skill.LatestDigest = release.Version, release.Digest
			}
		}
		items = append(items, skill)
	}
	return items, total, store.Error(rows.Err())
}

func (s *skillMarketplaceStore) ListSkillReleases(_ context.Context, skillID string) ([]*skilltypes.Release, error) {
	rows, err := s.slave.Query(`SELECT id, skill_id, version, digest, bundle_size, status, yanked,
		deprecated, IFNULL(signature, ''), signer_key_version, signed_at, published_at,
		registry_source_id, source_digest, scan_status, scan_evidence, create_by, ctime FROM skill_release WHERE skill_id = ?`, skillID)
	if err != nil {
		return nil, store.Error(err)
	}
	defer rows.Close()
	items := make([]*skilltypes.Release, 0)
	for rows.Next() {
		release, err := scanRelease(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, release)
	}
	return items, store.Error(rows.Err())
}

func (s *skillMarketplaceStore) GetSkillRelease(_ context.Context, publisher, name, version string) (*skilltypes.Release, error) {
	row := s.slave.QueryRow(`SELECT r.id, r.skill_id, r.version, r.digest, r.bundle_size, r.status, r.yanked,
		r.deprecated, IFNULL(r.signature, ''), r.signer_key_version, r.signed_at, r.published_at,
		r.registry_source_id, r.source_digest, r.scan_status, r.scan_evidence, r.create_by, r.ctime
		FROM skill_release r JOIN skill_marketplace_skill s ON s.id = r.skill_id
		JOIN skill_publisher p ON p.id = s.publisher_id WHERE p.name = ? AND s.name = ? AND r.version = ?`,
		publisher, name, version)
	return scanRelease(row)
}

func scanRelease(row skillRowScanner) (*skilltypes.Release, error) {
	release := &skilltypes.Release{}
	var signedAt, publishedAt sql.NullTime
	if err := row.Scan(&release.ID, &release.SkillID, &release.Version, &release.Digest,
		&release.BundleSize, &release.Status, &release.Yanked, &release.Deprecated, &release.Signature,
		&release.SignerKeyVersion, &signedAt, &publishedAt, &release.RegistrySourceID,
		&release.SourceDigest, &release.ScanStatus, &release.ScanEvidence, &release.CreateBy, &release.CreateTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, market.ErrNotFound
		}
		return nil, store.Error(err)
	}
	if signedAt.Valid {
		release.SignedAt = signedAt.Time
	}
	if publishedAt.Valid {
		release.PublishedAt = publishedAt.Time
	}
	return release, nil
}

func (s *skillMarketplaceStore) SetSkillReleaseLifecycle(_ context.Context, skillID, version string, yanked bool, deprecated string) error {
	result, err := s.master.Exec(`UPDATE skill_release SET yanked = ?, deprecated = ? WHERE skill_id = ? AND version = ?`,
		yanked, deprecated, skillID, version)
	if err != nil {
		return store.Error(err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return market.ErrNotFound
	}
	return nil
}

func (s *skillMarketplaceStore) CreateSkillReview(_ context.Context, review *skilltypes.Review,
	status skilltypes.ReleaseStatus, publishedAt time.Time) error {
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.Exec(`UPDATE skill_release SET status = ?, published_at = ?
		WHERE id = ? AND status = 'pending_review' AND scan_status = 'passed'`, status, nullableTime(publishedAt), review.ReleaseID)
	if err != nil {
		return store.Error(err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("release is not pending review")
	}
	_, err = tx.Exec(`INSERT INTO skill_release_review(id, release_id, reviewer_id, decision, comment)
		VALUES (?, ?, ?, ?, ?)`, review.ID, review.ReleaseID, review.ReviewerID, review.Decision, review.Comment)
	if err != nil {
		return store.Error(err)
	}
	return tx.Commit()
}

func (s *skillMarketplaceStore) ListSkillReviews(_ context.Context, status skilltypes.ReleaseStatus,
	limit uint32) ([]*skilltypes.ReviewQueueItem, error) {
	if limit == 0 || limit > 200 {
		limit = 100
	}
	rows, err := s.slave.Query(`SELECT r.id, p.name, s.name, r.version, r.digest, r.status,
		CASE WHEN LENGTH(IFNULL(r.signature, '')) > 0 THEN 'verified' ELSE 'missing' END,
		r.scan_status, r.scan_evidence, r.ctime FROM skill_release r
		JOIN skill_marketplace_skill s ON s.id = r.skill_id JOIN skill_publisher p ON p.id = s.publisher_id
		WHERE (? = '' OR r.status = ?) ORDER BY r.ctime LIMIT ?`, status, status, limit)
	if err != nil {
		return nil, store.Error(err)
	}
	defer rows.Close()
	items := make([]*skilltypes.ReviewQueueItem, 0)
	for rows.Next() {
		item := &skilltypes.ReviewQueueItem{}
		if err := rows.Scan(&item.ReleaseID,
			&item.Publisher, &item.Name, &item.Version, &item.Digest, &item.Status, &item.SignatureStatus,
			&item.ScanStatus, &item.ScanEvidence, &item.RequestedAt); err != nil {
			return nil, store.Error(err)
		}
		item.ID = item.ReleaseID
		items = append(items, item)
	}
	return items, store.Error(rows.Err())
}

func (s *skillMarketplaceStore) ListSkillGrants(_ context.Context, skillID string) ([]*skilltypes.Grant, error) {
	rows, err := s.slave.Query(`SELECT skill_id, principal_type, principal_id FROM skill_access_grant WHERE skill_id = ?`, skillID)
	if err != nil {
		return nil, store.Error(err)
	}
	defer rows.Close()
	items := make([]*skilltypes.Grant, 0)
	for rows.Next() {
		grant := &skilltypes.Grant{}
		if err := rows.Scan(&grant.SkillID, &grant.PrincipalType, &grant.PrincipalID); err != nil {
			return nil, store.Error(err)
		}
		items = append(items, grant)
	}
	return items, store.Error(rows.Err())
}

func (s *skillMarketplaceStore) ReplaceSkillGrants(_ context.Context, skillID string, grants []*skilltypes.Grant) error {
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM skill_access_grant WHERE skill_id = ?`, skillID); err != nil {
		return store.Error(err)
	}
	for _, grant := range grants {
		if grant == nil || grant.PrincipalID == "" {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO skill_access_grant(skill_id, principal_type, principal_id) VALUES (?, ?, ?)`,
			skillID, grant.PrincipalType, grant.PrincipalID); err != nil {
			return store.Error(err)
		}
	}
	return tx.Commit()
}

func (s *skillMarketplaceStore) CreateRegistrySource(_ context.Context, source *skilltypes.RegistrySource) error {
	_, err := s.master.Exec(`INSERT INTO skill_registry_source
		(id, name, type, url, enabled, trust_level, sync_interval_seconds, cursor_value,
		sync_status, last_sync_error, last_sync_at, next_sync_at, create_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, source.ID, source.Name, source.Type,
		source.URL, source.Enabled, source.TrustLevel, source.SyncIntervalSeconds, source.Cursor,
		source.SyncStatus, source.LastSyncError, nullableTime(source.LastSyncTime), source.NextSyncTime, source.CreateBy)
	return store.Error(err)
}

func (s *skillMarketplaceStore) UpdateRegistrySource(_ context.Context, source *skilltypes.RegistrySource) error {
	result, err := s.master.Exec(`UPDATE skill_registry_source SET name = ?, type = ?, url = ?, enabled = ?,
		trust_level = ?, sync_interval_seconds = ?, next_sync_at = ?, mtime = sysdate() WHERE id = ?`,
		source.Name, source.Type, source.URL, source.Enabled, source.TrustLevel,
		source.SyncIntervalSeconds, source.NextSyncTime, source.ID)
	if err != nil {
		return store.Error(err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return market.ErrNotFound
	}
	return nil
}

func (s *skillMarketplaceStore) DeleteRegistrySource(_ context.Context, id string) error {
	result, err := s.master.Exec(`DELETE FROM skill_registry_source WHERE id = ?`, id)
	if err != nil {
		return store.Error(err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return market.ErrNotFound
	}
	return nil
}

func (s *skillMarketplaceStore) GetRegistrySource(_ context.Context, id string) (*skilltypes.RegistrySource, error) {
	return scanRegistrySource(s.slave.QueryRow(registrySourceSelect+` WHERE id = ?`, id))
}

func (s *skillMarketplaceStore) ListRegistrySources(_ context.Context) ([]*skilltypes.RegistrySource, error) {
	rows, err := s.slave.Query(registrySourceSelect + ` ORDER BY name`)
	if err != nil {
		return nil, store.Error(err)
	}
	defer rows.Close()
	items := make([]*skilltypes.RegistrySource, 0)
	for rows.Next() {
		item, err := scanRegistrySource(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, store.Error(rows.Err())
}

func (s *skillMarketplaceStore) ListDueRegistrySources(_ context.Context, now time.Time, limit uint32) ([]*skilltypes.RegistrySource, error) {
	rows, err := s.master.Query(registrySourceSelect+` WHERE enabled = 1 AND sync_status != 'running'
		AND next_sync_at <= ? ORDER BY next_sync_at LIMIT ?`, now.UTC(), limit)
	if err != nil {
		return nil, store.Error(err)
	}
	defer rows.Close()
	items := make([]*skilltypes.RegistrySource, 0)
	for rows.Next() {
		item, err := scanRegistrySource(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, store.Error(rows.Err())
}

const registrySourceSelect = `SELECT id, name, type, url, enabled, trust_level, sync_interval_seconds,
	cursor_value, sync_status, last_sync_error, last_sync_at, next_sync_at, create_by, ctime, mtime FROM skill_registry_source`

func scanRegistrySource(row skillRowScanner) (*skilltypes.RegistrySource, error) {
	source := &skilltypes.RegistrySource{}
	var lastSync sql.NullTime
	if err := row.Scan(&source.ID, &source.Name, &source.Type, &source.URL, &source.Enabled,
		&source.TrustLevel, &source.SyncIntervalSeconds, &source.Cursor, &source.SyncStatus,
		&source.LastSyncError, &lastSync, &source.NextSyncTime, &source.CreateBy,
		&source.CreateTime, &source.ModifyTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, market.ErrNotFound
		}
		return nil, store.Error(err)
	}
	if lastSync.Valid {
		source.LastSyncTime = lastSync.Time
	}
	return source, nil
}

func (s *skillMarketplaceStore) MarkRegistrySyncStarted(_ context.Context, id string, startedAt time.Time) error {
	result, err := s.master.Exec(`UPDATE skill_registry_source SET sync_status = 'running', last_sync_error = '',
		mtime = ? WHERE id = ? AND sync_status != 'running'`, startedAt.UTC(), id)
	if err != nil {
		return store.Error(err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("registry source is already syncing or missing")
	}
	return nil
}

func (s *skillMarketplaceStore) MarkRegistrySyncFinished(_ context.Context, id, cursor, syncError string,
	finishedAt, nextSync time.Time) error {
	status := "idle"
	if syncError != "" {
		status = "failed"
	}
	_, err := s.master.Exec(`UPDATE skill_registry_source SET cursor_value = ?, sync_status = ?,
		last_sync_error = ?, last_sync_at = ?, next_sync_at = ?, mtime = sysdate() WHERE id = ?`,
		cursor, status, syncError, finishedAt.UTC(), nextSync.UTC(), id)
	return store.Error(err)
}

func (s *skillMarketplaceStore) RecoverInterruptedRegistrySyncs(_ context.Context, staleBefore time.Time) error {
	_, err := s.master.Exec(`UPDATE skill_registry_source SET sync_status = 'failed',
		last_sync_error = 'interrupted registry sync recovered after restart', next_sync_at = sysdate()
		WHERE sync_status = 'running' AND mtime < ?`, staleBefore.UTC())
	return store.Error(err)
}
