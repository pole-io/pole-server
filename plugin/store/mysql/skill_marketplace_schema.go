package sqldb

func ensureSkillMarketplaceSchema(db *BaseDB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS skill_publisher (
	id varchar(36) NOT NULL,
	name varchar(64) COLLATE utf8mb4_bin NOT NULL,
	display_name varchar(128) NOT NULL DEFAULT '',
	owner_id varchar(36) NOT NULL,
	public_key varbinary(64) DEFAULT NULL,
	public_key_version int unsigned NOT NULL DEFAULT 1,
	key_revoked tinyint(1) NOT NULL DEFAULT 0,
	trusted tinyint(1) NOT NULL DEFAULT 0,
	ctime timestamp NOT NULL DEFAULT current_timestamp,
	mtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp,
	PRIMARY KEY (id), UNIQUE KEY uk_skill_publisher_name (name), KEY idx_skill_publisher_owner (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`,
		`CREATE TABLE IF NOT EXISTS skill_marketplace_skill (
	id varchar(36) NOT NULL,
	publisher_id varchar(36) NOT NULL,
	name varchar(64) COLLATE utf8mb4_bin NOT NULL,
	description varchar(2048) NOT NULL DEFAULT '',
	visibility varchar(16) NOT NULL,
	owner_id varchar(36) NOT NULL,
	registry_source_id varchar(36) NOT NULL DEFAULT '',
	metadata_json longtext NOT NULL,
	ctime timestamp NOT NULL DEFAULT current_timestamp,
	mtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp,
	PRIMARY KEY (id), UNIQUE KEY uk_marketplace_skill_identity (publisher_id, name),
	KEY idx_marketplace_skill_visibility (visibility), KEY idx_marketplace_skill_owner (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`,
		`CREATE TABLE IF NOT EXISTS skill_publisher_key (
	publisher_id varchar(36) NOT NULL,
	key_version int unsigned NOT NULL,
	public_key varbinary(64) NOT NULL,
	revoked tinyint(1) NOT NULL DEFAULT 0,
	revoked_at datetime DEFAULT NULL,
	ctime timestamp NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY (publisher_id, key_version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`,
		`CREATE TABLE IF NOT EXISTS skill_publisher_member (
	publisher_id varchar(36) NOT NULL,
	principal_type varchar(16) NOT NULL,
	principal_id varchar(36) NOT NULL,
	member_role varchar(16) NOT NULL,
	ctime timestamp NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY (publisher_id, principal_type, principal_id),
	KEY idx_skill_publisher_member_principal (principal_type, principal_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`,
		`CREATE TABLE IF NOT EXISTS skill_bundle_blob (
	digest char(64) COLLATE ascii_bin NOT NULL,
	size bigint unsigned NOT NULL,
	bundle longblob NOT NULL,
	ctime timestamp NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY (digest), KEY idx_skill_bundle_ctime (ctime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`,
		`CREATE TABLE IF NOT EXISTS skill_release (
	id varchar(36) NOT NULL,
	skill_id varchar(36) NOT NULL,
	version varchar(128) COLLATE utf8mb4_bin NOT NULL,
	digest char(64) COLLATE ascii_bin NOT NULL,
	bundle_size bigint unsigned NOT NULL,
	status varchar(32) NOT NULL,
	yanked tinyint(1) NOT NULL DEFAULT 0,
	deprecated varchar(1024) NOT NULL DEFAULT '',
	signature varbinary(128) DEFAULT NULL,
	signer_key_version int unsigned NOT NULL DEFAULT 0,
	signed_at datetime DEFAULT NULL,
	published_at datetime DEFAULT NULL,
	registry_source_id varchar(36) NOT NULL DEFAULT '',
	source_digest char(64) COLLATE ascii_bin NOT NULL DEFAULT '',
	scan_status varchar(16) NOT NULL DEFAULT 'pending',
	scan_evidence longtext NOT NULL,
	create_by varchar(36) NOT NULL DEFAULT '',
	ctime timestamp NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY (id), UNIQUE KEY uk_skill_release_version (skill_id, version),
	KEY idx_skill_release_digest (digest), KEY idx_skill_release_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`,
		`CREATE TABLE IF NOT EXISTS skill_release_review (
	id varchar(36) NOT NULL,
	release_id varchar(36) NOT NULL,
	reviewer_id varchar(36) NOT NULL,
	decision varchar(16) NOT NULL,
	comment varchar(2048) NOT NULL DEFAULT '',
	ctime timestamp NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY (id), KEY idx_skill_review_release (release_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`,
		`CREATE TABLE IF NOT EXISTS skill_access_grant (
	skill_id varchar(36) NOT NULL,
	principal_type varchar(16) NOT NULL,
	principal_id varchar(36) NOT NULL,
	ctime timestamp NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY (skill_id, principal_type, principal_id), KEY idx_skill_grant_principal (principal_type, principal_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`,
		`CREATE TABLE IF NOT EXISTS skill_registry_source (
	id varchar(36) NOT NULL,
	name varchar(128) COLLATE utf8mb4_bin NOT NULL,
	type varchar(32) NOT NULL,
	url varchar(2048) NOT NULL,
	enabled tinyint(1) NOT NULL DEFAULT 1,
	trust_level varchar(16) NOT NULL DEFAULT 'untrusted',
	sync_interval_seconds int unsigned NOT NULL DEFAULT 3600,
	cursor_value varchar(2048) NOT NULL DEFAULT '',
	sync_status varchar(16) NOT NULL DEFAULT 'idle',
	last_sync_error varchar(2048) NOT NULL DEFAULT '',
	last_sync_at datetime DEFAULT NULL,
	next_sync_at datetime NOT NULL,
	create_by varchar(36) NOT NULL DEFAULT '',
	ctime timestamp NOT NULL DEFAULT current_timestamp,
	mtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp,
	PRIMARY KEY (id), UNIQUE KEY uk_skill_registry_name (name), KEY idx_skill_registry_due (enabled, next_sync_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}
