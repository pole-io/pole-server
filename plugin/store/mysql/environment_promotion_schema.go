package sqldb

func ensureEnvironmentPromotionSchema(db *BaseDB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS environment_promotion_topology (
	id tinyint unsigned NOT NULL,
	draft_revision bigint unsigned NOT NULL DEFAULT 0,
	published_revision bigint unsigned NOT NULL DEFAULT 0,
	draft_json longtext COLLATE utf8_bin NOT NULL,
	modify_by varchar(64) COLLATE utf8_bin NOT NULL DEFAULT '',
	mtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp,
	PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='global environment promotion topology draft'`,
		`CREATE TABLE IF NOT EXISTS environment_promotion_topology_revision (
	revision bigint unsigned NOT NULL,
	topology_json longtext COLLATE utf8_bin NOT NULL,
	comment varchar(512) COLLATE utf8_bin NOT NULL DEFAULT '',
	create_by varchar(64) COLLATE utf8_bin NOT NULL DEFAULT '',
	ctime timestamp NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY (revision)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='immutable environment promotion topology revisions'`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}
