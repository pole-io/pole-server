package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	store "github.com/pole-io/pole-server/console/pkg/observer"
)

var errSystemConfigConflict = errors.New("system configuration revision conflict")

func ensureSystemSettingsSchema(db *BaseDB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS console_system_config_domain (
			component VARCHAR(64) NOT NULL,
			domain_name VARCHAR(64) NOT NULL,
			active_revision BIGINT NULL,
			draft_revision BIGINT NULL,
			mtime DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
			PRIMARY KEY (component, domain_name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`,
		`CREATE TABLE IF NOT EXISTS console_system_config_revision (
			id BIGINT NOT NULL AUTO_INCREMENT,
			component VARCHAR(64) NOT NULL,
			domain_name VARCHAR(64) NOT NULL,
			payload_json JSON NOT NULL,
			secret_version_id BIGINT NOT NULL DEFAULT 0,
			state VARCHAR(16) NOT NULL,
			created_by VARCHAR(128) NOT NULL,
			published_by VARCHAR(128) NOT NULL DEFAULT '',
			ctime DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			published_at DATETIME(3) NULL,
			PRIMARY KEY (id),
			KEY idx_console_system_config_scope (component, domain_name, id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`,
		`CREATE TABLE IF NOT EXISTS console_system_secret_version (
			id BIGINT NOT NULL AUTO_INCREMENT,
			component VARCHAR(64) NOT NULL,
			domain_name VARCHAR(64) NOT NULL,
			purpose VARCHAR(128) NOT NULL,
			ciphertext LONGBLOB NOT NULL,
			data_nonce VARBINARY(32) NOT NULL,
			wrapped_dek VARBINARY(128) NOT NULL,
			wrap_nonce VARBINARY(32) NOT NULL,
			fingerprint VARCHAR(64) NOT NULL,
			created_by VARCHAR(128) NOT NULL,
			ctime DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			PRIMARY KEY (id),
			KEY idx_console_system_secret_scope (component, domain_name, purpose, id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *stableStore) GetSystemConfigDomain(ctx context.Context, component, domain string) (*store.SystemConfigDomain, error) {
	result := &store.SystemConfigDomain{Component: component, Domain: domain}
	var activeID, draftID sql.NullInt64
	err := s.slave.QueryRowContext(ctx,
		`SELECT active_revision, draft_revision FROM console_system_config_domain WHERE component = ? AND domain_name = ?`,
		component, domain).Scan(&activeID, &draftID)
	if errors.Is(err, sql.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return nil, err
	}
	if activeID.Valid {
		revision, loadErr := s.getSystemConfigRevision(ctx, activeID.Int64)
		if loadErr != nil {
			return nil, loadErr
		}
		result.ActiveRevision = revision
	}
	if draftID.Valid {
		revision, loadErr := s.getSystemConfigRevision(ctx, draftID.Int64)
		if loadErr != nil {
			return nil, loadErr
		}
		result.DraftRevision = revision
	}
	return result, nil
}

func (s *stableStore) getSystemConfigRevision(ctx context.Context, id int64) (*store.SystemConfigRevision, error) {
	revision := &store.SystemConfigRevision{}
	var createdAt, publishedAt []byte
	err := s.slave.QueryRowContext(ctx, `SELECT id, component, domain_name, payload_json, secret_version_id,
		state, created_by, published_by, ctime, published_at
		FROM console_system_config_revision WHERE id = ?`, id).Scan(
		&revision.ID, &revision.Component, &revision.Domain, &revision.Payload, &revision.SecretVersionID,
		&revision.State, &revision.CreatedBy, &revision.PublishedBy, &createdAt, &publishedAt)
	if err != nil {
		return nil, err
	}
	revision.CreatedAt, err = parseSystemSettingsTime(createdAt)
	if err != nil {
		return nil, err
	}
	if len(publishedAt) > 0 {
		value, parseErr := parseSystemSettingsTime(publishedAt)
		if parseErr != nil {
			return nil, parseErr
		}
		revision.PublishedAt = &value
	}
	return revision, nil
}

func (s *stableStore) SaveSystemConfigDraft(ctx context.Context, component, domain string,
	expectedDraftRevision int64, payload []byte, secretVersionID int64, actor string) (*store.SystemConfigRevision, error) {
	tx, err := s.master.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT IGNORE INTO console_system_config_domain(component, domain_name) VALUES(?, ?)`,
		component, domain); err != nil {
		return nil, err
	}
	var current sql.NullInt64
	if err = tx.QueryRowContext(ctx, `SELECT draft_revision FROM console_system_config_domain
		WHERE component = ? AND domain_name = ? FOR UPDATE`, component, domain).Scan(&current); err != nil {
		return nil, err
	}
	if (current.Valid && current.Int64 != expectedDraftRevision) || (!current.Valid && expectedDraftRevision != 0) {
		return nil, errSystemConfigConflict
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO console_system_config_revision
		(component, domain_name, payload_json, secret_version_id, state, created_by)
		VALUES(?, ?, ?, ?, 'draft', ?)`, component, domain, payload, secretVersionID, actor)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	if current.Valid {
		if _, err = tx.ExecContext(ctx, `UPDATE console_system_config_revision SET state = 'superseded'
			WHERE id = ? AND state = 'draft'`, current.Int64); err != nil {
			return nil, err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE console_system_config_domain SET draft_revision = ?
		WHERE component = ? AND domain_name = ?`, id, component, domain); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return s.getSystemConfigRevision(ctx, id)
}

func (s *stableStore) PublishSystemConfigDraft(ctx context.Context, component, domain string,
	draftRevision int64, actor string) (*store.SystemConfigRevision, error) {
	tx, err := s.master.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var current sql.NullInt64
	if err = tx.QueryRowContext(ctx, `SELECT draft_revision FROM console_system_config_domain
		WHERE component = ? AND domain_name = ? FOR UPDATE`, component, domain).Scan(&current); err != nil {
		return nil, err
	}
	if !current.Valid || current.Int64 != draftRevision {
		return nil, errSystemConfigConflict
	}
	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `UPDATE console_system_config_revision
		SET state = 'published', published_by = ?, published_at = ? WHERE id = ? AND state = 'draft'`,
		actor, now, draftRevision); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE console_system_config_domain
		SET active_revision = ?, draft_revision = NULL WHERE component = ? AND domain_name = ?`,
		draftRevision, component, domain); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return s.getSystemConfigRevision(ctx, draftRevision)
}

func (s *stableStore) ListSystemConfigReleases(ctx context.Context, component, domain string, limit int) ([]*store.SystemConfigRevision, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.slave.QueryContext(ctx, `SELECT id, component, domain_name, payload_json, secret_version_id,
		state, created_by, published_by, ctime, published_at
		FROM console_system_config_revision WHERE component = ? AND domain_name = ? AND state = 'published'
		ORDER BY id DESC LIMIT ?`, component, domain, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*store.SystemConfigRevision, 0)
	for rows.Next() {
		revision := &store.SystemConfigRevision{}
		var createdAt, publishedAt []byte
		if err := rows.Scan(&revision.ID, &revision.Component, &revision.Domain, &revision.Payload,
			&revision.SecretVersionID, &revision.State, &revision.CreatedBy, &revision.PublishedBy,
			&createdAt, &publishedAt); err != nil {
			return nil, err
		}
		revision.CreatedAt, err = parseSystemSettingsTime(createdAt)
		if err != nil {
			return nil, err
		}
		if len(publishedAt) > 0 {
			value, parseErr := parseSystemSettingsTime(publishedAt)
			if parseErr != nil {
				return nil, parseErr
			}
			revision.PublishedAt = &value
		}
		result = append(result, revision)
	}
	return result, rows.Err()
}

func (s *stableStore) CreateSystemSecretVersion(ctx context.Context,
	version *store.EncryptedSecretVersion) (*store.EncryptedSecretVersion, error) {
	result, err := s.master.ExecContext(ctx, `INSERT INTO console_system_secret_version
		(component, domain_name, purpose, ciphertext, data_nonce, wrapped_dek, wrap_nonce, fingerprint, created_by)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`, version.Component, version.Domain, version.Purpose,
		version.Ciphertext, version.DataNonce, version.WrappedDEK, version.WrapNonce,
		version.Fingerprint, version.CreatedBy)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetSystemSecretVersion(ctx, id)
}

func (s *stableStore) GetSystemSecretVersion(ctx context.Context, id int64) (*store.EncryptedSecretVersion, error) {
	version := &store.EncryptedSecretVersion{}
	var createdAt []byte
	err := s.slave.QueryRowContext(ctx, `SELECT id, component, domain_name, purpose, ciphertext,
		data_nonce, wrapped_dek, wrap_nonce, fingerprint, created_by, ctime
		FROM console_system_secret_version WHERE id = ?`, id).Scan(
		&version.ID, &version.Component, &version.Domain, &version.Purpose, &version.Ciphertext,
		&version.DataNonce, &version.WrappedDEK, &version.WrapNonce, &version.Fingerprint,
		&version.CreatedBy, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("get system secret version %d: %w", id, err)
	}
	version.CreatedAt, err = parseSystemSettingsTime(createdAt)
	if err != nil {
		return nil, fmt.Errorf("get system secret version %d time: %w", id, err)
	}
	return version, nil
}

func (s *stableStore) DeleteSystemSecretVersion(ctx context.Context, id int64) error {
	if id <= 0 {
		return nil
	}
	_, err := s.master.ExecContext(ctx, `DELETE FROM console_system_secret_version
		WHERE id = ? AND NOT EXISTS (
			SELECT 1 FROM console_system_config_revision WHERE secret_version_id = ?
		)`, id, id)
	return err
}

func parseSystemSettingsTime(value []byte) (time.Time, error) {
	for _, layout := range []string{"2006-01-02 15:04:05.999999", "2006-01-02 15:04:05"} {
		if parsed, err := time.ParseInLocation(layout, string(value), time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid system settings timestamp %q", string(value))
}
