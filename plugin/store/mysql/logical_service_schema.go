package sqldb

func ensureLogicalServiceSchema(db *BaseDB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS logical_service (
	id varchar(32) NOT NULL COMMENT 'control-plane logical service id',
	name varchar(128) COLLATE utf8_bin NOT NULL COMMENT 'logical service display name',
	comment varchar(1024) DEFAULT NULL,
	owner varchar(1024) NOT NULL DEFAULT '',
	business varchar(64) DEFAULT NULL,
	department varchar(1024) DEFAULT NULL,
	revision varchar(32) NOT NULL,
	flag tinyint(4) NOT NULL DEFAULT 0,
	active_name varchar(128) COLLATE utf8_bin GENERATED ALWAYS AS
		(CASE WHEN flag = 0 THEN name ELSE NULL END) STORED,
	ctime timestamp NOT NULL DEFAULT current_timestamp,
	mtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp,
	PRIMARY KEY (id),
	UNIQUE KEY uk_logical_service_active_name (active_name),
	KEY idx_logical_service_mtime (mtime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='control-plane logical services'`,
		`CREATE TABLE IF NOT EXISTS service_environment_binding (
	logical_service_id varchar(32) NOT NULL,
	service_id varchar(32) NOT NULL,
	namespace varchar(64) COLLATE utf8_bin NOT NULL,
	service_name varchar(128) COLLATE utf8_bin NOT NULL,
	ctime timestamp NOT NULL DEFAULT current_timestamp,
	mtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp,
	PRIMARY KEY (service_id),
	UNIQUE KEY uk_logical_service_namespace (logical_service_id, namespace),
	KEY idx_environment_binding_logical (logical_service_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='explicit environment service bindings'`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	if ok, err := hasTableColumn(db, "logical_service", "active_name"); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`ALTER TABLE logical_service ADD COLUMN active_name varchar(128)
			COLLATE utf8_bin GENERATED ALWAYS AS
			(CASE WHEN flag = 0 THEN name ELSE NULL END) STORED`); err != nil {
			return err
		}
	}
	if ok, err := hasTableIndex(db, "logical_service", "uk_logical_service_active_name"); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`ALTER TABLE logical_service
			ADD UNIQUE KEY uk_logical_service_active_name (active_name)`); err != nil {
			return err
		}
	}
	return nil
}
