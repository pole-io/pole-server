/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License.
 */

package sqldb

func ensureServiceIdentitySchema(db *BaseDB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS service_identity (
		service_id varchar(32) NOT NULL COMMENT 'Service resource ID',
		subject varchar(255) NOT NULL COMMENT 'Stable internal data-plane subject',
		revision varchar(32) NOT NULL COMMENT 'Identity descriptor revision',
		ctime timestamp NOT NULL DEFAULT current_timestamp COMMENT 'Create time',
		mtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp COMMENT 'Last updated time',
		PRIMARY KEY (service_id),
		UNIQUE KEY subject (subject)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='service internal data-plane identity'`); err != nil {
		return err
	}

	// Identity bootstrap resolves the authenticated service by its control-plane
	// token. A prefix index avoids a full service-table scan; the equality
	// predicate still verifies the complete token and detects duplicates.
	if ok, err := hasTableIndex(db, "service", "token"); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`alter table service add key token (token(255))`); err != nil {
			return err
		}
	}
	return nil
}
