package sqldb

func ensureNamespaceKindSchema(db *BaseDB) error {
	hasKind, err := hasTableColumn(db, "namespace", "kind")
	if err != nil {
		return err
	}
	if !hasKind {
		if _, err := db.Exec(`ALTER TABLE namespace ADD COLUMN kind
			tinyint(4) NOT NULL DEFAULT 0 COMMENT 'namespace kind: 0 business, 1 system'
			AFTER metadata`); err != nil {
			return err
		}
	}
	_, err = db.Exec(`UPDATE namespace SET kind = 1 WHERE name = 'pole-system' AND kind <> 1`)
	return err
}
