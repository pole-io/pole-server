package sqldb

type aiEnvironmentSchema struct {
	table           string
	definitionIndex string
	environmentKey  string
	backendIndex    string
}

func ensureAIResourceDefinitionSchema(db *BaseDB) error {
	statements := []string{
		aiResourceDefinitionDDL("mcp_server_definition", "MCP server"),
		aiResourceDefinitionDDL("a2a_agent_definition", "A2A agent"),
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}

	environments := []aiEnvironmentSchema{
		{
			table:           "mcp_server",
			definitionIndex: "idx_mcp_server_definition",
			environmentKey:  "uk_mcp_server_definition_namespace",
			backendIndex:    "idx_mcp_server_backend_service_id",
		},
		{
			table:           "a2a_agent",
			definitionIndex: "idx_a2a_agent_definition",
			environmentKey:  "uk_a2a_agent_definition_namespace",
			backendIndex:    "idx_a2a_agent_backend_service_id",
		},
	}
	for _, environment := range environments {
		if err := ensureAIEnvironmentDefinitionColumn(db, environment); err != nil {
			return err
		}
		if err := backfillAIBackendServiceID(db, environment.table); err != nil {
			return err
		}
	}
	return nil
}

func aiResourceDefinitionDDL(table, label string) string {
	return `CREATE TABLE IF NOT EXISTS ` + table + ` (
		id varchar(32) NOT NULL COMMENT 'control-plane logical definition id',
		name varchar(128) COLLATE utf8_bin NOT NULL COMMENT '` + label + ` logical name',
		description varchar(1024) DEFAULT NULL,
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
		UNIQUE KEY uk_active_name (active_name),
		KEY idx_mtime (mtime)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='` + label + ` logical definitions'`
}

func ensureAIEnvironmentDefinitionColumn(db *BaseDB, schema aiEnvironmentSchema) error {
	if ok, err := hasTableColumn(db, schema.table, "definition_id"); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`ALTER TABLE ` + schema.table +
			` ADD COLUMN definition_id varchar(32) DEFAULT NULL COMMENT 'control-plane logical definition id'`); err != nil {
			return err
		}
	}
	if ok, err := hasTableIndex(db, schema.table, schema.definitionIndex); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`ALTER TABLE ` + schema.table +
			` ADD KEY ` + schema.definitionIndex + ` (definition_id)`); err != nil {
			return err
		}
	}
	if ok, err := hasTableIndex(db, schema.table, schema.environmentKey); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`ALTER TABLE ` + schema.table +
			` ADD UNIQUE KEY ` + schema.environmentKey + ` (definition_id, namespace)`); err != nil {
			return err
		}
	}
	if ok, err := hasTableColumn(db, schema.table, "backend_service_id"); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`ALTER TABLE ` + schema.table +
			` ADD COLUMN backend_service_id varchar(32) DEFAULT NULL COMMENT 'stable Pole backend service id'`); err != nil {
			return err
		}
	}
	if ok, err := hasTableIndex(db, schema.table, schema.backendIndex); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`ALTER TABLE ` + schema.table +
			` ADD KEY ` + schema.backendIndex + ` (backend_service_id)`); err != nil {
			return err
		}
	}
	return nil
}

func backfillAIBackendServiceID(db *BaseDB, table string) error {
	_, err := db.Exec(`UPDATE ` + table + ` AS ai
		INNER JOIN service AS backend
			ON backend.namespace = ai.backend_service_namespace
			AND backend.name = ai.backend_service_name
			AND backend.flag != 1
			AND IFNULL(backend.reference, '') = ''
		SET ai.backend_service_id = backend.id
		WHERE ai.backend_service_id IS NULL
			AND ai.backend_type = 'service'
			AND ai.flag != 1
			AND ai.namespace = backend.namespace`)
	return err
}
